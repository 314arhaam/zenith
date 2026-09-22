package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	data "zenith/models"
	hd "zenith/server/handlers"
)

func TestAddCreated(t *testing.T) {
	h := hd.NewHandler()
	payload, _ := json.Marshal(data.AddRequest{ServiceName: "api"})
	w, r := responseAndRequestBuild(http.MethodPost, "/add", strings.NewReader(string(payload)))
	h.Add(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d; want %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestAddDuplicateReturnsConflict(t *testing.T) {
	h := hd.NewHandler()
	h.Core.Add("api")
	w, r := responseAndRequestBuild(http.MethodPost, "/add", strings.NewReader(`{"service_name":"api"}`))
	h.Add(w, r)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d; want %d; body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestAddMalformedJSONReturnsBadRequest(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodPost, "/add", strings.NewReader(`{"service_name":`))
	h.Add(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAddEmptyServiceReturnsBadRequest(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodPost, "/add", strings.NewReader(`{"service_name":"   "}`))
	h.Add(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAddMethodNotAllowed(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(http.MethodGet, "/add", nil)
	h.Add(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
