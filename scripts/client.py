#!/usr/bin/env python3

import argparse, json, time, requests

API_URL = "http://0.0.0.0:8080"


def add(service_name: str) -> None:
    r = requests.post(
        f"{API_URL}/add",
        headers={"Content-Type": "application/json"},
        json={
            "service_name": service_name,
        },
    )

    if not r.ok:
        raise ValueError(f"/add failed: expected 201, got {r.status_code}")

    print(r.status_code)


def remove(service_name: str) ->None:
    r = requests.post(
        f"{API_URL}/remove",
        headers={"Content-Type": "application/json"},
        json={
            "service_name": service_name,
        },
    )

    if not r.ok:
        raise ValueError(f"/remove failed: expected 201, got {r.status_code}")

    print(r.status_code)


def run(
    retry: int = 1,
    max_retry: int = 3,
    retry_dt: int = 1,
    dt: int = 2,
) -> int:
    counter = 0

    while True:
        res = requests.get(f"{API_URL}/status")

        if res.json() == {}:
            if retry > max_retry:
                print("[*] System shutdown")
                return 0

            print(f"[*] Retry attempt {retry} of {max_retry}")
            retry += 1
            time.sleep(retry_dt * retry)

        else:
            if retry > 1:
                retry = 1

            print(
                f"[*] Count: {counter}\n"
                f"    Data: {json.dumps(res.json(), indent=4)}"
            )

        time.sleep(dt)
        counter += 1


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="client",
        description="Service manager client",
    )

    subparsers = parser.add_subparsers(
        dest="command",
        required=True,
    )

    # add
    add_parser = subparsers.add_parser(
        "add",
        help="Add a service",
    )
    add_parser.add_argument(
        "service_name",
        help="Name of the service",
    )

    # remove
    remove_parser = subparsers.add_parser(
        "remove",
        help="Remove a service",
    )
    remove_parser.add_argument(
        "service_name",
        help="Name of the service",
    )

    # run
    run_parser = subparsers.add_parser(
        "run",
        help="Monitor service status",
    )

    run_parser.add_argument(
        "--retry",
        type=int,
        default=1,
        help="Initial retry count",
    )
    run_parser.add_argument(
        "--max-retry",
        type=int,
        default=3,
        help="Maximum retry attempts",
    )
    run_parser.add_argument(
        "--retry-dt",
        type=int,
        default=1,
        help="Retry delay multiplier (seconds)",
    )
    run_parser.add_argument(
        "--dt",
        type=int,
        default=2,
        help="Polling interval (seconds)",
    )

    return parser


def main():
    parser = build_parser()
    args = parser.parse_args()

    match args.command:
        case "add":
            add(args.service_name)

        case "remove":
            remove(args.service_name)

        case "run":
            run(
                retry=args.retry,
                max_retry=args.max_retry,
                retry_dt=args.retry_dt,
                dt=args.dt,
            )


if __name__ == "__main__":
    main()