package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	data "zenith/models"
)

func TestAddEndpoint(t *testing.T) {
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
	t.Log("\n[*] Server OK")
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
		t.Fatalf("\n[x] Errorin response: StatusCode %d", resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
}
