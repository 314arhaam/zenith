#!/usr/bin/env python3
"""Small portable HTTP example client. For continuous monitoring use zenith-client."""
from __future__ import annotations

import argparse
import json
import math
import sys
import urllib.error
import urllib.parse
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def build_request(base: str, action: str, name: str | None) -> urllib.request.Request:
    parsed = urllib.parse.urlsplit(base)
    if (parsed.scheme not in ("http", "https") or not parsed.netloc or
            parsed.username is not None or "?" in base or "#" in base):
        raise ValueError("--url must be an HTTP(S) base URL without credentials, query or fragment")
    path = {"ping": "/ping", "status": "/status", "add": "/add", "remove": "/remove"}[action]
    method, data, headers = "GET", None, {}
    if action in ("add", "remove"):
        if not name:
            raise ValueError(f"{action} requires SERVICE")
        method = "POST" if action == "add" else "DELETE"
        data = json.dumps({"service_name": name}, ensure_ascii=False).encode("utf-8")
        headers["Content-Type"] = "application/json"
    elif action == "status" and name is not None:
        path += "?" + urllib.parse.urlencode({"service": name})
    elif action == "ping" and name is not None:
        raise ValueError("ping does not accept SERVICE")
    return urllib.request.Request(base.rstrip("/") + path, data=data, headers=headers, method=method)


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    parser.add_argument("--timeout", type=float, default=10)
    parser.add_argument("action", choices=("ping", "add", "remove", "status"))
    parser.add_argument("service", nargs="?")
    args = parser.parse_args(argv)
    if not math.isfinite(args.timeout) or args.timeout <= 0:
        parser.error("--timeout must be a finite positive number")
    try:
        req = build_request(args.url, args.action, args.service)
        opener = urllib.request.build_opener(NoRedirect())  # HTTPS verification stays enabled.
        with opener.open(req, timeout=args.timeout) as response:
            body = response.read((8 << 20) + 1)
            if len(body) > 8 << 20:
                raise ValueError("Response exceeds 8 MiB")
            text = body.decode("utf-8") if body else f"Status Code: {response.status}\n"
            print(text, end="" if text.endswith("\n") else "\n")
        return 0
    except urllib.error.HTTPError as exc:
        with exc:
            detail = exc.read(4096).decode("utf-8", "replace").strip()
        print(f"HTTP {exc.code}: {detail}", file=sys.stderr)
    except (ValueError, OSError, urllib.error.URLError) as exc:
        print(str(exc), file=sys.stderr)
    return 1


if __name__ == "__main__":
    # Keep redirected output UTF-8 even on older Windows PowerShell hosts.
    for stream in (sys.stdout, sys.stderr):
        if hasattr(stream, "reconfigure"):
            stream.reconfigure(encoding="utf-8")
    raise SystemExit(main())
