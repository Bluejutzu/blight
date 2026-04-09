//go:build !windows

package commands

type SystemCommand struct {
	ID       string
	Name     string
	Subtitle string
	Icon     string
	Keywords []string
}

// SystemCommands is empty on non-Windows platforms. System-level commands
// (lock, sleep, shutdown, etc.) are handled by the desktop environment and
// are not surfaced through the launcher on Linux or macOS.
var SystemCommands = []SystemCommand{}

func ExecuteSystemCommand(_ string) error {
	return nil
}
