// Package main demonstrates provider switching and comparison
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
	// Example 1: Basic provider switching
	fmt.Println("=== Basic Provider Switching ===")
	basicProviderSwitching()

	fmt.Println("\n=== Provider Performance Comparison ===")
	providerPerformanceComparison()

	fmt.Println("\n=== Fallback Provider Strategy ===")
	fallbackProviderStrategy()

	fmt.Println("\n=== Provider Feature Comparison ===")
	providerFeatureComparison()
}

// basicProviderSwitching demonstrates switching between providers for the same task
func basicProviderSwitching() {
	providers := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	prompt := "Explain the difference between machine learning and artificial intelligence:"

	for _, p := range providers {
		fmt.Printf("\n--- Using %s ---\n", p.name)

		client, err := wrapper.NewClient(p.provider, wrapper.Config{
			APIKey: p.apiKey,
		})
		if err != nil {
			log.Printf("Failed to create %s client: %v", p.name, err)
			continue
		}

		req := wrapper.CompletionRequest{
			Prompt:      prompt,
			Temperature: func() *float64 { t := 0.7; return &t }(),
			MaxTokens:   func() *int { m := 200; return &m }(),
		}

		start := time.Now()
		resp, err := client.Complete(context.Background(), req)
		duration := time.Since(start)

		if err != nil {
			log.Printf("%s completion failed: %v", p.name, err)
			client.Close()
			continue
		}

		fmt.Printf("Response: %s\n", resp.Text)
		fmt.Printf("Tokens: %d (prompt: %d, completion: %d)\n",
			resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		fmt.Printf("Duration: %v\n", duration)
		fmt.Printf("Finish reason: %s\n", resp.FinishReason)

		client.Close()
	}
}

// providerPerformanceComparison compares response times and token usage
func providerPerformanceComparison() {
	providers := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	testPrompts := []string{
		"Write a short summary of quantum computing:",
		"Explain recursion in programming:",
		"What are the benefits of cloud computing?",
	}

	results := make(map[string][]PerformanceResult)

	for _, p := range providers {
		fmt.Printf("\n--- Performance Testing: %s ---\n", p.name)

		client, err := wrapper.NewClient(p.provider, wrapper.Config{
			APIKey: p.apiKey,
		})
		if err != nil {
			log.Printf("Failed to create %s client: %v", p.name, err)
			continue
		}

		var providerResults []PerformanceResult

		for i, prompt := range testPrompts {
			fmt.Printf("Test %d/%d: ", i+1, len(testPrompts))

			req := wrapper.CompletionRequest{
				Prompt:      prompt,
				Temperature: func() *float64 { t := 0.5; return &t }(),
				MaxTokens:   func() *int { m := 100; return &m }(),
			}

			start := time.Now()
			resp, err := client.Complete(context.Background(), req)
			duration := time.Since(start)

			if err != nil {
				fmt.Printf("Failed - %v\n", err)
				continue
			}

			result := PerformanceResult{
				Prompt:           prompt,
				Duration:         duration,
				TokensUsed:       resp.Usage.TotalTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				ResponseLength:   len(resp.Text),
			}

			providerResults = append(providerResults, result)
			fmt.Printf("Success - %v, %d tokens\n", duration, resp.Usage.TotalTokens)
		}

		results[p.name] = providerResults
		client.Close()
	}

	// Print comparison summary
	fmt.Println("\n=== Performance Summary ===")
	for providerName, providerResults := range results {
		if len(providerResults) == 0 {
			continue
		}

		var totalDuration time.Duration
		var totalTokens int
		for _, result := range providerResults {
			totalDuration += result.Duration
			totalTokens += result.TokensUsed
		}

		avgDuration := totalDuration / time.Duration(len(providerResults))
		avgTokens := totalTokens / len(providerResults)

		fmt.Printf("%s: Avg Duration: %v, Avg Tokens: %d\n",
			providerName, avgDuration, avgTokens)
	}
}

// PerformanceResult holds performance metrics for a single request
type PerformanceResult struct {
	Prompt           string
	Duration         time.Duration
	TokensUsed       int
	CompletionTokens int
	ResponseLength   int
}

// fallbackProviderStrategy demonstrates implementing a fallback strategy
func fallbackProviderStrategy() {
	// Define providers in order of preference
	providerChain := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI (Primary)", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic (Fallback)", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	prompt := "Write a haiku about programming:"

	// Try each provider in order until one succeeds
	for i, p := range providerChain {
		fmt.Printf("\nTrying %s...\n", p.name)

		client, err := wrapper.NewClient(p.provider, wrapper.Config{
			APIKey: p.apiKey,
		})
		if err != nil {
			fmt.Printf("Failed to create client: %v\n", err)
			if i < len(providerChain)-1 {
				fmt.Println("Falling back to next provider...")
			}
			continue
		}

		req := wrapper.CompletionRequest{
			Prompt: prompt,
		}

		resp, err := client.Complete(context.Background(), req)
		client.Close()

		if err != nil {
			fmt.Printf("Request failed: %v\n", err)
			if i < len(providerChain)-1 {
				fmt.Println("Falling back to next provider...")
			}
			continue
		}

		// Success!
		fmt.Printf("Success with %s!\n", p.name)
		fmt.Printf("Response: %s\n", resp.Text)
		fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
		return
	}

	fmt.Println("All providers failed!")
}

// providerFeatureComparison demonstrates different provider capabilities
func providerFeatureComparison() {
	providers := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	// Test different features and parameter ranges
	featureTests := []struct {
		name        string
		description string
		testFunc    func(wrapper.Client, string) error
	}{
		{
			name:        "High Temperature",
			description: "Testing temperature = 1.8",
			testFunc:    testHighTemperature,
		},
		{
			name:        "Large Token Limit",
			description: "Testing max_tokens = 1000",
			testFunc:    testLargeTokenLimit,
		},
		{
			name:        "Multiple Stop Sequences",
			description: "Testing 5 stop sequences",
			testFunc:    testMultipleStopSequences,
		},
		{
			name:        "Long Conversation",
			description: "Testing chat with 10 messages",
			testFunc:    testLongConversation,
		},
	}

	for _, provider := range providers {
		fmt.Printf("\n=== Testing %s Features ===\n", provider.name)

		client, err := wrapper.NewClient(provider.provider, wrapper.Config{
			APIKey: provider.apiKey,
		})
		if err != nil {
			log.Printf("Failed to create %s client: %v", provider.name, err)
			continue
		}

		for _, test := range featureTests {
			fmt.Printf("\n--- %s: %s ---\n", test.name, test.description)

			err := test.testFunc(client, provider.name)
			if err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
			} else {
				fmt.Printf("✅ Supported\n")
			}
		}

		client.Close()
	}
}

