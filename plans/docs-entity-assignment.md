<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Docs Entity Assignment (Issue #424)

## Problem

The top-level Documents tab shows all documents with an Entity column
("project #3", "appliance #1", etc.) but there is no way to:

1. Set the entity when creating a document from the Docs tab
2. Change the entity when editing a document from the Docs tab
3. Inline-edit the Entity column

Documents created from the top-level tab get no entity linkage at all.

## Design

### Single flat entity selector

A `huh.Select` with all active entities from every kind, grouped by type:

```
(none)
Kitchen Fridge (appliance)
Dishwasher (appliance)
Roof Leak (incident)
Kitchen Renovation (project)
HVAC Quote (quote)
Annual Furnace Filter (maintenance)
Bob's Plumbing (vendor)
```

Value type: `entityRef{Kind string, ID uint}` -- comparable, maps directly
to `Document.EntityKind` and `Document.EntityID`.

### Entity name resolution in table

Replace "project #3" labels with resolved names: "Kitchen Renovation (project)".
Build a name map at load time from existing list methods.

### Scope awareness

- **Top-level Docs tab**: entity selector appears in add/edit forms and as
  inline edit on the Entity column.
- **Scoped drilldowns** (e.g., Project > Docs): entity selector is hidden;
  the entity is fixed by context. No change to existing scoped behavior.

## Changes

1. `forms.go` -- Add `entityRef` type, `documentEntityOptions()` builder,
   wire entity selector into `startDocumentForm` (when unscoped) and
   `openEditDocumentForm`.
2. `forms.go` -- Update `documentFormData` to carry `EntityRef entityRef`.
3. `forms.go` -- Update `parseDocumentFormData` / `submitDocumentForm` to
   write `EntityKind`+`EntityID` from `EntityRef`.
4. `forms.go` -- Update `inlineEditDocument` for `documentColEntity` to
   open an inline entity select.
5. `tables.go` -- Add `documentEntityNameMap` builder, update
   `documentRows` to resolve names, update `documentEntityLabel`.
6. `handlers.go` -- Pass entity name map through `documentHandler.Load`.
