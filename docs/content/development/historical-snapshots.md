+++
title = "Historical Snapshots"
weight = 3
description = "Pre-freeze plan for preserving historical truth on mutable links."
linkTitle = "Historical Snapshots"
+++

Before schema freeze, micasa needs explicit snapshot rules so historical
records stay correct even when linked entities are later edited.

This document defines the snapshot strategy for quotes, service logs, and
future finance records.

## Problem

Some records represent historical events:

- A quote received from a vendor on a specific date
- A completed service visit with a cost and notes
- (Upcoming) an invoice or payment tied to a project

Those records currently link to mutable parent rows (for example, `vendors`).
If the linked row is edited later (name, email, phone), the historical record
can silently change its displayed meaning.

## Snapshot policy

When a record is historical, we persist both:

- **Foreign keys** for relational joins and integrity
- **Snapshot fields** for immutable historical display

Snapshot fields are write-once at create time and are never auto-refreshed from
the parent row.

## Entities and snapshot fields

### Quotes

For each `quotes` row, capture vendor identity at quote time:

- `vendor_snapshot_name`
- `vendor_snapshot_contact_name`
- `vendor_snapshot_email`
- `vendor_snapshot_phone`
- `vendor_snapshot_website`

Display rule:

- Show snapshot fields in quote history views
- Keep FK join to vendor for navigation and aggregate reporting

### Service log entries

For each `service_log_entries` row with a vendor, capture:

- `vendor_snapshot_name`
- `vendor_snapshot_contact_name`
- `vendor_snapshot_email`
- `vendor_snapshot_phone`

If the service was self-performed (`vendor_id` is null), snapshot fields stay
empty.

### Future finance entities

For planned invoice/payment models, snapshot:

- Vendor identity fields
- Source quote number/name (if invoice derived from quote)
- Monetary fields that represent the committed transaction

## Mutation rules

- **Create**: snapshot fields are copied from current linked rows
- **Update historical record**: snapshot fields may be edited manually on that
  row, but are not recomputed from parent FKs
- **Update parent row** (for example vendor rename): no automatic propagation
  to snapshot fields on historical children

This keeps old records historically accurate while still allowing parent cleanup.

## Migration and backfill

Additive migration only:

1. Add nullable snapshot columns
2. Backfill from existing FK joins
3. Move reads to prefer snapshot values
4. Keep null-safe fallback to FK values for any legacy gaps

Backfill query shape (conceptual):

- `UPDATE quotes SET vendor_snapshot_name = vendors.name ... WHERE quotes.vendor_id = vendors.id`
- Equivalent update for `service_log_entries`

No rows are deleted or rewritten in ways that lose original FK links.

## Cross-home compatibility

Snapshots are row-local and remain valid under multi-home support. They do not
depend on global uniqueness constraints or mutable lookup tables.

## Test requirements for implementation PR

When this strategy is implemented, add tests that verify:

- Editing a vendor does not alter quote/service-log snapshot display values
- Historical rows render snapshot values when present
- Backfill populates snapshots for pre-existing rows
- New records always write snapshots when created

## Status

This document is the agreed pre-freeze design. Implementation should land in a
follow-up code PR before model hardening.
