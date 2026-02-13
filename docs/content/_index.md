+++
title = "micasa docs"
description = "Documentation for micasa, a terminal UI for tracking everything about your home."
+++

Your house is quietly plotting to break while you sleep -- and you're dreaming
about redoing the kitchen. micasa tracks both from your terminal.

micasa is a keyboard-driven terminal UI for managing everything about your home:
maintenance schedules, projects, vendor quotes, appliances, warranties, and
service history. It stores all data in a single SQLite file on your machine.
No cloud. No account. No subscriptions.

## What it does

- **[Maintenance tracking]({{< ref "/guide/maintenance" >}})** with auto-computed due dates, service log history,
  and vendor records
- **[Project management]({{< ref "/guide/projects" >}})** from ideating through completion (or graceful
  abandonment), with budget tracking
- **[Quote comparison]({{< ref "/guide/quotes" >}})** across vendors, with cost breakdowns
- **[Appliance inventory]({{< ref "/guide/appliances" >}})** with warranty windows, purchase dates, and
  maintenance history tied to each one
- **[Dashboard]({{< ref "/guide/dashboard" >}})** showing overdue maintenance, active projects, expiring
  warranties, and YTD spending at a glance
- **[Vim-style modal navigation]({{< ref "/using/navigation" >}})** with Normal and Edit modes, multi-column
  sorting, column hiding, and cross-tab FK links
- **[LLM chat]({{< ref "/guide/llm-chat" >}})** -- ask questions about your home data in plain English,
  powered by a local LLM ([Ollama](https://ollama.com))

## What it doesn't do

micasa is not a smart home controller, a home automation platform, or a
property management SaaS. It's a personal tool for one house (yours), designed
to answer questions like "when did I last change the furnace filter?" and "is
the dishwasher still under warranty?"

## Quick start

```sh
go install github.com/cpcloud/micasa/cmd/micasa@latest
micasa --demo   # poke around with sample data
micasa          # start fresh with your own house
```

![micasa dashboard](/docs/images/dashboard.webp)

Read the full [Installation]({{< ref "/getting-started/installation" >}}) guide for
other options (binaries, Nix, container).
