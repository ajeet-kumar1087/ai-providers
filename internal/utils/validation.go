// Package utils provides parameter mapping and validation utilities
package utils

import (
	"fmt"
	"strings"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// ParameterMapper provides utilities for mapping parameters between providers
type ParameterMapper struct {
	sourceProvider ProviderType
	targetProvider ProviderType
}

// NewParameterMapper creates a new parameter mapper
func NewParameterMapper(sourceProvider, targetProvider ProviderType) *ParameterMapper {
	return &ParameterMapper{
		sourceProvider: sourceProvider,
		targetProvider: targetProvider,
	}
}

// MapTemperature maps temperature values between providers
func (pm *ParameterMapper) MapTemperature(temperature float64) float64 {
	// OpenAI: 0.0-2.0, Anthropic: 0.0-1.0, Google: 0.0-1.0
	switch pm.targetProvider {
	case types.ProviderAnthropic, types.ProviderGoogle:
		// If source allows higher range (like OpenAI), scale down
		if pm.sourceProvider == types.ProviderOpenAI && temperature > 1.0 {
			return temperature / 2.0 // Simple scaling
		}
		// Clamp to 1.0 for providers that don't support higher values
		if temperature > 1.0 {
			return 1.0
		}
	case types.ProviderOpenAI:
		// OpenAI supports up to 2.0, no scaling needed
		if temperature > 2.0 {
			return 2.0
		}
	}

	// Ensure minimum value
	if temperature < 0.0 {
		return 0.0
	}

	return temperature
}

// MapMaxTokens maps max token values between providers with their limits
func (pm *ParameterMapper) MapMaxTokens(maxTokens int) int {
	targetLimit := GetProviderTokenLimit(pm.targetProvider)

	if maxTokens > targetLimit {
		return targetLimit
	}

	if maxTokens <= 0 {
		return GetDefaultMaxTokens(pm.targetProvider)
	}

	return maxTokens
}

// MapStopSequences maps stop sequences between providers
func (pm *ParameterMapper) MapStopSequences(stopSequences []string) []string {
	maxStop := GetProviderMaxStopSequences(pm.targetProvider)

	if len(stopSequences) <= maxStop {
		return stopSequences
	}

	// Truncate to provider limit
	return stopSequences[:maxStop]
}

// Provider-specific limits and defaults

// GetProviderTokenLimit returns the maximum token limit for a provider
func GetProviderTokenLimit(provider ProviderType) int {
	switch provider {
	case types.ProviderOpenAI:
		return 4096 // Conservative limit for GPT-3.5/4
	case types.ProviderAnthropic:
		return 100000 // Claude models support up to 100k tokens
	case types.ProviderGoogle:
		return 8192 // Conservative limit for Gemini
	default:
		return 4096 // Conservative default
	}
}

// GetProviderMaxTemperature returns the maximum temperature for a provider
func GetProviderMaxTemperature(provider ProviderType) float64 {
	switch provider {
	case types.ProviderOpenAI:
		return 2.0
	case types.ProviderAnthropic:
		return 1.0
	case types.ProviderGoogle:
		return 1.0
	default:
		return 1.0 // Conservative default
	}
}

// GetProviderMaxStopSequences returns the maximum number of stop sequences for a provider
func GetProviderMaxStopSequences(provider ProviderType) int {
	switch provider {
	case types.ProviderOpenAI:
		return 4 // OpenAI supports up to 4 stop sequences
	case types.ProviderAnthropic:
		return 10 // Anthropic supports more stop sequences
	case types.ProviderGoogle:
		return 5 // Conservative limit for Google
	default:
		return 4 // Conservative default
	}
}

// GetDefaultMaxTokens returns a sensible default for max tokens for a provider
func GetDefaultMaxTokens(provider ProviderType) int {
	switch provider {
	case types.ProviderOpenAI:
		return 1024
	case types.ProviderAnthropic:
		return 1024
	case types.ProviderGoogle:
		return 1024
	default:
		return 1024
	}
}

// Validation utilities

// ValidateCompletionRequest validates a completion request (basic validation only)
func ValidateCompletionRequest(req types.CompletionRequest) error {
	if strings.TrimSpace(req.Prompt) == "" {
		return fmt.Errorf("prompt is required and cannot be empty")
	}

	if req.Temperature != nil {
		temp := *req.Temperature
		if temp < 0.0 {
			return fmt.Errorf("temperature must be non-negative, got: %f", temp)
		}
		// Don't validate upper bound here - let provider-specific validation handle it
	}

	if req.MaxTokens != nil {
		tokens := *req.MaxTokens
		if tokens <= 0 {
			return fmt.Errorf("max_tokens must be positive, got: %d", tokens)
		}
		// Don't validate upper bound here - let provider-specific validation handle it
	}

	return nil
}

// ValidateChatRequest validates a chat request (basic validation only)
func ValidateChatRequest(req types.ChatRequest) error {
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}

	for i, msg := range req.Messages {
		if err := ValidateMessage(msg, i); err != nil {
			return err
		}
	}

	if req.Temperature != nil {
		temp := *req.Temperature
		if temp < 0.0 {
			return fmt.Errorf("temperature must be non-negative, got: %f", temp)
		}
		// Don't validate upper bound here - let provider-specific validation handle it
	}

	if req.MaxTokens != nil {
		tokens := *req.MaxTokens
		if tokens <= 0 {
			return fmt.Errorf("max_tokens must be positive, got: %d", tokens)
		}
		// Don't validate upper bound here - let provider-specific validation handle it
	}

	return nil
}

