package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"zenith/server/handlers"
)

func TestCommandInstancesDoNotLeakFlags(t *testing.T) {
	first := NewRootCommand()
	first.SetArgs([]string{"--url", "http://example.invalid", "ping", "--until", "-1"})
	first.SetOut(io.Discard)
	first.SetErr(io.Discard)
	if err := first.Execute(); err == nil {
		t.Fatal("expected invalid --until")
	}
	second := NewRootCommand()
	if got, _ := second.PersistentFlags().GetString("url"); got != "http://127.0.0.1:8080" {
		t.Fatal("URL flag leaked")
	}
	ping, _, err := second.Find([]string{"ping"})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := ping.Flags().GetInt("until"); got != 0 {
		t.Fatal("--until leaked")
	}
}

func TestParallelCommandInvocations(t *testing.T) {
	server := httptest.NewServer(handlers.NewHandler().Routes())
	defer server.Close()
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("api-%d", i)
			for _, command := range []string{"add", "status", "remove"} {
				out, err := ExecuteWithArgs([]string{"--url", server.URL, command, name})
				if err != nil {
					t.Errorf("%s: %v %s", command, err, out)
					return
				}
				if command == "status" && !strings.Contains(out, `"`+name+`"`) {
					t.Errorf("wrong output %s", out)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestInvalidFlagsAndArguments(t *testing.T) {
	tests := [][]string{
		{"add"}, {"remove"}, {"add", "a", "b"}, {"status", "a", "b"}, {"ping", "extra"},
		{"ping", "--until", "-1"}, {"check", "--max-retry", "0"}, {"check", "--sleep", "0"},
		{"check", "--interval", "-1"}, {"--timeout", "0s", "ping"},
		{"--url", "ftp://host", "ping"}, {"add", "   "}, {"status", ""},
	}
	for _, args := range tests {
		if _, err := ExecuteWithArgs(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}

func TestPingDeadlineBoundsHTTPRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	started := time.Now()
	_, err := ExecuteWithArgs([]string{"--url", server.URL, "--timeout", "10s", "ping", "--until", "1"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 2500*time.Millisecond {
		t.Fatalf("one-second deadline took %s", elapsed)
	}
}

// Synchronize on an actual HTTP request rather than a 100ms scheduling guess.
func TestCheckCancellationAfterRequestStarts(t *testing.T) {
	for _, quiet := range []bool{false, true} {
		t.Run(fmt.Sprintf("quiet=%v", quiet), func(t *testing.T) {
			requested := make(chan struct{})
			var once sync.Once
			h := handlers.NewHandler()
			h.Core.Add("api")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.Routes().ServeHTTP(w, r)
				once.Do(func() { close(requested) })
			}))
			defer server.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			args := []string{"--url", server.URL, "check", "api", "--sleep", "3600"}
			if quiet {
				args = append(args, "--quiet")
			}
			type result struct {
				out string
				err error
			}
			done := make(chan result, 1)
			go func() {
				out, err := ExecuteWithContext(ctx, args)
				done <- result{out, err}
			}()
			select {
			case <-requested:
			case <-time.After(5 * time.Second):
				t.Fatal("no poll reached the test server")
			}
			cancel()
			select {
			case got := <-done:
				if !errors.Is(got.err, context.Canceled) {
					t.Fatalf("cancellation: %v", got.err)
				}
				if quiet && got.out != "" {
					t.Fatalf("quiet output=%q", got.out)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("cancelled check waited for its hour-long poll interval")
			}
		})
	}
}

func TestCheckRetryLimitAndQuietMethod(t *testing.T) {
	var calls atomic.Int64
	h := handlers.NewHandler()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != "HEAD" {
			t.Error("not a HEAD request")
		}
		h.Routes().ServeHTTP(w, r)
	}))
	defer server.Close()
	out, err := ExecuteWithArgs([]string{"--url", server.URL, "check", "missing", "--quiet", "--max-retry", "1"})
	if err == nil || !strings.Contains(err.Error(), "max retry") || out != "" || calls.Load() != 1 {
		t.Fatalf("out=%q err=%v calls=%d", out, err, calls.Load())
	}
}

func TestCheckRetryCounterResetsAfterSuccess(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 2 {
			w.Header().Set("X-Zenith-Service-Count", "1")
			w.WriteHeader(200)
		} else {
			w.Header().Set("X-Zenith-Service-Count", "0")
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	_, err := ExecuteWithArgs([]string{"--url", server.URL, "check", "api", "--quiet", "--max-retry", "2", "--interval", "1", "--sleep", "1"})
	if err == nil || !strings.Contains(err.Error(), "max retry") || calls.Load() != 4 {
		t.Fatalf("calls=%d err=%v", calls.Load(), err)
	}
}

func TestPingRetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(503)
			return
		}
		_, _ = io.WriteString(w, "Pong\n")
	}))
	defer server.Close()
	out, err := ExecuteWithArgs([]string{"--url", server.URL, "ping", "--until", "5"})
	if err != nil || !strings.Contains(out, "Pong\n") || calls.Load() != 2 {
		t.Fatalf("%q %v calls=%d", out, err, calls.Load())
	}
}

func TestSecondsDurationOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if int64(maxInt) > 9223372036 {
		if _, err := secondsDuration("sleep", maxInt, false); err == nil {
			t.Fatal("overflow accepted")
		}
	}
}
