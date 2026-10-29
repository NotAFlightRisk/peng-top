package config

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestFlagsBeatEnvAndEnvBeatsDefaults(t *testing.T) {
	t.Setenv("PENGTOP_PORT", "9000")
	t.Setenv("PENGTOP_INTERVAL", "5s")
	t.Setenv("PENGTOP_ALLOWED_HOSTS", " Box.Local , ,nas.")
	c, err := Parse([]string{"--port", "9100"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 9100 || c.Interval != 5*time.Second || c.Host != "127.0.0.1" {
		t.Fatalf("got %+v", c)
	}
	if strings.Join(c.AllowedHosts, ",") != "box.local,nas" {
		t.Fatalf("allowed hosts %q", c.AllowedHosts)
	}
}

func TestBadEnvIsReportedNotIgnored(t *testing.T) {
	t.Setenv("PENGTOP_PORT", "lots")
	_, err := Parse(nil, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "PENGTOP_PORT") {
		t.Fatalf("got %v", err)
	}
}

func TestNonsenseValuesAreRejected(t *testing.T) {
	for _, args := range [][]string{
		{"--port", "0"},
		{"--interval", "100ms"},
		{"--history", "0"},
		{"--auth", "nopassword"},
	} {
		if _, err := Parse(args, io.Discard); err == nil {
			t.Errorf("%v was let through", args)
		}
	}
}
