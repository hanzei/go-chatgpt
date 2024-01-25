package chatgpt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		request       *ChatCompletionRequest
		expectedError error
	}{
		{
			name:          "Valid request",
			request:       validRequest(),
			expectedError: nil,
		},
		{
			name: "Invalid model",
			request: &ChatCompletionRequest{
				Model:    "invalid-model",
				Messages: validRequest().Messages,
			},
			expectedError: ErrInvalidModel,
		},
		{
			name:          "No messages",
			request:       &ChatCompletionRequest{},
			expectedError: ErrNoMessages,
		},
		{
			name: "Invalid role",
			request: &ChatCompletionRequest{
				Model: GPT35Turbo,
				Messages: []ChatMessage{
					{
						Role:    "invalid-role",
						Content: "Hello",
					},
				},
			},
			expectedError: ErrInvalidRole,
		},
		{
			name: "Invalid temperature",
			request: &ChatCompletionRequest{
				Model:       GPT35Turbo,
				Messages:    validRequest().Messages,
				Temperature: -0.5,
			},
			expectedError: ErrInvalidTemperature,
		},
		{
			name: "Invalid presence penalty",
			request: &ChatCompletionRequest{
				Model:           GPT35Turbo,
				Messages:        validRequest().Messages,
				PresencePenalty: -3,
			},
			expectedError: ErrInvalidPresencePenalty,
		},
		{
			name: "Invalid frequency penalty",
			request: &ChatCompletionRequest{
				Model:            GPT35Turbo,
				Messages:         validRequest().Messages,
				FrequencyPenalty: -3,
			},
			expectedError: ErrInvalidFrequencyPenalty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.request)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func validRequest() *ChatCompletionRequest {
	return &ChatCompletionRequest{
		Model: GPT35Turbo,
		Messages: []ChatMessage{
			{
				Role:    ChatGPTModelRoleUser,
				Content: "Hello",
			},
		},
	}
}

func newTestServerAndClient(t *testing.T) (*httptest.Server, *Client) {
	// Create a new test HTTP server to handle requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{ "id": "chatcmpl-abcd", "object": "chat.completion", "created_at": 0, "choices": [ { "index": 0, "message": { "role": "assistant", "content": "\n\n Sample response" }, "finish_reason": "stop" } ], "usage": { "prompt_tokens": 19, "completion_tokens": 47, "total_tokens": 66 }}`))
		require.NoError(t, err)
	}))

	t.Cleanup(testServer.Close)

	// Create a new client with the test server's URL and a mock API key
	return testServer, &Client{
		client: http.DefaultClient,
		config: &Config{
			BaseURL:        testServer.URL,
			APIKey:         "mock_api_key",
			OrganizationID: "mock_organization_id",
		},
	}
}

func newTestClientWithInvalidResponse(t *testing.T) (*httptest.Server, *Client) {
	// Create a new test HTTP server to handle requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{ fakejson }`))
		require.NoError(t, err)
	}))
	t.Cleanup(testServer.Close)

	// Create a new client with the test server's URL and a mock API key
	return testServer, &Client{
		client: http.DefaultClient,
		config: &Config{
			BaseURL:        testServer.URL,
			APIKey:         "mock_api_key",
			OrganizationID: "mock_organization_id",
		},
	}
}

func newTestClientWithInvalidStatusCode(t *testing.T) (*httptest.Server, *Client) {
	// Create a new test HTTP server to handle requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{ "error": "bad request" }`))
		require.NoError(t, err)
	}))
	t.Cleanup(testServer.Close)

	// Create a new client with the test server's URL and a mock API key
	return testServer, &Client{
		client: http.DefaultClient,
		config: &Config{
			BaseURL:        testServer.URL,
			APIKey:         "mock_api_key",
			OrganizationID: "mock_organization_id",
		},
	}
}

func TestSend(t *testing.T) {
	_, client := newTestServerAndClient(t)

	_, err := client.Send(context.Background(), &ChatCompletionRequest{
		Model: GPT35Turbo,
		Messages: []ChatMessage{
			{
				Role:    ChatGPTModelRoleUser,
				Content: "Hello",
			},
		},
	})
	assert.NoError(t, err)

	_, err = client.Send(context.Background(), &ChatCompletionRequest{
		Model: "invalid model",
		Messages: []ChatMessage{
			{
				Role:    ChatGPTModelRoleUser,
				Content: "Hello",
			},
		},
	})
	assert.Error(t, err)

	_, client = newTestClientWithInvalidResponse(t)

	_, err = client.Send(context.Background(), &ChatCompletionRequest{
		Model: GPT35Turbo,
		Messages: []ChatMessage{
			{
				Role:    ChatGPTModelRoleUser,
				Content: "Hello",
			},
		},
	})
	assert.Error(t, err)

	_, client = newTestClientWithInvalidStatusCode(t)

	_, err = client.Send(context.Background(), &ChatCompletionRequest{
		Model: GPT35Turbo,
		Messages: []ChatMessage{
			{
				Role:    ChatGPTModelRoleUser,
				Content: "Hello",
			},
		},
	})
	assert.Error(t, err)

}

func TestSimpleSend(t *testing.T) {
	_, client := newTestServerAndClient(t)

	_, err := client.SimpleSend(context.Background(), "Hello")
	assert.NoError(t, err)
}
