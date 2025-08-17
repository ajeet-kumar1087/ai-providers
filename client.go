package main

import (
	"context"
	"fmt"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/adapters/anthropic"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/adapters/openai"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/internal/utils"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// client is the main implementation of the Client interface
type client struct {
	adapter  ProviderAdapter
	provider ProviderType
	config   Config
}

// NewClient creates a new client instance for the specified provider
func NewClient(provider ProviderType, config Config) (Client, error) {
	// Validate provider type
	if !IsValidProvider(provider) {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("unsupported provider: %s", provider),
			Provider: string(provider),
		}
	}

	// Validate configuration before creating adapter
	if err := config.Validate(provider); err != nil {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("invalid configuration: %v", err),
			Provider: string(provider),
			Wrapped:  err,
		}
	}

	// Create adapter for the provider
	adapter, err := CreateAdapter(provider, config)
	if err != nil {
		return nil, err
	}

	// Additional adapter-specific validation
	if err := adapter.ValidateConfig(config); err != nil {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("adapter validation failed: %v", err),
			Provider: string(provider),
			Wrapped:  err,
		}
	}

	return &client{
		adapter:  adapter,
		provider: provider,
		config:   config,
	}, nil
}

// Complete implements the Client interface
func (c *client) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Validate and normalize the request before delegation
	normalizedReq, err := c.validateAndNormalizeCompletionRequest(req)
	if err != nil {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("request validation failed: %v", err),
			Provider: string(c.provider),
			Wrapped:  err,
		}
	}

	// Delegate to the provider adapter
	return c.adapter.Complete(ctx, normalizedReq)
}

// ChatComplete implements the Client interface
func (c *client) ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Validate and normalize the request before delegation
	normalizedReq, err := c.validateAndNormalizeChatRequest(req)
	if err != nil {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("request validation failed: %v", err),
			Provider: string(c.provider),
			Wrapped:  err,
		}
	}

	// Delegate to the provider adapter
	return c.adapter.ChatComplete(ctx, normalizedReq)
}

// Close implements the Client interface
func (c *client) Close() error {
	// Currently no cleanup needed, but this allows for future resource management
	return nil
}

// defaultClientFactory is the default implementation of ClientFactory
type defaultClientFactory struct{}

// CreateClient implements the ClientFactory interface
func (f *defaultClientFactory) CreateClient(provider ProviderType, config Config) (Client, error) {
	return NewClient(provider, config)
}

// SupportedProviders implements the ClientFactory interface
func (f *defaultClientFactory) SupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGoogle,
	}
}

// NewClientFactory creates a new client factory
func NewClientFactory() ClientFactory {
	return &defaultClientFactory{}
}

// Helper functions

// IsValidProvider checks if the provider type is supported
func IsValidProvider(provider ProviderType) bool {
	switch provider {
	case ProviderOpenAI, ProviderAnthropic, ProviderGoogle:
		return true
	default:
		return false
	}
}

// GetSupportedProviders returns a list of all supported provider types
func GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGoogle,
	}
}

// ValidateProviderType validates that the provider type is supported
var ValidateProviderType = types.ValidateProviderType

// CreateAdapter creates the appropriate adapter for the provider
func CreateAdapter(provider ProviderType, config Config) (ProviderAdapter, error) {
	// Validate provider type first
	if err := ValidateProviderType(provider); err != nil {
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  err.Error(),
			Provider: string(provider),
			Wrapped:  err,
		}
	}

	switch provider {
	case ProviderOpenAI:
		// Import the OpenAI adapter package
		openaiAdapter, err := createOpenAIAdapter(config)
		if err != nil {
			return nil, &Error{
				Type:     ErrorTypeProvider,
				Message:  fmt.Sprintf("failed to create OpenAI adapter: %v", err),
				Provider: string(provider),
				Wrapped:  err,
			}
		}
		return openaiAdapter, nil
	case ProviderAnthropic:
		// Import the Anthropic adapter package
		anthropicAdapter, err := createAnthropicAdapter(config)
		if err != nil {
			return nil, &Error{
				Type:     ErrorTypeProvider,
				Message:  fmt.Sprintf("failed to create Anthropic adapter: %v", err),
				Provider: string(provider),
				Wrapped:  err,
			}
		}
		return anthropicAdapter, nil
	case ProviderGoogle:
		// Will be implemented in future tasks
		return nil, &Error{
			Type:     ErrorTypeProvider,
			Message:  "Google adapter not yet implemented",
			Provider: string(provider),
		}
	default:
		// This should never happen due to validation above, but included for safety
		return nil, &Error{
			Type:     ErrorTypeValidation,
			Message:  fmt.Sprintf("unsupported provider: %s", provider),
			Provider: string(provider),
		}
	}
}

// createOpenAIAdapter creates an OpenAI adapter from the generic config
func createOpenAIAdapter(config Config) (ProviderAdapter, error) {
	// The config is already the correct type since AdapterConfig = types.Config
	return openai.NewAdapter(config)
}

// createAnthropicAdapter creates an Anthropic adapter from the generic config
func createAnthropicAdapter(config Config) (ProviderAdapter, error) {
	// The config is already the correct type since AdapterConfig = types.Config
	return anthropic.NewAdapter(config)
}

// Parameter validation and mapping functions

