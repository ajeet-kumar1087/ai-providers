// Package main demonstrates error handling and retry logic
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	wrapper "github.com/ajeet-kumar1087/ai-providers"
)

func main() {
	// Example 1: Basic error handling
	fmt.Println("=== Basic Error Handling ===")
	basicErrorHandling()

	fmt.Println("\n=== Error Type Checking ===")
	errorTypeChecking()

	fmt.Println("\n=== Retry Logic with Rate Limiting ===")
	retryLogicExample()

	fmt.Println("\n=== Handling Different Error Types ===")
	handleDifferentErrorTypes()
}

// basicErrorHandling demonstrates basic error handling patterns
func basicErrorHandling() {
	// Create client with invalid API key to demonstrate error handling
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "invalid-api-key", // This will cause an authentication error
	})
	if err != nil {
		// Handle client creation errors
		if wrapperErr, ok := err.(*wrapper.Error); ok {
			fmt.Printf("Client creation failed - Type: %s, Message: %s\n",
				wrapperErr.Type, wrapperErr.Message)
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
		return
	}
	defer client.Close()

	// Attempt completion with invalid client
	req := wrapper.CompletionRequest{
		Prompt: "Hello, world!",
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		// Handle completion errors
		if wrapperErr, ok := err.(*wrapper.Error); ok {
			fmt.Printf("Completion failed - Type: %s, Provider: %s, Message: %s\n",
				wrapperErr.Type, wrapperErr.Provider, wrapperErr.Message)

			// Check if error has additional context
			if wrapperErr.Code != "" {
				fmt.Printf("Error code: %s\n", wrapperErr.Code)
			}
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
		return
	}

	fmt.Printf("Response: %s\n", resp.Text)
}

// errorTypeChecking demonstrates checking specific error types
func errorTypeChecking() {
	// Create client that might encounter various errors
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-invalid-key-format", // Invalid format
	})
	if err != nil {
		handleSpecificError(err)
		return
	}
	defer client.Close()

	// Test with various problematic requests
	testCases := []struct {
		name string
		req  wrapper.CompletionRequest
	}{
		{
			name: "Empty prompt",
			req: wrapper.CompletionRequest{
				Prompt: "", // This should cause a validation error
			},
		},
		{
			name: "Invalid temperature",
			req: wrapper.CompletionRequest{
				Prompt:      "Hello",
				Temperature: func() *float64 { t := 5.0; return &t }(), // Invalid temperature
			},
		},
		{
			name: "Excessive max tokens",
			req: wrapper.CompletionRequest{
				Prompt:    "Hello",
				MaxTokens: func() *int { m := 1000000; return &m }(), // Excessive tokens
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTesting: %s\n", tc.name)
		_, err := client.Complete(context.Background(), tc.req)
		if err != nil {
			handleSpecificError(err)
		}
	}
}

// handleSpecificError demonstrates handling different error types
func handleSpecificError(err error) {
	if wrapperErr, ok := err.(*wrapper.Error); ok {
		switch wrapperErr.Type {
		case wrapper.ErrorTypeAuth:
			fmt.Printf("Authentication Error: %s\n", wrapperErr.Message)
			fmt.Println("Action: Check your API key and ensure it's valid")

		case wrapper.ErrorTypeRateLimit:
			fmt.Printf("Rate Limit Error: %s\n", wrapperErr.Message)
			if wrapperErr.RetryAfter != nil {
				fmt.Printf("Retry after: %d seconds\n", *wrapperErr.RetryAfter)
			}
			fmt.Println("Action: Wait before retrying or implement exponential backoff")

		case wrapper.ErrorTypeValidation:
			fmt.Printf("Validation Error: %s\n", wrapperErr.Message)
			fmt.Println("Action: Check your request parameters")

		case wrapper.ErrorTypeNetwork:
			fmt.Printf("Network Error: %s\n", wrapperErr.Message)
			fmt.Println("Action: Check your internet connection and retry")

		case wrapper.ErrorTypeProvider:
			fmt.Printf("Provider Error: %s\n", wrapperErr.Message)
			if wrapperErr.Code != "" {
				fmt.Printf("Provider error code: %s\n", wrapperErr.Code)
			}
			fmt.Println("Action: Check provider status or contact support")

		case wrapper.ErrorTypeTokenLimit:
			fmt.Printf("Token Limit Error: %s\n", wrapperErr.Message)
			if wrapperErr.TokenCount != nil {
				fmt.Printf("Token count: %d\n", *wrapperErr.TokenCount)
			}
			fmt.Println("Action: Reduce prompt length or increase max tokens limit")

		default:
			fmt.Printf("Unknown Error Type: %s - %s\n", wrapperErr.Type, wrapperErr.Message)
		}
	} else {
		fmt.Printf("Non-wrapper error: %v\n", err)
	}
}

// retryLogicExample demonstrates implementing retry logic with exponential backoff
func retryLogicExample() {
	// Create client
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey:     "sk-your-openai-api-key-here", // Replace with your actual API key
		MaxRetries: 3,
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Request that might hit rate limits
	req := wrapper.CompletionRequest{
		Prompt:    "Generate a long story about artificial intelligence:",
		MaxTokens: func() *int { m := 1000; return &m }(),
	}

	// Implement custom retry logic with exponential backoff
	maxRetries := 3
	baseDelay := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		fmt.Printf("Attempt %d/%d\n", attempt+1, maxRetries+1)

		resp, err := client.Complete(context.Background(), req)
		if err == nil {
			// Success!
			fmt.Printf("Success! Response: %s...\n", resp.Text[:min(100, len(resp.Text))])
			fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
			return
		}

		// Check if we should retry
		if !shouldRetry(err, attempt, maxRetries) {
			fmt.Printf("Final failure after %d attempts: %v\n", attempt+1, err)
			return
		}

		// Calculate delay with exponential backoff
		delay := calculateBackoffDelay(baseDelay, attempt, err)
		fmt.Printf("Retrying in %v...\n", delay)
		time.Sleep(delay)
	}
}

