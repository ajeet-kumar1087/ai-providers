// Package main demonstrates parameter customization across providers
package main

import (
	"context"
	"fmt"
	"log"

	wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
	// Example 1: Temperature variations
	fmt.Println("=== Temperature Parameter Examples ===")
	temperatureExamples()

	fmt.Println("\n=== Max Tokens Parameter Examples ===")
	maxTokensExamples()

	fmt.Println("\n=== Stop Sequences Examples ===")
	stopSequencesExamples()

	fmt.Println("\n=== Cross-Provider Parameter Compatibility ===")
	crossProviderCompatibility()
}

// temperatureExamples demonstrates different temperature settings
func temperatureExamples() {
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	prompt := "Write a creative opening line for a story:"
	temperatures := []float64{0.0, 0.5, 1.0, 1.5}

	for _, temp := range temperatures {
		fmt.Printf("\n--- Temperature: %.1f ---\n", temp)

		req := wrapper.CompletionRequest{
			Prompt:      prompt,
			Temperature: &temp,
			MaxTokens:   func() *int { m := 50; return &m }(),
		}

		resp, err := client.Complete(context.Background(), req)
		if err != nil {
			log.Printf("Completion failed with temperature %.1f: %v", temp, err)
			continue
		}

		fmt.Printf("Response: %s\n", resp.Text)
		fmt.Printf("Tokens: %d\n", resp.Usage.TotalTokens)
	}
}

