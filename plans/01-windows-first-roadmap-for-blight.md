# Windows-First Flow/Raycast-Inspired Roadmap for Blight

## Summary
- Target a 2-4 week Windows-first roadmap with a balanced split between launcher polish and an extension foundation.
- Do not build a full plugin marketplace or external runtime yet. Build a command system and provider architecture that makes a later plugin model straightforward.
- Prioritize five outcomes: meaningful home view, better ranking, richer actions, user-defined commands/quicklinks/snippets, and a clearer command-palette UX.

## Scope
- In scope: command mode, user commands, command/search ranking overhaul, real home view, richer actions, settings/supporting migrations, and UI polish.
- Out of scope: extension marketplace, cloud sync, AI features, deep macOS parity, large preview pane, networked integrations beyond URL-based quicklinks.

## Implementation Plan
1. Replace the current “blank spotlight” idle state with a real home view.
- Change `frontend/src/main.ts` so `loadDefaultResults()` calls `Search("")` and renders grouped default results instead of leaving the launcher visually empty.
- Change `app.go:getDefaultResults()` to return four sections in this order: `Pinned`, `Recent Apps`, `Recent Commands`, `Clipboard`.
- Default counts: pinned all, recent apps 5, recent commands 4, clipboard 3.
- Success criteria: opening the launcher with no query always shows useful content and never an empty panel.

2. Add a first-class command system instead of treating aliases as a side feature.
- Introduce a new config-backed type `CommandDefinition` with fields: `ID`, `Title`, `Keyword`, `Description`, `ActionType`, `Template`, `Keywords`, `Icon`, `RequiresArgument`, `RunAsAdmin`, `Pinned`.
- Supported `ActionType` values: `open_url`, `copy_text`, `open_path`, `run_shell`.
- Add Wails methods in `app.go`: `GetCommands()`, `SaveCommand(cmd)`, `DeleteCommand(id)`.
- Keep `Aliases` as a legacy input only. On config load, migrate each alias into a `CommandDefinition` and persist the new shape on next save.
- Ship built-in commands inspired by Raycast/Flow quicklinks: `g`, `gh`, `yt`, `wiki`, `maps`.
- Template substitution is exactly `{{query}}`. If `RequiresArgument=true` and the query part is empty, render the result but block execution and show subtitle `Type an argument`.

3. Introduce explicit command mode.
- Reserve `>` as the command prefix.
- Query behavior: `>` shows all commands, `>g openai` filters commands and previews the resolved action, `g openai` also works as a keyword command without the `>` prefix.
- Command-mode results use category `Commands` and sort above apps/files for the same query.
- Frontend requirement: add a visible “command mode” treatment in the search field and footer hints when the query starts with `>`.

4. Refactor search into providers and a single ranking pass.
- Create internal provider interfaces for `apps`, `commands`, `system`, `files`, `folders`, `clipboard`, `web`.
- Each provider returns a common ranked result model before the final merge.
- Replace the current append-by-category behavior in `app.go:Search()` with a global ranking pass and then apply category caps.
- Scoring formula must combine exact match, prefix match, substring match, fuzzy score, keyword hit, usage score, and pin boost.
- Default caps: Commands 6, Applications 8, Files 6, Folders 4, Clipboard 6, System 5, Web 1.
- Lower file search threshold from 3 chars to 2 chars after ranking is in place.

5. Expand the result/action model so actions feel deliberate instead of incidental.
- Extend `SearchResult` with `Kind`, `Score`, `PrimaryActionLabel`, `SecondaryActionLabel`, and `SupportsActions`.
- Extend `ContextAction` with `Shortcut` and `Destructive`.
- Standardize result kinds: `app`, `command`, `file`, `folder`, `clipboard`, `system`, `web`, `calc`.
- Standardize default actions by kind. Example: app = `Open`, `Run as Administrator`, `Reveal`, `Copy Path`, `Pin/Unpin`; command = `Run`, `Edit`, `Duplicate`, `Pin/Unpin`; clipboard = `Copy`, `Delete`.
- Keep `Enter` as primary and `Ctrl+Enter` as secondary. Make `Tab` always open the action panel for the selected row.

6. Upgrade custom commands UI in Settings.
- Replace the current aliases UI with a `Commands` tab in `frontend/src/modules/settings.ts`.
- The tab must support create, edit, duplicate, delete, and pin.
- Form fields must expose: title, keyword, type, template, keywords, requires-argument, run-as-admin.
- For `open_url`, validate that the template contains either a valid absolute URL or `{{query}}`.
- For `run_shell`, show a Windows-only warning that execution is local and unsandboxed.

7. Tighten ranking and recency behavior across the existing launcher features.
- Continue using `internal/search/usage.go`, but record usage for commands and system actions as well as apps/files/folders.
- Pinned items should get a deterministic boost large enough to keep them in the top three unless there is an exact match for something else.
- Calculator results should only appear when the query is clearly mathematical; remove the frontend-side `Function(...)` preview and use backend evaluation only to avoid divergence from `internal/commands/calculator.go`.

8. Polish the launcher UX to feel closer to Raycast quality without redesigning the whole app.
- Keep the current visual language, but add a real empty/home state hierarchy, clearer selected-row affordances, and better footer action hints.
- On a selected result, always show the relevant footer hints: `Enter`, `Ctrl+Enter`, `Tab`.
- Preserve async icons, but ensure home-view pinned/recent results also request icons.
- Keep the current window layout; do not add a right-side preview pane in this phase.

## Public API / Interface Changes
- `SearchResult` gains: `Kind`, `Score`, `PrimaryActionLabel`, `SecondaryActionLabel`, `SupportsActions`.
- `ContextAction` gains: `Shortcut`, `Destructive`.
- `BlightConfig` gains: `Commands []CommandDefinition`.
- New exported Wails methods: `GetCommands()`, `SaveCommand(CommandDefinition)`, `DeleteCommand(id string)`.
- `Aliases` remains readable for migration but is no longer the primary authoring surface.

## File/Module Touchpoints
- Backend: `app.go`, `internal/search/*`, `internal/commands/*`, plus a new provider/ranking layer.
- Frontend: `frontend/src/main.ts`, `frontend/src/modules/settings.ts`, `frontend/src/modules/context-menu.ts`, and styles in `frontend/src/style.css`.
- Tests: existing Go search/calculator tests and frontend Vitest modules should be expanded rather than replaced.

## Tests and Acceptance Scenarios
- `Search("")` returns grouped home results with at least pinned/recent sections when data exists.
- `g cats` resolves the GitHub or Google quicklink command correctly depending on keyword.
- `>yt synthwave` and `yt synthwave` both produce a runnable command result.
- Legacy aliases migrate into commands exactly once and still work after restart.
- Pinned commands/apps outrank non-pinned fuzzy matches.
- File search returns relevant results for 2-character queries without flooding the list.
- `Tab` opens the action panel for every actionable result kind.
- `Ctrl+Enter` triggers the defined secondary action for apps, folders, commands, and clipboard items.
- Calculator preview and executed result always match because both use backend evaluation.
- Settings CRUD for commands persists correctly and updates search behavior without restart.

## Assumptions and Defaults
- Windows-first is the product stance for this phase.
- No third-party plugin runtime will be built now; the provider/command architecture is the extension foundation.
- Existing hotkey, tray, updater, clipboard polling, and file indexing remain in place and must not regress.
- User-created shell commands are allowed, but only through explicit settings authoring and only on Windows in this phase.
- The launcher keeps its current overall layout and theme system; this phase is UX refinement, not a visual rewrite.
