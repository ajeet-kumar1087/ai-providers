// Package anthropic provides Anthropic API adapter implementation
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	httputil "github.com/ai-provider-wrapper/ai-provider-wrapper/internal/http"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

const (
	// DefaultBaseURL is the default Anthropic API base URL
	DefaultBaseURL = "https://api.anthropic.com/v1"

	// DefaultModel is the default model to use for completions
	DefaultModel = "claude-3-haiku-20240307"

	// DefaultChatModel is the default model to use for chat completions
	DefaultChatModel = "claude-3-haiku-20240307"

	// MaxTokenLimit is the maximum number of tokens supported
	MaxTokenLimit = 100000

	// APIVersion is the Anthropic API version to use
	APIVersion = "2023-06-01"
)

// AdapterConfig represents the configuration needed for Anthropic adapter
type AdapterConfig = types.Config

// AnthropicAdapter implements the ProviderAdapter interface for Anthropic
type AnthropicAdapter struct {
	httpClient *httputil.Client
	config     AdapterConfig
	baseURL    string
	apiKey     string
}

// NewAdapter creates a new Anthropic adapter with the given configuration
func NewAdapter(config AdapterConfig) (*AnthropicAdapter, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid Anthropic configuration: %w", err)
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

	return &AnthropicAdapter{
		httpClient: httpClient,
		config:     config,
		baseURL:    baseURL,
		apiKey:     config.APIKey,
	}, nil
}

// validateConfig validates the Anthropic configuration
func validateConfig(config AdapterConfig) error {
	if strings.TrimSpace(config.APIKey) == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate API key format
	apiKey := strings.TrimSpace(config.APIKey)
	if !strings.HasPrefix(apiKey, "sk-ant-") {
		return fmt.Errorf("Anthropic API key should start with 'sk-ant-'")
	}
	if len(apiKey) < 20 {
		return fmt.Errorf("Anthropic API key appears to be too short")
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
		if temp < 0.0 || temp > 1.0 {
			return fmt.Errorf("temperature must be between 0.0 and 1.0 for Anthropic, got: %f", temp)
		}
	}

	// Validate max tokens
	if config.MaxTokens != nil {
		tokens := *config.MaxTokens
		if tokens <= 0 {
			return fmt.Errorf("max tokens must be positive, got: %d", tokens)
		}
		if tokens > MaxTokenLimit {
			return fmt.Errorf("max tokens exceeds Anthropic limit of %d, got: %d", MaxTokenLimit, tokens)
		}
	}

	return nil
}

// Name returns the name of the provider
func (a *AnthropicAdapter) Name() string {
	return "anthropic"
}

// SupportedFeatures returns a list of features supported by Anthropic
func (a *AnthropicAdapter) SupportedFeatures() []string {
	return []string{
		"completion",
		"chat_completion",
		"streaming",
		"temperature",
		"max_tokens",
		"stop_sequences",
		"system_messages",
	}
}

// ValidateConfig validates the configuration for Anthropic adapter
func (a *AnthropicAdapter) ValidateConfig(config AdapterConfig) error {
	return validateConfig(config)
}

// makeRequest makes an HTTP request to the Anthropic API
func (a *AnthropicAdapter) makeRequest(ctx context.Context, endpoint string, requestBody interface{}) (*http.Response, error) {
	// Marshal request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Prepare headers
	headers := map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": APIVersion,
		"Content-Type":      "application/json",
	}

	// Make the request
	url := a.baseURL + endpoint
	resp, err := a.httpClient.Post(ctx, url, headers, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	return resp, nil
}

