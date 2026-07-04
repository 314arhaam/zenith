package handlefuncs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
	serviceName := "test_service-01"
	url := "/remove"
	// mock request and writer
	d := data.CreateServiceData()
	d.Add(serviceName)
	w, r := responseAndRequestBuild(http.MethodGet, url, nil)
	Add(w, r, &d)
	if w.Result().StatusCode != 405 {
		t.Fatalf("Error: `Method Not Allowed` doesn't work.")
	}
}

func TestAdd(t *testing.T) {
	// mock data
	serviceName := "test_service-01"
	url := "/add?service=" + serviceName
	// mock request and writer
	d := data.CreateServiceData()
	w, r := responseAndRequestBuild(http.MethodPost, url, nil)
	// handle function
	Add(w, r, &d)
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
	res, ok := d[serviceName]
	if !ok {
		t.Fatal("Endpoint /add Failed.")
	} else {
		_res, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			t.Fatal("Error in marshal")
		}
		t.Logf("Data added: `%s` = `%v`\n`%v`", serviceName, string(_res), *w)
	}
}
