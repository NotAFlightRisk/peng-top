// Package server serves the UI and the readings
package server

import (
	"compress/gzip"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NotAFlightRisk/peng-top/internal/collect"
	"github.com/NotAFlightRisk/peng-top/internal/config"
	"github.com/NotAFlightRisk/peng-top/internal/logx"
)

type Server struct {
	cfg     config.Config
	col     *collect.Collector
	info    collect.Info
	version string
	hosts   map[string]bool
}

func New(cfg config.Config, col *collect.Collector, info collect.Info, version string) *Server {
	hosts := map[string]bool{"localhost": true}
	for _, h := range cfg.AllowedHosts {
		hosts[h] = true
	}
	if name, err := os.Hostname(); err == nil {
		hosts[strings.ToLower(name)], hosts[strings.ToLower(name)+".local"] = true, true
	}
	if info.Hostname != "" {
		hosts[strings.ToLower(info.Hostname)] = true
	}
	return &Server{cfg: cfg, col: col, info: info, version: version, hosts: hosts}
}

// Handler is the whole app, UI included
func (s *Server) Handler(ui fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /api/info", s.infoHandler)
	mux.HandleFunc("GET /api/snapshot", s.snapshot)
	mux.HandleFunc("GET /api/stream", s.stream)
	mux.Handle("GET /", static(ui))
	return s.guard(headers(s.auth(mux)))
}

func (s *Server) infoHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, struct {
		collect.Info
		Version  string            `json:"version"`
		Interval int64             `json:"interval"`
		Issues   map[string]string `json:"issues"`
	}{s.info, s.version, s.cfg.Interval.Milliseconds(), logx.Issues()})
}

func (s *Server) snapshot(w http.ResponseWriter, _ *http.Request) {
	latest := s.col.Latest()
	if latest == nil {
		http.Error(w, "no reading yet, give it a sec", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(latest)
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	var out io.Writer = w
	flush := rc.Flush
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gz, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		out = gz
		flush = func() error { return errors.Join(gz.Flush(), rc.Flush()) }
	}
	history, feed, cancel := s.col.Subscribe()
	defer cancel()

	send := func(format string, args ...any) bool {
		rc.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if _, err := fmt.Fprintf(out, format, args...); err != nil {
			return false
		}
		return flush() == nil
	}
	if !send("retry: 3000\nevent: history\ndata: %s\n\n", history) {
		return
	}
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		ok := true
		select {
		case <-r.Context().Done():
			return
		case reading := <-feed:
			ok = send("data: %s\n\n", reading)
		case <-ping.C:
			ok = send(": still here\n\n")
		}
		if !ok {
			slog.Debug("stream dropped", "client", r.RemoteAddr)
			return
		}
	}
}

// guard turns away Host headers we don't know, so a DNS rebinding page can't read us.
// With a login set the browser won't hand it to a stranger anyway, so it steps aside
func (s *Server) guard(next http.Handler) http.Handler {
	if s.cfg.Auth != "" || s.hosts["*"] {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		host = strings.Trim(host, "[]")
		if net.ParseIP(host) == nil && !s.hosts[host] && !strings.HasSuffix(host, ".localhost") {
			slog.Warn("turned away an unknown hostname", "host", host,
				"fix", "add it to --allowed-hosts / "+config.EnvName("allowed-hosts"))
			http.Error(w, fmt.Sprintf("peng-top doesn't know %q. Add it with --allowed-hosts", host),
				http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) auth(next http.Handler) http.Handler {
	if s.cfg.Auth == "" {
		return next
	}
	want := []byte(s.cfg.Auth)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		got := []byte(user + ":" + pass)
		if r.URL.Path != "/healthz" && subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="peng-top", charset="UTF-8"`)
			http.Error(w, "login needed", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data:; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func static(ui fs.FS) http.Handler {
	if _, err := fs.Stat(ui, "index.html"); err != nil {
		slog.Warn("no UI in this build, only the API's here", "fix", "make build")
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "This build has no UI. The API's at /api/snapshot", http.StatusNotFound)
		})
	}
	files := http.FileServerFS(ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Debug("couldn't send a response", "err", err)
	}
}
