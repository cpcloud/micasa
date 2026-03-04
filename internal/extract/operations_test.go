// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ParseOperations ---

func TestParseOperations_Valid(t *testing.T) {
	raw := `{"operations": [
		{"action": "create", "table": "vendors", "data": {"name": "Garcia Plumbing"}},
		{"action": "update", "table": "documents", "data": {"title": "Invoice", "notes": "Repair"}}
	]}`
	ops, err := ParseOperations(raw)
	require.NoError(t, err)
	require.Len(t, ops, 2)

	assert.Equal(t, ActionCreate, ops[0].Action)
	assert.Equal(t, "vendors", ops[0].Table)
	assert.Equal(t, "Garcia Plumbing", ops[0].Data["name"])

	assert.Equal(t, ActionUpdate, ops[1].Action)
	assert.Equal(t, "documents", ops[1].Table)
	assert.Equal(t, "Invoice", ops[1].Data["title"])
}

func TestParseOperations_RejectsCodeFences(t *testing.T) {
	raw := "```json\n" + `{"operations": [{"action": "create", "table": "vendors", "data": {"name": "Test"}}]}` + "\n```"
	_, err := ParseOperations(raw)
	assert.Error(t, err)
}

func TestParseOperations_Empty(t *testing.T) {
	_, err := ParseOperations("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty LLM output")
}

func TestParseOperations_InvalidJSON(t *testing.T) {
	_, err := ParseOperations("I don't understand the question")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse operations json")
}

