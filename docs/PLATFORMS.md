# Zenith platform and automation guide

## Target matrix

The server and CLI are native Go console applications, not Windows services or desktop applications. Release builds disable cgo and use baseline instruction sets (`GOAMD64=v1`, `GOARM64=v8.0`). No Go/Python installation or separately installed compiler runtime is needed to run Zenith's built executables; normal OS libraries are still used.

| Target | Typical machine | Native GitHub runner | Go race tests |
| --- | --- | --- | --- |
| `windows/amd64` | Windows x64 / Intel or AMD | `windows-2025` | Yes, with MinGW-w64 |
| `windows/arm64` | Windows on ARM64 | `windows-11-arm` | Not supported by Go; ordinary full tests run |
| `linux/amd64` | Linux x64 / Intel or AMD | `ubuntu-24.04` | Yes |
| `linux/arm64` | Linux ARM64 | `ubuntu-24.04-arm` | Yes |
| `darwin/amd64` | Intel Mac | `macos-15-intel` | Yes |
| `darwin/arm64` | Apple Silicon Mac | `macos-15` | Yes |

These are the native CI operating-system versions, not a promise about every historic Windows/Linux/macOS release. The configured Go 1.27 release toolchain requires macOS 13 or later. Check the Go port requirements before deploying to older systems. macOS builds are ordinary Go binaries; this project does not perform Developer ID signing or notarization. Follow your organization's software-trust policy rather than disabling OS protections.

## Tooling prerequisites

Use Go from `.go-version` for builds and releases. CI also tests the previous supported Go line on Linux. The `go 1.23.0` declaration in `go.mod` remains the language/module floor; it is not a recommendation to deploy an unsupported old toolchain.

The developer helpers use Python 3.10+ and only the standard library. `.python-version` selects 3.13 for CI. Windows examples use the Python launcher (`py -3`); an installation exposing `python` works too. Linux/macOS examples use `python3`. Bash, curl and GNU Make are not required for the Windows development workflow.

Native race tests require cgo and suitable platform tooling. On Windows x64, install a MinGW-w64 compiler with runtime version 8+; `scripts/dev.py doctor --race` checks `libsynchronization.a`. The hosted Windows image includes GCC, which is checked rather than silently assuming success. For a custom installation, set `CC` to the compiler executable and put its dependent DLL directory on PATH. On macOS, install Xcode command-line tools. On Linux, install the distribution's C build tools. Windows ARM64 must use ordinary tests, not `--race`.

## One command interface on every OS

Run from the repository root, replacing `python3` with `py -3` in Windows PowerShell:

```sh
python3 scripts/dev.py doctor
python3 scripts/dev.py modules
python3 scripts/dev.py fmt --check
python3 scripts/dev.py vet
python3 scripts/dev.py py-test
python3 scripts/dev.py test
python3 scripts/dev.py test --race
python3 scripts/dev.py e2e
python3 scripts/dev.py verify
python3 scripts/dev.py dist --all
```

`verify` runs module integrity/tidy checks, formatting, vet, helper tests, Go tests, a native build and full process E2E. Add `--race` on a supported target. Every failed subprocess propagates a nonzero exit; no shell pipeline hides failures. Use `fmt` without `--check` to apply Go formatting.

`build` and `e2e` use `dist/`; `build --output "dist/custom path"` selects another location. `e2e` builds both executables unless `--skip-build` is supplied. The developer tool resolves relative paths against the repository, preserves argument boundaries, and resets inherited cross-build targets for native commands. `GO` can name a different Go executable.

Optional wrappers are `scripts/dev.ps1` and `scripts/dev.sh`. On Linux/macOS run `sh scripts/dev.sh build`; in PowerShell, `.\scripts\dev.ps1 build` works when your execution policy allows local scripts. There is no need to relax an organization policy: invoke `py -3 scripts/dev.py build` directly instead. The Makefile only delegates to these portable Python commands.

The old `scripts/add.sh`, `remove.sh`, `status.sh`, and `status_single.sh` are optional Unix wrappers. The same HTTP examples are available on every OS through `scripts/api.py`, without curl or platform-specific JSON escaping:

```sh
python3 scripts/api.py --url http://127.0.0.1:8080 add "worker east"
python3 scripts/api.py --url http://127.0.0.1:8080 status "worker east"
python3 scripts/api.py --url http://127.0.0.1:8080 remove "worker east"
```

Use the Go CLI for monitoring (`check`), retrying pings, help and shell completion. The Python HTTP helper is only an example utility, not a second implementation of those CLI features.

## Release artifacts and extraction

`python3 scripts/dev.py dist --all` creates `dist/release/`. To build one target from any supported host, use `--target windows/arm64`, for example. This cross-compiles; it does not execute the foreign binaries.

Each target produces:

```text
zenith_server-OS-ARCH[.exe]       # existing raw release filename retained
zenith_client-OS-ARCH[.exe]
zenith-OS-ARCH.zip               # Windows bundle
zenith-OS-ARCH.tar.gz            # Linux/macOS bundle (instead of .zip)
checksums-OS-ARCH.txt
```

There are **12 raw binaries and six bundles** in a full release. `SHA256SUMS` contains all 18 hashes. Per-target checksum files are CI artifacts; the release publishes the combined checksum file. Packaging verifies PE, ELF or Mach-O headers against the requested architecture and verifies bundled bytes against raw assets. The checksums detect mismatches; they are not a substitute for publisher authenticity/signing.

Every bundle contains a single `zenith-OS-ARCH/` directory with `zenith-server[.exe]`, `zenith-client[.exe]`, README and these two documentation files. Unix archives preserve executable mode 0755 and document mode 0644. Binary contents are not text-normalized; documentation is normalized to UTF-8/LF.

