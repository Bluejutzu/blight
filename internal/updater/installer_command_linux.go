//go:build linux

package updater

import "os/exec"

func installerCommand(path string) *exec.Cmd {
	return exec.Command("xdg-open", path)
}
