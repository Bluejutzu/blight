//go:build !windows

package main

import "os/exec"

func startShellCommand(command string, asAdmin bool) error {
	if asAdmin {
		return exec.Command("sudo", "sh", "-lc", command).Start()
	}
	return exec.Command("sh", "-lc", command).Start()
}
