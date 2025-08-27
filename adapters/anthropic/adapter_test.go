package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	httputil "github.com/ajeet-kumar1087/ai-providers/internal/http"
)

// MockHTTPClient implements the HTTPClient interface for testing
type MockHTTPClient struct {
	responses []MockResponse
	requests  []*http.Request
	index     int
}

type MockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Store the request for verification
	m.requests = append(m.requests, req)

	if len(m.responses) == 0 {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader("No mock responses configured")),
		}, nil
	}

	// Use the last response if we've run out (for retries)
	responseIndex := m.index
	if responseIndex >= len(m.responses) {
		responseIndex = len(m.responses) - 1
	} else {
		m.index++
	}

	mockResp := m.responses[responseIndex]

	resp := &http.Response{
		StatusCode: mockResp.StatusCode,
		Body:       io.NopCloser(strings.NewReader(mockResp.Body)),
		Header:     make(http.Header),
	}

	// Set headers
	for key, value := range mockResp.Headers {
		resp.Header.Set(key, value)
	}

	return resp, nil
}

func (m *MockHTTPClient) Reset() {
	m.index = 0
	m.requests = nil
}

func (m *MockHTTPClient) GetLastRequest() *http.Request {
	if len(m.requests) == 0 {
		return nil
	}
	return m.requests[len(m.requests)-1]
}

// Test adapter creation and configuration
func TestNewAdapter(t *testing.T) {
	tests := []struct {
		name    string
		config  AdapterConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: AdapterConfig{
				APIKey:      "sk-ant-1234567890abcdef1234567890abcdef",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
			},
			wantErr: false,
		},
		{
			name: "valid config with custom base URL",
			config: AdapterConfig{
				APIKey:  "sk-ant-1234567890abcdef1234567890abcdef",
				BaseURL: "https://custom.anthropic.com/v1",
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: AdapterConfig{
				APIKey: "",
			},
			wantErr: true,
			errMsg:  "api key is required",
		},
		{
			name: "invalid API key format",
			config: AdapterConfig{
				APIKey: "sk-invalid-key",
			},
			wantErr: true,
			errMsg:  "anthropic API key should start with 'sk-ant-'",
		},
		{
			name: "short API key",
			config: AdapterConfig{
				APIKey: "sk-ant-short",
			},
			wantErr: true,
			errMsg:  "anthropic API key appears to be too short",
		},
		{
			name: "negative timeout",
			config: AdapterConfig{
				APIKey:  "sk-ant-1234567890abcdef1234567890abcdef",
				Timeout: -5 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be non-negative",
		},
		{
			name: "invalid temperature",
			config: AdapterConfig{
				APIKey:      "sk-ant-1234567890abcdef1234567890abcdef",
				Temperature: floatPtr(1.5), // Anthropic max is 1.0
			},
			wantErr: true,
			errMsg:  "temperature must be between 0.0 and 1.0",
		},
		{
			name: "invalid max tokens",
			config: AdapterConfig{
				APIKey:    "sk-ant-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(200000), // Exceeds Anthropic limit
			},
			wantErr: true,
			errMsg:  "max tokens exceeds Anthropic limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return
				}
				if adapter == nil {
					t.Errorf("Expected adapter to be created")
					return
				}

				// Verify adapter properties
				if adapter.Name() != "anthropic" {
					t.Errorf("Expected name 'anthropic', got %q", adapter.Name())
				}

				expectedBaseURL := tt.config.BaseURL
				if expectedBaseURL == "" {
					expectedBaseURL = DefaultBaseURL
				}
				if adapter.baseURL != expectedBaseURL {
					t.Errorf("Expected baseURL %q, got %q", expectedBaseURL, adapter.baseURL)
				}

				if adapter.apiKey != tt.config.APIKey {
					t.Errorf("Expected apiKey %q, got %q", tt.config.APIKey, adapter.apiKey)
				}
			}
		})
	}
}

// Test adapter methods
func TestAdapterMethods(t *testing.T) {
	config := AdapterConfig{
		APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
	}
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test Name
	if adapter.Name() != "anthropic" {
		t.Errorf("Expected name 'anthropic', got %q", adapter.Name())
	}

	// Test SupportedFeatures
	features := adapter.SupportedFeatures()
	expectedFeatures := []string{
		"completion",
		"chat_completion",
		"streaming",
		"temperature",
		"max_tokens",
		"stop_sequences",
		"system_messages",
	}

	if len(features) != len(expectedFeatures) {
		t.Errorf("Expected %d features, got %d", len(expectedFeatures), len(features))
	}

	for _, expected := range expectedFeatures {
		found := false
		for _, feature := range features {
			if feature == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected feature %q not found", expected)
		}
	}

	// Test ValidateConfig
	err = adapter.ValidateConfig(config)
	if err != nil {
		t.Errorf("Expected valid config, got error: %v", err)
	}

	invalidConfig := AdapterConfig{APIKey: "invalid"}
	err = adapter.ValidateConfig(invalidConfig)
	if err == nil {
		t.Errorf("Expected invalid config to return error")
	}
}

