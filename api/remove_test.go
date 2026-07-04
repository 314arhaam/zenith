package handlefuncs

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	data "zenith/models"
)

func TestRemove(t *testing.T) {
	// title
	log.Printf("*** REMOVE ENDPOINT ***")
	// mock data
	serviceName := "test_service-01"
	url := "/remove?service=" + serviceName
	// mock request and writer
	d := data.CreateServiceData()
	r := httptest.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	d.Add(serviceName)
	_d_, err := json.MarshalIndent(d, "", " ")
	log.Printf("ServiceData: %v", string(_d_))
	w := httptest.NewRecorder()
	// handle function
	Remove(w, r, &d)
	if w.Result().StatusCode != 201 {
		t.Errorf("StatusCode not %d", w.Result().StatusCode)
	} else {
		log.Printf("StatusCode: %d", w.Result().StatusCode)
	}
	_d, err := json.MarshalIndent(d, "", " ")
	if err != nil {
		t.Fatal("Error in marshal")
	}
	log.Printf("%v", string(_d))
	log.Printf("Data removed: `%s` = `%v`\n`%v`", serviceName, string(_d), *w)
}
