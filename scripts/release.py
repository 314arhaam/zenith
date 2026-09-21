#!/usr/bin/env python3
"""Portable release packaging and verification. No tar, zip or sha256sum executable needed."""
from __future__ import annotations

import argparse
import gzip
import hashlib
import io
from pathlib import Path, PurePosixPath
import re
import shutil
import stat
import struct
import tarfile
import zipfile

from dev import ROOT, TARGETS, exe_suffix, validate_target


def sha256(path: Path) -> str:
    with path.open("rb") as handle:
        return hashlib.file_digest(handle, "sha256").hexdigest() if hasattr(hashlib, "file_digest") else _hash(handle)


def _hash(handle) -> str:
    digest = hashlib.sha256()
    for block in iter(lambda: handle.read(1024 * 1024), b""):
        digest.update(block)
    return digest.hexdigest()


def asset_names(target: str) -> tuple[str, str, str]:
    stem = validate_target(target).replace("/", "-")
    suffix = exe_suffix(target)
    return (f"zenith_server-{stem}{suffix}", f"zenith_client-{stem}{suffix}",
            f"zenith-{stem}." + ("zip" if suffix else "tar.gz"))


def validate_binary(data: bytes, target: str) -> None:
    system, arch = validate_target(target).split("/")
    try:
        if system == "linux":
            valid = (data[:6] == b"\x7fELF\x02\x01" and
                     struct.unpack_from("<H", data, 18)[0] == {"amd64": 62, "arm64": 183}[arch])
        elif system == "darwin":
            valid = (data[:4] == b"\xcf\xfa\xed\xfe" and
                     struct.unpack_from("<I", data, 4)[0] == {"amd64": 0x01000007, "arm64": 0x0100000C}[arch])
        else:
            offset = struct.unpack_from("<I", data, 0x3C)[0]
            valid = (data[:2] == b"MZ" and data[offset:offset + 4] == b"PE\x00\x00" and
                     struct.unpack_from("<H", data, offset + 4)[0] == {"amd64": 0x8664, "arm64": 0xAA64}[arch])
    except (struct.error, IndexError):
        valid = False
    if not valid:
        raise ValueError(f"Executable header does not match {target}")


def bundle_members(target: str) -> set[str]:
    prefix = "zenith-" + validate_target(target).replace("/", "-") + "/"
    return {prefix + name for name in ("zenith-server" + exe_suffix(target), "zenith-client" + exe_suffix(target),
                                      "README.md", "docs/QUICK_REFERENCE.md", "docs/PLATFORMS.md")}


def package_target(directory: Path, target: str, server: Path, client: Path, root: Path = ROOT) -> Path:
    names = asset_names(target)
    directory.mkdir(parents=True, exist_ok=True)
    for source, name in zip((server, client), names[:2]):
        if source.resolve() != (directory / name).resolve():
            shutil.copyfile(source, directory / name)
    prefix = "zenith-" + target.replace("/", "-") + "/"
    content = {
        prefix + "zenith-server" + exe_suffix(target): ((directory / names[0]).read_bytes(), 0o755),
        prefix + "zenith-client" + exe_suffix(target): ((directory / names[1]).read_bytes(), 0o755),
    }
    for name in ("README.md", "docs/QUICK_REFERENCE.md", "docs/PLATFORMS.md"):
        # Normalize checkout line endings so Windows-hosted packages use the same docs.
        content[prefix + name] = ((root / name).read_text(encoding="utf-8").encode("utf-8"), 0o644)
    archive = directory / names[2]
    if target.startswith("windows/"):
        with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as bundle:
            for name, (data, mode) in sorted(content.items()):
                entry = zipfile.ZipInfo(name, (1980, 1, 1, 0, 0, 0))
                entry.create_system = 3
                entry.external_attr = (stat.S_IFREG | mode) << 16
                entry.compress_type = zipfile.ZIP_DEFLATED
                bundle.writestr(entry, data)
    else:
        with archive.open("wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", filename="", mtime=0) as zipped:
            with tarfile.open(fileobj=zipped, mode="w", format=tarfile.PAX_FORMAT) as bundle:
                for name, (data, mode) in sorted(content.items()):
                    entry = tarfile.TarInfo(name)
                    entry.size, entry.mode = len(data), mode
                    entry.mtime, entry.uid, entry.gid = 0, 0, 0
                    bundle.addfile(entry, io.BytesIO(data))
    manifest = directory / ("checksums-" + target.replace("/", "-") + ".txt")
    manifest.write_text("".join(f"{sha256(directory / name)}  {name}\n" for name in names), encoding="utf-8", newline="\n")
    return archive


