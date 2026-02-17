<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Complexity Audit

Issue: #306

## Executive Summary

The codebase is roughly 15k lines of Go across ~30 files. The architecture is
sound (Bubble Tea MVC, TabHandler interface, GORM data layer), but several areas
have accumulated structural duplication that inflates line count, makes bugs
harder to spot, and increases the cost of adding new entity types.

The biggest wins are in three areas:
1. **Detail-opening boilerplate** (model.go) -- 7 nearly identical methods
2. **Overlay composition** (view.go) -- 6 copy-paste overlay blocks
3. **Scoped handler duplication** (handlers.go) -- handlers that delegate 8/10
   methods to a parent

Smaller wins include collapsing repetitive inline-edit column switches,
deduplicating scoped column specs, and simplifying the status hint fitting
algorithm.

## Findings

### P0 -- High impact, low risk

#### 1. Consolidate detail-opening methods into a table-driven dispatch

**Files:** `internal/app/model.go:726-900`

Seven methods (`openServiceLogDetail`, `openApplianceMaintenanceDetail`,
`openVendorQuoteDetail`, `openVendorJobsDetail`, `openProjectQuoteDetail`,
`openProjectDocumentDetail`, `openApplianceDocumentDetail`) follow the exact
same pattern:

```go
func (m *Model) openXxxDetail(parentID uint, parentName string) error {
    specs := xxxColumnSpecs()
    return m.openDetailWith(detailContext{
        ParentTabIndex: m.active,
        ParentRowID:    parentID,
        Breadcrumb:     "Tab" + sep + parentName + sep + "Sub",
        Tab: Tab{
            Kind:    tabKind,
            Name:    "Sub",
            Handler: xxxHandler{parentID: parentID},
            Specs:   specs,
            Table:   newTable(specsToColumns(specs), m.styles),
        },
    })
}
```

And `openDetailForRow` is a 50-line switch dispatching to them, each arm loading
the parent entity just to get its name.

**Fix:** Define a `detailDef` struct and a table mapping `(TabKind, colTitle)` to
a detail definition. Each entry provides: handler constructor, column spec
function, breadcrumb template, and a store-getter for the parent name. Replace
the 7 methods + switch with a single `openDetail` that looks up the definition
and constructs the `detailContext`.

**Estimate:** -120 lines, eliminates the most repetitive block in the codebase.

#### 2. Consolidate overlay composition in buildView

**Files:** `internal/app/view.go:15-57`

Six overlay blocks follow the identical pattern:
```go
if <condition> {
    fg := cancelFaint(m.buildXxxOverlay())
    base = overlay.Composite(fg, dimBackground(base), overlay.Center, overlay.Center, 0, 0)
}
```

**Fix:** Extract a helper `applyOverlay(base *string, condition bool, render func() string)`
or a slice of `(condition, renderer)` pairs iterated in priority order.

**Estimate:** -20 lines, more importantly makes overlay priority explicit and
adding new overlays trivial.

#### 3. Reduce scoped handler boilerplate via embedding

**Files:** `internal/app/handlers.go:362-931`

Seven scoped handlers (e.g. `vendorQuoteHandler`, `projectQuoteHandler`,
`applianceMaintenanceHandler`, `projectDocumentHandler`,
`applianceDocumentHandler`) delegate 7-9 of their 10 interface methods to the
parent handler. The only methods that differ are `Load` (scoped query),
`InlineEdit` (column remapping), and sometimes `StartAddForm`/`SubmitForm`.

**Fix:** Create a `scopedHandler` struct that embeds a parent `TabHandler` and
adds `parentID uint`, `columnRemap func(int) int`, and `loadFn`. The parent
handler's methods are inherited; only the differing methods are overridden.
This eliminates ~400 lines of near-identical delegation code.

**Estimate:** -350 lines.

### P1 -- Medium impact

#### 4. Dedup scoped column specs in tables.go

**Files:** `internal/app/tables.go`

Scoped column spec functions (`vendorQuoteColumnSpecs`,
`projectQuoteColumnSpecs`, `applianceMaintenanceColumnSpecs`,
`entityDocumentColumnSpecs`) are near-copies of the parent spec with one column
removed.

