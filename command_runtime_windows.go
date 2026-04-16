//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

func startShellCommand(command string, asAdmin bool) error {
	if asAdmin {
		escalation := fmt.Sprintf(
			"Start-Process -Verb RunAs powershell.exe -ArgumentList @('-NoProfile','-WindowStyle','Hidden','-Command',%q)",
			command,
		)
		process := exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", escalation)
		process.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return process.Start()
	}

	process := exec.Command("powershell.exe", "-NoProfile", "-WindowStyle", "Hidden", "-Command", command)
	process.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return process.Start()
}