func TestParseOperations_EmptyArray(t *testing.T) {
	_, err := ParseOperations(`{"operations": []}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no operations found")
}

func TestParseOperations_InvalidWrapper(t *testing.T) {
	_, err := ParseOperations(`{"operations": "not an array"}`)
	assert.Error(t, err)
}

func TestParseOperations_RawArrayRejected(t *testing.T) {
	raw := `[{"action": "create", "table": "vendors", "data": {"name": "Test"}}]`
	_, err := ParseOperations(raw)
	assert.Error(t, err, "raw arrays should be rejected; schema requires object wrapper")
}

// --- OperationsSchema ---

func TestOperationsSchema_TopLevel(t *testing.T) {
	schema := OperationsSchema()
	assert.Equal(t, "object", schema["type"])
	assert.Equal(t, false, schema["additionalProperties"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	opsProp, ok := props["operations"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", opsProp["type"])

	items, ok := opsProp["items"].(map[string]any)
	require.True(t, ok)

	variants, ok := items["anyOf"].([]any)
	require.True(t, ok)
	assert.Len(t, variants, 7, "expected 7 variants: 5 create + 2 update")
}

func TestOperationsSchema_VariantStructure(t *testing.T) {
	variants := operationVariants()

	for i, v := range variants {
		variant, ok := v.(map[string]any)
		require.True(t, ok, "variant %d is not a map", i)
		assert.Equal(t, "object", variant["type"])
		assert.Equal(t, false, variant["additionalProperties"])

		required, ok := variant["required"].([]any)
		require.True(t, ok, "variant %d missing required", i)
		assert.Contains(t, required, "action")
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "data")

		props, ok := variant["properties"].(map[string]any)
		require.True(t, ok)

		actionProp, ok := props["action"].(map[string]any)
		require.True(t, ok)
		actionEnum, ok := actionProp["enum"].([]any)
		require.True(t, ok)
		assert.Len(t, actionEnum, 1, "each variant constrains action to one value")

		tableProp, ok := props["table"].(map[string]any)
		require.True(t, ok)
		tableEnum, ok := tableProp["enum"].([]any)
		require.True(t, ok)
		assert.Len(t, tableEnum, 1, "each variant constrains table to one value")

		dataProp, ok := props["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "object", dataProp["type"])
		assert.Equal(t, false, dataProp["additionalProperties"],
			"variant %d data must disallow additional properties", i)
	}
}

func TestOperationsSchema_CoversTables(t *testing.T) {
	type tableAction struct {
		table  string
		action Action
	}

	// Verify variants match ExtractionOps 1:1.
	variants := operationVariants()
	require.Len(t, variants, len(ExtractionOps))

	expected := []tableAction{
		{"vendors", ActionCreate},
		{"appliances", ActionCreate},
		{"quotes", ActionCreate},
		{"maintenance_items", ActionCreate},
		{"maintenance_items", ActionUpdate},
		{"documents", ActionCreate},
		{"documents", ActionUpdate},
	}

	seen := make(map[tableAction]bool, len(ExtractionOps))
	for _, op := range ExtractionOps {
		seen[tableAction{op.Table, op.Action}] = true
	}
	for _, ta := range expected {
		assert.True(t, seen[ta], "missing variant for %s/%s", ta.action, ta.table)
	}
}

func TestOperationsSchema_VendorsCreateColumns(t *testing.T) {
	variant := findVariant(t, ActionCreate, "vendors")
	dataProps := variantDataProps(t, variant)

	expected := []string{"name", "contact_name", "email", "phone", "website", "notes"}
	assert.Len(t, dataProps, len(expected))
	for _, col := range expected {
		_, ok := dataProps[col]
		assert.True(t, ok, "missing column %q", col)
	}

	dataRequired := variantDataRequired(t, variant)
	assert.Contains(t, dataRequired, "name")
}

func TestOperationsSchema_DocumentsUpdateRequiresID(t *testing.T) {
	variant := findVariant(t, ActionUpdate, "documents")
	dataRequired := variantDataRequired(t, variant)
	assert.Contains(t, dataRequired, "id")
}

func TestOperationsSchema_MaintenanceUpdateRequiresID(t *testing.T) {
	variant := findVariant(t, ActionUpdate, "maintenance_items")
	dataRequired := variantDataRequired(t, variant)
	assert.Contains(t, dataRequired, "id")
}

func TestOperationsSchema_EntityKindEnum(t *testing.T) {
	variant := findVariant(t, ActionUpdate, "documents")
	dataProps := variantDataProps(t, variant)

	ekProp, ok := dataProps["entity_kind"].(map[string]any)
	require.True(t, ok)
	ekEnum, ok := ekProp["enum"].([]any)
	require.True(t, ok)
	assert.Contains(t, ekEnum, "project")
	assert.Contains(t, ekEnum, "vendor")
	assert.Contains(t, ekEnum, "maintenance")
}

func TestOperationsSchema_QuotesCreateColumns(t *testing.T) {
	variant := findVariant(t, ActionCreate, "quotes")
	dataProps := variantDataProps(t, variant)
	dataRequired := variantDataRequired(t, variant)

	assert.Contains(t, dataRequired, "total_cents")

	expected := []string{
		"project_id", "vendor_id", "vendor_name",
		"total_cents", "labor_cents", "materials_cents", "notes",
	}
	assert.Len(t, dataProps, len(expected))
	for _, col := range expected {
		_, ok := dataProps[col]
		assert.True(t, ok, "missing column %q", col)
	}
}

// --- schema test helpers ---

// findVariant builds the schema variant for the given {action, table} pair
// by looking up ExtractionOps and calling buildVariant directly.
func findVariant(t *testing.T, action Action, table string) map[string]any {
	t.Helper()
	for _, op := range ExtractionOps {
		if op.Action == action && op.Table == table {
			return buildVariant(op)
		}
	}
	t.Fatalf("no variant for %s/%s", action, table)
	return nil
}

func variantDataProps(t *testing.T, variant map[string]any) map[string]any {
	t.Helper()
	props, ok := variant["properties"].(map[string]any)
	require.True(t, ok, "variant missing properties")
	dataProp, ok := props["data"].(map[string]any)
	require.True(t, ok, "variant missing data")
	dataProps, ok := dataProp["properties"].(map[string]any)
	require.True(t, ok, "data missing properties")
	return dataProps
}

func variantDataRequired(t *testing.T, variant map[string]any) []any {
	t.Helper()
	props, ok := variant["properties"].(map[string]any)
	require.True(t, ok, "variant missing properties")
	dataProp, ok := props["data"].(map[string]any)
	require.True(t, ok, "variant missing data")
	req, _ := dataProp["required"].([]any)
	return req
}

// --- ValidateOperations ---

var testAllowedOps = map[string]AllowedOps{
	"documents":         {Update: true},
	"vendors":           {Insert: true},
	"quotes":            {Insert: true},
	"maintenance_items": {Insert: true},
	"appliances":        {Insert: true},
}

func TestValidateOperations_Valid(t *testing.T) {
	ops := []Operation{
		{Action: ActionCreate, Table: "vendors", Data: map[string]any{"name": "Test"}},
		{Action: ActionUpdate, Table: "documents", Data: map[string]any{"title": "Doc"}},
	}
	err := ValidateOperations(ops, testAllowedOps)
	assert.NoError(t, err)
}

func TestValidateOperations_InvalidAction(t *testing.T) {
	ops := []Operation{
		{
			Action: "delete",
			Table:  "vendors",
			Data:   map[string]any{"id": 1},
		}, //nolint:exhaustive // intentionally invalid
	}
	err := ValidateOperations(ops, testAllowedOps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action must be")
}

func TestValidateOperations_UnknownTable(t *testing.T) {
	ops := []Operation{
		{Action: ActionCreate, Table: "users", Data: map[string]any{"name": "Test"}},
	}
	err := ValidateOperations(ops, testAllowedOps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in the allowed set")
}

func TestValidateOperations_CreateOnUpdateOnlyTable(t *testing.T) {
	ops := []Operation{
		{Action: ActionCreate, Table: "documents", Data: map[string]any{"title": "X"}},
	}
	err := ValidateOperations(ops, testAllowedOps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create not allowed")
}

func TestValidateOperations_UpdateOnInsertOnlyTable(t *testing.T) {
	ops := []Operation{
		{Action: ActionUpdate, Table: "vendors", Data: map[string]any{"name": "X"}},
	}
	err := ValidateOperations(ops, testAllowedOps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update not allowed")
}

func TestValidateOperations_EmptyData(t *testing.T) {
	ops := []Operation{
		{Action: ActionCreate, Table: "vendors", Data: map[string]any{}},
	}
	err := ValidateOperations(ops, testAllowedOps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data must not be empty")
}

// --- ParseUint ---

func TestParseUint(t *testing.T) {
	assert.Equal(t, uint(42), ParseUint(float64(42)))
	assert.Equal(t, uint(42), ParseUint("42"))
	assert.Equal(t, uint(42), ParseUint(" 42 "))
	assert.Equal(t, uint(0), ParseUint(float64(-1)))
	assert.Equal(t, uint(0), ParseUint("abc"))
	assert.Equal(t, uint(0), ParseUint(nil))
}
