<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Plans

Tracks in-flight feature plans and ad-hoc requests so context survives agent crashes/handoffs.

## Appliances Tab (remaining_work.md item 1) -- DONE

The first work item is a multi-part feature. Prior agent did most of the data + UI work but left the build broken. This session wired the remaining pieces.

**What was already done** (by prior agent, not logged):
- Data model: `Appliance` struct, store CRUD (Create/Get/Update/Delete/Restore/List)
- Table: `applianceColumnSpecs`, `applianceRows`, `NewTabs` includes Appliances
- Forms: `applianceFormData`, `startApplianceForm`, `startEditApplianceForm`, `openApplianceForm`, `submitApplianceForm`, `submitEditApplianceForm`
- Types: `formAppliance`, `tabAppliances`, `columnLink`, `cell.LinkID`
- Demo seed data: 7 appliances, 3 maintenance-appliance links
- Maintenance form: ApplianceID field, appliance select dropdown

**What this session added** (to fix build + complete wiring):
- `applianceOptions()` helper for huh select dropdowns
- `inlineEditAppliance()` for per-cell editing
- Switch cases in: `handleFormSubmit`, `startAddForm`, `startEditForm`, `deleteSelected`, `restoreByTab`, `deletionEntityForTab`, `reloadTab`, `tabLabel`, `tabIndex`, `buildSearchEntries`

**Cross-tab navigation (enter on linked cell)** -- DONE:
- `navigateToLink()` switches tab and selects target row by ID
- `selectedCell()` helper reads cell at current cursor position
- Header shows relation type (e.g. "m:1") in muted rose via `LinkIndicator` style
- Status bar shows "follow m:1" hint when cursor is on a linked cell with a target
- Works for Quotes.Project (m:1 -> Projects) and Maintenance.Appliance (m:1 -> Appliances)
- For empty links (e.g. maintenance with no appliance), falls through to normal edit

## House Profile UX Redesign (RW-HOUSE-UX)

**Problem**: Collapsed and expanded house profile views feel like a "wall of text tags." Every key-value pair is wrapped in a `RoundedBorder` chip box, creating dense visual noise.

**Collapsed (before)**: Title row + row of 6 bordered chip boxes (House, Loc, Yr, Sq Ft, Beds, Baths)
**Expanded (before)**: Title row + 2 chip rows + 3 section rows each packed with bordered chips

**Design**:

Collapsed -- single clean middot-separated line, no borders:
```
House Profile ▸  h toggle
Elm Street · Springfield, IL · 4bd / 2.5ba · 2,400 sqft · 1987
```
Nickname pops in orange (HeaderValue), stats in subdued gray (HeaderHint).

Expanded -- section headers with inline middot-separated values, no chip borders:
```
House Profile ▾  h toggle
Elm Street · 742 Elm Street, Springfield, IL 62704

 Structure  1987 · 2,400 sqft · 8,500 lot · 4bd / 2.5ba
            fnd Poured Concrete · wir Copper · roof Asphalt Shingle
            ext Vinyl Siding · bsmt Finished
 Utilities  heat Forced Air Gas · cool Central AC · water Municipal
            sewer Municipal · park Attached 2-Car
 Financial  ins Acme Insurance · policy HO-00-0000000 · renew 2026-08-15
            tax $4,850.00 · hoa Elm Street HOA ($150.00/mo)
```
Section headers use existing HeaderSection style. Values use dim label + bright value (`hlv` helper). Continuation lines indent to align with values.

**Implementation**:
1. Add helpers: `styledPart`, `bedBathLabel`, `sqftLabel`, `lotLabel`, `hlv`, `houseSection`
2. Rewrite `houseCollapsed` and `houseExpanded`
3. Remove now-unused `chip`, `sectionLine`, `renderHouseValue`, `HeaderChip` style

## Modal System

**Goal**: Vim-style modal keybindings that work *with* bubbles/table's built-in vim nav.

