package aiprovider

import (
	"context"
	"testing"
	"time"
)

// Test client creation and configuration
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		config   Config
		wantErr  bool
		errType  string
		errMsg   string
	}{
		{
			name:     "valid OpenAI client",
			provider: ProviderOpenAI,
			config: Config{
				APIKey:      "sk-1234567890abcdef1234567890abcdef",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				Temperature: floatPtr(0.7),
				MaxTokens:   intPtr(1000),
			},
			wantErr: false,
		},
		{
			name:     "valid Anthropic client",
			provider: ProviderAnthropic,
			config: Config{
				APIKey:      "sk-ant-1234567890abcdef1234567890abcdef",
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				Temperature: floatPtr(0.8),
				MaxTokens:   intPtr(2000),
			},
			wantErr: false,
		},
		{
			name:     "unsupported provider",
			provider: ProviderType("unsupported"),
			config: Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			wantErr: true,
			errType: "validation",
			errMsg:  "unsupported provider",
		},
		{
			name:     "invalid configuration - empty API key",
			provider: ProviderOpenAI,
			config: Config{
				APIKey: "",
			},
			wantErr: true,
			errType: "validation",
			errMsg:  "invalid configuration",
		},
		{
			name:     "invalid configuration - wrong API key format",
			provider: ProviderOpenAI,
			config: Config{
				APIKey: "invalid-key",
			},
			wantErr: true,
			errType: "validation",
			errMsg:  "invalid configuration",
		},
		{
			name:     "Google provider not implemented",
			provider: ProviderGoogle,
			config: Config{
				APIKey: "AIzaSyDaGmWKa4JsXZ-HjGw7ISLan_KqP8o",
			},
			wantErr: true,
			errType: "provider",
			errMsg:  "Google adapter not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientInstance, err := NewClient(tt.provider, tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}

				// Check if it's our custom error type
				if customErr, ok := err.(*Error); ok {
					if tt.errType != "" && string(customErr.Type) != tt.errType {
						t.Errorf("Expected error type %q, got %q", tt.errType, customErr.Type)
					}
					if !contains(customErr.Message, tt.errMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, customErr.Message)
					}
				} else {
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return
				}
				if clientInstance == nil {
					t.Errorf("Expected client to be created")
				}
			}
		})
	}
}

// Test client factory
func TestClientFactory(t *testing.T) {
	factory := NewClientFactory()

	// Test SupportedProviders
	providers := factory.SupportedProviders()
	expectedProviders := []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderGoogle}

	if len(providers) != len(expectedProviders) {
		t.Errorf("Expected %d providers, got %d", len(expectedProviders), len(providers))
	}

	for _, expected := range expectedProviders {
		found := false
		for _, provider := range providers {
			if provider == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected provider %q not found", expected)
		}
	}

	// Test CreateClient
	config := Config{
		APIKey: "sk-1234567890abcdef1234567890abcdef",
	}
	client, err := factory.CreateClient(ProviderOpenAI, config)
	if err != nil {
		t.Errorf("Expected successful client creation, got error: %v", err)
	}
	if client == nil {
		t.Errorf("Expected client to be created")
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	// Test IsValidProvider
	validProviders := []ProviderType{ProviderOpenAI, ProviderAnthropic, ProviderGoogle}
	for _, provider := range validProviders {
		if !IsValidProvider(provider) {
			t.Errorf("Expected %q to be valid provider", provider)
		}
	}

	if IsValidProvider(ProviderType("invalid")) {
		t.Errorf("Expected 'invalid' to be invalid provider")
	}

	// Test GetSupportedProviders
	supported := GetSupportedProviders()
	if len(supported) != 3 {
		t.Errorf("Expected 3 supported providers, got %d", len(supported))
	}

	// Test ValidateProviderType
	err := ValidateProviderType(ProviderOpenAI)
	if err != nil {
		t.Errorf("Expected valid provider type, got error: %v", err)
	}

	err = ValidateProviderType(ProviderType("invalid"))
	if err == nil {
		t.Errorf("Expected invalid provider type to return error")
	}
}

