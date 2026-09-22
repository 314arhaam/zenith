# Zenith optimization review

Review date: 2026-09-20.

## Scope and conclusion

This review used the actual supplied **zenith-fixed.zip**, not the earlier original archive or only the previous chat description. The baseline ZIP SHA-256 is:

```text
9c2f72a3bd25dbfc5e7b664520cfa845247d71746c4569ef5097863f6b33fff9
```

The earlier atomic registration/removal, locked snapshots, structured errors and basic timeouts were useful fixes and were preserved. The main remaining optimization opportunity was repeated full-registry encoding during polling, not the individual map lookup. Operational and test reliability also needed work in command state, cancellation, input bounds and workflow consistency.

The revised implementation is still a single-process in-memory registry. It is better suited to read-heavy, trusted-network coordination, but it is not a production service-discovery platform or a measure of registered applications' health.

## Findings and implemented changes

| Area | Finding in supplied updated code | Change |
| --- | --- | --- |
| All-service polling | Copies and JSON-encodes the whole registry on every request. | Versioned immutable encoding cache; invalidated only by successful changes. |
| Lock ownership | Constructor/handler values contain mutex-bearing state. | Pointer constructors and pointer ownership; zero-value registry supported. |
| CLI state | Global Cobra command and flag state reused between invocations. | Fresh command tree/options/output for each invocation. |
| API client | Repeated command-level request handling, unbounded body reads. | Shared typed client, context propagation, bounded reads, typed HTTP errors. Existing connection reuse was retained, not newly invented. |
| Polling | Uninterruptible sleeps; retry time limit did not bound the active HTTP request. | Context-aware timers, request cancellation and an overall ping deadline. |
| Membership checks | Broad 404 handling and insufficient validation of a named 200 response. | Structured missing-service code, requested-key validation; optional HEAD count protocol. |
| Flags | Zero/negative durations or retry counts could yield invalid/hot loops. | Positive timing/retry validation, overflow checks, explicit failure semantics. |
| Input validation | Unknown fields, unlimited body/name length, inconsistent name/query handling. | 4 KiB bodies, 256-byte normalized names, strict known fields, single JSON value and query validation. |
| Routing | Handler and production router semantics could diverge. | One shared router; consistent method errors, Allow headers and HEAD support. |
| Server lifecycle | No graceful termination or configurable bind host. | Signal-driven draining with timeout/force-close fallback; ZENITH_HOST. |
| Module metadata | Cobra classified indirect despite direct imports. | Corrected go.mod and formatted it; remote tidy validation remains pending. |
| Workflows | Old toolchains/actions and release paths, insufficient publication gates. | Full-SHA actions, current supported Go CI, native E2E, six builds, checksums, gated draft-to-public release. |
| Documentation | Described a Tailscale multi-runner E2E setup not present in the supplied updated workflow; wrong default client address. | Documents the actual local-per-runner test model and 127.0.0.1 client default. |

Telegram delivery was retained as an explicitly enabled reusable workflow after successful publication. It is not enabled or executed by this review. No service has been deployed and no GitHub workflow has been run here.

## Cache correctness and tradeoffs

`System.Marshal()` first reads the cached immutable string under an RLock. A cache miss takes a separate fill mutex, snapshots the map and its generation under RLock, releases the registry lock, and encodes outside it. The encoding is published only when the generation still matches. A concurrent reader may receive a coherent snapshot from before an overlapping write; a stale encoding cannot be installed for a newer generation. Calls starting after a completed mutation will not reuse the invalidated old cache.

Add/Set/Remove invalidation and concurrent readers/writers are exercised by race tests. Failed duplicate adds, missing removes and unchanged Set calls preserve the cache. `GetAll()` still returns an independent map. Network writes do not hold the registry lock, including for slow clients.

This exchanges additional retained memory for lower repeated-read CPU/allocation cost. A warm cache keeps one full JSON representation alive. Mutation-heavy workloads still incur cloning, key sorting and encoding; the new cache-fill lock serializes misses. No multi-reader/writer throughput claim is made from the sequential microbenchmarks. Sharding or a different map implementation was not added without workload evidence.

