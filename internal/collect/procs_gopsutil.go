//go:build !linux

package collect

import (
	"context"

	"github.com/shirou/gopsutil/v4/process"
)

func (r *procReader) describe(ctx context.Context, raw rawProc) about {
	p := &process.Process{Pid: raw.pid}
	a := about{name: raw.name}
	if a.name == "" {
		a.name, _ = p.NameWithContext(ctx)
	}
	a.user, _ = p.UsernameWithContext(ctx)
	cmd, _ := p.CmdlineWithContext(ctx)
	a.cmd = clip(cmd)
	return a
}
