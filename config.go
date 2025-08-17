package main

import (
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// Re-export functions from the types package for backward compatibility
var DefaultConfig = types.DefaultConfig
var LoadConfigFromEnv = types.LoadConfigFromEnv

// NewClientWithEnvConfig creates a new client using configuration loaded from environment variables
func NewClientWithEnvConfig(provider ProviderType) (Client, error) {
	config := LoadConfigFromEnv(provider)
	return NewClient(provider, config)
}

// NewClientWithDefaults creates a new client with default configuration and the specified API key
func NewClientWithDefaults(provider ProviderType, apiKey string) (Client, error) {
	config := DefaultConfig().WithAPIKey(apiKey)
	return NewClient(provider, config)
}
