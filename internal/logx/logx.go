// Package logx is the one place peng-top logs from, and where metric failures get remembered
package logx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	reset = "\033[0m"
	dim   = "\033[2m"
	bold  = "\033[1m"
)

var levelStyle = map[slog.Level]struct{ tag, colour string }{
	slog.LevelDebug: {"DBG", "\033[35m"},
	slog.LevelInfo:  {"INF", "\033[36m"},
	slog.LevelWarn:  {"WRN", "\033[33m"},
	slog.LevelError: {"ERR", "\033[31m"},
}

type handler struct {
	mu     *sync.Mutex
	out    io.Writer
	level  slog.Level
	colour bool
	attrs  []slog.Attr
}

// Setup swaps the default slog logger for a short, coloured, human one
func Setup(level string, colour bool) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(&handler{mu: &sync.Mutex{}, out: os.Stderr, level: lvl, colour: colour}))
}

func (h *handler) Enabled(_ context.Context, l slog.Level) bool { return l >= h.level }

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

func (h *handler) WithGroup(string) slog.Handler { return h }

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	style := levelStyle[r.Level]
	if style.tag == "" {
		style = levelStyle[slog.LevelError]
	}
	var b strings.Builder
	b.WriteString(h.paint(dim, r.Time.Format(time.TimeOnly)) + " ")
	b.WriteString(h.paint(bold+style.colour, style.tag) + " " + r.Message)
	write := func(a slog.Attr) bool {
		b.WriteString(" " + h.paint(dim, a.Key+"=") + fmt.Sprint(a.Value.Any()))
		return true
	}
	for _, a := range h.attrs {
		write(a)
	}
	r.Attrs(write)
	b.WriteByte('\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *handler) paint(style, s string) string {
	if !h.colour {
		return s
	}
	return style + s + reset
}

var (
	issuesMu sync.RWMutex
	issues   = map[string]string{}
)

// Check records whether an area works, warning once per new failure, and says if err was one
func Check(area string, err error) bool {
	issuesMu.Lock()
	prev, failing := issues[area]
	if err == nil {
		delete(issues, area)
	} else {
		issues[area] = err.Error()
	}
	issuesMu.Unlock()

	switch {
	case err == nil && failing:
		slog.Info("back in business", "area", area)
	case err != nil && prev != err.Error():
		slog.Warn("can't read this here, skipping it", "area", area, "why", err)
	case err != nil:
		slog.Debug("still unavailable", "area", area, "why", err)
	}
	return err != nil
}

// Issues lists every area currently failing, with why
func Issues() map[string]string {
	issuesMu.RLock()
	defer issuesMu.RUnlock()
	out := make(map[string]string, len(issues))
	for k, v := range issues {
		out[k] = v
	}
	return out
}
