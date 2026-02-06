- make a new tab, call it Appliances, for tracking appliance information; add
  an appliance field to Maintenance that links (by foreign key) to this new
  dataset; second, make it so that when i hit enter on cell that is linked
  (maybe also find a way to indicate that a column is linked to another table
  [including its relationship 1:1, 1:m, m:1 or m:n as the indicator would also
  │ be fuckin' dope]) that will move you to that cell in the other tab (if it's
  1:1 of course. if 1:m go to the first match, if m:1, there's only a single
  match, if m:n not sure, help me   figure out the ux for that )
- [RW-HOUSE-UX] redesign house profile collapsed/expanded views: remove chip borders,
  use middot-separated inline text, cleaner section layout in expanded view
- for maintenance items, compute the default ghost text for next due date from
  the last serviced date + the maintenance interval and default to that
- [RW-NOTRUNC] avoid truncating cell text when terminal is wide enough
- [RW-APPLIANCEAGE] Add an Age column to the Appliances table, it should be
  read-only and computed from purchase date and the current date based on the
  current time zone.
- [RW-HIDECOLS] Add the ability to hide and show columns.

## Completed

- [RW-SERVICELOG] maintenance log feature: service history sub-table per maintenance item (PENDING_SHA)
- [RW-VENDOR-SERVICE] vendor tracking in service log entries (PENDING_SHA)
- [RW-APPLIANCE-MAINT] Maint count column on appliances tab (PENDING_SHA)

- [RW-ROWHL] soften table row highlight color -- textMid too close to white fg; use surface bg instead (5406579)
- [RW-DBPATH] move db path from status bar to help overlay (5406579)
- [RW-SELFG] drop fg override from selected row so status/money colors show through (5406579)
- [RW-CURSOR] replace orange-bg cell cursor with underline+bold; fix ANSI leak from nested lipgloss Render (5406579)
- [RW-ULLEN] underline matches text length, not full column width (5406579)
- [RW-CONSOLIDATE] merge LEARNINGS.md and AGENT_LOG.md into AGENTS.md sections (5406579)
- [RW-ARROWS] use proper arrow/triangle characters for sort indicators (98384e0)
- [RW-SORTSTABLE] sort indicators render within existing column width, no layout shift (98384e0)
- [RW-STATUSBAR-ENTER] remove redundant `enter edit` from Normal mode status bar (98384e0)
- [RW-SORTPK] PK as implicit tiebreaker, skip priority number for single-column sorts (98384e0)
- [RW-EDITLABEL] shorten "edit mode" to "edit" in Normal mode status bar (98384e0)
- [RW-DELETEDANSI] fix ANSI leak on deleted rows: merge strikethrough into per-cell style instead of wrapping (98384e0)
- [RW-SORT] multi-column sorting (98384e0)
- [RW-STRIKELEN] strikethrough length matches text, not full column width (d1720a0)
- [RW-STRIKECLR] softer color for deleted row strikethrough (d1720a0)
- [RW-XEDIT] move x (show deleted) to Edit mode only (d1720a0)
- [RW-DELITALIC] add italic to deleted rows (d1720a0)
- [RW-DTOGGLE] d toggles delete/restore instead of separate d/u keys (d1720a0)
- [RW-ORDINAL] press 1-9 to jump to Nth option in select fields (60ec495)
- [RW-ORDINAL-LABEL] show ordinal numbers next to select options (60ec495)
- [RW-UNDERWAY] rename "in_progress" status to "underway" (ef87b74)
- [RW-SELECTCOLOR] color status labels in select menus to match table cell colors (d05836d)
- [RW-UNDO] undo cell/form edits with u in Edit mode (c6b6739)
- [RW-REDO] redo undone edits with r in Edit mode (c6b6739)
- [RW-COLWIDTH] stable column widths for fixed-option columns (c6b6739)
- [RW-TABHANDLER] TabHandler interface eliminates switch dispatch on TabKind/FormKind (67bfbe3)

- refactor forms.go and view.go: deduplicate submit/edit pairs, centering, inline edit boilerplate, form-data converters (9851c74)
- scrap the log-on-dash-v approach, just enable logging dynamically (and allow changing log level) with a keyboard shortcut and bring up the logger ui component when that key is pressed (it's a toggle obviously) (75b2c86)

- remove the v1 in Logs; remove the forward slashes; the ghost text should read type a Perl-compatible regex; put the log lines themselves in visually separate components (1c623d4)
- build a search engine (and sweet embedded UI for it) that MUST run locally, with spinner and selection (1c623d4)
- make a global search interface that works like google, pop up a box, show matches, select and jump to row, runs locally with background indexing and spinner (1c623d4)
- highlight the part of the string that the regex matched in log lines (4289fb7)
- i can't edit existing entries, make that work please (a457c44)
- with no logging the keystroke info is really tight up against the bottom of the data (a457c44)
- can you make the keystroke info always appear all the way at the bottom of the terminal? (a457c44)
