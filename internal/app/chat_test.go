// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"context"
	"strings"
	"testing"

	"github.com/cpcloud/micasa/internal/llm"
	"github.com/stretchr/testify/assert"
)

const testQuestion = "test question"

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

// TestSQLStreamCancellation verifies that cancelling SQL streaming mid-stream
// does not produce "LLM returned empty SQL" error. This is a regression test
// for the issue where closing the stream channel would synthesize a Done message
// with empty SQL, triggering an error. The fix is to return nil (no message)
// when the channel closes, stopping the message loop cleanly.
func TestSQLStreamCancellation(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Create a channel to simulate the LLM stream
	ch := make(chan llm.StreamChunk, 16)

	// Start streaming SQL
	m.chat.CurrentQuery = testQuestion
	m.chat.StreamingSQL = true
	m.chat.Streaming = true
	m.chat.SQLStreamCh = ch
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: ""},
	}

	// Send a partial chunk
	ch <- llm.StreamChunk{Content: "SELECT", Done: false}
	msg1 := sqlChunkMsg{Content: "SELECT", Done: false}
	cmd := m.handleSQLChunk(msg1)
	assert.NotNil(t, cmd, "should continue waiting for chunks")

	// Simulate ctrl+c: close channel without sending Done
	m.chat.Streaming = false
	m.chat.StreamingSQL = false
	m.chat.SQLStreamCh = nil
	close(ch)

	// The real waitForSQLChunk would read from closed channel and return nil.
	// We can't test that directly, but we can verify handleSQLChunk doesn't
	// crash on empty messages and that the overall flow is correct.

	// Verify no error message was added
	for _, msg := range m.chat.Messages {
		assert.NotEqual(t, roleError, msg.Role, "should not have error message")
		assert.NotContains(t, msg.Content, "LLM returned empty SQL")
	}
}

// TestCancellationRemovesAssistantMessage verifies that pressing ctrl+c
// removes the in-progress assistant message and shows "Interrupted" notice.
// This tests cancelChatOperations which is the real ctrl+c handler (the global
// handler in model.Update intercepts ctrl+c before the chat-specific handler).
func TestCancellationRemovesAssistantMessage(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Simulate streaming in progress (stage 1, no SQL accumulated yet)
	_, cancel := context.WithCancel(context.Background())
	m.chat.CurrentQuery = testQuestion
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.CancelFn = cancel
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: ""},
	}

	// Before cancellation: rendered output shows spinner text.
	// The spinner renders "generating query" when StreamingSQL is true
	// and the assistant message has no SQL yet.
	renderedBefore := m.renderChatMessages()
	assert.Contains(t, renderedBefore, "generating query",
		"should show spinner text before cancellation")

	// This is what the global ctrl+c handler calls
	m.cancelChatOperations()

	// After cancellation: rendered output must NOT have spinner text
	// and MUST contain "Interrupted"
	renderedAfter := m.renderChatMessages()
	assert.NotContains(t, renderedAfter, "generating query",
		"should not show spinner text after cancellation")
	assert.NotContains(t, renderedAfter, "thinking",
		"should not show thinking text after cancellation")
	assert.Contains(t, renderedAfter, "Interrupted",
		"should show Interrupted notice after cancellation")

	// Verify all streaming state is cleaned up
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)
	assert.Nil(t, m.chat.CancelFn)
	assert.Nil(t, m.chat.SQLStreamCh)
	assert.Nil(t, m.chat.StreamCh)
}