def read_manifest(path: Path, expected: set[str] | None = None) -> dict[str, str]:
    entries = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  ([A-Za-z0-9._-]+)", line)
        if not match:
            raise ValueError(f"Invalid checksum line in {path.name}: {line!r}")
        digest, name = match.groups()
        if name in (".", "..") or name in entries:
            raise ValueError(f"Duplicate/unsafe checksum filename: {name}")
        entries[name] = digest
    if not entries or (expected is not None and set(entries) != expected):
        raise ValueError(f"Missing or unexpected release assets in {path.name}")
    return entries


def verify_files(directory: Path, entries: dict[str, str]) -> None:
    for name, digest in entries.items():
        path = directory / name
        if path.is_symlink() or not path.is_file() or sha256(path) != digest:
            raise ValueError(f"Missing, symbolic-link or checksum-mismatched asset: {name}")


def read_bundle(archive: Path, target: str) -> dict[str, tuple[bytes, int]]:
    content = {}
    if target.startswith("windows/"):
        with zipfile.ZipFile(archive) as bundle:
            for entry in bundle.infolist():
                if entry.filename in content or entry.is_dir():
                    raise ValueError("Unexpected duplicate/directory in release bundle")
                if stat.S_IFMT(entry.external_attr >> 16) != stat.S_IFREG:
                    raise ValueError("Non-regular ZIP member")
                content[entry.filename] = (bundle.read(entry), (entry.external_attr >> 16) & 0o777)
    else:
        with tarfile.open(archive, "r:gz") as bundle:
            for entry in bundle:
                if not entry.isfile() or entry.name in content:
                    raise ValueError("Unexpected non-file/duplicate in release bundle")
                handle = bundle.extractfile(entry)
                assert handle is not None
                with handle:
                    content[entry.name] = (handle.read(), entry.mode)
    if set(content) != bundle_members(target):
        raise ValueError("Missing or unexpected bundle members")
    for name, (data, mode) in content.items():
        if not name.endswith(".md"):
            validate_binary(data, target)
        path = PurePosixPath(name)
        if path.is_absolute() or ".." in path.parts or "\\" in name or not data:
            raise ValueError("Unsafe/empty bundle member")
        expected_mode = 0o644 if name.endswith(".md") else 0o755
        if mode != expected_mode:
            raise ValueError(f"Incorrect permissions for {name}: {mode:o}")
    return content


def verify_target(directory: Path, target: str) -> None:
    names = asset_names(target)
    manifest = directory / ("checksums-" + target.replace("/", "-") + ".txt")
    entries = read_manifest(manifest, set(names))
    verify_files(directory, entries)
    content = read_bundle(directory / names[2], target)
    prefix = "zenith-" + target.replace("/", "-") + "/"
    for app, name in zip(("server", "client"), names[:2]):
        member = prefix + "zenith-" + app + exe_suffix(target)
        if hashlib.sha256(content[member][0]).hexdigest() != entries[name]:
            raise ValueError("Bundled executable differs from its raw release asset")


def combine(directory: Path) -> None:
    merged = {}
    for target in TARGETS:
        verify_target(directory, target)
        manifest = directory / ("checksums-" + target.replace("/", "-") + ".txt")
        merged.update(read_manifest(manifest, set(asset_names(target))))
    (directory / "SHA256SUMS").write_text("".join(f"{digest}  {name}\n" for name, digest in sorted(merged.items())),
                                         encoding="utf-8", newline="\n")


def unpack(directory: Path, target: str, destination: Path) -> Path:
    verify_target(directory, target)
    content = read_bundle(directory / asset_names(target)[2], target)
    # Never follow pre-existing symlinks or overwrite another extraction.
    if destination.exists():
        raise ValueError(f"Extraction destination already exists: {destination}")
    destination.mkdir(parents=True)
    for name, (data, mode) in content.items():
        path = destination.joinpath(*PurePosixPath(name).parts)
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
        path.chmod(mode)
    return destination / ("zenith-" + target.replace("/", "-"))


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("verify", "combine", "unpack"))
    parser.add_argument("--directory", type=Path, default=ROOT / "dist/release")
    parser.add_argument("--target", choices=TARGETS)
    parser.add_argument("--destination", type=Path)
    args = parser.parse_args()
    if args.command == "combine":
        combine(args.directory)
    elif args.command == "verify" and not args.target:
        entries = read_manifest(args.directory / "SHA256SUMS", {name for t in TARGETS for name in asset_names(t)})
        verify_files(args.directory, entries)
        for target in TARGETS:
            read_bundle(args.directory / asset_names(target)[2], target)
    else:
        if not args.target:
            parser.error("--target is required")
        if args.command == "unpack":
            if args.destination is None:
                parser.error("--destination is required")
            print(unpack(args.directory, args.target, args.destination))
        else:
            verify_target(args.directory, args.target)
    print(f"OK: {args.command}")


if __name__ == "__main__":
    main()
