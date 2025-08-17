package main

import (
	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

// Re-export types from the types package for backward compatibility
type CompletionRequest = types.CompletionRequest
type CompletionResponse = types.CompletionResponse
type ChatRequest = types.ChatRequest
type ChatResponse = types.ChatResponse
type Message = types.Message
type Usage = types.Usage
type ProviderType = types.ProviderType
type Config = types.Config

// Re-export constants
const (
	ProviderOpenAI    = types.ProviderOpenAI
	ProviderAnthropic = types.ProviderAnthropic
	ProviderGoogle    = types.ProviderGoogle
)
