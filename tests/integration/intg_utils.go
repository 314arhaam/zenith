package integration

import (
	"net/http/httptest"
	"zenith/server/handlers"
)

func NewTestServer() *httptest.Server {
	return httptest.NewServer(handlers.NewHandler().Routes())
}