// ValidateMessage validates a single message
func ValidateMessage(msg types.Message, index int) error {
	if strings.TrimSpace(msg.Role) == "" {
		return fmt.Errorf("message %d: role is required", index)
	}

	if strings.TrimSpace(msg.Content) == "" {
		return fmt.Errorf("message %d: content is required", index)
	}

	// Validate role values
	switch msg.Role {
	case "user", "assistant", "system":
		// Valid roles
	default:
		return fmt.Errorf("message %d: invalid role '%s', must be one of: user, assistant, system", index, msg.Role)
	}

	return nil
}

// ClampParameters clamps parameters to provider-specific ranges
func ClampParameters(req interface{}, provider ProviderType) interface{} {
	switch r := req.(type) {
	case types.CompletionRequest:
		return clampCompletionRequest(r, provider)
	case types.ChatRequest:
		return clampChatRequest(r, provider)
	default:
		return req
	}
}

// clampCompletionRequest clamps completion request parameters
func clampCompletionRequest(req types.CompletionRequest, provider ProviderType) types.CompletionRequest {
	clamped := req

	// Clamp temperature
	if clamped.Temperature != nil {
		temp := *clamped.Temperature
		maxTemp := GetProviderMaxTemperature(provider)
		if temp > maxTemp {
			temp = maxTemp
		}
		if temp < 0.0 {
			temp = 0.0
		}
		clamped.Temperature = &temp
	}

	// Clamp max tokens
	if clamped.MaxTokens != nil {
		tokens := *clamped.MaxTokens
		maxTokens := GetProviderTokenLimit(provider)
		if tokens > maxTokens {
			tokens = maxTokens
		}
		if tokens <= 0 {
			tokens = GetDefaultMaxTokens(provider)
		}
		clamped.MaxTokens = &tokens
	}

	// Clamp stop sequences
	maxStop := GetProviderMaxStopSequences(provider)
	if len(clamped.Stop) > maxStop {
		clamped.Stop = clamped.Stop[:maxStop]
	}

	return clamped
}

// clampChatRequest clamps chat request parameters
func clampChatRequest(req types.ChatRequest, provider ProviderType) types.ChatRequest {
	clamped := req

	// Clamp temperature
	if clamped.Temperature != nil {
		temp := *clamped.Temperature
		maxTemp := GetProviderMaxTemperature(provider)
		if temp > maxTemp {
			temp = maxTemp
		}
		if temp < 0.0 {
			temp = 0.0
		}
		clamped.Temperature = &temp
	}

	// Clamp max tokens
	if clamped.MaxTokens != nil {
		tokens := *clamped.MaxTokens
		maxTokens := GetProviderTokenLimit(provider)
		if tokens > maxTokens {
			tokens = maxTokens
		}
		if tokens <= 0 {
			tokens = GetDefaultMaxTokens(provider)
		}
		clamped.MaxTokens = &tokens
	}

	return clamped
}

// Type alias for convenience
type ProviderType = types.ProviderType
