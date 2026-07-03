package handlefuncs

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	data "zenith/models"
)

func TestAdd(t *testing.T) {
	// mock data
	serviceName := "test_service-01"
	url := "/add?service=" + serviceName
	// mock request and writer
	d := data.CreateServiceData()
	r := httptest.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	w := httptest.NewRecorder()
	Add(w, r, &d)
	_d, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		t.Fatal("Error in marshal")
	}
	log.Printf("%v", string(_d))
	res, ok := d[serviceName]
	if !ok {
		t.Fatal("Endpoint /add Failed.")
	} else {
		_res, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			t.Fatal("Error in marshal")
		}
		log.Printf("Data added: `%s` = `%v`", serviceName, string(_res))
	}
}
