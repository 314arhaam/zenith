package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddRequest(t *testing.T) {
	var rs AddRequest
	test_service_name := "test_service_1"
	payload := fmt.Sprintf(`{"service_name": "%s"}`, test_service_name)
	t.Logf("[*] Initialized `rs` %v", rs)
	rq := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(payload),
	)
	t.Logf("[*] Initialized `rq`")
	defer rq.Body.Close()
	if err := Decode(&rs, rq.Body); err != nil {
		t.Fatalf("Error in Decode: %s", err.Error())
	}
	if ok := rs.Validate(); !ok {
		t.Fatalf("Error in validation")
	}
	if rs.ServiceName != test_service_name {
		t.Fatalf("Data doesnt match")
	}
	t.Logf("[*] Payload: `%s` AddRequest `%v`", payload, rs)
}

func TestRemoveRequest(t *testing.T) {
	var rs RemoveRequest
	test_service_name := "test_service_2"
	payload := fmt.Sprintf(`{"service_name": "%s"}`, test_service_name)
	t.Logf("[*] Initialized `rs` %v", rs)
	rq := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(payload),
	)
	t.Logf("[*] Initialized `rq`")
	defer rq.Body.Close()
	if err := Decode(&rs, rq.Body); err != nil {
		t.Fatalf("Error in Decode: %s", err.Error())
	}
	if ok := rs.Validate(); !ok {
		t.Fatalf("Error in validation")
	}
	if rs.ServiceName != test_service_name {
		t.Fatalf("Data doesnt match")
	}
	t.Logf("[*] Payload: `%s` AddRequest `%v`", payload, rs)
}
