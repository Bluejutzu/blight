package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"blight/internal/apps"
	"blight/internal/commands"
	"blight/internal/files"
	"blight/internal/search"
)

const (
	resultKindApp       = "app"
	resultKindCommand   = "command"
	resultKindFile      = "file"
	resultKindFolder    = "folder"
	resultKindClipboard = "clipboard"
	resultKindSystem    = "system"
	resultKindWeb       = "web"
	resultKindCalc      = "calc"
)

type searchRequest struct {
	RawQuery          string
	NormalizedQuery   string
	CommandMode       bool
	CommandKeyword    string
	CommandArgument   string
	CommandExpression string
}

type rankedCandidate struct {
	Result       SearchResult
	CategoryKind string
}

func newSearchRequest(query string) searchRequest {
	trimmedQuery := strings.TrimSpace(query)
	request := searchRequest{
		RawQuery: trimmedQuery,
	}

	commandExpression := trimmedQuery
	if strings.HasPrefix(trimmedQuery, ">") {
		request.CommandMode = true
		commandExpression = strings.TrimSpace(trimmedQuery[1:])
		request.CommandExpression = commandExpression
	} else {
		request.CommandExpression = trimmedQuery
	}

	normalizedQuery := trimmedQuery
	if request.CommandMode {
		normalizedQuery = commandExpression
	}
	request.NormalizedQuery = strings.ToLower(strings.TrimSpace(normalizedQuery))

	if commandExpression == "" {
		return request
	}

	commandParts := strings.Fields(commandExpression)
	if len(commandParts) == 0 {
		return request
	}

	request.CommandKeyword = strings.ToLower(commandParts[0])
	if len(commandExpression) > len(commandParts[0]) {
		request.CommandArgument = strings.TrimSpace(commandExpression[len(commandParts[0]):])
	}

	return request
}

func providerCaps(maxResults int) map[string]int {
	if maxResults < 1 {
		maxResults = 1
	}

	folderCap := (maxResults + 1) / 2
	if folderCap < 1 {
		folderCap = 1
	}

	return map[string]int{
		resultKindCommand:   min(6, maxResults),
		resultKindApp:       maxResults,
		resultKindFile:      min(6, maxResults),
		resultKindFolder:    min(4, folderCap),
		resultKindClipboard: min(6, maxResults),
		resultKindSystem:    min(5, maxResults),
		resultKindWeb:       1,
		resultKindCalc:      1,
	}
}

func providerPriority(kind string) int {
	switch kind {
	case resultKindCommand:
		return 0
	case resultKindCalc:
		return 1
	case resultKindApp:
		return 2
	case resultKindFolder:
		return 3
	case resultKindFile:
		return 4
	case resultKindClipboard:
		return 5
	case resultKindSystem:
		return 6
	case resultKindWeb:
		return 7
	default:
		return 8
	}
}

func rankMatch(query string, targets []string, keywords []string, usageScore int, pinned bool, providerBoost int) int {
	baseScore := search.BestScore(query, targets, keywords)
	if baseScore == 0 && strings.TrimSpace(query) != "" {
		return 0
	}

	totalScore := baseScore + usageScore + providerBoost
	if pinned {
		totalScore += 6000
	}
	return totalScore
}

func buildAppResult(application apps.AppEntry, category string, score int) SearchResult {
	subtitle := "Application"
	if !application.IsLnk {
		subtitle = prettifyPath(application.Path)
	}

	return SearchResult{
		ID:                   application.Name,
		Title:                application.Name,
		Subtitle:             subtitle,
		Category:             category,
		Path:                 application.Path,
		Kind:                 resultKindApp,
		Score:                score,
		PrimaryActionLabel:   "Open",
		SecondaryActionLabel: elevateLabel(),
		SupportsActions:      true,
	}
}

func buildCommandResult(command CommandDefinition, argument string, category string, score int) SearchResult {
	secondaryActionLabel := "Copy Value"
	if command.ActionType == "open_path" {
		secondaryActionLabel = "Copy Path"
	}
	if command.ActionType == "run_shell" {
		secondaryActionLabel = "Copy Command"
	}

	return SearchResult{
		ID:                   commandResultID(command.ID, argument),
		Title:                command.Title,
		Subtitle:             resolveCommandPreview(command, argument),
		Category:             category,
		Kind:                 resultKindCommand,
		Score:                score,
		PrimaryActionLabel:   "Run",
		SecondaryActionLabel: secondaryActionLabel,
		SupportsActions:      true,
	}
}

