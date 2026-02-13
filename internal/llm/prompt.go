// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package llm

import (
	"fmt"
	"strings"
	"time"
)

// TableInfo describes a database table for context injection.
type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

// ColumnInfo describes a single column in a table.
type ColumnInfo struct {
	Name    string
	Type    string
	NotNull bool
	PK      bool
}

// BuildSQLPrompt creates a system prompt that instructs the LLM to translate
// a natural-language question into a single SELECT statement. The prompt
// includes the current date, the full schema as DDL, and few-shot examples.
// If extraContext is non-empty, it's appended at the end.
func BuildSQLPrompt(tables []TableInfo, now time.Time, extraContext string) string {
	var b strings.Builder
	b.WriteString(sqlSystemPreamble)
	b.WriteString(dateContext(now))
	b.WriteString("\n\n## Schema\n\n```sql\n")
	for _, t := range tables {
		b.WriteString(formatDDL(t))
		b.WriteString("\n")
	}
	b.WriteString("```\n")
	b.WriteString(sqlSchemaNotes)
	b.WriteString("\n\n")
	b.WriteString(sqlFewShot)
	if extraContext != "" {
		b.WriteString("\n\n## Additional context\n\n")
		b.WriteString(extraContext)
	}
	return b.String()
}

// BuildSummaryPrompt creates a system prompt for the second stage: turning
// SQL results into a concise natural-language answer.
// If extraContext is non-empty, it's appended at the end.
func BuildSummaryPrompt(question, sql, resultsTable string, now time.Time, extraContext string) string {
	var b strings.Builder
	b.WriteString(summarySystemPreamble)
	b.WriteString(dateContext(now))
	b.WriteString("\n\n## User question\n\n")
	b.WriteString(question)
	b.WriteString("\n\n## SQL executed\n\n```sql\n")
	b.WriteString(sql)
	b.WriteString("\n```\n\n## Results\n\n```\n")
	b.WriteString(resultsTable)
	b.WriteString("\n```\n\n")
	b.WriteString(summaryGuidelines)
	if extraContext != "" {
		b.WriteString("\n\n## Additional context\n\n")
		b.WriteString(extraContext)
	}
	return b.String()
}

// BuildSystemPrompt assembles the old single-stage system prompt, used as
// a fallback when the two-stage pipeline fails.
// If extraContext is non-empty, it's appended at the end.
func BuildSystemPrompt(tables []TableInfo, dataSummary string, now time.Time, extraContext string) string {
	var b strings.Builder
	b.WriteString(fallbackPreamble)
	b.WriteString(dateContext(now))
	b.WriteString("\n\n## Database Schema\n\n")
	for _, t := range tables {
		b.WriteString(formatTable(t))
		b.WriteString("\n")
	}
	b.WriteString(fallbackSchemaNotes)
	if dataSummary != "" {
		b.WriteString("\n\n## Current Data\n\n")
		b.WriteString(dataSummary)
	}
	b.WriteString("\n\n")
	b.WriteString(fallbackGuidelines)
	if extraContext != "" {
		b.WriteString("\n\n## Additional context\n\n")
		b.WriteString(extraContext)
	}
	return b.String()
}

// FormatResultsTable renders query results as a pipe-delimited text table,
// compact enough for an LLM context window.
func FormatResultsTable(columns []string, rows [][]string) string {
	if len(rows) == 0 {
		return "(no rows)\n"
	}
	var b strings.Builder
	b.WriteString(strings.Join(columns, " | "))
	b.WriteString("\n")
	for _, row := range rows {
		b.WriteString(strings.Join(row, " | "))
		b.WriteString("\n")
	}
	return b.String()
}

// ExtractSQL pulls the SQL statement from the LLM's response, handling both
// bare SQL and fenced code blocks. Returns the trimmed SQL string.
func ExtractSQL(raw string) string {
	s := strings.TrimSpace(raw)

	// Strip markdown code fences if present.
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		// Remove opening fence (```sql or ```)
		if len(lines) > 0 {
			lines = lines[1:]
		}
		// Remove closing fence
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) == "```" {
				lines = lines[:i]
				break
			}
		}
		s = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	// Strip trailing semicolons -- ReadOnlyQuery doesn't need them.
	s = strings.TrimRight(s, ";")
	return strings.TrimSpace(s)
}

// dateContext returns a short section telling the LLM what the current date
// is so it can reason about relative time ("last month", "overdue", etc.).
func dateContext(now time.Time) string {
	return fmt.Sprintf("\n\n## Current date\n\nToday is %s.", now.Format("Monday, January 2, 2006"))
}

// ---------- DDL formatting (for SQL generation prompt) ----------

func formatDDL(t TableInfo) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", t.Name))
	for i, c := range t.Columns {
		b.WriteString(fmt.Sprintf("  %s %s", c.Name, c.Type))
		if c.PK {
			b.WriteString(" PRIMARY KEY")
		}
		if c.NotNull {
			b.WriteString(" NOT NULL")
		}
		if i < len(t.Columns)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(");\n")
	return b.String()
}

// ---------- Markdown table formatting (for fallback prompt) ----------

