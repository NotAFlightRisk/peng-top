package collect

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/NotAFlightRisk/peng-top/internal/logx"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
)

type CPU struct {
	Total  pct       `json:"total"`
	User   pct       `json:"user"`
	System pct       `json:"system"`
	IOWait pct       `json:"iowait"`
	Steal  pct       `json:"steal"`
	Cores  []pct     `json:"cores"`
	Load   []float64 `json:"load,omitempty"`
	MHz    float64   `json:"mhz,omitempty"`
	Quota  float64   `json:"quota,omitempty"`
}

type cpuReader struct {
	env   Env
	total cpu.TimesStat
	cores []cpu.TimesStat
	own   map[string]uint64
}

func (r *cpuReader) read(ctx context.Context, s *Snapshot, elapsed float64) error {
	totals, err := times(ctx, false)
	if err == nil && len(totals) == 0 {
		err = errors.New("no cpu times")
	}
	if err != nil {
		return err
	}
	cores, err := times(ctx, true)
	if err != nil {
		return err
	}

	c := &CPU{Cores: make([]pct, len(cores))}
	now := totals[0]
	if spent := all(now) - all(r.total); spent > 0 {
		share := func(a, b float64) pct { return clampPct((a - b) / spent * 100) }
		c.Total = share(busy(now), busy(r.total))
		c.User = share(now.User+now.Nice, r.total.User+r.total.Nice)
		c.System = share(now.System+now.Irq+now.Softirq, r.total.System+r.total.Irq+r.total.Softirq)
		c.IOWait = share(now.Iowait, r.total.Iowait)
		c.Steal = share(now.Steal, r.total.Steal)
	}
	for i, core := range cores {
		if i < len(r.cores) {
			c.Cores[i] = usage(core, r.cores[i])
		}
	}
	r.total, r.cores = now, cores

	if avg, err := load.AvgWithContext(ctx); !logx.Check("load average", err) {
		c.Load = []float64{avg.Load1, avg.Load5, avg.Load15}
	}
	if runtime.GOOS == "linux" {
		c.MHz = r.mhz(cores)
		if r.env.Mode == ModeContainer {
			c.Quota = cgroupCPUQuota()
			r.ownShare(c, elapsed)
		}
	}
	s.CPU = c
	return nil
}

// times is gopsutil's, bar windows' interrupt time, as its kernel time already counts it
func times(ctx context.Context, perCore bool) ([]cpu.TimesStat, error) {
	all, err := cpu.TimesWithContext(ctx, perCore)
	if runtime.GOOS == "windows" {
		for i := range all {
			all[i].Irq = 0
		}
	}
	return all, err
}

// mhz averages every core's clock right now, from cpufreq or else /proc/cpuinfo
func (r *cpuReader) mhz(cores []cpu.TimesStat) float64 {
	var sum, n float64
	add := func(text string, perMHz float64) {
		if v, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err == nil && v > 0 {
			sum, n = sum+v/perMHz, n+1
		}
	}
	for i, core := range cores {
		khz, err := os.ReadFile(r.env.HostPath("/sys/devices/system/cpu/" + core.CPU +
			"/cpufreq/scaling_cur_freq"))
		if err != nil && i == 0 {
			break
		}
		add(string(khz), 1000)
	}
	if n == 0 {
		info, _ := os.ReadFile(r.env.HostPath("/proc/cpuinfo"))
		for _, line := range strings.Split(string(info), "\n") {
			if key, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(key) == "cpu MHz" {
				add(value, 1)
			}
		}
	}
	return sum / max(n, 1)
}

// ownShare swaps the host's totals for the container's own, out of the cores it's allowed
func (r *cpuReader) ownShare(c *CPU, elapsed float64) {
	now := cgroupStats("cpu.stat")
	if now["usage_usec"] == 0 {
		return
	}
	before := r.own
	r.own = now
	cores := float64(runtime.NumCPU())
	if c.Quota > 0 {
		cores = min(cores, c.Quota)
	}
	share := func(key string) pct {
		return clampPct(float64(rate(now[key], before[key], elapsed)) / 1e4 / cores)
	}
	c.Total, c.User, c.System = share("usage_usec"), share("user_usec"), share("system_usec")
}

// busy is non-idle time, minus guest which user already includes
func busy(t cpu.TimesStat) float64 {
	return t.User + t.Nice + t.System + t.Irq + t.Softirq + t.Steal
}

func all(t cpu.TimesStat) float64 { return busy(t) + t.Idle + t.Iowait }

func usage(now, before cpu.TimesStat) pct {
	spent := all(now) - all(before)
	if spent <= 0 {
		return 0
	}
	return clampPct((busy(now) - busy(before)) / spent * 100)
}

func clampPct(v float64) pct { return pct(min(max(v, 0), 100)) }

// cgroupCPUQuota is how many cores a cgroup v2 container is allowed, 0 if unlimited
func cgroupCPUQuota() float64 {
	raw, err := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(raw))
	if len(fields) != 2 || fields[0] == "max" {
		return 0
	}
	quota, err1 := strconv.ParseFloat(fields[0], 64)
	period, err2 := strconv.ParseFloat(fields[1], 64)
	if err1 != nil || err2 != nil || period == 0 {
		return 0
	}
	return quota / period
}
