package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/example/myservice/internal/handler"
)

// stubPinger implements handler.DBPinger for testing.
type stubPinger struct {
	err error
}

func (s *stubPinger) Ping(_ context.Context) error { return s.err }

func TestHealthz_ReturnsOK(t *testing.T) {
	r := chi.NewRouter()
	h := handler.NewHealthHandler(&stubPinger{})
	h.Routes(r)

	srv := httptest.NewServer(r)
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
	r := chi.NewRouter()
	h := handler.NewHealthHandler(&stubPinger{})
	h.Routes(r)

	srv := httptest.NewServer(r)
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
	r := chi.NewRouter()
	h := handler.NewHealthHandler(&stubPinger{err: errors.New("connection refused")})
	h.Routes(r)

	srv := httptest.NewServer(r)
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
	r := chi.NewRouter()
	h := handler.NewHealthHandler(nil)
	h.Routes(r)

	srv := httptest.NewServer(r)
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