// Test CreateAdapter function
func TestCreateAdapter(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		config   Config
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "create OpenAI adapter",
			provider: ProviderOpenAI,
			config: Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			wantErr: false,
		},
		{
			name:     "create Anthropic adapter",
			provider: ProviderAnthropic,
			config: Config{
				APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
			},
			wantErr: false,
		},
		{
			name:     "create Google adapter (not implemented)",
			provider: ProviderGoogle,
			config: Config{
				APIKey: "AIzaSyDaGmWKa4JsXZ-HjGw7ISLan_KqP8o",
			},
			wantErr: true,
			errMsg:  "Google adapter not yet implemented",
		},
		{
			name:     "invalid provider",
			provider: ProviderType("invalid"),
			config: Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			wantErr: true,
			errMsg:  "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := CreateAdapter(tt.provider, tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
					return
				}
				if adapter == nil {
					t.Errorf("Expected adapter to be created")
				}

				// Verify adapter name matches provider
				expectedName := string(tt.provider)
				if adapter.Name() != expectedName {
					t.Errorf("Expected adapter name %q, got %q", expectedName, adapter.Name())
				}
			}
		})
	}
}

// Test request validation and normalization
func TestRequestValidationAndNormalization(t *testing.T) {
	// Create a client for testing
	config := Config{
		APIKey:      "sk-1234567890abcdef1234567890abcdef",
		Temperature: floatPtr(0.5),
		MaxTokens:   intPtr(500),
	}
	clientInstance, err := NewClient(ProviderOpenAI, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Cast to internal client type to access private methods
	internalClient := clientInstance.(*client)

	t.Run("completion request validation", func(t *testing.T) {
		tests := []struct {
			name    string
			request CompletionRequest
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid request",
				request: CompletionRequest{
					Prompt:      "Hello, world!",
					Temperature: floatPtr(0.7),
					MaxTokens:   intPtr(100),
				},
				wantErr: false,
			},
			{
				name: "empty prompt",
				request: CompletionRequest{
					Prompt: "",
				},
				wantErr: true,
				errMsg:  "prompt is required",
			},
			{
				name: "negative temperature",
				request: CompletionRequest{
					Prompt:      "Hello",
					Temperature: floatPtr(-0.1),
				},
				wantErr: true,
				errMsg:  "temperature must be non-negative",
			},
			{
				name: "zero max tokens",
				request: CompletionRequest{
					Prompt:    "Hello",
					MaxTokens: intPtr(0),
				},
				wantErr: true,
				errMsg:  "max_tokens must be positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				normalized, err := internalClient.validateAndNormalizeCompletionRequest(tt.request)
				if tt.wantErr {
					if err == nil {
						t.Errorf("Expected error, got nil")
						return
					}
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got %v", err)
						return
					}

					// Verify normalization applied config defaults when not specified
					if tt.request.Temperature == nil && normalized.Temperature != nil {
						if *normalized.Temperature != *config.Temperature {
							t.Errorf("Expected default temperature %v, got %v", *config.Temperature, *normalized.Temperature)
						}
					}

					if tt.request.MaxTokens == nil && normalized.MaxTokens != nil {
						if *normalized.MaxTokens != *config.MaxTokens {
							t.Errorf("Expected default max tokens %v, got %v", *config.MaxTokens, *normalized.MaxTokens)
						}
					}
				}
			})
		}
	})

	t.Run("chat request validation", func(t *testing.T) {
		tests := []struct {
			name    string
			request ChatRequest
			wantErr bool
			errMsg  string
		}{
			{
				name: "valid request",
				request: ChatRequest{
					Messages: []Message{
						{Role: "user", Content: "Hello"},
					},
					Temperature: floatPtr(0.7),
					MaxTokens:   intPtr(100),
				},
				wantErr: false,
			},
			{
				name: "valid conversation with system message",
				request: ChatRequest{
					Messages: []Message{
						{Role: "system", Content: "You are helpful"},
						{Role: "user", Content: "Hello"},
						{Role: "assistant", Content: "Hi there!"},
						{Role: "user", Content: "How are you?"},
					},
				},
				wantErr: false,
			},
			{
				name: "empty messages",
				request: ChatRequest{
					Messages: []Message{},
				},
				wantErr: true,
				errMsg:  "messages are required",
			},
			{
				name: "conversation starting with assistant",
				request: ChatRequest{
					Messages: []Message{
						{Role: "assistant", Content: "Hello"},
					},
				},
				wantErr: true,
				errMsg:  "conversation cannot start with assistant message",
			},
			{
				name: "message with empty role",
				request: ChatRequest{
					Messages: []Message{
						{Role: "", Content: "Hello"},
					},
				},
				wantErr: true,
				errMsg:  "role is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				normalized, err := internalClient.validateAndNormalizeChatRequest(tt.request)
				if tt.wantErr {
					if err == nil {
						t.Errorf("Expected error, got nil")
						return
					}
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Errorf("Expected no error, got %v", err)
						return
					}

					// Verify normalization applied config defaults when not specified
					if tt.request.Temperature == nil && normalized.Temperature != nil {
						if *normalized.Temperature != *config.Temperature {
							t.Errorf("Expected default temperature %v, got %v", *config.Temperature, *normalized.Temperature)
						}
					}

					if tt.request.MaxTokens == nil && normalized.MaxTokens != nil {
						if *normalized.MaxTokens != *config.MaxTokens {
							t.Errorf("Expected default max tokens %v, got %v", *config.MaxTokens, *normalized.MaxTokens)
						}
					}
				}
			})
		}
	})
}

