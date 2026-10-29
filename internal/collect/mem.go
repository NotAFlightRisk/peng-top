package collect

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/NotAFlightRisk/peng-top/internal/logx"
	"github.com/shirou/gopsutil/v4/mem"
)

type Mem struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Available uint64 `json:"available"`
	Cached    uint64 `json:"cached"`
	Buffers   uint64 `json:"buffers"`
	Free      uint64 `json:"free"`
	SwapTotal uint64 `json:"swapTotal"`
	SwapUsed  uint64 `json:"swapUsed"`
	Limit     uint64 `json:"limit,omitempty"`
}

type memReader struct {
	env  Env
	swap every[*mem.SwapMemoryStat]
}

func (r *memReader) read(ctx context.Context, s *Snapshot, _ float64) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	m := &Mem{
		Total: vm.Total, Used: vm.Used, Available: vm.Available,
		Cached: vm.Cached, Buffers: vm.Buffers, Free: vm.Free,
	}
	swap, err := r.swap.get(func() (*mem.SwapMemoryStat, error) { return mem.SwapMemoryWithContext(ctx) })
	if !logx.Check("swap", err) {
		m.SwapTotal, m.SwapUsed = swap.Total, swap.Used
	}
	if r.env.Mode == ModeContainer {
		applyCgroup(m)
	}
	s.Mem = m
	return nil
}

// applyCgroup swaps in a cgroup v2 container's own limit and usage, when it has one
func applyCgroup(m *Mem) {
	limit, err1 := readCgroupNumber("memory.max")
	used, err2 := readCgroupNumber("memory.current")
	if err1 != nil || err2 != nil || limit == 0 || limit >= m.Total {
		return
	}
	stat := cgroupStats("memory.stat")
	// used leaves out inactive page cache, the same as docker stats
	used -= min(stat["inactive_file"], used)
	m.Limit, m.Used = limit, used
	m.Available = limit - min(used, limit)
	m.Free, m.Buffers, m.Cached = m.Available, 0, stat["file"]
	if swap, err := readCgroupNumber("memory.swap.current"); err == nil {
		m.SwapUsed = swap
		if allowed, _ := readCgroupNumber("memory.swap.max"); allowed > 0 {
			m.SwapTotal = min(allowed, m.SwapTotal)
		}
	}
}

// cgroupStats reads a cgroup file of "key value" lines, like memory.stat
func cgroupStats(name string) map[string]uint64 {
	raw, _ := os.ReadFile("/sys/fs/cgroup/" + name)
	stats := map[string]uint64{}
	for _, line := range strings.Split(string(raw), "\n") {
		if key, value, ok := strings.Cut(line, " "); ok {
			stats[key], _ = strconv.ParseUint(value, 10, 64)
		}
	}
	return stats
}

func readCgroupNumber(name string) (uint64, error) {
	raw, err := os.ReadFile("/sys/fs/cgroup/" + name)
	if err != nil {
		return 0, err
	}
	text := strings.TrimSpace(string(raw))
	if text == "max" {
		return 0, nil
	}
	return strconv.ParseUint(text, 10, 64)
}
