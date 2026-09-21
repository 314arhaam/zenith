#!/usr/bin/env python3
"""Portable Zenith developer commands. Python 3.10+, standard library only.

No shell is used for subprocesses. Paths are relative to the repository, not
whichever directory invoked the script. Go dependencies still need downloading.
"""
from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
TARGETS = tuple(f"{system}/{arch}" for system in ("linux", "windows", "darwin")
                for arch in ("amd64", "arm64"))
RACE_TARGETS = frozenset(TARGETS) - {"windows/arm64"}


def output(text: str) -> None:
    # Windows legacy consoles cannot necessarily encode a Unicode checkout path.
    try:
        print(text, flush=True)
    except UnicodeEncodeError:
        print(text.encode(sys.stdout.encoding or "ascii", "backslashreplace").decode(
            sys.stdout.encoding or "ascii"), flush=True)


def run(argv, *, env=None, log: Path | None = None) -> None:
    argv = [str(arg) for arg in argv]
    output("+ " + subprocess.list2cmdline(argv))
    if log is None:
        subprocess.run(argv, cwd=ROOT, env=env, check=True)
        return
    log.parent.mkdir(parents=True, exist_ok=True)
    with log.open("w", encoding="utf-8", newline="\n") as saved:
        process = subprocess.Popen(argv, cwd=ROOT, env=env, stdout=subprocess.PIPE,
                                   stderr=subprocess.STDOUT, text=True,
                                   encoding="utf-8", errors="replace")
        try:
            assert process.stdout is not None
            for line in process.stdout:
                saved.write(line)
                output(line.rstrip("\r\n"))
            code = process.wait()
        except BaseException:
            process.kill()
            process.wait()
            raise
        finally:
            if process.stdout is not None:
                process.stdout.close()
    if code:
        raise subprocess.CalledProcessError(code, argv)


def go_executable() -> str:
    executable = shutil.which(os.environ.get("GO", "go"))
    if executable is None:
        raise ValueError("Go is not on PATH. Install the version in .go-version; GO may name its executable.")
    return executable


def host_target() -> str:
    result = subprocess.run([go_executable(), "env", "GOHOSTOS", "GOHOSTARCH"],
                            cwd=ROOT, check=True, capture_output=True, text=True,
                            encoding="utf-8")
    fields = result.stdout.split()
    if len(fields) != 2:
        raise ValueError("Could not determine the native Go target")
    return validate_target("/".join(fields))


def validate_target(target: str) -> str:
    if target not in TARGETS:
        raise ValueError(f"Unsupported target {target!r}; choose one of {', '.join(TARGETS)}")
    return target


def exe_suffix(target: str) -> str:
    return ".exe" if validate_target(target).startswith("windows/") else ""


def build_env(target: str, *, race=False, base=None) -> dict[str, str]:
    target = validate_target(target)
    if race and target not in RACE_TARGETS:
        raise ValueError("Go's race detector does not support windows/arm64; run ordinary tests there")
    env = dict(os.environ if base is None else base)
    system, arch = target.split("/")
    env.update(GOOS=system, GOARCH=arch, CGO_ENABLED="1" if race else "0")
    # Never inherit host-specific ISA requirements when generating portable builds.
    env["GOAMD64"] = "v1"
    env["GOARM64"] = "v8.0"
    return env


def race_env(target: str) -> dict[str, str]:
    env = build_env(target, race=True)
    if target == "windows/amd64":
        compiler = env.get("CC") or shutil.which("gcc")
        if compiler is None:
            for candidate in (r"C:\mingw64\bin\gcc.exe", r"C:\msys64\ucrt64\bin\gcc.exe"):
                if Path(candidate).is_file():
                    compiler = candidate
                    break
        if compiler is None:
            raise ValueError("Windows amd64 race tests need a MinGW-w64 C compiler on PATH (runtime v8+)")
        compiler = shutil.which(compiler) or compiler
        result = subprocess.run([compiler, "--print-file-name=libsynchronization.a"],
                                capture_output=True, text=True, check=True)
        library = result.stdout.strip()
        if not library or library == "libsynchronization.a" or not Path(library).is_file():
            raise ValueError("The Windows C compiler lacks a supported MinGW-w64 synchronization runtime")
        # Go parses CC as a quoted command; retain spaces in compiler paths.
        env["CC"] = f'"{compiler}"' if any(char.isspace() for char in compiler) else compiler
        env["PATH"] = str(Path(compiler).resolve().parent) + os.pathsep + env.get("PATH", "")
    return env


def local_path(value: str) -> Path:
    path = Path(value)
    return path if path.is_absolute() else ROOT / path


def build(target: str, directory: Path, *, release=False) -> tuple[Path, Path]:
    directory.mkdir(parents=True, exist_ok=True)
    env = build_env(target)
    products = []
    for app in ("server", "client"):
        name = (f"zenith_{app}-{target.replace('/', '-')}" if release else f"zenith-{app}") + exe_suffix(target)
        path = directory / name
        args = [go_executable(), "build", "-mod=readonly", "-trimpath"]
        if release:
            args += ["-ldflags=-s -w"]
        run([*args, "-o", path, f"./{app}"], env=env)
        products.append(path)
    return products[0], products[1]


