package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

// Test Error struct creation and methods
func TestError(t *testing.T) {
	tests := []struct {
		name     string
		error    *Error
		expected string
	}{
		{
			name: "basic error",
			error: &Error{
				Type:     ErrorTypeAuth,
				Provider: "openai",
				Message:  "Invalid API key",
			},
			expected: "[openai] authentication: Invalid API key",
		},
		{
			name: "error with code",
			error: &Error{
				Type:     ErrorTypeRateLimit,
				Provider: "anthropic",
				Message:  "Rate limit exceeded",
				Code:     "rate_limit_exceeded",
			},
			expected: "[anthropic] rate_limit (rate_limit_exceeded): Rate limit exceeded",
		},
		{
			name: "error with retry after",
			error: &Error{
				Type:       ErrorTypeRateLimit,
				Provider:   "openai",
				Message:    "Too many requests",
				RetryAfter: intPtr(60),
			},
			expected: "[openai] rate_limit: Too many requests",
		},
		{
			name: "error with token count",
			error: &Error{
				Type:       ErrorTypeTokenLimit,
				Provider:   "anthropic",
				Message:    "Token limit exceeded",
				TokenCount: intPtr(4096),
			},
			expected: "[anthropic] token_limit: Token limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error.Error() != tt.expected {
				t.Errorf("Error() = %q, want %q", tt.error.Error(), tt.expected)
			}
		})
	}
}

// Test error wrapping and unwrapping
func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorTypeNetwork, "openai", "Network error occurred")

	// Test Unwrap
	if wrappedErr.Unwrap() != originalErr {
		t.Errorf("Unwrap() = %v, want %v", wrappedErr.Unwrap(), originalErr)
	}

	// Test errors.Is
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("errors.Is() should return true for wrapped error")
	}

	// Test errors.As
	var targetErr *Error
	if !errors.As(wrappedErr, &targetErr) {
		t.Errorf("errors.As() should return true for Error type")
	}
}

// Test Error.Is method
func TestErrorIs(t *testing.T) {
	err1 := &Error{Type: ErrorTypeAuth, Provider: "openai", Message: "Auth error"}
	err2 := &Error{Type: ErrorTypeAuth, Provider: "openai", Message: "Different message"}
	err3 := &Error{Type: ErrorTypeRateLimit, Provider: "openai", Message: "Rate limit"}
	err4 := &Error{Type: ErrorTypeAuth, Provider: "anthropic", Message: "Auth error"}

	// Same type and provider should match
	if !err1.Is(err2) {
		t.Errorf("err1.Is(err2) should return true (same type and provider)")
	}

	// Different type should not match
	if err1.Is(err3) {
		t.Errorf("err1.Is(err3) should return false (different type)")
	}

	// Different provider should not match
	if err1.Is(err4) {
		t.Errorf("err1.Is(err4) should return false (different provider)")
	}

	// Non-Error type should not match
	if err1.Is(errors.New("regular error")) {
		t.Errorf("err1.Is(regular error) should return false")
	}
}

// Test NewError functions
func TestNewError(t *testing.T) {
	err := NewError(ErrorTypeValidation, "openai", "Validation failed")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeValidation)
	}
	if err.Provider != "openai" {
		t.Errorf("Provider = %v, want %v", err.Provider, "openai")
	}
	if err.Message != "Validation failed" {
		t.Errorf("Message = %v, want %v", err.Message, "Validation failed")
	}
}

// Test NewErrorWithCode
func TestNewErrorWithCode(t *testing.T) {
	err := NewErrorWithCode(ErrorTypeProvider, "anthropic", "Server error", "500")

	if err.Type != ErrorTypeProvider {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeProvider)
	}
	if err.Provider != "anthropic" {
		t.Errorf("Provider = %v, want %v", err.Provider, "anthropic")
	}
	if err.Message != "Server error" {
		t.Errorf("Message = %v, want %v", err.Message, "Server error")
	}
	if err.Code != "500" {
		t.Errorf("Code = %v, want %v", err.Code, "500")
	}
}