// Test successful completion request
func TestComplete_Success(t *testing.T) {
	mockClient := &MockHTTPClient{
		responses: []MockResponse{
			{
				StatusCode: 200,
				Body: `{
					"id": "msg_test123",
					"type": "message",
					"role": "assistant",
					"content": [
						{
							"type": "text",
							"text": "Hello! How can I help you today?"
						}
					],
					"model": "claude-3-haiku-20240307",
					"stop_reason": "end_turn",
					"stop_sequence": null,
					"usage": {
						"input_tokens": 5,
						"output_tokens": 9
					}
				}`,
			},
		},
	}

	config := AdapterConfig{
		APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
	}
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Replace HTTP client with mock (no retries for predictable testing)
	adapter.httpClient = httputil.NewClientWithHTTPClient(mockClient, 30*time.Second, 0)

	req := CompletionRequest{
		Prompt:      "Hello",
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(100),
		Stop:        []string{"."},
	}

	resp, err := adapter.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected successful completion, got error: %v", err)
	}

	// Verify response
	if resp.Text != "Hello! How can I help you today?" {
		t.Errorf("Expected text 'Hello! How can I help you today?', got %q", resp.Text)
	}

	if resp.Usage.PromptTokens != 5 {
		t.Errorf("Expected prompt tokens 5, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 9 {
		t.Errorf("Expected completion tokens 9, got %d", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 14 {
		t.Errorf("Expected total tokens 14, got %d", resp.Usage.TotalTokens)
	}

	if resp.FinishReason != "end_turn" {
		t.Errorf("Expected finish reason 'end_turn', got %q", resp.FinishReason)
	}

	// Verify request was made correctly
	lastReq := mockClient.GetLastRequest()
	if lastReq == nil {
		t.Fatalf("No request was made")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Expected POST request, got %s", lastReq.Method)
	}

	if !strings.HasSuffix(lastReq.URL.Path, "/messages") {
		t.Errorf("Expected request to /messages endpoint, got %s", lastReq.URL.Path)
	}

	// Verify headers
	if auth := lastReq.Header.Get("x-api-key"); auth != config.APIKey {
		t.Errorf("Expected x-api-key header %q, got %q", config.APIKey, auth)
	}

	if version := lastReq.Header.Get("anthropic-version"); version != APIVersion {
		t.Errorf("Expected anthropic-version header %q, got %q", APIVersion, version)
	}

	if contentType := lastReq.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
	}
}

// Test completion request with parameter mapping
func TestComplete_ParameterMapping(t *testing.T) {
	mockClient := &MockHTTPClient{
		responses: []MockResponse{
			{
				StatusCode: 200,
				Body: `{
					"id": "msg_test123",
					"type": "message",
					"role": "assistant",
					"content": [{"type": "text", "text": "Response"}],
					"model": "claude-3-haiku-20240307",
					"stop_reason": "end_turn",
					"usage": {"input_tokens": 1, "output_tokens": 1}
				}`,
			},
		},
	}

	config := AdapterConfig{
		APIKey:      "sk-ant-1234567890abcdef1234567890abcdef",
		Temperature: floatPtr(0.5), // Default temperature
		MaxTokens:   intPtr(500),   // Default max tokens
	}
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	adapter.httpClient = httputil.NewClientWithHTTPClient(mockClient, 30*time.Second, 0)

	tests := []struct {
		name         string
		request      CompletionRequest
		expectTemp   *float64
		expectTokens int
	}{
		{
			name: "use request parameters",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(0.8),
				MaxTokens:   intPtr(200),
			},
			expectTemp:   floatPtr(0.8),
			expectTokens: 200,
		},
		{
			name: "clamp high temperature",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(1.5), // Should be clamped to 1.0
			},
			expectTemp:   floatPtr(1.0),
			expectTokens: 500, // From config default
		},
		{
			name: "clamp negative temperature",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(-0.5), // Should be clamped to 0.0
			},
			expectTemp:   floatPtr(0.0),
			expectTokens: 500, // From config default
		},
		{
			name: "clamp high max tokens",
			request: CompletionRequest{
				Prompt:    "Test",
				MaxTokens: intPtr(200000), // Should be clamped to MaxTokenLimit
			},
			expectTokens: MaxTokenLimit,
		},
		{
			name: "use config defaults",
			request: CompletionRequest{
				Prompt: "Test",
				// No temperature or max tokens - should use config defaults
			},
			expectTemp:   floatPtr(0.5), // From config
			expectTokens: 500,           // From config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.Reset()

			_, err := adapter.Complete(context.Background(), tt.request)
			if err != nil {
				t.Fatalf("Expected successful completion, got error: %v", err)
			}

			// Verify the mapped request by checking what was sent
			lastReq := mockClient.GetLastRequest()
			if lastReq == nil {
				t.Fatalf("No request was made")
			}

			// Read and parse the request body
			body, err := io.ReadAll(lastReq.Body)
			if err != nil {
				t.Fatalf("Failed to read request body: %v", err)
			}

			// Reset body for potential future reads
			lastReq.Body = io.NopCloser(bytes.NewReader(body))

			// Parse the Anthropic request to verify parameter mapping
			var anthropicReq AnthropicChatCompletionRequest
			if err := json.Unmarshal(body, &anthropicReq); err != nil {
				t.Fatalf("Failed to parse request body: %v", err)
			}

			// Verify temperature mapping
			if tt.expectTemp != nil {
				if anthropicReq.Temperature == nil {
					t.Errorf("Expected temperature %v, got nil", *tt.expectTemp)
				} else if *anthropicReq.Temperature != *tt.expectTemp {
					t.Errorf("Expected temperature %v, got %v", *tt.expectTemp, *anthropicReq.Temperature)
				}
			}

			// Verify max tokens mapping (always required for Anthropic)
			if anthropicReq.MaxTokens != tt.expectTokens {
				t.Errorf("Expected max tokens %v, got %v", tt.expectTokens, anthropicReq.MaxTokens)
			}

			// Verify prompt is converted to user message
			if len(anthropicReq.Messages) != 1 {
				t.Errorf("Expected 1 message, got %d", len(anthropicReq.Messages))
			} else {
				if anthropicReq.Messages[0].Role != "user" {
					t.Errorf("Expected message role 'user', got %q", anthropicReq.Messages[0].Role)
				}
				if anthropicReq.Messages[0].Content != tt.request.Prompt {
					t.Errorf("Expected message content %q, got %q", tt.request.Prompt, anthropicReq.Messages[0].Content)
				}
			}
		})
	}
}

