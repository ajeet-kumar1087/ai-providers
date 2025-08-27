package aiprovider

import (
	"encoding/json"
	"fmt"
)

// ErrorType represents the category of error that occurred.
//
// This enumeration provides a standardized way to categorize errors
// across different AI providers, enabling consistent error handling
// and retry logic regardless of the underlying provider.
type ErrorType string

const (
	// ErrorTypeAuth indicates authentication or authorization failures.
	// This includes invalid API keys, expired tokens, or insufficient permissions.
	ErrorTypeAuth ErrorType = "authentication"

	// ErrorTypeRateLimit indicates that the request was rejected due to rate limiting.
	// The RetryAfter field may contain suggested retry timing in seconds.
	ErrorTypeRateLimit ErrorType = "rate_limit"

	// ErrorTypeNetwork indicates network-related failures.
	// This includes connection timeouts, DNS resolution failures, or network unreachability.
	ErrorTypeNetwork ErrorType = "network"

	// ErrorTypeValidation indicates client-side validation failures.
	// This includes invalid parameters, malformed requests, or constraint violations.
	ErrorTypeValidation ErrorType = "validation"

	// ErrorTypeProvider indicates provider-side errors.
	// This includes internal server errors, service unavailability, or provider-specific issues.
	ErrorTypeProvider ErrorType = "provider"

	// ErrorTypeTokenLimit indicates that the request exceeded token limits.
	// The TokenCount field may contain the actual token count that caused the error.
	ErrorTypeTokenLimit ErrorType = "token_limit"
)

// Error represents a standardized error across all AI providers.
//
// This struct provides a consistent error interface that wraps provider-specific
// errors with additional context and categorization. It implements the standard
// Go error interface and provides additional methods for error inspection and
// retry logic.
//
// The error includes provider-specific codes and messages while maintaining
// a consistent structure for error handling across different providers.
type Error struct {
	// Type categorizes the error for consistent handling across providers
	Type ErrorType `json:"type"`

	// Message provides a human-readable description of the error
	Message string `json:"message"`

	// Code contains the provider-specific error code (optional)
	Code string `json:"code,omitempty"`

	// Provider identifies which AI provider generated this error
	Provider string `json:"provider"`

	// Wrapped contains the original error from the provider (not serialized)
	Wrapped error `json:"-"`

	// RetryAfter suggests retry timing in seconds for rate limit errors (optional)
	RetryAfter *int `json:"retry_after,omitempty"`

	// TokenCount contains the token count for token limit errors (optional)
	TokenCount *int `json:"token_count,omitempty"`
}

// Error implements the standard Go error interface.
//
// Returns a formatted string representation of the error that includes
// the provider name, error type, optional code, and message. This provides
// a consistent format for error display and logging.
//
// Format: "[provider] type (code): message" or "[provider] type: message"
//
// Example output:
//   - "[openai] authentication (invalid_api_key): Invalid API key provided"
//   - "[anthropic] rate_limit: Request rate limit exceeded"
//
// Returns:
//   - string: Formatted error message
func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (%s): %s", e.Provider, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Type, e.Message)
}

// Unwrap returns the original wrapped error.
//
// This method enables Go's error unwrapping functionality, allowing
// the use of errors.Is() and errors.As() to inspect the underlying
// provider-specific error while maintaining the standardized wrapper.
//
// Returns:
//   - error: The original error from the provider, or nil if none was wrapped
func (e *Error) Unwrap() error {
	return e.Wrapped
}

// Is checks if this error matches the target error.
//
// This method enables the use of errors.Is() for error comparison.
// Two Error instances are considered equal if they have the same
// Type and Provider, regardless of message or code differences.
//
// Example:
//
//	authErr := NewError(ErrorTypeAuth, "openai", "Invalid key")
//	if errors.Is(err, authErr) {
//		// Handle authentication error
//	}
//
// Parameters:
//   - target: The error to compare against
//
// Returns:
//   - bool: true if the errors match by type and provider
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Type == t.Type && e.Provider == t.Provider
	}
	return false
}

// NewError creates a new standardized error with the specified type, provider, and message.
//
// This is the basic constructor for creating standardized errors without
// wrapping an existing error or including provider-specific codes.
//
// Example:
//
//	err := NewError(ErrorTypeValidation, "openai", "Temperature must be between 0 and 2")
//
// Parameters:
//   - errorType: The category of error (authentication, rate_limit, etc.)
//   - provider: The name of the AI provider that generated the error
//   - message: A human-readable description of the error
//
// Returns:
//   - *Error: A new standardized error instance
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

// IsRetryable returns true if the error type indicates a retryable condition.
//
// This method helps determine whether a failed request should be retried
// based on the error type. Rate limit and network errors are typically
// retryable, while authentication and validation errors are not.
//
// Retryable error types:
//   - ErrorTypeRateLimit: Should retry after the suggested delay
//   - ErrorTypeNetwork: Should retry with exponential backoff
//
// Non-retryable error types:
//   - ErrorTypeAuth: Requires fixing credentials
//   - ErrorTypeValidation: Requires fixing request parameters
//   - ErrorTypeProvider: May indicate service outage (context-dependent)
//   - ErrorTypeTokenLimit: Requires reducing request size
//
// Returns:
//   - bool: true if the error condition is typically retryable
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
