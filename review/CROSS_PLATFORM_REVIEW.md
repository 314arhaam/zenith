# Zenith cross-platform implementation and validation

Date: 2026-09-20.

## Scope

This revision starts from the supplied `zenith-optimized.zip`, SHA-256 `5d71d3da697ee75f3f4756b52a6abb003660c83c91082978a5d4660f1527b7e8`, and updates the current code, automation and Markdown documentation together. It does not revert to the original or first fixed version.

The target set is Windows, Linux and macOS (`darwin`), each on amd64 and arm64. Existing HTTP endpoints, response schemas, registry/cache behavior and application CLI flags remain. CLI help/completion now use the actual executable name, `zenith-client`, instead of `zenith-cli`. The server logs `Shutdown complete` when its graceful path finishes.

## Implemented changes

- `scripts/dev.py`: one standard-library Python entrypoint for native builds, modules, formatting, vet, Go/helper tests, benchmarks, full process E2E, cross-builds and packages. Argument arrays avoid shell quoting differences. Paths resolve against the source directory. Native commands reset inherited GOOS/GOARCH; release targets are explicit.
- `scripts/dev.ps1`, `scripts/dev.sh`, Makefile: optional wrappers around the same commands. Windows does not require Bash, GNU Make, curl or a relaxed PowerShell execution policy; direct Python invocation is supported.
- `scripts/release.py`: raw executable names retained, documented Windows ZIP/Unix tar.gz bundles added, archive metadata normalized, executable bits preserved, architecture headers and checksums checked. Full releases contain 12 raw executables, six bundles and a combined checksum manifest.
- `scripts/processes.py` and `scripts/e2e.py`: owned process groups, POSIX signals and Windows targeted console events. Windows cancellation is no longer skipped. Temporary paths contain spaces and Unicode. Graceful shutdown requires the expected exit and completion log; forced cleanup is not counted as a pass. Machine-readable results and logs are retained on failure.
- `scripts/api.py`: optional portable HTTP examples with UTF-8 request bodies, URL-encoded names, verified HTTPS, bounded reads and explicit failure exits. Unix example scripts delegate to it.
- Runtime and tests: help/completion naming, shutdown completion logging, synchronized rather than 100 ms timer-dependent CLI cancellation test, Unicode/special-character CLI round trips, long-wait cancellation and four shell-completion generators.
- Portable-helper tests: target/environment/path handling, all six archive layouts, binary headers, permissions, checksums, tampering, subprocess failure, HTTP-helper input validation and mocked release-publication safety. No test actually publishes anything.
- `.gitattributes` and `.editorconfig`: consistent source endings across checkouts. Go/Python/Markdown/shell files use LF; PowerShell wrappers use CRLF on Git checkout.
- `README.md`, `docs/QUICK_REFERENCE.md`, `docs/PLATFORMS.md`: Windows PowerShell and Linux/macOS setup, all endpoints/subcommands/options, architecture selection, extraction, development, workflows and operating boundaries.

## Workflow coverage configured

| Target | Native runner | Native Go tests | Native packaged server/CLI E2E |
| --- | --- | --- | --- |
| linux/amd64 | ubuntu-24.04 | Full suite with race detector | Configured |
| linux/arm64 | ubuntu-24.04-arm | Full suite with race detector | Configured |
| windows/amd64 | windows-2025 | Full suite with race detector; MinGW check | Configured |
| windows/arm64 | windows-11-arm | Full ordinary suite; Go race detector unavailable | Configured |
| darwin/amd64 | macos-15-intel | Full suite with race detector | Configured |
| darwin/arm64 | macos-15 | Full suite with race detector | Configured |

CI also tests the previous supported Go line on Linux, performs module/tidy/vet/format checks, helper tests, wrapper smoke tests, actionlint and bounded fuzzing. Build is reusable and manual; push/PR E2E calls it automatically. E2E unpacks and runs each target's exact release bundle. Release requires both CI and E2E, verifies all assets, uploads a draft and publishes only after success. It does not rebuild after testing. The Telegram notifier is opt-in and post-publication only. Benchmarks remain manually triggered across the native matrix.

