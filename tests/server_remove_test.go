package tests

import (
	"net/http"
	"strings"
	"testing"
	hd "zenith/server/handlers"
)

func TestRemoveExisting(t *testing.T) {
	h := hd.NewHandler()
	h.Core.Add("api")
	w, r := responseAndRequestBuild(http.MethodDelete, "/remove", strings.NewReader(`{"service_name":"api"}`))
	h.Remove(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusNoContent)
	}
	if _, ok := h.Core.Get("api"); ok {
		t.Fatal("service still exists after remove")
	}
}

func TestRemoveMissingReturnsNotFound(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodDelete, "/remove", strings.NewReader(`{"service_name":"missing"}`))
	h.Remove(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusNotFound)
	}
}

func TestRemoveMalformedJSONReturnsBadRequest(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodDelete, "/remove", strings.NewReader(`{`))
	h.Remove(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}