Windows PowerShell, after downloading the matching release bundle:

```powershell
Get-FileHash .\zenith-windows-amd64.zip -Algorithm SHA256
# Compare the value with the matching line in SHA256SUMS from the same release.
Expand-Archive .\zenith-windows-amd64.zip -DestinationPath .\zenith-release
Set-Location .\zenith-release\zenith-windows-amd64
$env:ZENITH_HOST = "127.0.0.1"
.\zenith-server.exe 8080
```

Linux:

```sh
sha256sum zenith-linux-amd64.tar.gz
# Compare with the matching SHA256SUMS line before extraction.
tar -xzf zenith-linux-amd64.tar.gz
cd zenith-linux-amd64
ZENITH_HOST=127.0.0.1 ./zenith-server 8080
```

macOS (Apple Silicon; use `amd64` for Intel):

```sh
shasum -a 256 zenith-darwin-arm64.tar.gz
# Compare with the matching SHA256SUMS line before extraction.
tar -xzf zenith-darwin-arm64.tar.gz
cd zenith-darwin-arm64
ZENITH_HOST=127.0.0.1 ./zenith-server 8080
```

Use a second terminal in the extracted directory to run the matching client. Prefer the tar.gz bundles on Unix: downloading raw artifacts may lose executable bits. For a raw binary that you have verified and are permitted to run, restore its executable permission with `chmod +x FILENAME`.

## E2E and shutdown semantics

`scripts/e2e.py` launches only its own loopback server and CLI processes. It copies executables to a temporary path containing spaces and Unicode, with a working directory outside the repository. It checks API lifecycle/validation, concurrent registrations, CLI behavior, completion generation, polling, deadlines and graceful shutdown. Server logs and a machine-readable `results.json` are saved even when a test fails.

Linux/macOS tests send SIGINT to the CLI and SIGTERM to the server. Windows tests allocate a console when necessary, start each long-lived child in a separate process group, and target that child's group with CTRL_BREAK_EVENT. They never broadcast to group 0 or kill processes by name. The server must exit with code 0 through its graceful path; a cancelled CLI must report cancellation and exit 1. Forceful cleanup does not count as a passing graceful-shutdown test.

User-facing Ctrl+C works on all three platforms. Windows `Stop-Process`, `taskkill /F`, Task Manager termination and POSIX SIGKILL cannot be used to prove graceful shutdown. Windows console-close/logoff/OS-shutdown events may impose an earlier OS deadline than Zenith's 10-second drain period. This repository does not install or manage Windows SCM services, launchd agents or systemd units.

`--server-only` is an explicit partial diagnostic mode. Hosted E2E never uses it. Helper packaging tests use synthetic header fixtures; they do not prove that application binaries compile or execute. Cross-builds likewise do not replace native E2E.

## GitHub Actions

| File | Triggers and checks |
| --- | --- |
| `ci.yml` | Push/PR, manual and reusable. Native unit/integration/helper tests on all six targets; race tests on five. Previous supported Go on Linux, vet/format/module checks, coverage, actionlint and bounded fuzzing. |
| `build.yml` | Manual or reusable, automatically called by E2E. Cross-builds both applications for six targets, creates bundles, checks architecture, permissions and hashes. |
| `e2e.yml` | Push/PR, manual and reusable. Calls Build once, downloads each target's exact artifact and executes the unpacked server/CLI on that native runner. |
| `release.yml` | `v*` tags. Requires CI plus packaged-binary E2E; aggregates and verifies artifacts, stages a draft release, then publishes. Published releases are never overwritten. |
| `benchmarks.yml` | Manual. Runs native benchmark samples for all six targets. Results across different hardware are not directly comparable throughput claims. |
| `send_telegram.yml` | Optional post-publication notifier on a Linux management runner; unrelated to the user's OS. Disabled unless `ENABLE_TELEGRAM=true`. |

Release binaries are built once per E2E/release workflow run and tested as packaged, not rebuilt after validation. Workflow dependencies fail closed: failed native tests prevent publication. Keep runner availability/permissions enabled in repository settings. New commits or release tags can consume hosted-runner minutes; no workflows are launched by extracting this source ZIP or running local tests.

Actions use full commit pins and Dependabot remains configured. Go/Python version files are the central build settings. `.gitattributes` keeps Go, Python, Markdown and shell files LF across checkouts; PowerShell wrappers use CRLF. After adding attributes to an existing checkout, review the result of `git add --renormalize .` before committing.

The workflows do not provision Tailscale, use production infrastructure, or require network secrets for tests. For optional Telegram delivery, set the existing bot/chat secrets only when you intend to send release files to that chat. Code in `scripts/publish.py` is run only by the gated release job, never by `verify` or `dist`.

## Reference basis

Application behavior is defined by the accompanying source. External platform guidance used for this update (checked 2026-09-20):

- [Go release history](https://go.dev/doc/devel/release) and [Go 1.27 port requirements](https://go.dev/doc/go1.27).
- [Go race-detector requirements](https://go.dev/doc/articles/race_detector#Requirements).
- [Go Windows console signal behavior](https://pkg.go.dev/os/signal#hdr-Windows).
- [Python subprocess process groups and signals](https://docs.python.org/3/library/subprocess.html).
- [Microsoft GenerateConsoleCtrlEvent](https://learn.microsoft.com/en-us/windows/console/generateconsolectrlevent).
- [GitHub standard runner labels](https://docs.github.com/en/actions/reference/runners/github-hosted-runners).
