package collect

import (
	"context"
	"os"
	"runtime"
	"strings"

	"github.com/NotAFlightRisk/peng-top/internal/logx"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
)

// Info is everything that doesn't change between readings
type Info struct {
	Hostname string  `json:"hostname"`
	OS       string  `json:"os"`
	Platform string  `json:"platform"`
	Release  string  `json:"release"`
	Kernel   string  `json:"kernel"`
	Arch     string  `json:"arch"`
	Virt     string  `json:"virt,omitempty"`
	BootTime int64   `json:"bootTime"`
	CPUModel string  `json:"cpuModel"`
	Cores    int     `json:"cores"`
	Threads  int     `json:"threads"`
	MaxMHz   float64 `json:"maxMhz,omitempty"`
	Mode     Mode    `json:"mode"`
}

// ReadInfo is called once at startup, and anything it can't find is left blank
func ReadInfo(ctx context.Context, env Env) Info {
	info := Info{Mode: env.Mode, OS: runtime.GOOS, Arch: runtime.GOARCH}
	if h, err := host.InfoWithContext(ctx); !logx.Check("host info", err) {
		info.Hostname, info.Platform, info.Release = h.Hostname, h.Platform, h.PlatformVersion
		info.Kernel, info.BootTime = h.KernelVersion, int64(h.BootTime)*1000
		if h.KernelArch != "" {
			info.Arch = h.KernelArch
		}
		if h.VirtualizationRole == "guest" {
			info.Virt = h.VirtualizationSystem
		}
	}
	if name, err := os.ReadFile(env.HostPath("/etc/hostname")); err == nil && env.Root != "" {
		info.Hostname = strings.TrimSpace(string(name))
	}
	if cpus, err := cpu.InfoWithContext(ctx); !logx.Check("cpu info", err) && len(cpus) > 0 {
		info.CPUModel = cpus[0].ModelName
		info.MaxMHz = cpus[0].Mhz
	}
	info.Cores, _ = cpu.CountsWithContext(ctx, false)
	info.Threads, _ = cpu.CountsWithContext(ctx, true)
	return info
}
