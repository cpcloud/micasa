<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Schema evolution

## Current behavior

Schema management uses GORM's `AutoMigrate()`, called on every startup in
`cmd/micasa/main.go`. AutoMigrate is **additive-only**: it creates missing
tables and adds missing columns, but it never drops, renames, or changes the
type of anything that already exists. This is fine for development but
becomes a problem once real users have data.

### What AutoMigrate does

- Creates tables that don't exist yet
- Adds columns that are missing from existing tables
- Creates missing indexes
- Adds missing foreign key constraints (SQLite: only at table creation)

### What AutoMigrate does NOT do

- Drop columns
- Rename columns or tables
- Change column types (e.g. `text` to `integer`)
- Change nullability (e.g. add `NOT NULL` to an existing column)
- Drop indexes
- Modify foreign key constraints on existing tables
- Remove or reorder enum-like string values
- Rewrite data (backfills, splits, merges)

### Current schema (14 tables)

| Table                  | Soft-delete | FK references                      |
|------------------------|:-----------:|------------------------------------|
| `house_profiles`       |             | none                               |
| `project_types`        |             | none                               |
| `maintenance_categories` |           | none                               |
| `vendors`              | yes         | none                               |
| `appliances`           | yes         | none                               |
| `settings`             |             | none (key-value, `key` is PK)      |
| `chat_inputs`          |             | none                               |
| `deletion_records`     |             | none (polymorphic via entity+target_id) |
| `documents`            | yes         | polymorphic (entity_kind + entity_id) |
| `projects`             | yes         | project_types (RESTRICT), vendors (SET NULL) |
| `maintenance_items`    | yes         | maintenance_categories (RESTRICT), appliances (SET NULL) |
| `quotes`               | yes         | projects (RESTRICT), vendors (RESTRICT) |
| `service_log_entries`  | yes         | maintenance_items (CASCADE), vendors (SET NULL) |

### Seed data

`SeedDefaults()` populates `project_types` (12 rows) and
`maintenance_categories` (9 rows) using `FirstOrCreate`, making it
idempotent. This runs after AutoMigrate on every startup.

## Known unsupported schema changes

These are changes that would be needed eventually but cannot be done with
AutoMigrate. Listed roughly by likelihood of being needed.

### 1. Adding NOT NULL to existing columns

Several columns that should logically be required (e.g. `projects.title`,
`vendors.name`, `maintenance_items.name`) are nullable in SQLite because
GORM creates them as `text` without a NOT NULL constraint. AutoMigrate
cannot retroactively add NOT NULL.

**SQLite constraint:** Requires the 12-step table rebuild pattern (create
new table with NOT NULL, copy data, drop old, rename).

### 2. Renaming columns

If a column name turns out to be confusing or inconsistent (e.g.
`cost_cents` on maintenance_items vs `total_cents` on quotes), renaming
requires `ALTER TABLE RENAME COLUMN` (SQLite 3.25+, bundled in our driver).
AutoMigrate won't do this.

### 3. Changing column types

For example, storing monetary values as `integer` cents is correct, but if
we ever needed to change `bathrooms` from `real` to `integer` (storing
quarter-baths as an int), AutoMigrate cannot alter the column type. Requires
table rebuild.

### 4. Dropping columns

If a column becomes obsolete (e.g. `manual_url` and `manual_text` on
maintenance_items might be replaced by a document attachment), AutoMigrate
cannot drop it. `ALTER TABLE DROP COLUMN` is available in SQLite 3.35+
(bundled), but AutoMigrate doesn't use it.

### 5. Renaming enum-like string values

Project status values (`ideating`, `planned`, `quoted`, `underway`,
`delayed`, `completed`, `abandoned`) are stored as plain strings. If
`underway` were renamed to `in_progress`, a `UPDATE projects SET status =
'in_progress' WHERE status = 'underway'` migration would be needed.
AutoMigrate has no concept of data migrations.

### 6. Splitting or merging tables

If `house_profiles` grew too large and needed to be split into
`house_profiles` + `house_financials`, or if `settings` and `chat_inputs`
were merged, AutoMigrate cannot express this.

### 7. Changing foreign key behavior

SQLite foreign key constraints are baked into the CREATE TABLE statement.
Changing `ON DELETE RESTRICT` to `ON DELETE CASCADE` (or vice versa)
requires a full table rebuild. AutoMigrate does not modify constraints on
existing tables.

### 8. Adding composite unique constraints

If we needed a unique constraint on `(entity_kind, entity_id, file_name)`
in documents (preventing duplicate filenames per entity), this would be a
new index. AutoMigrate can add indexes but cannot add unique constraints
that might conflict with existing data.

## Backup story

The single-file principle (`micasa backup backup.db` is a complete backup)
means any future migration framework inherits a simple rollback story:
restore from the copy. No down-migrations needed if users back up before
upgrading.

## Parked work

A full migration framework is implemented on the parked branch
`feat/schema-migration-framework`. It replaces AutoMigrate with
a hand-rolled `schema_version` table + numbered Go migration functions. That
branch is ready to merge when the first non-additive schema change is needed
(likely a 2.0 milestone, since semrel already shipped 1.x).