// Test parameter clamping across providers
func TestParameterClamping(t *testing.T) {
	tests := []struct {
		name         string
		provider     ProviderType
		config       Config
		request      CompletionRequest
		expectTemp   *float64
		expectTokens *int
	}{
		{
			name:     "OpenAI - high temperature clamped",
			provider: ProviderOpenAI,
			config: Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(2.5), // Should be clamped to 2.0 for OpenAI
				MaxTokens:   intPtr(10000), // Should be clamped to 4096 for OpenAI
			},
			expectTemp:   floatPtr(2.0),
			expectTokens: intPtr(4096),
		},
		{
			name:     "Anthropic - high temperature clamped",
			provider: ProviderAnthropic,
			config: Config{
				APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
			},
			request: CompletionRequest{
				Prompt:      "Test",
				Temperature: floatPtr(1.5),  // Should be clamped to 1.0 for Anthropic
				MaxTokens:   intPtr(200000), // Should be clamped to 100000 for Anthropic
			},
			expectTemp:   floatPtr(1.0),
			expectTokens: intPtr(100000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientInstance, err := NewClient(tt.provider, tt.config)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			internalClient := clientInstance.(*client)
			normalized, err := internalClient.validateAndNormalizeCompletionRequest(tt.request)
			if err != nil {
				t.Fatalf("Expected successful normalization, got error: %v", err)
			}

			// Check temperature clamping
			if tt.expectTemp != nil {
				if normalized.Temperature == nil {
					t.Errorf("Expected temperature %v, got nil", *tt.expectTemp)
				} else if *normalized.Temperature != *tt.expectTemp {
					t.Errorf("Expected temperature %v, got %v", *tt.expectTemp, *normalized.Temperature)
				}
			}

			// Check max tokens clamping
			if tt.expectTokens != nil {
				if normalized.MaxTokens == nil {
					t.Errorf("Expected max tokens %v, got nil", *tt.expectTokens)
				} else if *normalized.MaxTokens != *tt.expectTokens {
					t.Errorf("Expected max tokens %v, got %v", *tt.expectTokens, *normalized.MaxTokens)
				}
			}
		})
	}
}

