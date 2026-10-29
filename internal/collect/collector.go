// Package collect takes readings on a timer and hands them out to whoever's listening
package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/NotAFlightRisk/peng-top/internal/logx"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Snapshot is one reading of everything. A nil section means it couldn't be read
type Snapshot struct {
	Time    int64    `json:"t"`
	CPU     *CPU     `json:"cpu,omitempty"`
	Mem     *Mem     `json:"mem,omitempty"`
	Disks   *Disks   `json:"disks,omitempty"`
	Net     []NIC    `json:"net,omitempty"`
	Sensors []Sensor `json:"sensors,omitempty"`
	Procs   *Procs   `json:"procs,omitempty"`
}

// source is one reader, handed the seconds since it last read cleanly
type source struct {
	area string
	read func(ctx context.Context, s *Snapshot, elapsed float64) error
}

type Collector struct {
	interval time.Duration
	sources  []source
	lastGood map[string]time.Time
	keep     int

	mu      sync.RWMutex
	latest  []byte
	history [][]byte
	subs    map[chan []byte]struct{}
}

func New(env Env, interval time.Duration, keep int) *Collector {
	return &Collector{
		interval: interval,
		keep:     keep,
		subs:     map[chan []byte]struct{}{},
		lastGood: map[string]time.Time{},
		sources: []source{
			{"cpu", (&cpuReader{env: env}).read},
			{"memory", (&memReader{env: env, swap: every[*mem.SwapMemoryStat]{period: 10 * time.Second}}).read},
			{"disk usage", (&mountReader{env: env, parts: every[[]disk.PartitionStat]{period: 30 * time.Second}}).read},
			{"disk io", (&diskIOReader{env: env}).read},
			{"network", (&netReader{ifaces: every[[]net.InterfaceStat]{period: 10 * time.Second}}).read},
			{"sensors", (&sensorReader{}).read},
			{"processes", (&procReader{env: env}).read},
		},
	}
}

// Run takes readings until ctx is done. The first one only primes the counters
func (c *Collector) Run(ctx context.Context) {
	c.sample(ctx, false)
	tick := time.NewTicker(c.interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			c.sample(ctx, true)
		}
	}
}

func (c *Collector) sample(ctx context.Context, publish bool) {
	started := time.Now()
	s := Snapshot{Time: started.UnixMilli()}
	for _, src := range c.sources {
		now := time.Now()
		var elapsed float64
		if last, ok := c.lastGood[src.area]; ok {
			elapsed = now.Sub(last).Seconds()
		}
		if !logx.Check(src.area, safely(func() error { return src.read(ctx, &s, elapsed) })) {
			c.lastGood[src.area] = now
		}
	}
	if !publish {
		return
	}

	full, err := json.Marshal(s)
	s.Procs = nil
	light, err2 := json.Marshal(s)
	if err != nil || err2 != nil {
		slog.Error("couldn't encode a reading, which really shouldn't happen", "err", err)
		return
	}

	c.mu.Lock()
	c.latest = full
	c.history = append(c.history, light)
	if len(c.history) > c.keep {
		c.history = c.history[len(c.history)-c.keep:]
	}
	for ch := range c.subs {
		select {
		case ch <- full:
		default:
		}
	}
	c.mu.Unlock()
	slog.Debug("reading taken", "took", time.Since(started).Round(time.Millisecond),
		"bytes", len(full))
}

// safely turns a panic in one reader into an error
func safely(read func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fell over reading it: %v", r)
		}
	}()
	return read()
}

// Latest is the newest full reading, or nil before the first one
func (c *Collector) Latest() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}

// Subscribe hands back the history so far as a JSON array, plus a feed of new readings
func (c *Collector) Subscribe() (history []byte, feed <-chan []byte, cancel func()) {
	ch := make(chan []byte, 2)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subs[ch] = struct{}{}
	history = append(append([]byte("["), bytes.Join(c.history, []byte(","))...), ']')
	return history, ch, func() {
		c.mu.Lock()
		delete(c.subs, ch)
		c.mu.Unlock()
	}
}

// every reruns a slow read at most once a period, handing back the last result in between
type every[T any] struct {
	period time.Duration
	next   time.Time
	val    T
	err    error
}

func (e *every[T]) get(read func() (T, error)) (T, error) {
	if now := time.Now(); !now.Before(e.next) {
		e.val, e.err = read()
		e.next = now.Add(e.period)
	}
	return e.val, e.err
}

// stale makes the next get read afresh
func (e *every[T]) stale() { e.next = time.Time{} }

// pct is a percentage, sent with one decimal place
type pct float64

func (p pct) MarshalJSON() ([]byte, error) {
	return strconv.AppendFloat(nil, math.Round(float64(p)*10)/10, 'f', -1, 64), nil
}

// perSec is a rate, sent as a whole number
type perSec float64

func (r perSec) MarshalJSON() ([]byte, error) {
	return strconv.AppendFloat(nil, math.Round(float64(r)), 'f', 0, 64), nil
}

func rate(now, before uint64, elapsed float64) perSec {
	if elapsed <= 0 || now < before {
		return 0
	}
	return perSec(float64(now-before) / elapsed)
}
