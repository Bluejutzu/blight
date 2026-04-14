//go:build linux

package commands

import (
	"os"
	"os/exec"
)

type SystemCommand struct {
	ID       string
	Name     string
	Subtitle string
	Icon     string
	Keywords []string
}

var SystemCommands = []SystemCommand{
	// --- Power ---
	{
		ID: "lock-screen", Name: "Lock Screen",
		Subtitle: "Lock this computer",
		Icon:     "🔒", Keywords: []string{"lock", "screen", "secure"},
	},
	{
		ID: "sleep", Name: "Suspend",
		Subtitle: "Suspend the computer",
		Icon:     "💤", Keywords: []string{"sleep", "suspend", "standby"},
	},
	{
		ID: "shutdown", Name: "Shut Down",
		Subtitle: "Shut down this computer",
		Icon:     "⏻", Keywords: []string{"shutdown", "shut down", "power off", "turn off"},
	},
	{
		ID: "restart", Name: "Restart",
		Subtitle: "Restart this computer",
		Icon:     "🔄", Keywords: []string{"restart", "reboot"},
	},
	{
		ID: "logout", Name: "Log Out",
		Subtitle: "Log out of the current session",
		Icon:     "🚪", Keywords: []string{"logout", "log out", "sign out", "signout"},
	},

	// --- Apps ---
	{
		ID: "file-manager", Name: "File Manager",
		Subtitle: "Open the file manager",
		Icon:     "📁", Keywords: []string{"files", "folder", "file manager", "nautilus", "dolphin"},
	},
	{
		ID: "terminal", Name: "Terminal",
		Subtitle: "Open a terminal emulator",
		Icon:     "⌨️", Keywords: []string{"terminal", "console", "shell", "bash", "zsh"},
	},
}

func ExecuteSystemCommand(id string) error {
	switch id {
	case "lock-screen":
		// loginctl works on systemd desktops; fall back to xdg-screensaver
		if err := exec.Command("loginctl", "lock-session").Start(); err != nil {
			return exec.Command("xdg-screensaver", "lock").Start()
		}
		return nil
	case "sleep":
		return exec.Command("systemctl", "suspend").Start()
	case "shutdown":
		return exec.Command("systemctl", "poweroff").Start()
	case "restart":
		return exec.Command("systemctl", "reboot").Start()
	case "logout":
		return exec.Command("loginctl", "terminate-session", "self").Start()
	case "file-manager":
		home, _ := os.UserHomeDir()
		return exec.Command("xdg-open", home).Start()
	case "terminal":
		for _, term := range []string{"x-terminal-emulator", "gnome-terminal", "konsole", "xfce4-terminal", "xterm"} {
			if p, err := exec.LookPath(term); err == nil {
				return exec.Command(p).Start()
			}
		}
		return nil
	}
	return nil
}
