package handlefuncs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	data "zenith/models"
)

func TestMethodNotAllowedRemove(t *testing.T) {
	serviceName := "test_service-01"
	url := "/remove"
	// mock request and writer
	d := data.CreateServiceData()
	d.Add(serviceName)
	d.Remove(serviceName)
	w, r := responseAndRequestBuild(http.MethodGet, url, nil)
	Remove(w, r, &d)
	if w.Result().StatusCode != 405 {
		t.Fatalf("Error: `Method Not Allowed` doesn't work.")
	}
}

func TestRemove(t *testing.T) {
	// mock data
	serviceName := "test_service-01"
	url := "/remove?service=" + serviceName
	// mock request and writer
	d := data.CreateServiceData()
	r := httptest.NewRequest(
		http.MethodPost,
		url,
		nil,
	)
	d.Add(serviceName)
	_d_, err := json.MarshalIndent(d, "", " ")
	t.Logf("ServiceData: %v", string(_d_))
	w := httptest.NewRecorder()
	// handle function
	Remove(w, r, &d)
	if w.Result().StatusCode != 201 {
		t.Errorf("StatusCode not %d", w.Result().StatusCode)
	} else {
		t.Logf("StatusCode: %d", w.Result().StatusCode)
	}
	_d, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		t.Fatal("Error in marshal")
	}
	t.Logf("%v", string(_d))
	t.Logf("Data removed: `%s` = `%v`\n`%v`", serviceName, string(_d), *w)
}
