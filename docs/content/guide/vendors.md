+++
title = "Vendors"
weight = 6
description = "Browse and manage your vendors."
linkTitle = "Vendors"
+++

The Vendors tab gives you a single view of everyone you've hired or gotten
quotes from.

![Vendors table](/docs/images/vendors.webp)

## Columns

| Column | Type | Description | Notes |
|-------:|------|-------------|-------|
| `ID` | auto | Auto-assigned | Read-only |
| `Name` | text | Company or person name | Required, unique |
| `Contact` | text | Contact person | Optional |
| `Email` | text | Email address | Optional |
| `Phone` | text | Phone number | Optional |
| `Website` | text | URL | Optional |
| `Quotes` | count | Number of linked quotes | Read-only |
| `Jobs` | count | Number of linked service log entries | Read-only |

## How vendors are created

Vendors can be created in two ways:

1. **Directly** on the Vendors tab: enter Edit mode (`i`), press `a`
2. **Implicitly** when adding a quote or service log entry -- type a vendor
   name and micasa finds or creates the record

## Editing a vendor

Navigate to the Vendors tab, enter Edit mode (`i`), and press `e` on the
cell you want to change. Edits to a vendor's contact info propagate to all
quotes and service log entries that reference that vendor.

## Cross-tab navigation

The `Vendor` column on the Quotes tab is a live link (shown with `→` in the header). Press `enter` on
a vendor name in the Quotes table to jump to that vendor's row in the Vendors
tab.

## Counts

The `Quotes` and `Jobs` columns show how many quotes and service log entries
reference each vendor. These are read-only aggregate counts.

## Notes

The edit form includes a `Notes` textarea for free-text annotations about the
vendor. Notes are stored on the vendor record but don't appear as a table
column.

## No deletion

Vendors cannot be deleted because they are referenced by quotes and service
log entries. If you need to retire a vendor, add a note to their record.
