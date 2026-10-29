package collect

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Mode says whose numbers you're looking at
type Mode string

const (
	ModeHost          Mode = "host"
	ModeContainer     Mode = "container"
	ModeContainerHost Mode = "host-via-container"
)

// Env is where we are and where the host's files live
type Env struct {
	Mode Mode   `json:"mode"`
	Root string `json:"-"`
}

// Detect works out the mode, and points gopsutil at the host's files when they're mounted
func Detect(hostRoot string) Env {
	if !inContainer() {
		return Env{Mode: ModeHost}
	}
	if hostRoot == "" || !exists(filepath.Join(hostRoot, "proc", "1")) {
		slog.Warn("in a container with no host mounted, so these are the container's own numbers",
			"fix", "see the README for the docker flags")
		return Env{Mode: ModeContainer}
	}
	for env, dir := range map[string]string{
		"HOST_PROC": "proc", "HOST_SYS": "sys", "HOST_ETC": "etc",
		"HOST_VAR": "var", "HOST_RUN": "run", "HOST_DEV": "dev",
	} {
		os.Setenv(env, filepath.Join(hostRoot, dir))
	}
	os.Setenv("HOST_ROOT", hostRoot)
	os.Setenv("HOST_PROC_MOUNTINFO", filepath.Join(hostRoot, "proc", "1", "mountinfo"))
	if !sameNetwork(hostRoot) {
		slog.Warn("not on the host's network, so network numbers are the container's",
			"fix", "run with --network host")
	}
	return Env{Mode: ModeContainerHost, Root: hostRoot}
}

// HostPath turns a path on the host into one we can open from here
func (e Env) HostPath(path string) string {
	if e.Root == "" {
		return path
	}
	return filepath.Join(e.Root, path)
}

func inContainer() bool {
	if exists("/.dockerenv") || exists("/run/.containerenv") || os.Getenv("container") != "" {
		return true
	}
	cgroup, _ := os.ReadFile("/proc/1/cgroup")
	for _, hint := range []string{"docker", "kubepods", "containerd", "lxc", "podman"} {
		if strings.Contains(string(cgroup), hint) {
			return true
		}
	}
	return false
}

func sameNetwork(hostRoot string) bool {
	ours, err1 := os.Readlink("/proc/self/ns/net")
	theirs, err2 := os.Readlink(filepath.Join(hostRoot, "proc", "1", "ns", "net"))
	return err1 != nil || err2 != nil || ours == theirs
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
