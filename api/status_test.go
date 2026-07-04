package handlefuncs

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	data "zenith/models"
)

func TestStatus(t *testing.T) {
	// title
	log.Printf("*** STATUS ENDPOINT ***")
	// mock data
	url := "/status"
	serviceName := "test_service_status-endpoint"
	dataPost, err := json.Marshal(data.RequestPayload{ServiceName: serviceName})
	if err != nil {
		t.Fatal("Error in Payload create")
	}
	payload := strings.NewReader(string(dataPost))
	// mock request and writer
	d := data.CreateServiceData()
	d.Add(serviceName)
	r := httptest.NewRequest(
		http.MethodGet,
		url,
		payload,
	)
	w := httptest.NewRecorder()
	// handle function
	Status(w, r, &d)
	defer w.Result().Body.Close()
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", body)
}
