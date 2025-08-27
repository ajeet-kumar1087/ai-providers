// Package types provides common types and utilities used across the AI provider wrapper.
//
// This package defines the core data structures for requests, responses, and
// configuration that are shared between the main client and provider-specific
// adapters. It also includes validation and configuration loading utilities.
package types

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// CompletionRequest represents a text completion request to an AI provider.
//
// This struct contains all the parameters needed to generate text completions,
// with automatic validation and provider-specific parameter mapping handled
// by the client implementation.
type CompletionRequest struct {
	// Prompt is the input text to generate a completion for (required)
	Prompt string `json:"prompt" validate:"required"`

	// Temperature controls randomness in the output (optional, 0.0-2.0)
	// Lower values make output more focused and deterministic
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`

	// MaxTokens limits the number of tokens in the generated completion (optional)
	// If not specified, the provider's default limit will be used
	MaxTokens *int `json:"max_tokens,omitempty" validate:"omitempty,min=1"`

	// Stop contains sequences where the API will stop generating further tokens (optional)
	// Maximum number of stop sequences varies by provider
	Stop []string `json:"stop,omitempty"`

	// Stream indicates whether to stream the response (optional, not yet implemented)
	// When true, the response will be streamed as it's generated
	Stream bool `json:"stream,omitempty"`
}

// CompletionResponse represents a text completion response from an AI provider.
//
// This struct contains the generated text along with metadata about the
// generation process, normalized across different providers.
type CompletionResponse struct {
	// Text contains the generated completion text
	Text string `json:"text"`

	// Usage provides token usage statistics for the request
	Usage Usage `json:"usage"`

	// FinishReason indicates why the generation stopped
	// Common values: "stop", "length", "content_filter"
	FinishReason string `json:"finish_reason"`
}

// ChatRequest represents a chat completion request with conversation history.
//
// This struct contains the conversation messages and parameters for generating
// a chat response, supporting multi-turn conversations with proper context management.
type ChatRequest struct {
	// Messages contains the conversation history (required, at least 1 message)
	// Should include user messages and any previous assistant responses
	Messages []Message `json:"messages" validate:"required,min=1"`

	// Temperature controls randomness in the output (optional, 0.0-2.0)
	// Lower values make output more focused and deterministic
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`

	// MaxTokens limits the number of tokens in the generated response (optional)
	// If not specified, the provider's default limit will be used
	MaxTokens *int `json:"max_tokens,omitempty" validate:"omitempty,min=1"`

	// Stream indicates whether to stream the response (optional, not yet implemented)
	// When true, the response will be streamed as it's generated
	Stream bool `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response from an AI provider.
//
// This struct contains the assistant's response message along with metadata
// about the generation process, normalized across different providers.
type ChatResponse struct {
	// Message contains the assistant's response message
	Message Message `json:"message"`

	// Usage provides token usage statistics for the request
	Usage Usage `json:"usage"`

	// FinishReason indicates why the generation stopped
	// Common values: "stop", "length", "content_filter"
	FinishReason string `json:"finish_reason"`
}

// Message represents a single message in a conversation.
//
// Messages form the building blocks of chat conversations, with different
// roles serving different purposes in the conversation flow.
type Message struct {
	// Role identifies the speaker of the message (required)
	// Valid values: "user", "assistant", "system"
	//   - "user": Messages from the human user
	//   - "assistant": Messages from the AI assistant
	//   - "system": System instructions or context (usually at the beginning)
	Role string `json:"role" validate:"required,oneof=user assistant system"`

	// Content contains the actual message text (required)
	Content string `json:"content" validate:"required"`
}

// Usage represents token usage information for API requests.
//
// This struct provides detailed information about token consumption,
// which is important for cost tracking and rate limit management.
type Usage struct {
	// PromptTokens is the number of tokens in the input prompt/messages
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the generated response
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the sum of prompt and completion tokens
	// This represents the total billable tokens for the request
	TotalTokens int `json:"total_tokens"`
}

// ProviderType represents the type of AI provider.
//
// This type is used to identify which AI provider to use when creating
// clients, allowing the system to select the appropriate adapter and
// apply provider-specific configurations.
type ProviderType string

const (
	// ProviderOpenAI represents the OpenAI provider.
	// Supports GPT-3.5, GPT-4, and other OpenAI models.
	// API Key format: starts with "sk-"
	ProviderOpenAI ProviderType = "openai"

	// ProviderAnthropic represents the Anthropic provider.
	// Supports Claude models (Claude-3, Claude-2, etc.).
	// API Key format: starts with "sk-ant-"
	ProviderAnthropic ProviderType = "anthropic"

	// ProviderGoogle represents the Google AI provider.
	// Supports Gemini models and other Google AI services.
	// API Key format: Google API key (typically 39 characters)
	ProviderGoogle ProviderType = "google"
)

// Config represents the configuration for an AI provider client.
//
// This struct contains all the settings needed to create and configure
// a client for any supported AI provider, with automatic validation
// and provider-specific defaults applied as needed.
type Config struct {
	// APIKey is the authentication key for the AI provider (required)
	// Format varies by provider (see ProviderType constants for details)
	APIKey string `json:"api_key" validate:"required"`

	// BaseURL allows overriding the default API endpoint (optional)
	// Useful for custom deployments or proxy configurations
	BaseURL string `json:"base_url,omitempty"`

	// Timeout sets the maximum duration for API requests (optional)
	// Default: 30 seconds if not specified
	Timeout time.Duration `json:"timeout,omitempty"`

	// MaxRetries sets the maximum number of retry attempts (optional)
	// Default: 3 retries if not specified
	MaxRetries int `json:"max_retries,omitempty"`

	// Temperature sets the default temperature for requests (optional, 0.0-2.0)
	// Can be overridden on individual requests
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`

	// MaxTokens sets the default maximum tokens for requests (optional)
	// Can be overridden on individual requests
	MaxTokens *int `json:"max_tokens,omitempty" validate:"omitempty,min=1"`
}