**Fix:** Write a `withoutColumn(specs []columnSpec, title string) []columnSpec`
helper that filters out a column by title. Each scoped spec becomes a one-liner:
`withoutColumn(quoteColumnSpecs(), "Vendor")`.

**Estimate:** -80 lines.

#### 5. Simplify inline edit dispatch in forms.go

**Files:** `internal/app/forms.go:654-900`

Each `inlineEditXxx` method is a column-index switch that maps column numbers to
field edits. The switches are verbose but each arm is a one-liner calling
`openInlineInput` or `openDatePicker`. The column-index magic numbers are
fragile -- adding a column silently shifts all the indices.

**Fix:** Define per-entity inline edit tables mapping column index to an
`inlineEditAction` (field pointer, validator, form kind). This is less urgent
since it's more of a maintainability improvement than a complexity reduction,
but it eliminates the column-index fragility.

**Estimate:** -60 lines, +robustness.

#### 6. Simplify renderStatusHints 3-pass algorithm

**Files:** `internal/app/view.go:429-514`

The status hint fitting algorithm does 3 passes: compact non-required → compact
required → drop by priority. This is clever but the 3 separate loops with
`compact[]` and `dropped[]` arrays are hard to follow.

**Fix:** Merge into a single priority-ordered pass: iterate hints by descending
priority, for each try compact first, then drop. One loop, same behavior.

**Estimate:** -30 lines, +readability.

### P2 -- Lower impact, nice to have

#### 7. Extract `reflect`-based form helpers

**Files:** `internal/app/forms.go` (cloneFormData), `internal/app/form_select.go`
(selectOptionCount)

Two places use `reflect`: `cloneFormData` for snapshot cloning and
`selectOptionCount` for inspecting huh's private `filteredOptions` field.

The `cloneFormData` usage is fine -- it's a clean shallow copy of a known struct.
The `selectOptionCount` accessing `filteredOptions` by name is fragile and will
break if huh renames that field.

**Fix:** For `selectOptionCount`, consider contributing upstream to huh to expose
the filtered option count, or accept the fragility with a test guard. Low
priority.

#### 8. Consolidate delete/restore guard patterns in store.go

**Files:** `internal/data/store.go`

Delete methods follow: count dependents → count doc dependents → soft delete.
Restore methods follow: load unscoped → check parent alive → restore.
`validateDocumentParent` is a switch on 6 entity kinds.

These are already reasonably well-factored with `requireParentAlive` and
`countDependents` helpers. The main issue is that adding a new entity type
requires touching multiple places. This is acceptable given the domain -- new
entity types are rare.

**Fix:** No change needed. The patterns are clear and the helpers provide
sufficient DRY.

#### 9. joinNonEmpty / joinWithSeparator duplication

**Files:** `internal/app/view.go:1057-1063`

`joinNonEmpty` is just a wrapper for `joinWithSeparator` with swapped argument
order. Having both is confusing.

**Fix:** Remove `joinNonEmpty`, update callers to use `joinWithSeparator`.

**Estimate:** -5 lines.

## Implementation Order

1. P0-1: Detail-opening consolidation (biggest win, cleanest change)
2. P0-2: Overlay composition consolidation
3. P0-3: Scoped handler embedding
4. P1-4: Scoped column spec dedup
5. P1-5: Inline edit tables (if time permits)
6. P1-6: Status hint algorithm simplification
7. P2-9: joinNonEmpty removal

Items P2-7 and P2-8 are deferred -- the reflect usage and store patterns are
acceptable as-is.

## Non-findings

These areas were reviewed and found to be appropriately complex for what they do:

- **chat.go** -- The chat state machine is complex but necessarily so. It
  handles streaming, cancellation, model pulling, and SQL execution. The
  complexity matches the problem.
- **table.go column width algorithm** -- Multi-phase width allocation is complex
  but well-documented with clear phase names. Simplifying it would sacrifice
  correct behavior.
- **filter.go** -- Clean, well-factored. Pin/filter logic is straightforward.
- **dashboard.go** -- Reasonable complexity for what it does.
- **calendar.go** -- Simple date picker state machine.
- **column_finder.go** -- Clean fuzzy matching with good scoring.
- **data/store.go** -- Well-factored with helpers; adding entities requires
  touching multiple places but that's inherent to the domain.
- **config package** -- Simple TOML parsing, no issues.
- **llm package** -- Clean client/prompt separation.
