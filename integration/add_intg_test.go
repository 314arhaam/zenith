package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	handlefuncs "zenith/api"
)

func TestAddIntegration(t *testing.T) {
	h := handlefuncs.NewHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("/add", h.Add)
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Post(
		server.URL+"/add",
		"application/json",
		strings.NewReader(`{"service_name":"test"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("got %d", resp.StatusCode)
	}
}
