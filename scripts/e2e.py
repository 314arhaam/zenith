#!/usr/bin/env python3
"""Process-level tests for the actual Zenith executables; stdlib only.

All traffic is loopback. The harness terminates only the processes it starts.
--server-only explicitly skips CLI coverage; CI never supplies that flag.
"""
from __future__ import annotations

import argparse
from concurrent.futures import ThreadPoolExecutor
from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import platform
import shutil
import sys
import tempfile
import socket
import subprocess
import threading
import time
import urllib.error
import urllib.parse
import urllib.request

from processes import console_for_children, force_cleanup, graceful_stop, interrupt, start


def require(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def request(base: str, method: str, path: str, payload=None, raw=None, content_type="application/json"):
    body = json.dumps(payload).encode("utf-8") if payload is not None else raw
    headers = {"Content-Type": content_type} if body is not None else {}
    req = urllib.request.Request(base + path, data=body, method=method, headers=headers)
    # Ignore machine-specific proxies for this local test server.
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    try:
        response = opener.open(req, timeout=3)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        return response.status, dict(response.headers), response.read()


def expect(base, method, path, status, **kwargs):
    actual, headers, body = request(base, method, path, **kwargs)
    require(actual == status, f"{method} {path}: wanted {status}, got {actual}: {body!r}")
    return headers, body


def wait_for_log(process, path, text, timeout=5):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if text in path.read_text(encoding="utf-8"):
            return
        if process.poll() is not None:
            raise AssertionError(f"monitor exited early: {path.read_text(encoding='utf-8')}")
        time.sleep(0.025)
    raise AssertionError(f"monitor never produced {text!r}")


@contextmanager
def monitor(client, base, log, *arguments):
    with log.open("w", encoding="utf-8") as output:
        process = start(
            [str(client), "--url", base, "check", *arguments],
            stdout=output, stderr=subprocess.STDOUT, cwd=client.parent,
        )
        try:
            yield process
        finally:
            if process.poll() is None:
                force_cleanup(process)


def run_suite(args, cases) -> None:
    server_path = args.server
    def passed(name):
        cases.append(name)
        print(f"PASS {name}", flush=True)

    # The program validates 1..65535. Reserve a random free port rather than
    # depending on a fixed shared port. Readiness also verifies process liveness.
    port = free_port()
    base = f"http://127.0.0.1:{port}"
    env = dict(os.environ, ZENITH_HOST="127.0.0.1")
    with (args.log_dir / "server.log").open("w", encoding="utf-8") as server_log:
        server = start([str(server_path), str(port)], env=env, cwd=server_path.parent, stdout=server_log, stderr=subprocess.STDOUT)
        try:
            deadline = time.monotonic() + 15
            while True:
                require(server.poll() is None, "server exited during startup; inspect server.log")
                try:
                    _, body = expect(base, "GET", "/ping", 200)
                    require(body == b"Pong\n", "unexpected ping body")
                    break
                except (OSError, urllib.error.URLError):
                    require(time.monotonic() < deadline, "server readiness timed out")
                    time.sleep(0.05)
            passed("server readiness")
            headers, body = expect(base, "HEAD", "/status", 200)
            require(body == b"" and headers.get("X-Zenith-Service-Count") == "0", "bad empty HEAD response")
            _, body = expect(base, "GET", "/status", 200)
            require(json.loads(body) == {}, "registry not empty")
            passed("empty registry and HEAD")

            name = "api & worker/+?=" + "\u062e\u062f\u0645\u062a"
            query = "/status?" + urllib.parse.urlencode({"service": name})
            _, body = expect(base, "POST", "/add", 201, payload={"service_name": name})
            service = json.loads(body)
            require(service["service_id"] > 0 and service["create_datetime"], "invalid service metadata")
            expect(base, "POST", "/add", 409, payload={"service_name": name})
            for path in [query, "/status", "/status"]:
                _, body = expect(base, "GET", path, 200)
                require(json.loads(body)[name] == service, "duplicate registration or cache changed service")
            _, body = expect(base, "DELETE", "/remove", 204, payload={"service_name": name})
            require(body == b"", "204 response had a body")
            expect(base, "GET", query, 404)
            headers, body = expect(base, "HEAD", query, 404)
            require(headers.get("X-Zenith-Service-Count") == "0" and not body, "bad missing HEAD")
            _, body = expect(base, "GET", "/status", 200)
            require(json.loads(body) == {}, "cached deleted service")
            passed("lifecycle, special-character names and cache invalidation")

            for method, path in [("POST", "/add"), ("DELETE", "/remove")]:
                for raw, status, code in [
                    (b"{", 400, "invalid_json"),
                    (b'{"service_name":"api","unknown":1}', 400, "invalid_json"),
                    (b'{"service_name":"api"} {}', 400, "invalid_json"),
                    (b'{"service_name":"   "}', 400, "invalid_payload"),
                    (json.dumps({"service_name": "x" * 8192}).encode(), 413, "payload_too_large"),
                ]:
                    _, body = expect(base, method, path, status, raw=raw)
                    require(json.loads(body)["error"] == code, f"wrong structured error: {body!r}")
                expect(base, method, path, 415, raw=b'{}', content_type="text/plain")
            for path in ["/status?service=", "/status?service=a&service=b", "/status?service=%ZZ"]:
                expect(base, "GET", path, 400)
            passed("malformed JSON, payload limits and query validation")

            for method, path, allow in [("GET", "/add", "POST"), ("POST", "/status", "GET, HEAD"), ("POST", "/remove", "DELETE")]:
                headers, body = expect(base, method, path, 405)
                require(headers.get("Allow") == allow and json.loads(body)["error"] == "method_not_allowed", "bad 405 response")
            expect(base, "GET", "/unknown", 404)
            passed("routing and allowed methods")

            with ThreadPoolExecutor(max_workers=16) as pool:
                codes = list(pool.map(lambda _: request(base, "POST", "/add", {"service_name": "same"})[0], range(32)))
            require(codes.count(201) == 1 and codes.count(409) == 31, f"non-atomic add: {codes}")
            expect(base, "DELETE", "/remove", 204, payload={"service_name": "same"})
            passed("concurrent duplicate registration: one winner")

            names = [f"worker-{i}" for i in range(64)]
            def add(worker):
                _, body = expect(base, "POST", "/add", 201, payload={"service_name": worker})
                return json.loads(body)["service_id"]
            with ThreadPoolExecutor(max_workers=16) as pool:
                ids = list(pool.map(add, names))
            require(len(set(ids)) == len(names), "duplicate service IDs")
            _, body = expect(base, "GET", "/status", 200)
            require(set(json.loads(body)) == set(names), "lost concurrent registrations")
            with ThreadPoolExecutor(max_workers=16) as pool:
                list(pool.map(lambda worker: expect(base, "DELETE", "/remove", 204, payload={"service_name": worker}), names))
            _, body = expect(base, "GET", "/status", 200)
            require(json.loads(body) == {}, "concurrent removal left data behind")
            passed("64 concurrent service lifecycles and unique IDs")

            if not args.server_only:
                client = args.client
                def cli(*arguments, success=True, timeout=5):
                    result = subprocess.run([str(client), "--url", base, *arguments], capture_output=True, text=True, encoding="utf-8", timeout=timeout, cwd=client.parent)
                    require((result.returncode == 0) == success, f"CLI {arguments}: {result.returncode}\n{result.stdout}\n{result.stderr}")
                    return result
                require(cli("ping").stdout == "Pong\n", "CLI ping changed")
                require(json.loads(cli("add", name).stdout)["service_id"] > 0, "bad CLI registration")
                require(name in json.loads(cli("status", name).stdout), "CLI URL encoding failed")
                cli("add", name, success=False)
                cli("remove", name)
                cli("status", name, success=False)
                require(json.loads(cli("status").stdout) == {}, "CLI list not empty")
                cli("ping", "--until", "-1", success=False)
                cli("check", "--sleep", "0", success=False)
                result = cli("check", "missing", "--quiet", "--max-retry", "1", success=False)
                require(result.stdout == "", "quiet mode wrote normal output")
                passed("CLI lifecycle, error exits, URL escaping and flag validation")

                expect(base, "POST", "/add", 201, payload={"service_name": "monitor"})
                log = args.log_dir / "monitor.log"
                with monitor(client, base, log, "monitor", "--sleep", "1", "--max-retry", "1") as process:
                    wait_for_log(process, log, "Step 0")
                    expect(base, "DELETE", "/remove", 204, payload={"service_name": "monitor"})
                    require(process.wait(timeout=5) != 0, "check failed to notice removal")
                passed("CLI monitor detects removed registration")

                expect(base, "POST", "/add", 201, payload={"service_name": "cancel"})
                log = args.log_dir / "cancel.log"
                with monitor(client, base, log, "cancel", "--sleep", "3600") as process:
                    wait_for_log(process, log, "Step 0", timeout=10)
                    interrupt(process)
                    require(process.wait(timeout=5) == 1, "cancellation did not return CLI error exit 1")
                    require("context canceled" in log.read_text(encoding="utf-8"), "CLI did not handle the console signal")
                expect(base, "DELETE", "/remove", 204, payload={"service_name": "cancel"})
                passed("CLI console signal interrupts an hour-long polling wait")

                for shell in ("bash", "zsh", "fish", "powershell"):
                    require("zenith-client" in cli("completion", shell, timeout=10).stdout,
                            f"completion generation failed for {shell}")
                require("zenith-client" in cli("--help").stdout, "help uses the wrong executable name")
                passed("CLI help and four shell completion generators")

                finished = threading.Event()
                class SlowPing(BaseHTTPRequestHandler):
                    def log_message(self, *_args):
                        pass
                    def do_GET(self):
                        finished.wait(10)
                        try:
                            self.send_response(200); self.end_headers(); self.wfile.write(b"Pong\n")
                        except (BrokenPipeError, ConnectionResetError):
                            pass
                slow = ThreadingHTTPServer(("127.0.0.1", 0), SlowPing)
                thread = threading.Thread(target=slow.serve_forever, daemon=True)
                thread.start()
                try:
                    start_time = time.monotonic()
                    result = subprocess.run([str(client), "--url", f"http://127.0.0.1:{slow.server_port}", "--timeout", "10s", "ping", "--until", "1"], capture_output=True, timeout=8)
                    require(result.returncode != 0 and time.monotonic() - start_time < 5, "--until did not bound the HTTP request")
                finally:
                    finished.set(); slow.shutdown(); slow.server_close(); thread.join(timeout=2)
                passed("CLI overall ping deadline")

            code = graceful_stop(server)
            require(code == 0, f"server graceful exit code {code}; expected zero")
            require("Shutdown complete" in (args.log_dir / "server.log").read_text(encoding="utf-8"),
                    "server did not finish its graceful-shutdown path")
            passed("server graceful console/SIGTERM shutdown")
        finally:
            if server.poll() is None:
                force_cleanup(server)

    for arguments in (["0"], ["65536"], ["-1"], ["abc"], ["8080", "extra"]):
        result = subprocess.run([str(server_path), *arguments], capture_output=True, timeout=10)
        require(result.returncode == 1, f"invalid startup arguments did not fail: {arguments}")
    passed("server rejects invalid ports and excess arguments")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--server", type=Path, required=True)
    parser.add_argument("--client", type=Path)
    parser.add_argument("--server-only", action="store_true")
    parser.add_argument("--require-executable", action="store_true", help="check unpacked executable bits on POSIX")
    parser.add_argument("--log-dir", type=Path, default=Path("artifacts/e2e"))
    args = parser.parse_args()
    args.server = args.server.resolve()
    if not args.server_only and args.client is None:
        parser.error("--client is required unless --server-only is explicit")
    if args.client is not None:
        args.client = args.client.resolve()
    args.log_dir = args.log_dir.resolve()
    args.log_dir.mkdir(parents=True, exist_ok=True)
    cases, error = [], None
    try:
        sources = [args.server] + ([args.client] if not args.server_only else [])
        for source in sources:
            require(source.is_file(), f"binary missing: {source}")
            if args.require_executable and os.name != "nt":
                require(os.access(source, os.X_OK), f"bundle lost executable permissions: {source}")
        # Exercise paths with spaces and Unicode on every OS, away from the repo.
        with console_for_children(), tempfile.TemporaryDirectory(prefix="zenith e2e ") as temporary:
            work = Path(temporary) / "bin \u03a9"
            work.mkdir()
            args.server = Path(shutil.copy2(args.server, work / args.server.name))
            if not args.server_only:
                args.client = Path(shutil.copy2(args.client, work / args.client.name))
            run_suite(args, cases)
    except BaseException as exc:
        error = f"{type(exc).__name__}: {exc}"
        raise
    finally:
        summary = {
            "mode": "server-only" if args.server_only else "server-and-cli",
            "status": "failed" if error else "passed",
            "platform": sys.platform, "machine": platform.machine(),
            "passed_groups": len(cases), "cases": cases,
            "skipped_groups": ["all CLI tests (--server-only)"] if args.server_only else [],
            "error": error,
        }
        (args.log_dir / "results.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
        print(json.dumps(summary, indent=2), flush=True)


if __name__ == "__main__":
    main()
