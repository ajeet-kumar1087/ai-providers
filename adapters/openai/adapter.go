// Package openai provides OpenAI API adapter implementation
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	httputil "github.com/ajeet-kumar1087/ai-providers/internal/http"
	"github.com/ajeet-kumar1087/ai-providers/types"
)

const (
	// DefaultBaseURL is the default OpenAI API base URL
	DefaultBaseURL = "https://api.openai.com/v1"

	// DefaultModel is the default model to use for completions
	DefaultModel = "gpt-3.5-turbo-instruct"

	// DefaultChatModel is the default model to use for chat completions
	DefaultChatModel = "gpt-3.5-turbo"

	// MaxTokenLimit is the maximum number of tokens supported
	MaxTokenLimit = 4096
)

// AdapterConfig represents the configuration needed for OpenAI adapter
type AdapterConfig = types.Config

// OpenAIAdapter implements the ProviderAdapter interface for OpenAI
type OpenAIAdapter struct {
	httpClient *httputil.Client
	config     AdapterConfig
	baseURL    string
	apiKey     string
}

// NewAdapter creates a new OpenAI adapter with the given configuration
func NewAdapter(config AdapterConfig) (*OpenAIAdapter, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid OpenAI configuration: %w", err)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	httpClient := httputil.NewClient(timeout, maxRetries)

	return &OpenAIAdapter{
		httpClient: httpClient,
		config:     config,
		baseURL:    baseURL,
		apiKey:     config.APIKey,
	}, nil
}

// validateConfig validates the OpenAI configuration
func validateConfig(config AdapterConfig) error {
	if strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate API key format
	apiKey := strings.TrimSpace(config.APIKey)
	if !strings.HasPrefix(apiKey, "sk-") {
		return fmt.Errorf("OpenAI API key should start with 'sk-'")
	}
	if len(apiKey) < 20 {
		return fmt.Errorf("OpenAI API key appears to be too short")
	}

	// Validate timeout
	if config.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	// Validate max retries
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	// Validate temperature
	if config.Temperature != nil {
		temp := *config.Temperature
		if temp < 0.0 || temp > 2.0 {
			return fmt.Errorf("temperature must be between 0.0 and 2.0, got: %f", temp)
		}
	}

	// Validate max tokens
	if config.MaxTokens != nil {
		tokens := *config.MaxTokens
		if tokens <= 0 {
			return fmt.Errorf("max tokens must be positive, got: %d", tokens)
		}
		if tokens > MaxTokenLimit {
			return fmt.Errorf("max tokens exceeds OpenAI limit of %d, got: %d", MaxTokenLimit, tokens)
		}
	}

	return nil
}

// Name returns the name of the provider
func (a *OpenAIAdapter) Name() string {
	return "openai"
}

// SupportedFeatures returns a list of features supported by OpenAI
func (a *OpenAIAdapter) SupportedFeatures() []string {
	return []string{
		"completion",
		"chat_completion",
		"streaming",
		"temperature",
		"max_tokens",
		"stop_sequences",
		"system_messages",
		"function_calling",
	}
}

// ValidateConfig validates the configuration for OpenAI adapter
func (a *OpenAIAdapter) ValidateConfig(config AdapterConfig) error {
	return validateConfig(config)
}

// makeRequest makes an HTTP request to the OpenAI API
func (a *OpenAIAdapter) makeRequest(ctx context.Context, endpoint string, requestBody interface{}) (*http.Response, error) {
	// Marshal request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Prepare headers
	headers := map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}

	// Make the request
	url := a.baseURL + endpoint
	resp, err := a.httpClient.Post(ctx, url, headers, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}

// parseErrorResponse parses an OpenAI error response
func (a *OpenAIAdapter) parseErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	// Parse OpenAI error format
	var openaiError struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &openaiError); err != nil {
		// If we can't parse the error, return a generic error with the status code
		return fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	message := openaiError.Error.Message
	if message == "" {
		message = "Unknown OpenAI error"
	}

	// Map to our standardized error types
	switch resp.StatusCode {
	case 401, 403:
		return &Error{
			Type:     "authentication",
			Message:  message,
			Code:     openaiError.Error.Code,
			Provider: "openai",
		}
	case 429:
		return &Error{
			Type:       "rate_limit",
			Message:    message,
			Code:       openaiError.Error.Code,
			Provider:   "openai",
			RetryAfter: getRetryAfter(resp.Header),
		}
	case 400:
		return &Error{
			Type:     "validation",
			Message:  message,
			Code:     openaiError.Error.Code,
			Provider: "openai",
		}
	default:
		return &Error{
			Type:     "provider",
			Message:  message,
			Code:     openaiError.Error.Code,
			Provider: "openai",
		}
	}
}

