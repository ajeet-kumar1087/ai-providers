package openai

import (
	"testing"

	"github.com/ai-provider-wrapper/ai-provider-wrapper/types"
)

func TestMapCompletionRequest(t *testing.T) {
	config := types.Config{
		APIKey: "sk-test123456789012345678901234567890",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test basic request mapping
	temp := 0.7
	maxTokens := 100
	req := types.CompletionRequest{
		Prompt:      "Hello, world!",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Stop:        []string{".", "!"},
	}

	openaiReq := adapter.mapCompletionRequest(req)

	// Verify mapping
	if openaiReq.Model != DefaultModel {
		t.Errorf("Expected model %s, got %s", DefaultModel, openaiReq.Model)
	}

	if openaiReq.Prompt != req.Prompt {
		t.Errorf("Expected prompt %s, got %s", req.Prompt, openaiReq.Prompt)
	}

	if openaiReq.Temperature == nil || *openaiReq.Temperature != temp {
		t.Errorf("Expected temperature %f, got %v", temp, openaiReq.Temperature)
	}

	if openaiReq.MaxTokens == nil || *openaiReq.MaxTokens != maxTokens {
		t.Errorf("Expected max tokens %d, got %v", maxTokens, openaiReq.MaxTokens)
	}

	if len(openaiReq.Stop) != len(req.Stop) {
		t.Errorf("Expected %d stop sequences, got %d", len(req.Stop), len(openaiReq.Stop))
	}
}

func TestMapCompletionRequestWithClamping(t *testing.T) {
	config := types.Config{
		APIKey: "sk-test123456789012345678901234567890",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test temperature clamping
	temp := 3.0        // Above OpenAI's max of 2.0
	maxTokens := 10000 // Above OpenAI's limit
	req := types.CompletionRequest{
		Prompt:      "Test",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	openaiReq := adapter.mapCompletionRequest(req)

	// Verify clamping
	if openaiReq.Temperature == nil || *openaiReq.Temperature != 2.0 {
		t.Errorf("Expected temperature to be clamped to 2.0, got %v", openaiReq.Temperature)
	}

	if openaiReq.MaxTokens == nil || *openaiReq.MaxTokens != MaxTokenLimit {
		t.Errorf("Expected max tokens to be clamped to %d, got %v", MaxTokenLimit, openaiReq.MaxTokens)
	}
}

func TestNormalizeCompletionResponse(t *testing.T) {
	config := types.Config{
		APIKey: "sk-test123456789012345678901234567890",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test response normalization
	openaiResp := OpenAICompletionResponse{
		ID:      "cmpl-test",
		Object:  "text_completion",
		Created: 1234567890,
		Model:   "gpt-3.5-turbo-instruct",
		Choices: []struct {
			Text         string `json:"text"`
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Text:         "Hello! How can I help you today?",
				Index:        0,
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}

	resp := adapter.normalizeCompletionResponse(openaiResp)

	// Verify normalization
	expectedText := "Hello! How can I help you today?"
	if resp.Text != expectedText {
		t.Errorf("Expected text %s, got %s", expectedText, resp.Text)
	}

	if resp.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %s", resp.FinishReason)
	}

	if resp.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 8 {
		t.Errorf("Expected completion tokens 8, got %d", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 18 {
		t.Errorf("Expected total tokens 18, got %d", resp.Usage.TotalTokens)
	}
}
func TestMapCompletionRequestWithDefaults(t *testing.T) {
	// Test with config defaults
	temp := 0.5
	maxTokens := 200
	config := types.Config{
		APIKey:      "sk-test123456789012345678901234567890",
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Request without temperature/maxTokens should use config defaults
	req := types.CompletionRequest{
		Prompt: "Test prompt",
	}

	openaiReq := adapter.mapCompletionRequest(req)

	// Verify defaults from config are used
	if openaiReq.Temperature == nil || *openaiReq.Temperature != temp {
		t.Errorf("Expected temperature from config %f, got %v", temp, openaiReq.Temperature)
	}

	if openaiReq.MaxTokens == nil || *openaiReq.MaxTokens != maxTokens {
		t.Errorf("Expected max tokens from config %d, got %v", maxTokens, openaiReq.MaxTokens)
	}
}

func TestNormalizeCompletionResponseEmpty(t *testing.T) {
	config := types.Config{
		APIKey: "sk-test123456789012345678901234567890",
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Test response with no choices
	openaiResp := OpenAICompletionResponse{
		ID:      "cmpl-test",
		Object:  "text_completion",
		Created: 1234567890,
		Model:   "gpt-3.5-turbo-instruct",
		Choices: []struct {
			Text         string `json:"text"`
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
		}{},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     5,
			CompletionTokens: 0,
			TotalTokens:      5,
		},
	}

	resp := adapter.normalizeCompletionResponse(openaiResp)

	// Should handle empty choices gracefully
	if resp.Text != "" {
		t.Errorf("Expected empty text for no choices, got %s", resp.Text)
	}

	if resp.FinishReason != "" {
		t.Errorf("Expected empty finish reason for no choices, got %s", resp.FinishReason)
	}

	// Usage should still be populated
	if resp.Usage.TotalTokens != 5 {
		t.Errorf("Expected total tokens 5, got %d", resp.Usage.TotalTokens)
	}
}
