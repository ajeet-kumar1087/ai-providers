package utils

import (
	"testing"

	"github.com/ajeet-kumar1087/ai-providers/types"
)

// Test ParameterMapper creation
func TestNewParameterMapper(t *testing.T) {
	mapper := NewParameterMapper(types.ProviderOpenAI, types.ProviderAnthropic)

	if mapper.sourceProvider != types.ProviderOpenAI {
		t.Errorf("sourceProvider = %v, want %v", mapper.sourceProvider, types.ProviderOpenAI)
	}
	if mapper.targetProvider != types.ProviderAnthropic {
		t.Errorf("targetProvider = %v, want %v", mapper.targetProvider, types.ProviderAnthropic)
	}
}

// Test MapTemperature
func TestMapTemperature(t *testing.T) {
	tests := []struct {
		name           string
		sourceProvider ProviderType
		targetProvider ProviderType
		temperature    float64
		expected       float64
	}{
		{
			name:           "OpenAI to Anthropic - scale down high value",
			sourceProvider: types.ProviderOpenAI,
			targetProvider: types.ProviderAnthropic,
			temperature:    1.8,
			expected:       0.9, // 1.8 / 2.0
		},
		{
			name:           "OpenAI to Anthropic - normal value",
			sourceProvider: types.ProviderOpenAI,
			targetProvider: types.ProviderAnthropic,
			temperature:    0.7,
			expected:       0.7,
		},
		{
			name:           "Anthropic to OpenAI - no scaling needed",
			sourceProvider: types.ProviderAnthropic,
			targetProvider: types.ProviderOpenAI,
			temperature:    0.8,
			expected:       0.8,
		},
		{
			name:           "OpenAI to OpenAI - clamp high value",
			sourceProvider: types.ProviderOpenAI,
			targetProvider: types.ProviderOpenAI,
			temperature:    2.5,
			expected:       2.0,
		},
		{
			name:           "negative temperature - clamp to zero",
			sourceProvider: types.ProviderOpenAI,
			targetProvider: types.ProviderAnthropic,
			temperature:    -0.5,
			expected:       0.0,
		},
		{
			name:           "Google to Anthropic - clamp high value",
			sourceProvider: types.ProviderGoogle,
			targetProvider: types.ProviderAnthropic,
			temperature:    1.5,
			expected:       1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewParameterMapper(tt.sourceProvider, tt.targetProvider)
			result := mapper.MapTemperature(tt.temperature)
			if result != tt.expected {
				t.Errorf("MapTemperature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test MapMaxTokens
func TestMapMaxTokens(t *testing.T) {
	tests := []struct {
		name           string
		targetProvider ProviderType
		maxTokens      int
		expected       int
	}{
		{
			name:           "OpenAI - within limit",
			targetProvider: types.ProviderOpenAI,
			maxTokens:      2000,
			expected:       2000,
		},
		{
			name:           "OpenAI - exceeds limit",
			targetProvider: types.ProviderOpenAI,
			maxTokens:      10000,
			expected:       4096, // OpenAI limit
		},
		{
			name:           "Anthropic - within limit",
			targetProvider: types.ProviderAnthropic,
			maxTokens:      50000,
			expected:       50000,
		},
		{
			name:           "Anthropic - exceeds limit",
			targetProvider: types.ProviderAnthropic,
			maxTokens:      200000,
			expected:       100000, // Anthropic limit
		},
		{
			name:           "Google - within limit",
			targetProvider: types.ProviderGoogle,
			maxTokens:      4000,
			expected:       4000,
		},
		{
			name:           "Google - exceeds limit",
			targetProvider: types.ProviderGoogle,
			maxTokens:      20000,
			expected:       8192, // Google limit
		},
		{
			name:           "zero tokens - use default",
			targetProvider: types.ProviderOpenAI,
			maxTokens:      0,
			expected:       1024, // Default
		},
		{
			name:           "negative tokens - use default",
			targetProvider: types.ProviderOpenAI,
			maxTokens:      -100,
			expected:       1024, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewParameterMapper(types.ProviderOpenAI, tt.targetProvider)
			result := mapper.MapMaxTokens(tt.maxTokens)
			if result != tt.expected {
				t.Errorf("MapMaxTokens() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test MapStopSequences
func TestMapStopSequences(t *testing.T) {
	tests := []struct {
		name           string
		targetProvider ProviderType
		stopSequences  []string
		expected       []string
	}{
		{
			name:           "OpenAI - within limit",
			targetProvider: types.ProviderOpenAI,
			stopSequences:  []string{".", "!", "?"},
			expected:       []string{".", "!", "?"},
		},
		{
			name:           "OpenAI - exceeds limit",
			targetProvider: types.ProviderOpenAI,
			stopSequences:  []string{".", "!", "?", ";", ":", ","},
			expected:       []string{".", "!", "?", ";"}, // Truncated to 4
		},
		{
			name:           "Anthropic - within limit",
			targetProvider: types.ProviderAnthropic,
			stopSequences:  []string{".", "!", "?", ";", ":", ",", "\n", "\t"},
			expected:       []string{".", "!", "?", ";", ":", ",", "\n", "\t"},
		},
		{
			name:           "Google - exceeds limit",
			targetProvider: types.ProviderGoogle,
			stopSequences:  []string{".", "!", "?", ";", ":", ",", "\n"},
			expected:       []string{".", "!", "?", ";", ":"}, // Truncated to 5
		},
		{
			name:           "empty stop sequences",
			targetProvider: types.ProviderOpenAI,
			stopSequences:  []string{},
			expected:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewParameterMapper(types.ProviderOpenAI, tt.targetProvider)
			result := mapper.MapStopSequences(tt.stopSequences)
			if !equalStringSlices(result, tt.expected) {
				t.Errorf("MapStopSequences() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test provider limit functions
func TestGetProviderTokenLimit(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected int
	}{
		{types.ProviderOpenAI, 4096},
		{types.ProviderAnthropic, 100000},
		{types.ProviderGoogle, 8192},
		{ProviderType("unknown"), 4096}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetProviderTokenLimit(tt.provider)
			if result != tt.expected {
				t.Errorf("GetProviderTokenLimit(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetProviderMaxTemperature(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected float64
	}{
		{types.ProviderOpenAI, 2.0},
		{types.ProviderAnthropic, 1.0},
		{types.ProviderGoogle, 1.0},
		{ProviderType("unknown"), 1.0}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetProviderMaxTemperature(tt.provider)
			if result != tt.expected {
				t.Errorf("GetProviderMaxTemperature(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetProviderMaxStopSequences(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected int
	}{
		{types.ProviderOpenAI, 4},
		{types.ProviderAnthropic, 10},
		{types.ProviderGoogle, 5},
		{ProviderType("unknown"), 4}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetProviderMaxStopSequences(tt.provider)
			if result != tt.expected {
				t.Errorf("GetProviderMaxStopSequences(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultMaxTokens(t *testing.T) {
	tests := []struct {
		provider ProviderType
		expected int
	}{
		{types.ProviderOpenAI, 1024},
		{types.ProviderAnthropic, 1024},
		{types.ProviderGoogle, 1024},
		{ProviderType("unknown"), 1024}, // Default
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			result := GetDefaultMaxTokens(tt.provider)
			if result != tt.expected {
				t.Errorf("GetDefaultMaxTokens(%v) = %v, want %v", tt.provider, result, tt.expected)
			}
		})
	}
}

// Test ValidateCompletionRequest
func TestValidateCompletionRequest(t *testing.T) {
	tests := []struct {
		name    string
		request types.CompletionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: types.CompletionRequest{
				Prompt:      "Hello, world!",
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(100),
			},
			wantErr: false,
		},
		{
			name: "empty prompt",
			request: types.CompletionRequest{
				Prompt: "",
			},
			wantErr: true,
			errMsg:  "prompt is required and cannot be empty",
		},
		{
			name: "whitespace prompt",
			request: types.CompletionRequest{
				Prompt: "   ",
			},
			wantErr: true,
			errMsg:  "prompt is required and cannot be empty",
		},
		{
			name: "negative temperature",
			request: types.CompletionRequest{
				Prompt:      "Hello",
				Temperature: floatPtr(-0.1),
			},
			wantErr: true,
			errMsg:  "temperature must be non-negative",
		},
		{
			name: "zero max tokens",
			request: types.CompletionRequest{
				Prompt:    "Hello",
				MaxTokens: intPtr(0),
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
		{
			name: "negative max tokens",
			request: types.CompletionRequest{
				Prompt:    "Hello",
				MaxTokens: intPtr(-100),
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompletionRequest(tt.request)
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
				}
			}
		})
	}
}

// Test ValidateChatRequest
func TestValidateChatRequest(t *testing.T) {
	tests := []struct {
		name    string
		request types.ChatRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(100),
			},
			wantErr: false,
		},
		{
			name: "empty messages",
			request: types.ChatRequest{
				Messages: []types.Message{},
			},
			wantErr: true,
			errMsg:  "messages are required",
		},
		{
			name: "nil messages",
			request: types.ChatRequest{
				Messages: nil,
			},
			wantErr: true,
			errMsg:  "messages are required",
		},
		{
			name: "message with empty role",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "", Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "message 0: role is required",
		},
		{
			name: "message with empty content",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: ""},
				},
			},
			wantErr: true,
			errMsg:  "message 0: content is required",
		},
		{
			name: "message with invalid role",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "invalid", Content: "Hello"},
				},
			},
			wantErr: true,
			errMsg:  "message 0: invalid role 'invalid'",
		},
		{
			name: "negative temperature",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: floatPtr(-0.1),
			},
			wantErr: true,
			errMsg:  "temperature must be non-negative",
		},
		{
			name: "zero max tokens",
			request: types.ChatRequest{
				Messages: []types.Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens: intPtr(0),
			},
			wantErr: true,
			errMsg:  "max_tokens must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChatRequest(tt.request)
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
				}
			}
		})
	}
}

// Test ValidateMessage
func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		message types.Message
		index   int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid user message",
			message: types.Message{Role: "user", Content: "Hello"},
			index:   0,
			wantErr: false,
		},
		{
			name:    "valid assistant message",
			message: types.Message{Role: "assistant", Content: "Hi there!"},
			index:   1,
			wantErr: false,
		},
		{
			name:    "valid system message",
			message: types.Message{Role: "system", Content: "You are helpful"},
			index:   0,
			wantErr: false,
		},
		{
			name:    "empty role",
			message: types.Message{Role: "", Content: "Hello"},
			index:   2,
			wantErr: true,
			errMsg:  "message 2: role is required",
		},
		{
			name:    "whitespace role",
			message: types.Message{Role: "   ", Content: "Hello"},
			index:   1,
			wantErr: true,
			errMsg:  "message 1: role is required",
		},
		{
			name:    "empty content",
			message: types.Message{Role: "user", Content: ""},
			index:   0,
			wantErr: true,
			errMsg:  "message 0: content is required",
		},
		{
			name:    "whitespace content",
			message: types.Message{Role: "user", Content: "   "},
			index:   3,
			wantErr: true,
			errMsg:  "message 3: content is required",
		},
		{
			name:    "invalid role",
			message: types.Message{Role: "invalid", Content: "Hello"},
			index:   1,
			wantErr: true,
			errMsg:  "message 1: invalid role 'invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.message, tt.index)
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
				}
			}
		})
	}
}

