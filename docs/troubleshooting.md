# Troubleshooting Guide

This document provides solutions to common issues and error scenarios when using the AI Provider Wrapper.

## Quick Diagnostics

### Check Your Setup
```bash
# Verify Go version (requires 1.19+)
go version

# Verify module installation
go list -m github.com/ai-provider-wrapper/ai-provider-wrapper

# Check environment variables
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY
```

### Test Basic Connectivity
```go
package main

import (
    "context"
    "fmt"
    "log"
    wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
    // Test with minimal request
    client, err := wrapper.NewClientWithDefaults(wrapper.ProviderOpenAI, "your-api-key")
    if err != nil {
        log.Fatal("Client creation failed:", err)
    }
    defer client.Close()

    resp, err := client.Complete(context.Background(), wrapper.CompletionRequest{
        Prompt: "Hello",
        MaxTokens: &[]int{5}[0],
    })
    if err != nil {
        log.Fatal("Request failed:", err)
    }
    
    fmt.Println("Success:", resp.Text)
}
```

---

## Authentication Issues

### Invalid API Key Format

#### Error Message
```
[openai] authentication (invalid_api_key): Invalid API key provided
```

#### Causes & Solutions

**OpenAI API Key Issues:**
```go
// ❌ Wrong format
apiKey := "your-openai-key"  // Missing sk- prefix

// ✅ Correct format
apiKey := "sk-1234567890abcdef..."  // Must start with sk-
```

**Anthropic API Key Issues:**
```go
// ❌ Wrong format
apiKey := "sk-1234567890abcdef..."  // Wrong prefix

// ✅ Correct format
apiKey := "sk-ant-api03-1234567890abcdef..."  // Must start with sk-ant-
```

#### Validation Function
```go
func validateAPIKey(provider wrapper.ProviderType, apiKey string) error {
    switch provider {
    case wrapper.ProviderOpenAI:
        if !strings.HasPrefix(apiKey, "sk-") {
            return fmt.Errorf("OpenAI API key must start with 'sk-'")
        }
    case wrapper.ProviderAnthropic:
        if !strings.HasPrefix(apiKey, "sk-ant-") {
            return fmt.Errorf("Anthropic API key must start with 'sk-ant-'")
        }
    }
    return nil
}
```

### Missing API Key

#### Error Message
```
[openai] validation: API key is required
```

#### Solutions
```go
// ❌ Empty API key
config := wrapper.Config{}

// ✅ Set API key directly
config := wrapper.Config{
    APIKey: "sk-your-api-key",
}

// ✅ Load from environment
config := wrapper.LoadConfigFromEnv(wrapper.ProviderOpenAI)

// ✅ Use convenience function
client, err := wrapper.NewClientWithDefaults(wrapper.ProviderOpenAI, "sk-your-api-key")
```

### Expired or Invalid Credentials

#### Error Message
```
[openai] authentication (invalid_api_key): Incorrect API key provided
```

#### Solutions
1. **Verify API Key**: Check your provider dashboard
2. **Check Permissions**: Ensure API key has required permissions
3. **Regenerate Key**: Create a new API key if needed
4. **Check Billing**: Ensure account is in good standing

---

## Rate Limiting Issues

### Rate Limit Exceeded

#### Error Message
```
[openai] rate_limit: Request rate limit exceeded. Please try again in 20s
```

#### Automatic Retry Implementation
```go
func makeRequestWithRetry(client wrapper.Client, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    maxRetries := 3
    baseDelay := time.Second

    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := client.Complete(context.Background(), req)
        if err == nil {
            return resp, nil
        }

        if wrapperErr, ok := err.(*wrapper.Error); ok && wrapperErr.Type == wrapper.ErrorTypeRateLimit {
            if attempt == maxRetries {
                return nil, fmt.Errorf("max retries exceeded: %w", err)
            }

            // Use suggested retry delay or exponential backoff
            delay := baseDelay * time.Duration(1<<attempt)
            if wrapperErr.RetryAfter != nil {
                delay = time.Duration(*wrapperErr.RetryAfter) * time.Second
            }

            log.Printf("Rate limited, retrying in %v (attempt %d/%d)", delay, attempt+1, maxRetries)
            time.Sleep(delay)
            continue
        }

        // Non-retryable error
        return nil, err
    }

    return nil, fmt.Errorf("unexpected retry loop exit")
}
```

