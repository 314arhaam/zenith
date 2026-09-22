package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"zenith/core"
	"zenith/server/handlers"
)

func TestHelpAndCompletionUsePortableBinaryName(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help", "check"}, {"completion", "bash"}, {"completion", "zsh"}, {"completion", "fish"}, {"completion", "powershell"}} {
		t.Run(strings.Join(args, "-"), func(t *testing.T) {
			out, err := ExecuteWithArgs(args)
			if err != nil || !strings.Contains(out, "zenith-client") {
				t.Fatalf("help/completion: err=%v output=%q", err, out)
			}
		})
	}
}

func TestServiceNamesSurviveCLIAndURLRoundTrip(t *testing.T) {
	server := httptest.NewServer(handlers.NewHandler().Routes())
	defer server.Close()
	for _, name := range []string{"worker east", "worker \u03a9", "\u062e\u062f\u0645\u062a", "api & worker/+?=", "worker\\directory"} {
		t.Run(name, func(t *testing.T) {
			if out, err := ExecuteWithArgs([]string{"--url", server.URL, "add", name}); err != nil {
				t.Fatalf("add: %v %s", err, out)
			}
			out, err := ExecuteWithArgs([]string{"--url", server.URL, "status", name})
			var services map[string]core.Service
			if err != nil || json.Unmarshal([]byte(out), &services) != nil {
				t.Fatalf("status: %v %q", err, out)
			}
			if _, exists := services[name]; !exists || len(services) != 1 {
				t.Fatalf("name changed across CLI/HTTP: %q => %v", name, services)
			}
			if _, err := ExecuteWithArgs([]string{"--url", server.URL, "remove", name}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWaitCancelsLongDuration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- wait(ctx, time.Hour) }()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("wait cancellation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled wait did not stop")
	}
}
