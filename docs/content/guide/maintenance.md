+++
title = "Maintenance"
weight = 5
description = "Recurring upkeep tasks with schedules and service logs."
linkTitle = "Maintenance"
+++

The Maintenance tab tracks recurring upkeep tasks -- things that need to happen
on a schedule to keep your house running.

![Maintenance table](/docs/images/maintenance.png)

## Adding a maintenance item

1. Switch to the Maintenance tab
2. Enter Edit mode (`i`), press `a`
3. Fill in the schedule form

The `Item` name is required. Set a `Category`, optionally link an
`Appliance`, and set the `Last` serviced date and `Every` (interval months) to
enable auto-computed due dates.

## Fields

| Column | Type | Description | Notes |
|-------:|------|-------------|-------|
| `ID` | auto | Auto-assigned | Read-only |
| `Item` | text | Task name | Required. E.g., "HVAC filter replacement" |
| `Category` | select | Task type | Pre-seeded categories (HVAC, Plumbing, etc.) |
| `Appliance` | select | Linked appliance | Optional. Links to Appliances tab |
| `Last` | date | Last serviced date | YYYY-MM-DD |
| `Next` | date | Next due date | Auto-computed: `Last` + `Every`. Read-only |
| `Every` | number | Interval | Shown as "N mo" (e.g., "6 mo") |
| `Log` | drilldown | Service log count | Press `enter` to open |

## Next due date

The `Next` column is computed automatically from `Last` serviced +
`Every` (interval months). You don't edit it directly. If either `Last` or
`Every` is empty, `Next` is blank.

Items that are overdue or coming due soon appear on the
[Dashboard]({{< ref "/guide/dashboard" >}}) with urgency indicators.

## Service log

Each maintenance item has a service log -- a history of when the work was
actually performed. The `Log` column shows the entry count.

To view the service log, navigate to the `Log` column in Normal mode and press
`enter`. This opens a detail view with its own table:

![Service log drilldown](/docs/images/service-log.png)

| Column | Type | Description |
|-------:|------|-------------|
| `ID` | auto | Auto-assigned |
| `Date` | date | When the work was done (required) |
| `Performed By` | select | "Self" or a vendor name |
| `Cost` | money | Dollar amount |
| `Notes` | text | Free text |

The detail view supports all the same operations as a regular tab: add, edit,
delete, sort, undo. Press `esc` to close the detail view and return to the
Maintenance table.

### Vendors in service logs

The "Performed By" field is a select. The first option is always "Self
(homeowner)." All existing vendors appear as additional options. To add a new
vendor, create one via the Quotes form or Vendors tab first -- vendors are
shared across quotes and service logs.

The `Performed By` column is also a foreign key link (shown with `→` in the
header). In Normal mode, press `enter` on a vendor name to jump to that
vendor's row in the Vendors tab.

## Additional form fields

The edit form includes fields that don't appear as table columns:

| Field | Type | Description |
|------:|------|-------------|
| `Manual URL` | text | Link to the product or service manual |
| `Manual notes` | text | Free-text manual excerpts or reminders |
| `Cost` | money | Estimated or typical cost per service |
| `Notes` | text | General notes about this maintenance item |

These fields are accessible when editing a maintenance item (press `e` on the
`ID` column or any read-only column to open the full form).

## Appliance link

When a maintenance item is linked to an appliance, the `Appliance` column shows
the appliance name. This column is a foreign key link -- in Normal mode, press
`enter` on it to jump to that appliance in the Appliances tab.