#### Rate Limiting Best Practices
```go
// 1. Implement client-side rate limiting
type RateLimitedClient struct {
    client wrapper.Client
    limiter *rate.Limiter
}

func NewRateLimitedClient(client wrapper.Client, requestsPerSecond float64) *RateLimitedClient {
    return &RateLimitedClient{
        client:  client,
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (r *RateLimitedClient) Complete(ctx context.Context, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    if err := r.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    return r.client.Complete(ctx, req)
}

// 2. Batch requests when possible
func batchRequests(client wrapper.Client, prompts []string) []wrapper.CompletionResponse {
    results := make([]wrapper.CompletionResponse, len(prompts))
    
    // Process in smaller batches
    batchSize := 5
    for i := 0; i < len(prompts); i += batchSize {
        end := i + batchSize
        if end > len(prompts) {
            end = len(prompts)
        }
        
        // Process batch with delay
        for j := i; j < end; j++ {
            resp, err := client.Complete(context.Background(), wrapper.CompletionRequest{
                Prompt: prompts[j],
            })
            if err != nil {
                log.Printf("Request %d failed: %v", j, err)
                continue
            }
            results[j] = *resp
        }
        
        // Delay between batches
        if end < len(prompts) {
            time.Sleep(time.Second)
        }
    }
    
    return results
}
```

---

## Network Issues

### Connection Timeouts

#### Error Message
```
[openai] network: context deadline exceeded
```

#### Solutions
```go
// 1. Increase timeout
config := wrapper.DefaultConfig().
    WithAPIKey("your-api-key").
    WithTimeout(60 * time.Second)  // Increase from default 30s

// 2. Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

resp, err := client.Complete(ctx, request)

// 3. Implement retry with exponential backoff
func makeRequestWithNetworkRetry(client wrapper.Client, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    maxRetries := 3
    baseDelay := 2 * time.Second

    for attempt := 0; attempt <= maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
        resp, err := client.Complete(ctx, req)
        cancel()

        if err == nil {
            return resp, nil
        }

        if wrapperErr, ok := err.(*wrapper.Error); ok && wrapperErr.Type == wrapper.ErrorTypeNetwork {
            if attempt < maxRetries {
                delay := baseDelay * time.Duration(1<<attempt)
                log.Printf("Network error, retrying in %v", delay)
                time.Sleep(delay)
                continue
            }
        }

        return nil, err
    }

    return nil, fmt.Errorf("max retries exceeded")
}
```

### DNS Resolution Issues

#### Error Message
```
[openai] network: no such host
```

#### Solutions
1. **Check Internet Connection**: Verify basic connectivity
2. **DNS Configuration**: Check DNS settings
3. **Firewall Rules**: Ensure outbound HTTPS is allowed
4. **Proxy Configuration**: Set up proxy if required

```go
// Custom HTTP client with proxy
import (
    "net/http"
    "net/url"
)

func createClientWithProxy(proxyURL string) wrapper.Client {
    proxy, _ := url.Parse(proxyURL)
    transport := &http.Transport{
        Proxy: http.ProxyURL(proxy),
    }
    
    // Note: This requires custom HTTP client support in adapters
    // Currently not directly supported, but can be added
}
```

---

## Validation Errors

### Parameter Out of Range

#### Error Message
```
[openai] validation: temperature must be between 0.0 and 2.0, got: 3.0
```

#### Solutions
```go
// ❌ Invalid parameter
req := wrapper.CompletionRequest{
    Prompt: "Hello",
    Temperature: &[]float64{3.0}[0],  // Too high
}

// ✅ Valid parameter
req := wrapper.CompletionRequest{
    Prompt: "Hello",
    Temperature: &[]float64{0.8}[0],  // Within range
}

// ✅ Parameter validation function
func validateTemperature(temp *float64, provider wrapper.ProviderType) error {
    if temp == nil {
        return nil
    }
    
    var maxTemp float64
    switch provider {
    case wrapper.ProviderOpenAI:
        maxTemp = 2.0
    case wrapper.ProviderAnthropic:
        maxTemp = 1.0
    default:
        maxTemp = 1.0
    }
    
    if *temp < 0.0 || *temp > maxTemp {
        return fmt.Errorf("temperature must be between 0.0 and %.1f for %s", maxTemp, provider)
    }
    
    return nil
}
```

