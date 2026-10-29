package collect

import (
	"context"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/unix"
)

type procOS struct{}

const (
	procPIDTaskInfo = 4
	statStop        = 4
	statZombie      = 5
)

type timebase struct{ numer, denom uint32 }

// libproc is proc_pidinfo plus the mach tick length, looked up once
var libproc = sync.OnceValues(func() (func(pid int32, task *process.ProcTaskInfo) bool, error) {
	lib, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}
	var pidinfo func(pid, flavor int32, arg uint64, buf unsafe.Pointer, size int32) int32
	var machTimebase func(*timebase) int32
	for name, fn := range map[string]any{"proc_pidinfo": &pidinfo, "mach_timebase_info": &machTimebase} {
		sym, err := purego.Dlsym(lib, name)
		if err != nil {
			return nil, err
		}
		purego.RegisterFunc(fn, sym)
	}
	tb := timebase{1, 1}
	if machTimebase(&tb) != 0 || tb.denom == 0 {
		tb = timebase{1, 1}
	}
	size := int32(unsafe.Sizeof(process.ProcTaskInfo{}))
	return func(pid int32, task *process.ProcTaskInfo) bool {
		if pidinfo(pid, procPIDTaskInfo, 0, unsafe.Pointer(task), size) != size {
			return false
		}
		task.Total_user = task.Total_user * uint64(tb.numer) / uint64(tb.denom)
		task.Total_system = task.Total_system * uint64(tb.numer) / uint64(tb.denom)
		return true
	}, nil
})

// list reads the kernel's process table, then one task info per process
func (r *procReader) list(context.Context) ([]rawProc, error) {
	taskInfo, err := libproc()
	if err != nil {
		return nil, err
	}
	kprocs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, err
	}
	out := make([]rawProc, 0, len(kprocs))
	for i := range kprocs {
		k := &kprocs[i]
		raw := rawProc{
			pid:     k.Proc.P_pid,
			ppid:    k.Eproc.Ppid,
			started: k.Proc.P_starttime.Sec*1000 + int64(k.Proc.P_starttime.Usec)/1000,
			kernel:  k.Proc.P_pid == 0,
		}
		var task process.ProcTaskInfo
		if taskInfo(raw.pid, &task) {
			raw.cpuSecs = float64(task.Total_user+task.Total_system) / 1e9
			raw.rss, raw.threads = task.Resident_size, task.Threadnum
			raw.state = "sleep"
			if task.Numrunning > 0 {
				raw.state = "running"
			}
		}
		switch k.Proc.P_stat {
		case statStop:
			raw.state = "stop"
		case statZombie:
			raw.state = "zombie"
		}
		out = append(out, raw)
	}
	return out, nil
}
