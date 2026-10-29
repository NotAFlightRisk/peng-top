package collect

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// burn keeps a cpu busy for a bit so there's some time on the clock
func burn(d time.Duration) (x float64) {
	for end := time.Now().Add(d); time.Now().Before(end); {
		for i := range 10_000 {
			x += math.Sqrt(float64(i))
		}
	}
	return x
}

func TestListAgreesWithGopsutilAboutUs(t *testing.T) {
	burn(300 * time.Millisecond)
	ctx := context.Background()
	oracle := &process.Process{Pid: int32(os.Getpid())}
	before, err := oracle.TimesWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	r := &procReader{}
	took := time.Now()
	raws, err := r.list(ctx)
	t.Logf("read %d processes in %s", len(raws), time.Since(took))
	if err != nil {
		t.Fatal(err)
	}

	seen := map[int32]bool{}
	var us rawProc
	for _, raw := range raws {
		if seen[raw.pid] {
			t.Fatalf("pid %d turned up twice", raw.pid)
		}
		seen[raw.pid] = true
		if raw.pid == int32(os.Getpid()) {
			us = raw
		}
	}
	if us.pid == 0 {
		t.Fatal("couldn't find ourselves")
	}
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := len(raws) - len(pids); diff < -10-len(pids)/4 || diff > 10+len(pids)/4 {
		t.Errorf("found %d processes, gopsutil found %d", len(raws), len(pids))
	}

	after, err := oracle.TimesWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ppid, _ := oracle.PpidWithContext(ctx)
	started, _ := oracle.CreateTimeWithContext(ctx)
	mem, err := oracle.MemoryInfoWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if us.ppid != ppid {
		t.Errorf("ppid %d, gopsutil says %d", us.ppid, ppid)
	}
	if d := us.started - started; d < -1000 || d > 1000 {
		t.Errorf("started %d, gopsutil says %d", us.started, started)
	}
	// clocks tick every 10-16ms, so allow a couple of ticks either side
	low, high := before.User+before.System-0.05, after.User+after.System+0.05
	if us.cpuSecs < 0.2 || us.cpuSecs < low || us.cpuSecs > high {
		t.Errorf("cpu %.3fs, gopsutil says between %.3fs and %.3fs", us.cpuSecs, low, high)
	}
	if us.rss < mem.RSS/2 || us.rss > mem.RSS*2 {
		t.Errorf("rss %d, gopsutil says %d", us.rss, mem.RSS)
	}
	if us.threads < 1 || us.state != "running" {
		t.Errorf("threads %d, state %q", us.threads, us.state)
	}
	if name := r.describe(ctx, us).name; !strings.HasPrefix(name, "collect.test") {
		t.Errorf("called %q", name)
	}
}

func BenchmarkList(b *testing.B) {
	r := &procReader{}
	for b.Loop() {
		if _, err := r.list(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