// Test chat completion request
func TestChatComplete_Success(t *testing.T) {
	mockClient := &MockHTTPClient{
		responses: []MockResponse{
			{
				StatusCode: 200,
				Body: `{
					"id": "msg_test123",
					"type": "message",
					"role": "assistant",
					"content": [
						{
							"type": "text",
							"text": "Hello! How can I help you today?"
						}
					],
					"model": "claude-3-haiku-20240307",
					"stop_reason": "end_turn",
					"usage": {
						"input_tokens": 15,
						"output_tokens": 9
					}
				}`,
			},
		},
	}

	config := AdapterConfig{
		APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
	}
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	adapter.httpClient = httputil.NewClientWithHTTPClient(mockClient, 30*time.Second, 0)

	req := ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "Hello"},
		},
		Temperature: floatPtr(0.7),
		MaxTokens:   intPtr(100),
	}

	resp, err := adapter.ChatComplete(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected successful chat completion, got error: %v", err)
	}

	// Verify response
	if resp.Message.Role != "assistant" {
		t.Errorf("Expected message role 'assistant', got %q", resp.Message.Role)
	}

	if resp.Message.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected message content 'Hello! How can I help you today?', got %q", resp.Message.Content)
	}

	if resp.Usage.PromptTokens != 15 {
		t.Errorf("Expected prompt tokens 15, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 9 {
		t.Errorf("Expected completion tokens 9, got %d", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 24 {
		t.Errorf("Expected total tokens 24, got %d", resp.Usage.TotalTokens)
	}

	// Verify request mapping
	lastReq := mockClient.GetLastRequest()
	if lastReq == nil {
		t.Fatalf("No request was made")
	}

	body, err := io.ReadAll(lastReq.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}

	var anthropicReq AnthropicChatCompletionRequest
	if err := json.Unmarshal(body, &anthropicReq); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	// Verify system message is handled separately
	if anthropicReq.System != "You are a helpful assistant" {
		t.Errorf("Expected system message 'You are a helpful assistant', got %q", anthropicReq.System)
	}

	// Verify user message is in messages array
	if len(anthropicReq.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(anthropicReq.Messages))
	} else {
		if anthropicReq.Messages[0].Role != "user" {
			t.Errorf("Expected message role 'user', got %q", anthropicReq.Messages[0].Role)
		}
		if anthropicReq.Messages[0].Content != "Hello" {
			t.Errorf("Expected message content 'Hello', got %q", anthropicReq.Messages[0].Content)
		}
	}
}

// Test error handling
func TestComplete_ErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		responseBody    string
		expectedErrType string
		expectedMsg     string
	}{
		{
			name:       "authentication error",
			statusCode: 401,
			responseBody: `{
				"type": "authentication_error",
				"message": "Invalid API key"
			}`,
			expectedErrType: "authentication",
			expectedMsg:     "Invalid API key",
		},
		{
			name:       "rate limit error",
			statusCode: 429,
			responseBody: `{
				"type": "rate_limit_error",
				"message": "Rate limit exceeded"
			}`,
			expectedErrType: "rate_limit",
			expectedMsg:     "Rate limit exceeded",
		},
		{
			name:       "validation error",
			statusCode: 400,
			responseBody: `{
				"type": "invalid_request_error",
				"message": "Invalid request format"
			}`,
			expectedErrType: "validation",
			expectedMsg:     "Invalid request format",
		},
		{
			name:       "server error",
			statusCode: 500,
			responseBody: `{
				"type": "internal_server_error",
				"message": "Internal server error"
			}`,
			expectedErrType: "provider",
			expectedMsg:     "Internal server error",
		},
		{
			name:            "invalid JSON response",
			statusCode:      500,
			responseBody:    `{"invalid": json}`,
			expectedErrType: "",
			expectedMsg:     "anthropic api error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				responses: []MockResponse{
					{
						StatusCode: tt.statusCode,
						Body:       tt.responseBody,
						Headers: map[string]string{
							"Retry-After": "60",
						},
					},
				},
			}

			config := AdapterConfig{
				APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
			}
			adapter, err := NewAdapter(config)
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			adapter.httpClient = httputil.NewClientWithHTTPClient(mockClient, 30*time.Second, 0)

			req := CompletionRequest{
				Prompt: "Test",
			}

			_, err = adapter.Complete(context.Background(), req)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}

			// Check if it's our custom error type
			if anthropicErr, ok := err.(*Error); ok {
				if tt.expectedErrType != "" && anthropicErr.Type != tt.expectedErrType {
					t.Errorf("Expected error type %q, got %q", tt.expectedErrType, anthropicErr.Type)
				}

				if !contains(anthropicErr.Message, tt.expectedMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.expectedMsg, anthropicErr.Message)
				}

				if anthropicErr.Provider != "anthropic" {
					t.Errorf("Expected provider 'anthropic', got %q", anthropicErr.Provider)
				}

				// Check retry after for rate limit errors
				if tt.expectedErrType == "rate_limit" && anthropicErr.RetryAfter == nil {
					t.Errorf("Expected RetryAfter to be set for rate limit error")
				}
			} else {
				// For non-custom errors, just check the message
				if !contains(err.Error(), tt.expectedMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.expectedMsg, err.Error())
				}
			}
		})
	}
}

