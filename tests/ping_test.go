package handlers

import (
	"io"
	"net/http"
	"testing"
	hd "zenith/server/handlers"
)

func TestPing(t *testing.T) {
	h := hd.NewHandler()
	w, r := responseAndRequestBuild(
		http.MethodGet,
		"/ping",
		nil,
	)
	h.Ping(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("Error in testing /ping. StatusCode %d", w.Result().StatusCode)
	}
	defer w.Result().Body.Close()
	if val, err := io.ReadAll(w.Result().Body); err != nil {
		t.Fatalf("Error in reading response: %s", err.Error())
	} else if string(val) != "Pong" {
		t.Fatalf("Error in reponse: not Pong: %s", val)
	} else {
		t.Logf("%s", val)
	}
}
