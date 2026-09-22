"""Helper tests use fixture bytes, not substitutes for building/testing Zenith."""
import contextlib
import io
import json
import os
from pathlib import Path
import struct
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
import urllib.parse

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import api
import dev
import processes
import release


def fixture_binary(target):
    data = bytearray(256)
    system, arch = target.split("/")
    if system == "linux":
        data[:6] = b"\x7fELF\x02\x01"
        struct.pack_into("<H", data, 18, {"amd64": 62, "arm64": 183}[arch])
    elif system == "darwin":
        data[:4] = b"\xcf\xfa\xed\xfe"
        struct.pack_into("<I", data, 4, {"amd64": 0x01000007, "arm64": 0x0100000C}[arch])
    else:
        data[:2] = b"MZ"
        struct.pack_into("<I", data, 0x3C, 128)
        data[128:132] = b"PE\x00\x00"
        struct.pack_into("<H", data, 132, {"amd64": 0x8664, "arm64": 0xAA64}[arch])
    return bytes(data)


class DeveloperTests(unittest.TestCase):
    def test_all_six_targets_and_extensions(self):
        self.assertEqual(len(dev.TARGETS), 6)
        for target in dev.TARGETS:
            self.assertEqual(dev.exe_suffix(target), ".exe" if target.startswith("windows/") else "")
        for target in ("macos/amd64", "windows/x64", "linux/386", "../linux"):
            with self.assertRaises(ValueError):
                dev.validate_target(target)

    def test_build_environment_is_explicit_and_not_mutated(self):
        original = {"GOOS": "bad", "GOARCH": "bad", "CGO_ENABLED": "1", "GOAMD64": "v4", "PATH": "test"}
        env = dev.build_env("windows/arm64", base=original)
        self.assertEqual((env["GOOS"], env["GOARCH"], env["CGO_ENABLED"], env["GOAMD64"], env["GOARM64"]),
                         ("windows", "arm64", "0", "v1", "v8.0"))
        self.assertEqual(original["GOOS"], "bad")
        self.assertEqual(env["PATH"], "test")

    def test_race_does_not_silently_skip_unsupported_target(self):
        with self.assertRaisesRegex(ValueError, "windows/arm64"):
            dev.build_env("windows/arm64", race=True)
        for target in dev.RACE_TARGETS:
            self.assertEqual(dev.build_env(target, race=True)["CGO_ENABLED"], "1")

    def test_paths_are_relative_to_repository_not_cwd(self):
        self.assertEqual(dev.local_path("dist"), dev.ROOT / "dist")
        self.assertEqual(dev.local_path(str(Path(tempfile.gettempdir()).resolve())), Path(tempfile.gettempdir()).resolve())

    def test_subprocess_failure_is_propagated_and_logged(self):
        with tempfile.TemporaryDirectory(prefix="zenith tests ") as folder:
            log = Path(folder) / "test \u03a9.log"
            with contextlib.redirect_stdout(io.StringIO()):
                with self.assertRaises(subprocess.CalledProcessError) as failure:
                    dev.run([sys.executable, "-c", "print('evidence'); raise SystemExit(7)"], log=log)
            self.assertEqual(failure.exception.returncode, 7)
            self.assertIn("evidence", log.read_text(encoding="utf-8"))

    def test_unicode_and_space_arguments_are_not_shell_interpreted(self):
        argument = "worker \u03a9 & /+?="
        with tempfile.TemporaryDirectory() as folder:
            log = Path(folder) / "arguments.log"
            with contextlib.redirect_stdout(io.StringIO()):
                dev.run([sys.executable, "-c", "import sys; print(ascii(sys.argv[1]))", argument], log=log)
            self.assertEqual(log.read_text(encoding="utf-8").strip(), ascii(argument))

    def test_process_options_are_native(self):
        with patch.object(processes.subprocess, "Popen") as popen:
            processes.start(["example"])
            kwargs = popen.call_args.kwargs
            if os.name == "nt":
                self.assertEqual(kwargs["creationflags"], subprocess.CREATE_NEW_PROCESS_GROUP)
            else:
                self.assertIs(kwargs["start_new_session"], True)

    def test_windows_control_branch_is_explicit(self):
        # This tests dispatch only. Actual Windows events are tested in native E2E.
        with patch.object(processes.os, "name", "nt"), \
             patch.object(processes.subprocess, "CREATE_NEW_PROCESS_GROUP", 512, create=True), \
             patch.object(processes.subprocess, "Popen") as popen:
            processes.start(["example.exe"])
            self.assertEqual(popen.call_args.kwargs["creationflags"], 512)
        from unittest.mock import Mock
        child = Mock()
        child.poll.return_value = None
        with patch.object(processes.os, "name", "nt"), \
             patch.object(processes.signal, "CTRL_BREAK_EVENT", 1, create=True):
            processes.interrupt(child, server=True)
        child.send_signal.assert_called_once_with(1)


class PackagingTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="zenith package tests ")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.docs = self.root / "source"
        (self.docs / "docs").mkdir(parents=True)
        for name in ("README.md", "docs/QUICK_REFERENCE.md", "docs/PLATFORMS.md"):
            (self.docs / name).write_bytes(b"# Fixture documentation\r\n\r\nTest only.\r\n")
        self.dist = self.root / "release \u03a9"

    def pack(self, target):
        server, client = self.root / "server", self.root / "client"
        server.write_bytes(fixture_binary(target))
        client.write_bytes(fixture_binary(target) + b"client")
        return release.package_target(self.dist, target, server, client, self.docs)

    def test_all_target_packages_checksums_headers_and_modes(self):
        for target in dev.TARGETS:
            with self.subTest(target=target):
                archive = self.pack(target)
                release.verify_target(self.dist, target)
                members = release.read_bundle(archive, target)
                self.assertEqual(len(members), 5)
                for name, (data, mode) in members.items():
                    if name.endswith(".md"):
                        self.assertNotIn(b"\r\n", data)
                        self.assertEqual(mode, 0o644)
                    else:
                        self.assertEqual(mode, 0o755)
                extracted = release.unpack(self.dist, target, self.root / ("unpacked-" + target.replace("/", "-")))
                self.assertTrue((extracted / ("zenith-server" + dev.exe_suffix(target))).is_file())
                if os.name != "nt":
                    self.assertTrue(os.access(extracted / ("zenith-server" + dev.exe_suffix(target)), os.X_OK))
        release.combine(self.dist)
        entries = release.read_manifest(self.dist / "SHA256SUMS")
        self.assertEqual(len(entries), 18)
        release.verify_files(self.dist, entries)

    def test_bundles_are_deterministic_for_identical_inputs(self):
        for target in ("linux/amd64", "windows/arm64"):
            archive = self.pack(target)
            digest = release.sha256(archive)
            self.pack(target)
            self.assertEqual(release.sha256(archive), digest)

    def test_tampered_asset_fails(self):
        self.pack("linux/amd64")
        (self.dist / release.asset_names("linux/amd64")[0]).write_bytes(b"tampered")
        with self.assertRaises(ValueError):
            release.verify_target(self.dist, "linux/amd64")

    def test_checksum_duplicate_traversal_and_wrong_set_rejected(self):
        file = self.root / "SHA256SUMS"
        for content in ("a" * 64 + "  ../outside\n", "a" * 64 + "  ..\n",
                        ("a" * 64 + "  same\n") * 2, "bad checksum\n", ""):
            file.write_text(content)
            with self.assertRaises(ValueError):
                release.read_manifest(file)
        file.write_text("a" * 64 + "  unexpected\n")
        with self.assertRaises(ValueError):
            release.read_manifest(file, {"expected"})

    def test_wrong_binary_architecture_rejected(self):
        for target in dev.TARGETS:
            other = target.split("/")[0] + ("/arm64" if target.endswith("amd64") else "/amd64")
            release.validate_binary(fixture_binary(target), target)
            with self.assertRaises(ValueError):
                release.validate_binary(fixture_binary(other), target)
            with self.assertRaises(ValueError):
                release.validate_binary(b"truncated", target)

    def test_existing_extraction_destination_rejected(self):
        self.pack("darwin/amd64")
        with self.assertRaises(ValueError):
            release.unpack(self.dist, "darwin/amd64", self.root)


class APITests(unittest.TestCase):
    def test_unicode_names_and_base_path(self):
        name = "worker \u03a9 &/+?="
        req = api.build_request("http://127.0.0.1:8080/prefix/", "status", name)
        parsed = urllib.parse.urlsplit(req.full_url)
        self.assertEqual(parsed.path, "/prefix/status")
        self.assertEqual(urllib.parse.parse_qs(parsed.query)["service"], [name])
        req = api.build_request("https://example.test", "add", name)
        self.assertEqual(req.method, "POST")
        self.assertEqual(json.loads(req.data), {"service_name": name})

    def test_methods_and_validation(self):
        self.assertEqual(api.build_request("http://localhost", "remove", "api").method, "DELETE")
        self.assertEqual(api.build_request("http://localhost", "ping", None).method, "GET")
        for url in ("file:///tmp", "http://user:pass@host", "http://host?x=1", "http://host#x"):
            with self.assertRaises(ValueError):
                api.build_request(url, "status", None)
        for action, name in (("add", None), ("remove", None), ("ping", "extra")):
            with self.assertRaises(ValueError):
                api.build_request("http://localhost", action, name)


if __name__ == "__main__":
    unittest.main()
