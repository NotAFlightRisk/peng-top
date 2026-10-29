package collect

import (
	"bufio"
	"context"
	"os"
	"os/user"
	"strconv"
	"strings"
)

type Proc struct {
	PID     int32  `json:"pid"`
	PPID    int32  `json:"ppid"`
	Name    string `json:"name"`
	User    string `json:"user"`
	CPU     pct    `json:"cpu"`
	Mem     pct    `json:"mem"`
	RSS     uint64 `json:"rss"`
	Threads int32  `json:"threads"`
	State   string `json:"state"`
	Cmd     string `json:"cmd,omitempty"`
	Started int64  `json:"started"`
	Kernel  bool   `json:"kernel,omitempty"`
}

type Procs struct {
	Total   int    `json:"total"`
	Running int    `json:"running"`
	Threads int    `json:"threads"`
	List    []Proc `json:"list"`
}

// rawProc is what each platform reads every time round
type rawProc struct {
	pid, ppid int32
	started   int64
	cpuSecs   float64
	rss       uint64
	threads   int32
	state     string
	kernel    bool
}

// about is what never changes for a process, so it's only read once
type about struct {
	name, cmd, user string
}

type tracked struct {
	about
	started int64
	cpuSecs float64
}

// long enough to tell two java processes apart, short enough not to ship someone's whole prompt
const maxCmd = 300

type procReader struct {
	env   Env
	known map[int32]tracked
	users map[uint32]string
}

func (r *procReader) read(ctx context.Context, s *Snapshot, elapsed float64) error {
	raws, err := r.list(ctx)
	if err != nil {
		return err
	}
	var memTotal float64
	if s.Mem != nil {
		memTotal = float64(s.Mem.Total)
	}

	known := make(map[int32]tracked, len(raws))
	out := &Procs{List: make([]Proc, 0, len(raws))}
	for _, raw := range raws {
		t, ok := r.known[raw.pid]
		if !ok || t.started != raw.started {
			t = tracked{about: r.describe(ctx, raw.pid), started: raw.started, cpuSecs: raw.cpuSecs}
		}
		p := Proc{
			PID: raw.pid, PPID: raw.ppid, Name: t.name, User: t.user, Cmd: t.cmd,
			RSS: raw.rss, Threads: raw.threads, State: raw.state, Started: raw.started,
			Kernel: raw.kernel,
		}
		if elapsed > 0 {
			p.CPU = pct(max(raw.cpuSecs-t.cpuSecs, 0) / elapsed * 100)
		}
		if memTotal > 0 {
			p.Mem = pct(float64(raw.rss) / memTotal * 100)
		}
		t.cpuSecs = raw.cpuSecs
		known[raw.pid] = t

		out.Threads += int(raw.threads)
		if raw.state == "running" {
			out.Running++
		}
		out.List = append(out.List, p)
	}
	r.known = known
	out.Total = len(out.List)
	s.Procs = out
	return nil
}

// username checks the host's passwd first (we might be in a container), then the OS
func (r *procReader) username(uid uint32) string {
	if r.users == nil {
		r.users = readPasswd(r.env.HostPath("/etc/passwd"))
	}
	if name, ok := r.users[uid]; ok {
		return name
	}
	name := strconv.FormatUint(uint64(uid), 10)
	if u, err := user.LookupId(name); err == nil && r.env.Root == "" {
		name = u.Username
	}
	r.users[uid] = name
	return name
}

func readPasswd(path string) map[uint32]string {
	users := map[uint32]string{}
	f, err := os.Open(path)
	if err != nil {
		return users
	}
	defer f.Close()
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		fields := strings.Split(lines.Text(), ":")
		if len(fields) < 3 {
			continue
		}
		if uid, err := strconv.ParseUint(fields[2], 10, 32); err == nil {
			users[uint32(uid)] = fields[0]
		}
	}
	return users
}

func clip(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) <= maxCmd {
		return cmd
	}
	return strings.ToValidUTF8(cmd[:maxCmd], "") + "…"
}
