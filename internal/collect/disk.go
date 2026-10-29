package collect

import (
	"context"
	"errors"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

type Disks struct {
	Mounts []Mount  `json:"mounts,omitempty"`
	IO     []DiskIO `json:"io,omitempty"`
}

type Mount struct {
	Path   string `json:"path"`
	Device string `json:"device"`
	FS     string `json:"fs"`
	Total  uint64 `json:"total"`
	Used   uint64 `json:"used"`
	Free   uint64 `json:"free"`
	Inodes pct    `json:"inodes,omitempty"`
}

type DiskIO struct {
	Name       string `json:"name"`
	Read       perSec `json:"read"`
	Write      perSec `json:"write"`
	ReadOps    perSec `json:"readOps"`
	WriteOps   perSec `json:"writeOps"`
	Busy       *pct   `json:"busy,omitempty"`
	ReadTotal  uint64 `json:"readTotal"`
	WriteTotal uint64 `json:"writeTotal"`
}

// snaps, squashfs images and friends that'd only clutter the list
var skipFS = []string{"squashfs", "overlay", "tmpfs", "devtmpfs", "iso9660", "nsfs", "tracefs"}

type mountReader struct{ env Env }

func (r *mountReader) read(ctx context.Context, s *Snapshot, _ float64) error {
	parts, err := disk.PartitionsWithContext(ctx, false)
	if len(parts) == 0 {
		return errors.Join(err, errors.New("no filesystems found"))
	}
	seen := map[string]int{}
	var mounts []Mount
	for _, p := range parts {
		if slices.Contains(skipFS, p.Fstype) || macSystemVolume(p.Mountpoint) {
			continue
		}
		u, err := disk.UsageWithContext(ctx, r.env.HostPath(p.Mountpoint))
		if err != nil || u.Total == 0 {
			continue
		}
		m := Mount{
			Path: p.Mountpoint, Device: p.Device, FS: p.Fstype,
			Total: u.Total, Used: u.Used, Free: u.Free, Inodes: pct(u.InodesUsedPercent),
		}
		if i, dup := seen[p.Device]; dup {
			if len(m.Path) < len(mounts[i].Path) {
				mounts[i] = m
			}
			continue
		}
		seen[p.Device] = len(mounts)
		mounts = append(mounts, m)
	}
	slices.SortFunc(mounts, func(a, b Mount) int { return strings.Compare(a.Path, b.Path) })
	if s.Disks == nil {
		s.Disks = &Disks{}
	}
	s.Disks.Mounts = mounts
	return nil
}

// macOS mounts half a dozen system volumes nobody wants to see, Data is the one that fills up
func macSystemVolume(path string) bool {
	return strings.HasPrefix(path, "/System/Volumes/") && path != "/System/Volumes/Data"
}

type diskIOReader struct {
	env  Env
	prev map[string]disk.IOCountersStat
}

func (r *diskIOReader) read(ctx context.Context, s *Snapshot, elapsed float64) error {
	now, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return err
	}
	var out []DiskIO
	for name, c := range now {
		if !r.wholeDisk(name) {
			continue
		}
		before, ok := r.prev[name]
		if !ok {
			before = c
		}
		out = append(out, DiskIO{
			Name:       name,
			Read:       rate(c.ReadBytes, before.ReadBytes, elapsed),
			Write:      rate(c.WriteBytes, before.WriteBytes, elapsed),
			ReadOps:    rate(c.ReadCount, before.ReadCount, elapsed),
			WriteOps:   rate(c.WriteCount, before.WriteCount, elapsed),
			Busy:       busyShare(c.IoTime, before.IoTime, elapsed),
			ReadTotal:  c.ReadBytes,
			WriteTotal: c.WriteBytes,
		})
	}
	r.prev = now
	slices.SortFunc(out, func(a, b DiskIO) int { return strings.Compare(a.Name, b.Name) })
	if s.Disks == nil {
		s.Disks = &Disks{}
	}
	s.Disks.IO = out
	return nil
}

// busyShare only means anything on linux, where io_ticks is wall time the disk spent busy
func busyShare(now, before uint64, elapsed float64) *pct {
	if runtime.GOOS != "linux" {
		return nil
	}
	busy := clampPct(float64(rate(now, before, elapsed)) / 10)
	return &busy
}

// wholeDisk drops partitions and loop/ram devices on linux, so a disk isn't counted twice
func (r *diskIOReader) wholeDisk(name string) bool {
	if runtime.GOOS != "linux" {
		return true
	}
	for _, prefix := range []string{"loop", "ram", "zram", "sr", "fd"} {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	_, err := os.Stat(r.env.HostPath("/sys/block/" + name))
	return err == nil
}