func buildSystemResult(systemCommand commands.SystemCommand, score int) SearchResult {
	return SearchResult{
		ID:                 "sys-" + systemCommand.ID,
		Title:              systemCommand.Name,
		Subtitle:           systemCommand.Subtitle,
		Icon:               systemCommand.Icon,
		Category:           "System",
		Kind:               resultKindSystem,
		Score:              score,
		PrimaryActionLabel: "Run",
		SupportsActions:    true,
	}
}

func buildFileResult(fileEntry files.FileEntry, score int) SearchResult {
	return SearchResult{
		ID:                   "file-open:" + fileEntry.Path,
		Title:                fileEntry.Name,
		Subtitle:             prettifyPath(fileEntry.Dir),
		Category:             "Files",
		Path:                 fileEntry.Path,
		Kind:                 resultKindFile,
		Score:                score,
		PrimaryActionLabel:   "Open",
		SecondaryActionLabel: revealLabel(),
		SupportsActions:      true,
	}
}

func buildFolderResult(folderEntry files.FileEntry, score int) SearchResult {
	return SearchResult{
		ID:                   "dir-open:" + folderEntry.Path,
		Title:                folderEntry.Name,
		Subtitle:             prettifyPath(folderEntry.Path),
		Category:             "Folders",
		Path:                 folderEntry.Path,
		Kind:                 resultKindFolder,
		Score:                score,
		PrimaryActionLabel:   "Open",
		SecondaryActionLabel: "Open in Terminal",
		SupportsActions:      true,
	}
}

func buildClipboardResult(entry commands.ClipboardEntry, entryIndex int, score int) SearchResult {
	preview := entry.Content
	if len(preview) > 80 {
		preview = preview[:80] + "..."
	}
	return SearchResult{
		ID:                   fmt.Sprintf("clip-%d", entryIndex),
		Title:                preview,
		Subtitle:             "Clipboard history",
		Category:             "Clipboard",
		Kind:                 resultKindClipboard,
		Score:                score,
		PrimaryActionLabel:   "Copy",
		SecondaryActionLabel: "Delete",
		SupportsActions:      true,
	}
}

func buildCalcResult(calcResult commands.CalcResult) SearchResult {
	return SearchResult{
		ID:                 "calc:" + calcResult.Result,
		Title:              calcResult.Result,
		Subtitle:           calcResult.Expression + " - press Enter to copy",
		Category:           "Calculator",
		Kind:               resultKindCalc,
		Score:              12000,
		PrimaryActionLabel: "Copy",
	}
}

func (app *App) searchAll(query string) []SearchResult {
	if query == "" {
		return app.defaultHomeResults()
	}
	if strings.HasPrefix(query, "~") || isAbsPath(query) {
		return app.searchPathResults(query)
	}

	request := newSearchRequest(query)
	if request.CommandMode && request.CommandExpression == "" {
		return app.commandCandidates(request)
	}

	var candidates []rankedCandidate
	if request.CommandMode {
		candidates = append(candidates, app.wrapResults(app.commandCandidates(request), resultKindCommand)...)
	} else {
		candidates = append(candidates, app.commandProviderCandidates(request)...)
		candidates = append(candidates, app.appProviderCandidates(request)...)
		candidates = append(candidates, app.calcProviderCandidates(request)...)
		candidates = append(candidates, app.folderProviderCandidates(request)...)
		candidates = append(candidates, app.fileProviderCandidates(request)...)
		candidates = append(candidates, app.clipboardProviderCandidates(request)...)
		candidates = append(candidates, app.systemProviderCandidates(request)...)
		candidates = append(candidates, app.urlProviderCandidates(request)...)
	}

	results := applyCapsAndSort(candidates, app.maxResults())
	if request.CommandMode {
		return results
	}

	webResult := SearchResult{
		ID:                 "web-search:" + query,
		Title:              `Search the web for "` + query + `"`,
		Subtitle:           "Open in your default browser",
		Category:           "Web",
		Kind:               resultKindWeb,
		PrimaryActionLabel: "Search",
	}

	if len(results) == 0 {
		return []SearchResult{webResult}
	}
	return append(results, webResult)
}