### Token Limit Exceeded

#### Error Message
```
[anthropic] token_limit: Maximum context length exceeded (150000 tokens)
```

#### Solutions
```go
// 1. Reduce prompt length
func truncatePrompt(prompt string, maxTokens int) string {
    // Rough estimation: 1 token ≈ 4 characters
    maxChars := maxTokens * 4
    if len(prompt) <= maxChars {
        return prompt
    }
    return prompt[:maxChars] + "..."
}

// 2. Split long requests
func splitLongRequest(prompt string, maxTokens int) []string {
    chunks := []string{}
    maxChars := maxTokens * 4
    
    for i := 0; i < len(prompt); i += maxChars {
        end := i + maxChars
        if end > len(prompt) {
            end = len(prompt)
        }
        chunks = append(chunks, prompt[i:end])
    }
    
    return chunks
}

// 3. Use streaming for long responses (when available)
req := wrapper.CompletionRequest{
    Prompt: "Long prompt...",
    MaxTokens: &[]int{1000}[0],  // Limit response length
    Stream: true,  // Enable streaming (future feature)
}
```

### Empty or Invalid Prompt

#### Error Message
```
[openai] validation: prompt is required
```

#### Solutions
```go
// ❌ Empty prompt
req := wrapper.CompletionRequest{
    Prompt: "",  // Empty
}

// ❌ Whitespace only
req := wrapper.CompletionRequest{
    Prompt: "   ",  // Only whitespace
}

// ✅ Valid prompt
req := wrapper.CompletionRequest{
    Prompt: "Write a haiku about programming",
}

// ✅ Prompt validation
func validatePrompt(prompt string) error {
    if strings.TrimSpace(prompt) == "" {
        return fmt.Errorf("prompt cannot be empty")
    }
    return nil
}
```

---

## Provider-Specific Issues

### OpenAI Issues

#### Model Not Found
```
[openai] provider (model_not_found): The model 'gpt-5' does not exist
```

**Solution**: Use supported model names:
```go
// ❌ Invalid model (models are set in adapter, not directly configurable yet)
// This will be addressed in future versions

// ✅ Current approach - models are handled automatically by adapters
client, err := wrapper.NewClient(wrapper.ProviderOpenAI, config)
```

#### Quota Exceeded
```
[openai] provider (insufficient_quota): You exceeded your current quota
```

**Solutions**:
1. Check billing in OpenAI dashboard
2. Upgrade plan or add credits
3. Implement usage monitoring

### Anthropic Issues

#### Overloaded Service
```
[anthropic] provider (overloaded_error): The service is temporarily overloaded
```

**Solution**: Implement retry with longer delays:
```go
func retryWithBackoff(client wrapper.Client, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    delays := []time.Duration{1*time.Second, 5*time.Second, 15*time.Second, 30*time.Second}
    
    for i, delay := range delays {
        resp, err := client.Complete(context.Background(), req)
        if err == nil {
            return resp, nil
        }
        
        if wrapperErr, ok := err.(*wrapper.Error); ok {
            if wrapperErr.Type == wrapper.ErrorTypeProvider && strings.Contains(wrapperErr.Message, "overloaded") {
                if i < len(delays)-1 {
                    log.Printf("Service overloaded, retrying in %v", delay)
                    time.Sleep(delay)
                    continue
                }
            }
        }
        
        return nil, err
    }
    
    return nil, fmt.Errorf("service remained overloaded after all retries")
}
```

---

## Performance Issues

### Slow Response Times

#### Diagnosis
```go
func benchmarkProvider(client wrapper.Client, prompt string) {
    start := time.Now()
    
    resp, err := client.Complete(context.Background(), wrapper.CompletionRequest{
        Prompt: prompt,
        MaxTokens: &[]int{100}[0],
    })
    
    duration := time.Since(start)
    
    if err != nil {
        log.Printf("Request failed after %v: %v", duration, err)
        return
    }
    
    log.Printf("Request completed in %v, tokens: %d", duration, resp.Usage.TotalTokens)
}
```

#### Solutions
1. **Reduce Max Tokens**: Lower token limits for faster responses
2. **Optimize Prompts**: Use shorter, more specific prompts
3. **Use Faster Models**: Choose speed-optimized models when available
4. **Implement Caching**: Cache responses for repeated requests

