// Package main demonstrates chat conversation usage
package main

import (
	"context"
	"fmt"
	"log"

	wrapper "github.com/ajeet-kumar1087/ai-providers"
)

func main() {
	// Example 1: Simple chat conversation
	fmt.Println("=== Simple Chat Conversation ===")
	simpleChatConversation()

	fmt.Println("\n=== Multi-turn Conversation ===")
	multiTurnConversation()

	fmt.Println("\n=== Chat with System Message ===")
	chatWithSystemMessage()
}

// simpleChatConversation demonstrates a basic chat interaction
func simpleChatConversation() {
	// Create client with OpenAI provider
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Create a simple chat request
	req := wrapper.ChatRequest{
		Messages: []wrapper.Message{
			{
				Role:    "user",
				Content: "Hello! Can you help me understand what Go programming language is?",
			},
		},
	}

	// Send the chat request
	resp, err := client.ChatComplete(context.Background(), req)
	if err != nil {
		log.Printf("Chat completion failed: %v", err)
		return
	}

	// Print the response
	fmt.Printf("Assistant: %s\n", resp.Message.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
}

// multiTurnConversation demonstrates a conversation with multiple exchanges
func multiTurnConversation() {
	// Create client with Anthropic provider
	client, err := wrapper.NewClient(wrapper.ProviderAnthropic, wrapper.Config{
		APIKey: "sk-ant-your-anthropic-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Build conversation history
	messages := []wrapper.Message{
		{
			Role:    "user",
			Content: "What's the difference between a slice and an array in Go?",
		},
		{
			Role:    "assistant",
			Content: "In Go, arrays have a fixed size that's part of their type, while slices are dynamic and built on top of arrays. Arrays are value types, but slices are reference types that point to an underlying array.",
		},
		{
			Role:    "user",
			Content: "Can you give me a code example showing this difference?",
		},
	}

	// Create chat request with conversation history
	temperature := 0.3 // Lower temperature for more focused code examples
	req := wrapper.ChatRequest{
		Messages:    messages,
		Temperature: &temperature,
	}

	// Send the chat request
	resp, err := client.ChatComplete(context.Background(), req)
	if err != nil {
		log.Printf("Chat completion failed: %v", err)
		return
	}

	// Print the conversation
	fmt.Println("Conversation:")
	for i, msg := range messages {
		fmt.Printf("%d. %s: %s\n", i+1, msg.Role, msg.Content)
	}
	fmt.Printf("%d. %s: %s\n", len(messages)+1, resp.Message.Role, resp.Message.Content)

	fmt.Printf("\nTokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
}

// chatWithSystemMessage demonstrates using system messages to set context
func chatWithSystemMessage() {
	// Create client with OpenAI provider
	client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
		APIKey: "sk-your-openai-api-key-here", // Replace with your actual API key
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close()

	// Create chat request with system message
	maxTokens := 150
	req := wrapper.ChatRequest{
		Messages: []wrapper.Message{
			{
				Role:    "system",
				Content: "You are a helpful programming tutor who explains concepts clearly and concisely. Always provide practical examples when possible.",
			},
			{
				Role:    "user",
				Content: "How do goroutines work in Go?",
			},
		},
		MaxTokens: &maxTokens,
	}

	// Send the chat request
	resp, err := client.ChatComplete(context.Background(), req)
	if err != nil {
		log.Printf("Chat completion failed: %v", err)
		return
	}

	// Print the response
	fmt.Printf("System context: %s\n", req.Messages[0].Content)
	fmt.Printf("User: %s\n", req.Messages[1].Content)
	fmt.Printf("Assistant: %s\n", resp.Message.Content)
	fmt.Printf("Tokens used: %d (prompt: %d, completion: %d)\n",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
}
