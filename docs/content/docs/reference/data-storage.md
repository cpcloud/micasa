+++
title = "Data Storage"
weight = 3
description = "SQLite database file, schema, backup, and portability."
linkTitle = "Data Storage"
+++

micasa stores everything in a single SQLite database file. This page covers
how data is stored and how to manage it.

## Database file

By default, the database lives in your platform's data directory:

| Platform | Default path |
|----------|-------------|
| Linux    | `~/.local/share/micasa/micasa.db` |
| macOS    | `~/Library/Application Support/micasa/micasa.db` |
| Windows  | `%LOCALAPPDATA%\micasa\micasa.db` |

See [Configuration]({{< ref "/docs/reference/configuration" >}}) for how to customize the path.

The database path is shown in the tab row (right-aligned) so you always know
which file you're working with.

## Schema

micasa uses [GORM](https://gorm.io) for database access with automatic schema
migration. The database is created and migrated on startup -- you never need to
run migrations manually.

### Tables

| Table                  | Description |
|------------------------|-------------|
| `house_profiles`       | Single row with your home's details |
| `projects`             | Home improvement projects |
| `project_types`        | Pre-seeded project categories |
| `quotes`               | Vendor quotes linked to projects |
| `vendors`              | Shared vendor records |
| `maintenance_items`    | Recurring maintenance tasks |
| `maintenance_categories` | Pre-seeded maintenance categories |
| `appliances`           | Physical equipment |
| `service_log_entries`  | Service history per maintenance item |
| `documents`            | File metadata + attachments linked to records |
| `deletion_records`     | Audit trail for soft deletes/restores |

### Pre-seeded data

On first run, micasa seeds default **project types** (Renovation, Repair,
Landscaping, etc.) and **maintenance categories** (HVAC, Plumbing, Electrical,
etc.). These are reference data used in select dropdowns.

## Backup

Your database is a single file. Back it up with `cp`:

```sh
cp "$(micasa --print-path)" ~/backups/micasa-$(date +%F).db
```

Since SQLite supports [hot
backup](https://www.sqlite.org/backup.html), you can safely copy the file
while micasa is running.

## Soft delete

micasa uses GORM's soft delete feature. When you delete an item, it sets a
`deleted_at` timestamp rather than removing the row. This means:

- Deleted items can be restored (press `d` on a deleted item in Edit mode)
- The `deletion_records` table tracks when items were deleted and restored
- Toggle `x` in Edit mode to show/hide deleted items
- Soft deletions persist across runs -- quit and reopen, and your deleted items
  are still hidden (but restorable). Nothing is ever permanently lost unless
  you edit the database file directly

### Referential integrity guards

Soft deletion respects foreign key relationships:

- **Delete guards**: you cannot delete a parent record that has active
  children. For example, deleting a project that still has quotes is refused
  with an actionable error message ("delete its quotes first").
- **Restore guards**: you cannot restore a child record whose parent is
  deleted. For example, restoring a quote whose project is deleted is refused
  ("restore the project first").
- These guards apply at every FK level: projects/quotes,
  maintenance/service logs, and appliances/maintenance (including nullable
  links where a value was set).

## Portability

The database is a standard SQLite file. You can:

- Open it with any SQLite client (`sqlite3`, DB Browser for SQLite, etc.)
- Move it between machines by copying the file
- Query it directly with SQL if needed

The file uses a pure-Go SQLite driver (no CGO), so the binary has zero
native dependencies.

## LLM data exposure

If you enable the optional [LLM chat]({{< ref "/docs/guide/llm-chat" >}}),
micasa sends database contents to the configured LLM endpoint. By default this
is localhost (Ollama), so data stays on your machine. If you configure a remote
endpoint, your home data travels over the network. See the
[Data and privacy]({{< ref "/docs/guide/llm-chat#data-and-privacy" >}})
section of the LLM chat guide for details on exactly what is sent.

## Demo mode

`micasa --demo` creates an in-memory database populated with fictitious sample
data. To persist demo data, pass a file path:

```sh
micasa --demo /tmp/demo.db
```

Demo data includes sample addresses, phone numbers (all `555-xxxx`), and
`example.com` email addresses.
