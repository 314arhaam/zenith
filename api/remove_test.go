package handlefuncs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethodNotAllowedRemove(t *testing.T) {
	h := NewHandler()
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

func TestRemove(t *testing.T) {
	h := NewHandler()
	// mock data
	serviceName := "test_service-01"
	url := "/remove?service=" + serviceName
	// mock request and writer
	h.Core.Add(serviceName)
	r := httptest.NewRequest(
		http.MethodPost,
		url,
		nil,
	)
	h.Core.Add(serviceName)
	_d_, err := json.MarshalIndent(h.Core, "", " ")
	t.Logf("ServiceData: %v", string(_d_))
	w := httptest.NewRecorder()
	// handle function
	h.Remove(w, r)
	if w.Result().StatusCode != 201 {
		t.Errorf("StatusCode not %d", w.Result().StatusCode)
	} else {
		t.Logf("StatusCode: %d", w.Result().StatusCode)
	}
	_d, err := json.MarshalIndent(h.Core, "", " ")
	if err != nil {
		t.Fatal("Error in marshal")
	}
	t.Logf("%v", string(_d))
	t.Logf("Data removed: `%s` = `%v`\n`%v`", serviceName, string(_d), *w)
}
