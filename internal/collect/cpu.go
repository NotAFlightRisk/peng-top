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
}

func (r *cpuReader) read(ctx context.Context, s *Snapshot, _ float64) error {
	totals, err := cpu.TimesWithContext(ctx, false)
	if err == nil && len(totals) == 0 {
		err = errors.New("no cpu times")
	}
	if err != nil {
		return err
	}
	cores, err := cpu.TimesWithContext(ctx, true)
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
		c.MHz = r.mhz(ctx)
		if r.env.Mode == ModeContainer {
			c.Quota = cgroupCPUQuota()
		}
	}
	s.CPU = c
	return nil
}

func (r *cpuReader) mhz(ctx context.Context) float64 {
	infos, err := cpu.InfoWithContext(ctx)
	if logx.Check("cpu frequency", err) || len(infos) == 0 {
		return 0
	}
	var sum float64
	for _, info := range infos {
		sum += info.Mhz
	}
	return sum / float64(len(infos))
}

// busy leaves out guest time, since linux already counts it inside user
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
