package collect

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/sensors"
)

type Sensor struct {
	Name string  `json:"name"`
	Temp float64 `json:"temp"`
	High float64 `json:"high,omitempty"`
	Crit float64 `json:"crit,omitempty"`
}

func readSensors(ctx context.Context, s *Snapshot, _ float64) error {
	temps, err := sensors.TemperaturesWithContext(ctx)
	var out []Sensor
	for _, t := range temps {
		if t.Temperature <= 0 || t.Temperature > 150 {
			continue
		}
		out = append(out, Sensor{Name: t.SensorKey, Temp: t.Temperature, High: t.High, Crit: t.Critical})
	}
	if len(out) == 0 {
		return errors.Join(err, errors.New("no temperature sensors exposed"))
	}
	if err != nil {
		slog.Debug("some sensors wouldn't read", "why", err)
	}
	slices.SortFunc(out, func(a, b Sensor) int { return strings.Compare(a.Name, b.Name) })
	s.Sensors = out
	return nil
}
