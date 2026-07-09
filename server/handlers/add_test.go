package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zenith/core"
	data "zenith/models"
)

func responseAndRequestBuild(method string, url string, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
	r := httptest.NewRequest(
		method,
		url,
		body,
	)
	w := httptest.NewRecorder()
	return w, r
}

func TestMethodNotAllowedAdd(t *testing.T) {
	// intentionally requets GET method on /add
	h := NewHandler()
	serviceName := "test_service-01"
	url := "/add"
	// mock request and writer
	h.Core.Add(serviceName)
	w, r := responseAndRequestBuild(http.MethodGet, url, nil)
	h.Add(w, r)
	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Error: `TestMethodNotAllowedAdd` failed. StatusCode %d", w.Result().StatusCode)
	}
}

func TestEmptyFetch(t *testing.T) {
	// invalid payload is pased to /add
	h := NewHandler()
	invalidPayload := strings.NewReader(`{"key": "val"}`)
	url := "/add"
	// mock request and writer
	w, r := responseAndRequestBuild(http.MethodPost, url, invalidPayload)
	h.Add(w, r)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("Error: `TestEmptyFetch` failed. StatusCode %d", w.Result().StatusCode)
	}
}

func TestAdd(t *testing.T) {
	h := NewHandler()
	// mock data
	serviceName := "test_service-01"
	url := "/add"
	// mock request and writer
	h.Core.Add(serviceName)
	payload, err := json.Marshal(data.AddRequest{ServiceName: serviceName})
	if err != nil {
		t.Fatal("Error in request payload marshal")
	}
	payloadString := string(payload)
	w, r := responseAndRequestBuild(http.MethodPost, url, strings.NewReader(payloadString))
	// handle function
	h.Add(w, r)
	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("Error: `` failed. StatusCode %d", w.Result().StatusCode)
	}
	defer w.Result().Body.Close()
	if d, err := io.ReadAll(w.Result().Body); err != nil {
		t.Fatalf("Error: `` failed. Cannot fetch data. Error: %v", err)
	} else {
		var fetchData core.Service
		if err := json.Unmarshal([]byte(strings.ReplaceAll(string(d), "\n", "")), &fetchData); err != nil {
			t.Fatalf("Error in fetch data validation: %s", err)
		}
		if fetchData.CreateDateTime == "" {
			t.Fatal("Error in system core: Service data generated is empty")
		}
		t.Logf("Result: %v %s", fetchData, string(d))
		t.Log("Success")
	}
}
