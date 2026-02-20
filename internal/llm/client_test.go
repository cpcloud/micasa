// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Second

func TestPingSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/models", r.URL.Path)
		_, _ = fmt.Fprint(w, `{"data":[{"id":"qwen3:latest"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "qwen3", "", testTimeout)
	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestPingModelNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[{"id":"llama3:latest"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "qwen3", "", testTimeout)
	err := client.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "ollama pull", "should include actionable remediation")
}

func TestPingServerDown(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", "qwen3", "", testTimeout)
	err := client.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reach")
	assert.Contains(t, err.Error(), "ollama serve", "should include actionable remediation")
}

func TestChatStreamSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintln(
			w,
			`data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
		)
		_, _ = fmt.Fprintln(
			w,
			`data: {"choices":[{"delta":{"content":" world"},"finish_reason":null}]}`,
		)
		_, _ = fmt.Fprintln(w, `data: {"choices":[{"delta":{},"finish_reason":"stop"}]}`)
		_, _ = fmt.Fprintln(w, `data: [DONE]`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	ch, err := client.ChatStream(context.Background(), []Message{
		{Role: "user", Content: "hi"},
	})
	require.NoError(t, err)

	var content string
	for chunk := range ch {
		require.NoError(t, chunk.Err)
		content += chunk.Content
		if chunk.Done {
			break
		}
	}
	assert.Equal(t, "Hello world", content)
}

func TestChatStreamCancellation(t *testing.T) {
	handlerDone := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintln(
			w,
			`data: {"choices":[{"delta":{"content":"start"},"finish_reason":null}]}`,
		)
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
		// Block until the client disconnects.
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	ch, err := client.ChatStream(ctx, []Message{
		{Role: "user", Content: "hi"},
	})
	require.NoError(t, err)

	// Read the first chunk.
	chunk := <-ch
	assert.Equal(t, "start", chunk.Content)

	// Cancel and drain -- channel should close promptly.
	cancel()
	for range ch {
		// drain
	}
	<-handlerDone
}

func TestChatStreamServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "model crashed")
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	_, err := client.ChatStream(context.Background(), []Message{
		{Role: "user", Content: "hi"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestModelAndBaseURL(t *testing.T) {
	client := NewClient("http://localhost:11434/v1/", "qwen3", "", testTimeout)
	assert.Equal(t, "qwen3", client.Model())
	assert.Equal(t, "http://localhost:11434/v1", client.BaseURL())
	assert.Equal(t, testTimeout, client.Timeout())
}

func TestSetModel(t *testing.T) {
	client := NewClient("http://localhost:11434/v1", "qwen3", "", testTimeout)
	assert.Equal(t, "qwen3", client.Model())

	client.SetModel("llama3")
	assert.Equal(t, "llama3", client.Model())
}

func TestListModelsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/models", r.URL.Path)
		_, _ = fmt.Fprint(
			w,
			`{"data":[{"id":"qwen3:latest"},{"id":"llama3:8b"},{"id":"mistral:7b"}]}`,
		)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "qwen3", "", testTimeout)
	models, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"qwen3:latest", "llama3:8b", "mistral:7b"}, models)
}

func TestListModelsServerDown(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", "qwen3", "", testTimeout)
	_, err := client.ListModels(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reach")
}

func TestListModelsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "qwen3", "", testTimeout)
	models, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.Empty(t, models)
}

func TestChatCompleteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_, _ = fmt.Fprint(
			w,
			`{"choices":[{"message":{"content":"SELECT COUNT(*) FROM projects"}}]}`,
		)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	result, err := client.ChatComplete(context.Background(), []Message{
		{Role: "user", Content: "how many projects?"},
	})
	require.NoError(t, err)
	assert.Equal(t, "SELECT COUNT(*) FROM projects", result)
}

func TestChatCompleteServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "model crashed")
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	_, err := client.ChatComplete(context.Background(), []Message{
		{Role: "user", Content: "hi"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestChatCompleteEmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"choices":[]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "test-model", "", testTimeout)
	_, err := client.ChatComplete(context.Background(), []Message{
		{Role: "user", Content: "hi"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestCleanErrorResponseOpenAIStyle(t *testing.T) {
	body := []byte(`{"error": {"message": "model not found", "type": "invalid_request_error"}}`)
	err := cleanErrorResponse(404, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model not found")
	assert.Contains(t, err.Error(), "404")
	assert.NotContains(t, err.Error(), `"error"`)
}

func TestCleanErrorResponseOllamaStyle(t *testing.T) {
	body := []byte(`{"error": "model 'qwen3' not found"}`)
	err := cleanErrorResponse(404, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "qwen3")
	assert.NotContains(t, err.Error(), `"error"`)
}

func TestCleanErrorResponsePlainText(t *testing.T) {
	body := []byte("not found")
	err := cleanErrorResponse(404, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "404")
}

func TestIsLocalServer(t *testing.T) {
	local := NewClient("http://localhost:11434/v1", "qwen3", "", testTimeout)
	assert.True(t, local.IsLocalServer())

	cloud := NewClient(
		"https://api.anthropic.com/v1",
		"claude-sonnet-4-5-20250929",
		"sk-ant-test",
		testTimeout,
	)
	assert.False(t, cloud.IsLocalServer())
}

func TestAuthHeaderSentWithAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = fmt.Fprint(w, `{"data":[{"id":"claude-sonnet-4-5-20250929"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "claude-sonnet-4-5-20250929", "sk-ant-test123", testTimeout)
	err := client.Ping(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer sk-ant-test123", gotAuth)
}

func TestNoAuthHeaderWithoutAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = fmt.Fprint(w, `{"data":[{"id":"qwen3:latest"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "qwen3", "", testTimeout)
	err := client.Ping(context.Background())
	require.NoError(t, err)
	assert.Empty(t, gotAuth)
}

func TestPingModelNotFoundCloud(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":[{"id":"claude-sonnet-4-5-20250929"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "gpt-4o", "sk-test", testTimeout)
	err := client.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
	assert.Contains(t, err.Error(), "check the model name")
	assert.NotContains(t, err.Error(), "ollama", "cloud error should not mention ollama")
}

func TestPingServerDownCloud(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", "claude-sonnet-4-5-20250929", "sk-test", testTimeout)
	err := client.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reach")
	assert.Contains(t, err.Error(), "check your base_url")
	assert.NotContains(t, err.Error(), "ollama", "cloud error should not mention ollama")
}

func TestChatStreamAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintln(
			w,
			`data: {"choices":[{"delta":{"content":"hi"},"finish_reason":"stop"}]}`,
		)
		_, _ = fmt.Fprintln(w, `data: [DONE]`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "claude-sonnet-4-5-20250929", "sk-ant-key", testTimeout)
	ch, err := client.ChatStream(context.Background(), []Message{
		{Role: "user", Content: "hello"},
	})
	require.NoError(t, err)
	for range ch {
	}
	assert.Equal(t, "Bearer sk-ant-key", gotAuth)
}

func TestChatCompleteAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"result"}}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "gpt-4o", "sk-openai-key", testTimeout)
	result, err := client.ChatComplete(context.Background(), []Message{
		{Role: "user", Content: "test"},
	})
	require.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, "Bearer sk-openai-key", gotAuth)
}

func TestListModelsAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = fmt.Fprint(w, `{"data":[{"id":"model-a"}]}`)
	}))
	defer srv.Close()

	client := NewClient(srv.URL+"/v1", "model-a", "sk-key", testTimeout)
	models, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"model-a"}, models)
	assert.Equal(t, "Bearer sk-key", gotAuth)
}

func TestCleanErrorResponseUnparsableJSON(t *testing.T) {
	// Long noisy JSON that doesn't match our expected shape should fall back to
	// generic message without dumping the body.
	body := []byte(`{"status":"failed","details":{"code":42,"trace":["a","b","c"]}}`)
	err := cleanErrorResponse(500, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	// Should NOT dump the raw JSON.
	assert.NotContains(t, err.Error(), `"status"`)
	assert.NotContains(t, err.Error(), `"details"`)
}
