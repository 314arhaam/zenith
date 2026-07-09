package tests

import (
	"testing"
	"zenith/client/cmd"
	"zenith/tests/integration"
)

func TestSubcommandAdd(t *testing.T) {
	var out string
	var err error
	mockServiceName := "cli_test_service_123"
	serviceToRemove := "service_to_remove"
	server := integration.NewTestServer()
	defer server.Close()
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "add", mockServiceName})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status"})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status", mockServiceName})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "add", serviceToRemove})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status"})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "remove", serviceToRemove})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status"})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "status", serviceToRemove})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Log(string(out))
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "ping"})
	if err != nil {
		t.Fatal(err, out)
	}
	t.Logf("%s\n", string(out))
}

func TestRemoveError(t *testing.T) {
	var out string
	var err error
	serviceToRemove := "service_to_remove"
	server := integration.NewTestServer()
	defer server.Close()
	//
	out, err = cmd.ExecuteWithArgs([]string{"--url", server.URL, "remove", serviceToRemove})
	if err == nil {
		t.Fatalf("must return 404 error: %s | %s", err, out)
	}
}
