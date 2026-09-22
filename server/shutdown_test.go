package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestResolvePortRejectsExtraArguments(t *testing.T) {
	if _, err := resolvePort([]string{"server", "8080", "ignored"}); err == nil {
		t.Fatal("ignored extra argument")
	}
}

func TestServerTimeouts(t *testing.T) {
	server := newHTTPServer(":8080", http.NewServeMux())
	if server.ReadHeaderTimeout <= 0 || server.ReadTimeout <= 0 || server.WriteTimeout <= 0 || server.IdleTimeout <= 0 || server.MaxHeaderBytes <= 0 {
		t.Fatal("missing resource limits")
	}
}

func TestGracefulShutdownWaitsForActiveRequest(t *testing.T) {
	started, release, shuttingDown := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := newHTTPServer(listener.Addr().String(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	}))
	server.RegisterOnShutdown(func() { close(shuttingDown) })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { unblock(); cancel(); _ = server.Close() })
	serveDone := make(chan error, 1)
	go func() { serveDone <- serve(ctx, server, listener, 2*time.Second) }()
	requestDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			defer resp.Body.Close()
			var body []byte
			body, err = io.ReadAll(resp.Body)
			if string(body) != "done" {
				err = errors.New("response lost during shutdown")
			}
		}
		requestDone <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case <-shuttingDown:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not start")
	}
	select {
	case err := <-serveDone:
		t.Fatalf("shutdown returned before request finished: %v", err)
	default:
	}
	unblock()
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-serveDone; err != nil {
		t.Fatal(err)
	}
}

func TestShutdownDeadlineClosesConnections(t *testing.T) {
	started := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := newHTTPServer(listener.Addr().String(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done() }))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer server.Close()
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, listener, 20*time.Millisecond) }()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			resp.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("shutdown error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown hung")
	}
	<-requestDone
}
