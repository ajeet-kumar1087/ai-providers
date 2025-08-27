package openai

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
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
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
				APIKey:  "sk-1234567890abcdef1234567890abcdef",
				BaseURL: "https://custom.openai.com/v1",
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: AdapterConfig{
				APIKey: "",
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "invalid API key format",
			config: AdapterConfig{
				APIKey: "invalid-key",
			},
			wantErr: true,
			errMsg:  "OpenAI API key should start with 'sk-'",
		},
		{
			name: "short API key",
			config: AdapterConfig{
				APIKey: "sk-short",
			},
			wantErr: true,
			errMsg:  "OpenAI API key appears to be too short",
		},
		{
			name: "negative timeout",
			config: AdapterConfig{
				APIKey:  "sk-1234567890abcdef1234567890abcdef",
				Timeout: -5 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be non-negative",
		},
		{
			name: "invalid temperature",
			config: AdapterConfig{
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
				Temperature: floatPtr(2.5),
			},
			wantErr: true,
			errMsg:  "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid max tokens",
			config: AdapterConfig{
				APIKey:    "sk-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(10000),
			},
			wantErr: true,
			errMsg:  "max tokens exceeds OpenAI limit",
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
				if adapter.Name() != "openai" {
					t.Errorf("Expected name 'openai', got %q", adapter.Name())
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
		APIKey: "sk-1234567890abcdef1234567890abcdef",
	}
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test Name
	if adapter.Name() != "openai" {
		t.Errorf("Expected name 'openai', got %q", adapter.Name())
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
		"function_calling",
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
					"id": "cmpl-test123",
					"object": "text_completion",
					"created": 1677652288,
					"model": "gpt-3.5-turbo-instruct",
					"choices": [
						{
							"text": "Hello! How can I help you today?",
							"index": 0,
							"finish_reason": "stop"
						}
					],
					"usage": {
						"prompt_tokens": 5,
						"completion_tokens": 9,
						"total_tokens": 14
					}
				}`,
			},
		},
	}

	config := AdapterConfig{
		APIKey: "sk-1234567890abcdef1234567890abcdef",
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

	if resp.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %q", resp.FinishReason)
	}

	// Verify request was made correctly
	lastReq := mockClient.GetLastRequest()
	if lastReq == nil {
		t.Fatalf("No request was made")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Expected POST request, got %s", lastReq.Method)
	}

	if !strings.HasSuffix(lastReq.URL.Path, "/completions") {
		t.Errorf("Expected request to /completions endpoint, got %s", lastReq.URL.Path)
	}

	// Verify headers
	if auth := lastReq.Header.Get("Authorization"); auth != "Bearer "+config.APIKey {
		t.Errorf("Expected Authorization header 'Bearer %s', got %q", config.APIKey, auth)
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
					"id": "cmpl-test123",
					"object": "text_completion",
					"created": 1677652288,
					"model": "gpt-3.5-turbo-instruct",
					"choices": [{"text": "Response", "index": 0, "finish_reason": "stop"}],
					"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
				}`,
			},
		},
	}

	config := AdapterConfig{
		APIKey:      "sk-1234567890abcdef1234567890abcdef",
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
		expectTokens *int
	}{
		{
			name: "use request parameters",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(1.5),
				MaxTokens:   intPtr(200),
			},
			expectTemp:   floatPtr(1.5),
			expectTokens: intPtr(200),
		},
		{
			name: "clamp high temperature",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(2.5), // Should be clamped to 2.0
			},
			expectTemp: floatPtr(2.0),
		},
		{
			name: "clamp negative temperature",
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(-0.5), // Should be clamped to 0.0
			},
			expectTemp: floatPtr(0.0),
		},
		{
			name: "clamp high max tokens",
			request: CompletionRequest{
				Prompt:    "Test",
				MaxTokens: intPtr(10000), // Should be clamped to MaxTokenLimit
			},
			expectTokens: intPtr(MaxTokenLimit),
		},
		{
			name: "use config defaults",
			request: CompletionRequest{
				Prompt: "Test",
				// No temperature or max tokens - should use config defaults
			},
			expectTemp:   floatPtr(0.5), // From config
			expectTokens: intPtr(500),   // From config
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

			// Parse the OpenAI request to verify parameter mapping
			var openaiReq OpenAICompletionRequest
			if err := json.Unmarshal(body, &openaiReq); err != nil {
				t.Fatalf("Failed to parse request body: %v", err)
			}

			// Verify temperature mapping
			if tt.expectTemp != nil {
				if openaiReq.Temperature == nil {
					t.Errorf("Expected temperature %v, got nil", *tt.expectTemp)
				} else if *openaiReq.Temperature != *tt.expectTemp {
					t.Errorf("Expected temperature %v, got %v", *tt.expectTemp, *openaiReq.Temperature)
				}
			}

			// Verify max tokens mapping
			if tt.expectTokens != nil {
				if openaiReq.MaxTokens == nil {
					t.Errorf("Expected max tokens %v, got nil", *tt.expectTokens)
				} else if *openaiReq.MaxTokens != *tt.expectTokens {
					t.Errorf("Expected max tokens %v, got %v", *tt.expectTokens, *openaiReq.MaxTokens)
				}
			}
		})
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
				"error": {
					"message": "Invalid API key provided",
					"type": "invalid_request_error",
					"code": "invalid_api_key"
				}
			}`,
			expectedErrType: "authentication",
			expectedMsg:     "Invalid API key provided",
		},
		{
			name:       "rate limit error",
			statusCode: 429,
			responseBody: `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_exceeded",
					"code": "rate_limit"
				}
			}`,
			expectedErrType: "rate_limit",
			expectedMsg:     "Rate limit exceeded",
		},
		{
			name:       "validation error",
			statusCode: 400,
			responseBody: `{
				"error": {
					"message": "Invalid request format",
					"type": "invalid_request_error",
					"code": "invalid_request"
				}
			}`,
			expectedErrType: "validation",
			expectedMsg:     "Invalid request format",
		},
		{
			name:       "server error",
			statusCode: 500,
			responseBody: `{
				"error": {
					"message": "Internal server error",
					"type": "server_error",
					"code": "internal_error"
				}
			}`,
			expectedErrType: "provider",
			expectedMsg:     "Internal server error",
		},
		{
			name:            "invalid JSON response",
			statusCode:      500,
			responseBody:    `{"invalid": json}`,
			expectedErrType: "",
			expectedMsg:     "OpenAI API error (status 500)",
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
				APIKey: "sk-1234567890abcdef1234567890abcdef",
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
			if openaiErr, ok := err.(*Error); ok {
				if tt.expectedErrType != "" && openaiErr.Type != tt.expectedErrType {
					t.Errorf("Expected error type %q, got %q", tt.expectedErrType, openaiErr.Type)
				}

				if !contains(openaiErr.Message, tt.expectedMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.expectedMsg, openaiErr.Message)
				}

				if openaiErr.Provider != "openai" {
					t.Errorf("Expected provider 'openai', got %q", openaiErr.Provider)
				}

				// Check retry after for rate limit errors
				if tt.expectedErrType == "rate_limit" && openaiErr.RetryAfter == nil {
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
	adapter := &OpenAIAdapter{}

	tests := []struct {
		name     string
		response OpenAICompletionResponse
		expected CompletionResponse
	}{
		{
			name: "normal response",
			response: OpenAICompletionResponse{
				Choices: []struct {
					Text         string `json:"text"`
					Index        int    `json:"index"`
					FinishReason string `json:"finish_reason"`
				}{
					{
						Text:         "Hello world!",
						Index:        0,
						FinishReason: "stop",
					},
				},
				Usage: struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				}{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
			},
			expected: CompletionResponse{
				Text: "Hello world!",
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
				FinishReason: "stop",
			},
		},
		{
			name: "empty choices",
			response: OpenAICompletionResponse{
				Choices: []struct {
					Text         string `json:"text"`
					Index        int    `json:"index"`
					FinishReason string `json:"finish_reason"`
				}{},
				Usage: struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				}{
					PromptTokens:     5,
					CompletionTokens: 0,
					TotalTokens:      5,
				},
			},
			expected: CompletionResponse{
				Text: "",
				Usage: Usage{
					PromptTokens:     5,
					CompletionTokens: 0,
					TotalTokens:      5,
				},
				FinishReason: "",
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
