package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"blight/internal/apps"
	"blight/internal/commands"
	"blight/internal/debug"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func shortcutAction(id, label, iconName, shortcut string, destructive bool) ContextAction {
	return ContextAction{
		ID:          id,
		Label:       label,
		Icon:        iconName,
		Shortcut:    shortcut,
		Destructive: destructive,
	}
}

func (app *App) executeResult(id string) string {
	debug.Get().Info("execute", map[string]interface{}{"id": id})

	switch {
	case strings.HasPrefix(id, "command:"):
		commandID, commandArgument, _ := parseCommandResultID(id)
		return app.executeCommand(commandID, commandArgument)
	case strings.HasPrefix(id, "calc:"):
		runtime.ClipboardSetText(app.ctx, strings.TrimPrefix(id, "calc:"))
		return "copied"
	case strings.HasPrefix(id, "web-search:"):
		query := strings.TrimPrefix(id, "web-search:")
		template := app.config.SearchEngineURL
		if template == "" {
			template = "https://www.google.com/search?q=%s"
		}
		searchURL := strings.ReplaceAll(template, "%s", url.QueryEscape(query))
		runtime.BrowserOpenURL(app.ctx, searchURL)
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	case strings.HasPrefix(id, "url-open:"):
		target := strings.TrimPrefix(id, "url-open:")
		runtime.BrowserOpenURL(app.ctx, target)
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	case strings.HasPrefix(id, "clip-"):
		var entryIndex int
		fmt.Sscanf(strings.TrimPrefix(id, "clip-"), "%d", &entryIndex)
		if app.clipboard.CopyToClipboard(entryIndex) {
			return "copied"
		}
		return "error"
	case strings.HasPrefix(id, "sys-"):
		app.usage.Record(id)
		systemID := strings.TrimPrefix(id, "sys-")
		if err := commands.ExecuteSystemCommand(systemID); err != nil {
			return err.Error()
		}
		return "ok"
	case strings.HasPrefix(id, "file-open:"):
		filePath := strings.TrimPrefix(id, "file-open:")
		app.usage.Record(id)
		shellOpen(filePath)
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	case strings.HasPrefix(id, "file-reveal:"):
		explorerSelect(strings.TrimPrefix(id, "file-reveal:"))
		return "ok"
	case strings.HasPrefix(id, "dir-open:"):
		directoryPath := strings.TrimPrefix(id, "dir-open:")
		app.usage.Record(id)
		shellOpen(directoryPath)
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	}

	for _, application := range app.scanner.Apps() {
		if application.Name != id {
			continue
		}
		app.usage.Record(id)
		if err := apps.Launch(application); err != nil {
			return err.Error()
		}
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	}

	return "not found"
}

func (app *App) contextActionsFor(id string) []ContextAction {
	switch {
	case strings.HasPrefix(id, "command:"):
		commandID, commandArgument, _ := parseCommandResultID(id)
		commandDefinition, found := app.findCommand(commandID)
		if !found {
			return nil
		}
		pinLabel := "Pin to Home"
		pinIcon := icon("\uE718", "P")
		if slices.Contains(app.config.PinnedItems, commandPinnedID(commandID)) {
			pinLabel = "Unpin from Home"
			pinIcon = icon("\uE77A", "P")
		}
		secondaryLabel := "Copy Value"
		if commandDefinition.ActionType == "open_path" {
			secondaryLabel = "Copy Path"
		}
		if commandDefinition.ActionType == "run_shell" {
			secondaryLabel = "Copy Command"
		}

		actions := []ContextAction{
			shortcutAction("run", "Run", icon("\uE768", ">"), "Enter", false),
			shortcutAction("copy", secondaryLabel, icon("\uE8C8", "C"), "Ctrl+Enter", false),
			shortcutAction("edit-command", "Edit Command", icon("\uE70F", "E"), "", false),
			shortcutAction("duplicate-command", "Duplicate Command", icon("\uE8C8", "D"), "", false),
			shortcutAction("pin", pinLabel, pinIcon, "", false),
		}
		if _, _, isUserCommand := app.findUserCommand(commandID); isUserCommand {
			actions = append(actions, shortcutAction("delete-command", "Delete Command", icon("\uE74D", "X"), "", true))
		}
		if commandDefinition.RequiresArgument && strings.TrimSpace(commandArgument) == "" {
			actions[0].Label = "Run (type an argument first)"
		}
		return actions
	case strings.HasPrefix(id, "dir-open:"):
		return []ContextAction{
			shortcutAction("open", "Open", icon("\uE768", ">"), "Enter", false),
			shortcutAction("terminal", "Open in Terminal", icon("\uE756", "T"), "Ctrl+Enter", false),
			shortcutAction("copy-path", "Copy Path", icon("\uE8C8", "C"), "", false),
		}
	case strings.HasPrefix(id, "file-open:"):
		return []ContextAction{
			shortcutAction("open", "Open", icon("\uE768", ">"), "Enter", false),
			shortcutAction("explorer", revealLabel(), icon("\uE8B7", "F"), "Ctrl+Enter", false),
			shortcutAction("copy-path", "Copy Path", icon("\uE8C8", "C"), "", false),
			shortcutAction("copy-name", "Copy Name", icon("\uE70F", "N"), "", false),
		}
	case strings.HasPrefix(id, "clip-"):
		return []ContextAction{
			shortcutAction("copy", "Copy", icon("\uE8C8", "C"), "Enter", false),
			shortcutAction("delete", "Delete", icon("\uE74D", "X"), "Ctrl+Enter", true),
		}
	case strings.HasPrefix(id, "sys-"):
		return []ContextAction{
			shortcutAction("run", "Run", icon("\uE768", ">"), "Enter", false),
		}
	case strings.HasPrefix(id, "calc:") || strings.HasPrefix(id, "web-search:"):
		return nil
	default:
		pinLabel := "Pin to Home"
		pinIcon := icon("\uE718", "P")
		if slices.Contains(app.config.PinnedItems, canonicalPinnedID(id)) {
			pinLabel = "Unpin from Home"
			pinIcon = icon("\uE77A", "P")
		}
		return []ContextAction{
			shortcutAction("open", "Open", icon("\uE768", ">"), "Enter", false),
			shortcutAction("admin", elevateLabel(), icon("\uE7EF", "A"), "Ctrl+Enter", false),
			shortcutAction("explorer", revealLabel(), icon("\uE8B7", "F"), "", false),
			shortcutAction("copy-path", "Copy Path", icon("\uE8C8", "C"), "", false),
			shortcutAction("pin", pinLabel, pinIcon, "", false),
		}
	}
}

