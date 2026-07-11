package integration

import (
	"net/http"
	"net/http/httptest"
	"zenith/server/handlers"
)

func NewTestServer() *httptest.Server {
	h := handlers.NewHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", h.Add)
	mux.HandleFunc("DELETE /remove", h.Remove)
	mux.HandleFunc("GET /status", h.Status)
	mux.HandleFunc("GET /ping", h.Ping)
	server := httptest.NewServer(mux)
	return server
}
