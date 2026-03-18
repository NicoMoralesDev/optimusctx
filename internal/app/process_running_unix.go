//go:build !windows

package app

import (
	"errors"
	"syscall"
)

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	if err := syscall.Kill(pid, 0); err != nil {
		return !errors.Is(err, syscall.ESRCH)
	}
	return true
}