// Test NewRateLimitError
func TestNewRateLimitError(t *testing.T) {
	err := NewRateLimitError("openai", "Rate limit exceeded", 120)

	if err.Type != ErrorTypeRateLimit {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeRateLimit)
	}
	if err.Provider != "openai" {
		t.Errorf("Provider = %v, want %v", err.Provider, "openai")
	}
	if err.Message != "Rate limit exceeded" {
		t.Errorf("Message = %v, want %v", err.Message, "Rate limit exceeded")
	}
	if err.RetryAfter == nil || *err.RetryAfter != 120 {
		t.Errorf("RetryAfter = %v, want %v", err.RetryAfter, 120)
	}
}

// Test NewTokenLimitError
func TestNewTokenLimitError(t *testing.T) {
	err := NewTokenLimitError("anthropic", "Token limit exceeded", 8192)

	if err.Type != ErrorTypeTokenLimit {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeTokenLimit)
	}
	if err.Provider != "anthropic" {
		t.Errorf("Provider = %v, want %v", err.Provider, "anthropic")
	}
	if err.Message != "Token limit exceeded" {
		t.Errorf("Message = %v, want %v", err.Message, "Token limit exceeded")
	}
	if err.TokenCount == nil || *err.TokenCount != 8192 {
		t.Errorf("TokenCount = %v, want %v", err.TokenCount, 8192)
	}
}

// Test IsRetryable method
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		retryable bool
	}{
		{"rate limit is retryable", ErrorTypeRateLimit, true},
		{"network is retryable", ErrorTypeNetwork, true},
		{"auth is not retryable", ErrorTypeAuth, false},
		{"validation is not retryable", ErrorTypeValidation, false},
		{"provider is not retryable", ErrorTypeProvider, false},
		{"token limit is not retryable", ErrorTypeTokenLimit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Type: tt.errorType}
			if err.IsRetryable() != tt.retryable {
				t.Errorf("IsRetryable() = %v, want %v", err.IsRetryable(), tt.retryable)
			}
		})
	}
}

// Test MapHTTPStatusToErrorType
func TestMapHTTPStatusToErrorType(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   ErrorType
	}{
		{401, ErrorTypeAuth},
		{403, ErrorTypeAuth},
		{429, ErrorTypeRateLimit},
		{400, ErrorTypeValidation},
		{422, ErrorTypeValidation},
		{500, ErrorTypeProvider},
		{502, ErrorTypeProvider},
		{503, ErrorTypeProvider},
		{200, ErrorTypeNetwork}, // Default case
		{0, ErrorTypeNetwork},   // Default case
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			result := MapHTTPStatusToErrorType(tt.statusCode)
			if result != tt.expected {
				t.Errorf("MapHTTPStatusToErrorType(%d) = %v, want %v", tt.statusCode, result, tt.expected)
			}
		})
	}
}

// Test ParseProviderError with OpenAI format
func TestParseProviderErrorOpenAI(t *testing.T) {
	openAIErrorJSON := `{
		"error": {
			"message": "Invalid API key provided",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`

	err := ParseProviderError("openai", 401, []byte(openAIErrorJSON))

	if err.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuth)
	}
	if err.Provider != "openai" {
		t.Errorf("Provider = %v, want %v", err.Provider, "openai")
	}
	if err.Message != "Invalid API key provided" {
		t.Errorf("Message = %v, want %v", err.Message, "Invalid API key provided")
	}
	if err.Code != "invalid_api_key" {
		t.Errorf("Code = %v, want %v", err.Code, "invalid_api_key")
	}
}

// Test ParseProviderError with Anthropic format
func TestParseProviderErrorAnthropic(t *testing.T) {
	anthropicErrorJSON := `{
		"error": {
			"type": "authentication_error",
			"message": "Invalid API key"
		}
	}`

	err := ParseProviderError("anthropic", 401, []byte(anthropicErrorJSON))

	if err.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuth)
	}
	if err.Provider != "anthropic" {
		t.Errorf("Provider = %v, want %v", err.Provider, "anthropic")
	}
	if err.Message != "Invalid API key" {
		t.Errorf("Message = %v, want %v", err.Message, "Invalid API key")
	}
	if err.Code != "authentication_error" {
		t.Errorf("Code = %v, want %v", err.Code, "authentication_error")
	}
}

