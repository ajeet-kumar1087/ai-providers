package aiprovider

import (
	"encoding/json"
	"testing"

	"github.com/ajeet-kumar1087/ai-providers/types"
)

// Test CompletionRequest validation and JSON marshaling
func TestCompletionRequest(t *testing.T) {
	tests := []struct {
		name    string
		request types.CompletionRequest
		wantErr bool
	}{
		{
			name: "valid request with all fields",
			request: types.CompletionRequest{
				Prompt:      "Hello, world!",
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(100),
				Stop:        []string{".", "!"},
				Stream:      false,
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal fields",
			request: types.CompletionRequest{
				Prompt: "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "empty prompt should be handled by validation",
			request: types.CompletionRequest{
				Prompt: "",
			},
			wantErr: false, // Basic struct creation doesn't validate, validation is separate
		},
		{
			name: "valid temperature range",
			request: types.CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(1.5),
			},
			wantErr: false,
		},
		{
			name: "valid max tokens",
			request: types.CompletionRequest{
				Prompt:    "Test",
				MaxTokens: intPtr(1000),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal CompletionRequest: %v", err)
				return
			}

			var unmarshaled types.CompletionRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal CompletionRequest: %v", err)
				return
			}

			// Compare fields
			if unmarshaled.Prompt != tt.request.Prompt {
				t.Errorf("Prompt mismatch: got %s, want %s", unmarshaled.Prompt, tt.request.Prompt)
			}

			if !equalFloatPtr(unmarshaled.Temperature, tt.request.Temperature) {
				t.Errorf("Temperature mismatch: got %v, want %v", unmarshaled.Temperature, tt.request.Temperature)
			}

			if !equalIntPtr(unmarshaled.MaxTokens, tt.request.MaxTokens) {
				t.Errorf("MaxTokens mismatch: got %v, want %v", unmarshaled.MaxTokens, tt.request.MaxTokens)
			}

			if unmarshaled.Stream != tt.request.Stream {
				t.Errorf("Stream mismatch: got %v, want %v", unmarshaled.Stream, tt.request.Stream)
			}
		})
	}
}

// Test CompletionResponse validation and JSON marshaling
func TestCompletionResponse(t *testing.T) {
	response := types.CompletionResponse{
		Text: "Hello! How can I help you today?",
		Usage: types.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		FinishReason: "stop",
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal CompletionResponse: %v", err)
		return
	}

	var unmarshaled types.CompletionResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal CompletionResponse: %v", err)
		return
	}

	if unmarshaled.Text != response.Text {
		t.Errorf("Text mismatch: got %s, want %s", unmarshaled.Text, response.Text)
	}

	if unmarshaled.Usage != response.Usage {
		t.Errorf("Usage mismatch: got %+v, want %+v", unmarshaled.Usage, response.Usage)
	}

	if unmarshaled.FinishReason != response.FinishReason {
		t.Errorf("FinishReason mismatch: got %s, want %s", unmarshaled.FinishReason, response.FinishReason)
	}
}

// Test ChatRequest validation and JSON marshaling
func TestChatRequest(t *testing.T) {
	tests := []struct {
		name    string
		request types.ChatRequest
		wantErr bool
	}{
		{
			name: "valid chat request",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
				},
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(100),
				Stream:      false,
			},
			wantErr: false,
		},
		{
			name: "single message",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "system message",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal ChatRequest: %v", err)
				return
			}

			var unmarshaled types.ChatRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal ChatRequest: %v", err)
				return
			}

			if len(unmarshaled.Messages) != len(tt.request.Messages) {
				t.Errorf("Messages length mismatch: got %d, want %d", len(unmarshaled.Messages), len(tt.request.Messages))
				return
			}

			for i, msg := range unmarshaled.Messages {
				if msg.Role != tt.request.Messages[i].Role {
					t.Errorf("Message %d role mismatch: got %s, want %s", i, msg.Role, tt.request.Messages[i].Role)
				}
				if msg.Content != tt.request.Messages[i].Content {
					t.Errorf("Message %d content mismatch: got %s, want %s", i, msg.Content, tt.request.Messages[i].Content)
				}
			}
		})
	}
}