// shouldRetry determines if an error should be retried
func shouldRetry(err error, attempt, maxRetries int) bool {
	if attempt >= maxRetries {
		return false
	}

	if wrapperErr, ok := err.(*wrapper.Error); ok {
		// Only retry certain error types
		switch wrapperErr.Type {
		case wrapper.ErrorTypeRateLimit, wrapper.ErrorTypeNetwork:
			return true
		case wrapper.ErrorTypeProvider:
			// Retry server errors (5xx) but not client errors (4xx)
			return wrapperErr.Code >= "500"
		default:
			return false
		}
	}

	return false
}

// calculateBackoffDelay calculates the delay for exponential backoff
func calculateBackoffDelay(baseDelay time.Duration, attempt int, err error) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := baseDelay * time.Duration(1<<uint(attempt))

	// Check if error provides specific retry timing
	if wrapperErr, ok := err.(*wrapper.Error); ok {
		if wrapperErr.Type == wrapper.ErrorTypeRateLimit && wrapperErr.RetryAfter != nil {
			// Use the retry-after value from the error
			suggestedDelay := time.Duration(*wrapperErr.RetryAfter) * time.Second
			if suggestedDelay > delay {
				delay = suggestedDelay
			}
		}
	}

	// Cap the maximum delay
	maxDelay := 60 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// handleDifferentErrorTypes demonstrates handling various error scenarios
func handleDifferentErrorTypes() {
	// Test different providers and error scenarios
	testScenarios := []struct {
		name     string
		provider wrapper.ProviderType
		config   wrapper.Config
		request  wrapper.CompletionRequest
	}{
		{
			name:     "Invalid API Key Format",
			provider: wrapper.ProviderOpenAI,
			config: wrapper.Config{
				APIKey: "invalid-key",
			},
			request: wrapper.CompletionRequest{
				Prompt: "Hello",
			},
		},
		{
			name:     "Empty Prompt Validation",
			provider: wrapper.ProviderAnthropic,
			config: wrapper.Config{
				APIKey: "sk-ant-valid-format-but-fake",
			},
			request: wrapper.CompletionRequest{
				Prompt: "", // Empty prompt should fail validation
			},
		},
		{
			name:     "Parameter Out of Range",
			provider: wrapper.ProviderOpenAI,
			config: wrapper.Config{
				APIKey: "sk-valid-format-but-fake",
			},
			request: wrapper.CompletionRequest{
				Prompt:      "Hello",
				Temperature: func() *float64 { t := -1.0; return &t }(), // Invalid temperature
			},
		},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("\n--- Testing: %s ---\n", scenario.name)

		client, err := wrapper.NewClient(scenario.provider, scenario.config)
		if err != nil {
			fmt.Printf("Client creation error: ")
			handleSpecificError(err)
			continue
		}

		_, err = client.Complete(context.Background(), scenario.request)
		if err != nil {
			fmt.Printf("Request error: ")
			handleSpecificError(err)
		} else {
			fmt.Println("Request succeeded unexpectedly")
		}

		client.Close()
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