// maxTokensExamples demonstrates different token limits
func maxTokensExamples() {
	client, err := wrapper.NewClient(wrapper.ProviderAnthropic, wrapper.Config{
		APIKey: "sk-ant-your-anthropic-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	prompt := "Explain the concept of machine learning in detail:"
	tokenLimits := []int{50, 100, 200, 500}

	for _, limit := range tokenLimits {
		fmt.Printf("\n--- Max Tokens: %d ---\n", limit)

		req := wrapper.CompletionRequest{
			Prompt:    prompt,
			MaxTokens: &limit,
		}

		resp, err := client.Complete(context.Background(), req)
		if err != nil {
			log.Printf("Completion failed with max tokens %d: %v", limit, err)
			continue
		}

		fmt.Printf("Response length: %d characters\n", len(resp.Text))
		fmt.Printf("Tokens used: %d/%d\n", resp.Usage.CompletionTokens, limit)
		fmt.Printf("Finish reason: %s\n", resp.FinishReason)
		fmt.Printf("Response preview: %s...\n", resp.Text[:min(100, len(resp.Text))])
	}
}

// stopSequencesExamples demonstrates using stop sequences
func stopSequencesExamples() {
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Example 1: Stop at specific punctuation
	fmt.Println("\n--- Stop at period ---")
	req1 := wrapper.CompletionRequest{
		Prompt:    "List three benefits of exercise: 1.",
		MaxTokens: func() *int { m := 100; return &m }(),
		Stop:      []string{"."},
	}

	resp1, err := client.Complete(context.Background(), req1)
	if err != nil {
		log.Printf("Completion failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", resp1.Text)
		fmt.Printf("Finish reason: %s\n", resp1.FinishReason)
	}

	// Example 2: Stop at multiple sequences
	fmt.Println("\n--- Stop at multiple sequences ---")
	req2 := wrapper.CompletionRequest{
		Prompt:    "Write a dialogue:\nAlice: Hello!\nBob:",
		MaxTokens: func() *int { m := 100; return &m }(),
		Stop:      []string{"\nAlice:", "\nBob:", "\n\n"},
	}

	resp2, err := client.Complete(context.Background(), req2)
	if err != nil {
		log.Printf("Completion failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", resp2.Text)
		fmt.Printf("Finish reason: %s\n", resp2.FinishReason)
	}

	// Example 3: Stop at custom markers
	fmt.Println("\n--- Stop at custom markers ---")
	req3 := wrapper.CompletionRequest{
		Prompt:    "Generate code with comments:\n```python\n# Function to calculate factorial\ndef factorial(n):",
		MaxTokens: func() *int { m := 150; return &m }(),
		Stop:      []string{"```", "# End", "\n\n\n"},
	}

	resp3, err := client.Complete(context.Background(), req3)
	if err != nil {
		log.Printf("Completion failed: %v", err)
	} else {
		fmt.Printf("Response: %s\n", resp3.Text)
		fmt.Printf("Finish reason: %s\n", resp3.FinishReason)
	}
}

// crossProviderCompatibility demonstrates parameter handling across providers
func crossProviderCompatibility() {
	providers := []struct {
		name     string
		provider wrapper.ProviderType
		apiKey   string
	}{
		{"OpenAI", wrapper.ProviderOpenAI, "sk-your-openai-api-key-here"},
		{"Anthropic", wrapper.ProviderAnthropic, "sk-ant-your-anthropic-api-key-here"},
	}

	// Test the same parameters across different providers
	testParams := struct {
		prompt      string
		temperature float64
		maxTokens   int
		stop        []string
	}{
		prompt:      "Explain quantum computing in simple terms:",
		temperature: 0.7,
		maxTokens:   150,
		stop:        []string{"\n\n", "In conclusion"},
	}

	for _, p := range providers {
		fmt.Printf("\n--- Testing %s with standard parameters ---\n", p.name)

		client, err := wrapper.NewClient(p.provider, wrapper.Config{
			APIKey: p.apiKey,
		})
		if err != nil {
			log.Printf("Failed to create %s client: %v", p.name, err)
			continue
		}

		req := wrapper.CompletionRequest{
			Prompt:      testParams.prompt,
			Temperature: &testParams.temperature,
			MaxTokens:   &testParams.maxTokens,
			Stop:        testParams.stop,
		}

		resp, err := client.Complete(context.Background(), req)
		if err != nil {
			log.Printf("%s completion failed: %v", p.name, err)
			client.Close()
			continue
		}

		fmt.Printf("Response: %s\n", resp.Text)
		fmt.Printf("Tokens: %d/%d (prompt: %d, completion: %d)\n",
			resp.Usage.TotalTokens, testParams.maxTokens,
			resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		fmt.Printf("Finish reason: %s\n", resp.FinishReason)

		client.Close()
	}
}

// demonstrateParameterClamping shows how parameters are clamped to provider limits
func demonstrateParameterClamping() {
	fmt.Println("\n=== Parameter Clamping Examples ===")

	// Test with parameters that exceed provider limits
	testCases := []struct {
		name        string
		provider    wrapper.ProviderType
		apiKey      string
		temperature float64
		maxTokens   int
	}{
		{
			name:        "OpenAI with high temperature",
			provider:    wrapper.ProviderOpenAI,
			apiKey:      "sk-your-openai-api-key-here",
			temperature: 3.0, // Will be clamped to 2.0 for OpenAI
			maxTokens:   100,
		},
		{
			name:        "Anthropic with high temperature",
			provider:    wrapper.ProviderAnthropic,
			apiKey:      "sk-ant-your-anthropic-api-key-here",
			temperature: 2.0, // Will be clamped to 1.0 for Anthropic
			maxTokens:   100,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)

		client, err := wrapper.NewClient(tc.provider, wrapper.Config{
			APIKey: tc.apiKey,
		})
		if err != nil {
			log.Printf("Failed to create client: %v", err)
			continue
		}

		fmt.Printf("Requested temperature: %.1f\n", tc.temperature)

		req := wrapper.CompletionRequest{
			Prompt:      "Write a short poem:",
			Temperature: &tc.temperature,
			MaxTokens:   &tc.maxTokens,
		}

		resp, err := client.Complete(context.Background(), req)
		if err != nil {
			log.Printf("Completion failed: %v", err)
			client.Close()
			continue
		}

		fmt.Printf("Response: %s\n", resp.Text)
		fmt.Printf("Note: Temperature was automatically clamped to provider limits\n")

		client.Close()
	}
}

// demonstrateChatParameterCustomization shows parameter customization for chat
func demonstrateChatParameterCustomization() {
	fmt.Println("\n=== Chat Parameter Customization ===")

	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-your-openai-api-key-here",
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Different parameter combinations for chat
	scenarios := []struct {
		name        string
		temperature float64
		maxTokens   int
		messages    []wrapper.Message
	}{
		{
			name:        "Creative storytelling (high temperature)",
			temperature: 1.2,
			maxTokens:   200,
			messages: []wrapper.Message{
				{Role: "system", Content: "You are a creative storyteller."},
				{Role: "user", Content: "Tell me a unique story about a robot learning to paint."},
			},
		},
		{
			name:        "Technical explanation (low temperature)",
			temperature: 0.2,
			maxTokens:   150,
			messages: []wrapper.Message{
				{Role: "system", Content: "You are a technical expert who gives precise explanations."},
				{Role: "user", Content: "How does a hash table work?"},
			},
		},
		{
			name:        "Conversational (medium temperature)",
			temperature: 0.7,
			maxTokens:   100,
			messages: []wrapper.Message{
				{Role: "user", Content: "What's your favorite programming language and why?"},
			},
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n--- %s ---\n", scenario.name)
		fmt.Printf("Temperature: %.1f, Max Tokens: %d\n", scenario.temperature, scenario.maxTokens)

		req := wrapper.ChatRequest{
			Messages:    scenario.messages,
			Temperature: &scenario.temperature,
			MaxTokens:   &scenario.maxTokens,
		}

		resp, err := client.ChatComplete(context.Background(), req)
		if err != nil {
			log.Printf("Chat completion failed: %v", err)
			continue
		}

		fmt.Printf("Response: %s\n", resp.Message.Content)
		fmt.Printf("Tokens: %d/%d\n", resp.Usage.CompletionTokens, scenario.maxTokens)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
