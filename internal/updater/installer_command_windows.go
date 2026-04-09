//go:build windows

package updater

import "os/exec"

func installerCommand(path string) *exec.Cmd {
	return exec.Command(path)
}
