package handlefuncs

import (
	"io"
	"log"
	"net/http"
	"testing"
	data "zenith/models"
)

func TestMethodNotAllowedStatus(t *testing.T) {
	// mock data
	url := "/status"
	// mock request and writer
	d := data.CreateServiceData()
	d.Add("mock_service")
	w, r := responseAndRequestBuild(
		http.MethodPost,
		url,
		nil,
	)
	// handle function
	Status(w, r, &d)
	if w.Result().StatusCode != 405 {
		t.Fatalf("Error: `Method Not Allowed` doesn't work.")
	}
}

func TestStatus(t *testing.T) {
	// mock data
	url := "/status"
	// mock request and writer
	d := data.CreateServiceData()
	d.Add("mock_service")
	w, r := responseAndRequestBuild(
		http.MethodGet,
		url,
		nil,
	)
	// handle function
	Status(w, r, &d)
	defer w.Result().Body.Close()
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("%s", body)
}
