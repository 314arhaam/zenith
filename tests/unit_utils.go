package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
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
