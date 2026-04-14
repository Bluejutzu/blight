//go:build darwin

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
		Subtitle: "Lock this Mac",
		Icon:     "🔒", Keywords: []string{"lock", "screen", "secure"},
	},
	{
		ID: "sleep", Name: "Sleep",
		Subtitle: "Put this Mac to sleep",
		Icon:     "💤", Keywords: []string{"sleep", "suspend", "standby"},
	},
	{
		ID: "shutdown", Name: "Shut Down",
		Subtitle: "Shut down this Mac",
		Icon:     "⏻", Keywords: []string{"shutdown", "shut down", "power off", "turn off"},
	},
	{
		ID: "restart", Name: "Restart",
		Subtitle: "Restart this Mac",
		Icon:     "🔄", Keywords: []string{"restart", "reboot"},
	},
	{
		ID: "logout", Name: "Log Out",
		Subtitle: "Log out of this account",
		Icon:     "🚪", Keywords: []string{"logout", "log out", "sign out", "signout"},
	},

	// --- Apps ---
	{
		ID: "finder", Name: "Finder",
		Subtitle: "Open Finder",
		Icon:     "📁", Keywords: []string{"finder", "files", "folder", "file manager"},
	},
	{
		ID: "terminal", Name: "Terminal",
		Subtitle: "Open Terminal",
		Icon:     "⌨️", Keywords: []string{"terminal", "console", "shell", "bash", "zsh"},
	},
}

func ExecuteSystemCommand(id string) error {
	switch id {
	case "lock-screen":
		// CGSession -suspend is the standard way to lock on macOS
		return exec.Command(
			"/System/Library/CoreServices/Menu Extras/User.menu/Contents/Resources/CGSession",
			"-suspend",
		).Start()
	case "sleep":
		return exec.Command("pmset", "sleepnow").Start()
	case "shutdown":
		return exec.Command("osascript", "-e", `tell app "System Events" to shut down`).Start()
	case "restart":
		return exec.Command("osascript", "-e", `tell app "System Events" to restart`).Start()
	case "logout":
		return exec.Command("osascript", "-e", `tell app "System Events" to log out`).Start()
	case "finder":
		home, _ := os.UserHomeDir()
		return exec.Command("open", home).Start()
	case "terminal":
		return exec.Command("open", "-a", "Terminal").Start()
	}
	return nil
}
