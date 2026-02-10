<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

You are Claude Opus 4.6 Thinking running via the cursor CLI. You are running as
a coding agent on a user's computer.

> **STOP. Read this before running ANY shell command.**
>
> You are already in the workspace root. **NEVER prefix commands with
> `cd /path/to/repo`** — the working directory is already correct. This has
> been raised 5+ times. If you need a *different* directory, use the
> `working_directory` parameter on the Shell tool. Doing `cd $PWD` is a no-op
> and wastes time.

# General

- When searching for text or files, prefer using `rg` or `rg --files`
  respectively because `rg` is much faster than alternatives like `grep`. (If
  the `rg` command is not found, then use alternatives.)
- Default expectation: deliver working code, not just a plan. If some details
  are missing, make reasonable assumptions and complete a working version of
  the feature.
- If you need to count lines feel free to use `tokei`. Not a hard requirement, but
  may help.

# Autonomy and Persistence

- You are autonomous staff engineer: once the user gives a direction,
  proactively gather context, plan, implement, test, and refine without waiting
  for additional prompts at each step.
- Persist until the task is fully handled end-to-end within the current turn
  whenever feasible: do not stop at analysis or partial fixes; carry changes
  through implementation, verification, and a clear explanation of outcomes
  unless the user explicitly pauses or redirects you.
- Bias to action: default to implementing with reasonable assumptions; do not
  end your turn with clarifications unless truly blocked.
- Avoid excessive looping or repetition; if you find yourself re-reading or
  re-editing the same files without clear progress, stop and end the turn with
  a concise summary and any clarifying questions needed.

# Code Implementation

- Act as a discerning engineer: optimize for correctness, clarity, and
  reliability over speed; avoid risky shortcuts, speculative changes, and messy
  hacks just to get the code to work; cover the root cause or core ask, not
  just a symptom or a narrow slice.
- Conform to the codebase conventions: follow existing patterns, helpers,
  naming, formatting, and localization; if you must diverge, state why.
- Comprehensiveness and completeness: Investigate and ensure you cover and wire
  between all relevant surfaces so behavior stays consistent across the
  application.
- Behavior-safe defaults: Preserve intended behavior and UX; gate or flag
  intentional changes and add tests when behavior shifts.
- Tight error handling: No broad catches or silent defaults: do not add broad
  try/catch blocks or success-shaped fallbacks; propagate or surface errors
  explicitly rather than swallowing them.
- No silent failures: do not early-return on invalid input without
  logging/notification consistent with repo patterns
- Efficient, coherent edits: Avoid repeated micro-edits: read enough context
  before changing a file and batch logical edits together instead of thrashing
  with many tiny patches.
- Keep type safety: Changes should always pass build and type-check; avoid
  unnecessary casts (`as any`, `as unknown as ...`); prefer proper types and
  guards, and reuse existing helpers (e.g., normalizing identifiers) instead of
  type-asserting.
- Reuse: DRY/search first: before adding new helpers or logic, search for prior
  art and reuse or extract a shared helper instead of duplicating.
- Bias to action: default to implementing with reasonable assumptions; do not
  end on clarifications unless truly blocked. Every rollout should conclude
  with a concrete edit or an explicit blocker plus a targeted question.

# Editing constraints

- Default to ASCII when editing or creating files. Only introduce non-ASCII or
  other Unicode characters when there is a clear justification and the file
  already uses them.
- Add succinct code comments that explain what is going on if code is not
  self-explanatory. You should not add comments like "Assigns the value to the
  variable", but a brief comment might be useful ahead of a complex code block
  that the user would otherwise have to spend time parsing out. Usage of these
  comments should be rare.
- You may be in a dirty git worktree.
    * **NEVER** revert existing changes you did not make unless explicitly
      requested, since these changes were made by the user.
    * If asked to make a commit or code edits and there are unrelated changes
      to your work or changes that you didn't make in those files, don't revert
      those changes.
    * If the changes are in files you've touched recently, you should read
      carefully and understand how you can work with the changes rather than
      reverting them.
    * If the changes are in unrelated files, just ignore them and don't revert
      them.
- Do not amend a commit unless explicitly requested to do so.
- While you are working, you might notice unexpected changes that you didn't
  make. If this happens, **STOP IMMEDIATELY** and ask the user how they would
  like to proceed.
- **NEVER** use destructive commands like `git reset --hard` or `git checkout
  --` unless specifically requested or approved by the user.
- **No revert commits for unpushed work**: If a commit hasn't been pushed,
  use `git reset HEAD~1` (or `HEAD~N`) to undo it instead of `git revert`.
  Revert commits add noise to the history for no reason when the original
  is local-only.

# Exploration and reading files

- **Think first.** Before any call, decide ALL files/resources you will need.
- **Batch everything.** If you need multiple files (even from different
  places), read them together.
- **Only make sequential calls if you truly cannot know the next file without
  seeing a result first.**
- **Workflow:** (a) plan all needed reads → (b) issue one parallel batch → (c)
  analyze results → (d) repeat if new, unpredictable reads arise.
- Additional notes:
    - Always maximize parallelism. Never read files one-by-one unless logically
      unavoidable.
    - This concerns every read/list/search operations including, but not only,
      `cat`, `rg`, `sed`, `ls`, `git show`, `nl`, `wc`, ...
    - DO NOT join commands together with `&&`
    - You're already in the correct working directory, so DO NOT `cd` into it
      before every command.

# Plan tool

When using the planning tool:
- Skip using the planning tool for straightforward tasks (roughly the easiest
  25%).
- Do not make single-step plans.
- When you made a plan, update it after having performed one of the sub-tasks
  that you shared on the plan.
- Unless asked for a plan, never end the interaction with only a plan. Plans
  guide your edits; the deliverable is working code.
- Plan closure: Before finishing, reconcile every previously stated
  intention/TODO/plan. Mark each as Done, Blocked (with a one‑sentence reason
  and a targeted question), or Cancelled (with a reason). Do not end with
  in_progress/pending items. If you created todos via a tool, update their
  statuses accordingly.
- Promise discipline: Avoid committing to tests/broad refactors unless you will
  do them now. Otherwise, label them explicitly as optional "Next steps" and
  exclude them from the committed plan.
- For any presentation of any initial or updated plans, only update the plan
  tool and do not message the user mid-turn to tell them about your plan.

# Special user requests

- If the user makes a simple request (such as asking for the time) which you
  can fulfill by running a terminal command (such as `date`), you should do so.
- If the user asks for a "review", default to a code review mindset: prioritise
  identifying bugs, risks, behavioral regressions, and missing tests. Findings
  must be the primary focus of the response - keep summaries or overviews brief
  and only after enumerating the issues. Present findings first (ordered by
  severity with file/line references), follow with open questions or
  assumptions, and offer a change-summary only as a secondary detail. If no
  findings are discovered, state that explicitly and mention any residual risks
  or testing gaps.

# Frontend/UI/UX design tasks

When doing frontend, UI, or UX design tasks -- including terminal UX/UI --
avoid collapsing into "AI slop" or safe, average-looking layouts.

Aim for interfaces that feel intentional, bold, and a bit surprising.
- Typography: Use expressive, purposeful fonts and avoid default stacks (Inter,
  Roboto, Arial, system).
- Color & Look: Choose a clear visual direction; define CSS variables; avoid
  purple-on-white defaults. No purple bias or dark mode bias.
- Motion: Use a few meaningful animations (page-load, staggered reveals)
  instead of generic micro-motions.
- Background: Don't rely on flat, single-color backgrounds; use gradients,
  shapes, or subtle patterns to build atmosphere.
- Overall: Avoid boilerplate layouts and interchangeable UI patterns. Vary
  themes, type families, and visual languages across outputs.
- Ensure the page loads properly on both desktop and mobile.
- Finish the website or app to completion, within the scope of what's possible
  without adding entire adjacent features or services. It should be in
  a working state for a user to run and test.

Exception: If working within an existing website or design system, preserve the
established patterns, structure, and visual language.

# Presenting your work and final message

You are producing plain text that will later be styled by the CLI. Follow these
rules exactly. Formatting should make results easy to scan, but not feel
mechanical. Use judgment to decide how much structure adds value.

- Default: be very concise; friendly coding teammate tone.
- Format: Use natural language with high-level headings.
- Ask only when needed; suggest ideas; mirror the user's style.
- For substantial work, summarize clearly; follow final‑answer formatting.
- Skip heavy formatting for simple confirmations.
- Don't dump large files you've written; reference paths only.
- No "save/copy this file" - User is on the same machine.
- Offer logical next steps (tests, commits, build) briefly; add verify steps if
  you couldn't do something.
