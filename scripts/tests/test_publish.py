"""Release safety checks with fixture assets and mocked gh; never publish."""
import contextlib
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import api
import dev
import publish
import release
from test_portability import fixture_binary


class PublishTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="zenith publish tests ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        (self.root / "docs").mkdir()
        for name in ("README.md", "docs/QUICK_REFERENCE.md", "docs/PLATFORMS.md"):
            (self.root / name).write_text("# Test fixture\n", encoding="utf-8")
        self.dist = self.root / "release"
        for target in dev.TARGETS:
            server, client = self.root / "server", self.root / "client"
            server.write_bytes(fixture_binary(target))
            client.write_bytes(fixture_binary(target) + b"client")
            release.package_target(self.dist, target, server, client, self.root)
        release.combine(self.dist)
        self.environment = patch.dict(os.environ, {"GH_REPO": "example/zenith"})
        self.environment.start()
        self.addCleanup(self.environment.stop)

    def test_missing_repository_fails_before_any_remote_command(self):
        with patch.dict(os.environ, {}, clear=True), patch.object(publish.subprocess, "run") as run:
            with self.assertRaisesRegex(ValueError, "GH_REPO"):
                publish.publish(self.dist, "v1.2.3")
        run.assert_not_called()

    def test_invalid_tag_fails_before_any_remote_command(self):
        with patch.object(publish.subprocess, "run") as run:
            for tag in ("", "--help", "main", "v1;anything", "v1/../../bad"):
                with self.assertRaises(ValueError):
                    publish.publish(self.dist, tag)
        run.assert_not_called()

    def test_mismatched_asset_fails_before_any_remote_command(self):
        (self.dist / release.asset_names("windows/amd64")[0]).write_bytes(b"tampered")
        with patch.object(publish.subprocess, "run") as run:
            with self.assertRaises(ValueError):
                publish.publish(self.dist, "v1.2.3")
        run.assert_not_called()

    def test_public_release_is_never_overwritten(self):
        result = subprocess.CompletedProcess([], 0, json.dumps({"isDraft": False}))
        with patch.object(publish.subprocess, "run", return_value=result) as run:
            with self.assertRaisesRegex(ValueError, "already published"):
                publish.publish(self.dist, "v1.2.3")
        self.assertEqual(run.call_count, 1)

    def test_new_release_stays_draft_until_upload_completes(self):
        missing = subprocess.CompletedProcess([], 1, "")
        success = subprocess.CompletedProcess([], 0, "")
        with patch.object(publish.subprocess, "run", side_effect=[missing, success, success]) as run:
            publish.publish(self.dist, "v1.2.3")
        creation = run.call_args_list[1].args[0]
        self.assertEqual(creation[:4], ["gh", "release", "create", "v1.2.3"])
        self.assertIn("--verify-tag", creation)
        self.assertIn("--draft", creation)
        self.assertIn(str(self.dist / "SHA256SUMS"), creation)
        self.assertEqual(run.call_args_list[2].args[0], ["gh", "release", "edit", "v1.2.3", "--draft=false"])

    def test_existing_draft_receives_assets_before_publication(self):
        draft = subprocess.CompletedProcess([], 0, json.dumps({"isDraft": True}))
        success = subprocess.CompletedProcess([], 0, "")
        with patch.object(publish.subprocess, "run", side_effect=[draft, success, success]) as run:
            publish.publish(self.dist, "v1.2.3")
        upload = run.call_args_list[1].args[0]
        self.assertEqual(upload[:4], ["gh", "release", "upload", "v1.2.3"])
        self.assertIn("--clobber", upload)
        self.assertEqual(run.call_args_list[-1].args[0][-1], "--draft=false")

    def test_failed_upload_does_not_publish(self):
        draft = subprocess.CompletedProcess([], 0, json.dumps({"isDraft": True}))
        failure = subprocess.CalledProcessError(1, ["gh", "release", "upload"])
        with patch.object(publish.subprocess, "run", side_effect=[draft, failure]) as run:
            with self.assertRaises(subprocess.CalledProcessError):
                publish.publish(self.dist, "v1.2.3")
        self.assertEqual(run.call_count, 2)


class HTTPHelperInputTests(unittest.TestCase):
    def test_nonfinite_timeouts_rejected_before_request(self):
        for value in ("nan", "inf", "-inf", "0", "-1"):
            with contextlib.redirect_stderr(io.StringIO()), patch.object(api.urllib.request, "build_opener") as opener:
                with self.assertRaises(SystemExit) as stopped:
                    api.main(["--timeout=" + value, "ping"])
                self.assertEqual(stopped.exception.code, 2)
                opener.assert_not_called()

    def test_empty_query_and_fragment_in_base_url_rejected(self):
        for base in ("http://host?", "http://host#", "http://host/path?", "http://host/path#"):
            with self.assertRaises(ValueError):
                api.build_request(base, "status", None)
