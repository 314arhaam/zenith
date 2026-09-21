package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func call(t *testing.T, h http.HandlerFunc, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestAddDuplicateAndMalformedJSON(t *testing.T) {
	h := NewHandler()
	rec := call(t, h.Add, http.MethodPost, "/add", `{"service_name":"api"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first add status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = call(t, h.Add, http.MethodPost, "/add", `{"service_name":"api"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate add status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = call(t, h.Add, http.MethodPost, "/add", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed add status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLifecycle(t *testing.T) {
	h := NewHandler()
	if rec := call(t, h.Add, http.MethodPost, "/add", `{"service_name":"api"}`); rec.Code != http.StatusCreated {
		t.Fatalf("add status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := call(t, h.Status, http.MethodGet, "/status?service=api", ""); rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := call(t, h.Remove, http.MethodDelete, "/remove", `{"service_name":"api"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("remove=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec := call(t, h.Status, http.MethodGet, "/status?service=api", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", rec.Code, rec.Body.String())
	}
}
