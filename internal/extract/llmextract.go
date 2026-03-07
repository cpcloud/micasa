// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"fmt"
	"strings"

	"github.com/cpcloud/micasa/internal/data"
	"github.com/cpcloud/micasa/internal/llm"
)

// ExtractionPromptInput holds the inputs for building an extraction prompt.
type ExtractionPromptInput struct {
	DocID     uint
	Filename  string
	MIME      string
	SizeBytes int64
	Schema    SchemaContext
	Sources   []TextSource
}

// BuildExtractionPrompt creates the system and user messages for document
// extraction. The system prompt includes the database DDL and existing entity
// rows; the LLM outputs a JSON array of operations.
func BuildExtractionPrompt(in ExtractionPromptInput) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: operationExtractionSystemPrompt(in.Schema)},
		{Role: "user", Content: operationExtractionUserMessage(in)},
	}
}

func operationExtractionSystemPrompt(ctx SchemaContext) string {
	var b strings.Builder
	b.WriteString(operationExtractionPreamble)

	b.WriteString("\n\n## Database schema\n\n")
	b.WriteString(FormatDDLBlock(ctx.DDL, ExtractionTables))

	hasRows := len(ctx.Vendors) > 0 || len(ctx.Projects) > 0 ||
		len(ctx.Appliances) > 0 || len(ctx.MaintenanceItems) > 0 ||
		len(ctx.MaintenanceCategories) > 0 || len(ctx.ProjectTypes) > 0
	if hasRows {
		b.WriteString("\n## Existing rows (use these IDs for foreign keys)\n\n")
		b.WriteString(FormatEntityRows(data.TableVendors, ctx.Vendors))
		b.WriteString(FormatEntityRows(data.TableProjects, ctx.Projects))
		b.WriteString(FormatEntityRows(data.TableAppliances, ctx.Appliances))
		b.WriteString(FormatEntityRows(data.TableMaintenanceItems, ctx.MaintenanceItems))
		b.WriteString(FormatEntityRows(data.TableMaintenanceCategories, ctx.MaintenanceCategories))
		b.WriteString(FormatEntityRows(data.TableProjectTypes, ctx.ProjectTypes))
	}

	b.WriteString("\n")
	b.WriteString(operationExtractionRules())
	return b.String()
}

func operationExtractionUserMessage(in ExtractionPromptInput) string {
	var b strings.Builder
	if in.DocID > 0 {
		fmt.Fprintf(&b, "Document ID: %d\n", in.DocID)
	}
	fmt.Fprintf(&b, "Filename: %s\n", in.Filename)
	fmt.Fprintf(&b, "MIME: %s\n", in.MIME)
	fmt.Fprintf(&b, "Size: %d bytes\n", in.SizeBytes)

	for _, src := range in.Sources {
		if strings.TrimSpace(src.Text) == "" {
			continue
		}
		fmt.Fprintf(&b, "\n---\n\n## Source: %s\n", src.Tool)
		if src.Desc != "" {
			b.WriteString(src.Desc + "\n\n")
		}
		b.WriteString(src.Text)
	}

	return b.String()
}

const operationExtractionPreamble = `You are a document extraction assistant for a home management application. Given a document's metadata and extracted text, output a JSON array of operations to store the extracted information in the database.

Note: In this application, "quotes" means contractor/vendor cost estimates (bids for home projects), not quoted text or quotation marks. Create a quotes row when a document contains a cost estimate from a contractor or vendor, but not when dollar amounts appear in other contexts (e.g. receipts, manuals, general text).

You may receive text from multiple extraction sources. Each source is labeled with its tool and a description. Multiple OCR sources may contain overlapping or duplicate text because different extraction methods (digital text extraction, OCR) process the same pages independently. Deduplicate the information: extract each fact once regardless of how many sources mention it. When multiple sources are present, prefer digital text extraction for clean output, and use OCR output for scanned content. Reconcile any conflicts by trusting the more plausible reading.`

