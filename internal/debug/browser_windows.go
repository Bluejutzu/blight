//go:build windows

package debug

import (
	"os/exec"
	"strings"
)

func openBrowser(url string) {
	exec.Command("cmd", "/c", "start", "", strings.ReplaceAll(url, "&", "^&")).Start()
}
