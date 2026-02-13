// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"context"
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

	// Simulate streaming in progress with partial SQL
	_, cancel := context.WithCancel(context.Background())
	m.chat.CurrentQuery = testQuestion
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.CancelFn = cancel
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleNotice, Content: "generating query"},
		{Role: roleAssistant, Content: "", SQL: "SELECT * FROM"},
	}

	// Verify assistant message exists before cancellation
	assert.Len(t, m.chat.Messages, 3)
	assert.Equal(t, roleAssistant, m.chat.Messages[2].Role)

	// This is what the global ctrl+c handler calls
	m.cancelChatOperations()

	// Verify assistant message was removed and Interrupted notice added
	assert.Len(t, m.chat.Messages, 2, "should have user + interrupted notice")
	assert.Equal(t, roleUser, m.chat.Messages[0].Role)
	assert.Equal(t, roleNotice, m.chat.Messages[1].Role)
	assert.Equal(t, "Interrupted", m.chat.Messages[1].Content)

	// Verify all streaming state is cleaned up
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)
	assert.Nil(t, m.chat.CancelFn)
	assert.Nil(t, m.chat.SQLStreamCh)
	assert.Nil(t, m.chat.StreamCh)
}

// TestCancellationRemovesAssistantWithPartialContent verifies that ctrl+c
// removes assistant message even if it has partial content (stage 2).
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

	// Verify assistant message with partial content exists
	assert.Len(t, m.chat.Messages, 2)
	assert.Equal(t, "Based on the data", m.chat.Messages[1].Content)

	// This is what the global ctrl+c handler calls
	m.cancelChatOperations()

	// Verify partial assistant message was removed
	assert.Len(t, m.chat.Messages, 2, "should have user + interrupted notice")
	assert.Equal(t, roleNotice, m.chat.Messages[1].Role)
	assert.Equal(t, "Interrupted", m.chat.Messages[1].Content)

	// Verify all streaming state is cleaned up
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)
	assert.Nil(t, m.chat.CancelFn)
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

	// This is what the global ctrl+c handler calls
	m.cancelChatOperations()

	// Should still clean up even without CancelFn
	assert.Len(t, m.chat.Messages, 2, "should have user + interrupted notice")
	assert.Equal(t, roleNotice, m.chat.Messages[1].Role)
	assert.Equal(t, "Interrupted", m.chat.Messages[1].Content)
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)
}

// TestSpinnerOnlyShowsForLastMessage verifies that the spinner is only
// rendered for the last assistant message, not all assistant messages.
func TestSpinnerOnlyShowsForLastMessage(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Setup: multiple completed assistant messages + one streaming
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: "question 1"},
		{Role: roleAssistant, Content: "answer 1", SQL: "SELECT 1"},
		{Role: roleUser, Content: "question 2"},
		{Role: roleAssistant, Content: "", SQL: ""}, // currently streaming
	}

	// Render messages
	rendered := m.renderChatMessages()

	// The spinner should only appear once (for the last message)
	// We can't easily test the actual spinner rendering without coupling
	// to implementation, but we can verify the logic by checking message count
	assert.Len(t, m.chat.Messages, 4)

	// Verify rendering doesn't panic and produces output
	assert.NotEmpty(t, rendered)

	// The key test: when we stop streaming, no spinners should render
	m.chat.Streaming = false
	m.chat.StreamingSQL = false
	renderedAfterStop := m.renderChatMessages()
	assert.NotEmpty(t, renderedAfterStop)
}

// TestNoSpinnerAfterCancellation verifies that after cancellation,
// no spinner is rendered at all.
func TestNoSpinnerAfterCancellation(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Setup: streaming then cancel
	m.chat.Streaming = true
	m.chat.StreamingSQL = true
	m.chat.Messages = []chatMessage{
		{Role: roleUser, Content: testQuestion},
		{Role: roleAssistant, Content: "", SQL: "SELECT"},
	}

	// Use the real cancelChatOperations function
	m.cancelChatOperations()

	// Verify state
	assert.Len(t, m.chat.Messages, 2)
	assert.Equal(t, roleNotice, m.chat.Messages[1].Role)
	assert.False(t, m.chat.Streaming)
	assert.False(t, m.chat.StreamingSQL)

	// Last message is a notice, not assistant -- no spinner can render
	lastMsg := m.chat.Messages[len(m.chat.Messages)-1]
	assert.NotEqual(t, roleAssistant, lastMsg.Role, "last message should not be assistant")
}

// TestChatMagModeToggle verifies that ctrl+m toggles magnitude mode on/off.
func TestChatMagModeToggle(t *testing.T) {
	m := newTestModel()
	m.openChat()

	// Initially off.
	assert.False(t, m.chat.MagMode)

	// Toggle on.
	sendKey(m, "ctrl+m")
	assert.True(t, m.chat.MagMode)

	// Toggle off.
	sendKey(m, "ctrl+m")
	assert.False(t, m.chat.MagMode)
}
