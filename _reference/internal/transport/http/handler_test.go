package http_test

import (
	"context"
	"encoding/json"
	"errors"

	"net/http"
	"net/http/httptest"
	"testing"

	transporthttp "github.com/example/myservice/internal/transport/http"
)

// stubPinger implements transporthttp.DBPinger for testing.
type stubPinger struct {
	err error
}

func (s *stubPinger) Ping(_ context.Context) error { return s.err }

func TestHealthz_ReturnsOK(t *testing.T) {
	handler := transporthttp.NewHandler(nil, &stubPinger{}, false)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status \"ok\", got %q", body["status"])
	}
}

func TestReadyz_DBHealthy(t *testing.T) {
	handler := transporthttp.NewHandler(nil, &stubPinger{}, false)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status \"ok\", got %q", body["status"])
	}
}

func TestReadyz_DBUnhealthy(t *testing.T) {
	handler := transporthttp.NewHandler(nil, &stubPinger{err: errors.New("connection refused")}, false)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}
}

func TestReadyz_NilDB(t *testing.T) {
	handler := transporthttp.NewHandler(nil, nil, false)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}
}
