package main

import "context"

// Client represents the main interface for interacting with AI providers
type Client interface {
	// Complete sends a text completion request
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// ChatComplete sends a chat completion request
	ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Close cleans up resources and closes the client
	Close() error
}

// ProviderAdapter represents the interface that each provider must implement
type ProviderAdapter interface {
	// Complete handles text completion requests for the specific provider
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// ChatComplete handles chat completion requests for the specific provider
	ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ValidateConfig validates the configuration for this provider
	ValidateConfig(config Config) error

	// Name returns the name of the provider
	Name() string

	// SupportedFeatures returns a list of features supported by this provider
	SupportedFeatures() []string
}

// ClientFactory represents the interface for creating clients
type ClientFactory interface {
	// CreateClient creates a new client for the specified provider and configuration
	CreateClient(provider ProviderType, config Config) (Client, error)

	// SupportedProviders returns a list of supported provider types
	SupportedProviders() []ProviderType
}
