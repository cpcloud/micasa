<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Confirmation State Machine

Replace the three boolean confirmation flags with a single enum so that
illegal states (multiple confirmations active simultaneously) are
unrepresentable.

Resolves: #688

## Current State

Three independent booleans control confirmation dialogs:

| Flag | Location | Purpose |
|------|----------|---------|
| `confirmHardDelete` | `Model` | Permanent incident deletion (y/n) |
| `confirmDiscard` | `formState` | Discard dirty form changes (y/n) |
| `confirmQuit` | `formState` | Modifier on discard: also quit the app |

`confirmQuit` is always set together with `confirmDiscard` (never alone).
Although `confirmHardDelete` (normal mode) and `confirmDiscard` (form mode)
should never overlap in practice, the type system does not enforce this.

## Design

### New enum

```go
type confirmKind int

const (
    confirmNone            confirmKind = iota
    confirmHardDelete      // permanent incident deletion
    confirmFormDiscard     // discard dirty form, stay in app
    confirmFormQuitDiscard // discard dirty form and quit
)
```

### Structural changes

- Add `confirm confirmKind` field on `Model` (replaces `confirmHardDelete bool`).
- Keep `hardDeleteID uint` on `Model` (only meaningful when `confirm == confirmHardDelete`).
- Remove `confirmDiscard bool` and `confirmQuit bool` from `formState`.
- `resetFormState()` clears `m.confirm` to `confirmNone` (only the form-related kinds).

### Dispatch changes

- `Update()` top-level: `m.confirmHardDelete` -> `m.confirm == confirmHardDelete`
- `updateForm()`: `m.fs.confirmDiscard` -> `m.confirm.isFormConfirm()`
- `handleConfirmDiscard()`: `m.fs.confirmQuit` check -> `m.confirm == confirmFormQuitDiscard`
- Ctrl+Q in dirty form: sets `m.confirm = confirmFormQuitDiscard`
- ESC on dirty form: sets `m.confirm = confirmFormDiscard`
- `promptHardDeleteIncident()`: sets `m.confirm = confirmHardDelete`

### Rendering

- `statusView()`: switch on `m.confirm` instead of checking separate bools.

### Tests

Update all references in:
- `form_save_test.go` (`m.fs.confirmDiscard`, `m.fs.confirmQuit`)
- `handler_crud_test.go` (`m.confirmHardDelete`)
- `lighter_forms_test.go` (`m.fs.confirmDiscard`)
