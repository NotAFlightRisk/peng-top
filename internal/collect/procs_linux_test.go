package collect

import "testing"

func TestParseStatCopesWithBracketsAndSpacesInTheName(t *testing.T) {
	line := []byte("4242 (tmux: server) (x)) S 1 4242 4242 0 -1 4194560 5000 0 0 0 " +
		"250 150 0 0 20 0 3 0 1000 123456 2048 18446744073709551615")
	got, ok := parseStat(line, 1_700_000_000, 4096)
	if !ok {
		t.Fatal("wouldn't parse a valid line")
	}
	want := rawProc{
		pid: 4242, ppid: 1, cpuSecs: 4, threads: 3, state: "sleep",
		started: 1_700_000_000_000 + 10_000, rss: 2048 * 4096,
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseStatFlagsKernelThreads(t *testing.T) {
	line := []byte("31 (kworker/0:1) I 2 0 0 0 -1 0 0 0 0 0 0 0 0 0 20 0 1 0 50 0 0 0")
	got, ok := parseStat(line, 0, 4096)
	if !ok || !got.kernel || got.state != "idle" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseStatRejectsJunk(t *testing.T) {
	for _, line := range []string{"", "12 (truncated", "x (y) S 1 2 3", "12 (short) S 1 2 3"} {
		if _, ok := parseStat([]byte(line), 0, 4096); ok {
			t.Errorf("parsed %q", line)
		}
	}
}