func (app *App) wrapResults(results []SearchResult, categoryKind string) []rankedCandidate {
	wrappedResults := make([]rankedCandidate, 0, len(results))
	for _, result := range results {
		wrappedResults = append(wrappedResults, rankedCandidate{
			Result:       result,
			CategoryKind: categoryKind,
		})
	}
	return wrappedResults
}

func applyCapsAndSort(candidates []rankedCandidate, maxResults int) []SearchResult {
	sort.SliceStable(candidates, func(leftIndex, rightIndex int) bool {
		leftCandidate := candidates[leftIndex]
		rightCandidate := candidates[rightIndex]
		if leftCandidate.Result.Score != rightCandidate.Result.Score {
			return leftCandidate.Result.Score > rightCandidate.Result.Score
		}
		leftPriority := providerPriority(leftCandidate.CategoryKind)
		rightPriority := providerPriority(rightCandidate.CategoryKind)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return strings.ToLower(leftCandidate.Result.Title) < strings.ToLower(rightCandidate.Result.Title)
	})

	categoryCaps := providerCaps(maxResults)
	categoryCounts := make(map[string]int)
	results := make([]SearchResult, 0, len(candidates))

	for _, candidate := range candidates {
		categoryLimit, hasCategoryLimit := categoryCaps[candidate.CategoryKind]
		if hasCategoryLimit && categoryCounts[candidate.CategoryKind] >= categoryLimit {
			continue
		}
		categoryCounts[candidate.CategoryKind]++
		results = append(results, candidate.Result)
	}

	return results
}

func (app *App) commandCandidates(request searchRequest) []SearchResult {
	candidates := app.commandProviderCandidates(request)
	results := applyCapsAndSort(candidates, app.maxResults())
	return results
}

func (app *App) commandProviderCandidates(request searchRequest) []rankedCandidate {
	var candidates []rankedCandidate
	for _, commandDefinition := range app.commandDefinitions() {
		isPinned := slices.Contains(app.config.PinnedItems, commandPinnedID(commandDefinition.ID))
		usageScore := app.usage.Score(commandPinnedID(commandDefinition.ID))

		if request.CommandMode && request.CommandExpression == "" {
			score := 4000 + usageScore
			if isPinned {
				score += 6000
			}
			candidates = append(candidates, rankedCandidate{
				Result:       buildCommandResult(commandDefinition, "", "Commands", score),
				CategoryKind: resultKindCommand,
			})
			continue
		}

		score := 0
		argument := ""
		if request.CommandKeyword != "" && strings.EqualFold(request.CommandKeyword, commandDefinition.Keyword) {
			argument = request.CommandArgument
			score = 14000 + usageScore
			if isPinned {
				score += 6000
			}
		} else {
			score = rankMatch(
				request.NormalizedQuery,
				commandSearchBlob(commandDefinition),
				append([]string{commandDefinition.Keyword}, commandDefinition.Keywords...),
				usageScore,
				isPinned,
				2500,
			)
		}

		if score == 0 {
			continue
		}

		if request.CommandMode {
			score += 2000
		}

		candidates = append(candidates, rankedCandidate{
			Result:       buildCommandResult(commandDefinition, argument, "Commands", score),
			CategoryKind: resultKindCommand,
		})
	}
	return candidates
}

func (app *App) appProviderCandidates(request searchRequest) []rankedCandidate {
	var candidates []rankedCandidate
	for _, application := range app.scanner.Apps() {
		isPinned := slices.Contains(app.config.PinnedItems, application.Name)
		usageScore := app.usage.Score(application.Name)
		score := rankMatch(
			request.NormalizedQuery,
			[]string{application.Name, application.Path},
			nil,
			usageScore,
			isPinned,
			0,
		)
		if score == 0 {
			continue
		}
		candidates = append(candidates, rankedCandidate{
			Result:       buildAppResult(application, "Applications", score),
			CategoryKind: resultKindApp,
		})
	}
	return candidates
}

