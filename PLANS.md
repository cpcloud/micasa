<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Local LLM Interface (#126)

100% local-inference chat for wrangling home data. Talks to any
OpenAI-compatible endpoint (Ollama, llama.cpp, LM Studio, etc.)
-- no data leaves the machine.

## Config

TOML at `$XDG_CONFIG_HOME/micasa/config.toml`:

```toml
[llm]
base_url = "http://localhost:11434/v1"
model = "qwen3"
```

Env var overrides: `OLLAMA_HOST` (auto-appends `/v1`), `MICASA_LLM_MODEL`.

## Architecture: Two-Stage Pipeline

```
User question (English)
  │
  ▼
Stage 1: NL → SQL
  • System prompt = schema DDL + few-shot NL→SQL examples
  • Non-streaming completion (need full SQL before executing)
  • LLM returns a single SELECT statement
  │
  ▼
Validation
  • Must start with SELECT
  • No INSERT/UPDATE/DELETE/DROP/ALTER/CREATE
  • Row cap via existing ReadOnlyQuery
  │
  ▼
Execute against SQLite
  • Returns columns + rows
  │
  ▼
Stage 2: Results → English
  • System prompt = question + SQL + tabular results
  • Streaming completion (nice incremental UX)
  • LLM summarizes structured data into natural language
  │
  ▼
Display in chat viewport (markdown-rendered)
```

### Module responsibilities

```
llm/prompt.go    - BuildSQLPrompt (schema + few-shot → SQL generation)
                 - BuildSummaryPrompt (question + results → English)
llm/client.go    - ChatComplete (non-streaming, for SQL generation)
                 - ChatStream (streaming, for summary -- already exists)
data/query.go    - ReadOnlyQuery (already exists, validates + executes)
app/chat.go      - Two-stage pipeline orchestration in submitChat
```

### Fallback

If SQL generation fails (bad SQL, validation error, execution error),
show the error and fall back to the old single-stage approach so the
user still gets an answer.

## UX Flow

1. Press `@` to open the chat overlay (centered, bordered).
2. Text input at the bottom. Type a question, press Enter.
3. Notice appears: "generating query..."
4. Stage 1 runs (non-streaming). SQL result shown as a notice.
5. SQL executes. Stage 2 streams the natural-language summary.
6. Conversation persists across messages in the session.
7. Esc closes the overlay (conversation preserved).
8. Ctrl+C cancels an in-flight stream.
9. If no LLM configured: shows config hint with example TOML.

## Non-goals (v1)

- Cloud/hosted LLM support (future: Anthropic Messages API)
- Persistent chat history across sessions
