// Package main provides a unified interface for interacting with multiple AI providers.
//
// This package abstracts away provider-specific implementations and offers a consistent
// API for developers to integrate AI capabilities into their applications without being
// locked into a single provider.
//
// Supported Providers:
//   - OpenAI (GPT models)
//   - Anthropic (Claude models)
//   - Google AI (Gemini models)
//
// Basic Usage:
//
//	// Create a client for OpenAI
//	client, err := NewClient(ProviderOpenAI, Config{
//		APIKey: "your-openai-api-key",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Send a completion request
//	resp, err := client.Complete(context.Background(), CompletionRequest{
//		Prompt: "Hello, world!",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(resp.Text)
//
// Configuration can be loaded from environment variables:
//
//	client, err := NewClientWithEnvConfig(ProviderOpenAI)
//
// The package provides consistent error handling across all providers with detailed
// error categorization for authentication, rate limiting, network issues, and more.
package main

import (
	"context"
	"fmt"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/adapters/anthropic"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/adapters/openai"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/internal/utils"
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// client is the main implementation of the Client interface.
// It delegates requests to provider-specific adapters while providing
// unified parameter validation and error handling.
type client struct {
	adapter  ProviderAdapter // The provider-specific adapter
	provider ProviderType    // The provider type for this client
	config   Config          // The configuration used to create this client
}

// NewClient creates a new client instance for the specified provider.
//
// The function validates the provider type and configuration before creating
// the appropriate adapter. It returns an error if the provider is unsupported
// or if the configuration is invalid.
//
// Example:
//
//	client, err := NewClient(ProviderOpenAI, Config{
//		APIKey: "sk-your-openai-key",
//		Timeout: 30 * time.Second,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// Parameters:
//   - provider: The AI provider to use (ProviderOpenAI, ProviderAnthropic, or ProviderGoogle)
//   - config: Configuration including API key, timeout, and optional parameters
//
// Returns:
//   - Client: A configured client instance ready for use
//   - error: An error if provider is unsupported or configuration is invalid
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

// Complete sends a text completion request to the configured AI provider.
//
// The method validates and normalizes the request parameters before delegating
// to the provider-specific adapter. Parameters are automatically clamped to
// provider-specific limits and default values from the client configuration
// are applied when not specified in the request.
//
// Example:
//
//	resp, err := client.Complete(ctx, CompletionRequest{
//		Prompt: "Write a haiku about programming",
//		Temperature: &[]float64{0.7}[0],
//		MaxTokens: &[]int{100}[0],
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(resp.Text)
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - req: The completion request with prompt and optional parameters
//
// Returns:
//   - *CompletionResponse: The completion response with generated text and usage info
//   - error: An error if the request fails or parameters are invalid
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

// ChatComplete sends a chat completion request to the configured AI provider.
//
// The method handles conversation history with proper role mapping and validates
// the conversation structure. Message roles are automatically mapped to
// provider-specific formats, and parameters are normalized across providers.
//
// Example:
//
//	resp, err := client.ChatComplete(ctx, ChatRequest{
//		Messages: []Message{
//			{Role: "user", Content: "Hello, how are you?"},
//		},
//		Temperature: &[]float64{0.7}[0],
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(resp.Message.Content)
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - req: The chat request with messages and optional parameters
//
// Returns:
//   - *ChatResponse: The chat response with the assistant's message and usage info
//   - error: An error if the request fails or conversation structure is invalid
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

// Close cleans up resources and closes the client.
//
// Currently, this method performs no cleanup as the client uses stateless
// HTTP connections. However, it's provided for future compatibility and
// should always be called when the client is no longer needed.
//
// Example:
//
//	client, err := NewClient(ProviderOpenAI, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close() // Always close the client
//
// Returns:
//   - error: Always returns nil in the current implementation
func (c *client) Close() error {
	// Currently no cleanup needed, but this allows for future resource management
	return nil
}

// defaultClientFactory is the default implementation of ClientFactory.
// It provides a factory interface for creating clients with different providers.
type defaultClientFactory struct{}

// CreateClient creates a new client for the specified provider and configuration.
//
// This method is equivalent to calling NewClient directly but provides a
// factory interface for dependency injection scenarios.
//
// Parameters:
//   - provider: The AI provider to use
//   - config: Configuration for the client
//
// Returns:
//   - Client: A configured client instance
//   - error: An error if client creation fails
func (f *defaultClientFactory) CreateClient(provider ProviderType, config Config) (Client, error) {
	return NewClient(provider, config)
}

// SupportedProviders returns a list of all supported provider types.
//
// Returns:
//   - []ProviderType: A slice containing all supported provider types
func (f *defaultClientFactory) SupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGoogle,
	}
}

// NewClientFactory creates a new client factory instance.
//
// The factory provides an alternative interface for creating clients,
// useful in dependency injection scenarios or when you need to query
// supported providers programmatically.
//
// Example:
//
//	factory := NewClientFactory()
//	providers := factory.SupportedProviders()
//	client, err := factory.CreateClient(providers[0], config)
//
// Returns:
//   - ClientFactory: A new client factory instance
func NewClientFactory() ClientFactory {
	return &defaultClientFactory{}
}

// Helper functions

// IsValidProvider checks if the provider type is supported.
//
// This function can be used to validate provider types before attempting
// to create a client, providing a quick way to check support without
// triggering the full client creation process.
//
// Example:
//
//	if IsValidProvider(ProviderOpenAI) {
//		client, err := NewClient(ProviderOpenAI, config)
//		// ...
//	}
//
// Parameters:
//   - provider: The provider type to validate
//
// Returns:
//   - bool: true if the provider is supported, false otherwise
func IsValidProvider(provider ProviderType) bool {
	switch provider {
	case ProviderOpenAI, ProviderAnthropic, ProviderGoogle:
		return true
	default:
		return false
	}
}

// GetSupportedProviders returns a list of all supported provider types.
//
// This function provides a convenient way to enumerate all available
// providers, useful for building configuration UIs or validation logic.
//
// Example:
//
//	providers := GetSupportedProviders()
//	for _, provider := range providers {
//		fmt.Printf("Supported provider: %s\n", provider)
//	}
//
// Returns:
//   - []ProviderType: A slice containing all supported provider types
func GetSupportedProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderGoogle,
	}
}

// ValidateProviderType validates that the provider type is supported
var ValidateProviderType = types.ValidateProviderType

// CreateAdapter creates the appropriate adapter for the specified provider.
//
// This function is used internally by NewClient to instantiate the correct
// provider-specific adapter. It validates the provider type and delegates
// to the appropriate adapter constructor.
//
// Parameters:
//   - provider: The provider type to create an adapter for
//   - config: Configuration to pass to the adapter
//
// Returns:
//   - ProviderAdapter: The created adapter instance
//   - error: An error if the provider is unsupported or adapter creation fails
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
