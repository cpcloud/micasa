+++
title = "Appliances"
weight = 5
description = "Track physical equipment, warranties, and linked maintenance."
linkTitle = "Appliances"
+++

The Appliances tab tracks physical equipment in your home -- everything from
the water heater to the dishwasher.

![Appliances table](/docs/images/appliances.png)

## Adding an appliance

1. Switch to the Appliances tab
2. Enter Edit mode (`i`), press `a`
3. Fill in the identity and details forms

Only the `Name` is required.

## Fields

| Column      | Description | Notes |
|-------------|-------------|-------|
| `ID`        | Auto-assigned | Read-only |
| `Name`      | Appliance name | Required. E.g., "Kitchen Refrigerator" |
| `Brand`     | Manufacturer | E.g., "LG" |
| `Model`     | Model number | For warranty lookups and replacements |
| `Serial`    | Serial number | |
| `Location`  | Where in the house | E.g., "Kitchen", "Basement" |
| `Purchased` | Purchase date | YYYY-MM-DD |
| `Warranty`  | Warranty expiry | YYYY-MM-DD. Shows on dashboard when expiring. |
| `Cost`      | Purchase price | Dollar amount |
| `Maint`     | Maintenance count | Drilldown -- press `enter` to view linked maintenance. |

## Warranty tracking

Enter the warranty expiry date when you add an appliance. The
[Dashboard]({{< ref "/guide/dashboard" >}}) shows appliances with warranties expiring within
90 days (or recently expired within 30 days) in the "Expiring Soon" section.

## Maintenance drilldown

The `Maint` column shows how many maintenance items are linked to this
appliance. In Normal mode, navigate to the `Maint` column and press `enter` to
open a detail view showing those maintenance items (without the Appliance
column, since it's redundant).

From the detail view you can add, edit, or delete maintenance items. Press
`esc` to return to the Appliances table.

## Inline editing

All columns except `ID` and `Maint` support inline editing. Press `e` in Edit
mode on a cell to edit just that field.
