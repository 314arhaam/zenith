# Zenith

**Windows, Linux and macOS - overview, HTTP API and CLI reference**

Zenith is a lightweight, single-process service registry written in Go. A central HTTP server stores named registrations; a Cobra CLI lets applications, developers and CI runners register, inspect, monitor and remove them across a trusted network.

```text
zenith-client -> HTTP API -> in-memory registry
```

Each registration has a name, numeric ID and creation timestamp. Registering or removing a name **does not start or stop an application**. Zenith does not store application addresses or ports, and registration does not prove that an application is healthy.

## 1. Build and run

The server and CLI target **amd64 and arm64 on Windows, Linux and macOS**. `amd64` means x64/Intel/AMD; `darwin` is Go's name for macOS. Choose `darwin/arm64` for Apple Silicon.

To build from source, install the Go version in `.go-version`. Python 3.10+ is needed only for the optional developer tools and E2E tests; **built Zenith executables do not require Go, Python, Bash, Make or curl**. Initial builds need access to the Go module dependencies.

### Windows - PowerShell

From the repository root:

```powershell
go mod download
py -3 scripts/dev.py build
```

Terminal 1:

```powershell
$env:ZENITH_HOST = "127.0.0.1"
.\dist\zenith-server.exe 8080
```

Terminal 2:

```powershell
.\dist\zenith-client.exe ping
.\dist\zenith-client.exe add api
.\dist\zenith-client.exe status api
.\dist\zenith-client.exe remove api
```

When Python is installed as `python` rather than `py`, replace `py -3` with `python`.

### Linux and macOS - Bash or Zsh

From the repository root:

```sh
go mod download
python3 scripts/dev.py build
```

Terminal 1:

```sh
ZENITH_HOST=127.0.0.1 ./dist/zenith-server 8080
```

Terminal 2:

```sh
./dist/zenith-client ping
./dist/zenith-client add api
./dist/zenith-client status api
./dist/zenith-client remove api
```

Without Python, build directly with `go build -trimpath -o zenith-server ./server` and `go build -trimpath -o zenith-client ./client`; add `.exe` to output filenames on Windows. These direct commands write to the repository root instead of `dist/`.

For a remote server, add `--url http://SERVER:8080` to the CLI and configure the server to listen on a reachable, access-controlled interface. Do not use `0.0.0.0` as the remote client address.

## 2. Server configuration and shutdown

| Setting | Behavior |
| --- | --- |
| `zenith-server [PORT]` | Optional positional port; default `8080`, valid range `1-65535`. |
| `ZENITH_HOST` | Bind host/interface. Unset: all interfaces. `127.0.0.1`: local-only. The server does not load `.env` files automatically. |
| Windows shutdown | Console Ctrl+C or Ctrl+Break requests graceful shutdown. `Stop-Process` and Task Manager termination are forceful, not graceful. |
| Linux/macOS shutdown | Ctrl+C (`SIGINT`) or `SIGTERM` requests graceful shutdown. `SIGKILL` cannot be handled. |
| Grace period | Up to 10 seconds for active requests to drain; overdue connections are closed. Windows console-close/logoff/system-shutdown events can terminate the process sooner. |

`ping` checks Zenith itself. `check` checks **registry membership**, not application health. Crashed applications remain registered until removed. Registrations disappear when Zenith restarts. There is no automatic expiry, persistence, replication, built-in authentication or TLS. Restrict network access; do not expose Zenith directly to untrusted clients.

## 3. HTTP API

Default local base URL: `http://127.0.0.1:8080`. Responses use `application/json`, except `/ping`, which returns plain text.

| Method | Endpoint | Response / behavior |
| --- | --- | --- |
| `GET` / `HEAD` | `/ping` | `200`. GET: `Pong` plus newline. HEAD: no body. |
| `POST` | `/add` | `201` with a service object. Existing name: `409`; metadata is not replaced. |
| `DELETE` | `/remove` | `204`, no body. Missing name: `404`. |
| `GET` | `/status` | `200`, registrations keyed by name. Empty registry: `{}`. |
| `GET` | `/status?service=NAME` | `200`, one-entry object. Missing name: `404`. |
| `HEAD` | `/status` | `200`; `X-Zenith-Service-Count` gives the total. No body. |
| `HEAD` | `/status?service=NAME` | Present: `200`, count `1`; absent: `404`, count `0`. No body. |

### Requests and responses

`POST /add` and `DELETE /remove` use the same JSON body:

```json
{"service_name":"api"}
```

Illustrative `POST /add` response:

```json
{"service_id":123,"create_datetime":"2026-09-20T08:00:00Z"}
```

Illustrative `GET /status?service=api` response; all-service GET uses the same shape:

```json
{"api":{"service_id":123,"create_datetime":"2026-09-20T08:00:00Z"}}
```

`service_id` is an unsigned 64-bit process-local number, not an authentication token or durable global ID. Preserve integer precision when consuming it. `create_datetime` is a UTC RFC 3339 timestamp, with fractional seconds when present.

### Validation and errors

