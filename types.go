package aiprovider

import (
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// Re-export types from the types package for backward compatibility.
// These type aliases allow users to import only the main package while
// still having access to all necessary types for API interactions.

// CompletionRequest represents a text completion request.
// See types.CompletionRequest for detailed documentation.
type CompletionRequest = types.CompletionRequest

// CompletionResponse represents a text completion response.
// See types.CompletionResponse for detailed documentation.
type CompletionResponse = types.CompletionResponse

// ChatRequest represents a chat completion request with conversation history.
// See types.ChatRequest for detailed documentation.
type ChatRequest = types.ChatRequest

// ChatResponse represents a chat completion response.
// See types.ChatResponse for detailed documentation.
type ChatResponse = types.ChatResponse

// Message represents a single message in a conversation.
// See types.Message for detailed documentation.
type Message = types.Message

// Usage represents token usage information for API requests.
// See types.Usage for detailed documentation.
type Usage = types.Usage

// ProviderType represents the type of AI provider.
// See types.ProviderType for detailed documentation.
type ProviderType = types.ProviderType

// Config represents the configuration for an AI provider client.
// See types.Config for detailed documentation.
type Config = types.Config

// Re-export provider type constants for convenient access.
// These constants identify the supported AI providers.
const (
	// ProviderOpenAI represents the OpenAI provider (GPT models).
	ProviderOpenAI = types.ProviderOpenAI

	// ProviderAnthropic represents the Anthropic provider (Claude models).
	ProviderAnthropic = types.ProviderAnthropic

	// ProviderGoogle represents the Google AI provider (Gemini models).
	ProviderGoogle = types.ProviderGoogle
)