// DefaultConfig returns a configuration with sensible defaults.
//
// This function provides a starting point for configuration with reasonable
// default values. The API key must still be set before the configuration
// can be used to create a client.
//
// Default values:
//   - Timeout: 30 seconds
//   - MaxRetries: 3
//   - Temperature: not set (uses provider default)
//   - MaxTokens: not set (uses provider default)
//
// Example:
//
//	config := DefaultConfig()
//	config.APIKey = "your-api-key"
//	client, err := NewClient(ProviderOpenAI, config)
//
// Returns:
//   - Config: A configuration struct with default values
func DefaultConfig() Config {
	return Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// LoadConfigFromEnv loads configuration from environment variables.
//
// This function creates a configuration by reading from environment variables,
// starting with default values and overriding them with any environment
// variables that are set. This provides a convenient way to configure
// clients in containerized or cloud environments.
//
// Environment Variables by Provider:
//   - OpenAI: OPENAI_API_KEY, OPENAI_BASE_URL
//   - Anthropic: ANTHROPIC_API_KEY, ANTHROPIC_BASE_URL
//   - Google: GOOGLE_API_KEY, GOOGLE_BASE_URL
//
// Common Environment Variables:
//   - AI_TIMEOUT: Request timeout (e.g., "30s", "1m")
//   - AI_MAX_RETRIES: Maximum retry attempts (integer)
//   - AI_TEMPERATURE: Default temperature (float, 0.0-2.0)
//   - AI_MAX_TOKENS: Default max tokens (integer)
//
// Example:
//
//	// Set environment: export OPENAI_API_KEY="sk-your-key"
//	config := LoadConfigFromEnv(ProviderOpenAI)
//	client, err := NewClient(ProviderOpenAI, config)
//
// Parameters:
//   - provider: The provider type to load configuration for
//
// Returns:
//   - Config: A configuration loaded from environment variables
func LoadConfigFromEnv(provider ProviderType) Config {
	config := DefaultConfig()

	// Load API key based on provider
	switch provider {
	case ProviderOpenAI:
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			config.APIKey = key
		}
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}
	case ProviderAnthropic:
		if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
			config.APIKey = key
		}
		if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}
	case ProviderGoogle:
		if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
			config.APIKey = key
		}
		if baseURL := os.Getenv("GOOGLE_BASE_URL"); baseURL != "" {
			config.BaseURL = baseURL
		}
	}

	// Load common configuration from environment
	if timeout := os.Getenv("AI_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Timeout = duration
		}
	}

	if retries := os.Getenv("AI_MAX_RETRIES"); retries != "" {
		if maxRetries, err := strconv.Atoi(retries); err == nil && maxRetries >= 0 {
			config.MaxRetries = maxRetries
		}
	}

	if temp := os.Getenv("AI_TEMPERATURE"); temp != "" {
		if temperature, err := strconv.ParseFloat(temp, 64); err == nil {
			config.Temperature = &temperature
		}
	}

	if tokens := os.Getenv("AI_MAX_TOKENS"); tokens != "" {
		if maxTokens, err := strconv.Atoi(tokens); err == nil && maxTokens > 0 {
			config.MaxTokens = &maxTokens
		}
	}

	return config
}

