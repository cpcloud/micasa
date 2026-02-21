<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Service Log Document Attachments (Issue #427)

## Goal

Allow users to attach documents (receipts, invoices, etc.) to individual
service log entries. The data layer already supports this via the polymorphic
`Document` model with `EntityKind = "service_log"`. This plan wires up the
terminal UI.

## Current State

- `DocumentEntityServiceLog` constant exists in `models.go`
- `Store.ListDocumentsByEntity`, `CountDocumentsByEntity`, etc. all work
  with service logs
- `validateDocumentParent` handles `DocumentEntityServiceLog` for restore
  guards
- Store-level tests verify delete/restore behavior with service log documents

## Changes

### 1. Add "Docs" drilldown column to service log table

**`tables.go`**:
- Add `serviceLogColDocs` to the `serviceLogCol` iota enum
- Add `{Title: "Docs", Min: 5, Max: 8, Align: alignRight, Kind: cellDrilldown}`
  to `serviceLogColumnSpecs()`
- Update `serviceLogRows()` to accept `docCounts map[uint]int` and produce
  a drilldown cell for the Docs column

### 2. Count documents in the service log handler

**`handlers.go`**:
- In `serviceLogHandler.Load`, after listing entries, collect IDs and call
  `store.CountDocumentsByEntity(data.DocumentEntityServiceLog, ids)` to get
  doc counts, then pass them to `serviceLogRows`

### 3. Create the service log document detail definition

**`model.go`**:
- Add `serviceLogDocumentDef` following the same pattern as
  `projectDocumentDef` / `applianceDocumentDef`
- `tabKind: tabMaintenance`, `subName: tabDocuments.String()`
- Uses `newEntityDocumentHandler(data.DocumentEntityServiceLog, id)`
- Breadcrumb appends "Docs" to the existing service log breadcrumb context
  (service logs are already nested, so this is a 3rd-level drilldown)
- `getName` resolves service log entry → formatted date string

### 4. Wire routing

**`model.go`**:
- Add `{[]TabKind{tabMaintenance}, tabDocuments.String(), serviceLogDocumentDef}`
  to `detailRoutes`

### 5. Tests

- `TestServiceLogDocumentHandlerFormKind` (mirrors existing pattern)
- `TestOpenDetailForRow_ServiceLogDocuments` (integration test for drilldown
  routing)
- Verify the new column spec has "Docs" title and `cellDrilldown` kind
