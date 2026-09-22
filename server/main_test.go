package main

import "testing"

func TestResolvePortDefault(t *testing.T) {
	got, err := resolvePort([]string{"zenith-server"})
	if err != nil || got != ":8080" {
		t.Fatalf("resolvePort default = %q, %v; want :8080, nil", got, err)
	}
}

func TestResolvePortExplicit(t *testing.T) {
	got, err := resolvePort([]string{"zenith-server", "9090"})
	if err != nil || got != ":9090" {
		t.Fatalf("resolvePort explicit = %q, %v; want :9090, nil", got, err)
	}
}

func TestResolvePortInvalid(t *testing.T) {
	for _, value := range []string{"abc", "0", "65536", "-1"} {
		if _, err := resolvePort([]string{"zenith-server", value}); err == nil {
			t.Fatalf("resolvePort(%q) should fail", value)
		}
	}
}
