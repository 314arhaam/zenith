package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	data "zenith/models"
)

func TestRemoveEndPoint(t *testing.T) {
	endpoint := "/remove"
	serviceName := "test_service"
	var request string
	var err error
	if request, err = data.ToJson(&data.AddRequest{ServiceName: serviceName}); err != nil {
		t.Fatalf("Cannot convert request struct to json: %s", err)
	}
	t.Log("\n[*] Request body OK")
	server := NewTestServer()
	defer server.Close()
	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete,
		server.URL+endpoint,
		strings.NewReader(request),
	)
	if err != nil {
		t.Fatalf("\n[x] Error in DELETE request creation: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("\n[x] Error in DELETE method: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("\n[x] Error in response: StatusCode %d", resp.StatusCode)
	}
	if v, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("\n[x] Error in io.ReadAll on response %v", err)
	} else {
		t.Logf("\n[*] Response OK, StatusCode %d, Body %s", resp.StatusCode, string(v))
	}
}
