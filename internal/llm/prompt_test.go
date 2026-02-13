// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testTables = []TableInfo{
	{
		Name: "projects",
		Columns: []ColumnInfo{
			{Name: "id", Type: "integer", PK: true},
			{Name: "title", Type: "text", NotNull: true},
			{Name: "budget_ct", Type: "integer"},
			{Name: "status", Type: "text"},
		},
	},
	{
		Name: "appliances",
		Columns: []ColumnInfo{
			{Name: "id", Type: "integer", PK: true},
			{Name: "name", Type: "text", NotNull: true},
		},
	},
}

var testNow = time.Date(2026, 2, 13, 10, 0, 0, 0, time.UTC)

// --- BuildSystemPrompt (fallback) ---

func TestBuildSystemPromptIncludesSchema(t *testing.T) {
	prompt := BuildSystemPrompt(testTables, "", testNow, "")
	assert.Contains(t, prompt, "projects")
	assert.Contains(t, prompt, "id integer PK")
	assert.Contains(t, prompt, "title text NOT NULL")
	assert.Contains(t, prompt, "status text")
	assert.Contains(t, prompt, "home management")
}

func TestBuildSystemPromptIncludesData(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "### projects (3 rows)\n\n- id: 1, title: Fix roof\n", testNow, "")
	assert.Contains(t, prompt, "Fix roof")
	assert.Contains(t, prompt, "Current Data")
}

func TestBuildSystemPromptOmitsDataWhenEmpty(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "", testNow, "")
	assert.NotContains(t, prompt, "Current Data")
}

func TestBuildSystemPromptIncludesCurrentDate(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "", testNow, "")
	assert.Contains(t, prompt, "Friday, February 13, 2026")
}

func TestBuildSystemPromptIncludesExtraContext(t *testing.T) {
	prompt := BuildSystemPrompt(nil, "", testNow, "House is a 1920s craftsman.")
	assert.Contains(t, prompt, "Additional context")
	assert.Contains(t, prompt, "1920s craftsman")
}

// --- BuildSQLPrompt ---

func TestBuildSQLPromptIncludesDDL(t *testing.T) {
	prompt := BuildSQLPrompt(testTables, testNow, "")
	assert.Contains(t, prompt, "CREATE TABLE projects")
	assert.Contains(t, prompt, "id integer PRIMARY KEY")
	assert.Contains(t, prompt, "title text NOT NULL")
	assert.Contains(t, prompt, "budget_ct integer")
	assert.Contains(t, prompt, "CREATE TABLE appliances")
}

func TestBuildSQLPromptIncludesFewShotExamples(t *testing.T) {
	prompt := BuildSQLPrompt(testTables, testNow, "")
	assert.Contains(t, prompt, "SELECT COUNT(*)")
	assert.Contains(t, prompt, "budget_ct / 100.0")
	assert.Contains(t, prompt, "deleted_at IS NULL")
}

func TestBuildSQLPromptIncludesRules(t *testing.T) {
	prompt := BuildSQLPrompt(testTables, testNow, "")
	assert.Contains(t, prompt, "single SELECT statement")
	assert.Contains(t, prompt, "never INSERT")
}

func TestBuildSQLPromptIncludesCurrentDate(t *testing.T) {
	prompt := BuildSQLPrompt(testTables, testNow, "")
	assert.Contains(t, prompt, "Friday, February 13, 2026")
}

func TestBuildSQLPromptIncludesExtraContext(t *testing.T) {
	prompt := BuildSQLPrompt(testTables, testNow, "Budgets are in CAD.")
	assert.Contains(t, prompt, "Additional context")
	assert.Contains(t, prompt, "Budgets are in CAD")
}

// --- BuildSummaryPrompt ---

func TestBuildSummaryPromptIncludesAllParts(t *testing.T) {
	prompt := BuildSummaryPrompt(
		"How many projects?",
		"SELECT COUNT(*) AS count FROM projects",
		"count\n3\n",
		testNow,
		"",
	)
	assert.Contains(t, prompt, "How many projects?")
	assert.Contains(t, prompt, "SELECT COUNT(*)")
	assert.Contains(t, prompt, "count\n3")
	assert.Contains(t, prompt, "concise")
}

func TestBuildSummaryPromptIncludesCurrentDate(t *testing.T) {
	prompt := BuildSummaryPrompt("test", "SELECT 1", "1\n", testNow, "")
	assert.Contains(t, prompt, "Friday, February 13, 2026")
}

func TestBuildSummaryPromptIncludesExtraContext(t *testing.T) {
	prompt := BuildSummaryPrompt("test", "SELECT 1", "1\n", testNow, "Currency is CAD.")
	assert.Contains(t, prompt, "Additional context")
	assert.Contains(t, prompt, "Currency is CAD")
}

// --- FormatResultsTable ---

func TestFormatResultsTableWithRows(t *testing.T) {
	result := FormatResultsTable(
		[]string{"name", "budget"},
		[][]string{
			{"Kitchen", "$5000"},
			{"Deck", "$3000"},
		},
	)
	assert.Contains(t, result, "name | budget")
	assert.Contains(t, result, "Kitchen | $5000")
	assert.Contains(t, result, "Deck | $3000")
}

func TestFormatResultsTableEmpty(t *testing.T) {
	result := FormatResultsTable([]string{"name"}, nil)
	assert.Equal(t, "(no rows)\n", result)
}

// --- ExtractSQL ---

func TestExtractSQLBare(t *testing.T) {
	sql := ExtractSQL("SELECT * FROM projects")
	assert.Equal(t, "SELECT * FROM projects", sql)
}

func TestExtractSQLWithFences(t *testing.T) {
	raw := "```sql\nSELECT * FROM projects;\n```"
	sql := ExtractSQL(raw)
	assert.Equal(t, "SELECT * FROM projects", sql)
}

func TestExtractSQLWithBareBackticks(t *testing.T) {
	raw := "```\nSELECT COUNT(*) FROM appliances\n```"
	sql := ExtractSQL(raw)
	assert.Equal(t, "SELECT COUNT(*) FROM appliances", sql)
}

func TestExtractSQLStripsTrailingSemicolons(t *testing.T) {
	sql := ExtractSQL("SELECT 1;;;")
	assert.Equal(t, "SELECT 1", sql)
}

func TestExtractSQLTrimsWhitespace(t *testing.T) {
	sql := ExtractSQL("  \n  SELECT 1  \n  ")
	assert.Equal(t, "SELECT 1", sql)
}
