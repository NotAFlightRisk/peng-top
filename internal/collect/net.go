package collect

import (
	"context"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
)

type NIC struct {
	Name    string   `json:"name"`
	Rx      perSec   `json:"rx"`
	Tx      perSec   `json:"tx"`
	RxTotal uint64   `json:"rxTotal"`
	TxTotal uint64   `json:"txTotal"`
	Errors  uint64   `json:"errors"`
	Drops   uint64   `json:"drops"`
	Addrs   []string `json:"addrs,omitempty"`
	MAC     string   `json:"mac,omitempty"`
	MTU     int      `json:"mtu,omitempty"`
	Up      bool     `json:"up"`
	Virtual bool     `json:"virtual"`
}

var virtualPrefixes = []string{
	"veth", "docker", "br-", "virbr", "vnet", "cni", "flannel", "cali", "tap", "utun",
	"awdl", "llw", "anpi", "gif", "stf", "bridge", "ap1", "isatap", "teredo",
}

type netReader struct{ prev map[string]net.IOCountersStat }

func (r *netReader) read(ctx context.Context, s *Snapshot, elapsed float64) error {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return err
	}
	ifaces, _ := net.InterfacesWithContext(ctx)
	byName := map[string]net.InterfaceStat{}
	for _, i := range ifaces {
		byName[i.Name] = i
	}

	now := map[string]net.IOCountersStat{}
	nics := make([]NIC, 0, len(counters))
	for _, c := range counters {
		now[c.Name] = c
		before, ok := r.prev[c.Name]
		if !ok {
			before = c
		}
		iface := byName[c.Name]
		nic := NIC{
			Name:    c.Name,
			Rx:      rate(c.BytesRecv, before.BytesRecv, elapsed),
			Tx:      rate(c.BytesSent, before.BytesSent, elapsed),
			RxTotal: c.BytesRecv,
			TxTotal: c.BytesSent,
			Errors:  c.Errin + c.Errout,
			Drops:   c.Dropin + c.Dropout,
			MAC:     iface.HardwareAddr,
			MTU:     iface.MTU,
			Up:      slices.Contains(iface.Flags, "up") || len(iface.Flags) == 0,
			Virtual: isVirtual(c.Name),
		}
		for _, a := range iface.Addrs {
			nic.Addrs = append(nic.Addrs, a.Addr)
		}
		nics = append(nics, nic)
	}
	r.prev = now
	slices.SortFunc(nics, func(a, b NIC) int { return strings.Compare(a.Name, b.Name) })
	s.Net = nics
	return nil
}

func isVirtual(name string) bool {
	lower := strings.ToLower(name)
	if strings.TrimRight(lower, "0123456789") == "lo" {
		return true
	}
	for _, p := range virtualPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return strings.Contains(lower, "loopback") || strings.Contains(lower, "pseudo")
}
