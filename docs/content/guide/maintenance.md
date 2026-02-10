+++
title = "Maintenance"
weight = 4
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

| Column      | Description | Notes |
|-------------|-------------|-------|
| `ID`        | Auto-assigned | Read-only |
| `Item`      | Task name | Required. E.g., "HVAC filter replacement" |
| `Category`  | Task type | Select from pre-seeded categories (HVAC, Plumbing, etc.) |
| `Appliance` | Linked appliance | Optional. Select from your appliances. Links to Appliances tab. |
| `Last`      | Last serviced date | YYYY-MM-DD |
| `Next`      | Next due date | Auto-computed: `Last` + interval. Read-only. |
| `Every`     | Interval | Shown as "N mo" (e.g., "6 mo") |
| `Log`       | Service log count | Drilldown -- press `enter` to open. |

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

| Column         | Description |
|----------------|-------------|
| `ID`           | Auto-assigned |
| `Date`         | When the work was done (required) |
| `Performed By` | "Self" or a vendor name |
| `Cost`         | Dollar amount |
| `Notes`        | Free text |

The detail view supports all the same operations as a regular tab: add, edit,
delete, sort, undo. Press `esc` to close the detail view and return to the
Maintenance table.

### Vendors in service logs

The "Performed By" field is a select. The first option is always "Self
(homeowner)." All existing vendors appear as additional options. To add a new
vendor, create one via the Quotes form first -- vendors are shared across
quotes and service logs.

## Appliance link

When a maintenance item is linked to an appliance, the `Appliance` column shows
the appliance name. This column is a foreign key link -- in Normal mode, press
`enter` on it to jump to that appliance in the Appliances tab.
