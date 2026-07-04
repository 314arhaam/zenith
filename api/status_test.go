package handlefuncs

import (
	"io"
	"log"
	"net/http"
	"testing"
)

func TestMethodNotAllowedStatus(t *testing.T) {
	h := NewHandler()
	// mock data
	url := "/status"
	// mock request and writer
	h.Core.Add("mock_service")
	w, r := responseAndRequestBuild(
		http.MethodPost,
		url,
		nil,
	)
	// handle function
	h.Status(w, r)
	if w.Result().StatusCode != 405 {
		t.Fatalf("Error: `Method Not Allowed` doesn't work.")
	}
}

func TestStatus(t *testing.T) {
	h := NewHandler()
	// mock data
	url := "/status"
	// mock request and writer
	h.Core.Add("mock_service")
	w, r := responseAndRequestBuild(
		http.MethodGet,
		url,
		nil,
	)
	// handle function
	h.Status(w, r)
	defer w.Result().Body.Close()
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		log.Fatal(err)
	}
	t.Logf("%s", body)
}