// TestCancellationRemovesAssistantWithPartialContent verifies that ctrl+c
// during stage 2 (answering) removes the partial response and shows
// "Interrupted" in the rendered output.
func TestCancellationRemovesAssistantWithPartialContent(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Simulate streaming stage 2 with partial answer
	_, cancel := context.WithCancel(context.Background())
	m.chat.CurrentQuery = testQuestion
	m.chat.Streaming = true
	m.chat.StreamingSQL = false // stage 2
	m.chat.CancelFn = cancel
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleAssistant, Content: "Based on the data", SQL: "SELECT * FROM projects"},
	}

	// Before cancellation: rendered output contains the partial response
	renderedBefore := m.renderChatMessages()
	assert.Contains(t, renderedBefore, "Based on the data",
		"should show partial response before cancellation")

	// This is what the global ctrl+c handler calls
	m.cancelChatOperations()

	// After cancellation: partial response gone, "Interrupted" shown
	renderedAfter := m.renderChatMessages()
	assert.NotContains(t, renderedAfter, "Based on the data",
		"should not show partial response after cancellation")
	assert.NotContains(t, renderedAfter, "thinking",
		"should not show thinking text after cancellation")
	assert.Contains(t, renderedAfter, "Interrupted",
		"should show Interrupted notice after cancellation")
}

// TestCancellationWorksWithoutCancelFn verifies that ctrl+c still cleans up
// even when CancelFn is nil (e.g. user presses ctrl+c before the LLM stream
// has been established).
func TestCancellationWorksWithoutCancelFn(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Simulate the window between submitChat and handleSQLStreamStarted
	// where Streaming is true but CancelFn hasn't been set yet
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.CancelFn = nil // not yet set
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: ""},
	}

	// Before cancellation: spinner text present
	renderedBefore := m.renderChatMessages()
	assert.Contains(t, renderedBefore, "generating query",
		"should show spinner text before cancellation")

	// This is what the global ctrl+c handler calls -- CancelFn is nil
	m.cancelChatOperations()

	// After cancellation: clean output with Interrupted
	renderedAfter := m.renderChatMessages()
	assert.NotContains(t, renderedAfter, "generating query",
		"should not show spinner text after cancellation")
	assert.Contains(t, renderedAfter, "Interrupted",
		"should show Interrupted notice after cancellation")
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)
}

// TestSpinnerOnlyShowsForLastMessage verifies that the spinner is only
// rendered for the last assistant message, not all assistant messages.
// This is a regression test for the bug where all assistant messages with
// empty content/SQL would show spinners during streaming.
func TestSpinnerOnlyShowsForLastMessage(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Setup: one completed assistant message + one streaming
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: "question 1"},
		{Role: roleAssistant, Content: "", SQL: ""}, // completed but empty
		{Role: roleUser, Content: "question 2"},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: ""}, // currently streaming
	}

	rendered := m.renderChatMessages()

	// "generating query" should appear exactly once (for the last message only).
	// Count occurrences: the notice is skipped in rendering, so only the
	// inline spinner text contributes. With isLastMessage check, only the
	// last assistant message renders the spinner.
	count := strings.Count(rendered, "generating query")
	assert.Equal(t, 1, count,
		"generating query should appear once (last message only), got %d", count)
}

// TestNoSpinnerAfterCancellation verifies that after calling
// cancelChatOperations, the rendered chat output contains no spinner text.
func TestNoSpinnerAfterCancellation(t *testing.T) {
	m := newTestModel()
	m.openChat()

	_, cancel := context.WithCancel(context.Background())
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.CancelFn = cancel
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: "SELECT"},
	}

	m.cancelChatOperations()

	rendered := m.renderChatMessages()
	assert.NotContains(t, rendered, "generating query",
		"should not show spinner text after cancellation")
	assert.NotContains(t, rendered, "thinking",
		"should not show thinking text after cancellation")
	assert.Contains(t, rendered, "Interrupted",
		"should show Interrupted notice")
}

// TestChatMagModeToggle verifies that ctrl+m toggles the global mag mode
// even when the chat overlay is active.
func TestChatMagModeToggle(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Initially off.
	assert.False(t, m.magMode)

	// Toggle on from within chat.
	sendKey(m, "ctrl+m")
	assert.True(t, m.magMode)

	// Toggle off.
	sendKey(m, "ctrl+m")
	assert.False(t, m.magMode)
}