All-service GET still transfers the entire JSON payload; it is not O(1) end-to-end. `check --quiet` uses a separate O(1) metadata lookup/count and HEAD response, so it avoids that payload when only membership matters. It does not perform application health checks.

## Measured performance

Same benchmark source files were applied to the supplied updated baseline and optimized tree. The baseline tree received only the two benchmark harnesses, not production changes.

Environment: Linux amd64, local Go 1.23.2, GOMAXPROCS=4. Five samples per case, 200 ms target per sample, medians below. Cache warmed before timed read-only cases; no mutations during those cases. The handler benchmark includes constructing an httptest response recorder and writing the complete JSON, not network transport, TLS or application work. These are local microbenchmarks, not production load-test results or confidence intervals.

```bash
GOMAXPROCS=4 go test ./core ./server/handlers -run '^$' \
  -bench 'Benchmark(SystemMarshal|StatusAll)$' -benchmem -benchtime=200ms -count=5
```

| Services | Baseline us/op | Optimized us/op | Baseline / optimized | Allocated B/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: |
| 10 | 3.809 | 0.853 | 4.47x | 3,337 -> 1,776 | 33 -> 11 |
| 1,000 | 397.596 | 18.355 | 21.66x | 262,725 -> 82,928 | 2,014 -> 11 |
| 10,000 | 6,182.862 | 188.972 | 32.72x | 3,138,206 -> 771,056 | 20,025 -> 11 |

At 1,000 services, the median warm full-status handler was about **21.7x faster**, with approximately **68.4% fewer allocated bytes per request**. Allocated bytes per operation are not retained heap size.

The counterexample matters: with a mutation before every Marshal at 1,000 services, baseline median was **397.255 us/op** and optimized median **405.297 us/op** (about 2% slower). Treat that as roughly unchanged in this environment, not an improvement. Cold/write-heavy callers should not expect the warm-cache speedup.

The raw core-only warm Marshal results in the JSON/logs measure returning a cached string reference and exclude output transfer; their very large ratios must not be advertised as whole-application speedups. Re-run benchmarks under Go 1.27 and a representative read/write mix before capacity planning.

Raw evidence: [baseline](evidence/bench-baseline.txt), [optimized](evidence/bench-optimized.txt), [computed medians](evidence/benchmark-summary.json). The baseline and optimized code use the same service names and fixed timestamps in these harnesses.

## Validation actually completed

The local runtime is Go 1.23.2. These commands succeeded:

```bash
go vet ./core ./models ./server/... ./client/api ./tests/integration
go test -race -shuffle=on -count=1 -timeout=3m \
  -coverprofile=coverage-local.out \
  ./core ./models ./server/... ./client/api ./tests/integration
```

| Package | Locally measured statement coverage |
| --- | ---: |
| core | 98.5% |
| models | 96.6% |
| server/handlers | 95.3% |
| client/api | 87.8% |
| server | 57.5% |

This table is partial-package coverage, not whole-repository coverage. The server entrypoint is separately exercised by the process test, which was not built with coverage instrumentation. The 100% output for tests/integration in the log covers its tiny test helper, not 100% of application paths.

The bounded parser fuzz run completed **194,347 executions** without a reported failure. This is a smoke test, not proof that the parser is bug-free.

A built real server passed **eight server-only E2E groups**: readiness, empty GET/HEAD state, lifecycle and cache invalidation, malformed/oversized inputs, routing, atomic duplicate registration, 64 concurrent distinct lifecycles/IDs, and graceful SIGTERM exit. Loopback networking was used; no external server was touched.

