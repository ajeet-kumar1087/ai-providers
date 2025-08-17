// Package types provides common types used across the AI provider wrapper
package types

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// CompletionRequest represents a text completion request
type CompletionRequest struct {
	Prompt      string   `json:"prompt" validate:"required"`
	Temperature *float64 `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`
	MaxTokens   *int     `json:"max_tokens,omitempty" validate:"omitempty,min=1"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

// CompletionResponse represents a text completion response
type CompletionResponse struct {
	Text         string `json:"text"`
	Usage        Usage  `json:"usage"`
	FinishReason string `json:"finish_reason"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []Message `json:"messages" validate:"required,min=1"`
	Temperature *float64  `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`
	MaxTokens   *int      `json:"max_tokens,omitempty" validate:"omitempty,min=1"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Message      Message `json:"message"`
	Usage        Usage   `json:"usage"`
	FinishReason string  `json:"finish_reason"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role" validate:"required,oneof=user assistant system"`
	Content string `json:"content" validate:"required"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderGoogle    ProviderType = "google"
)

// Config represents the configuration for an AI provider client
type Config struct {
	APIKey      string        `json:"api_key" validate:"required"`
	BaseURL     string        `json:"base_url,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
	MaxRetries  int           `json:"max_retries,omitempty"`
	Temperature *float64      `json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`
	MaxTokens   *int          `json:"max_tokens,omitempty" validate:"omitempty,min=1"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
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

// Validate validates the configuration for the specified provider
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

// WithAPIKey returns a new config with the specified API key
func (c Config) WithAPIKey(apiKey string) Config {
	c.APIKey = apiKey
	return c
}

// WithBaseURL returns a new config with the specified base URL
func (c Config) WithBaseURL(baseURL string) Config {
	c.BaseURL = baseURL
	return c
}

// WithTimeout returns a new config with the specified timeout
func (c Config) WithTimeout(timeout time.Duration) Config {
	c.Timeout = timeout
	return c
}

// WithMaxRetries returns a new config with the specified max retries
func (c Config) WithMaxRetries(maxRetries int) Config {
	c.MaxRetries = maxRetries
	return c
}

// WithTemperature returns a new config with the specified temperature
func (c Config) WithTemperature(temperature float64) Config {
	c.Temperature = &temperature
	return c
}

// WithMaxTokens returns a new config with the specified max tokens
func (c Config) WithMaxTokens(maxTokens int) Config {
	c.MaxTokens = &maxTokens
	return c
}

// ValidateProviderType validates that the provider type is supported
func ValidateProviderType(provider ProviderType) error {
	switch provider {
	case ProviderOpenAI, ProviderAnthropic, ProviderGoogle:
		return nil
	default:
		return fmt.Errorf("unsupported provider '%s', supported providers: %v", provider, []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderGoogle})
	}
}
