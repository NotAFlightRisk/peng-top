//go:build !linux && !windows && !darwin

package collect

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
)

type procOS struct{}

func (r *procReader) list(ctx context.Context) ([]rawProc, error) {
	all, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]rawProc, 0, len(all))
	for _, p := range all {
		times, err := p.TimesWithContext(ctx)
		if err != nil {
			continue
		}
		raw := rawProc{pid: p.Pid, cpuSecs: times.User + times.System}
		raw.started, _ = p.CreateTimeWithContext(ctx)
		raw.ppid, _ = p.PpidWithContext(ctx)
		raw.threads, _ = p.NumThreadsWithContext(ctx)
		if m, err := p.MemoryInfoWithContext(ctx); err == nil {
			raw.rss = m.RSS
		}
		if states, err := p.StatusWithContext(ctx); err == nil && len(states) > 0 {
			raw.state = states[0]
		}
		out = append(out, raw)
	}
	return out, nil
}
