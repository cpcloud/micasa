+++
title = "Documents"
weight = 8
description = "Attach files to projects, appliances, and other records."
linkTitle = "Documents"
+++

Attach files to your home records -- warranties, manuals, invoices, photos.

![Documents table](/images/documents.webp)

## Adding a document

1. Switch to the Docs tab (`f` to cycle forward)
2. Enter Edit mode (`i`), press `a`
3. Fill in a title and optional file path, then save (`ctrl+s`)

If you provide a file path, micasa reads the file into the database as a BLOB
(up to 50 MB). The title auto-fills from the filename when left blank.

You can also add documents from within a project or appliance detail view --
drill into the `Docs` column and press `a`. Documents added this way are
automatically linked to that record.

## Fields

| Column | Type | Description | Notes |
|-------:|------|-------------|-------|
| `ID` | auto | Auto-assigned | Read-only |
| `Title` | text | Document name | Required. Auto-filled from filename if blank |
| `Entity` | text | Linked record | E.g., "project #3". Only shown on top-level Docs tab |
| `Type` | text | MIME type | E.g., "application/pdf", "image/jpeg" |
| `Size` | text | File size | Human-readable (e.g., "2.5 MB"). Read-only |
| `Notes` | notes | Free-text annotations | Press `enter` to preview |
| `Updated` | date | Last modified | Read-only |

## File handling

- **Storage**: files are stored as BLOBs inside the SQLite database, so
  `cp micasa.db backup.db` backs up everything -- no sidecar files
- **Size limit**: 50 MB per file
- **MIME detection**: automatic from file contents and extension
- **Checksum**: SHA-256 hash stored for integrity
- **Cache**: when you open a document (`enter` on the row), micasa extracts it
  to the XDG cache directory and opens it with your OS viewer

## Entity linking

Documents can be linked to any record type: projects, appliances, quotes,
maintenance items, vendors, or service log entries. The link is set
automatically when adding from a drill view, or can be left empty for
standalone documents.

The `Entity` column on the top-level Docs tab shows which record a document
belongs to (e.g., "project #3", "appliance #7").

## Drill columns

The `Docs` column appears on the **Projects** and **Appliances** tabs, showing
how many documents are linked to each record. In Nav mode, press `enter` to
drill into a scoped document list for that record.

## Inline editing

In Edit mode, press `e` on the `Title` or `Notes` column to edit inline. Press
`e` on any other column to open the full edit form. The file attachment cannot
be changed after creation.
