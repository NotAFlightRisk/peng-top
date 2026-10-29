package collect

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/sensors"
)

type Sensor struct {
	Name string  `json:"name"`
	Temp float64 `json:"temp"`
	High float64 `json:"high,omitempty"`
	Crit float64 `json:"crit,omitempty"`
}

// sensors are read every few seconds, and ever less often while they keep failing
const sensorEvery, sensorMaxWait = 5 * time.Second, 5 * time.Minute

type sensorReader struct {
	next time.Time
	wait time.Duration
	last []Sensor
	err  error
}

func (r *sensorReader) read(ctx context.Context, s *Snapshot, _ float64) error {
	if now := time.Now(); !now.Before(r.next) {
		r.last, r.err = readSensors(ctx)
		if r.err == nil {
			r.wait = sensorEvery
		} else {
			r.wait = min(max(r.wait*2, sensorEvery), sensorMaxWait)
		}
		r.next = now.Add(r.wait)
	}
	s.Sensors = r.last
	return r.err
}

func readSensors(ctx context.Context) ([]Sensor, error) {
	temps, err := sensors.TemperaturesWithContext(ctx)
	var out []Sensor
	for _, t := range temps {
		if t.Temperature <= 0 || t.Temperature > 150 {
			continue
		}
		out = append(out, Sensor{Name: t.SensorKey, Temp: t.Temperature, High: t.High, Crit: t.Critical})
	}
	if len(out) == 0 {
		return nil, errors.Join(err, errors.New("no temperature sensors exposed"))
	}
	if err != nil {
		slog.Debug("some sensors wouldn't read", "why", err)
	}
	slices.SortFunc(out, func(a, b Sensor) int { return strings.Compare(a.Name, b.Name) })
	return out, nil
}
