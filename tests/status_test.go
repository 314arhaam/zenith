package handlers

import (
	"io"
	"log"
	"net/http"
	"testing"
	hd "zenith/server/handlers"
)

func TestMethodNotAllowedStatus(t *testing.T) {
	h := hd.NewHandler()
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
	h := hd.NewHandler()
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

func TestStatusSignleService(t *testing.T) {
	h := hd.NewHandler()
	// mock data
	serviceName := "mock_single_service"
	url := "/status?service=" + serviceName
	// mock request and writer
	h.Core.Add(serviceName)
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
