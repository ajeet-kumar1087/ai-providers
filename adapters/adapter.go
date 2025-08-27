// Package adapters provides the provider adapter interface and implementations.
//
// This package contains the ProviderAdapter interface that defines the contract
// for all AI provider implementations, as well as the concrete adapter
// implementations for each supported provider (OpenAI, Anthropic, Google).
//
// Each adapter is responsible for:
//   - Converting generic requests to provider-specific formats
//   - Making HTTP requests to the provider's API
//   - Parsing and normalizing responses back to generic formats
//   - Handling provider-specific error conditions
//   - Validating provider-specific configuration requirements
//
// The adapter pattern allows the main client to provide a unified interface
// while delegating provider-specific details to specialized implementations.
//
// Supported Adapters:
//   - openai: OpenAI GPT models (GPT-3.5, GPT-4, etc.)
//   - anthropic: Anthropic Claude models (Claude-3, Claude-2, etc.)
//   - google: Google AI Gemini models (implementation in progress)
//
// Example of using an adapter directly (not recommended for normal use):
//
//	adapter, err := openai.NewAdapter(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	response, err := adapter.Complete(ctx, request)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Note: It's recommended to use the main client interface rather than
// adapters directly, as the client provides additional validation,
// parameter normalization, and error handling.
package adapters
