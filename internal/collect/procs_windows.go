package collect

import (
	"context"
	"unsafe"

	"golang.org/x/sys/windows"
)

// procOS keeps the process table buffer between readings
type procOS struct{ buf []uint64 }

// systemThread is SYSTEM_THREAD_INFORMATION, a run of which follows each process entry
type systemThread struct {
	KernelTime, UserTime, CreateTime int64
	WaitTime                         uint32
	StartAddress                     uintptr
	ClientID                         [2]uintptr
	Priority, BasePriority           int32
	ContextSwitches                  uint32
	ThreadState                      uint32
	WaitReason                       uint32
}

const (
	procInfoSize   = int(unsafe.Sizeof(windows.SYSTEM_PROCESS_INFORMATION{}))
	threadInfoSize = int(unsafe.Sizeof(systemThread{}))
	threadReady    = 1
	threadRunning  = 2
	// 100ns ticks between 1601 and 1970
	filetimeEpoch = 116444736000000000
)

// list reads the whole process table in one go
func (r *procReader) list(context.Context) ([]rawProc, error) {
	var need uint32
	for {
		var ptr unsafe.Pointer
		if len(r.buf) > 0 {
			ptr = unsafe.Pointer(&r.buf[0])
		}
		err := windows.NtQuerySystemInformation(windows.SystemProcessInformation, ptr, uint32(len(r.buf)*8), &need)
		if err == nil {
			break
		}
		if err != windows.STATUS_INFO_LENGTH_MISMATCH {
			return nil, err
		}
		r.buf = make([]uint64, int(need)*5/4/8+1)
	}

	table := unsafe.Slice((*byte)(unsafe.Pointer(&r.buf[0])), min(int(need), len(r.buf)*8))
	out := make([]rawProc, 0, len(table)/1024)
	for at := 0; at+procInfoSize <= len(table); {
		p := (*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&table[at]))
		end := len(table)
		if p.NextEntryOffset > 0 {
			end = min(at+int(p.NextEntryOffset), end)
		}
		// pid 0 is the idle process, whose cpu time is everyone's idle time
		if p.UniqueProcessID != 0 {
			out = append(out, rawProc{
				pid:     int32(p.UniqueProcessID),
				ppid:    int32(p.InheritedFromUniqueProcessID),
				name:    imageName(table, &p.ImageName),
				started: max(p.CreateTime-filetimeEpoch, 0) / 10_000,
				cpuSecs: float64(p.UserTime+p.KernelTime) / 1e7,
				rss:     uint64(p.WorkingSetSize),
				threads: int32(p.NumberOfThreads),
				state:   threadsState(p, (end-at-procInfoSize)/threadInfoSize),
				kernel:  p.UniqueProcessID == 4,
			})
		}
		next := int(p.NextEntryOffset)
		if next < procInfoSize || next%8 != 0 {
			break
		}
		at += next
	}
	return out, nil
}

// imageName copies the name, as long as it really points inside the table
func imageName(table []byte, s *windows.NTUnicodeString) string {
	start := uintptr(unsafe.Pointer(s.Buffer)) - uintptr(unsafe.Pointer(&table[0]))
	if s.Buffer == nil || s.Length%2 != 0 || s.Length > s.MaximumLength ||
		start > uintptr(len(table)) || uintptr(s.Length) > uintptr(len(table))-start {
		return ""
	}
	return s.String()
}

// threadsState is running if any thread is on a cpu or queued for one
func threadsState(p *windows.SYSTEM_PROCESS_INFORMATION, room int) string {
	first := unsafe.Add(unsafe.Pointer(p), procInfoSize)
	for _, t := range unsafe.Slice((*systemThread)(first), min(int(p.NumberOfThreads), max(room, 0))) {
		if t.ThreadState == threadReady || t.ThreadState == threadRunning {
			return "running"
		}
	}
	return "sleep"
}
