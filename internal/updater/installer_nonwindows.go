//go:build !windows

package updater

import (
	"strings"
	"syscall"
)

func isInstallerAsset(name string) bool {
	return strings.HasSuffix(name, ".dmg") ||
		strings.HasSuffix(name, ".pkg") ||
		strings.HasSuffix(name, ".appimage") ||
		strings.HasSuffix(name, ".deb") ||
		strings.HasSuffix(name, ".rpm") ||
		strings.HasSuffix(name, ".tar.gz")
}

func installerTempName() string {
	return "blight-installer"
}

func installerSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
