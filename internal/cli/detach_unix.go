//go:build !windows
// +build !windows

package cli

import (
	"os/exec"
	"syscall"
)

func startDetached(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	return cmd.Start()
}
