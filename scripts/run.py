import requests, time, json, sys

API_URL = "http://0.0.0.0:8080"

if __name__ == "__main__":
    retry = 1
    max_retry = 3
    retry_dt = 1
    dt = 2
    counter = 0
    while True:
        res = requests.get(
            f"{API_URL}/status"
        )
        if res.json() == {}:
            if retry > max_retry:
                print("[*] System shutdown")
                sys.exit(0)
            else:
                print(f"[*] Retry attempt {retry} of {max_retry}")
                retry += 1
                time.sleep(retry_dt * retry)
        else:
            if retry > 1:
                retry = 1
            print(f"[*] Count: {counter}\n  Data {json.dumps(res.json(), indent = 4)}")
        time.sleep(dt)
        counter += 1