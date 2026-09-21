package data

import (
	"strings"
	"testing"
)

func TestDecodeAndValidate(t *testing.T) {
	var req AddRequest
	if err := Decode(&req, strings.NewReader(`{"service_name":" api "}`)); err != nil {
		t.Fatal(err)
	}
	if !req.Validate() || req.ServiceName != "api" {
		t.Fatalf("unexpected request after validation: %+v", req)
	}
}

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	var req AddRequest
	if err := Decode(&req, strings.NewReader(`{"service_name":"api"} {"service_name":"other"}`)); err == nil {
		t.Fatal("expected trailing JSON to be rejected")
	}
}
