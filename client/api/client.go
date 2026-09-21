// Package api provides a context-aware client for Zenith's HTTP API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"zenith/core"
	data "zenith/models"
)

const MaxResponseBytes int64 = 8 << 20

// Client is safe for concurrent use; its configuration is immutable after New.
// The supplied http.Client must not be mutated while requests are running.
type Client struct {
	base *url.URL
	http *http.Client
}

// APIError preserves the HTTP status and machine-readable server error code.
type APIError struct {
	StatusCode    int
	Code, Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("server returned %d %s: %s", e.StatusCode, http.StatusText(e.StatusCode), e.Message)
}

func New(baseURL string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, errors.New("base URL must be http(s) with a host, optional path, and no credentials, query or fragment")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second, CheckRedirect: NoRedirect}
	}
	return &Client{base: u, http: httpClient}, nil
}

// NoRedirect prevents silently following API redirects to another endpoint.
func NoRedirect(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }

func (c *Client) do(ctx context.Context, method, path string, query url.Values, payload any) (*http.Response, []byte, error) {
	target := c.base.JoinPath(path)
	target.RawQuery = query.Encode()
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json, text/plain")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s /%s: %w", method, path, err)
	}
	defer resp.Body.Close()
	// A HEAD Content-Length describes the corresponding GET representation,
	// not a body that this response will transfer.
	if method != http.MethodHead && resp.ContentLength > MaxResponseBytes {
		return nil, nil, fmt.Errorf("response exceeds %d bytes; query a single service or use check --quiet", MaxResponseBytes)
	}
	encoded, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(encoded)) > MaxResponseBytes {
		return nil, nil, fmt.Errorf("response exceeds %d bytes; query a single service or use check --quiet", MaxResponseBytes)
	}
	return resp, encoded, nil
}

func requireStatus(resp *http.Response, body []byte, expected int) error {
	if resp.StatusCode == expected {
		return nil
	}
	var serverError data.ErrorResponse
	_ = json.Unmarshal(body, &serverError)
	if serverError.Message == "" {
		serverError.Message = strings.TrimSpace(string(body))
		if len(serverError.Message) > 512 {
			serverError.Message = serverError.Message[:512] + "..."
		}
	}
	if serverError.Message == "" {
		serverError.Message = http.StatusText(resp.StatusCode)
	}
	return &APIError{StatusCode: resp.StatusCode, Code: serverError.Error, Message: serverError.Message}
}

func nameQuery(name string) (url.Values, error) {
	if name == "" {
		return nil, nil
	}
	normalized, valid := data.NormalizeServiceName(name)
	if !valid {
		return nil, errors.New("invalid service name")
	}
	return url.Values{"service": {normalized}}, nil
}

func (c *Client) Ping(ctx context.Context) ([]byte, error) {
	resp, body, err := c.do(ctx, http.MethodGet, "ping", nil, nil)
	if err != nil {
		return nil, err
	}
	if err := requireStatus(resp, body, http.StatusOK); err != nil {
		return nil, err
	}
	if string(body) != "Pong\n" {
		return nil, errors.New("unexpected ping response from server")
	}
	return body, nil
}

func (c *Client) Add(ctx context.Context, name string) ([]byte, error) {
	payload := data.AddRequest{ServiceName: name}
	if !payload.Validate() {
		return nil, errors.New("invalid service name")
	}
	resp, body, err := c.do(ctx, http.MethodPost, "add", nil, payload)
	if err != nil {
		return nil, err
	}
	if err := requireStatus(resp, body, http.StatusCreated); err != nil {
		return nil, err
	}
	var service core.Service
	if err := json.Unmarshal(body, &service); err != nil || service.ServiceID == 0 || service.CreateDateTime == "" {
		return nil, errors.New("invalid service registration response")
	}
	return body, nil
}

func (c *Client) Remove(ctx context.Context, name string) error {
	payload := data.RemoveRequest{ServiceName: name}
	if !payload.Validate() {
		return errors.New("invalid service name")
	}
	resp, body, err := c.do(ctx, http.MethodDelete, "remove", nil, payload)
	if err != nil {
		return err
	}
	return requireStatus(resp, body, http.StatusNoContent)
}

// Status preserves the server's JSON for display, and validates its structure.
// Only a structured service_not_found 404 means the named service is absent.
func (c *Client) Status(ctx context.Context, name string) ([]byte, bool, error) {
	query, err := nameQuery(name)
	if err != nil {
		return nil, false, err
	}
	resp, body, err := c.do(ctx, http.MethodGet, "status", query, nil)
	if err != nil {
		return nil, false, err
	}
	if err := requireStatus(resp, body, http.StatusOK); err != nil {
		var serverError *APIError
		if name != "" && errors.As(err, &serverError) && serverError.StatusCode == http.StatusNotFound && serverError.Code == "service_not_found" {
			return body, false, nil
		}
		return nil, false, err
	}
	var services map[string]core.Service
	if err := json.Unmarshal(body, &services); err != nil || services == nil {
		return nil, false, errors.New("invalid status response: expected an object of services")
	}
	for key, service := range services {
		if _, ok := data.NormalizeServiceName(key); !ok || service.ServiceID == 0 || service.CreateDateTime == "" {
			return nil, false, errors.New("invalid service in status response")
		}
	}
	if name != "" {
		if _, ok := services[query.Get("service")]; !ok {
			return nil, false, errors.New("status response does not contain the requested service")
		}
	}
	return body, len(services) > 0, nil
}

// Present performs an O(1), body-free availability probe on the updated server.
// The header also distinguishes a missing service from an unrelated proxy 404.
func (c *Client) Present(ctx context.Context, name string) (bool, error) {
	query, err := nameQuery(name)
	if err != nil {
		return false, err
	}
	resp, body, err := c.do(ctx, http.MethodHead, "status", query, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK && !(name != "" && resp.StatusCode == http.StatusNotFound) {
		return false, requireStatus(resp, body, http.StatusOK)
	}
	count, err := strconv.ParseUint(resp.Header.Get("X-Zenith-Service-Count"), 10, 64)
	if err != nil {
		return false, errors.New("server does not support Zenith HEAD probes; omit --quiet")
	}
	if name != "" && ((resp.StatusCode == http.StatusNotFound && count != 0) || (resp.StatusCode == http.StatusOK && count != 1)) {
		return false, errors.New("inconsistent service count in HEAD response")
	}
	return count > 0, nil
}