// Test provider switching
func TestProviderSwitching(t *testing.T) {
	providers := []struct {
		provider ProviderType
		config   Config
	}{
		{
			provider: ProviderOpenAI,
			config: Config{
				APIKey: "sk-1234567890abcdef1234567890abcdef",
			},
		},
		{
			provider: ProviderAnthropic,
			config: Config{
				APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
			},
		},
	}

	for _, p := range providers {
		t.Run(string(p.provider), func(t *testing.T) {
			clientInstance, err := NewClient(p.provider, p.config)
			if err != nil {
				t.Fatalf("Failed to create %s client: %v", p.provider, err)
			}

			// Verify client was created successfully
			if clientInstance == nil {
				t.Errorf("Expected client to be created for %s", p.provider)
			}

			// Test that the client can be closed without error
			err = clientInstance.Close()
			if err != nil {
				t.Errorf("Expected successful close for %s client, got error: %v", p.provider, err)
			}
		})
	}
}

// Test error propagation through client layer
func TestErrorPropagation(t *testing.T) {
	// Test with invalid configuration that should be caught at client level
	_, err := NewClient(ProviderOpenAI, Config{APIKey: ""})
	if err == nil {
		t.Errorf("Expected error for empty API key")
		return
	}

	// Verify error is properly wrapped
	if customErr, ok := err.(*Error); ok {
		if customErr.Type != ErrorTypeValidation {
			t.Errorf("Expected validation error, got %s", customErr.Type)
		}
		if customErr.Provider != string(ProviderOpenAI) {
			t.Errorf("Expected provider %s, got %s", ProviderOpenAI, customErr.Provider)
		}
	} else {
		t.Errorf("Expected custom error type, got %T", err)
	}

	// Test with valid client but invalid request
	validClient, err := NewClient(ProviderOpenAI, Config{
		APIKey: "sk-1234567890abcdef1234567890abcdef",
	})
	if err != nil {
		t.Fatalf("Failed to create valid client: %v", err)
	}

	// Test completion with invalid request
	_, err = validClient.Complete(context.Background(), CompletionRequest{
		Prompt: "", // Empty prompt should cause validation error
	})
	if err == nil {
		t.Errorf("Expected error for empty prompt")
		return
	}

	// Verify error propagation
	if customErr, ok := err.(*Error); ok {
		if customErr.Type != ErrorTypeValidation {
			t.Errorf("Expected validation error, got %s", customErr.Type)
		}
	} else {
		t.Errorf("Expected custom error type, got %T", err)
	}

	// Test chat completion with invalid request
	_, err = validClient.ChatComplete(context.Background(), ChatRequest{
		Messages: []Message{}, // Empty messages should cause validation error
	})
	if err == nil {
		t.Errorf("Expected error for empty messages")
		return
	}

	// Verify error propagation
	if customErr, ok := err.(*Error); ok {
		if customErr.Type != ErrorTypeValidation {
			t.Errorf("Expected validation error, got %s", customErr.Type)
		}
	} else {
		t.Errorf("Expected custom error type, got %T", err)
	}
}

// Test conversation structure validation
func TestConversationStructureValidation(t *testing.T) {
	config := Config{
		APIKey: "sk-ant-1234567890abcdef1234567890abcdef",
	}
	clientInstance, err := NewClient(ProviderAnthropic, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	internalClient := clientInstance.(*client)

	tests := []struct {
		name     string
		messages []Message
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid conversation",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
				{Role: "user", Content: "How are you?"},
			},
			wantErr: false,
		},
		{
			name: "valid with system message",
			messages: []Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hello"},
			},
			wantErr: false,
		},
		{
			name:     "empty conversation",
			messages: []Message{},
			wantErr:  true,
			errMsg:   "conversation must have at least one message",
		},
		{
			name: "starts with assistant",
			messages: []Message{
				{Role: "assistant", Content: "Hello"},
			},
			wantErr: true,
			errMsg:  "conversation cannot start with assistant message",
		},
		{
			name: "too many system messages for Anthropic",
			messages: []Message{
				{Role: "system", Content: "System 1"},
				{Role: "system", Content: "System 2"},
				{Role: "system", Content: "System 3"},
				{Role: "system", Content: "System 4"},
				{Role: "system", Content: "System 5"},
				{Role: "system", Content: "System 6"}, // 6 system messages
				{Role: "user", Content: "Hello"},
			},
			wantErr: true,
			errMsg:  "too many system messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := internalClient.validateConversationStructure(tt.messages)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// Helper functions are in test_utils.go