func (app *App) systemProviderCandidates(request searchRequest) []rankedCandidate {
	var candidates []rankedCandidate
	for _, systemCommand := range commands.SystemCommands {
		resultID := "sys-" + systemCommand.ID
		score := rankMatch(
			request.NormalizedQuery,
			[]string{systemCommand.Name, systemCommand.Subtitle},
			systemCommand.Keywords,
			app.usage.Score(resultID),
			false,
			0,
		)
		if score == 0 {
			continue
		}
		candidates = append(candidates, rankedCandidate{
			Result:       buildSystemResult(systemCommand, score),
			CategoryKind: resultKindSystem,
		})
	}
	return candidates
}

func (app *App) fileProviderCandidates(request searchRequest) []rankedCandidate {
	if len(request.RawQuery) < 2 {
		return nil
	}
	if app.fileIdx.Status().State != "ready" {
		return nil
	}

	allUsageScores := app.usage.AllScores()
	fileUsageScores := make(map[string]int, len(allUsageScores))
	for usageKey, usageScore := range allUsageScores {
		if strings.HasPrefix(usageKey, "file-open:") {
			fileUsageScores[strings.TrimPrefix(usageKey, "file-open:")] = usageScore
		}
	}

	var candidates []rankedCandidate
	for _, fileEntry := range app.fileIdx.SearchFiles(request.RawQuery, fileUsageScores) {
		score := rankMatch(
			request.NormalizedQuery,
			[]string{fileEntry.Name, fileEntry.Path, fileEntry.Dir},
			[]string{strings.TrimPrefix(strings.ToLower(filepath.Ext(fileEntry.Name)), ".")},
			fileUsageScores[fileEntry.Path],
			false,
			0,
		)
		if score == 0 {
			continue
		}
		candidates = append(candidates, rankedCandidate{
			Result:       buildFileResult(fileEntry, score),
			CategoryKind: resultKindFile,
		})
	}
	return candidates
}

func (app *App) folderProviderCandidates(request searchRequest) []rankedCandidate {
	if len(request.RawQuery) < 2 || app.config.DisableFolderIndex {
		return nil
	}
	if app.fileIdx.Status().State != "ready" {
		return nil
	}

	allUsageScores := app.usage.AllScores()
	folderUsageScores := make(map[string]int, len(allUsageScores))
	for usageKey, usageScore := range allUsageScores {
		if strings.HasPrefix(usageKey, "dir-open:") {
			folderUsageScores[strings.TrimPrefix(usageKey, "dir-open:")] = usageScore
		}
	}

	var candidates []rankedCandidate
	for _, folderEntry := range app.fileIdx.SearchDirs(request.RawQuery, folderUsageScores) {
		score := rankMatch(
			request.NormalizedQuery,
			[]string{folderEntry.Name, folderEntry.Path},
			nil,
			folderUsageScores[folderEntry.Path],
			false,
			0,
		)
		if score == 0 {
			continue
		}
		candidates = append(candidates, rankedCandidate{
			Result:       buildFolderResult(folderEntry, score),
			CategoryKind: resultKindFolder,
		})
	}
	return candidates
}

func (app *App) clipboardProviderCandidates(request searchRequest) []rankedCandidate {
	if len(request.RawQuery) < 2 {
		return nil
	}

	var candidates []rankedCandidate
	for entryIndex, clipboardEntry := range app.clipboard.Entries() {
		score := rankMatch(
			request.NormalizedQuery,
			[]string{clipboardEntry.Content},
			[]string{"clipboard", "copy"},
			0,
			false,
			0,
		)
		if score == 0 {
			continue
		}
		candidates = append(candidates, rankedCandidate{
			Result:       buildClipboardResult(clipboardEntry, entryIndex, score),
			CategoryKind: resultKindClipboard,
		})
	}
	return candidates
}

