import requests, json

API_URL = "http://0.0.0.0:8080"

def test_add_405() -> None:
    r = requests.post(
        f"{API_URL}/add",
        headers = {"Content-Type": "application/json"},
        data = {
            "name": "test",
        }
    )
    if r.ok:
        raise ValueError(f"/add test failed: must be 405: {r.status_code}")

def test_add(payload: dict) -> None:
    r = requests.post(
        f"{API_URL}/add",
        headers = {"Content-Type": "application/json"},
        json = payload
    )
    if not r.ok:
        raise ValueError(f"/add test failed: must be 201: {r.status_code}")
    else:
        print(r.status_code)

def test_remove(payload: dict) -> None:
    r = requests.post(
    f"{API_URL}/remove",
    headers = {"Content-Type": "application/json"},
    json = payload
)
    if not r.ok:
        raise ValueError(f"/remove test failed: must be 201: {r.status_code}")
    else:
        print(r.status_code)

def test_status():
    r = requests.get(
        f"{API_URL}/status"
    )
    if not r.ok:
        raise ValueError("/status test failed")
    else:
        print(json.dumps(r.json(), indent = 4))

if __name__ == '__main__':
    payloads = [
        {"service_name": "test-00"},
        {"service_name": "test-01"},
        {"service_name": "test-02"},
        {"service_name": "test-03"},
        {"service_name": "test-04"},
        {"service_name": "test-05"},
    ]
    for i in range(3):
        test_add(payload=payloads[i])
    test_status()
    test_add(payload=payloads[3])
    test_add(payload=payloads[4])
    test_status()
    test_remove(payload=payloads[2])
    test_status()
    for p in payloads:
        test_remove(payload=p)
    test_status()