```go
// Simple response cache
type ResponseCache struct {
    cache map[string]*wrapper.CompletionResponse
    mutex sync.RWMutex
}

func (c *ResponseCache) Get(prompt string) (*wrapper.CompletionResponse, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    resp, exists := c.cache[prompt]
    return resp, exists
}

func (c *ResponseCache) Set(prompt string, resp *wrapper.CompletionResponse) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.cache[prompt] = resp
}
```

### Memory Usage Issues

#### High Memory Consumption
```go
// ❌ Creating many clients
for i := 0; i < 1000; i++ {
    client, _ := wrapper.NewClient(wrapper.ProviderOpenAI, config)
    // ... use client
    client.Close()
}

// ✅ Reuse single client
client, err := wrapper.NewClient(wrapper.ProviderOpenAI, config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

for i := 0; i < 1000; i++ {
    // ... use same client for multiple requests
}
```

---

## Debugging Tips

### Enable Verbose Logging
```go
import "log"

// Set up detailed logging
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Log all requests and responses
func loggedRequest(client wrapper.Client, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    log.Printf("Sending request: %+v", req)
    
    resp, err := client.Complete(context.Background(), req)
    
    if err != nil {
        log.Printf("Request failed: %v", err)
        return nil, err
    }
    
    log.Printf("Request succeeded: %d tokens used", resp.Usage.TotalTokens)
    return resp, nil
}
```

### Inspect Error Details
```go
func inspectError(err error) {
    if wrapperErr, ok := err.(*wrapper.Error); ok {
        log.Printf("Error Type: %s", wrapperErr.Type)
        log.Printf("Provider: %s", wrapperErr.Provider)
        log.Printf("Message: %s", wrapperErr.Message)
        log.Printf("Code: %s", wrapperErr.Code)
        
        if wrapperErr.RetryAfter != nil {
            log.Printf("Retry After: %d seconds", *wrapperErr.RetryAfter)
        }
        
        if wrapperErr.TokenCount != nil {
            log.Printf("Token Count: %d", *wrapperErr.TokenCount)
        }
        
        if wrapperErr.Wrapped != nil {
            log.Printf("Underlying Error: %v", wrapperErr.Wrapped)
        }
    }
}
```

### Test with Minimal Examples
```go
// Minimal test for each provider
func testProvider(provider wrapper.ProviderType, apiKey string) {
    client, err := wrapper.NewClientWithDefaults(provider, apiKey)
    if err != nil {
        log.Printf("%s client creation failed: %v", provider, err)
        return
    }
    defer client.Close()
    
    resp, err := client.Complete(context.Background(), wrapper.CompletionRequest{
        Prompt: "Say hello",
        MaxTokens: &[]int{5}[0],
    })
    
    if err != nil {
        log.Printf("%s request failed: %v", provider, err)
        return
    }
    
    log.Printf("%s success: %s", provider, resp.Text)
}
```

---

## Getting Help

### Before Asking for Help
1. **Check this troubleshooting guide**
2. **Review the [Provider Documentation](providers.md)**
3. **Look at [Examples](../examples/)**
4. **Search existing GitHub issues**

### When Reporting Issues
Include the following information:

```go
// Version information
go version
go list -m github.com/ai-provider-wrapper/ai-provider-wrapper

// Minimal reproduction case
package main

import (
    "context"
    "log"
    wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
    // Your minimal failing code here
    client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
        APIKey: "sk-test...", // Redacted API key
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // The specific request that fails
    _, err = client.Complete(context.Background(), wrapper.CompletionRequest{
        Prompt: "test prompt",
    })
    if err != nil {
        log.Fatal(err) // Include full error output
    }
}
```

### Support Channels
1. **GitHub Issues**: For bugs and feature requests
2. **Discussions**: For questions and community help
3. **Documentation**: Check API docs and examples
4. **Provider Support**: For provider-specific API issues

### Emergency Debugging
```bash
# Quick environment check
echo "Go version: $(go version)"
echo "Module path: $(go list -m)"
echo "OpenAI key set: $([ -n "$OPENAI_API_KEY" ] && echo "Yes" || echo "No")"
echo "Anthropic key set: $([ -n "$ANTHROPIC_API_KEY" ] && echo "Yes" || echo "No")"

# Test basic connectivity
curl -s https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY" | head -n 5
```