func (app *App) performContextAction(resultID string, actionID string) string {
	switch {
	case strings.HasPrefix(resultID, "command:"):
		commandID, commandArgument, parsed := parseCommandResultID(resultID)
		if !parsed {
			return "not found"
		}
		commandDefinition, found := app.findCommand(commandID)
		if !found {
			return "not found"
		}
		switch actionID {
		case "run":
			return app.executeCommand(commandID, commandArgument)
		case "copy":
			copyValue := resolveCommandPreview(commandDefinition, commandArgument)
			if copyValue == "Type an argument" {
				return copyValue
			}
			runtime.ClipboardSetText(app.ctx, copyValue)
			return "copied"
		case "edit-command":
			return "edit-command:" + commandID
		case "duplicate-command":
			newCommand := duplicateCommand(commandDefinition)
			if err := app.SaveCommand(newCommand); err != nil {
				return err.Error()
			}
			return "duplicated:" + newCommand.ID
		case "delete-command":
			if err := app.DeleteCommand(commandID); err != nil {
				return err.Error()
			}
			return "deleted"
		case "pin":
			if app.TogglePinned(resultID) {
				return "pinned"
			}
			return "unpinned"
		}
		return "unknown action"
	case strings.HasPrefix(resultID, "dir-open:"):
		directoryPath := strings.TrimPrefix(resultID, "dir-open:")
		switch actionID {
		case "open":
			shellOpen(directoryPath)
			runtime.WindowHide(app.ctx)
			app.visible.Store(false)
			return "ok"
		case "terminal":
			openInTerminal(directoryPath)
			return "ok"
		case "copy-path":
			runtime.ClipboardSetText(app.ctx, directoryPath)
			return "ok"
		}
		return "unknown action"
	case strings.HasPrefix(resultID, "file-open:"):
		filePath := strings.TrimPrefix(resultID, "file-open:")
		switch actionID {
		case "open":
			shellOpen(filePath)
			runtime.WindowHide(app.ctx)
			app.visible.Store(false)
			return "ok"
		case "explorer":
			explorerSelect(filePath)
			return "ok"
		case "copy-path":
			runtime.ClipboardSetText(app.ctx, filePath)
			return "ok"
		case "copy-name":
			runtime.ClipboardSetText(app.ctx, filepath.Base(filePath))
			return "ok"
		}
		return "unknown action"
	case strings.HasPrefix(resultID, "clip-"):
		var entryIndex int
		fmt.Sscanf(strings.TrimPrefix(resultID, "clip-"), "%d", &entryIndex)
		switch actionID {
		case "copy", "open":
			if app.clipboard.CopyToClipboard(entryIndex) {
				return "copied"
			}
			return "error"
		case "delete":
			app.clipboard.Delete(entryIndex)
			return "ok"
		}
		return "unknown action"
	case strings.HasPrefix(resultID, "sys-"):
		if actionID != "run" {
			return "unknown action"
		}
		app.usage.Record(resultID)
		systemID := strings.TrimPrefix(resultID, "sys-")
		if err := commands.ExecuteSystemCommand(systemID); err != nil {
			return err.Error()
		}
		return "ok"
	}

	var targetApplication apps.AppEntry
	found := false
	for _, application := range app.scanner.Apps() {
		if application.Name != resultID {
			continue
		}
		targetApplication = application
		found = true
		break
	}
	if !found {
		return "not found"
	}

	switch actionID {
	case "open":
		app.usage.Record(resultID)
		if err := apps.Launch(targetApplication); err != nil {
			return err.Error()
		}
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	case "admin":
		app.usage.Record(resultID)
		if err := runAsAdmin(targetApplication.Path); err != nil {
			return err.Error()
		}
		runtime.WindowHide(app.ctx)
		app.visible.Store(false)
		return "ok"
	case "explorer":
		explorerSelect(targetApplication.Path)
		return "ok"
	case "copy-path":
		runtime.ClipboardSetText(app.ctx, targetApplication.Path)
		return "ok"
	case "pin":
		if app.TogglePinned(resultID) {
			return "pinned"
		}
		return "unpinned"
	}

	return "unknown action"
}