Send `Content-Type: application/json`. An absent header is accepted; an explicitly incompatible type returns `415`. Add/remove bodies are limited to **4 KiB** and one JSON object; unknown fields and extra JSON values are rejected. Standard Go JSON duplicate-key/case-matching behavior is unchanged.

Names are case-sensitive, trimmed at both ends, and must contain **1-256 UTF-8 bytes**, with no remaining control characters. Interior spaces and Unicode are allowed. URL-encode names in queries. Omit `service` to list everything; an explicit empty/repeated `service` or malformed query encoding returns `400`.

```json
{"error":"service_not_found","message":"service not found: api"}
```

| HTTP code | Error codes |
| --- | --- |
| `400` | `invalid_json`, `invalid_payload`, `invalid_query` |
| `404` | `service_not_found`, `not_found` |
| `405` | `method_not_allowed`; includes `Allow` header |
| `409` | `service_exists` |
| `413` | `payload_too_large` |
| `415` | `unsupported_media_type` |
| `500` | `encode_error` |

HEAD responses never include an error body.

## 4. CLI commands, subcommands and flags

```text
zenith-client [GLOBAL FLAGS] COMMAND [ARGUMENTS] [FLAGS]
```

On Windows, use `zenith-client.exe`. `SERVICE` is required where shown; `[SERVICE]` is optional. Quote names containing spaces or shell punctuation.

| Command | Meaning and normal output |
| --- | --- |
| `add SERVICE` | Register once; print created-service JSON. Duplicate names fail. |
| `remove SERVICE` | Remove a registration; print `Status Code: 204`. Missing names fail. |
| `status [SERVICE]` | Print named/all registration JSON. Empty registry prints `{}`; a missing named service fails. |
| `ping` | Check Zenith once; print `Pong`. `--until` enables retries. |
| `check [SERVICE]` | Monitor membership continuously. Without a name, any registration counts as available. |
| `help [COMMAND]` / `COMMAND --help` | Show general or command-specific help. |
| `completion SHELL` | Print completion for `bash`, `zsh`, `fish` or `powershell`; does not automatically install it. |

### Global flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `--url` | `http://127.0.0.1:8080` | HTTP(S) base URL; optional base path. No credentials, query or fragment. |
| `--timeout` | `10s` | Positive per-request duration, e.g. `500ms` or `2s`. |
| `--help` / `-h` | `false` | Show help. |

### Command-specific flags

| Command | Flag / alias | Default | Meaning |
| --- | --- | --- | --- |
| `ping` | `--until` / `-u` | `0` | Integer seconds for the overall deadline. `0`: one attempt. Positive: retry, waiting one second between failures. |
| `check` | `--max-retry` / `-r` | `3` | Fail after this many consecutive absent observations; minimum `1`, including the first observation. |
| `check` | `--interval` / `-t` | `5` | Positive integer seconds between absent observations. |
| `check` | `--sleep` / `-s` | `5` | Positive integer seconds between successful polls. |
| `check` | `--quiet` / `-q` | `false` | Use body-free HEAD probes and suppress normal output. Errors still appear. |

A successful `check` resets its absence counter. Connection errors, server errors and invalid responses fail immediately. `check` does not finish after the first success. Ctrl+C cancels requests/waits; failures and cancellations exit with code `1`. On Windows, Ctrl+Break also cancels.

```powershell
# Windows
.\dist\zenith-client.exe ping --until 30
.\dist\zenith-client.exe --timeout 2s check "worker east" -q -s 10 -r 3
```

```sh
# Linux/macOS
./dist/zenith-client ping --until 30
./dist/zenith-client --timeout 2s check "worker east" -q -s 10 -r 3
```

The CLI does not follow redirects and caps responses at **8 MiB**. Use named `status` or `check --quiet` for large registries. `ping --until` bounds active requests as well as retry waits. HTTPS certificate verification stays enabled.

## 5. Development and automation

Use `py -3` on Windows or `python3` on Linux/macOS before `scripts/dev.py`:

| Task | Command suffix |
| --- | --- |
| Build both native applications | `scripts/dev.py build` |
| Unit/integration tests | `scripts/dev.py test` |
| Race tests (supported targets and C compiler required) | `scripts/dev.py test --race` |
| Developer-tool tests | `scripts/dev.py py-test` |
| Full local verification, including real CLI E2E | `scripts/dev.py verify` |
| Six-target builds and release bundles | `scripts/dev.py dist --all` |

GitHub Actions runs unit/integration tests on all six targets, and process E2E against the exact packaged binaries. Windows arm64 runs ordinary tests because Go's race detector does not support it; the other five targets run race tests. Tagged releases are published only after these checks succeed.

Implementation reference: `server/main.go`, `server/handlers/*.go`, `models/*.go`, `core/servicedata.go`, `client/cmd/*.go`, `client/api/client.go`. See [PLATFORMS.md](PLATFORMS.md) for packaging, test prerequisites and workflow details. Native workflow configuration is not evidence that a hosted run has already passed; actual local results are recorded in `review/CROSS_PLATFORM_REVIEW.md` in the source tree.
