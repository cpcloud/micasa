+++
title = "Projects"
weight = 2
description = "Track home improvement projects from idea to completion."
linkTitle = "Projects"
+++

The Projects tab tracks things you want to do to your house, from small
repairs to major renovations.

![Projects table](/docs/images/projects.webp)

## Adding a project

1. Switch to the Projects tab (`tab` until it's active)
2. Enter Edit mode (`i`)
3. Press `a` to open the add form
4. Fill in the fields and save (`ctrl+s`)

The `Title` field is required. Everything else is optional or has a default.

## Fields

| Column | Type | Description | Notes |
|-------:|------|-------------|-------|
| `ID` | auto | Auto-assigned primary key | Read-only |
| `Type` | select | Project category | Pre-seeded types (Renovation, Repair, etc.) |
| `Title` | text | Project name | Required |
| `Status` | select | Lifecycle stage | See [status lifecycle](#status-lifecycle) below |
| `Budget` | money | Planned cost | Dollar amount (e.g., 1250.00) |
| `Actual` | money | Real cost | Over-budget is highlighted on the dashboard |
| `Start` | date | Start date | YYYY-MM-DD |
| `End` | date | End date | YYYY-MM-DD |

## Status lifecycle

Projects move through these statuses. Each has a distinct color in the table:

- <span class="status-ideating">**ideating**</span> -- just an idea, not committed
- <span class="status-planned">**planned**</span> -- decided to do it, working out details
- <span class="status-quoted">**quoted**</span> -- have vendor quotes, comparing options
- <span class="status-underway">**underway**</span> -- work in progress
- <span class="status-delayed">**delayed**</span> -- stalled for some reason
- <span class="status-completed">**completed**</span> -- done
- <span class="status-abandoned">**abandoned**</span> -- decided not to do it

## Status filters

In Normal mode on the Projects tab:

- Press `z` to toggle hiding projects with status `completed`
- Press `a` to toggle hiding projects with status `abandoned`
- Press `t` to toggle hiding **settled projects** (`completed` + `abandoned`)

## Description

The edit form includes a `Description` textarea (in the "Timeline" group) for
longer notes about the project. The description is stored on the project record
but doesn't appear as a table column.

## Inline editing

In Edit mode, press `e` on any non-`ID` column to edit just that cell inline.
Press `e` on the `ID` column (or any read-only column) to open the full edit
form, which includes the description field.

## Linked quotes

Quotes reference projects. On the Quotes tab, the `Project` column shows
which project a quote belongs to. The column header shows `→` indicating it's
a navigable link.
