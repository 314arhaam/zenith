package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	data "zenith/models"
)

func TestStatusEndpoint(t *testing.T) {
	endpoint := "/add"
	serviceName := "test_service"
	var request string
	var err error
	if request, err = data.ToJson(&data.AddRequest{ServiceName: serviceName}); err != nil {
		t.Fatalf("Cannot convert request struct to json: %s", err)
	}
	t.Log("\n[*] Request body OK")
	server := NewTestServer()
	defer server.Close()
	resp, err := http.Post(
		server.URL+endpoint,
		"application/json",
		strings.NewReader(request),
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin POST method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("\n[x] Errorin response /add: StatusCode %d", resp.StatusCode)
	}
	//
	endpoint = "/status"
	resp, err = http.Get(
		server.URL + endpoint,
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin GET method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("\n[x] Errorin response /status: StatusCode %d", resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
	//
	endpoint = "/status?service=" + serviceName
	resp, err = http.Get(
		server.URL + endpoint,
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin GET method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("\n[x] Errorin response /status?service=%s: StatusCode %d", serviceName, resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
	//
	endpoint = "/remove"
	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete,
		server.URL+endpoint,
		strings.NewReader(request),
	)
	if err != nil {
		t.Fatalf("\n[x] Error in DELETE request make: %v", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("\n[x] Error in DELETE Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("\n[x] Errorin response /status: StatusCode %d", resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
	//
	endpoint = "/status"
	resp, err = http.Get(
		server.URL + endpoint,
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin GET method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("\n[x] Errorin response /status: StatusCode %d", resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
	//
	endpoint = "/status?service=" + serviceName
	resp, err = http.Get(
		server.URL + endpoint,
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin GET method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("\n[x] Errorin response /status?service=%s: StatusCode %d", serviceName, resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
}
