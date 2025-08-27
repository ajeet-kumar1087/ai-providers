// Package main demonstrates basic completion usage
// This example shows how to use the AI Provider Wrapper library
//
// To use this example with the actual library:
// 1. Replace the API keys with your actual keys
// 2. Import the library: go get github.com/ai-provider-wrapper/ai-provider-wrapper
// 3. Import as: import wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
// 4. Replace mock types with wrapper.* types
//
// Note: This example uses mock implementations for demonstration purposes
package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Mock types for demonstration - replace with actual library imports
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)

type Config struct {
	APIKey      string
	BaseURL     string
	Timeout     time.Duration
	MaxRetries  int
	Temperature *float64
	MaxTokens   *int
}

type CompletionRequest struct {
	Prompt      string
	Temperature *float64
	MaxTokens   *int
	Stop        []string
	Stream      bool
}

type CompletionResponse struct {
	Text         string
	Usage        Usage
	FinishReason string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type Client interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Close() error
}

// Mock client implementation for demonstration
type mockClient struct {
	provider ProviderType
	config   Config
}

func NewClient(provider ProviderType, config Config) (Client, error) {
	fmt.Printf("Creating %s client with API key: %s...\n", provider, config.APIKey[:10])
	return &mockClient{provider: provider, config: config}, nil
}

func (c *mockClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Simulate API call
	fmt.Printf("Sending request to %s: %s\n", c.provider, req.Prompt)

	// Mock response based on provider
	var responseText string
	switch c.provider {
	case ProviderOpenAI:
		responseText = "Code flows like water,\nBugs dance in morning sunlight,\nDebug brings peace."
	case ProviderAnthropic:
		responseText = "Recursion is a programming technique where a function calls itself to solve smaller instances of the same problem, eventually reaching a base case that stops the recursion."
	default:
		responseText = "Mock response from " + string(c.provider)
	}

	return &CompletionResponse{
		Text: responseText,
		Usage: Usage{
			PromptTokens:     len(req.Prompt) / 4, // Rough token estimate
			CompletionTokens: len(responseText) / 4,
			TotalTokens:      (len(req.Prompt) + len(responseText)) / 4,
		},
		FinishReason: "stop",
	}, nil
}

func (c *mockClient) Close() error {
	fmt.Printf("Closing %s client\n", c.provider)
	return nil
}

func main() {
	fmt.Println("AI Provider Wrapper - Basic Usage Examples")
	fmt.Println("==========================================")
	fmt.Println()

	// Example 1: Basic completion with OpenAI
	fmt.Println("=== Basic OpenAI Completion ===")
	basicOpenAICompletion()

	fmt.Println("\n=== Basic Anthropic Completion ===")
	basicAnthropicCompletion()

	fmt.Println("\n=== Basic Completion with Custom Parameters ===")
	completionWithParameters()
}

// basicOpenAICompletion demonstrates basic text completion with OpenAI
func basicOpenAICompletion() {
	// Create client with OpenAI provider
	client, err := NewClient(ProviderOpenAI, Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create OpenAI client: %v", err)
		return
	}
	defer client.Close()

	// Create a simple completion request
	req := CompletionRequest{
		Prompt: "Write a haiku about programming:",
	}

	// Send the completion request
	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		log.Printf("Completion failed: %v", err)
		return
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp.Text)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
}

// basicAnthropicCompletion demonstrates basic text completion with Anthropic
func basicAnthropicCompletion() {
	// Create client with Anthropic provider
	client, err := NewClient(ProviderAnthropic, Config{
		APIKey: "sk-ant-your-anthropic-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create Anthropic client: %v", err)
		return
	}
	defer client.Close()

	// Create a simple completion request
	req := CompletionRequest{
		Prompt: "Explain the concept of recursion in programming in simple terms:",
	}

	// Send the completion request
	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		log.Printf("Completion failed: %v", err)
		return
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp.Text)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
}

// completionWithParameters demonstrates completion with custom parameters
func completionWithParameters() {
	// Create client with default configuration
	client, err := NewClient(ProviderOpenAI, Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Create completion request with custom parameters
	temperature := 0.7
	maxTokens := 100
	req := CompletionRequest{
		Prompt:      "Write a creative story beginning with 'Once upon a time in a digital world':",
		Temperature: &temperature,     // More creative output
		MaxTokens:   &maxTokens,       // Limit response length
		Stop:        []string{"\n\n"}, // Stop at double newline
	}

	// Send the completion request
	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		log.Printf("Completion failed: %v", err)
		return
	}

	// Print the response
	fmt.Printf("Response: %s\n", resp.Text)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
}