// getRetryAfter extracts retry-after information from response headers
func getRetryAfter(headers http.Header) *int {
	if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
		// Try to parse as seconds (simple implementation)
		// In a production system, you might want to handle different formats
		if seconds := parseRetryAfterSeconds(retryAfter); seconds > 0 {
			return &seconds
		}
	}

	// Default retry after 60 seconds for rate limits
	defaultRetry := 60
	return &defaultRetry
}

// parseRetryAfterSeconds parses retry-after header value as seconds
func parseRetryAfterSeconds(value string) int {
	// Simplified implementation - just return 60 seconds as default
	// In production, you'd want to properly parse the header value
	return 60
}

// Error represents a standardized error for OpenAI adapter
type Error struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	Provider   string `json:"provider"`
	RetryAfter *int   `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s (%s): %s", e.Provider, e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Type, e.Message)
}

// Type aliases for imported types
type CompletionRequest = types.CompletionRequest
type CompletionResponse = types.CompletionResponse
type ChatRequest = types.ChatRequest
type ChatResponse = types.ChatResponse
type Message = types.Message
type Usage = types.Usage

// OpenAI API request/response types

// OpenAICompletionRequest represents an OpenAI completion request
type OpenAICompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

// OpenAICompletionResponse represents an OpenAI completion response
type OpenAICompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIChatCompletionRequest represents an OpenAI chat completion request
type OpenAIChatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// OpenAIChatCompletionResponse represents an OpenAI chat completion response
type OpenAIChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIMessage represents a chat message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Complete implements the ProviderAdapter interface for text completions
func (a *OpenAIAdapter) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Map generic request to OpenAI format
	openaiReq := a.mapCompletionRequest(req)

	// Make HTTP request to OpenAI API
	resp, err := a.makeRequest(ctx, "/completions", openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make completion request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		return nil, a.parseErrorResponse(resp)
	}

	// Parse successful response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var openaiResp OpenAICompletionResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Normalize response to generic format
	return a.normalizeCompletionResponse(openaiResp), nil
}

// mapCompletionRequest maps a generic CompletionRequest to OpenAI format
func (a *OpenAIAdapter) mapCompletionRequest(req CompletionRequest) OpenAICompletionRequest {
	openaiReq := OpenAICompletionRequest{
		Model:  DefaultModel,
		Prompt: req.Prompt,
		Stream: req.Stream,
	}

	// Apply temperature with range clamping
	if req.Temperature != nil {
		temp := *req.Temperature
		// Clamp to OpenAI's supported range (0.0-2.0)
		if temp < 0.0 {
			temp = 0.0
		}
		if temp > 2.0 {
			temp = 2.0
		}
		openaiReq.Temperature = &temp
	} else if a.config.Temperature != nil {
		// Use default from config if available
		temp := *a.config.Temperature
		if temp >= 0.0 && temp <= 2.0 {
			openaiReq.Temperature = &temp
		}
	}

	// Apply max tokens with provider-specific limits
	if req.MaxTokens != nil {
		tokens := *req.MaxTokens
		// Clamp to OpenAI's limit
		if tokens > MaxTokenLimit {
			tokens = MaxTokenLimit
		}
		if tokens > 0 {
			openaiReq.MaxTokens = &tokens
		}
	} else if a.config.MaxTokens != nil {
		// Use default from config if available
		tokens := *a.config.MaxTokens
		if tokens > 0 && tokens <= MaxTokenLimit {
			openaiReq.MaxTokens = &tokens
		}
	}

	// Apply stop sequences
	if len(req.Stop) > 0 {
		openaiReq.Stop = req.Stop
	}

	return openaiReq
}

// normalizeCompletionResponse converts OpenAI response to generic format
func (a *OpenAIAdapter) normalizeCompletionResponse(resp OpenAICompletionResponse) *CompletionResponse {
	// Extract text from first choice (OpenAI typically returns one choice for completions)
	text := ""
	finishReason := ""
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Text
		finishReason = resp.Choices[0].FinishReason
	}

	return &CompletionResponse{
		Text: text,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: finishReason,
	}
}

// ChatComplete implements the ProviderAdapter interface for chat completions
func (a *OpenAIAdapter) ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// This will be implemented in task 5.3
	return nil, fmt.Errorf("ChatComplete method not yet implemented")
}
