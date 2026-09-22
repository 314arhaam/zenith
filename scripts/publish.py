#!/usr/bin/env python3
"""Publish only a complete verified release. Used by the gated tag workflow.

Requires the GitHub CLI plus GH_TOKEN/GH_REPO. Never called by local verification.
"""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import re
import subprocess

from dev import TARGETS
from release import asset_names, read_manifest, verify_files, verify_target


def publish(directory: Path, tag: str) -> None:
    if not re.fullmatch(r"v[0-9][0-9A-Za-z._+-]*", tag):
        raise ValueError("ZENITH_RELEASE_TAG must start with v and a digit, e.g. v1.2.3")
    if not os.environ.get("GH_REPO"):
        raise ValueError("GH_REPO is required")
    expected = {name for target in TARGETS for name in asset_names(target)}
    entries = read_manifest(directory / "SHA256SUMS", expected)
    verify_files(directory, entries)
    for target in TARGETS:
        verify_target(directory, target)
    assets = [str(directory / name) for name in sorted(entries)] + [str(directory / "SHA256SUMS")]
    result = subprocess.run(["gh", "release", "view", tag, "--json", "isDraft"],
                            capture_output=True, text=True, encoding="utf-8")
    if result.returncode == 0:
        if not json.loads(result.stdout)["isDraft"]:
            raise ValueError("Release is already published; refusing to replace public assets")
        subprocess.run(["gh", "release", "upload", tag, *assets, "--clobber"], check=True)
    else:
        # --verify-tag prevents creating a new tag after a failed release lookup.
        subprocess.run(["gh", "release", "create", tag, *assets, "--verify-tag", "--draft",
                        "--generate-notes", "--title", tag], check=True)
    subprocess.run(["gh", "release", "edit", tag, "--draft=false"], check=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, required=True)
    args = parser.parse_args()
    publish(args.directory, os.environ.get("ZENITH_RELEASE_TAG", ""))
