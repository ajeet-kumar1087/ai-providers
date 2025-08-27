# AI Provider Wrapper

A Go package that provides a unified interface for interacting with multiple AI providers (OpenAI, Anthropic, Google). This package abstracts away provider-specific implementations and offers a consistent API for developers to integrate AI capabilities into their applications without being locked into a single provider.

[![Go Reference](https://pkg.go.dev/badge/github.com/ajeet-kumar1087/ai-providers.svg)](https://pkg.go.dev/github.com/ajeet-kumar1087/ai-providers)
[![Go Report Card](https://goreportcard.com/badge/github.com/ajeet-kumar1087/ai-providers)](https://goreportcard.com/report/github.com/ajeet-kumar1087/ai-providers)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **🔄 Unified Interface**: Single API for multiple AI providers
- **🔀 Provider Abstraction**: Switch between providers without code changes
- **⚡ Consistent Error Handling**: Standardized error types across all providers
- **🎛️ Parameter Mapping**: Automatic translation between generic and provider-specific parameters
- **📦 Minimal Dependencies**: Lightweight design using standard library components
- **🚀 Easy Installation**: Simple `go get` installation with Go modules support
- **🔧 Flexible Configuration**: Environment variables, defaults, and custom configurations
- **🛡️ Robust Error Handling**: Comprehensive error categorization and retry logic
- **📊 Usage Tracking**: Token usage statistics for cost monitoring

## Supported Providers

| Provider | Status | Models | Features |
|----------|--------|--------|----------|
| **OpenAI** | ✅ Ready | GPT-3.5, GPT-4, GPT-4 Turbo | Text completion, Chat completion |
| **Anthropic** | ✅ Ready | Claude-3, Claude-2, Claude Instant | Text completion, Chat completion |
| **Google AI** | 🚧 In Progress | Gemini Pro, Gemini Pro Vision | Coming soon |

## Installation

### Prerequisites

- Go 1.19 or later
- Valid API key for at least one supported provider

### Install the Package

```bash
go get github.com/ajeet-kumar1087/ai-providers
```

### Verify Installation

```bash
go mod tidy
```

## Quick Start

### Basic Text Completion

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
    // Create client with OpenAI
    client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
        APIKey: "sk-your-openai-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send completion request
    resp, err := client.Complete(context.Background(), wrapper.CompletionRequest{
        Prompt: "Write a haiku about programming:",
        Temperature: &[]float64{0.7}[0],
        MaxTokens: &[]int{100}[0],
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Generated text: %s\n", resp.Text)
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

### Chat Completion

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
    // Create client with Anthropic
    client, err := wrapper.NewClient(wrapper.ProviderAnthropic, wrapper.Config{
        APIKey: "sk-ant-your-anthropic-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send chat request
    resp, err := client.ChatComplete(context.Background(), wrapper.ChatRequest{
        Messages: []wrapper.Message{
            {Role: "system", Content: "You are a helpful assistant."},
            {Role: "user", Content: "Explain quantum computing in simple terms."},
        },
        Temperature: &[]float64{0.5}[0],
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Assistant: %s\n", resp.Message.Content)
}
```

## Configuration

### Environment Variables

The package supports configuration through environment variables:

```bash
# Provider-specific API keys
export OPENAI_API_KEY="sk-your-openai-key"
export ANTHROPIC_API_KEY="sk-ant-your-anthropic-key"
export GOOGLE_API_KEY="your-google-api-key"

# Optional: Custom endpoints
export OPENAI_BASE_URL="https://api.openai.com/v1"
export ANTHROPIC_BASE_URL="https://api.anthropic.com"

# Optional: Global settings
export AI_TIMEOUT="30s"
export AI_MAX_RETRIES="3"
export AI_TEMPERATURE="0.7"
export AI_MAX_TOKENS="1000"
```

Then use the convenience function:

```go
client, err := wrapper.NewClientWithEnvConfig(wrapper.ProviderOpenAI)
```

### Configuration Builder Pattern

```go
config := wrapper.DefaultConfig().
    WithAPIKey("your-api-key").
    WithTimeout(60 * time.Second).
    WithMaxRetries(5).
    WithTemperature(0.8).
    WithMaxTokens(2000)

client, err := wrapper.NewClient(wrapper.ProviderOpenAI, config)
```

### Quick Setup with Defaults

```go
client, err := wrapper.NewClientWithDefaults(wrapper.ProviderOpenAI, "sk-your-api-key")
```

## Advanced Usage

### Provider Switching

```go
providers := []wrapper.ProviderType{
    wrapper.ProviderOpenAI,
    wrapper.ProviderAnthropic,
}

for _, provider := range providers {
    client, err := wrapper.NewClientWithEnvConfig(provider)
    if err != nil {
        log.Printf("Failed to create %s client: %v", provider, err)
        continue
    }
    
    resp, err := client.Complete(ctx, request)
    if err != nil {
        log.Printf("Request failed with %s: %v", provider, err)
        client.Close()
        continue
    }
    
    fmt.Printf("Success with %s: %s\n", provider, resp.Text)
    client.Close()
    break
}
```

### Error Handling with Retry Logic

```go
func makeRequestWithRetry(client wrapper.Client, req wrapper.CompletionRequest) (*wrapper.CompletionResponse, error) {
    maxRetries := 3
    baseDelay := time.Second
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        resp, err := client.Complete(context.Background(), req)
        if err == nil {
            return resp, nil
        }
        
        // Check if error is retryable
        if wrapperErr, ok := err.(*wrapper.Error); ok {
            switch wrapperErr.Type {
            case wrapper.ErrorTypeRateLimit:
                // Use suggested retry delay if available
                delay := baseDelay * time.Duration(1<<attempt)
                if wrapperErr.RetryAfter != nil {
                    delay = time.Duration(*wrapperErr.RetryAfter) * time.Second
                }
                log.Printf("Rate limited, retrying in %v", delay)
                time.Sleep(delay)
                continue
                
            case wrapper.ErrorTypeNetwork:
                // Exponential backoff for network errors
                delay := baseDelay * time.Duration(1<<attempt)
                log.Printf("Network error, retrying in %v", delay)
                time.Sleep(delay)
                continue
                
            case wrapper.ErrorTypeAuth:
                // Don't retry authentication errors
                return nil, fmt.Errorf("authentication failed: %w", err)
                
            case wrapper.ErrorTypeValidation:
                // Don't retry validation errors
                return nil, fmt.Errorf("invalid request: %w", err)
            }
        }
        
        // Unknown error type, don't retry
        return nil, err
    }
    
    return nil, fmt.Errorf("max retries exceeded")
}
```

### Multi-turn Conversation

```go
func chatConversation(client wrapper.Client) {
    messages := []wrapper.Message{
        {Role: "system", Content: "You are a helpful coding assistant."},
    }
    
    scanner := bufio.NewScanner(os.Stdin)
    
    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        
        userInput := scanner.Text()
        if userInput == "quit" {
            break
        }
        
        // Add user message
        messages = append(messages, wrapper.Message{
            Role: "user",
            Content: userInput,
        })
        
        // Get assistant response
        resp, err := client.ChatComplete(context.Background(), wrapper.ChatRequest{
            Messages: messages,
            Temperature: &[]float64{0.7}[0],
        })
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        // Add assistant response to conversation
        messages = append(messages, resp.Message)
        
        fmt.Printf("Assistant: %s\n\n", resp.Message.Content)
    }
}
```

## Error Handling

The package provides comprehensive error categorization:

```go
if err != nil {
    if wrapperErr, ok := err.(*wrapper.Error); ok {
        switch wrapperErr.Type {
        case wrapper.ErrorTypeAuth:
            // Invalid API key or authentication failure
            log.Printf("Authentication error: %s", wrapperErr.Message)
            
        case wrapper.ErrorTypeRateLimit:
            // Rate limit exceeded
            if wrapperErr.RetryAfter != nil {
                log.Printf("Rate limited, retry after %d seconds", *wrapperErr.RetryAfter)
            }
            
        case wrapper.ErrorTypeNetwork:
            // Network connectivity issues
            log.Printf("Network error: %s", wrapperErr.Message)
            
        case wrapper.ErrorTypeValidation:
            // Invalid request parameters
            log.Printf("Validation error: %s", wrapperErr.Message)
            
        case wrapper.ErrorTypeProvider:
            // Provider-side errors (5xx responses)
            log.Printf("Provider error: %s", wrapperErr.Message)
            
        case wrapper.ErrorTypeTokenLimit:
            // Token limit exceeded
            if wrapperErr.TokenCount != nil {
                log.Printf("Token limit exceeded: %d tokens", *wrapperErr.TokenCount)
            }
        }
        
        // Check if error is retryable
        if wrapperErr.IsRetryable() {
            log.Println("This error can be retried")
        }
    }
}
```

## Provider Capabilities

### OpenAI
- **Models**: GPT-3.5-turbo, GPT-4, GPT-4-turbo
- **Max Tokens**: Up to 4,096 (varies by model)
- **Temperature Range**: 0.0 - 2.0
- **Special Features**: Function calling (future), JSON mode (future)

### Anthropic
- **Models**: Claude-3 (Haiku, Sonnet, Opus), Claude-2
- **Max Tokens**: Up to 100,000
- **Temperature Range**: 0.0 - 1.0
- **Special Features**: Large context windows, constitutional AI

### Google AI (Coming Soon)
- **Models**: Gemini Pro, Gemini Pro Vision
- **Max Tokens**: Up to 8,192
- **Temperature Range**: 0.0 - 1.0
- **Special Features**: Multimodal input (future)

## Examples

The [examples](examples/) directory contains comprehensive usage examples:

- **[Basic Usage](examples/basic/)** - Simple completion requests
- **[Chat Conversations](examples/chat/)** - Multi-turn conversations
- **[Configuration](examples/config/)** - Advanced configuration options
- **[Error Handling](examples/errors/)** - Robust error handling patterns
- **[Parameters](examples/parameters/)** - Parameter customization
- **[Provider Switching](examples/provider-switching/)** - Working with multiple providers

Run any example:

```bash
go run ./examples/basic/
go run ./examples/chat/
go run ./examples/config/
```

## API Reference

### Core Types

```go
// Client interface
type Client interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    ChatComplete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Close() error
}

// Request types
type CompletionRequest struct {
    Prompt      string
    Temperature *float64
    MaxTokens   *int
    Stop        []string
    Stream      bool
}

type ChatRequest struct {
    Messages    []Message
    Temperature *float64
    MaxTokens   *int
    Stream      bool
}

// Response types
type CompletionResponse struct {
    Text         string
    Usage        Usage
    FinishReason string
}

type ChatResponse struct {
    Message      Message
    Usage        Usage
    FinishReason string
}
```

### Configuration

```go
type Config struct {
    APIKey      string
    BaseURL     string
    Timeout     time.Duration
    MaxRetries  int
    Temperature *float64
    MaxTokens   *int
}
```

### Error Types

```go
type Error struct {
    Type       ErrorType
    Message    string
    Code       string
    Provider   string
    RetryAfter *int
    TokenCount *int
}

const (
    ErrorTypeAuth       ErrorType = "authentication"
    ErrorTypeRateLimit  ErrorType = "rate_limit"
    ErrorTypeNetwork    ErrorType = "network"
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeProvider   ErrorType = "provider"
    ErrorTypeTokenLimit ErrorType = "token_limit"
)
```

## Troubleshooting

### Common Issues

#### Authentication Errors
```
Error: [openai] authentication: Invalid API key provided
```
**Solution**: Verify your API key is correct and has the proper format:
- OpenAI: starts with `sk-`
- Anthropic: starts with `sk-ant-`

#### Rate Limiting
```
Error: [openai] rate_limit: Request rate limit exceeded
```
**Solution**: Implement retry logic with exponential backoff or reduce request frequency.

#### Network Timeouts
```
Error: [openai] network: context deadline exceeded
```
**Solution**: Increase timeout in configuration:
```go
config := wrapper.DefaultConfig().WithTimeout(60 * time.Second)
```

#### Token Limits
```
Error: [anthropic] token_limit: Maximum context length exceeded
```
**Solution**: Reduce prompt length or max_tokens parameter.

### Getting Help

1. **Check the [Troubleshooting Guide](docs/troubleshooting.md)**
2. **Review [Provider Documentation](docs/providers.md)**
3. **Look at [Examples](examples/)**
4. **Check [API Documentation](https://pkg.go.dev/github.com/ai-provider-wrapper/ai-provider-wrapper)**
5. **Open an issue on GitHub**

## Performance Tips

1. **Reuse Clients**: Create clients once and reuse them for multiple requests
2. **Set Appropriate Timeouts**: Balance responsiveness with reliability
3. **Use Context Cancellation**: Implement proper context handling for request cancellation
4. **Monitor Token Usage**: Track usage statistics to optimize costs
5. **Implement Caching**: Cache responses for repeated requests when appropriate

## Security Considerations

1. **API Key Management**: Never commit API keys to version control
2. **Environment Variables**: Use environment variables or secure secret management
3. **Network Security**: Use HTTPS endpoints (default)
4. **Input Validation**: Validate user inputs before sending to AI providers
5. **Rate Limiting**: Implement client-side rate limiting to prevent abuse

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
git clone https://github.com/ai-provider-wrapper/ai-provider-wrapper.git
cd ai-provider-wrapper
go mod download
go test ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./adapters/openai/
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

---

**Made with ❤️ for the Go community**