**Problem**: bubbles/table defaults bind `d` (half-page-down) and `u` (half-page-up), which
conflict with our delete and undo keys. Single-mode apps must intercept these before the table
sees them, losing useful navigation. A modal system resolves this cleanly.

**Modes**:

### Normal mode (default, `-- NORMAL --`)

All table vim keys work natively: `j`/`k` rows, `d`/`ctrl+d` half-page-down,
`u`/`ctrl+u` half-page-up, `g`/`G` top/bottom, `space`/`b` page-down/up. Plus:
- `h`/`l` or `left`/`right` = column movement (free keys, table doesn't bind them)
- `tab`/`shift+tab` = switch tabs
- `H` = toggle house profile
- `x` = toggle show deleted (view-only toggle)
- `enter` = edit current cell (convenience; opens form directly)
- `i` = enter Edit mode
- `?` = help
- `q` = quit
- `esc` = clear status

### Edit mode (`-- EDIT --`)

Same navigation, but `d`/`u` rebound from table nav to data actions:
- `a` = add new entry
- `e`/`enter` = edit cell/row
- `d` = delete
- `u` = undo/restore
- `p` = edit house profile
- `esc` = back to Normal mode

Table KeyMap is dynamically updated: entering Edit mode strips `d`/`u` from
HalfPageDown/HalfPageUp (keeps `ctrl+d`/`ctrl+u`). Returning to Normal restores them.

### Form mode (unchanged)

`ctrl+s` save, `esc` cancel. Returns to whichever mode (Normal/Edit) was active before.

**Also in this change**:
- Remove logging feature (files, state, UI, keybindings)
- Remove search feature (files, state, UI, keybindings)
- Mode indicator badge in status bar (accent for Normal, secondary for Edit)
- Per-mode help items in status bar

## Multi-Column Sort (Option 1: Simple Stack)

**UX**:
- `s` on current column in Normal mode cycles: none -> asc -> desc -> none
- Multiple columns sortable; priority = insertion order
- Column headers show direction + priority: `Name ^1`, `Cost v2`
- `S` (shift+s) clears all sorts, back to default PK asc
- Normal mode only (ignored in Edit mode)

**Data model**:
- New types in `types.go`: `sortDir` (asc/desc), `sortEntry` (col index + dir)
- New field on `Tab`: `Sorts []sortEntry`

**Sort logic** (new file `sort.go`):
- `toggleSort(tab, colIdx)`: find col in tab.Sorts; cycle none->asc->desc->none; append if new, update if exists, remove if cycling to none
- `clearSorts(tab)`: empty the slice
- `applySorts(tab)`: sort tab.CellRows, tab.Rows (meta), and tab.Table rows in sync using `sort.SliceStable` with a multi-key comparator
- Comparator is cell-kind-aware: numeric for cellMoney (parse cents), date-aware for cellDate (parse date string), lexicographic for everything else
- Called from `reloadTab` after rows are loaded, and from `toggleSort`/`clearSorts`

**Header rendering**:
- `renderHeaderRow` gets the tab's `Sorts` slice
- After title text, append sort indicator: `^1` or `v2` (using arrow chars)
- Indicator styled with accent color so it pops

**Key handling**:
- `s` in `handleNormalKeys`: call `toggleSort` on current column, re-render
- `S` in `handleNormalKeys`: call `clearSorts`, re-render

**Help overlay**:
- Add `s` / `S` to Normal mode section

**Tests**:
- `toggleSort` cycling: none->asc->desc->none
- Multi-column: add col 0 asc, add col 2 desc, verify both present with correct priority
- `clearSorts`: empties slice
- Comparator: numeric, date, string ordering
- `S` key ignored in Edit mode

## TabHandler Interface Refactoring -- DONE

**Problem**: Entity-specific logic was scattered across 9+ `switch tab.Kind` /
`switch m.formKind` dispatch sites in model.go, forms.go, and undo.go.
Adding a new entity type required touching every dispatch site ("shotgun surgery").

**Solution**: `TabHandler` interface with 10 methods encapsulating all entity-specific
operations. Each entity type implements the interface as a stateless struct.
The `Tab` struct embeds the handler. Dispatch sites replaced with single
`tab.Handler.Method(...)` calls.

**Interface**: `FormKind`, `Load`, `Delete`, `Restore`, `StartAddForm`,
`StartEditForm`, `InlineEdit`, `SubmitForm`, `Snapshot`, `SyncFixedValues`

**Special case**: House form (formHouse) has no tab. It stays as explicit
special-case code in `handleFormSubmit` and `snapshotEntity`.

**Removed dead code**: `restoreByTab`, `startInlineCellEdit` (dispatch wrapper),
unused `table` import from model.go.

## Remaining Work Items (from remaining_work.md)

1. **Appliance tab + cross-tab FK navigation** -- tab done, navigation TBD
2. **Column sorting** -- DONE
3. **Maintenance ghost text** -- compute next_due from last_serviced + interval as default

## Service Log + Vendor Tracking (RW-SERVICELOG, RW-VENDOR-SERVICE, RW-APPLIANCE-MAINT)

### Problem

Maintenance items track *what* needs doing and *when*, but there's no history of
*when it was actually done*, *who did it*, and *what it cost each time*. Homeowners
need a running log per maintenance task. Many tasks are vendor-performed (too
dangerous or specialized for DIY), so the log should capture that naturally.

### Data Model

**New model: `ServiceLogEntry`**

```
ServiceLogEntry
  ID             uint (PK)
  MaintenanceItemID  uint (FK -> MaintenanceItem, NOT NULL)
  ServicedAt     time.Time (required - when the work was performed)
  VendorID       *uint (FK -> Vendor, nullable; nil = homeowner did it)
  CostCents      *int64
  Notes          string
  CreatedAt, UpdatedAt, DeletedAt (soft-delete)
```

The existing `Vendor` model is already in the DB and used by Quotes. We reuse
it directly -- no new vendor table needed.

**Changes to MaintenanceItem**: None structurally. The existing `LastServicedAt`
and `CostCents` remain as user-entered schedule hints. A future enhancement could
auto-update `LastServicedAt` from the most recent log entry, but that's separate
scope.

### Architecture: Detail View

Instead of adding a new top-level tab (which the user explicitly doesn't want),
we introduce a **detail view** concept: a secondary table that temporarily
replaces the main tab table when the user "drills in."

**New state on Model**:
- `detailContext *detailContext` -- when non-nil, we're in a detail view
- `detailContext` struct holds:
  - `ParentTab TabKind` -- which tab we came from
  - `ParentRowID uint` -- which row we drilled into
  - `Breadcrumb string` -- e.g. "Maintenance > HVAC filter replacement"
  - `Tab Tab` -- a full Tab struct (with Handler, Table, Specs, etc.)

**Key behavior when detailContext is active**:
- View renders the detail tab's table instead of the main tab's
- Breadcrumb bar replaces the tab bar (shows path back to parent)
- `esc` in Normal mode: close detail, return to parent tab + row
- All other keys: delegated to the detail tab's handler (add, edit, delete,
  sort, undo, etc.)

This means `serviceLogHandler` implements `TabHandler` just like the other
entity handlers. The detail view is just a tab that isn't in the top-level
tab bar.

### UX Flow

1. User is on the Maintenance tab, Normal mode
2. Navigates to a row (e.g. "HVAC filter replacement")
3. Presses `enter` -> detail view opens showing service log entries for that item
4. Breadcrumb: `Maintenance > HVAC filter replacement`
5. The service log table has columns: Date, Performed By, Cost, Notes
6. User can:
   - `i` to enter Edit mode, `a` to add a new entry, `e`/`enter` to edit
   - `s` to sort, `d` to delete, `u` to undo
   - `esc` (Normal) returns to the maintenance list
7. "Performed By" shows "Self" or the vendor name

**Visual distinction**:
- The breadcrumb bar replaces the tab underline, uses a distinct accent
- The detail table header uses a slightly different style (or the same -- TBD
  based on how it looks)
- A subtle visual indicator on the parent table's row (like a count badge in a
  "Log" column) shows how many entries exist

