// Package config reads flags, falling back to PENGTOP_* env vars, then defaults
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host         string
	Port         int
	Interval     time.Duration
	History      int
	Auth         string
	AllowedHosts []string
	HostRoot     string
	LogLevel     string
	NoColour     bool
	Open         bool
	Version      bool
	Healthcheck  bool
}

// Addr is what the server listens on
func (c Config) Addr() string { return net.JoinHostPort(c.Host, strconv.Itoa(c.Port)) }

// EnvName is the env var that sets a flag, e.g. allowed-hosts is PENGTOP_ALLOWED_HOSTS
func EnvName(flagName string) string {
	return "PENGTOP_" + strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}

var flagOnly = map[string]bool{"version": true, "healthcheck": true}

// Parse reads args (without the program name), with the environment filling the gaps
func Parse(args []string, usage io.Writer) (Config, error) {
	var c Config
	var allowed string
	fs := flag.NewFlagSet("peng-top", flag.ContinueOnError)
	fs.SetOutput(usage)
	fs.StringVar(&c.Host, "host", "127.0.0.1", "address to listen on")
	fs.IntVar(&c.Port, "port", 7340, "port to listen on")
	fs.DurationVar(&c.Interval, "interval", 2*time.Second, "how often to take a reading")
	fs.IntVar(&c.History, "history", 150, "readings to keep for the graphs")
	fs.StringVar(&c.Auth, "auth", "", "require a login, as user:password")
	fs.StringVar(&allowed, "allowed-hosts", "",
		"extra hostnames to answer to, comma separated, or * for any")
	fs.StringVar(&c.HostRoot, "host-root", "",
		"where the host's / is mounted, when running in a container")
	fs.StringVar(&c.LogLevel, "log-level", "info", "debug, info, warn or error")
	fs.BoolVar(&c.NoColour, "no-color", os.Getenv("NO_COLOR") != "", "plain log output")
	fs.BoolVar(&c.Open, "open", false, "open it in your browser once it's up")
	fs.BoolVar(&c.Version, "version", false, "print the version and exit")
	fs.BoolVar(&c.Healthcheck, "healthcheck", false, "check a running instance answers, then exit")

	var errs []error
	fs.VisitAll(func(f *flag.Flag) {
		v, ok := os.LookupEnv(EnvName(f.Name))
		if ok && !flagOnly[f.Name] {
			if err := fs.Set(f.Name, v); err != nil {
				errs = append(errs, fmt.Errorf("%s=%q: %w", EnvName(f.Name), v, err))
			}
		}
	})
	if err := fs.Parse(args); err != nil {
		return c, err
	}

	for _, h := range strings.Split(allowed, ",") {
		if h = strings.TrimSpace(h); h != "" {
			c.AllowedHosts = append(c.AllowedHosts, strings.ToLower(h))
		}
	}
	return c, errors.Join(append(errs, c.validate()...)...)
}

func (c Config) validate() []error {
	var errs []error
	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port %d isn't a port", c.Port))
	}
	if c.Interval < 500*time.Millisecond {
		errs = append(errs, fmt.Errorf("interval %s is a bit keen, 500ms is the floor", c.Interval))
	}
	if c.History < 1 {
		errs = append(errs, errors.New("history needs to be at least 1"))
	}
	if c.Auth != "" && !strings.Contains(c.Auth, ":") {
		errs = append(errs, errors.New("auth wants user:password"))
	}
	return errs
}
