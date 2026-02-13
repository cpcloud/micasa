// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"

	"github.com/cpcloud/micasa/internal/llm"
	"github.com/stretchr/testify/assert"
)

// TestHandleSQLChunkCompletionUsesCurrentQuery verifies that when SQL
// streaming completes, executeSQLQuery uses CurrentQuery instead of
// attempting to index into the messages array. This is a regression test
// for a panic that occurred when the code tried to access messages[-3]
// after the message structure changed during streaming.
func TestHandleSQLChunkCompletionUsesCurrentQuery(t *testing.T) {
	m := newTestModel()

	// Set up minimal chat state as if streaming has started.
	m.openChat()
	m.chat.CurrentQuery = "test question"
	m.chat.StreamingSQL = true
	m.chat.Streaming = true

	// Add user, notice, and assistant messages (mimics submitChat setup).
	m.chat.Messages = []chatMessage{
		{Role: "user", Content: "test question"},
		{Role: "notice", Content: "generating query"},
		{Role: "assistant", Content: "", SQL: "SELECT * FROM projects"},
	}

	// Simulate SQL streaming completion with valid SQL.
	msg := sqlChunkMsg{
		Content: "",   // All content already accumulated
		Done:    true, // SQL generation complete
		Err:     nil,
	}

	// This is the critical test: handleSQLChunk calls executeSQLQuery which
	// MUST use CurrentQuery. If it tries to access messages[-3] with only 3
	// messages, it would panic with index out of range.
	cmd := m.handleSQLChunk(msg)

	// Should return a command (executeSQLQuery wrapped in tea.Cmd).
	assert.NotNil(t, cmd, "handleSQLChunk should return a command when Done=true")

	// Verify state was updated correctly.
	assert.False(t, m.chat.StreamingSQL, "StreamingSQL should be false after completion")
	assert.Nil(t, m.chat.SQLStreamCh, "SQLStreamCh should be nil after completion")
}

// TestHandleSQLChunkWithNoMessagesDoesNotPanic is a more extreme case
// where the messages array is unexpectedly empty. This should not panic.
func TestHandleSQLChunkWithNoMessagesDoesNotPanic(t *testing.T) {
	m := newTestModel()

	m.openChat()
	m.chat.CurrentQuery = "test question"
	m.chat.StreamingSQL = true
	m.chat.Streaming = true

	// Empty messages array - unexpected state but shouldn't panic.
	m.chat.Messages = []chatMessage{}

	msg := sqlChunkMsg{
		Done: true,
		Err:  nil,
	}

	// Should not panic even with no messages.
	cmd := m.handleSQLChunk(msg)

	// Should return nil because SQL extraction will fail (no assistant message).
	assert.Nil(t, cmd, "handleSQLChunk should handle empty messages gracefully")
	assert.False(t, m.chat.Streaming, "Streaming should be false after error")
}

// TestSQLStreamStartedStoresCurrentQuery verifies that when SQL streaming
// starts, the question is stored in CurrentQuery.
func TestSQLStreamStartedStoresCurrentQuery(t *testing.T) {
	m := newTestModel()
	m.openChat()

	testQuestion := "how much did I spend on projects?"

	// Create a mock stream channel.
	ch := make(chan llm.StreamChunk, 1)
	close(ch) // Close immediately since we're not actually streaming.

	msg := sqlStreamStartedMsg{
		Question: testQuestion,
		Channel:  ch,
		CancelFn: func() {},
		Err:      nil,
	}

	_ = m.handleSQLStreamStarted(msg)

	assert.Equal(
		t,
		testQuestion,
		m.chat.CurrentQuery,
		"CurrentQuery should be set from sqlStreamStartedMsg",
	)
}