// Test ParseProviderError with Google format
func TestParseProviderErrorGoogle(t *testing.T) {
	googleErrorJSON := `{
		"error": {
			"code": 401,
			"message": "Request had invalid authentication credentials",
			"status": "UNAUTHENTICATED"
		}
	}`

	err := ParseProviderError("google", 401, []byte(googleErrorJSON))

	if err.Type != ErrorTypeAuth {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeAuth)
	}
	if err.Provider != "google" {
		t.Errorf("Provider = %v, want %v", err.Provider, "google")
	}
	if err.Message != "Request had invalid authentication credentials" {
		t.Errorf("Message = %v, want %v", err.Message, "Request had invalid authentication credentials")
	}
	if err.Code != "UNAUTHENTICATED" {
		t.Errorf("Code = %v, want %v", err.Code, "UNAUTHENTICATED")
	}
}

// Test ParseProviderError with invalid JSON
func TestParseProviderErrorInvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`

	err := ParseProviderError("openai", 500, []byte(invalidJSON))

	if err.Type != ErrorTypeProvider {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeProvider)
	}
	if err.Provider != "openai" {
		t.Errorf("Provider = %v, want %v", err.Provider, "openai")
	}
	if err.Message != "Failed to parse error response" {
		t.Errorf("Message = %v, want %v", err.Message, "Failed to parse error response")
	}
}

// Test ParseProviderError with unknown provider
func TestParseProviderErrorUnknownProvider(t *testing.T) {
	err := ParseProviderError("unknown", 500, []byte("{}"))

	if err.Type != ErrorTypeProvider {
		t.Errorf("Type = %v, want %v", err.Type, ErrorTypeProvider)
	}
	if err.Provider != "unknown" {
		t.Errorf("Provider = %v, want %v", err.Provider, "unknown")
	}
	if err.Message != "Unknown provider error" {
		t.Errorf("Message = %v, want %v", err.Message, "Unknown provider error")
	}
}

// Test ShouldRetry function
func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		attemptCount int
		maxRetries   int
		expected     bool
	}{
		{
			name:         "retryable error within limit",
			err:          &Error{Type: ErrorTypeRateLimit},
			attemptCount: 1,
			maxRetries:   3,
			expected:     true,
		},
		{
			name:         "retryable error at limit",
			err:          &Error{Type: ErrorTypeRateLimit},
			attemptCount: 3,
			maxRetries:   3,
			expected:     false,
		},
		{
			name:         "non-retryable error",
			err:          &Error{Type: ErrorTypeAuth},
			attemptCount: 1,
			maxRetries:   3,
			expected:     false,
		},
		{
			name:         "non-Error type",
			err:          errors.New("regular error"),
			attemptCount: 1,
			maxRetries:   3,
			expected:     false,
		},
		{
			name:         "network error within limit",
			err:          &Error{Type: ErrorTypeNetwork},
			attemptCount: 2,
			maxRetries:   3,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRetry(tt.err, tt.attemptCount, tt.maxRetries)
			if result != tt.expected {
				t.Errorf("ShouldRetry() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test Error JSON marshaling
func TestErrorJSONMarshaling(t *testing.T) {
	err := &Error{
		Type:       ErrorTypeRateLimit,
		Provider:   "openai",
		Message:    "Rate limit exceeded",
		Code:       "rate_limit_exceeded",
		RetryAfter: intPtr(60),
	}

	// Test marshaling
	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Errorf("Failed to marshal Error: %v", marshalErr)
		return
	}

	// Test unmarshaling
	var unmarshaled Error
	unmarshalErr := json.Unmarshal(data, &unmarshaled)
	if unmarshalErr != nil {
		t.Errorf("Failed to unmarshal Error: %v", unmarshalErr)
		return
	}

	// Compare fields (excluding Wrapped which is not marshaled)
	if unmarshaled.Type != err.Type {
		t.Errorf("Type mismatch: got %v, want %v", unmarshaled.Type, err.Type)
	}
	if unmarshaled.Provider != err.Provider {
		t.Errorf("Provider mismatch: got %v, want %v", unmarshaled.Provider, err.Provider)
	}
	if unmarshaled.Message != err.Message {
		t.Errorf("Message mismatch: got %v, want %v", unmarshaled.Message, err.Message)
	}
	if unmarshaled.Code != err.Code {
		t.Errorf("Code mismatch: got %v, want %v", unmarshaled.Code, err.Code)
	}
	if !equalIntPtr(unmarshaled.RetryAfter, err.RetryAfter) {
		t.Errorf("RetryAfter mismatch: got %v, want %v", unmarshaled.RetryAfter, err.RetryAfter)
	}
}

// Helper functions are in test_utils.go