// operationExtractionRules builds the rules and allowed-operations section
// of the extraction prompt dynamically from ExtractionTableDefs so the prompt
// stays in sync with the model definitions.
func operationExtractionRules() string {
	var b strings.Builder
	b.WriteString(operationExtractionRulesStatic)
	b.WriteString("\n\n## Allowed operations\n\n")
	b.WriteString(formatAllowedOps())
	return b.String()
}

// formatAllowedOps generates the "Allowed operations" section from
// ExtractionTableDefs, excluding the documents table (handled separately
// in the static rules).
func formatAllowedOps() string {
	var b strings.Builder
	b.WriteString("Operations array (non-document tables):\n")
	for _, td := range ExtractionTableDefs {
		if td.Table == data.TableDocuments {
			continue
		}
		actions := make([]string, 0, len(td.Actions))
		for _, ad := range td.Actions {
			actions = append(actions, string(ad.Action))
		}
		fmt.Fprintf(&b, "- %s: %s.", td.Table, strings.Join(actions, " or "))
		for _, ad := range td.Actions {
			if ad.Action == ActionUpdate {
				b.WriteString(` Include "id" in data when updating.`)
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(`
Document field (separate from operations array):
- create or update. Include "id" in data when updating an existing document.

No other tables may be written to.`)
	return b.String()
}

const operationExtractionRulesStatic = `## Output format

Output ONLY a JSON object. No code fences, no markdown, no commentary.

The object has two top-level fields:
- "operations" (required): array of entity operations. Each element has "action", "table", and "data".
- "document" (optional): a single object for document create/update. Has "action" and "data" only (no "table" -- it is always "documents").

Example:

{"operations": [
  {"action": "create", "table": "vendors", "data": {"name": "Garcia Plumbing"}},
  {"action": "create", "table": "projects", "data": {"title": "Kitchen Renovation"}},
  {"action": "create", "table": "quotes", "data": {"total_cents": 150000, "project_id": 3, "vendor_id": 5}}
], "document": {"action": "update", "data": {"id": 42, "title": "Invoice", "notes": "Repair"}}}

## Rules

1. Output ONLY valid JSON. No code fences, no markdown, no commentary.
2. Only write fields you can confidently extract. Do not guess.
3. Money values MUST be in CENTS (integer). $1,500.00 = 150000.
4. Dates are ISO 8601: YYYY-MM-DD.
5. Use real IDs from the existing rows above for foreign keys to existing entities. When you need to reference an entity you are creating in the same batch, use the ID it will receive: IDs are assigned sequentially starting from max(existing IDs) + 1 per table.
6. If a vendor is mentioned but does not exist, create it in the operations array before referencing it.
7. If a project is mentioned but does not exist, create it. Same for appliances, maintenance items, and incidents.
8. When a Document ID is provided, use "update" in the "document" field and include "id" in data. When no document exists yet, use "create".
9. To link a document to an entity, set "entity_kind" and "entity_id" in the "document" field.
10. For maintenance schedules (from manuals), create maintenance_items.
11. For contractor/vendor cost estimates (bids, proposals), create quotes with the correct project_id and vendor_id. Incidental dollar amounts (e.g. in receipts or manuals) are not quotes.
12. Only use "create" and "update". No other actions.`

// StripCodeFences removes markdown code fences that LLMs sometimes wrap
// around JSON output. Handles fences anywhere in the text (not just at
// the start), since LLMs may produce commentary before the fenced block.
func StripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")

	// Find the opening fence (``` or ```json etc.) anywhere in the text.
	fenceStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			fenceStart = i
			break
		}
	}
	if fenceStart < 0 {
		return s
	}

	// Find the closing fence after the opening one.
	fenceEnd := -1
	for i := len(lines) - 1; i > fenceStart; i-- {
		if strings.TrimSpace(lines[i]) == "```" {
			fenceEnd = i
			break
		}
	}
	if fenceEnd < 0 {
		// Opening fence but no closing fence: strip the opening and return rest.
		return strings.TrimSpace(strings.Join(lines[fenceStart+1:], "\n"))
	}

	return strings.TrimSpace(strings.Join(lines[fenceStart+1:fenceEnd], "\n"))
}