All legacy tests in tests/*.go other than the Cobra-dependent client_add_test.go passed in an explicit named-file run; see [that limited-scope result](evidence/legacy-tests-partial.log).

The server was also cross-compiled for all six Linux/Windows/macOS amd64/arm64 targets recorded in [server-cross-builds.log](evidence/server-cross-builds.log). Cross-compilation is not native execution on those operating systems. These local Go 1.23 binaries are test outputs, not release artifacts, and are not included in the source archive.

GitHub workflow YAML, local reusable references, needs names, full action commit pins, bash syntax and Python script syntax were checked locally. No local actionlint result is claimed. The generated CI includes a pinned actionlint run.

Evidence: [local test/vet/coverage log](evidence/local-validation.log), [coverage data](evidence/coverage-local.out), [fuzz log](evidence/fuzz.log), [server-only E2E](evidence/e2e-server-only.log), [workflow syntax checks](evidence/workflow-validation.log).

## Validation blocked or not performed

The sandbox cannot resolve/download `github.com/spf13/cobra v1.10.2` from proxy.golang.org. The original baseline full-suite attempt failed on that network access. An explicit offline final attempt also failed because the module was not cached. See [baseline failure](evidence/baseline-full-suite-blocked.log) and [final offline failure](evidence/full-suite-blocked.log).

Therefore **no passing result is claimed** for the full `go test ./...`, full-repository vet, Cobra command tests, full CLI build, full server-plus-CLI process E2E, full module tidy verification, actionlint, Windows/macOS native runtime tests, or actual GitHub Actions/release execution. The Cobra code was inspected statically; no stub or replacement library was used to manufacture a passing test result.

New CLI tests and the full E2E runner are supplied for execution in a network-enabled environment. The workflows gate publication on those checks; until they run, this is a tested server/API revision with partially unverified CLI/CI integration, not an unconditional production-readiness statement.

## Compatibility notes

Endpoint names and existing successful response shapes are preserved. The following are deliberate changes to review before upgrade:

- NewSystem/NewHandler and Handler.Core now use pointers. Explicit Go value-typed callers may need edits; inferred constructor assignments continue to work.
- Unknown JSON fields, >4 KiB requests, >256-byte normalized names, malformed/repeated/empty service queries and invalid explicit media types are rejected.
- The client limits received bodies to 8 MiB, validates service response structures, and does not follow redirects. Set the final URL directly; use named status or quiet probes for oversized inventories.
- Polling/retry timing flags must be positive, except ping --until=0 still means one attempt. Failures now return consistently; cancelled commands exit nonzero.
- `check --quiet` depends on the updated server's HEAD count header. Normal GET polling remains available for the previous fixed server.
- ZENITH_HOST is additive; all-interface bind is still the default for compatibility. Deployment configuration must restrict access.
- Telegram notification is opt-in; configure the repository variable and existing secrets only when intended.

## Remaining operating limits / maintainer decisions

No TTL/heartbeat expiry, persistence, replication, authentication, TLS termination, registration quota or pagination was added. These are product/deployment decisions beyond optimizing the supplied registry. A large or hostile client population can still exhaust resources; the request-size limit is not a total-memory or concurrency limit. Use only within a suitably restricted deployment.

Service identifiers remain process-local uint64 values and may exceed JavaScript Number precision. A switch to durable UUIDs or string-encoded IDs would require an explicit API migration and was not silently made.

The archive's MIT statement is not accompanied by a complete license text: client/LICENSE is empty and no root LICENSE is present. The maintainer must supply the intended license and copyright holder; they were not invented.

## Reproduce full validation after dependencies are available

```bash
go mod download
go mod verify
go mod tidy -diff
go vet ./...
go test -race -shuffle=on -count=1 -timeout=3m ./...
go build -trimpath -o zenith-server ./server
go build -trimpath -o zenith-client ./client
python3 scripts/e2e.py --server ./zenith-server --client ./zenith-client
```

Use the supported toolchain in .go-version for these deployment checks. On Windows give the binaries .exe names and run the Python script with those paths; race tests need a suitable C compiler.

## External references (separate from source-derived findings)

Go release history was verified on 2026-09-20: [Go release policy and versions](https://go.dev/doc/devel/release). Go 1.27.1 was released 2026-09-01; the supported major lines are 1.26 and 1.27. This motivated the CI/build update rather than deployment on the locally available Go 1.23.2.

HTTP lifecycle/API implementation reference: [net/http](https://pkg.go.dev/net/http). Release staging reference: [GitHub CLI gh release create](https://cli.github.com/manual/gh_release_create). Action version pins were checked against their upstream release commits; workflow execution remains unverified locally as stated above.