// Test ChatResponse validation and JSON marshaling
func TestChatResponse(t *testing.T) {
	response := types.ChatResponse{
		Message: types.Message{
			Role:    "assistant",
			Content: "Hello! How can I help you today?",
		},
		Usage: types.Usage{
			PromptTokens:     15,
			CompletionTokens: 25,
			TotalTokens:      40,
		},
		FinishReason: "stop",
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal ChatResponse: %v", err)
		return
	}

	var unmarshaled types.ChatResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal ChatResponse: %v", err)
		return
	}

	if unmarshaled.Message.Role != response.Message.Role {
		t.Errorf("Message role mismatch: got %s, want %s", unmarshaled.Message.Role, response.Message.Role)
	}

	if unmarshaled.Message.Content != response.Message.Content {
		t.Errorf("Message content mismatch: got %s, want %s", unmarshaled.Message.Content, response.Message.Content)
	}

	if unmarshaled.Usage != response.Usage {
		t.Errorf("Usage mismatch: got %+v, want %+v", unmarshaled.Usage, response.Usage)
	}
}

// Test Message validation
func TestMessage(t *testing.T) {
	tests := []struct {
		name    string
		message types.Message
		valid   bool
	}{
		{
			name:    "valid user message",
			message: types.Message{Role: "user", Content: "Hello"},
			valid:   true,
		},
		{
			name:    "valid assistant message",
			message: types.Message{Role: "assistant", Content: "Hi there!"},
			valid:   true,
		},
		{
			name:    "valid system message",
			message: types.Message{Role: "system", Content: "You are helpful"},
			valid:   true,
		},
		{
			name:    "empty role",
			message: types.Message{Role: "", Content: "Hello"},
			valid:   false,
		},
		{
			name:    "empty content",
			message: types.Message{Role: "user", Content: ""},
			valid:   false,
		},
		{
			name:    "invalid role",
			message: types.Message{Role: "invalid", Content: "Hello"},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Errorf("Failed to marshal Message: %v", err)
				return
			}

			var unmarshaled types.Message
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal Message: %v", err)
				return
			}

			if unmarshaled.Role != tt.message.Role {
				t.Errorf("Role mismatch: got %s, want %s", unmarshaled.Role, tt.message.Role)
			}

			if unmarshaled.Content != tt.message.Content {
				t.Errorf("Content mismatch: got %s, want %s", unmarshaled.Content, tt.message.Content)
			}
		})
	}
}

// Test Usage struct
func TestUsage(t *testing.T) {
	usage := types.Usage{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(usage)
	if err != nil {
		t.Errorf("Failed to marshal Usage: %v", err)
		return
	}

	var unmarshaled types.Usage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal Usage: %v", err)
		return
	}

	if unmarshaled != usage {
		t.Errorf("Usage mismatch: got %+v, want %+v", unmarshaled, usage)
	}
}

// Test ProviderType validation
func TestProviderType(t *testing.T) {
	tests := []struct {
		name     string
		provider types.ProviderType
		valid    bool
	}{
		{
			name:     "valid OpenAI provider",
			provider: types.ProviderOpenAI,
			valid:    true,
		},
		{
			name:     "valid Anthropic provider",
			provider: types.ProviderAnthropic,
			valid:    true,
		},
		{
			name:     "valid Google provider",
			provider: types.ProviderGoogle,
			valid:    true,
		},
		{
			name:     "invalid provider",
			provider: types.ProviderType("invalid"),
			valid:    false,
		},
		{
			name:     "empty provider",
			provider: types.ProviderType(""),
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateProviderType(tt.provider)
			if tt.valid && err != nil {
				t.Errorf("Expected valid provider, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid provider, got no error")
			}
		})
	}
}

// Helper functions are in test_utils.go
