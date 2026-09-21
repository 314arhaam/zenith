package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	data "zenith/models"
)

func TestRequestBoundaries(t *testing.T) {
	for _, endpoint := range []struct{ path, method string }{{"/add", "POST"}, {"/remove", "DELETE"}} {
		for _, tt := range []struct {
			name, body, contentType string
			status                  int
			code                    string
		}{
			{"malformed", "{", "", 400, "invalid_json"},
			{"unknown_field", `{"service_name":"api","typo":true}`, "", 400, "invalid_json"},
			{"trailing", `{"service_name":"api"}{}`, "", 400, "invalid_json"},
			{"missing", `null`, "", 400, "invalid_payload"},
			{"too_long", `{"service_name":"` + strings.Repeat("x", 257) + `"}`, "", 400, "invalid_payload"},
			{"control", `{"service_name":"a\nb"}`, "", 400, "invalid_payload"},
			{"too_large", `{"service_name":"` + strings.Repeat("x", 8192) + `"}`, "application/json", 413, "payload_too_large"},
			{"wrong_type", `{"service_name":"api"}`, "text/plain", 415, "unsupported_media_type"},
		} {
			t.Run(endpoint.path+"/"+tt.name, func(t *testing.T) {
				h := NewHandler()
				req := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", tt.contentType)
				rec := httptest.NewRecorder()
				h.Routes().ServeHTTP(rec, req)
				var payload data.ErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if rec.Code != tt.status || payload.Error != tt.code {
					t.Fatalf("got %d %s", rec.Code, rec.Body.String())
				}
				if h.Core.Len() != 0 {
					t.Fatal("invalid request mutated registry")
				}
			})
		}
	}
}

func TestRoutingMethodsAndHEAD(t *testing.T) {
	h := NewHandler()
	h.Core.Add("api")
	for _, tt := range []struct {
		method, path string
		status       int
		allow, count string
	}{
		{"GET", "/add", 405, "POST", ""}, {"POST", "/remove", 405, "DELETE", ""},
		{"POST", "/status", 405, "GET, HEAD", ""}, {"POST", "/ping", 405, "GET, HEAD", ""},
		{"GET", "/unknown", 404, "", ""}, {"GET", "/status/unknown", 404, "", ""},
		{"HEAD", "/ping", 200, "", ""}, {"HEAD", "/status", 200, "", "1"},
		{"HEAD", "/status?service=api", 200, "", "1"}, {"HEAD", "/status?service=missing", 404, "", "0"},
	} {
		rec := httptest.NewRecorder()
		h.Routes().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
		if rec.Code != tt.status || rec.Header().Get("Allow") != tt.allow || rec.Header().Get("X-Zenith-Service-Count") != tt.count {
			t.Errorf("%s %s: %d %v", tt.method, tt.path, rec.Code, rec.Header())
		}
		if tt.method == "HEAD" && rec.Body.Len() != 0 {
			t.Errorf("HEAD response had body: %s", rec.Body)
		}
	}
	h.Core.Remove("api")
	rec := call(t, h.Status, "HEAD", "/status", "")
	if rec.Header().Get("X-Zenith-Service-Count") != "0" {
		t.Fatal("empty count was not zero")
	}
}

func TestStatusCacheLifecycleAndQueries(t *testing.T) {
	h := NewHandler()
	for _, path := range []string{"/status?service=", "/status?service=a&service=b", "/status?service=%ZZ", "/status?service=a%00b"} {
		if rec := call(t, h.Status, "GET", path, ""); rec.Code != 400 {
			t.Errorf("invalid query %s: %d", path, rec.Code)
		}
	}
	if rec := call(t, h.Status, "GET", "/status", ""); rec.Body.String() != "{}\n" {
		t.Fatal("expected empty JSON")
	}
	h.Core.Add("api")
	for i := 0; i < 2; i++ {
		rec := call(t, h.Status, "GET", "/status", "")
		if !strings.Contains(rec.Body.String(), `"api"`) {
			t.Fatal("cache omitted addition")
		}
	}
	if rec := call(t, h.Status, "GET", "/status?service=%20api%20", ""); rec.Code != 200 {
		t.Fatal("normalization inconsistent")
	}
	h.Core.Remove("api")
	if rec := call(t, h.Status, "GET", "/status", ""); rec.Body.String() != "{}\n" {
		t.Fatal("cache retained removed service")
	}
}

type slowWriter struct {
	header           http.Header
	entered, release chan struct{}
	blocked          bool
}

func (w *slowWriter) Header() http.Header { return w.header }
func (w *slowWriter) WriteHeader(int)     {}
func (w *slowWriter) Write(p []byte) (int, error) {
	if !w.blocked {
		w.blocked = true
		close(w.entered)
		<-w.release
	}
	return len(p), nil
}

func TestSlowResponseDoesNotBlockWrites(t *testing.T) {
	h := NewHandler()
	h.Core.Add("api")
	writer := &slowWriter{header: make(http.Header), entered: make(chan struct{}), release: make(chan struct{})}
	done := make(chan struct{})
	go func() { defer close(done); h.Status(writer, httptest.NewRequest("GET", "/status", nil)) }()
	defer func() { close(writer.release); <-done }()
	select {
	case <-writer.entered:
	case <-time.After(time.Second):
		t.Fatal("handler did not write")
	}
	registered := make(chan struct{})
	go func() { h.Core.Add("worker"); close(registered) }()
	select {
	case <-registered:
	case <-time.After(time.Second):
		t.Fatal("network write held registry lock")
	}
}

func TestHTTPHeadHasNoBody(t *testing.T) {
	server := httptest.NewServer(NewHandler().Routes())
	defer server.Close()
	resp, err := server.Client().Head(server.URL + "/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) != 0 || resp.Header.Get("X-Zenith-Service-Count") != "0" {
		t.Fatal("invalid HEAD response")
	}
}
