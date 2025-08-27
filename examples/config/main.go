// Package main demonstrates configuration examples
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	wrapper "github.com/ajeet-kumar1087/ai-providers"
)

func main() {
	// Example 1: Configuration with environment variables
	fmt.Println("=== Configuration from Environment Variables ===")
	configFromEnvironment()

	fmt.Println("\n=== Custom Configuration ===")
	customConfiguration()

	fmt.Println("\n=== Configuration with Defaults ===")
	configurationWithDefaults()

	fmt.Println("\n=== Provider Switching ===")
	providerSwitching()

}

// configFromEnvironment demonstrates loading configuration from environment variables
func configFromEnvironment() {
	// Set environment variables (in real usage, these would be set externally)
	os.Setenv("OPENAI_API_KEY", "sk-your-openai-api-key-here")
	os.Setenv("AI_TEMPERATURE", "0.7")
	os.Setenv("AI_MAX_TOKENS", "100")
	os.Setenv("AI_TIMEOUT", "45s")

	// Load configuration from environment
	config := wrapper.LoadConfigFromEnv(wrapper.ProviderOpenAI)

	fmt.Printf("Loaded config - API Key: %s...\n", config.APIKey[:10])
	fmt.Printf("Temperature: %v\n", config.Temperature)
	fmt.Printf("Max Tokens: %v\n", config.MaxTokens)
	fmt.Printf("Timeout: %v\n", config.Timeout)

	// Create client with environment configuration
	client, err := wrapper.NewClientWithEnvConfig(wrapper.ProviderOpenAI)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Test the configuration
	req := wrapper.CompletionRequest{
		Prompt: "Say hello in a creative way:",
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		log.Printf("Completion failed: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Text)
}

// customConfiguration demonstrates creating a custom configuration
func customConfiguration() {
	// Create custom configuration with specific settings
	config := wrapper.Config{
		APIKey:      "sk-your-openai-api-key-here",
		BaseURL:     "https://api.openai.com/v1",               // Custom base URL
		Timeout:     60 * time.Second,                          // 60 second timeout
		MaxRetries:  5,                                         // 5 retry attempts
		Temperature: func() *float64 { t := 0.2; return &t }(), // Low temperature for consistent output
		MaxTokens:   func() *int { m := 200; return &m }(),     // Limit to 200 tokens
	}

	// Create client with custom configuration
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, config)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	fmt.Printf("Custom config - Timeout: %v, Max Retries: %d\n", config.Timeout, config.MaxRetries)
	fmt.Printf("Default Temperature: %v, Default Max Tokens: %v\n", config.Temperature, config.MaxTokens)

	// Test with the custom configuration
	req := wrapper.CompletionRequest{
		Prompt: "Explain what a REST API is:",
		// Note: Temperature and MaxTokens from config will be used as defaults
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		log.Printf("Completion failed: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Text)
}

// configurationWithDefaults demonstrates using default configuration with API key
func configurationWithDefaults() {
	// Create client with defaults (just provide API key)
	client, err := wrapper.NewClientWithDefaults(wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here")
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	fmt.Println("Using default configuration with Anthropic")

	// Test with default settings
	req := wrapper.ChatRequest{
		Messages: []wrapper.Message{
			{
				Role:    "user",
				Content: "What are the benefits of using Go for backend development?",
			},
		},
	}

	resp, err := client.ChatComplete(context.Background(), req)
	if err != nil {
		log.Printf("Chat completion failed: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Message.Content)
}

// providerSwitching demonstrates switching between different providers
func providerSwitching() {
	providers := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	prompt := "Write a one-sentence summary of machine learning:"

	for _, p := range providers {
		fmt.Printf("\n--- Testing with %s ---\n", p.name)

		// Create client for this provider
		client, err := wrapper.NewClientWithDefaults(p.provider, p.apiKey)
		if err != nil {
			log.Printf("Failed to create %s client: %v", p.name, err)
			continue
		}

		// Test completion
		req := wrapper.CompletionRequest{
			Prompt: prompt,
		}

		resp, err := client.Complete(context.Background(), req)
		if err != nil {
			log.Printf("%s completion failed: %v", p.name, err)
			client.Close()
			continue
		}

		fmt.Printf("%s Response: %s\n", p.name, resp.Text)
		fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)

		client.Close()
	}
}
