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

type memReader struct{ env Env }

func (r *memReader) read(ctx context.Context, s *Snapshot, _ float64) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	m := &Mem{
		Total: vm.Total, Used: vm.Used, Available: vm.Available,
		Cached: vm.Cached, Buffers: vm.Buffers, Free: vm.Free,
	}
	if swap, err := mem.SwapMemoryWithContext(ctx); !logx.Check("swap", err) {
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
	m.Limit, m.Used = limit, used
	m.Available = limit - min(used, limit)
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
