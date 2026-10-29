package collect

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

func TestUsageIgnoresGuestTimeAndCountsIOWaitAsIdle(t *testing.T) {
	before := cpu.TimesStat{User: 100, Idle: 100, Iowait: 0, Guest: 50}
	now := cpu.TimesStat{User: 130, Idle: 160, Iowait: 10, Guest: 80}
	if got := usage(now, before); got != 30 {
		t.Fatalf("got %v%%, want 30%%", got)
	}
}

func TestClockSpeedAveragesEachCoresCurrentClock(t *testing.T) {
	root := t.TempDir()
	write := func(path, text string) {
		path = filepath.Join(root, path)
		os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := &cpuReader{env: Env{Root: root}}
	cores := []cpu.TimesStat{{CPU: "cpu0"}, {CPU: "cpu1"}, {CPU: "cpu2"}}
	write("proc/cpuinfo", "cpu MHz\t\t: 1000.000\nflags\t\t: fpu\ncpu MHz\t\t: 3000.000\n")
	if got := r.mhz(cores); got != 2000 {
		t.Errorf("got %v MHz from cpuinfo, want 2000", got)
	}
	write("sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq", "800000\n")
	write("sys/devices/system/cpu/cpu2/cpufreq/scaling_cur_freq", "4000000\n")
	if got := r.mhz(cores); got != 2400 {
		t.Errorf("got %v MHz from cpufreq, want 2400", got)
	}
}

func TestRateNeverGoesNegativeWhenACounterResets(t *testing.T) {
	if got := rate(10, 500, 2); got != 0 {
		t.Fatalf("got %v", got)
	}
	if got := rate(500, 100, 2); got != 200 {
		t.Fatalf("got %v", got)
	}
}

func TestNumbersGoOutShort(t *testing.T) {
	out, _ := json.Marshal(struct {
		P pct
		R perSec
	}{12.3456, 1234.56})
	if string(out) != `{"P":12.3,"R":1235}` {
		t.Fatalf("got %s", out)
	}
}

func TestSubscribersGetHistoryThenNewReadings(t *testing.T) {
	c := &Collector{keep: 2, subs: map[chan []byte]struct{}{}}
	c.history = [][]byte{[]byte(`{"t":1}`), []byte(`{"t":2}`)}
	history, feed, cancel := c.Subscribe()
	defer cancel()
	if string(history) != `[{"t":1},{"t":2}]` {
		t.Fatalf("history was %s", history)
	}

	c.sources = nil
	c.sample(t.Context(), true)
	if got := <-feed; len(got) == 0 {
		t.Fatal("no reading came through")
	}
	if len(c.history) != 2 {
		t.Fatalf("history grew past its cap to %d", len(c.history))
	}
}

func TestAReadAfterAFailureCountsFromTheLastGoodOne(t *testing.T) {
	var got []float64
	fail := true
	flaky := func(_ context.Context, _ *Snapshot, elapsed float64) error {
		got = append(got, elapsed)
		if fail {
			return errors.New("not today")
		}
		return nil
	}
	c := &Collector{
		sources:  []source{{"flaky", flaky}},
		lastGood: map[string]time.Time{"flaky": time.Now().Add(-time.Minute)},
	}
	c.sample(t.Context(), false)
	fail = false
	c.sample(t.Context(), false)
	if got[1] < 60 {
		t.Fatalf("got %.1fs, but the last good read was a minute ago", got[1])
	}
}

func TestOnlyLoopbackNamesCountAsLoopback(t *testing.T) {
	for name, want := range map[string]bool{
		"lo": true, "lo0": true, "Local Area Connection": false, "eth0": false,
		"veth1a2b": true, "Loopback Pseudo-Interface 1": true,
	} {
		if got := isVirtual(name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

func TestAPanickingReaderBecomesAnError(t *testing.T) {
	err := safely(func() error { panic("boom") })
	if err == nil {
		t.Fatal("the panic got swallowed")
	}
}
