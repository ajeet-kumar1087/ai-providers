package main

import (
	"context"
	"fmt"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/adapters/openai"
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
	return c.adapter.Complete(ctx, req)
}

// ChatComplete implements the Client interface
func (c *client) ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return c.adapter.ChatComplete(ctx, req)
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
		// Will be implemented in task 6
		return nil, &Error{
			Type:     ErrorTypeProvider,
			Message:  "Anthropic adapter not yet implemented",
			Provider: string(provider),
		}
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