// Validate validates the configuration for the specified provider.
//
// This method performs comprehensive validation of all configuration fields,
// including provider-specific validation such as API key format checking
// and parameter range validation. It should be called before using a
// configuration to create a client.
//
// Validation includes:
//   - Required field validation (API key)
//   - Provider type validation
//   - API key format validation (provider-specific)
//   - Parameter range validation (temperature, max tokens, etc.)
//   - Provider-specific limits and constraints
//
// Example:
//
//	config := Config{APIKey: "sk-invalid"}
//	if err := config.Validate(ProviderOpenAI); err != nil {
//		log.Fatal("Invalid configuration:", err)
//	}
//
// Parameters:
//   - provider: The provider type to validate the configuration against
//
// Returns:
//   - error: A validation error if the configuration is invalid, nil otherwise
func (c Config) Validate(provider ProviderType) error {
	// Validate required fields
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate provider type
	if err := ValidateProviderType(provider); err != nil {
		return err
	}

	// Validate API key format based on provider
	if err := c.validateAPIKeyFormat(provider); err != nil {
		return fmt.Errorf("invalid API key format: %w", err)
	}

	// Validate timeout
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative, got: %v", c.Timeout)
	}
	if c.Timeout == 0 {
		// Set default timeout if not specified
		c.Timeout = 30 * time.Second
	}

	// Validate max retries
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative, got: %d", c.MaxRetries)
	}

	// Validate temperature
	if c.Temperature != nil {
		temp := *c.Temperature
		if temp < 0.0 || temp > 2.0 {
			return fmt.Errorf("temperature must be between 0.0 and 2.0, got: %f", temp)
		}
	}

	// Validate max tokens
	if c.MaxTokens != nil {
		tokens := *c.MaxTokens
		if tokens <= 0 {
			return fmt.Errorf("max tokens must be positive, got: %d", tokens)
		}

		// Provider-specific token limits
		maxLimit := c.getProviderTokenLimit(provider)
		if tokens > maxLimit {
			return fmt.Errorf("max tokens exceeds provider limit of %d, got: %d", maxLimit, tokens)
		}
	}

	return nil
}

// validateAPIKeyFormat validates the API key format for the specific provider
func (c Config) validateAPIKeyFormat(provider ProviderType) error {
	key := strings.TrimSpace(c.APIKey)

	switch provider {
	case ProviderOpenAI:
		// OpenAI API keys typically start with "sk-"
		if !strings.HasPrefix(key, "sk-") {
			return fmt.Errorf("OpenAI API key should start with 'sk-'")
		}
		if len(key) < 20 {
			return fmt.Errorf("OpenAI API key appears to be too short")
		}
	case ProviderAnthropic:
		// Anthropic API keys typically start with "sk-ant-"
		if !strings.HasPrefix(key, "sk-ant-") {
			return fmt.Errorf("Anthropic API key should start with 'sk-ant-'")
		}
		if len(key) < 20 {
			return fmt.Errorf("Anthropic API key appears to be too short")
		}
	case ProviderGoogle:
		// Google API keys are typically 39 characters long
		if len(key) < 20 {
			return fmt.Errorf("Google API key appears to be too short")
		}
	}

	return nil
}

// getProviderTokenLimit returns the maximum token limit for the provider
func (c Config) getProviderTokenLimit(provider ProviderType) int {
	switch provider {
	case ProviderOpenAI:
		return 4096 // Conservative limit for GPT-3.5/4
	case ProviderAnthropic:
		return 100000 // Claude models support up to 100k tokens
	case ProviderGoogle:
		return 8192 // Conservative limit for Gemini
	default:
		return 4096 // Default conservative limit
	}
}

