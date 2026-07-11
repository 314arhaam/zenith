package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	data "zenith/models"
	hd "zenith/server/handlers"
)

func TestMethodNotAllowedRemove(t *testing.T) {
	h := hd.NewHandler()
	serviceName := "test_service-01"
	url := "/remove"
	// mock request and writer
	h.Core.Add(serviceName)
	h.Core.Add(serviceName)
	h.Core.Remove(serviceName)
	w, r := responseAndRequestBuild(http.MethodGet, url, nil)
	h.Remove(w, r)
	if w.Result().StatusCode != 405 {
		t.Fatalf("Error: `Method Not Allowed` doesn't work.")
	}
}

func TestRemove404(t *testing.T) {
	h := hd.NewHandler()
	// mock data
	serviceNameNotFound := "test_service-02"
	url := "/remove"
	// mock request and writer
	payload, err := json.Marshal(data.RemoveRequest{ServiceName: serviceNameNotFound})
	if err != nil {
		t.Fatal("Error in request payload marshall")
	}
	w, r := responseAndRequestBuild(http.MethodDelete, url, strings.NewReader(string(payload)))
	// handle function
	h.Remove(w, r)
	if w.Result().StatusCode != 404 {
		t.Fatalf("StatusCode not 404: %d", w.Result().StatusCode)
	}
	t.Logf("StatusCode: %d", w.Result().StatusCode)
}

func TestRemove(t *testing.T) {
	h := hd.NewHandler()
	// mock data
	serviceName := "test_service-01"
	url := "/remove"
	// mock request and writer
	h.Core.Add(serviceName)
	payload, err := json.Marshal(data.RemoveRequest{ServiceName: serviceName})
	if err != nil {
		t.Fatal("Error in request payload marshall")
	}
	w, r := responseAndRequestBuild(http.MethodDelete, url, strings.NewReader(string(payload)))
	_d_, err := json.Marshal(h.Core.GetAll())
	t.Logf("ServiceData: %v", string(_d_))
	// handle function
	h.Remove(w, r)
	if w.Result().StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode not %d", w.Result().StatusCode)
	} else {
		t.Logf("StatusCode: %d", w.Result().StatusCode)
	}
	_d, err := json.Marshal(h.Core.GetAll())
	if err != nil {
		t.Fatal("Error in marshal")
	}
	t.Logf("%v", string(_d))
	t.Logf("Data removed: `%s` = `%v`\n`%v`", serviceName, string(_d), *w)
}
