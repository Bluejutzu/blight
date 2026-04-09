//go:build darwin

package updater

import "os/exec"

func installerCommand(path string) *exec.Cmd {
	return exec.Command("open", path)
}
