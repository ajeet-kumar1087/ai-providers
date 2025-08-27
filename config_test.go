package aiprovider

import (
	"os"
	"testing"
	"time"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// Test Config validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   types.Config
		provider types.ProviderType
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid OpenAI config",
			config: types.Config{
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
			},
			provider: types.ProviderOpenAI,
			wantErr:  false,
		},
		{
			name: "valid Anthropic config",
			config: types.Config{
				APIKey:     "sk-ant-1234567890abcdef1234567890abcdef",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			provider: types.ProviderAnthropic,
			wantErr:  false,
		},
		{
			name: "valid Google config",
			config: types.Config{
				APIKey:     "AIzaSyDaGmWKa4JsXZ-HjGw7ISLan_KqP8o",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
			},
			provider: types.ProviderGoogle,
			wantErr:  false,
		},
		{
			name: "empty API key",
			config: types.Config{
				APIKey: "",
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "api key is required",
		},
		{
			name: "whitespace API key",
			config: types.Config{
				APIKey: "   ",
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "api key is required",
		},
		{
			name: "invalid provider type",
			config: types.Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			provider: types.ProviderType("invalid"),
			wantErr:  true,
			errMsg:   "unsupported provider",
		},
		{
			name: "invalid OpenAI API key format",
			config: types.Config{
				APIKey: "invalid-key",
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "openAI API key should start with 'sk-'",
		},
		{
			name: "short OpenAI API key",
			config: types.Config{
				APIKey: "sk-short",
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "openAI API key appears to be too short",
		},
		{
			name: "invalid Anthropic API key format",
			config: types.Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			provider: types.ProviderAnthropic,
			wantErr:  true,
			errMsg:   "anthropic API key should start with 'sk-ant-'",
		},
		{
			name: "short Anthropic API key",
			config: types.Config{
				APIKey: "sk-ant-short",
			},
			provider: types.ProviderAnthropic,
			wantErr:  true,
			errMsg:   "anthropic API key appears to be too short",
		},
		{
			name: "short Google API key",
			config: types.Config{
				APIKey: "short",
			},
			provider: types.ProviderGoogle,
			wantErr:  true,
			errMsg:   "google API key appears to be too short",
		},
		{
			name: "negative timeout",
			config: types.Config{
				APIKey:  "sk-1234567890abcdef1234567890abcdef",
				Timeout: -5 * time.Second,
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "timeout must be non-negative",
		},
		{
			name: "negative max retries",
			config: types.Config{
				APIKey:     "sk-1234567890abcdef1234567890abcdef",
				MaxRetries: -1,
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "max retries must be non-negative",
		},
		{
			name: "temperature too low",
			config: types.Config{
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
				Temperature: floatPtr(-0.1),
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "temperature must be between 0.0 and 2.0",
		},
		{
			name: "temperature too high",
			config: types.Config{
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
				Temperature: floatPtr(2.1),
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "temperature must be between 0.0 and 2.0",
		},
		{
			name: "zero max tokens",
			config: types.Config{
				APIKey:    "sk-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(0),
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "max tokens must be positive",
		},
		{
			name: "negative max tokens",
			config: types.Config{
				APIKey:    "sk-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(-100),
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "max tokens must be positive",
		},
		{
			name: "max tokens exceeds OpenAI limit",
			config: types.Config{
				APIKey:    "sk-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(10000),
			},
			provider: types.ProviderOpenAI,
			wantErr:  true,
			errMsg:   "max tokens exceeds provider limit",
		},
		{
			name: "max tokens exceeds Google limit",
			config: types.Config{
				APIKey:    "AIzaSyDaGmWKa4JsXZ-HjGw7ISLan_KqP8o",
				MaxTokens: intPtr(20000),
			},
			provider: types.ProviderGoogle,
			wantErr:  true,
			errMsg:   "max tokens exceeds provider limit",
		},
		{
			name: "valid max tokens for Anthropic",
			config: types.Config{
				APIKey:    "sk-ant-1234567890abcdef1234567890abcdef",
				MaxTokens: intPtr(50000),
			},
			provider: types.ProviderAnthropic,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate(tt.provider)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// Test DefaultConfig
func TestDefaultConfig(t *testing.T) {
	config := types.DefaultConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("Default timeout = %v, want %v", config.Timeout, 30*time.Second)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Default MaxRetries = %v, want %v", config.MaxRetries, 3)
	}

	if config.APIKey != "" {
		t.Errorf("Default APIKey should be empty, got %q", config.APIKey)
	}

	if config.BaseURL != "" {
		t.Errorf("Default BaseURL should be empty, got %q", config.BaseURL)
	}

	if config.Temperature != nil {
		t.Errorf("Default Temperature should be nil, got %v", config.Temperature)
	}

	if config.MaxTokens != nil {
		t.Errorf("Default MaxTokens should be nil, got %v", config.MaxTokens)
	}
}

// Test LoadConfigFromEnv
func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"OPENAI_API_KEY", "OPENAI_BASE_URL",
		"ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL",
		"GOOGLE_API_KEY", "GOOGLE_BASE_URL",
		"AI_TIMEOUT", "AI_MAX_RETRIES", "AI_TEMPERATURE", "AI_MAX_TOKENS",
	}

	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	tests := []struct {
		name     string
		provider types.ProviderType
		envVars  map[string]string
		expected types.Config
	}{
		{
			name:     "OpenAI with environment variables",
			provider: types.ProviderOpenAI,
			envVars: map[string]string{
				"OPENAI_API_KEY":  "sk-test123",
				"OPENAI_BASE_URL": "https://api.openai.com/v1",
				"AI_TIMEOUT":      "45s",
				"AI_MAX_RETRIES":  "5",
				"AI_TEMPERATURE":  "0.8",
				"AI_MAX_TOKENS":   "2000",
			},
			expected: types.Config{
				APIKey:      "sk-test123",
				BaseURL:     "https://api.openai.com/v1",
				Timeout:     45 * time.Second,
				MaxRetries:  5,
				Temperature: floatPtr(0.8),
				MaxTokens:   intPtr(2000),
			},
		},
		{
			name:     "Anthropic with environment variables",
			provider: types.ProviderAnthropic,
			envVars: map[string]string{
				"ANTHROPIC_API_KEY":  "sk-ant-test123",
				"ANTHROPIC_BASE_URL": "https://api.anthropic.com",
			},
			expected: types.Config{
				APIKey:     "sk-ant-test123",
				BaseURL:    "https://api.anthropic.com",
				Timeout:    30 * time.Second, // Default
				MaxRetries: 3,                // Default
			},
		},
		{
			name:     "Google with environment variables",
			provider: types.ProviderGoogle,
			envVars: map[string]string{
				"GOOGLE_API_KEY":  "AIzaSyTest123",
				"GOOGLE_BASE_URL": "https://generativelanguage.googleapis.com",
			},
			expected: types.Config{
				APIKey:     "AIzaSyTest123",
				BaseURL:    "https://generativelanguage.googleapis.com",
				Timeout:    30 * time.Second, // Default
				MaxRetries: 3,                // Default
			},
		},
		{
			name:     "no environment variables",
			provider: types.ProviderOpenAI,
			envVars:  map[string]string{},
			expected: types.Config{
				Timeout:    30 * time.Second, // Default
				MaxRetries: 3,                // Default
			},
		},
		{
			name:     "invalid environment values ignored",
			provider: types.ProviderOpenAI,
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-test123",
				"AI_TIMEOUT":     "invalid",
				"AI_MAX_RETRIES": "invalid",
				"AI_TEMPERATURE": "invalid",
				"AI_MAX_TOKENS":  "invalid",
			},
			expected: types.Config{
				APIKey:     "sk-test123",
				Timeout:    30 * time.Second, // Default (invalid ignored)
				MaxRetries: 3,                // Default (invalid ignored)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := types.LoadConfigFromEnv(tt.provider)

			// Compare fields
			if config.APIKey != tt.expected.APIKey {
				t.Errorf("APIKey = %q, want %q", config.APIKey, tt.expected.APIKey)
			}
			if config.BaseURL != tt.expected.BaseURL {
				t.Errorf("BaseURL = %q, want %q", config.BaseURL, tt.expected.BaseURL)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
			}
			if config.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", config.MaxRetries, tt.expected.MaxRetries)
			}
			if !equalFloatPtr(config.Temperature, tt.expected.Temperature) {
				t.Errorf("Temperature = %v, want %v", config.Temperature, tt.expected.Temperature)
			}
			if !equalIntPtr(config.MaxTokens, tt.expected.MaxTokens) {
				t.Errorf("MaxTokens = %v, want %v", config.MaxTokens, tt.expected.MaxTokens)
			}

			// Clean up environment variables for next test
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

// Test Config builder methods
func TestConfigBuilderMethods(t *testing.T) {
	baseConfig := types.Config{
		APIKey:     "sk-test",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	// Test WithAPIKey
	newConfig := baseConfig.WithAPIKey("sk-new")
	if newConfig.APIKey != "sk-new" {
		t.Errorf("WithAPIKey: APIKey = %q, want %q", newConfig.APIKey, "sk-new")
	}
	// Original should be unchanged
	if baseConfig.APIKey != "sk-test" {
		t.Errorf("WithAPIKey modified original config")
	}

	// Test WithBaseURL
	newConfig = baseConfig.WithBaseURL("https://api.example.com")
	if newConfig.BaseURL != "https://api.example.com" {
		t.Errorf("WithBaseURL: BaseURL = %q, want %q", newConfig.BaseURL, "https://api.example.com")
	}

	// Test WithTimeout
	newConfig = baseConfig.WithTimeout(60 * time.Second)
	if newConfig.Timeout != 60*time.Second {
		t.Errorf("WithTimeout: Timeout = %v, want %v", newConfig.Timeout, 60*time.Second)
	}

	// Test WithMaxRetries
	newConfig = baseConfig.WithMaxRetries(5)
	if newConfig.MaxRetries != 5 {
		t.Errorf("WithMaxRetries: MaxRetries = %v, want %v", newConfig.MaxRetries, 5)
	}

	// Test WithTemperature
	newConfig = baseConfig.WithTemperature(0.8)
	if newConfig.Temperature == nil || *newConfig.Temperature != 0.8 {
		t.Errorf("WithTemperature: Temperature = %v, want %v", newConfig.Temperature, 0.8)
	}

	// Test WithMaxTokens
	newConfig = baseConfig.WithMaxTokens(2000)
	if newConfig.MaxTokens == nil || *newConfig.MaxTokens != 2000 {
		t.Errorf("WithMaxTokens: MaxTokens = %v, want %v", newConfig.MaxTokens, 2000)
	}
}

// Test ValidateProviderType
func TestValidateProviderType(t *testing.T) {
	tests := []struct {
		name     string
		provider types.ProviderType
		wantErr  bool
	}{
		{"valid OpenAI", types.ProviderOpenAI, false},
		{"valid Anthropic", types.ProviderAnthropic, false},
		{"valid Google", types.ProviderGoogle, false},
		{"invalid provider", types.ProviderType("invalid"), true},
		{"empty provider", types.ProviderType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateProviderType(tt.provider)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for provider %q", tt.provider)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for provider %q: %v", tt.provider, err)
			}
		})
	}
}

// Test NewClientWithEnvConfig
func TestNewClientWithEnvConfig(t *testing.T) {
	// This test requires a valid client implementation, so we'll test the function exists
	// and handles the basic case of missing API key
	_, err := NewClientWithEnvConfig(types.ProviderOpenAI)
	if err == nil {
		t.Errorf("Expected error when no API key is set in environment")
	}
}

// Test NewClientWithDefaults
func TestNewClientWithDefaults(t *testing.T) {
	// This test requires a valid client implementation, so we'll test the function exists
	// and handles the basic case of invalid API key format
	_, err := NewClientWithDefaults(types.ProviderOpenAI, "invalid-key")
	if err == nil {
		t.Errorf("Expected error for invalid API key format")
	}
}

// Helper functions are in test_utils.go
