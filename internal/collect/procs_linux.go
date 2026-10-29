package collect

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
)

// clock ticks a second, which is USER_HZ on every arch we ship
const ticks = 100

var states = map[byte]string{
	'R': "running", 'S': "sleep", 'D': "blocked", 'Z': "zombie",
	'T': "stop", 't': "stop", 'I': "idle", 'X': "dead",
}

// list reads one stat file per process
func (r *procReader) list(ctx context.Context) ([]rawProc, error) {
	root := r.env.HostPath("/proc")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	boot, err := host.BootTimeWithContext(ctx)
	if err != nil {
		return nil, err
	}
	page := uint64(os.Getpagesize())
	out := make([]rawProc, 0, len(entries))
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		stat, err := os.ReadFile(filepath.Join(root, e.Name(), "stat"))
		if err != nil {
			continue
		}
		raw, ok := parseStat(stat, int64(boot), page)
		if !ok {
			continue
		}
		out = append(out, raw)
	}
	return out, nil
}

// parseStat reads one /proc/<pid>/stat line
func parseStat(stat []byte, boot int64, page uint64) (rawProc, bool) {
	// fields start after the last bracket, as comm can hold spaces and brackets
	start, end := bytes.IndexByte(stat, ' '), bytes.LastIndexByte(stat, ')')
	if start < 0 || end < start {
		return rawProc{}, false
	}
	pid, err := strconv.ParseInt(string(stat[:start]), 10, 32)
	f := strings.Fields(string(stat[end+1:]))
	if err != nil || len(f) < 22 {
		return rawProc{}, false
	}
	num := func(i int) uint64 { n, _ := strconv.ParseUint(f[i], 10, 64); return n }
	raw := rawProc{
		pid:     int32(pid),
		ppid:    int32(num(1)),
		cpuSecs: float64(num(11)+num(12)) / ticks,
		threads: int32(num(17)),
		started: boot*1000 + int64(num(19))*1000/ticks,
		rss:     num(21) * page,
		state:   states[f[0][0]],
	}
	raw.kernel = raw.pid == 2 || raw.ppid == 2
	return raw, true
}

func (r *procReader) describe(_ context.Context, pid int32) about {
	dir := filepath.Join(r.env.HostPath("/proc"), strconv.Itoa(int(pid)))
	var a about
	status, _ := os.ReadFile(filepath.Join(dir, "status"))
	for _, line := range strings.Split(string(status), "\n") {
		key, value, _ := strings.Cut(line, ":")
		fields := strings.Fields(value)
		switch {
		case key == "Name" && len(fields) > 0:
			a.name = strings.TrimSpace(value)
		case key == "Uid" && len(fields) > 0:
			if uid, err := strconv.ParseUint(fields[0], 10, 32); err == nil {
				a.user = r.username(uint32(uid))
			}
		}
	}
	cmdline, _ := os.ReadFile(filepath.Join(dir, "cmdline"))
	args := strings.Split(strings.TrimRight(string(cmdline), "\x00"), "\x00")
	a.cmd = clip(strings.Join(args, " "))
	// names stop at 15 characters, so the full one comes from argv when it fits
	if exe := filepath.Base(args[0]); len(a.name) == 15 && strings.HasPrefix(exe, a.name) {
		a.name = exe
	}
	return a
}