// Test response normalization
func TestNormalizeCompletionResponse(t *testing.T) {
	adapter := &AnthropicAdapter{}

	tests := []struct {
		name     string
		response AnthropicChatCompletionResponse
		expected CompletionResponse
	}{
		{
			name: "normal response",
			response: AnthropicChatCompletionResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{
						Type: "text",
						Text: "Hello world!",
					},
				},
				StopReason: "end_turn",
				Usage: struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				}{
					InputTokens:  10,
					OutputTokens: 20,
				},
			},
			expected: CompletionResponse{
				Text: "Hello world!",
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
				FinishReason: "end_turn",
			},
		},
		{
			name: "empty content",
			response: AnthropicChatCompletionResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{},
				StopReason: "max_tokens",
				Usage: struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				}{
					InputTokens:  5,
					OutputTokens: 0,
				},
			},
			expected: CompletionResponse{
				Text: "",
				Usage: Usage{
					PromptTokens:     5,
					CompletionTokens: 0,
					TotalTokens:      5,
				},
				FinishReason: "max_tokens",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.normalizeCompletionResponse(tt.response)

			if result.Text != tt.expected.Text {
				t.Errorf("Expected text %q, got %q", tt.expected.Text, result.Text)
			}

			if result.Usage != tt.expected.Usage {
				t.Errorf("Expected usage %+v, got %+v", tt.expected.Usage, result.Usage)
			}

			if result.FinishReason != tt.expected.FinishReason {
				t.Errorf("Expected finish reason %q, got %q", tt.expected.FinishReason, result.FinishReason)
			}
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
