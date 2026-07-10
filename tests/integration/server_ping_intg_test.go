package integration

import (
	"io"
	"net/http"
	"testing"
)

func TestPingEndpoint(t *testing.T) {
	endpoint := "/ping"
	server := NewTestServer()
	defer server.Close()
	resp, err := http.Get(
		server.URL + endpoint,
	)
	if err != nil {
		t.Fatalf("\n[x] Errorin GET: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("\n[x] Errorin response: StatusCode %d", resp.StatusCode)
	}
	if pong, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Errorin response body io.ReadAll: %v", err)
	} else {
		if string(pong) != "Pong\n" {
			t.Fatalf("Invalid response %s", string(pong))
		} else {
			t.Logf("\n[*] Success: %s", string(pong))
		}
	}
}
