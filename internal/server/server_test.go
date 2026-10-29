package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/NotAFlightRisk/peng-top/internal/collect"
	"github.com/NotAFlightRisk/peng-top/internal/config"
)

func serve(t *testing.T, cfg config.Config, host, path string, login ...string) *httptest.ResponseRecorder {
	t.Helper()
	ui := fstest.MapFS{"index.html": {Data: []byte("<!doctype html>")}}
	h := New(cfg, collect.New(collect.Env{}, 0, 1), collect.Info{}, "test").Handler(ui)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = host
	if len(login) == 2 {
		req.SetBasicAuth(login[0], login[1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHostGuard(t *testing.T) {
	cfg := config.Config{AllowedHosts: []string{"nas.lan"}}
	for host, want := range map[string]int{
		"localhost:7340":     http.StatusOK,
		"127.0.0.1:7340":     http.StatusOK,
		"[::1]:7340":         http.StatusOK,
		"192.168.1.20":       http.StatusOK,
		"nas.lan:7340":       http.StatusOK,
		"app.localhost":      http.StatusOK,
		"evil.example:7340":  http.StatusForbidden,
		"rebind.attacker.io": http.StatusForbidden,
	} {
		if got := serve(t, cfg, host, "/healthz").Code; got != want {
			t.Errorf("%s: got %d, want %d", host, got, want)
		}
	}
}

func TestStarAllowsAnyHost(t *testing.T) {
	cfg := config.Config{AllowedHosts: []string{"*"}}
	if got := serve(t, cfg, "whatever.example", "/healthz").Code; got != http.StatusOK {
		t.Fatalf("got %d", got)
	}
}

func TestAuthGuardsEverythingButHealthz(t *testing.T) {
	cfg := config.Config{Auth: "dino:hunter2"}
	if got := serve(t, cfg, "anywhere.example", "/api/info").Code; got != http.StatusUnauthorized {
		t.Errorf("no login: got %d", got)
	}
	if got := serve(t, cfg, "anywhere.example", "/api/info", "dino", "wrong").Code; got != http.StatusUnauthorized {
		t.Errorf("wrong password: got %d", got)
	}
	if got := serve(t, cfg, "anywhere.example", "/api/info", "dino", "hunter2").Code; got != http.StatusOK {
		t.Errorf("right password: got %d", got)
	}
	if got := serve(t, cfg, "anywhere.example", "/healthz").Code; got != http.StatusOK {
		t.Errorf("healthz: got %d", got)
	}
}

func TestEveryResponseCarriesTheSecurityHeaders(t *testing.T) {
	rec := serve(t, config.Config{}, "localhost", "/")
	for _, h := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Referrer-Policy"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("missing %s", h)
		}
	}
}