// WithAPIKey returns a new config with the specified API key.
//
// This method provides a fluent interface for setting the API key on a
// configuration. It returns a new Config instance rather than modifying
// the existing one, following immutable design patterns.
//
// Example:
//
//	config := DefaultConfig().WithAPIKey("sk-your-key")
//
// Parameters:
//   - apiKey: The API key to set in the configuration
//
// Returns:
//   - Config: A new configuration with the specified API key
func (c Config) WithAPIKey(apiKey string) Config {
	c.APIKey = apiKey
	return c
}

// WithBaseURL returns a new config with the specified base URL.
//
// This method allows overriding the default API endpoint, useful for
// custom deployments, proxy configurations, or testing environments.
//
// Example:
//
//	config := DefaultConfig().
//		WithAPIKey("sk-your-key").
//		WithBaseURL("https://api.custom-deployment.com")
//
// Parameters:
//   - baseURL: The base URL to use for API requests
//
// Returns:
//   - Config: A new configuration with the specified base URL
func (c Config) WithBaseURL(baseURL string) Config {
	c.BaseURL = baseURL
	return c
}

// WithTimeout returns a new config with the specified timeout.
//
// This method sets the maximum duration for API requests. Requests that
// take longer than this timeout will be cancelled and return an error.
//
// Example:
//
//	config := DefaultConfig().
//		WithAPIKey("sk-your-key").
//		WithTimeout(60 * time.Second)
//
// Parameters:
//   - timeout: The timeout duration for API requests
//
// Returns:
//   - Config: A new configuration with the specified timeout
func (c Config) WithTimeout(timeout time.Duration) Config {
	c.Timeout = timeout
	return c
}

// WithMaxRetries returns a new config with the specified max retries.
//
// This method sets the maximum number of retry attempts for failed requests.
// Only retryable errors (network issues, rate limits) will be retried.
//
// Example:
//
//	config := DefaultConfig().
//		WithAPIKey("sk-your-key").
//		WithMaxRetries(5)
//
// Parameters:
//   - maxRetries: The maximum number of retry attempts (must be >= 0)
//
// Returns:
//   - Config: A new configuration with the specified max retries
func (c Config) WithMaxRetries(maxRetries int) Config {
	c.MaxRetries = maxRetries
	return c
}

// WithTemperature returns a new config with the specified temperature.
//
// This method sets the default temperature for all requests made with this
// configuration. Individual requests can still override this value.
// Temperature controls the randomness of the output (0.0 = deterministic, 2.0 = very random).
//
// Example:
//
//	config := DefaultConfig().
//		WithAPIKey("sk-your-key").
//		WithTemperature(0.7)
//
// Parameters:
//   - temperature: The default temperature (must be between 0.0 and 2.0)
//
// Returns:
//   - Config: A new configuration with the specified temperature
func (c Config) WithTemperature(temperature float64) Config {
	c.Temperature = &temperature
	return c
}

// WithMaxTokens returns a new config with the specified max tokens.
//
// This method sets the default maximum number of tokens for all requests
// made with this configuration. Individual requests can still override
// this value. The actual limit depends on the provider and model used.
//
// Example:
//
//	config := DefaultConfig().
//		WithAPIKey("sk-your-key").
//		WithMaxTokens(1000)
//
// Parameters:
//   - maxTokens: The default maximum tokens (must be > 0)
//
// Returns:
//   - Config: A new configuration with the specified max tokens
func (c Config) WithMaxTokens(maxTokens int) Config {
	c.MaxTokens = &maxTokens
	return c
}

// ValidateProviderType validates that the provider type is supported.
//
// This function checks if the given provider type is one of the supported
// providers. It's used internally by other validation functions and can
// be used by applications that need to validate provider types independently.
//
// Supported providers:
//   - ProviderOpenAI
//   - ProviderAnthropic
//   - ProviderGoogle
//
// Example:
//
//	if err := ValidateProviderType(ProviderOpenAI); err != nil {
//		log.Fatal("Unsupported provider:", err)
//	}
//
// Parameters:
//   - provider: The provider type to validate
//
// Returns:
//   - error: An error if the provider is unsupported, nil if valid
func ValidateProviderType(provider ProviderType) error {
	switch provider {
	case ProviderOpenAI, ProviderAnthropic, ProviderGoogle:
		return nil
	default:
		return fmt.Errorf("unsupported provider '%s', supported providers: %v", provider, []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderGoogle})
	}
}