func (app *App) calcProviderCandidates(request searchRequest) []rankedCandidate {
	if !commands.IsCalcQuery(request.RawQuery) {
		return nil
	}
	calcResult := commands.Evaluate(request.RawQuery)
	if !calcResult.Valid {
		return nil
	}
	return []rankedCandidate{{
		Result:       buildCalcResult(calcResult),
		CategoryKind: resultKindCalc,
	}}
}

func (app *App) urlProviderCandidates(request searchRequest) []rankedCandidate {
	if !isURL(request.RawQuery) {
		return nil
	}
	return []rankedCandidate{{
		Result: SearchResult{
			ID:                 "url-open:" + request.RawQuery,
			Title:              "Open URL",
			Subtitle:           request.RawQuery,
			Category:           "Web",
			Kind:               resultKindWeb,
			Score:              15000,
			PrimaryActionLabel: "Open",
		},
		CategoryKind: resultKindWeb,
	}}
}

func (app *App) defaultHomeResults() []SearchResult {
	var results []SearchResult

	for _, pinnedID := range app.config.PinnedItems {
		switch {
		case strings.HasPrefix(pinnedID, "command:"):
			commandID := strings.TrimPrefix(pinnedID, "command:")
			commandDefinition, found := app.findCommand(commandID)
			if !found {
				continue
			}
			results = append(results, buildCommandResult(commandDefinition, "", "Pinned", 0))
		default:
			for _, application := range app.scanner.Apps() {
				if application.Name != pinnedID {
					continue
				}
				results = append(results, buildAppResult(application, "Pinned", 0))
				break
			}
		}
	}

	allUsageScores := app.usage.AllScores()
	var rankedApps []SearchResult
	for _, application := range app.scanner.Apps() {
		if slices.Contains(app.config.PinnedItems, application.Name) {
			continue
		}
		usageScore := allUsageScores[application.Name]
		if usageScore == 0 {
			continue
		}
		rankedApps = append(rankedApps, buildAppResult(application, "Recent Apps", usageScore))
	}
	sort.SliceStable(rankedApps, func(leftIndex, rightIndex int) bool {
		return rankedApps[leftIndex].Score > rankedApps[rightIndex].Score
	})
	if len(rankedApps) > 5 {
		rankedApps = rankedApps[:5]
	}
	results = append(results, rankedApps...)

	var rankedCommands []SearchResult
	for _, commandDefinition := range app.commandDefinitions() {
		pinnedID := commandPinnedID(commandDefinition.ID)
		if slices.Contains(app.config.PinnedItems, pinnedID) {
			continue
		}
		usageScore := allUsageScores[pinnedID]
		if usageScore == 0 {
			continue
		}
		rankedCommands = append(rankedCommands, buildCommandResult(commandDefinition, "", "Recent Commands", usageScore))
	}
	sort.SliceStable(rankedCommands, func(leftIndex, rightIndex int) bool {
		return rankedCommands[leftIndex].Score > rankedCommands[rightIndex].Score
	})
	if len(rankedCommands) > 4 {
		rankedCommands = rankedCommands[:4]
	}
	results = append(results, rankedCommands...)

	clipboardEntries := app.clipboard.Entries()
	if len(clipboardEntries) > 3 {
		clipboardEntries = clipboardEntries[:3]
	}
	for entryIndex, clipboardEntry := range clipboardEntries {
		results = append(results, buildClipboardResult(clipboardEntry, entryIndex, 0))
		results[len(results)-1].Category = "Clipboard"
	}

	return results
}

func (app *App) searchPathResults(query string) []SearchResult {
	pathResults := app.searchPath(query)
	for index := range pathResults {
		switch {
		case strings.HasPrefix(pathResults[index].ID, "dir-open:"):
			pathResults[index].Kind = resultKindFolder
			pathResults[index].PrimaryActionLabel = "Open"
			pathResults[index].SecondaryActionLabel = "Open in Terminal"
			pathResults[index].SupportsActions = true
		case strings.HasPrefix(pathResults[index].ID, "file-open:"):
			pathResults[index].Kind = resultKindFile
			pathResults[index].PrimaryActionLabel = "Open"
			pathResults[index].SecondaryActionLabel = revealLabel()
			pathResults[index].SupportsActions = true
		default:
			pathResults[index].Kind = resultKindFile
		}
	}
	return pathResults
}