def format_sources(check: bool) -> None:
    formatter = str(Path(go_executable()).with_name("gofmt.exe" if os.name == "nt" else "gofmt"))
    if not Path(formatter).is_file():
        formatter = shutil.which("gofmt") or formatter
    files = sorted(path for path in ROOT.rglob("*.go")
                   if not {"vendor", ".git", "dist", "artifacts"}.intersection(path.relative_to(ROOT).parts))
    # Keep command lines comfortably below Windows' process-creation limit.
    changed = []
    for start in range(0, len(files), 20):
        result = subprocess.run([formatter, "-l" if check else "-w", *map(str, files[start:start + 20])],
                                cwd=ROOT, check=True, capture_output=True, text=True, encoding="utf-8")
        changed.extend(result.stdout.splitlines())
    if changed:
        raise ValueError("Run scripts/dev.py fmt to format:\n" + "\n".join(changed))


def python_tests() -> None:
    run([sys.executable, "-m", "unittest", "discover", "-s", "scripts/tests", "-v"], log=ROOT / "artifacts/python-tests.txt")


def test(target: str, *, race=False, packages=None) -> None:
    directory = ROOT / "artifacts" / "tests"
    directory.mkdir(parents=True, exist_ok=True)
    flags = ["-race"] if race else []
    run([go_executable(), "test", "-mod=readonly", *flags, "-shuffle=on", "-count=1", "-timeout=5m",
         f"-coverprofile={directory / 'coverage.out'}", "-json", *(packages or ["./..."])],
        env=race_env(target) if race else build_env(target), log=directory / "tests.jsonl")
    run([go_executable(), "tool", "cover", f"-func={directory / 'coverage.out'}"],
        log=directory / "coverage.txt")


def e2e(target: str, directory: Path, *, skip_build=False, from_bundle=False, log_dir=None) -> None:
    if not skip_build:
        build(target, directory)
    env = dict(os.environ, PYTHONUTF8="1")
    command = [sys.executable, ROOT / "scripts/e2e.py",
               "--server", directory / ("zenith-server" + exe_suffix(target)),
               "--client", directory / ("zenith-client" + exe_suffix(target)),
               "--log-dir", local_path(log_dir or "artifacts/e2e")]
    if from_bundle:
        command.append("--require-executable")
    run(command, env=env)


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    p = commands.add_parser("build", help="build both native executables")
    p.add_argument("--output", default="dist")
    p = commands.add_parser("test", help="Go unit/integration tests; optional native race detector")
    p.add_argument("--race", action="store_true")
    p.add_argument("packages", nargs="*", default=["./..."])
    commands.add_parser("vet")
    p = commands.add_parser("fmt")
    p.add_argument("--check", action="store_true")
    commands.add_parser("py-test", help="test the portable developer and packaging scripts")
    commands.add_parser("modules", help="download, verify, and check tidy without changing files")
    p = commands.add_parser("e2e", help="build and exercise real server/CLI processes")
    p.add_argument("--output", default="dist")
    p.add_argument("--skip-build", action="store_true")
    p.add_argument("--log-dir", default="artifacts/e2e")
    p = commands.add_parser("verify", help="modules, format, vet, Python/Go tests, build and full E2E")
    p.add_argument("--race", action="store_true")
    p = commands.add_parser("dist", help="build raw binaries plus documented zip/tar.gz bundles")
    p.add_argument("--target", choices=TARGETS, action="append")
    p.add_argument("--all", action="store_true")
    p.add_argument("--output", default="dist/release")
    p = commands.add_parser("doctor", help="verify native architecture and optional race prerequisites")
    p.add_argument("--expect-target", choices=TARGETS)
    p.add_argument("--race", action="store_true")
    commands.add_parser("bench")
    args = parser.parse_args(argv)
    if args.command == "fmt":
        format_sources(args.check)
        return 0
    if args.command == "py-test":
        python_tests()
        return 0
    target = host_target()
    if args.command == "doctor":
        run([go_executable(), "version"])
        output(f"Python {sys.version.split()[0]}; native Go target {target}")
        if args.expect_target and args.expect_target != target:
            raise ValueError(f"Expected native {args.expect_target}, got {target}; install a native Go toolchain")
        if args.race:
            race_env(target)
    elif args.command == "build":
        build(target, local_path(args.output))
    elif args.command == "test":
        test(target, race=args.race, packages=args.packages)
    elif args.command == "vet":
        run([go_executable(), "vet", "-mod=readonly", "./..."], env=build_env(target))
    elif args.command in ("modules", "verify"):
        env = build_env(target)
        for tail in (["mod", "download"], ["mod", "verify"], ["mod", "tidy", "-diff"]):
            run([go_executable(), *tail], env=env)
        if args.command == "verify":
            format_sources(True)
            run([go_executable(), "vet", "-mod=readonly", "./..."], env=env)
            python_tests()
            test(target, race=args.race)
            e2e(target, ROOT / "dist")
    elif args.command == "e2e":
        e2e(target, local_path(args.output), skip_build=args.skip_build, log_dir=args.log_dir)
    elif args.command == "dist":
        from release import package_target, verify_target, combine
        if args.all and args.target:
            parser.error("Choose --all or --target, not both")
        targets = TARGETS if args.all else tuple(args.target or [target])
        directory = local_path(args.output)
        for item in dict.fromkeys(targets):
            server, client = build(item, directory, release=True)
            package_target(directory, item, server, client, ROOT)
            verify_target(directory, item)
        if set(targets) == set(TARGETS):
            combine(directory)
    elif args.command == "bench":
        run([go_executable(), "test", "./core", "./server/handlers", "-run=^$",
             "-bench=Benchmark(SystemMarshal|StatusAll)$", "-benchmem", "-count=5"],
            env=build_env(target), log=ROOT / "artifacts/benchmarks.txt")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (ValueError, OSError, subprocess.CalledProcessError) as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
    except KeyboardInterrupt:
        raise SystemExit(130)
