+++
title = "Projects"
weight = 3
description = "Track home improvement projects from idea to completion."
linkTitle = "Projects"
+++

The Projects tab tracks things you want to do to your house, from small
repairs to major renovations.

![Projects table](/docs/images/projects.png)

## Adding a project

1. Switch to the Projects tab (`tab` until it's active)
2. Enter Edit mode (`i`)
3. Press `a` to open the add form
4. Fill in the fields and save (`ctrl+s`)

The **Title** field is required. Everything else is optional or has a default.

## Fields

| Column  | Description | Notes |
|---------|-------------|-------|
| ID      | Auto-assigned primary key | Read-only |
| Type    | Project category | Select from pre-seeded types (Renovation, Repair, etc.) |
| Title   | Project name | Required. Free text. |
| Status  | Lifecycle stage | Select: ideating, planned, quoted, underway, delayed, completed, abandoned |
| Budget  | Planned cost | Dollar amount (e.g., 1250.00) |
| Actual  | Real cost | Dollar amount. Over-budget is highlighted on the dashboard. |
| Start   | Start date | YYYY-MM-DD |
| End     | End date | YYYY-MM-DD |

## Status lifecycle

Projects move through these statuses. Each has a distinct color in the table:

- **ideating** -- just an idea, not committed
- **planned** -- decided to do it, working out details
- **quoted** -- have vendor quotes, comparing options
- **underway** -- work in progress
- **delayed** -- stalled for some reason
- **completed** -- done
- **abandoned** -- decided not to do it

## Inline editing

In Edit mode, press `e` on any non-ID column to edit just that cell inline.
Press `e` on the ID column (or any read-only column) to open the full edit
form.

## Linked quotes

Quotes reference projects. On the Quotes tab, the **Project** column shows
which project a quote belongs to. The column header shows `m:1` indicating the
many-to-one relationship.
