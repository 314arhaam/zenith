# Zenith

Zenith is a small Go HTTP service registry with a Cobra command-line client. It tracks named registrations for applications, developers and CI runners on a trusted network.

```text
zenith-client -> HTTP API -> in-memory registry
```

**Windows, Linux and macOS; amd64 and arm64.** Both applications are native console executables. A registration contains a name, process-local numeric ID and creation timestamp. Registering/removing a service does not start/stop its application. Zenith tracks membership, not application health.

## Documentation

- [Quick reference](docs/QUICK_REFERENCE.md): introduction, setup on all three OSes, endpoints, request/response shapes, CLI commands and flags.
- [Platform and automation guide](docs/PLATFORMS.md): build/test tools, release archives, runner matrix, signals, prerequisites and workflows.
- [Cross-platform validation report](review/CROSS_PLATFORM_REVIEW.md): what was actually checked locally and what still requires hosted/native validation.
- [Previous optimization report](review/OPTIMIZATION_REVIEW.md): historical cache benchmarks and prior validation scope, not a result for the current revision.

## Build and start

Install Go as specified by `.go-version` and download module dependencies. Python 3.10+ is only needed for the developer helpers/E2E, not to run release binaries.

Windows PowerShell:

```powershell
go mod download
py -3 scripts/dev.py build
$env:ZENITH_HOST = "127.0.0.1"
.\dist\zenith-server.exe 8080
# In a second terminal:
.\dist\zenith-client.exe add api
.\dist\zenith-client.exe status api
.\dist\zenith-client.exe remove api
```

Linux/macOS:

```sh
go mod download
python3 scripts/dev.py build
ZENITH_HOST=127.0.0.1 ./dist/zenith-server 8080
# In a second terminal:
./dist/zenith-client add api
./dist/zenith-client status api
./dist/zenith-client remove api
```

No Python? Run `go build -trimpath -o zenith-server ./server` and `go build -trimpath -o zenith-client ./client`, adding `.exe` to output filenames on Windows. Direct builds write to the current directory.

The port is optional (default 8080). `ZENITH_HOST` is optional; unset means all interfaces. For remote use, bind a reachable private interface and pass `--url http://SERVER:8080` to the client. Apply access controls before exposing the listener.

## HTTP API and CLI

| API | CLI equivalent | Purpose |
| --- | --- | --- |
| `POST /add` | `add SERVICE` | Register a name; duplicates fail with 409. |
| `DELETE /remove` | `remove SERVICE` | Remove a registration. |
| `GET /status[?service=NAME]` | `status [SERVICE]` | Read one/all registrations. |
| `GET /ping` | `ping [--until SECONDS]` | Check Zenith, optionally retrying. |
| Repeated `GET /status` | `check [SERVICE]` | Monitor registration until absence/error/cancellation. |
| `HEAD /status[?service=NAME]` | `check [SERVICE] --quiet` | Body-free membership polling. |

Add/remove request: `{"service_name":"api"}`. JSON error schema: `{"error":"CODE","message":"DETAIL"}`. Full status returns objects keyed by service name. `HEAD /ping` is also supported. All formats, flags, limits and status codes are in the quick reference.

The CLI's global defaults are `--url http://127.0.0.1:8080` and `--timeout 10s`. Use `help [COMMAND]`, `COMMAND --help`, or `completion bash|zsh|fish|powershell`. Shell completion is printed, not installed automatically.

## Portable development

```sh
# On Windows replace python3 with py -3.
python3 scripts/dev.py verify
python3 scripts/dev.py test --race    # supported platforms/C compiler required
python3 scripts/dev.py dist --all
```

`verify` performs module checks, formatting, vet, helper tests, Go tests, builds and real server/CLI E2E. Windows arm64 supports the ordinary tests but not Go's race detector. See the platform guide for C-compiler requirements. The optional Makefile and shell/PowerShell wrappers delegate to the same Python entrypoint.

Release builds create 12 raw binaries, six Windows ZIP/Unix tar.gz bundles, and checksums. Existing raw artifact names are retained; extracted bundles use `zenith-server` and `zenith-client` (with `.exe` on Windows). Packaging checks the binary OS/architecture and executable permissions.

## Workflows

CI runs native unit/integration tests on all six targets. E2E calls the reusable build workflow, then tests the exact packaged server and CLI on each target's native runner. Releases on `v*` tags are gated on both CI and E2E; no separate rebuild happens after those checks. Manual multi-platform benchmarks and an opt-in Telegram notifier remain available. Platform/workflow configuration is not a claim that those hosted jobs have already executed.

Tests use loopback and do not require Tailscale or production secrets. The notifier only runs after publication when `ENABLE_TELEGRAM=true`, using `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID_CHANNEL`; enabling it sends release files to that chat. Local `verify` and `dist` never publish or notify.

## Operating boundaries

The registry is thread-safe and full-status JSON is cached between mutations. Snapshots are independent and network writes do not hold the registry lock. Names are case-sensitive, trimmed and limited to 256 UTF-8 bytes; JSON requests are limited to 4 KiB. The CLI caps responses at 8 MiB, does not follow redirects, and keeps HTTPS certificate verification enabled.

Ctrl+C cancels the CLI or starts server draining on all supported OSes. Windows Ctrl+Break and POSIX SIGTERM are also handled; force-killing a process is not graceful shutdown. The server allows up to 10 seconds to drain active requests. See platform-specific OS termination limits in the platform guide.

There is no persistence, replication, automatic TTL/heartbeat expiry, authentication, built-in TLS, pagination or registration quota. A crash does not remove a registration; a Zenith restart removes all of them. Do not expose this server directly to untrusted clients. Windows service/launchd/systemd installation is not implemented.

## License

The original README identified MIT, but the supplied archive contained an empty `client/LICENSE` and no complete root license. The maintainer must confirm and supply the intended license/copyright text before relying on that designation for redistribution. This update does not invent those details.
