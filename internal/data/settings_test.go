// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSettingMissing(t *testing.T) {
	store := newTestStore(t)
	val, err := store.GetSetting("nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "", val)
}

func TestPutAndGetSetting(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.PutSetting("color", "blue"))

	val, err := store.GetSetting("color")
	require.NoError(t, err)
	assert.Equal(t, "blue", val)
}

func TestPutSettingUpserts(t *testing.T) {
	store := newTestStore(t)
	require.NoError(t, store.PutSetting("color", "blue"))
	require.NoError(t, store.PutSetting("color", "red"))

	val, err := store.GetSetting("color")
	require.NoError(t, err)
	assert.Equal(t, "red", val)
}

func TestLastModelRoundTrip(t *testing.T) {
	store := newTestStore(t)

	// Initially empty.
	model, err := store.GetLastModel()
	require.NoError(t, err)
	assert.Equal(t, "", model)

	// Set and retrieve.
	require.NoError(t, store.PutLastModel("qwen3:8b"))
	model, err = store.GetLastModel()
	require.NoError(t, err)
	assert.Equal(t, "qwen3:8b", model)

	// Overwrite.
	require.NoError(t, store.PutLastModel("llama3.3"))
	model, err = store.GetLastModel()
	require.NoError(t, err)
	assert.Equal(t, "llama3.3", model)
}

func TestAppendChatInputAndLoad(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.AppendChatInput("how many projects?"))
	require.NoError(t, store.AppendChatInput("oldest appliance?"))

	history, err := store.LoadChatHistory()
	require.NoError(t, err)
	assert.Equal(t, []string{"how many projects?", "oldest appliance?"}, history)
}

func TestAppendChatInputDeduplicatesConsecutive(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.AppendChatInput("hello"))
	require.NoError(t, store.AppendChatInput("hello"))
	require.NoError(t, store.AppendChatInput("hello"))

	history, err := store.LoadChatHistory()
	require.NoError(t, err)
	assert.Equal(t, []string{"hello"}, history)
}

func TestAppendChatInputAllowsNonConsecutiveDuplicates(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.AppendChatInput("a"))
	require.NoError(t, store.AppendChatInput("b"))
	require.NoError(t, store.AppendChatInput("a"))

	history, err := store.LoadChatHistory()
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "a"}, history)
}

func TestLoadChatHistoryEmpty(t *testing.T) {
	store := newTestStore(t)

	history, err := store.LoadChatHistory()
	require.NoError(t, err)
	assert.Empty(t, history)
}