// Test ClampParameters
func TestClampParameters(t *testing.T) {
	// Test CompletionRequest clamping
	completionReq := types.CompletionRequest{
		Prompt:      "Hello",
		Temperature: floatPtr(2.5),                                // Too high for Anthropic
		MaxTokens:   intPtr(200000),                               // Too high for OpenAI
		Stop:        []string{".", "!", "?", ";", ":", ",", "\n"}, // Too many for OpenAI
	}

	clampedCompletion := ClampParameters(completionReq, types.ProviderOpenAI).(types.CompletionRequest)

	if clampedCompletion.Temperature == nil || *clampedCompletion.Temperature != 2.0 {
		t.Errorf("Temperature not clamped correctly: got %v, want 2.0", clampedCompletion.Temperature)
	}
	if clampedCompletion.MaxTokens == nil || *clampedCompletion.MaxTokens != 4096 {
		t.Errorf("MaxTokens not clamped correctly: got %v, want 4096", clampedCompletion.MaxTokens)
	}
	if len(clampedCompletion.Stop) != 4 {
		t.Errorf("Stop sequences not clamped correctly: got %d, want 4", len(clampedCompletion.Stop))
	}

	// Test ChatRequest clamping
	chatReq := types.ChatRequest{
		Messages: []types.Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: floatPtr(-0.5), // Too low
		MaxTokens:   intPtr(0),      // Invalid
	}

	clampedChat := ClampParameters(chatReq, types.ProviderAnthropic).(types.ChatRequest)

	if clampedChat.Temperature == nil || *clampedChat.Temperature != 0.0 {
		t.Errorf("Temperature not clamped correctly: got %v, want 0.0", clampedChat.Temperature)
	}
	if clampedChat.MaxTokens == nil || *clampedChat.MaxTokens != 1024 {
		t.Errorf("MaxTokens not set to default correctly: got %v, want 1024", clampedChat.MaxTokens)
	}

	// Test unknown type (should return unchanged)
	unknownType := "unknown"
	result := ClampParameters(unknownType, types.ProviderOpenAI)
	if result != unknownType {
		t.Errorf("Unknown type should be returned unchanged")
	}
}

// Helper functions (keeping local copies since this is in a different package)
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