These are configured checks, not hosted results obtained during this edit. Runner access and available minutes are repository prerequisites. Nothing has been pushed, published, notified or deployed.

## Validation completed in this environment

Environment: Linux amd64, Go 1.23.2, Python 3.13.5. Only this local Go toolchain was available; `.go-version` remains **1.27.1** for CI/releases. Local binaries are diagnostic build outputs and are not included as release executables in the source ZIP.

| Check actually run | Result and scope |
| --- | --- |
| Portable Python unit tests | **25 passed** on Linux. Windows dispatch and release operations use mocks; archive tests use synthetic headers, not executable application substitutes. |
| Go race tests, shuffled and uncached | **Passed** for `core`, `models`, `server`, `server/handlers`, `client/api`, `tests/integration`. Not the full repository suite. |
| Go vet | Passed for those same non-Cobra packages. |
| Real-server process E2E | **9 groups passed** on Linux; explicitly `--server-only`. Includes paths with spaces/Unicode, JSON/name validation, concurrent registration/removal, cache behavior, shutdown and invalid startup arguments. |
| Updated server cross-builds | **All six targets passed**; PE/ELF/Mach-O OS/architecture headers checked. This is compilation, not foreign-platform execution. |
| Windows ARM64 test compilation | All six non-Cobra package test executables compiled. They were not executed. |
| Workflow structural checks | All workflow YAML parsed; declared dependencies, full action pins, local workflow references, six-target matrices, release gates and packaged-artifact paths checked. Not actionlint. |
| Formatting/syntax | Go formatting, Python compilation and POSIX shell syntax checks passed. |

An initial cold cross-build batch hit the command time limit; the isolated Linux ARM64 retry succeeded. No failed or timed-out attempt is counted as a passing build.

Evidence is in [cross-platform-evidence/](cross-platform-evidence/): `python-tests.log`, `go-tests.log`, `vet.log`, `e2e-server-only.log`, `e2e-server-only/results.json`, `cross-build-summary.log`, per-target cross-build logs, `windows-arm64-test-compilation.log`, `workflow-structure.log` and `syntax.log`.

## Not validated here

The complete `go test ./...` attempt is blocked because `github.com/spf13/cobra v1.10.2` is not cached and external dependency downloads are unavailable. `full-suite-attempt.log` records the offline failure. No dependency stub or replacement was used.

Therefore no passing result is claimed for the complete Cobra CLI build or CLI tests, the full server-plus-CLI process E2E, complete-repository vet/tidy, native Windows/macOS execution, real Windows console events, PowerShell wrapper execution, Go 1.26/1.27 execution, actionlint, hosted GitHub workflows, signing/notarization, actual release publication or Telegram delivery. Native CI is configured to execute these application tests, but has not run here. Signing/notarization is not implemented.

This is a source implementation with verified Linux server/API behavior and six server cross-builds, not a guarantee that all hosted jobs already pass.

## Reproduce full validation

From the repository root with the configured Go toolchain, Python 3.10+ and module-download access:

Windows PowerShell:

```powershell
py -3 scripts/dev.py verify
# Windows amd64 with supported MinGW-w64 only:
py -3 scripts/dev.py test --race
py -3 scripts/dev.py dist --all
```

Linux/macOS:

```sh
python3 scripts/dev.py verify --race
python3 scripts/dev.py dist --all
```

`verify` never publishes. `dist` cross-builds/packages but cannot prove foreign native execution. Use the generated native CI/E2E matrix for that. Detailed prerequisites and primary-source platform references are in [docs/PLATFORMS.md](../docs/PLATFORMS.md).

## Compatibility and operating limits

Registration is not application health; adding/removing does not launch or terminate applications. No TTL, persistence, replication, built-in authentication/TLS, pagination or Windows SCM/systemd/launchd installation has been added. OS-forced termination cannot guarantee a graceful drain. Restrict network access. macOS packages are not Developer ID signed/notarized. Existing process-local ID semantics remain.

The supplied archive's MIT claim still has no complete license text. No copyright holder was invented; the maintainer needs to resolve this before redistribution. Historical optimization evidence remains under `review/` and describes its own earlier revision, not this run.