// parseErrorResponse parses an Anthropic error response
func (a *AnthropicAdapter) parseErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	// Parse Anthropic error format
	var anthropicError struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &anthropicError); err != nil {
		// If we can't parse the error, return a generic error with the status code
		return fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	message := anthropicError.Message
	if message == "" {
		message = "Unknown Anthropic error"
	}

	// Map to our standardized error types
	switch resp.StatusCode {
	case 401, 403:
		return &Error{
			Type:     "authentication",
			Message:  message,
			Code:     anthropicError.Type,
			Provider: "anthropic",
		}
	case 429:
		return &Error{
			Type:       "rate_limit",
			Message:    message,
			Code:       anthropicError.Type,
			Provider:   "anthropic",
			RetryAfter: getRetryAfter(resp.Header),
		}
	case 400:
		return &Error{
			Type:     "validation",
			Message:  message,
			Code:     anthropicError.Type,
			Provider: "anthropic",
		}
	default:
		return &Error{
			Type:     "provider",
			Message:  message,
			Code:     anthropicError.Type,
			Provider: "anthropic",
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

// Error represents a standardized error for Anthropic adapter
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

// Anthropic API request/response types

// AnthropicCompletionRequest represents an Anthropic completion request
type AnthropicCompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature *float64 `json:"temperature,omitempty"`
	StopSeq     []string `json:"stop_sequences,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

// AnthropicCompletionResponse represents an Anthropic completion response
type AnthropicCompletionResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// AnthropicChatCompletionRequest represents an Anthropic chat completion request
type AnthropicChatCompletionRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []AnthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	StopSeq     []string           `json:"stop_sequences,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// AnthropicChatCompletionResponse represents an Anthropic chat completion response
type AnthropicChatCompletionResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// AnthropicMessage represents a chat message in Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Complete implements the ProviderAdapter interface for text completions
func (a *AnthropicAdapter) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Map generic request to Anthropic format
	anthropicReq := a.mapCompletionRequest(req)

	// Make HTTP request to Anthropic API
	resp, err := a.makeRequest(ctx, "/messages", anthropicReq)
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

	var anthropicResp AnthropicChatCompletionResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	// Normalize response to generic format
	return a.normalizeCompletionResponse(anthropicResp), nil
}

// mapCompletionRequest maps a generic CompletionRequest to Anthropic format
func (a *AnthropicAdapter) mapCompletionRequest(req CompletionRequest) AnthropicChatCompletionRequest {
	// Anthropic uses the messages API for both completion and chat
	// Convert prompt to a user message
	messages := []AnthropicMessage{
		{
			Role:    "user",
			Content: req.Prompt,
		},
	}

	anthropicReq := AnthropicChatCompletionRequest{
		Model:    DefaultModel,
		Messages: messages,
		Stream:   req.Stream,
	}

	// Set max tokens (required for Anthropic)
	maxTokens := 1024 // Default value
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
		// Clamp to Anthropic's limit
		if maxTokens > MaxTokenLimit {
			maxTokens = MaxTokenLimit
		}
	} else if a.config.MaxTokens != nil {
		// Use default from config if available
		tokens := *a.config.MaxTokens
		if tokens > 0 && tokens <= MaxTokenLimit {
			maxTokens = tokens
		}
	}
	anthropicReq.MaxTokens = maxTokens

	// Apply temperature with range clamping
	if req.Temperature != nil {
		temp := *req.Temperature
		// Clamp to Anthropic's supported range (0.0-1.0)
		if temp < 0.0 {
			temp = 0.0
		}
		if temp > 1.0 {
			temp = 1.0
		}
		anthropicReq.Temperature = &temp
	} else if a.config.Temperature != nil {
		// Use default from config if available
		temp := *a.config.Temperature
		if temp >= 0.0 && temp <= 1.0 {
			anthropicReq.Temperature = &temp
		}
	}

	// Apply stop sequences
	if len(req.Stop) > 0 {
		anthropicReq.StopSeq = req.Stop
	}

	return anthropicReq
}

// normalizeCompletionResponse converts Anthropic response to generic format
func (a *AnthropicAdapter) normalizeCompletionResponse(resp AnthropicChatCompletionResponse) *CompletionResponse {
	// Extract text from content array
	text := ""
	if len(resp.Content) > 0 && resp.Content[0].Type == "text" {
		text = resp.Content[0].Text
	}

	return &CompletionResponse{
		Text: text,
		Usage: Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		FinishReason: resp.StopReason,
	}
}

// ChatComplete implements the ProviderAdapter interface for chat completions
func (a *AnthropicAdapter) ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Map generic request to Anthropic format
	anthropicReq := a.mapChatRequest(req)

	// Make HTTP request to Anthropic API
	resp, err := a.makeRequest(ctx, "/messages", anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make chat completion request: %w", err)
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

	var anthropicResp AnthropicChatCompletionResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	// Normalize response to generic format
	return a.normalizeChatResponse(anthropicResp), nil
}

// mapChatRequest maps a generic ChatRequest to Anthropic format
func (a *AnthropicAdapter) mapChatRequest(req ChatRequest) AnthropicChatCompletionRequest {
	anthropicReq := AnthropicChatCompletionRequest{
		Model:  DefaultChatModel,
		Stream: req.Stream,
	}

	// Set max tokens (required for Anthropic)
	maxTokens := 1024 // Default value
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
		// Clamp to Anthropic's limit
		if maxTokens > MaxTokenLimit {
			maxTokens = MaxTokenLimit
		}
	} else if a.config.MaxTokens != nil {
		// Use default from config if available
		tokens := *a.config.MaxTokens
		if tokens > 0 && tokens <= MaxTokenLimit {
			maxTokens = tokens
		}
	}
	anthropicReq.MaxTokens = maxTokens

	// Apply temperature with range clamping
	if req.Temperature != nil {
		temp := *req.Temperature
		// Clamp to Anthropic's supported range (0.0-1.0)
		if temp < 0.0 {
			temp = 0.0
		}
		if temp > 1.0 {
			temp = 1.0
		}
		anthropicReq.Temperature = &temp
	} else if a.config.Temperature != nil {
		// Use default from config if available
		temp := *a.config.Temperature
		if temp >= 0.0 && temp <= 1.0 {
			anthropicReq.Temperature = &temp
		}
	}

	// Convert messages and handle system messages
	var systemMessage string
	var messages []AnthropicMessage

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			// Anthropic handles system messages separately
			if systemMessage == "" {
				systemMessage = msg.Content
			} else {
				// If multiple system messages, concatenate them
				systemMessage += "\n\n" + msg.Content
			}
		case "user", "assistant":
			messages = append(messages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		default:
			// For unsupported roles, convert to user message with role prefix
			messages = append(messages, AnthropicMessage{
				Role:    "user",
				Content: fmt.Sprintf("[%s]: %s", msg.Role, msg.Content),
			})
		}
	}

	anthropicReq.Messages = messages
	if systemMessage != "" {
		anthropicReq.System = systemMessage
	}

	return anthropicReq
}

// normalizeChatResponse converts Anthropic response to generic format
func (a *AnthropicAdapter) normalizeChatResponse(resp AnthropicChatCompletionResponse) *ChatResponse {
	// Extract text from content array
	text := ""
	if len(resp.Content) > 0 && resp.Content[0].Type == "text" {
		text = resp.Content[0].Text
	}

	return &ChatResponse{
		Message: Message{
			Role:    "assistant",
			Content: text,
		},
		Usage: Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		FinishReason: resp.StopReason,
	}
}
