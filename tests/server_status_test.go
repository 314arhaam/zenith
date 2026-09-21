package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"zenith/core"
	hd "zenith/server/handlers"
)

func TestStatusAll(t *testing.T) {
	h := hd.NewHandler()
	h.Core.Add("api")
	w, r := responseAndRequestBuild(http.MethodGet, "/status", nil)
	h.Status(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
	}
	var got map[string]core.Service
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := got["api"]; !ok {
		t.Fatalf("api missing from response: %+v", got)
	}
}

func TestStatusSingle(t *testing.T) {
	h := hd.NewHandler()
	h.Core.Add("api")
	w, r := responseAndRequestBuild(http.MethodGet, "/status?service=api", nil)
	h.Status(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestStatusMissingReturnsJSONNotFound(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodGet, "/status?service=missing", nil)
	h.Status(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusNotFound)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q; want application/json", got)
	}
}
