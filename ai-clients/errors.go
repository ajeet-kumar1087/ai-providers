package main

import (
	"encoding/json"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeAuth       ErrorType = "authentication"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeProvider   ErrorType = "provider"
	ErrorTypeTokenLimit ErrorType = "token_limit"
)

// Error represents a standardized error across all providers
type Error struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Code       string    `json:"code,omitempty"`
	Provider   string    `json:"provider"`
	Wrapped    error     `json:"-"`
	RetryAfter *int      `json:"retry_after,omitempty"` // Seconds for rate limit errors
	TokenCount *int      `json:"token_count,omitempty"` // For token limit errors
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (%s): %s", e.Provider, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Wrapped
}

// Is checks if the error matches the target error
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Type == t.Type && e.Provider == t.Provider
	}
	return false
}

// NewError creates a new standardized error
func NewError(errorType ErrorType, provider, message string) *Error {
	return &Error{
		Type:     errorType,
		Provider: provider,
		Message:  message,
	}
}

// NewErrorWithCode creates a new standardized error with a provider-specific code
func NewErrorWithCode(errorType ErrorType, provider, message, code string) *Error {
	return &Error{
		Type:     errorType,
		Provider: provider,
		Message:  message,
		Code:     code,
	}
}

// WrapError wraps an existing error with standardized error information
func WrapError(err error, errorType ErrorType, provider, message string) *Error {
	return &Error{
		Type:     errorType,
		Provider: provider,
		Message:  message,
		Wrapped:  err,
	}
}

// WrapErrorWithCode wraps an existing error with standardized error information and code
func WrapErrorWithCode(err error, errorType ErrorType, provider, message, code string) *Error {
	return &Error{
		Type:     errorType,
		Provider: provider,
		Message:  message,
		Code:     code,
		Wrapped:  err,
	}
}

// NewRateLimitError creates a rate limit error with retry information
func NewRateLimitError(provider, message string, retryAfter int) *Error {
	return &Error{
		Type:       ErrorTypeRateLimit,
		Provider:   provider,
		Message:    message,
		RetryAfter: &retryAfter,
	}
}

// NewTokenLimitError creates a token limit error with token count information
func NewTokenLimitError(provider, message string, tokenCount int) *Error {
	return &Error{
		Type:       ErrorTypeTokenLimit,
		Provider:   provider,
		Message:    message,
		TokenCount: &tokenCount,
	}
}

// IsRetryable returns true if the error type is retryable
func (e *Error) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeRateLimit, ErrorTypeNetwork:
		return true
	default:
		return false
	}
}

// MapHTTPStatusToErrorType maps HTTP status codes to error types
func MapHTTPStatusToErrorType(statusCode int) ErrorType {
	switch {
	case statusCode == 401 || statusCode == 403:
		return ErrorTypeAuth
	case statusCode == 429:
		return ErrorTypeRateLimit
	case statusCode >= 400 && statusCode < 500:
		return ErrorTypeValidation
	case statusCode >= 500:
		return ErrorTypeProvider
	default:
		return ErrorTypeNetwork
	}
}

// ParseProviderError parses provider-specific error responses
func ParseProviderError(provider string, statusCode int, body []byte) *Error {
	errorType := MapHTTPStatusToErrorType(statusCode)

	switch provider {
	case "openai":
		return parseOpenAIError(errorType, statusCode, body)
	case "anthropic":
		return parseAnthropicError(errorType, statusCode, body)
	case "google":
		return parseGoogleError(errorType, statusCode, body)
	default:
		return NewErrorWithCode(errorType, provider, "Unknown provider error", fmt.Sprintf("%d", statusCode))
	}
}

// parseOpenAIError parses OpenAI-specific error responses
func parseOpenAIError(errorType ErrorType, statusCode int, body []byte) *Error {
	// Basic OpenAI error structure parsing
	var openAIError struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &openAIError); err != nil {
		return NewErrorWithCode(errorType, "openai", "Failed to parse error response", fmt.Sprintf("%d", statusCode))
	}

	message := openAIError.Error.Message
	if message == "" {
		message = "Unknown OpenAI error"
	}

	code := openAIError.Error.Code
	if code == "" {
		code = openAIError.Error.Type
	}

	// Handle rate limiting with retry information
	if errorType == ErrorTypeRateLimit {
		// OpenAI typically includes retry information in headers, but we'll use a default
		return NewRateLimitError("openai", message, 60) // Default 60 seconds
	}

	return NewErrorWithCode(errorType, "openai", message, code)
}

// parseAnthropicError parses Anthropic-specific error responses
func parseAnthropicError(errorType ErrorType, statusCode int, body []byte) *Error {
	// Basic Anthropic error structure parsing
	var anthropicError struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &anthropicError); err != nil {
		return NewErrorWithCode(errorType, "anthropic", "Failed to parse error response", fmt.Sprintf("%d", statusCode))
	}

	message := anthropicError.Error.Message
	if message == "" {
		message = "Unknown Anthropic error"
	}

	code := anthropicError.Error.Type
	if code == "" {
		code = fmt.Sprintf("%d", statusCode)
	}

	// Handle rate limiting
	if errorType == ErrorTypeRateLimit {
		return NewRateLimitError("anthropic", message, 60) // Default 60 seconds
	}

	return NewErrorWithCode(errorType, "anthropic", message, code)
}

// parseGoogleError parses Google AI-specific error responses
func parseGoogleError(errorType ErrorType, statusCode int, body []byte) *Error {
	// Basic Google AI error structure parsing
	var googleError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &googleError); err != nil {
		return NewErrorWithCode(errorType, "google", "Failed to parse error response", fmt.Sprintf("%d", statusCode))
	}

	message := googleError.Error.Message
	if message == "" {
		message = "Unknown Google AI error"
	}

	code := googleError.Error.Status
	if code == "" {
		code = fmt.Sprintf("%d", googleError.Error.Code)
	}

	// Handle rate limiting
	if errorType == ErrorTypeRateLimit {
		return NewRateLimitError("google", message, 60) // Default 60 seconds
	}

	return NewErrorWithCode(errorType, "google", message, code)
}

// ShouldRetry determines if an error should be retried based on error type and attempt count
func ShouldRetry(err error, attemptCount, maxRetries int) bool {
	if attemptCount >= maxRetries {
		return false
	}

	if e, ok := err.(*Error); ok {
		return e.IsRetryable()
	}

	return false
}
