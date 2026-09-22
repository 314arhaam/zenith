package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"zenith/server/handlers"
)

func TestClientLifecycle(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(handlers.NewHandler().Routes())
	defer server.Close()
	client, err := New(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if body, err := client.Ping(ctx); err != nil || string(body) != "Pong\n" {
		t.Fatalf("ping: %s %v", body, err)
	}
	if present, err := client.Present(ctx, ""); err != nil || present {
		t.Fatalf("empty: %v %v", present, err)
	}
	name := "api & /+?=\u062e\u062f\u0645\u062a"
	if _, err := client.Add(ctx, "  "+name+"  "); err != nil {
		t.Fatal(err)
	}
	var apiErr *APIError
	if _, err := client.Add(ctx, name); !errors.As(err, &apiErr) || apiErr.StatusCode != 409 || apiErr.Code != "service_exists" {
		t.Fatalf("duplicate: %v", err)
	}
	for _, query := range []string{"", name} {
		if _, present, err := client.Status(ctx, query); err != nil || !present {
			t.Fatalf("status: %v %v", present, err)
		}
		if present, err := client.Present(ctx, query); err != nil || !present {
			t.Fatalf("HEAD: %v %v", present, err)
		}
	}
	if err := client.Remove(ctx, name); err != nil {
		t.Fatal(err)
	}
	if _, present, err := client.Status(ctx, name); err != nil || present {
		t.Fatalf("missing: %v %v", present, err)
	}
	if present, err := client.Present(ctx, name); err != nil || present {
		t.Fatalf("missing HEAD: %v %v", present, err)
	}
	if err := client.Remove(ctx, name); !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
		t.Fatalf("remove missing: %v", err)
	}
}

func TestInvalidBaseURLs(t *testing.T) {
	for _, raw := range []string{"", "://", "localhost:8080", "file:///tmp/x", "ftp://host", "http://", "http://user:pass@host", "http://host?x=1", "http://host?", "http://host#frag"} {
		if _, err := New(raw, nil); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}

func TestBasePath(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/prefix/", http.StripPrefix("/prefix", handlers.NewHandler().Routes()))
	server := httptest.NewServer(mux)
	defer server.Close()
	client, err := New(server.URL+"/prefix/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidAndMisleadingStatusResponses(t *testing.T) {
	tests := []struct {
		name, body, query string
		code              int
	}{
		{"null", "null", "", 200}, {"array", "[]", "", 200}, {"html", "<html>", "", 200},
		{"wrong_key", `{"other":{"service_id":1,"create_datetime":"now"}}`, "api", 200},
		{"missing_fields", `{"api":{}}`, "api", 200},
		{"proxy_404", `<html>Not Found</html>`, "api", 404},
		{"wrong_404", `{"error":"not_found"}`, "api", 404},
		{"server_error", `{"error":"internal","message":"broken"}`, "", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()
			client, _ := New(server.URL, nil)
			if _, _, err := client.Status(context.Background(), tt.query); err == nil {
				t.Fatal("accepted malformed response")
			}
		})
	}
}

func TestRequestCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	client, _ := New(server.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := client.Ping(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("cancellation: %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type countedBody struct {
	remaining, read int64
	closed          bool
}

func (b *countedBody) Read(p []byte) (int, error) {
	if b.remaining == 0 {
		return 0, io.EOF
	}
	n := int64(len(p))
	if n > b.remaining {
		n = b.remaining
	}
	b.remaining -= n
	b.read += n
	return int(n), nil
}
func (b *countedBody) Close() error { b.closed = true; return nil }

func TestResponseReadIsBoundedAndClosed(t *testing.T) {
	body := &countedBody{remaining: MaxResponseBytes * 2}
	client, _ := New("http://example.invalid", &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, ContentLength: -1}, nil
	})})
	if _, err := client.Ping(context.Background()); err == nil || !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("got %v", err)
	}
	if body.read != MaxResponseBytes+1 || !body.closed {
		t.Fatalf("read %d closed %v", body.read, body.closed)
	}
}

func TestKnownOversizeResponseClosedWithoutRead(t *testing.T) {
	body := &countedBody{remaining: MaxResponseBytes * 2}
	client, _ := New("http://example.invalid", &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, ContentLength: MaxResponseBytes * 2}, nil
	})})
	if _, err := client.Ping(context.Background()); err == nil {
		t.Fatal("accepted oversized response")
	}
	if body.read != 0 || !body.closed {
		t.Fatal("did not reject early")
	}
}

func TestRedirectNotFollowed(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	defer server.Close()
	client, _ := New(server.URL, nil)
	var apiErr *APIError
	if _, err := client.Ping(context.Background()); !errors.As(err, &apiErr) || apiErr.StatusCode != 302 {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("followed redirect")
	}
}

func TestHEADRequiresServerHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer server.Close()
	client, _ := New(server.URL, nil)
	if _, err := client.Present(context.Background(), ""); err == nil {
		t.Fatal("accepted non-Zenith HEAD")
	}
}

func TestClientInvalidNames(t *testing.T) {
	client, _ := New("http://example.invalid", nil)
	for _, name := range []string{"   ", "a\nb", strings.Repeat("x", 257)} {
		if _, err := client.Add(context.Background(), name); err == nil {
			t.Fatal("invalid add accepted")
		}
		if err := client.Remove(context.Background(), name); err == nil {
			t.Fatal("invalid remove accepted")
		}
		if _, _, err := client.Status(context.Background(), name); err == nil {
			t.Fatal("invalid status accepted")
		}
		if _, err := client.Present(context.Background(), name); err == nil {
			t.Fatal("invalid HEAD accepted")
		}
	}
}

func TestHEADAllowsLargeRepresentationLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Error("expected HEAD")
		}
		w.Header().Set("Content-Length", "20000000")
		w.Header().Set("X-Zenith-Service-Count", "1")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client, _ := New(server.URL, nil)
	if present, err := client.Present(context.Background(), "api"); err != nil || !present {
		t.Fatalf("HEAD representation length is not a response body: %v %v", present, err)
	}
}