func formatTable(t TableInfo) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("### %s\n", t.Name))
	for _, c := range t.Columns {
		flags := ""
		if c.PK {
			flags += " PK"
		}
		if c.NotNull {
			flags += " NOT NULL"
		}
		b.WriteString(fmt.Sprintf("- %s %s%s\n", c.Name, c.Type, flags))
	}
	return b.String()
}

// ---------- Stage 1: NL → SQL ----------

const sqlSystemPreamble = `You are a SQL generator for a SQLite database. Given a user question in English, output ONLY a single SELECT statement. No explanation, no markdown, no commentary -- just the SQL.

RULES:
1. Output exactly one SELECT statement. Nothing else.
2. Use only tables and columns from the schema below.
3. Only SELECT -- never INSERT, UPDATE, DELETE, DROP, ALTER, or CREATE.
4. Money columns end in "_ct" and store values in cents. Divide by 100.0 for display.
5. Soft-deleted rows have deleted_at IS NOT NULL. Exclude them unless asked about deleted items.
6. For date math, use SQLite date functions (date, julianday, etc.). Use the current date from the "Current date" section -- do NOT hardcode dates.
7. When comparing dates (past vs future, oldest, newest, overdue), always compare against the current date provided above.
8. If the question cannot be answered from the schema, output: SELECT 'I cannot answer that from the available data' AS answer`

const sqlSchemaNotes = `
Notes:
- Maintenance scheduling: next_due = date(last_serviced, '+' || interval_months || ' months')
- Project statuses: ideating, planned, quoted, underway, delayed, completed, abandoned
- Warranty expiry is in the warranty_exp column (date string)`

const sqlFewShot = `## Examples

User: How many projects are underway?
SQL: SELECT COUNT(*) AS count FROM projects WHERE status = 'underway' AND deleted_at IS NULL

User: What's my most expensive project?
SQL: SELECT name, budget_ct / 100.0 AS budget_dollars FROM projects WHERE deleted_at IS NULL ORDER BY budget_ct DESC LIMIT 1

User: When is the HVAC filter due?
SQL: SELECT m.name, m.last_serviced, m.interval_months, date(m.last_serviced, '+' || m.interval_months || ' months') AS next_due FROM maintenance_items m WHERE m.name LIKE '%HVAC%' AND m.deleted_at IS NULL

User: Which appliances have expiring warranties in the next 90 days?
SQL: SELECT name, warranty_exp FROM appliances WHERE warranty_exp IS NOT NULL AND warranty_exp BETWEEN date('now') AND date('now', '+90 days') AND deleted_at IS NULL

User: How much have I spent on plumbing?
SQL: SELECT SUM(q.amount_ct) / 100.0 AS total_dollars FROM quotes q JOIN projects p ON q.project_id = p.id WHERE p.project_type = 'plumbing' AND p.deleted_at IS NULL AND q.deleted_at IS NULL

User: Show me all maintenance items and when they're next due
SQL: SELECT name, last_serviced, interval_months, date(last_serviced, '+' || interval_months || ' months') AS next_due FROM maintenance_items WHERE deleted_at IS NULL ORDER BY next_due

Now generate SQL for the user's question.`

// ---------- Stage 2: Results → English ----------

const summarySystemPreamble = `You are a helpful assistant. The user asked a question about their home data. A SQL query was run and the results are below. Summarize the results as a concise, natural-language answer.`

const summaryGuidelines = `RULES:
1. Be concise. One short paragraph or a bullet list.
2. Format money as dollars (e.g. $1,234.56).
3. Format dates in a readable way (e.g. "March 3, 2025" or "3 months ago"). Use the current date above to calculate relative time correctly.
4. If the result set is empty, say you didn't find any matching data.
5. Do NOT show raw SQL or table formatting. Speak naturally.
6. Do NOT invent data that isn't in the results.`

// ---------- Fallback (single-stage) ----------

const fallbackPreamble = `You are micasa-assistant, a factual Q&A bot for a home management app. ` +
	`All data is below. Answer ONLY from this data. If the data doesn't contain the answer, say "I don't see that in your data."

RULES:
1. Be concise. One short paragraph or a bullet list. No preamble.
2. Money values in the data are already formatted as dollars.
3. Dates are YYYY-MM-DD.
4. Do NOT invent, assume, or hallucinate data not shown below.
5. If asked to change data, say: "Use the micasa edit mode to make changes."
6. Do NOT repeat the raw data back. Summarize or answer the specific question.`

const fallbackSchemaNotes = `
Schema notes:
- Soft-deleted rows have a non-NULL deleted_at and should be treated as removed.
- Maintenance scheduling: next_due = last_serviced + interval_months.
- Project statuses: ideating, planned, quoted, underway, delayed, completed, abandoned.`

const fallbackGuidelines = `## How to answer
Look at the data above, find the relevant rows, and answer the question directly.

Example question: "How many projects are underway?"
Example answer: "You have 2 projects underway: Kitchen Remodel and Deck Repair."

Example question: "What's my most expensive project?"
Example answer: "Kitchen Remodel at $12,500.00."

Now answer the user's question based solely on the data provided.`
