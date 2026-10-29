package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/NotAFlightRisk/peng-top/internal/collect"
	"github.com/NotAFlightRisk/peng-top/internal/config"
	"github.com/NotAFlightRisk/peng-top/internal/logx"
	"github.com/NotAFlightRisk/peng-top/internal/server"
	"github.com/NotAFlightRisk/peng-top/web"
)

var version = "dev"

const banner = `                         _
 _ __  ___ _ _  __ _ ___| |_ ___ _ __
| '_ \/ -_) ' \/ _' |___|  _/ _ \ '_ \
| .__/\___|_||_\__, |    \__\___/ .__/
|_|            |___/            |_|`

func main() {
	cfg, err := config.Parse(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	logx.Setup(cfg.LogLevel, !cfg.NoColour)
	if err != nil {
		slog.Error("that config won't fly", "why", err)
		os.Exit(2)
	}
	switch {
	case cfg.Version:
		fmt.Println("peng-top", version)
	case cfg.Healthcheck:
		os.Exit(healthcheck(cfg))
	default:
		if err := run(cfg); err != nil {
			slog.Error("peng-top fell over", "why", err)
			os.Exit(1)
		}
	}
}

func run(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("can't listen on %s (is something else using the port?): %w", cfg.Addr(), err)
	}
	env := collect.Detect(cfg.HostRoot)
	col := collect.New(env, cfg.Interval, cfg.History)
	go col.Run(ctx)

	srv := &http.Server{
		Handler:           server.New(cfg, col, collect.ReadInfo(ctx, env), version).Handler(web.UI()),
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()

	url := "http://" + displayHost(cfg.Host) + ":" + strconv.Itoa(cfg.Port)
	printBanner(cfg, env, url)
	if cfg.Open {
		openBrowser(url)
	}

	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}
	stop()
	slog.Info("shutting down, ctrl+c again if you're in a hurry")
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		return fmt.Errorf("didn't stop cleanly: %w", err)
	}
	slog.Info("bye 👋")
	return nil
}

func printBanner(cfg config.Config, env collect.Env, url string) {
	paint := func(code, s string) string {
		if cfg.NoColour {
			return s
		}
		return "\033[" + code + "m" + s + "\033[0m"
	}
	fmt.Fprintf(os.Stderr, "\n%s %s\n\n", paint("36", banner), paint("2", version))
	fmt.Fprintf(os.Stderr, "  %s %s\n", paint("1;32", "➜"), paint("1", url))
	if cfg.Host == "0.0.0.0" || cfg.Host == "::" {
		for _, ip := range lanAddrs() {
			fmt.Fprintf(os.Stderr, "  %s %s\n", paint("2", "➜"), "http://"+net.JoinHostPort(ip,
				strconv.Itoa(cfg.Port)))
		}
	}
	fmt.Fprintf(os.Stderr, "  %s\n\n", paint("2", fmt.Sprintf("watching the %s every %s",
		strings.ReplaceAll(string(env.Mode), "-", " "), cfg.Interval)))
}

func displayHost(host string) string {
	switch host {
	case "", "0.0.0.0", "::", "127.0.0.1", "::1":
		return "localhost"
	}
	if strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}

func lanAddrs() []string {
	addrs, _ := net.InterfaceAddrs()
	var out []string
	for _, a := range addrs {
		if ip, ok := a.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
			out = append(out, ip.IP.String())
		}
	}
	return out[:min(len(out), 3)]
}

func openBrowser(url string) {
	cmd := map[string][]string{
		"darwin":  {"open", url},
		"windows": {"rundll32", "url.dll,FileProtocolHandler", url},
	}[runtime.GOOS]
	if cmd == nil {
		cmd = []string{"xdg-open", url}
	}
	if err := exec.Command(cmd[0], cmd[1:]...).Start(); err != nil {
		slog.Warn("couldn't open a browser, the link's above", "why", err)
	}
}

func healthcheck(cfg config.Config) int {
	host := cfg.Host
	if host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	client := http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, "http://"+net.JoinHostPort(host,
		strconv.Itoa(cfg.Port))+"/healthz", nil)
	res, err := client.Do(req)
	if err != nil {
		slog.Error("not answering", "why", err)
		return 1
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		slog.Error("answered, but not happily", "status", res.Status)
		return 1
	}
	return 0
}