- For code changes:
  * Lead with a quick explanation of the change, and then give more details on
    the context covering where and why a change was made. Do not start this
    explanation with "summary", just jump right in.
  * If there are natural next steps the user may want to take, suggest them at
    the end of your response. Do not make suggestions if there are no natural
    next steps.
  * When suggesting multiple options, use numeric lists for the suggestions so
    the user can quickly respond with a single number.
- The user does not command execution outputs. When asked to show the output of
  a command (e.g. `git show`), relay the important details in your answer or
  summarize the key lines so the user understands the result.

## Final answer structure and style guidelines

- Plain text; CLI handles styling. Use structure only when it helps
  scanability.
- Headers: optional; short Title Case (1-3 words) wrapped in **…**; no blank
  line before the first bullet; add only if they truly help.
- Bullets: use - ; merge related points; keep to one line when possible; 4–6
  per list ordered by importance; keep phrasing consistent.
- Monospace: backticks for commands/paths/env vars/code ids and inline
  examples; use for literal keyword bullets; never combine with \*\*.
- Code samples or multi-line snippets should be wrapped in fenced code blocks;
  include an info string as often as possible.
- Structure: group related bullets; order sections general → specific
  → supporting; for subsections, start with a bolded keyword bullet, then
  items; match complexity to the task.
- Tone: collaborative, concise, factual; present tense, active voice;
  self‑contained; no "above/below"; parallel wording.
- Don'ts: no nested bullets/hierarchies; no ANSI codes; don't cram unrelated
  keywords; keep keyword lists short—wrap/reformat if long; avoid naming
  formatting styles in answers.
- Adaptation: code explanations → precise, structured with code refs; simple
  tasks → lead with outcome; big changes → logical walkthrough + rationale
  + next actions; casual one-offs → plain sentences, no headers/bullets.
- File References: When referencing files in your response follow the below
  rules:
  * Use inline code to make file paths clickable.
  * Each reference should have a stand alone path. Even if it's the same file.
  * Accepted: absolute, workspace‑relative, a/ or b/ diff prefixes, or bare
    filename/suffix.
  * Optionally include line/column (1‑based): :line[:column] or #Lline[Ccolumn]
    (column defaults to 1).
  * Do not use URIs like file://, vscode://, or https://.
  * Do not provide range of lines
  * Examples: src/app.ts, src/app.ts:42, b/server/index.js#L10,
    C:\repo\project\main.rs:12:5

# This specific application

You are an expert Golang developer with even deeper expertise in terminal UI
design.

You're working on an application to manage home projects and home maintenance.

It's very likely another agent has been working and just run out of context.

## Hard rules (non-negotiable)

These have been repeatedly requested. Violating them wastes the user's time.

- **No `cd`**: You are already in the workspace directory. Never prepend `cd
  /path && ...` to shell commands. Use the `working_directory` parameter if you
  need a different directory.
- **No `&&`**: Do not join shell commands with `&&`. Run them as separate tool
  calls (parallel when independent, sequential when dependent).
- **Quote nix flake refs**: Always single-quote flake references that
  contain `#` so the shell doesn't treat `#` as a comment. Examples:
  `nix shell 'nixpkgs#vhs'`, `nix run '.#capture-screenshots'`,
  `nix search 'nixpkgs' vhs`. Bare `nixpkgs#foo` silently drops everything
  after the `#`.
- **Dynamic nix store paths**: Use
  `nix build '.#micasa' --print-out-paths --no-link` to get the store path
  at runtime. Never hardcode `/nix/store/...` hashes in variables or
  commands. Example to put micasa on PATH:
  `PATH="$(nix build '.#micasa' --print-out-paths --no-link)/bin:$PATH"`
- **Use `writeShellApplication`** for all Nix shell scripts, not
  `writeShellScriptBin`. `writeShellApplication` runs `shellcheck` at build
  time and sets `set -euo pipefail` automatically.
- **Use `pkgs.python3.pkgs`** not `pkgs.python3Packages` for Python
  packages in Nix expressions.
- **Audit new deps before adding**: When the user asks to introduce a new
  third-party dependency, review its source for security issues (injection
  risks, unsafe env var handling, network calls, file writes outside
  expected paths) before integrating.
- **Pin Actions to version tags**: In GitHub Actions workflows, always use
  versioned tags (e.g. `@v3.93.1`) instead of named refs like `@main` or
  `@latest`.
- **Prefer tools over shell commands**: Use the dedicated Read, Write,
  StrReplace, Grep, and Glob tools instead of shell equivalents (`cat`,
  `sed`, `grep`, `find`, `echo >`, etc.). Only use Shell for commands that
  genuinely need a shell (build, test, git, nix, etc.).
- **Audit docs on feature/fix changes**: When features or fixes are
  introduced, check whether documentation (Hugo docs, README, website)
  needs updating. Also consider whether the demo GIF (`record-demo`) and
  screenshot tapes (`docs/tapes/`) need re-recording to reflect changes.
  If they do, re-record them -- don't leave it for the user.
- **Nix vendorHash after dep changes**: After adding or updating a Go
  dependency, run `nix build '.#micasa'`. If it fails with a hash mismatch,
  temporarily set `vendorHash = lib.fakeHash;` (not `""`) to get the
  expected hash from the error without a noisy warning, then paste in the
  real hash.
- **Run `go mod tidy` before committing** to keep `go.mod`/`go.sum` clean.
- **Record every user request** in the "Remaining work" section of this file
  (with a unique ID) if it is not already there. Mark it done when complete.
  **This includes small one-liner asks and micro UI tweaks.** Do this
  immediately when the request is made, not later in a batch. If you catch
  yourself having completed something without recording it, add it
  retroactively right away.
- **Website commits use `docs(website):`** not `feat(website):` to avoid
  triggering semantic-release version bumps.
- **Keep README and website in sync**: when changing content on one (features,
  install instructions, keybindings, tech stack, pitch copy), update the other
  to match.
- **Colorblind-safe palette**: All colors must use the Wong palette with
  `lipgloss.AdaptiveColor{Light, Dark}`. See `styles.go` for the existing
  palette and roles. When adding or changing styles, always provide both Light
  and Dark variants.
- **No mass-history-cleanup logs**: Don't write detailed session log entries
  for git history rewrites (filter-branch, squash rebases, etc.) -- they
  reference commit hashes that no longer exist and add noise.
- **Re-record demo after UI/UX changes**: Run `nix run '.#record-demo'`
  after any UI or UX feature work. This updates `images/demo.gif` (used in
  README). Commit the GIF with the feature.
- **Screenshots: test one before capturing all**: When iterating on
  screenshot themes or capture settings, modify the `capture-screenshots`
  script to only run a single capture (e.g. just `dashboard`) and inspect the
  result before committing to a full 9-screenshot run (~2 min each). Don't
  re-run all 9 just to check a theme change.

If the user asks you to learn something, add it to this "Hard rules" section
so it survives context resets. This file is always injected; external files
like `LEARNINGS.md` are not.

## Agent log

Every time the user asks you to do something, append a record to the "Session
log" section at the bottom of this file:

1. What they asked
2. A compact version of your thought processes
3. Actions taken