// validateAndNormalizeCompletionRequest validates and normalizes a completion request
func (c *client) validateAndNormalizeCompletionRequest(req CompletionRequest) (CompletionRequest, error) {
	// First, perform basic validation using utilities
	if err := utils.ValidateCompletionRequest(req); err != nil {
		return req, err
	}

	// Create a copy to avoid modifying the original request
	normalized := req

	// Apply parameter clamping for the target provider
	clamped := utils.ClampParameters(normalized, c.provider).(CompletionRequest)

	// Apply default values from config if not specified in request
	if clamped.Temperature == nil && c.config.Temperature != nil {
		// Apply default temperature from config, ensuring it's within provider limits
		temp := *c.config.Temperature
		maxTemp := utils.GetProviderMaxTemperature(c.provider)
		if temp >= 0.0 && temp <= maxTemp {
			clamped.Temperature = &temp
		}
	}

	if clamped.MaxTokens == nil && c.config.MaxTokens != nil {
		// Apply default max tokens from config, ensuring it's within provider limits
		tokens := *c.config.MaxTokens
		maxLimit := utils.GetProviderTokenLimit(c.provider)
		if tokens > 0 && tokens <= maxLimit {
			clamped.MaxTokens = &tokens
		}
	}

	return clamped, nil
}

// validateAndNormalizeChatRequest validates and normalizes a chat request
func (c *client) validateAndNormalizeChatRequest(req ChatRequest) (ChatRequest, error) {
	// First, perform basic validation using utilities
	if err := utils.ValidateChatRequest(req); err != nil {
		return req, err
	}

	// Validate conversation structure (provider-specific logic)
	if err := c.validateConversationStructure(req.Messages); err != nil {
		return req, fmt.Errorf("invalid conversation structure: %w", err)
	}

	// Create a copy to avoid modifying the original request
	normalized := req

	// Apply parameter clamping for the target provider
	clamped := utils.ClampParameters(normalized, c.provider).(ChatRequest)

	// Apply default values from config if not specified in request
	if clamped.Temperature == nil && c.config.Temperature != nil {
		// Apply default temperature from config, ensuring it's within provider limits
		temp := *c.config.Temperature
		maxTemp := utils.GetProviderMaxTemperature(c.provider)
		if temp >= 0.0 && temp <= maxTemp {
			clamped.Temperature = &temp
		}
	}

	if clamped.MaxTokens == nil && c.config.MaxTokens != nil {
		// Apply default max tokens from config, ensuring it's within provider limits
		tokens := *c.config.MaxTokens
		maxLimit := utils.GetProviderTokenLimit(c.provider)
		if tokens > 0 && tokens <= maxLimit {
			clamped.MaxTokens = &tokens
		}
	}

	return clamped, nil
}

// validateConversationStructure validates the structure of a conversation
func (c *client) validateConversationStructure(messages []Message) error {
	if len(messages) == 0 {
		return fmt.Errorf("conversation must have at least one message")
	}

	// Check for alternating user/assistant pattern (with optional system messages)
	var lastNonSystemRole string
	systemMessageCount := 0

	for i, msg := range messages {
		switch msg.Role {
		case "system":
			systemMessageCount++
			// System messages are typically at the beginning, but some providers allow them anywhere
			// We'll be permissive here and just count them
		case "user":
			if lastNonSystemRole == "user" {
				// Allow consecutive user messages (some use cases require this)
			}
			lastNonSystemRole = "user"
		case "assistant":
			if lastNonSystemRole == "" {
				return fmt.Errorf("conversation cannot start with assistant message at position %d", i)
			}
			lastNonSystemRole = "assistant"
		}
	}

	// Provider-specific validation
	switch c.provider {
	case ProviderAnthropic:
		// Anthropic prefers system messages to be separate, but we handle this in the adapter
		if systemMessageCount > 5 {
			return fmt.Errorf("too many system messages (%d), Anthropic recommends fewer system messages", systemMessageCount)
		}
	case ProviderOpenAI:
		// OpenAI is more flexible with message structure
	}

	return nil
}

// Provider-specific parameter limits

// getMaxTemperature returns the maximum temperature for the current provider
func (c *client) getMaxTemperature() float64 {
	switch c.provider {
	case ProviderOpenAI:
		return 2.0
	case ProviderAnthropic:
		return 1.0
	case ProviderGoogle:
		return 1.0
	default:
		return 1.0 // Conservative default
	}
}

// getMaxTokenLimit returns the maximum token limit for the current provider
func (c *client) getMaxTokenLimit() int {
	switch c.provider {
	case ProviderOpenAI:
		return 4096 // Conservative limit for GPT-3.5/4
	case ProviderAnthropic:
		return 100000 // Claude models support up to 100k tokens
	case ProviderGoogle:
		return 8192 // Conservative limit for Gemini
	default:
		return 4096 // Conservative default
	}
}

// getMaxStopSequences returns the maximum number of stop sequences for the current provider
func (c *client) getMaxStopSequences() int {
	switch c.provider {
	case ProviderOpenAI:
		return 4 // OpenAI supports up to 4 stop sequences
	case ProviderAnthropic:
		return 10 // Anthropic supports more stop sequences
	case ProviderGoogle:
		return 5 // Conservative limit for Google
	default:
		return 4 // Conservative default
	}
}
