package tests

import (
	"strings"
	"testing"
	"zenith/client/cmd"
	"zenith/tests/integration"
)

func TestClientLifecycle(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	out, err := cmd.ExecuteWithArgs([]string{"--url", server.URL, "add", "api"})
	if err != nil {
		t.Fatalf("add failed: %v; output=%s", err, out)
	}
	if !strings.Contains(out, "service_id") {
		t.Fatalf("add output missing service data: %s", out)
	}

	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status", "api"})
	if err != nil {
		t.Fatalf("status failed: %v; output=%s", err, out)
	}
	if !strings.Contains(out, `"api"`) {
		t.Fatalf("status output missing api: %s", out)
	}

	if _, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "add", "api"}); err == nil {
		t.Fatal("duplicate add should return an error")
	}

	if _, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "remove", "api"}); err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if _, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status", "api"}); err == nil {
		t.Fatal("status for removed service should return an error")
	}
}

func TestClientPing(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()
	out, err := cmd.ExecuteWithArgs([]string{"--url", server.URL, "ping"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Pong\n" {
		t.Fatalf("ping output = %q; want Pong\\n", out)
	}
}
