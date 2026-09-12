package data_test

import (
	"io"
	"strings"
	"testing"
	data "zenith/models"
)

func TestAddRequest(t *testing.T) {
	a := data.AddRequest{
		ServiceName: "test_service",
	}
	if a.Validate() == false {
		t.Fatalf("[x] error in validation: %v", a)
	}
	t.Logf("[*] created request: %v", a)
	as, err := data.ToJson(a)
	if err != nil {
		t.Fatalf("[x] error in converting request to json")
	}
	t.Logf("[*] request as string: %s", as)
	var aplus data.AddRequest
	var rc io.ReadCloser = io.NopCloser(
		strings.NewReader(as),
	)
	defer rc.Close()
	if err := data.Decode(&aplus, rc); err != nil {
		t.Fatalf("[x] error in Decode(...): %s", err.Error())
	}
	t.Logf("[*] decoded result: %v", aplus)
}

func TestRemoveRequest(t *testing.T) {
	r := data.RemoveRequest{
		ServiceName: "test_service",
	}
	if r.Validate() == false {
		t.Fatalf("[x] error in validation: %v", r)
	}
	t.Logf("[*] created request: %v", r)
	as, err := data.ToJson(r)
	if err != nil {
		t.Fatalf("[x] error in converting request to json")
	}
	t.Logf("[*] request as string: %s", as)
	var rplus data.RemoveRequest
	var rc io.ReadCloser = io.NopCloser(
		strings.NewReader(as),
	)
	defer rc.Close()
	if err := data.Decode(&rplus, rc); err != nil {
		t.Fatalf("[x] error in Decode(...): %s", err.Error())
	}
	t.Logf("[*] decoded result: %v", rplus)
}