### Maintenance Tab Changes

Add a **"Log" column** (rightmost or near-right) to the maintenance table:
- Shows the count of service log entries: "3", "1", or empty
- This column has `Kind: cellReadonly` (can't inline-edit a count)
- Pressing `enter` on any column in Normal mode on Maintenance opens the
  detail view (since Normal mode enter = "drill down / navigate")
- In Edit mode, `enter`/`e` still inline-edits cells as before

### Service Log Form

**Add form** (when pressing `a` in the service log detail view):
- ServicedAt (date, required, default: today)
- Performed By (select: "Self", then list of existing vendors)
- If new vendor needed: vendor name input (find-or-create, same as quotes)
- Cost (money, optional)
- Notes (text, optional)

**Edit form**: same fields, pre-populated.

### Vendor Handling

Reuse the `findOrCreateVendor` pattern from quotes. The form has a select
dropdown: first option "Self (homeowner)", then all known vendors by name.
If the user picks a vendor, `VendorID` is set; if "Self", it's nil.

For a future enhancement, we could add a "New vendor..." option that expands
into vendor detail fields, but for now selecting from existing vendors (or
using the quote form to create new ones) is sufficient. Actually, we should
support inline vendor creation in the service log form too -- let's include
a vendor name field that does find-or-create when it doesn't match an existing
vendor.

Simpler approach: The select has "Self" + existing vendors. To add a new vendor,
the user types the name in a text field that appears when "New vendor" is
selected. This keeps the form clean.

### Appliance Integration

**Option chosen**: Add a "Maint" column to the Appliances table showing the count
of linked maintenance items. This is a 1:m relationship indicator.

Additionally: in Normal mode, pressing `enter` on an appliance row opens a
detail view showing that appliance's linked maintenance items. This uses the
same detail view architecture. From *that* detail view, pressing `enter` on a
maintenance item opens its service log (nested detail).

Wait -- nested detail adds complexity. For v1, let's keep it to one level:
- Maintenance tab `enter` -> service log detail
- Appliance tab: "Maint" column shows count, enter on the Maint column follows
  an `m:1` link to the Maintenance tab (uses existing cross-tab nav)
- Or better: enter on appliance row shows linked maintenance items in detail view

For v1: just add the "Maint" count column on Appliances and use the existing
cross-tab FK navigation from Maintenance -> Appliance. The Appliance tab already
has a link from Maintenance. We add a reverse indicator.

**Simplest v1 for appliances**: Show maintenance count in Appliances, clicking
through navigates to Maintenance tab filtered (or just unfiltered with a status
message). Full detail-view for appliances = future scope.

### Implementation Phases

**Phase 1: Data layer**
- Add `ServiceLogEntry` model to `models.go`
- Add `DeletionEntityServiceLog` constant
- Add store CRUD: `ListServiceLog(maintenanceItemID)`, `GetServiceLog(id)`,
  `CreateServiceLog`, `UpdateServiceLog`, `DeleteServiceLog`, `RestoreServiceLog`
- Add `CountServiceLogs(maintenanceItemIDs []uint) map[uint]int` for batch count
- Migrate, seed demo data (a few entries for existing maintenance items)

**Phase 2: Detail view architecture**
- Add `detailContext` struct and field to Model
- Modify `Update()` to delegate to detail tab when active
- Modify `View()` to render detail table + breadcrumb when active
- `esc` in Normal mode when detail is active = close detail
- `enter` in Normal mode on Maintenance tab = open detail

**Phase 3: Service log handler + UI**
- `serviceLogColumnSpecs()`, `serviceLogRows()`
- `serviceLogHandler` implementing `TabHandler`
- Service log form (add/edit) with vendor selector
- Service log inline edit
- Demo seed data

**Phase 4: Visual polish + appliance integration**
- Breadcrumb styling
- "Log" column on Maintenance with entry count
- "Maint" column on Appliances with item count
- Visual distinction for detail view
- Help overlay updates
