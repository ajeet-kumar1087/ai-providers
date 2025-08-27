package aiprovider

import "context"

// Client represents the main interface for interacting with AI providers.
//
// This interface provides a unified API for text and chat completions across
// different AI providers. All provider-specific details are abstracted away,
// allowing seamless switching between providers without code changes.
//
// The interface supports both simple text completions and conversational
// chat completions with proper context management and parameter normalization.
type Client interface {
	// Complete sends a text completion request to the AI provider.
	//
	// This method generates text based on a prompt, supporting various
	// parameters like temperature and token limits. The implementation
	// handles provider-specific parameter mapping and response normalization.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout control
	//   - req: The completion request containing prompt and optional parameters
	//
	// Returns:
	//   - *CompletionResponse: Generated text with usage statistics
	//   - error: Provider-specific error wrapped in standardized error type
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// ChatComplete sends a chat completion request with conversation history.
	//
	// This method handles multi-turn conversations with proper role management
	// (user, assistant, system). Message formats are automatically converted
	// to provider-specific formats while maintaining conversation context.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout control
	//   - req: The chat request containing messages and optional parameters
	//
	// Returns:
	//   - *ChatResponse: Assistant's response message with usage statistics
	//   - error: Provider-specific error wrapped in standardized error type
	ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Close cleans up resources and closes the client connection.
	//
	// This method should be called when the client is no longer needed
	// to ensure proper resource cleanup. It's safe to call multiple times.
	//
	// Returns:
	//   - error: An error if cleanup fails (currently always returns nil)
	Close() error
}

// ProviderAdapter represents the interface that each AI provider must implement.
//
// This interface defines the contract for provider-specific adapters, enabling
// the main client to delegate requests while maintaining a consistent API.
// Each adapter handles the specifics of communicating with its respective
// AI provider's API, including authentication, request formatting, and response parsing.
type ProviderAdapter interface {
	// Complete handles text completion requests for the specific provider.
	//
	// Implementations should convert the generic CompletionRequest to the
	// provider's specific format, make the API call, and normalize the
	// response back to the standard CompletionResponse format.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout
	//   - req: Generic completion request to be converted to provider format
	//
	// Returns:
	//   - *CompletionResponse: Normalized response from the provider
	//   - error: Standardized error with provider-specific details
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// ChatComplete handles chat completion requests for the specific provider.
	//
	// Implementations should handle conversation history, role mapping,
	// and any provider-specific chat features while maintaining compatibility
	// with the generic chat interface.
	//
	// Parameters:
	//   - ctx: Context for request cancellation and timeout
	//   - req: Generic chat request with conversation history
	//
	// Returns:
	//   - *ChatResponse: Normalized chat response from the provider
	//   - error: Standardized error with provider-specific details
	ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ValidateConfig validates the configuration for this specific provider.
	//
	// This method performs provider-specific validation beyond the generic
	// config validation, such as API key format verification or endpoint
	// accessibility checks.
	//
	// Parameters:
	//   - config: Configuration to validate for this provider
	//
	// Returns:
	//   - error: Validation error if configuration is invalid for this provider
	ValidateConfig(config Config) error

	// Name returns the human-readable name of the provider.
	//
	// This is used for logging, error messages, and debugging purposes.
	// Should return a consistent string identifier for the provider.
	//
	// Returns:
	//   - string: The provider's name (e.g., "OpenAI", "Anthropic")
	Name() string

	// SupportedFeatures returns a list of features supported by this provider.
	//
	// This allows clients to query provider capabilities and adapt behavior
	// accordingly. Features might include "streaming", "function_calling",
	// "image_input", etc.
	//
	// Returns:
	//   - []string: List of supported feature identifiers
	SupportedFeatures() []string
}

// ClientFactory represents the interface for creating AI provider clients.
//
// This interface provides a factory pattern for client creation, useful in
// dependency injection scenarios or when you need to programmatically
// discover and create clients for different providers.
type ClientFactory interface {
	// CreateClient creates a new client for the specified provider and configuration.
	//
	// This method provides the same functionality as NewClient but through
	// a factory interface, enabling dependency injection and testing scenarios.
	//
	// Parameters:
	//   - provider: The AI provider type to create a client for
	//   - config: Configuration including API keys and optional parameters
	//
	// Returns:
	//   - Client: A configured client instance ready for use
	//   - error: An error if provider is unsupported or configuration is invalid
	CreateClient(provider ProviderType, config Config) (Client, error)

	// SupportedProviders returns a list of all supported provider types.
	//
	// This method allows programmatic discovery of available providers,
	// useful for building configuration interfaces or validation logic.
	//
	// Returns:
	//   - []ProviderType: Slice of all supported provider type constants
	SupportedProviders() []ProviderType
}