// testHighTemperature tests high temperature values
func testHighTemperature(client wrapper.Client, providerName string) error {
	temperature := 1.8
	req := wrapper.CompletionRequest{
		Prompt:      "Write a creative story opening:",
		Temperature: &temperature,
		MaxTokens:   func() *int { m := 50; return &m }(),
	}

	_, err := client.Complete(context.Background(), req)
	return err
}

// testLargeTokenLimit tests large token limits
func testLargeTokenLimit(client wrapper.Client, providerName string) error {
	maxTokens := 1000
	req := wrapper.CompletionRequest{
		Prompt:    "Write a detailed explanation of machine learning:",
		MaxTokens: &maxTokens,
	}

	_, err := client.Complete(context.Background(), req)
	return err
}

// testMultipleStopSequences tests multiple stop sequences
func testMultipleStopSequences(client wrapper.Client, providerName string) error {
	req := wrapper.CompletionRequest{
		Prompt:    "List programming languages: 1.",
		MaxTokens: func() *int { m := 100; return &m }(),
		Stop:      []string{".", "!", "?", "\n\n", "END"},
	}

	_, err := client.Complete(context.Background(), req)
	return err
}

// testLongConversation tests chat with many messages
func testLongConversation(client wrapper.Client, providerName string) error {
	messages := []wrapper.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm doing well, thank you!"},
		{Role: "user", Content: "What's the weather like?"},
		{Role: "assistant", Content: "I don't have access to current weather data."},
		{Role: "user", Content: "That's okay. Can you help me with programming?"},
		{Role: "assistant", Content: "Absolutely! I'd be happy to help with programming."},
		{Role: "user", Content: "What's your favorite programming language?"},
	}

	req := wrapper.ChatRequest{
		Messages:  messages,
		MaxTokens: func() *int { m := 100; return &m }(),
	}

	_, err := client.ChatComplete(context.Background(), req)
	return err
}
