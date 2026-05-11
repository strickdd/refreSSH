//go:build windows
// +build windows

package cli

import (
	"os/exec"
	"syscall"
)

func startDetached(cmd *exec.Cmd) error {
	// CREATE_BREAKAWAY_FROM_JOB | DETACHED_PROCESS
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x01000000 | 0x00000008,
	}
	return cmd.Start()
}
