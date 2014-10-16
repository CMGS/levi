package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

func NetNsSynchronize(pid string, fn func() error) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	f, err := os.OpenFile("/proc/self/ns/net", os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer func() {
		system.Setns(f.Fd(), 0)
		f.Close()
	}()
	if err = setNetNs(pid); err != nil {
		return
	}
	return fn()
}

func setNetNs(pid string) (err error) {
	path := filepath.Join("/proc", pid, "ns", "net")
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()
	if err = system.Setns(f.Fd(), 0); err != nil {
		return
	}
	return
}
