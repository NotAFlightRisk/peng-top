package collect

import (
	"testing"
	"unsafe"
)

func TestProcessTableLayoutMatchesWindows(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("only checked on 64-bit")
	}
	if procInfoSize != 256 || threadInfoSize != 80 {
		t.Fatalf("process entry %d bytes, thread %d, want 256 and 80", procInfoSize, threadInfoSize)
	}
}