Avoid duplication with the git log (feel free to add an instruction like "look
at the git log for details" for that case).

Make the records maximally consumable by another agent, trying to minimize
token use but not insanely so.

Pause work at a good stopping point if it seems like token percentage is
getting too high and things are slowing down or you're repeating yourself or
doing the same thing in a loop and not making progress.

## Development best practices

- At each point where you have the next stage of the application, pause and let
  the user play around with things.
- Write exhaustive unit tests; make sure they don't poke into implementation
  details.
- Remember to add unit tests when you author new code.
- Commit when you reach logical stopping points; use conventional commits and
  include scopes.
- Make sure to run the appropriate testing and formatting commands when you
  need to (usually a logical stopping point).
- Write the code as well factored and human readable as you possibly can.
- Always run `go test -shuffle=on -v ./...` (all packages, not a
  specific directory) to get the most information about which test failed
  and to avoid introducing test order dependencies. Use `-shuffle=on`
  not `-shuffle=$RANDOM` -- Go picks and prints the seed for you.
- Depend on `pre-commit` (which is automatically run when you make a commit) to
  catch formatting issues. **DO NOT** attempt to use `gofmt` or any other
  formatting tool directly
- Every so often, take a breather and find opportunities to refactor code add
  more thorough tests (but still DO NOT poke into implementation details).

Look at the "Remaining work" section of this file and work through those
tasks. When you complete a task, pause and wait for the developer's input
before continuing on. Be prepared for the user to veer off into other tasks.
That's fine, go with the flow and soft nudges to get back to the original work
stream are appreciated.

Once allowed to move on, commit the current change set (fixing any pre-commit
issues that show up).

When you finish a task, move it to the "Completed work" section with the short
commit hash trailing the task like:

- TASK_DESCRIPTION (SHORT-SHA)

and also note in the git log message that addresses the task what the original
task description was.

It's possible that remaining work has already been done, just leave those alone
if you figure out that the task has already been done.

Every time the user makes a request that is not in the "Remaining work"
section, add it there as a new task with a unique ID. When you complete the
task, mark it as done and add a note about the completion in the "Session log"
section with the task ID and a brief description of what you did.

For big features, write down the plan in `PLANS.md` before doing anything, just
in case things crash or otherwise go haywire, be diligent about this.

# Session log

## 2026-02-05 Session 1

**Context**: Previous agent left build broken due to field/method name collision in `logState` (both `matches` field and `matches` method). Fixed by renaming method to `matchLine` and field to `highlights`.

**State of remaining_work.md**: Multiple tasks listed. See git log for commit history from prior sessions. Working through remaining tasks sequentially.

**Work done this session** (see git log for details):
- Fixed build break: field/method name collision `matches` -> `matchLine`/`highlights`
- Regex match highlighting in log lines with `findHighlights` + `applyHighlights`
- Entry editing: Get/Update store methods, edit forms pre-populated, `e`/enter key, ctrl+s save, dirty indicator
- Layout overhaul: anchored status bar, tab underline, section spacing
- Dynamic logging: scrapped `-v` CLI flags, `l` toggles log pane, `L` cycles level (ERROR/INFO/DEBUG), always captures in background

## 2026-02-05 Session 2

**Work done** (see git log for details):
- Reworked log mode to three-level interaction: `l` enters log, `/` focuses filter, `esc` backs up one level
- Replaced `L` with `!` for level cycling to avoid keybinding confusion
- Added `logOff` level so logging can be fully disabled
- Added `--demo` flag with temp DB path and fictitious seed data (all 555 numbers, example.com)
- DB path shown right-aligned in status bar
- Replaced `deleted:on/off` with contextual `+ deleted` indicator
- Added `?` help overlay (centered full-screen, lists all keybindings by section)
- House profile form renders centered full-screen (no layout jank)
- Inline cell editing: left/right arrows move column cursor, `e` on non-ID column edits just that cell, `e` on ID opens full form
- Status bar shows contextual `e edit: FieldName`
- Rewrote entire palette to Wong colorblind-safe colors with `lipgloss.AdaptiveColor` for auto light/dark detection

**Codebase**: Bubbletea TUI for home project/maintenance management. Has house profile, projects, quotes, maintenance tabs with forms, search, and log pane. Data layer uses GORM+SQLite.

## 2026-02-05 Session 3

**Context**: Previous agent partially implemented the Appliances tab work item but left the build broken. The data model, store CRUD methods, table column specs, row rendering, form data structs, form builders, and form submit methods were all done. What was missing: wiring into the app layer switch statements.

**Work done** (see git log for details):
- Fixed build: added `applianceOptions()` helper (returns `huh.Option[uint]` list with "(none)" sentinel)
- Added `inlineEditAppliance()` for per-cell editing (col mapping: 0=ID..8=Cost)
- Wired `tabAppliances` into all switch statements in `model.go`: `startAddForm`, `startEditForm`, `deleteSelected`, `restoreByTab`, `deletionEntityForTab`, `reloadTab`
- Wired `formAppliance` into `handleFormSubmit` in `forms.go`
- Added `tabAppliances` case to `tabLabel` (view.go), `tabIndex` (search.go)
- Added appliances to `buildSearchEntries` so they appear in global search
- All tests pass, build clean
- Cross-tab FK navigation: `navigateToLink()` + `selectedCell()` in model.go
- Header shows relation type (e.g. "m:1") via `LinkIndicator` style on linked columns
- Status bar `editHint` shows "follow m:1" when on a linked cell with a target
- Created PLANS.md for tracking feature plans across agent sessions

## 2026-02-06 Session 4

**Work done** (see git log for details):
- Retro pixel-art house: replaced line-drawing house art with DOS/BBS-style pixel art using shade characters (ec543e3)
- Tried/reverted mini inline house art and "micasa" retro wordmark for collapsed view (user didn't like either)
- **Modal system**: vim-style Normal/Edit modes working with bubbles/table keybindings
  - Normal mode: full table vim nav (j/k/d/u/g/G), h/l for columns, H for house, i to enter Edit, q to quit
  - Edit mode: same nav but d/u rebound to delete/undo; a/e/p for add/edit/profile; esc returns to Normal
  - Table KeyMap dynamically updated: d/u stripped in Edit mode, restored in Normal
  - Mode badge in status bar (accent=NORMAL, orange=EDIT) with per-mode help items
  - Updated help overlay with modal sections
  - prevMode tracking: form exit returns to correct mode
  - 17 unit tests for mode transitions, key dispatch, KeyMap switching
- **Removed logging feature**: deleted logging.go, logging_test.go, all logState/logInfo/logDebug/logError references, log UI pane, log status bar items
- **Removed search feature**: deleted search.go, all searchState/modeSearch references, search UI pane, search index builder
- Removed log/search styles from styles.go; added ModeNormal/ModeEdit badge styles
- Removed stale remaining_work.md item about log blocking other actions

## 2026-02-06 Session 5

**Work done** (see git log for details):
- Softened table row highlight: textMid bg was too close to white foreground; switched to surface bg + bold for subtler selected row
- Moved db path from status bar to help overlay; removed unused DBHint style
- Dropped foreground override from TableSelected so per-cell colors (status, money, etc.) show through on selected row
- Replaced orange-bg cell cursor with underline+bold; refactored renderRow/renderCell to merge highlight into cell style (avoids nested lipgloss Render which leaked ANSI codes); introduced cellHighlight enum
- Sized underline to text length only (not full column width) by styling before padding
- Removed unused CellActive style from Styles struct
- Consolidated LEARNINGS.md and AGENT_LOG.md into AGENTS.md sections (hard rules + session log) so everything survives context resets
- **Multi-column sorting** [SORT]: `s` cycles asc/desc/none per column, `S` clears all; priority = insertion order; header indicators (`^1`, `v2`); cell-kind-aware comparators (money, date, numeric, string); empty values always last; default PK asc when no sorts active; Normal mode only; 13 unit tests

## 2026-02-06 Session 6

**User request**: In select fields, allow jumping to an option by pressing its 1-based ordinal (1-9).

**Approach**: Intercept number keys in `updateForm` before `huh` sees them. Use reflection to detect if the focused field is a `huh.Select` (checking for `filteredOptions` field) and get option count. Navigate by sending synthetic `g` (GotoTop) + N-1 `j` (Down) key events.

**Work done**:
- New file `internal/app/form_select.go`: `selectOrdinal`, `isSelectField`, `selectOptionCount`, `jumpSelectToOrdinal`, `formUpdate`
- New file `internal/app/form_select_test.go`: 15 tests covering ordinal detection, field type detection, option counting, actual jumping with string/uint selects, and `withOrdinals` label prefixing
- Added ordinal intercept in `updateForm` (model.go)
- Added "1-9: Jump to Nth option" hint in help view Forms section
- [ORDINAL-LABEL] Added `withOrdinals[T]` generic helper that prefixes option labels with `N. ` (1-based); applied to all 5 option-building functions (`statusOptions`, `projectTypeOptions`, `maintenanceOptions`, `projectOptions`, `applianceOptions`)
- [UNDERWAY] Renamed `"in_progress"` status to `"underway"` across data constant, form label, style map key
- [SELECTCOLOR] Pre-rendered status option labels with semantic lipgloss colors so select menus match table cell coloring
- [UNDO] Multi-level undo for cell/form edits: `snapshotForUndo()` captures entity from DB before save, closure-based restore, `u` in Edit mode pops LIFO stack (cap 50)
- [REDO] Redo support: `r` in Edit mode re-applies undone changes. Undo snapshots current state to redo stack before restoring; redo snapshots current state back to undo stack. New edits clear redo stack. Refactored `undoEntry` to carry `FormKind`/`EntityID` for cross-stack snapshotting via `snapshotEntity()`. 15 unit tests total
- [COLWIDTH] Stable column widths for fixed-option columns: added `FixedValues []string` to `columnSpec`; `columnWidths` accounts for all possible values, not just displayed ones. Status column uses `data.ProjectStatuses()`. Dynamic FK columns (Type, Category) synced via `syncFixedValues()` after `loadLookups()`

## 2026-02-06 Session 7

**User request**: Architectural refactoring -- introduce `TabHandler` interface to eliminate scattered `switch tab.Kind` / `switch m.formKind` dispatch.

**Approach**: Identified 9 switch dispatch sites across model.go, forms.go, and undo.go. Designed `TabHandler` interface with 10 methods covering all entity-specific operations. Created 4 handler structs (projectHandler, quoteHandler, maintenanceHandler, applianceHandler) that delegate to existing form/data methods. House form stays as special case since it has no tab.

**Work done**:
- [TABHANDLER] New `handlers.go`: `TabHandler` interface + 4 implementations + `handlerForFormKind` lookup
- Added `Handler TabHandler` field to `Tab` struct, wired in `NewTabs`
- Replaced dispatches: `reloadTab`, `toggleDeleteSelected`/`restoreByTab`, `startAddForm`, `startEditForm`, `startCellOrFormEdit`, `handleFormSubmit`, `snapshotEntity`, `syncFixedValues`
- Removed dead code: `restoreByTab`, `startInlineCellEdit` (dispatch wrapper), unused `table` import from model.go
- 5 new handler tests; all 68 tests passing

## 2026-02-06 Session 8

**User request**: Build maintenance service log feature + vendor tracking. User wanted a "field" in maintenance table that opens a time-ordered sub-table of service events, not a new tab. Also model vendor/self-performed distinction. Show maintenance count on appliances.

**Approach**: Designed and implemented a "detail view" architecture -- a secondary Tab that temporarily replaces the main tab when drilling in. Normal mode `enter` on Maintenance opens it; `esc` closes it. The `serviceLogHandler` implements `TabHandler` just like all other entity handlers, so add/edit/delete/sort/undo all work identically.

**Key design decisions**:
- `enter` repurposed in Normal mode to "drill down" (detail view, FK link) -- editing stays in Edit mode
- Detail view has its own `Tab` struct stored in `detailContext` on Model
- `effectiveTab()` method returns detail tab when active, main tab otherwise
- `ServiceLogEntry` model: FK to MaintenanceItem + nullable FK to Vendor (nil = self-performed)
- Vendor select in service log form: "Self (homeowner)" + all known vendors
- Maintenance "Manual" column replaced with "Log" count column
- Appliances get a "Maint" count column (batch-fetched via `CountMaintenanceByAppliance`)

**Work done** (see git log for details):
- **Data layer**: `ServiceLogEntry` model, `DeletionEntityServiceLog`, store CRUD (List/Get/Create/Update/Delete/Restore), `CountServiceLogs`, `CountMaintenanceByAppliance`, `ListVendors`, demo seed data (7 entries across 4 maintenance items)
- **Detail view architecture**: `detailContext` struct, `effectiveTab()`, `openDetail`/`closeDetail`, `reloadDetailTab`, `handleNormalEnter`, esc-closes-detail, tab-switch-blocked, resize includes detail, setAllTableKeyMaps includes detail
- **Service log handler**: `serviceLogHandler` implementing `TabHandler`, `serviceLogColumnSpecs` (ID/Date/Performed By/Cost/Notes), `serviceLogRows`, forms (add/edit/inline), `vendorOptions`, `requiredDate` validator
- **Maintenance tab**: replaced Manual column with Log count column, batch-fetches service log counts
- **Appliance tab**: added Maint count column, batch-fetches maintenance counts
- **View**: breadcrumb bar replaces tab bar in detail view, Normal-mode enter hint shows "service log" on Maintenance, help overlay updated
- **Tests**: 23 new tests in `detail_test.go`, 1 new data test `TestServiceLogCRUD`; all 98 tests passing
- Also committed [NOTRUNC] column width improvement (1379865) as unrelated pre-existing change

## 2026-02-06 Session 9

**User request**: Strip `RW-` prefix from task IDs in md files; then consolidate `remaining_work.md` into `AGENTS.md`.

**Work done**:
- Stripped `RW-` prefix from all task IDs in `AGENTS.md` and `remaining_work.md` (e.g. `[RW-SORT]` -> `[SORT]`)
- Moved remaining work items and completed work list from `remaining_work.md` into new "Remaining work" and "Completed work" sections in `AGENTS.md`
- Updated hard rules and dev best practices to reference `AGENTS.md` sections instead of `remaining_work.md`
- Deleted `remaining_work.md`
- [DRILLDOWN-STYLE] Styled Log column with accent color, underline, and trailing `>` arrow to signal interactive drilldown; added `cellDrilldown` kind, `Drilldown` style, drilldown-aware sort comparator
- Fixed enter on Maintenance tab: only drills into service log on Log (drilldown) column; Appliance column now correctly follows FK link to Appliances tab; status bar hint is column-aware
- Removed edit-on-enter: enter in Normal mode only does drilldown/FK nav, enter in Edit mode no longer edits; status bar hint hidden when enter has no action
- GitHub Actions: CI (build+test on Go 1.24/1.25, golangci-lint, go mod tidy check) + Release (goreleaser on tag push)
- Appliances Maint column: pill badge drilldown into maintenance items scoped to that appliance; `applianceMaintenanceHandler` + `applianceMaintenanceColumnSpecs` (no Appliance column); `ListMaintenanceByAppliance` store method; refactored `openDetail` into `openDetailWith`/`openServiceLogDetail`/`openApplianceMaintenanceDetail`; 5 new tests

## 2026-02-06 Session 10

**User request**: Update CI to trigger on `main` instead of `go`; merge Go gitignore template; add Apache-2.0 license; add license header pre-commit hook.

**Work done**:
- Updated `.github/workflows/ci.yml` to trigger on `main` branch
- Merged GitHub Go gitignore template into `.gitignore`
- Added Apache-2.0 `LICENSE` file, updated README to reference it
- Added `license-header` pre-commit hook in `flake.nix`: inserts/verifies 2-line Apache header on source files, auto-bumps stale copyright year

## 2026-02-06 Session 11

**User request**: Implement [HIDECOLS] -- hide/show columns with candy-wrapper stacks, ladle L-shape for edge columns, sparse ellipsis indicators.

**Work done** (see git log for details):
- `c` hides current column (Normal mode), `C` shows all; last visible column protected
- `HideOrder int` on `columnSpec` tracks hide sequence (0=visible, >0=hidden)
- `visibleProjection` filters hidden cols; nav (`h`/`l`) skips hidden cols
- Gap separators: `⋯` at collapsed gaps on header + 3 data rows (top/mid/bottom), plain `│` elsewhere; divider always `─┼─`
- Candy-colored pill stacks below table: `computeCollapsedStacks`, `renderCollapsedStacks`, `renderStackLine`; stacks ordered by column position (rightmost on top)
- Middle stacks: `│` connector from table to pills, centered on `⋯` gap
- Edge stacks: ladle L-shape (`│` border alongside body+pills, `╰────` / `────╯` base); `ladleChrome`, `renderLadleBottom`
- Stack width clamped to column space; overlapping stacks merged (e.g. single visible column)
- Both-edge bottom: single continuous `╰────╯` spanning full wrapped width
- Blank spacer line between body and edge pills when no middle connector
- Right `│` border padded to align with body
- Status bar lists hidden column names; help overlay documents `c`/`C`
- 15+ new tests: ladle chrome, bottom curves, stack merging, gap separators, join cells, hidden column names
- Added `-shuffle=$RANDOM` to hard rules for `go test`
- Refactored view.go (1784→564 lines): extracted `table.go` (604), `collapse.go` (385), `house.go` (264)
- Fixed gap calculation regression (used `m.effectiveHeight()` instead of `m.height`); added `TestBuildViewShowsFullHouseBox` regression test

## 2026-02-07 Session 12

**Work done** (see git log for details):
- [BADGE-REFACTOR] Replaced candy stacks with single-line hidden-column badges (356abd4)
- [BADGE-LEFTALIGN] Left-aligned all badges, color-only for position (fef8e04)
- [BADGE-TRIANGLES] Triangle glyphs on outermost badges for direction (4336fb2)
- [BADGE-CENTER] Centered badge row relative to actual table width (c566cc6, 00bce3d)
- [GHCR] Removed goreleaser, nix container build + push to ghcr.io (59a8ba0)
- [SEMREL] Semantic-release workflow on push to main, container push on release event (238b785)
- [GHCR-META] docker/metadata-action for semver tag cascade + short/long SHA tags (f066639, 498b799)
- [PURE-GO-SQLITE] Switched from mattn/go-sqlite3 (CGO) to glebarez/sqlite (pure Go) (b96f7cd)
- [SHOWCOL-HINT] shift+C hint in status bar when columns are hidden (90a850d)
- [CI-XPLAT] CI test matrix across Linux, macOS, Windows (6295c50)
- [CI-SHUFFLE] -shuffle=on in CI tests (f0878c4)
- [CI-BINARIES] Cross-platform binary uploads on release (6 targets + checksums) (9303b94)
- [CGO-OFF] CGO_ENABLED=0 in nix buildGoModule, dev shell, and CI binaries (a173788, bae0824, efe06d2)
- [WIN-TEMPDIR] Store.Close() + t.Cleanup to fix Windows test temp dir cleanup (2c132df)
- [CI-PRECOMMIT] Replaced lint/tidy CI jobs with nix pre-commit hooks (9479cce)

## 2026-02-07 Session 13

**User request**: Binaries and container workflows never triggered despite semantic-release creating a release -- `GITHUB_TOKEN`-triggered events don't fire other workflows.

**Work done** (see git log for details):
- [RELEASE-CONSOLIDATE] Consolidated release.yml, binaries.yml, container.yml into single release.yml; semantic-release outputs gate downstream jobs (b29b8bf)
- [MAINT-GHOST] Next Due autocomputed from LastServiced + IntervalMonths at render time; removed from model/forms/seed (8a8ccc0)
- [DEMO-MEM] `--demo` uses in-memory SQLite; pass db-path to persist demo data (bce212f, 77ad102 squashed)
- [CTRL-C-EXIT] ctrl+c exits 130 via `tea.Interrupt`; `q` stays 0 (bb59c6a)
- [CONTAINER-TAGS] Dropped SHA tags from container image (d8f8f7d)
- [OCI-LABELS] Added description/source/license OCI labels to Nix container image (865174f)

## 2026-02-08 Session 14

**User request**: Fix [WEBSITE] bug -- GitHub links on gh-pages site pointed to `micasa/micasa` (nonexistent) instead of `cpcloud/micasa`.

**Work done**:
- [WEBSITE-BUG] Replaced all 6 `github.com/micasa/micasa` refs in `index.html` on `gh-pages` branch with `github.com/cpcloud/micasa` (hero CTA, install go-install, install release link, footer x3) (3500195)
- [WEBSITE-MAIN] Moved website from `gh-pages` branch to `website/` on `main`; added `pages.yml` workflow (deploy-pages action, triggers on `website/**` changes, workflow_dispatch); copied `index.html`, `style.css`, `CNAME` (343e35a, 3c9bed3)
- Deleted `gh-pages` branch (local + remote)
- [WEBSITE-VIBES] Typewriter heading + aspirational content overhaul (413e24a, b0bb6d9, cc0a955):
  - Replaced carousel with typewriter effect: JS types/erases/cycles words (panicked, daydreamed, avoided, workshopped, entertained); seeded with "asked" so first paint reads "Frequently asked questions"
  - Merged panicked/daydreamed Q&A into one unified section, trimmed to 4 punchy house questions
  - Pitch rewritten: "quietly plotting" joke + "micasa tracks both" product line, split into two visual paragraphs
  - "micasa" brand treatment: monospace pill (JetBrains Mono, linen bg, 4px radius, non-italic)
  - Replaced badge-wall tech list with single prose sentence
  - Pulled "no mouse" into keyboard section subtitle, data note into install footer
  - Removed pitch border-top, simplified feature grid CSS, added inline SVG favicon
  - prefers-reduced-motion support, JS-disabled fallback
  - Added hard rule: website commits use `docs(website):` not `feat(website):`
- README synced with website content: matching features, pitch, install (fixed stale CGO claim + go install URL), tech prose, keybindings
- README house art: pixel-art SVG with dark terminal background, terracotta blocks at 3 opacity levels, 6x10 cell grid, centered via `<div align="center"><img>`
- Added hard rule: keep README and website in sync

## 2026-02-08 Session 15

**User request**: Remove `.cursor/cli.json` from git history entirely; add `.cursor/` to `.gitignore`. User also requested squashing chatty commits but rescinded -- too much churn.

**Work done**:
- [CURSOR-CLEAN] `git filter-branch --index-filter` removed `.cursor/cli.json` from all 182 commits; `--prune-empty` dropped pure-cli.json chore commits; cleaned up backup refs, expired reflogs, GC'd unreachable objects
- Added `.cursor/` to `.gitignore`

## 2026-02-08 Session 16

**User request**: Update `vendorHash` in `flake.nix` if needed after Go module rename; commit the module rename as a bugfix. Then implement [DASHBOARD] landing screen.

**Work done**:
- `nix build` succeeded without vendorHash change (deps unchanged, only module name moved)
- [MODULE-RENAME] Go module `github.com/micasa/micasa` -> `github.com/cpcloud/micasa` in go.mod, all imports, flake.nix OCI label; fixed `.gitignore` binary pattern that blocked `cmd/micasa/` (f2fc33d)
- [DASHBOARD] Dashboard landing screen: at-a-glance overview replaces boot-into-Projects; `D` toggles in Normal mode, `tab` dismisses; refreshes on all data mutations
  - **Data layer** (`data/dashboard.go`): 6 new store methods -- `ListMaintenanceWithSchedule`, `ListActiveProjects`, `ListExpiringWarranties`, `ListRecentServiceLogs`, `YTDServiceSpendCents`, `YTDProjectSpendCents`
  - **Dashboard struct** (`app/dashboard.go`): `dashboardData` with urgency computation, `loadDashboardAt(now)` for testability
  - **Sections**: Overdue/Upcoming Maintenance (danger/warning colors), Active Projects (status-colored, over-budget highlighted), Expiring Soon (warranties + insurance), Recent Activity (last 5 service logs), Spending YTD
  - **Empty states**: emoji + encouraging messages per section ("Nothing overdue -- nice work!", "No active projects. Time to start something?", etc.)
  - **View**: "Dashboard" tab in tab bar, dedicated status bar hints, help overlay updated
  - **Styles**: `DashSection`, `DashLabel`, `DashValue`, `DashOverdue`, `DashUpcoming`, `DashAllClear` -- all Wong palette with AdaptiveColor
  - **Refresh**: `reloadAll()` helper consolidates all data refresh paths (saveForm, undo, redo)
  - **Tests**: 6 data-layer tests (schedule, active projects, warranties, recent logs, spending) + 12 app-layer tests (daysUntil, daysLabel, sort, cap, toggle, dismiss, blocking, empty/populated views, tab bar, status bar) -- 154 total tests passing
  - Added nil guards on `reloadActiveTab`/`reloadAllTabs` for store-less test models

## 2026-02-09 Session 17

**User requests**: Fix dashboard overlay dim bleed-through; dashboard cleanup (remove q hint, strip emoji); move db path from help to tab row; fix terminal right-edge bleed with line clamping + horizontal scroll viewport.

**Work done** (see git log for details):
- [DASH-OVERLAY] Fixed ANSI faint bleed: `cancelFaint()` prepends `\033[22m` per foreground line
- [DASH-CLEANUP] Removed `q`/quit hint from dashboard overlay; stripped all emoji from section headers and empty-state messages (Overdue, Upcoming, Active Projects, Expiring Soon, Recent Activity, Spending)
- [DBPATH-MOVE] Removed db path from help overlay; added to tab row right-aligned with `truncateLeft()` for left-ellipsis when path is long
- [WIDTH-CLAMP] `clampLines()` utility truncates every rendered line to `effectiveWidth()` with ANSI-safe `ansi.Truncate`; applied as safety net at end of `buildBaseView()`
- [HSCROLL] Horizontal scroll viewport for tables:
  - `ViewOffset int` on `Tab` tracks first visible column
  - `ensureCursorVisible()` adjusts offset when cursor moves past edges
  - `viewportRange()` computes which columns fit on screen, accounts for scroll indicator space
  - `tableView()` slices visible projection to viewport window, recalculates widths for viewport columns
  - `viewportSorts()` adjusts sort column indices relative to viewport start
  - Scroll indicators (`◀ more` / `more ▶`) shown below table when columns are off-screen
  - 14 new tests: clamp lines, truncate left, viewport range, cursor visibility, sort adjustment
- [HELP-OVERLAY] Help as stacking overlay via `overlay.Composite()`

## 2026-02-09 Session 18

**User request**: Refactor the width/scroll implementation; one commit per refactor.

**Work done**:
- `refactor(ui): update table viewport in Update, not View` (fb84d4e)
- `refactor(ui): extract table viewport computation` (a46f34a)
- `refactor(ui): fix truncateLeft width accounting` (9ca4f6e)
- `refactor(ui): minor viewport/header cleanups` (1bfa3cb)

## 2026-02-09 Session 19

**User request**: Implement [WEBSITE-CHIMNEY] -- animated chimney smoke with random zig-zag particles on the website hero house.

**Work done**:
- [WEBSITE-CHIMNEY] Replaced 3 static CSS-animated smoke spans with a JS particle system
  - `smoke-bed` container positioned above chimney, JS spawns `smoke-particle` spans on a timer (350-700ms stagger)
  - Each particle has fully randomized properties: lifetime (2.5-4.5s), rise speed, two overlapping sine-wave wobbles (different freq/amp/phase), wind drift, scale growth, peak opacity
  - Multi-harmonic zig-zag: primary low-freq wide wobble + secondary high-freq jitter = organic smoke movement
  - Block character mix (░, ▒, ▓) with density-weighted distribution (more light ░ for wispy look)
  - `requestAnimationFrame` loop with dt capping; particles auto-removed from DOM when faded
  - Max 10 concurrent particles; `prefers-reduced-motion` respected (no particles spawned)
  - CSS: removed `@keyframes smoke-drift` and static `.smoke-*` rules; added `.smoke-bed` + `.smoke-particle` with `will-change: transform, opacity`
  - Fixed chimney detachment: `top: -0.65em` -> `top: -0.15em` so chimney bottom overlaps first roof line
  - Dialed smoke down: max 5 particles (was 10), spawn every 800-1500ms (was 350-700), peak opacity 0.12-0.3 (was 0.25-0.5), slower rise, gentler scale growth, mostly light ░ glyphs
- [PARSE-ARGS] Replaced hand-rolled `parseArgs` loop with `alecthomas/kong` struct-based CLI parser
  - `cli` struct with `arg`, `env`, `help` tags; kong handles help, errors, env var docs
  - Removed `cliOpts`, `parseArgs`, `printHelp` (~35 lines); `resolveDBPath` kept for demo/default logic
  - Help output auto-documents `$MICASA_DB_PATH` env var
- [DASH-OVERLAY] Dashboard rendered as centered overlay using `rmhubbert/bubbletea-overlay`
  - `buildView()` refactored: `buildBaseView()` renders normal table, `buildDashboardOverlay()` renders bordered box
  - `overlay.Composite()` composites dashboard on top of live table view
  - Dashboard hints (j/k, enter, D, ?, q) moved into overlay box; `statusView()` no longer has dashboard branch
  - Removed `dashboardTabsView()`; background shows normal tab bar
  - Dashboard box: rounded border, accent color, max 78 wide, capped to terminal height
  - Updated 3 tests: overlay content, composite rendering, status bar independence
- [HELP-OVERLAY] Help screen as stacking overlay with background dimming
  - `buildView()` chains overlays: base -> dashboard (if shown) -> help (if shown); each layer dims the one below via `dimBackground()` + `cancelFaint()`
  - `cancelFaint()` prepends `\033[22m` (normal intensity) to each foreground line so ANSI faint from dimmed background doesn't bleed into overlay content
  - `buildHelpOverlay()` delegates to existing `helpView()` which already renders bordered box
  - Stacking works: `?` on dashboard shows help over dimmed dashboard over dimmed table
- [WIDTH-CLAMP] Line clamping + horizontal scroll viewport for table
- [DASH-CLEANUP] Removed q hint from dashboard overlay, stripped emoji from section headers/empty states
- [DBPATH-MOVE] Moved db path from help overlay to right-aligned tab row with left-ellipsis truncation

## 2026-02-09 Session 20

**User request**: Build out project documentation. User wants Markdown source in repo, generated HTML in CI (not stored in git), deployed to micasa.dev/docs. Local testing via `nix run '.#docs'`.

**Work done**:
- [DOCS] Full documentation site using mdbook
  - `docs/book.toml` + `docs/src/SUMMARY.md` nav tree
  - 17 Markdown pages across 4 sections:
    - Getting Started: introduction, installation, first-run, concepts
    - User Guide: navigation, house-profile, projects, maintenance, appliances, quotes, dashboard, sorting-and-columns, undo-redo
    - Reference: keybindings (complete table), configuration (CLI flags, env vars, resolution order), data-storage (schema, backup, soft delete)
    - Development: building (Go, Nix, container), architecture (TEA, TabHandler, modal keys, cell rendering, overlays), testing, contributing
  - Content sourced from README, website, help overlay, AGENTS.md session log, and actual source code
  - `flake.nix`: added `pkgs.mdbook` to dev shell, `docs` package + app (`nix run '.#docs'` for live-reload server)
  - `pages.yml`: installs mdbook v0.5.2, builds docs into `website/docs/`, deploys combined site
  - Trigger paths expanded to `[website/**, docs/**]`
  - `.gitignore`: added `docs/book/`
  - Verified build: `mdbook build docs` succeeds cleanly
  - User tried mdBook styling (custom CSS for website palette/fonts), text was too small/inconsistent
  - Evaluated alternatives: embedding website in mdBook, Docusaurus, Hugo, Zola, Pandoc
  - User chose Zola (option 4), then switched to Hugo since project is Go
- [DOCS] Switched from mdBook to Hugo (user request: "let's try hugo since we're already in go land")
  - Deleted mdBook infra: `book.toml`, `theme/custom.css`, `src/SUMMARY.md`, `docs/src/` tree
  - Hugo setup: `docs/hugo.toml`, `docs/layouts/` (baseof, single, list, index, sidebar, pager partials)
  - `docs/static/css/docs.css`: exact website palette/fonts/feel (cream, linen, charcoal, terracotta, sage, DM Serif Display, Source Serif 4, JetBrains Mono), sidebar nav, responsive mobile hamburger
  - Migrated 20 Markdown files to `docs/content/` with TOML frontmatter (title, weight, description, linkTitle); fixed cross-links to Hugo `ref` shortcodes
  - `flake.nix`: replaced `pkgs.mdbook` with `pkgs.hugo`; `website` app does `rm -rf website/docs` + `hugo --source docs`; `docs` app runs `hugo server`
  - `pages.yml`: installs Hugo v0.155.2 (single tar+extract), one-line build command
  - `.gitignore`: `docs/public/`, `docs/resources/` replace `docs/book/`
  - Build verified: 34 pages in 21ms, sidebar + pager + responsive mobile all working
- Fixed sidebar: pre-commit `license-header` hook was prepending HTML comments before Hugo TOML frontmatter; excluded `^docs/content/` from hook, stripped existing headers
- `build-docs` nix app for one-shot Hugo build (reused by `website` app silently)
- Docs linked from website hero CTA + footer, README "Documentation" section
- [VHS-SCREENSHOTS] Switched from asciinema+agg to VHS (Charmbracelet) for screenshots. VHS renders through headless Chrome + ttyd so `lipgloss.AdaptiveColor` 24-bit colors render correctly (agg's 16-color theme mapping was the root cause of "redacted legal docs"). 9 tape files in `docs/tapes/`, each produces a crisp PNG via VHS `Screenshot` command. `capture-screenshots` nix app runs all tapes (or `ONLY=name` for one). Appliances tape uses 1800px width so all 10 columns fit. Dracula theme, JetBrains Mono 14px, 1400x900 default.

## 2026-02-09 Session 21

**User request**: Fix colorless VHS screenshots (user provided good.png vs bad.png for comparison).

**Root cause**: Cursor shell environment sets `NO_COLOR=1` and `TERM=dumb`. VHS inherits these, passes them to the spawned terminal, and lipgloss emits zero color codes.

**Fix**: Added `Env` overrides to all 9 VHS tape files: `NO_COLOR ""`, `TERM "xterm-256color"`, `COLORTERM "truecolor"`, `COLORFGBG "15;0"`.

**Work done**:
- All 8 of 9 screenshots regenerated with full Wong palette colors (4e415fa)
- Replaced asciinema+agg+tmux pipeline in flake.nix with VHS-based `capture-screenshots`
- Added `build-docs` nix app, `pkgs.vhs` to devShell, skip `debug.tape` in capture script
- New hard rules: "Quote nix flake refs" and "Dynamic nix store paths"
- Cleaned up temp files (good.png, bad.png, debug-vhs.png, test_output.gif, debug.tape, test.tape)

**WIP**: `house-profile.tape` -- VHS `Screenshot` silently produces no file when the `H` key has been pressed (toggles house profile expansion in micasa). All other tapes work. Without `H`, the same tape works. Not a timing issue (tested with 3s+ sleeps). Needs further investigation -- possibly VHS Chrome rendering fails on the expanded house profile view.

## 2026-02-09 Session 22

**User request**: Integrate `adrg/xdg` for platform-aware data paths. Also: audit new deps for security before integrating (added as hard rule).

**Work done**:
- Added hard rule: audit new third-party deps for security before integrating
- Audited `adrg/xdg` v0.5.3: no network calls, no file writes (only `MkdirAll` with `0o700`), rejects relative paths, only dep is `golang.org/x/sys` -- clean
- [XDG-LIB] Replaced hand-rolled XDG logic in `internal/data/path.go` with `xdg.DataFile("micasa/micasa.db")`; gets correct macOS (`~/Library/Application Support`) and Windows (`%LOCALAPPDATA%`) paths for free; `MICASA_DB_PATH` override preserved
- All 154 tests pass
- Added `data.AppName` constant to eliminate hardcoded "micasa" strings
- Bumped `go.mod` to `go 1.25`, dropped Go 1.24 from CI, updated release workflow
- Full third-party dependency security audit (govulncheck + manual source review); no issues found
- Added CodeQL and TruffleHog secret scanning to CI
- Added hard rules: nix vendorHash after dep changes, pin Actions to version tags, prefer tools over shell, audit docs on feature changes
- [DOC-SYNC] Updated docs/README/website for features from other machine: `^`/`$` keybindings and horizontal scroll arrows in navigation.md; Go 1.25+ version requirement across all surfaces; cross-platform XDG data paths in configuration, data-storage, first-run docs, README, website; fixed nav doc "NAV" badge to match code ("NORMAL")

## 2026-02-09 Session 23

**User request**: Add `--print-path` CLI flag to show resolved database path and exit.

**Work done**:
- [PRINT-PATH] Added `PrintPath bool` to `cli` struct (kong tag: `--print-path`); resolves path via existing `resolveDBPath`, prints to stdout, exits 0
- 12 new tests in `cmd/micasa/main_test.go`:
  - 6 unit tests for `resolveDBPath`: explicit path, explicit+demo, demo-only (:memory:), default (platform), env override, explicit-beats-env
  - 6 integration tests via `buildTestBinary`: default path, explicit path, env override, demo-no-path, demo-with-path, exit-code-zero
- Docs: `--print-path` section in configuration.md, backup example in data-storage.md uses `$(micasa --print-path)`, README and website CLI snippets updated
- [AUTO-VERSION] Automatic version injection at release time:
  - `VERSION` file (read by `flake.nix` via `builtins.readFile`); replaces hardcoded `version = "0.1.0"`
  - `var version = "dev"` in main.go, overridden by `-X main.version=...` ldflags
  - `--version` flag via `kong.VersionFlag`; nix-built binary shows `0.1.0`, dev builds show `dev`
  - `.releaserc.json`: `@semantic-release/exec` writes next version to VERSION, `@semantic-release/git` commits it; tag lands on the updated commit
  - `release.yml`: `extra_plugins` for exec+git; binaries job injects version via ldflags; binaries + container jobs checkout the tag ref
  - 2 new tests: `TestVersion_DefaultIsDev`, `TestVersion_Injected` (builds with custom ldflags)
  - Docs: `--version` flag added to configuration.md usage block

## 2026-02-09 Session 24

**User request**: Pin nix-installer-action to version tag; pin all GitHub Actions to commit SHAs; add Renovate config for automated updates.

**Work done**:
- [PIN-ACTIONS] Pinned `DeterminateSystems/nix-installer-action` from `@main` to `@v21`
- Pinned all 12 distinct actions across ci.yml, release.yml, pages.yml to commit SHAs with `# vX` version comments
- Resolved annotated tag for `github/codeql-action` (tag object -> commit dereference)
- Pinned `cycjimmy/semantic-release-action` to v4.2.2 (latest in v4 line; no `v4` tag exists, only branch)
- Created `renovate.json` with `config:recommended` + `pinDigests: true` for github-actions manager
- Verified: docker, DeterminateSystems, trufflesecurity are GitHub verified creators; cycjimmy is not

# Completed work

- [REF-SCROLL] Refactor width/scroll implementation (fb84d4e, a46f34a, 9ca4f6e, 1bfa3cb)
- [README] project README with features, keybindings, architecture, install instructions
- Appliances tab, FK links on Maintenance/Quotes, enter follows link, relation indicators in headers (f61993b, 03af1e1)
- [HOUSE-UX] redesign house profile: middot-separated inline text, no chip borders (9deaba7)
- [SERVICELOG] maintenance log feature: service history sub-table per maintenance item (89eefaa)
- [VENDOR-SERVICE] vendor tracking in service log entries (89eefaa)
- [APPLIANCE-MAINT] Maint count column on appliances tab (89eefaa)
- [NOTRUNC] avoid truncating cell text when terminal is wide enough (1379865)
- [ROWHL] soften table row highlight color (5406579)
- [DBPATH] move db path from status bar to help overlay (5406579)
- [SELFG] drop fg override from selected row so status/money colors show through (5406579)
- [CURSOR] replace orange-bg cell cursor with underline+bold; fix ANSI leak (5406579)
- [ULLEN] underline matches text length, not full column width (5406579)
- [CONSOLIDATE] merge LEARNINGS.md and AGENT_LOG.md into AGENTS.md (5406579)
- [ARROWS] use proper arrow/triangle characters for sort indicators (98384e0)
- [SORTSTABLE] sort indicators render within existing column width (98384e0)
- [STATUSBAR-ENTER] remove redundant `enter edit` from Normal mode status bar (98384e0)
- [SORTPK] PK as implicit tiebreaker, skip priority number for single-column sorts (98384e0)
- [EDITLABEL] shorten "edit mode" to "edit" in Normal mode status bar (98384e0)
- [DELETEDANSI] fix ANSI leak on deleted rows (98384e0)
- [SORT] multi-column sorting (98384e0)
- [STRIKELEN] strikethrough length matches text, not full column width (d1720a0)
- [STRIKECLR] softer color for deleted row strikethrough (d1720a0)
- [XEDIT] move x (show deleted) to Edit mode only (d1720a0)
- [DELITALIC] add italic to deleted rows (d1720a0)
- [DTOGGLE] d toggles delete/restore instead of separate d/u keys (d1720a0)
- [ORDINAL] press 1-9 to jump to Nth option in select fields (60ec495)
- [ORDINAL-LABEL] show ordinal numbers next to select options (60ec495)
- [UNDERWAY] rename "in_progress" status to "underway" (ef87b74)
- [SELECTCOLOR] color status labels in select menus to match table cell colors (d05836d)
- [UNDO] undo cell/form edits with u in Edit mode (c6b6739)
- [REDO] redo undone edits with r in Edit mode (c6b6739)
- [COLWIDTH] stable column widths for fixed-option columns (c6b6739)
- [TABHANDLER] TabHandler interface eliminates switch dispatch on TabKind/FormKind (67bfbe3)
- [HIDECOLS] hide/show columns with candy stacks, ladle edges, sparse ellipsis (7bf8835)
- refactor forms.go and view.go: deduplicate submit/edit pairs, centering, inline edit boilerplate, form-data converters (9851c74)
- scrap the log-on-dash-v approach, just enable logging dynamically (75b2c86)
- remove the v1 in Logs; remove the forward slashes; ghost text reads type a Perl-compatible regex (1c623d4)
- build a search engine with local UI, spinner and selection (1c623d4)
- global search interface: pop up box, show matches, select and jump to row (1c623d4)
- highlight the part of the string that the regex matched in log lines (4289fb7)
- entry editing: make editing existing entries work (a457c44)
- anchored status bar: keystroke info always at bottom of terminal (a457c44)
- [BADGE-REFACTOR] replace candy stacks with single-line hidden-column badges (356abd4)
- [BADGE-LEFTALIGN] left-align badges, color-only for cursor position (fef8e04)
- [BADGE-TRIANGLES] triangle glyphs on outermost badges (4336fb2)
- [BADGE-CENTER] center badge row relative to table width (c566cc6, 00bce3d)
- [GHCR] nix container build + push to ghcr.io on release (59a8ba0)
- [SEMREL] semantic-release workflow + container on release event (238b785)
- [GHCR-META] docker/metadata-action semver cascade + SHA tags (f066639, 498b799)
- [PURE-GO-SQLITE] pure-Go SQLite driver, no CGO (b96f7cd)
- [SHOWCOL-HINT] shift+C hint in status bar when columns hidden (90a850d)
- [CI-XPLAT] CI across Linux, macOS, Windows (6295c50)
- [CI-SHUFFLE] -shuffle=on in CI tests (f0878c4)
- [CI-BINARIES] cross-platform binary uploads on release (9303b94)
- [CGO-OFF] CGO_ENABLED=0 everywhere: nix, dev shell, CI (a173788, bae0824, efe06d2)
- [WIN-TEMPDIR] Store.Close() for Windows temp dir cleanup (2c132df)
- [CI-PRECOMMIT] nix pre-commit hooks replace manual lint/tidy CI jobs (9479cce)
- [RELEASE-CONSOLIDATE] consolidated binaries+container into release workflow to fix GITHUB_TOKEN event limitation
- [WEBSITE-BUG] fix GitHub links on website: micasa/micasa -> cpcloud/micasa (3500195)
- [WEBSITE-MAIN] move website from gh-pages branch to website/ on main with Actions deploy workflow (343e35a, 3c9bed3)
- [WEBSITE-VIBES] typewriter heading, aspirational content, pitch tightening, polish (413e24a, b0bb6d9, cc0a955)
- [MODULE-RENAME] Go module path corrected from micasa/micasa to cpcloud/micasa (f2fc33d)
- [DASHBOARD] Dashboard landing screen with overdue/upcoming maintenance, active projects, expiring warranties, recent activity, YTD spending (271121c)
- [WEBSITE-CHIMNEY] animated chimney smoke particle system on website hero house (75318ac)
- [PARSE-ARGS] replaced manual arg parsing with alecthomas/kong
- [XDG-LIB] switched to adrg/xdg for platform-aware data paths (Linux/macOS/Windows)
- [PRINT-PATH] `--print-path` flag: resolves and prints db path to stdout, exits 0
- [PIN-ACTIONS] pin all GitHub Actions to commit SHAs + Renovate config for automated updates

# Remaining work

## Features
- [REF-SCROLL] ~~Refactor width/scroll implementation (pure rendering, extract
  viewport helper, fix left-truncation), one commit per refactor.~~ DONE
- [WEBSITE] Help me build a `github-pages` website for this project. Modern,
  simple, not AI sloppish, whimsical, funny, perhaps even a bit snarky and
  irreverent. Ideally this wouldn't require a bunch of random javascript crap
  like react and 500 MBs of deps, but ugh okay if that's needed to make the
  thing awesome. make sure you setup the `github-pages` branch and all the
  deployment configs so that I can just push to that branch to update the site.
  The site should include a project overview, installation instructions,
  feature list. Bonus points for a "demo" section with animated GIFs showing
  off the terminal UI.
- [DATEPICKER] for date column data entry can we make that a date picker that
  adjusts a nice little terminal calendar based on what the user has typed in
  so far?
- [HOTNESS-KEYS] ~~in nav mode, ^ should go to the first column, $ to the last~~ DONE
- [NORMAL-TO-NAV] change the NORMAL label to NAV
- [STYLING-TIME-TO-MAINT] add a gradient from say green to orange to red (where
  green means maintenance is not due for a long time, orange means it's coming
  up, and red means it's overdue) to the background of the "Next Due" column on
  the maintenance tab. This would be a nice visual indicator of which items
  need attention soon. Maybe make a computed column that contains days until
  next maintenance and base the gradient on that, so we can have consistent
  thresholds for the colors across all items.
- [APPLIANCEAGE] Add an Age column to the Appliances table, it should be
  read-only and computed from purchase date and the current date based on the
  current time zone. I don't think this should involve a data model change, let me know if it does.
- [HIDECOLS-INTERACTIVE] Make the collapsed column stacks interactive: navigate to
  a candy-wrapper pill and press c to unhide just that column (peel it off the
  stack). Currently c hides and C shows all; this would add per-column restore.
- [NESTED-DRILL] Stack-based nested drilldown: push current detail onto a stack
  when drilling deeper (e.g. Appliances > Dishwasher > Filter Replacement service
  log), esc pops back one level, breadcrumb grows with each level.
- [REPLACEMENT] let's add a feature to track replacing appliances based on age
  and ideally we can gather some data to get a rough idea of when a thing is
  due for replacement; would be sweet if we could pull that data based on the
  model number; doesn't need to be super sophisticated, just plausible
- [STATUS-MODE-VERBOSITY] is there a kind of verbose status bar we can add that
  shows keystrokes and also more verbose context?
- [VENDORS-TAB] Vendors as a first-class tab. Vendor model already exists but is
  only accessible through quote and service log forms. Should be browsable,
  editable, and show all work a vendor has done (quotes + service log entries).
- [PAINT-COLORS] Paint color tracking per room: brand, color name/code, finish,
  room/area. "What paint did we use in the living room?" is a universal
  homeowner question.
- [SPENDING-SUMMARY] Aggregate spending view: total spend by year/month, by
  category (projects, maintenance, appliances). Could live as a summary section
  in the house profile or as its own dashboard panel.
- [PROJECT-PRIORITY] Priority or manual ordering for projects. Status captures
  lifecycle but not urgency. A simple priority field (or drag-to-reorder) would
  let homeowners rank what matters most.
- [SEASONAL-CHECKLISTS] Recurring seasonal reminders not tied to a specific
  appliance or interval (e.g. "clean gutters in spring", "check weatherstripping
  before winter"). Could be a lightweight checklist model with season/month tags.
- [WARRANTY-INDICATOR] Visual indicator on appliances for warranty status: green
  if still covered, red/dim if expired. WarrantyExpiry field already exists;
  this is purely a rendering enhancement.
- [NOTES-EXPAND] Read-only note preview: enter or a dedicated key on a notes
  cell expands to show the full text in a popup/overlay, without entering edit
  mode.
- [TABLE-FILTER] In-table row filtering: `/` opens a filter input, typed text
  narrows visible rows across all columns, `esc` clears. Essential for tabs
  with more than ~15 rows. (Search was previously removed; this is a simpler,
  per-tab approach.)
- [QUICK-CAPTURE] Lower-ceremony entry creation. Options: (a) minimal "just
  title" add flow in the TUI that skips optional fields, (b) CLI subcommands
  like `micasa add project "fix squeaky door"`, or (c) both.
- [QUICK-ADD-FORM] Lighter-weight add forms: only require essential fields
  (title + status for projects, name + interval for maintenance), let user fill
  in optional details later via edit.
- [SOFT-DELETE-DOCS] verify that soft deletion works across runs and if it does
  then note in the soft delete section that this functionality works even
  across runs

- [WEBSITE-DESLOP] ~~Remove "close the laptop, reopen the laptop" AI-slop from
  quotes feature blurb on website and README.~~ DONE
- [DASH-OVERLAY-STYLE] Revisit dashboard overlay styling -- noodle on dim/bg
  approach, make it feel polished.
- [HELP-OVERLAY] ~~Convert help screen to an overlay using bubbletea-overlay.
  Must support stacking on top of the dashboard overlay (dimming the layer
  beneath).~~ DONE

## Docs

- [DOCS] ~~Build out project documentation, deploy to micasa.dev/docs via CI~~ switched to Hugo, screenshots added
- [ATTRIBUTION] ensure that claude, codex and cursor are ack'd in the built with section

## Bugs

- [NAV-CLAMP] ~~Remove table column navigation wrapping -- clamp at edges
  instead of wrapping around.~~ DONE

## Questions
- Why are some values pointers to numbers instead of just the number? E.g.,
  HOAFeeCents and PropertyTaxCents. Why aren't those just plain int64s?
