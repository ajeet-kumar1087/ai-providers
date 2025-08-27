package aiprovider

import (
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// Re-export functions from the types package for backward compatibility.
// These provide convenient access to configuration utilities without
// requiring direct imports of the types package.
var (
	// DefaultConfig returns a configuration with sensible defaults.
	// Equivalent to types.DefaultConfig().
	DefaultConfig = types.DefaultConfig

	// LoadConfigFromEnv loads configuration from environment variables.
	// Equivalent to types.LoadConfigFromEnv().
	LoadConfigFromEnv = types.LoadConfigFromEnv
)

// NewClientWithEnvConfig creates a new client using configuration loaded from environment variables.
//
// This is a convenience function that combines LoadConfigFromEnv and NewClient
// into a single call. It automatically loads the appropriate environment
// variables based on the provider type.
//
// Environment Variables by Provider:
//   - OpenAI: OPENAI_API_KEY, OPENAI_BASE_URL
//   - Anthropic: ANTHROPIC_API_KEY, ANTHROPIC_BASE_URL
//   - Google: GOOGLE_API_KEY, GOOGLE_BASE_URL
//   - Common: AI_TIMEOUT, AI_MAX_RETRIES, AI_TEMPERATURE, AI_MAX_TOKENS
//
// Example:
//
//	// Set environment variable: export OPENAI_API_KEY="sk-your-key"
//	client, err := NewClientWithEnvConfig(ProviderOpenAI)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// Parameters:
//   - provider: The AI provider to create a client for
//
// Returns:
//   - Client: A configured client instance using environment configuration
//   - error: An error if required environment variables are missing or invalid
func NewClientWithEnvConfig(provider ProviderType) (Client, error) {
	config := LoadConfigFromEnv(provider)
	return NewClient(provider, config)
}

// NewClientWithDefaults creates a new client with default configuration and the specified API key.
//
// This is a convenience function for quick client setup when you only need
// to specify the API key and want to use default values for all other
// configuration options (30s timeout, 3 retries, etc.).
//
// Example:
//
//	client, err := NewClientWithDefaults(ProviderOpenAI, "sk-your-openai-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// Parameters:
//   - provider: The AI provider to create a client for
//   - apiKey: The API key for the specified provider
//
// Returns:
//   - Client: A configured client instance with default settings
//   - error: An error if the provider is unsupported or API key is invalid
func NewClientWithDefaults(provider ProviderType, apiKey string) (Client, error) {
	config := DefaultConfig().WithAPIKey(apiKey)
	return NewClient(provider, config)
}
