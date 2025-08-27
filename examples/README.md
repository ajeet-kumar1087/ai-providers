# AI Provider Wrapper Examples

This directory contains comprehensive examples demonstrating how to use the AI Provider Wrapper library. Each example is self-contained and includes mock implementations for demonstration purposes.

## Examples Overview

### Basic Usage (`basic/`)
- **Purpose**: Demonstrates fundamental completion operations
- **Features**: 
  - Simple text completion with OpenAI and Anthropic
  - Basic parameter usage (temperature, max tokens, stop sequences)
  - Client initialization and cleanup

### Chat Conversations (`chat/`)
- **Purpose**: Shows chat completion functionality
- **Features**:
  - Simple chat interactions
  - Multi-turn conversations with history
  - System messages for context setting
  - Role management (user, assistant, system)

### Configuration (`config/`)
- **Purpose**: Demonstrates various configuration options
- **Features**:
  - Environment variable configuration
  - Custom configuration with timeouts and retries
  - Default configuration usage
  - Provider switching
  - Configuration method chaining

### Error Handling (`errors/`)
- **Purpose**: Shows comprehensive error handling patterns
- **Features**:
  - Basic error handling and type checking
  - Retry logic with exponential backoff
  - Rate limiting handling
  - Different error type scenarios
  - Custom retry strategies

### Parameter Customization (`parameters/`)
- **Purpose**: Demonstrates advanced parameter usage
- **Features**:
  - Temperature variations and effects
  - Token limit examples
  - Stop sequence usage
  - Cross-provider parameter compatibility
  - Parameter clamping demonstration

### Provider Switching (`provider-switching/`)
- **Purpose**: Shows how to work with multiple providers
- **Features**:
  - Basic provider switching
  - Performance comparison between providers
  - Fallback provider strategies
  - Provider feature comparison
  - Provider selection based on requirements

## Running the Examples

Each example is self-contained and can be run independently:

```bash
# Run basic usage examples
go run ./examples/basic/

# Run chat examples
go run ./examples/chat/

# Run configuration examples
go run ./examples/config/

# Run error handling examples
go run ./examples/errors/

# Run parameter customization examples
go run ./examples/parameters/

# Run provider switching examples
go run ./examples/provider-switching/
```

## Using with the Actual Library

These examples use mock implementations for demonstration. To use with the actual library:

1. **Install the library**:
   ```bash
   go get github.com/ai-provider-wrapper/ai-provider-wrapper
   ```

2. **Update imports**:
   ```go
   import wrapper "github.com/ai-provider-wrapper/ai-provider-wrapper"
   ```

3. **Replace mock types**:
   - `ProviderType` → `wrapper.ProviderType`
   - `Config` → `wrapper.Config`
   - `CompletionRequest` → `wrapper.CompletionRequest`
   - `CompletionResponse` → `wrapper.CompletionResponse`
   - `ChatRequest` → `wrapper.ChatRequest`
   - `ChatResponse` → `wrapper.ChatResponse`
   - `NewClient()` → `wrapper.NewClient()`

4. **Add your API keys**:
   - Replace `"sk-your-openai-api-key-here"` with your actual OpenAI API key
   - Replace `"sk-ant-your-anthropic-api-key-here"` with your actual Anthropic API key

## Example Code Patterns

### Basic Client Creation
```go
client, err := wrapper.NewClient(wrapper.ProviderOpenAI, wrapper.Config{
    APIKey: "your-api-key-here",
})
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Completion Request
```go
req := wrapper.CompletionRequest{
    Prompt:      "Your prompt here",
    Temperature: &temperature,
    MaxTokens:   &maxTokens,
}

resp, err := client.Complete(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Text)
```

### Chat Request
```go
req := wrapper.ChatRequest{
    Messages: []wrapper.Message{
        {Role: "user", Content: "Hello!"},
    },
}

resp, err := client.ChatComplete(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Message.Content)
```

### Error Handling
```go
if err != nil {
    if wrapperErr, ok := err.(*wrapper.Error); ok {
        switch wrapperErr.Type {
        case wrapper.ErrorTypeAuth:
            // Handle authentication errors
        case wrapper.ErrorTypeRateLimit:
            // Handle rate limiting
        case wrapper.ErrorTypeValidation:
            // Handle validation errors
        }
    }
}
```

## Requirements Addressed

These examples address the following requirements from the specification:

- **Requirement 7.1**: Clear API documentation with examples
- **Requirement 7.2**: Code samples for common use cases  
- **Requirement 4.1**: Provider-agnostic parameter configuration
- **Requirement 5.1**: Consistent error handling across providers
- **Requirements 1.1-1.4**: Client initialization and provider selection
- **Requirements 2.1-2.2**: Text completion functionality
- **Requirements 3.1-3.3**: Chat completion with conversation history

## Notes

- All examples include comprehensive comments explaining each step
- Mock implementations simulate real API behavior for demonstration
- Examples show both success and error scenarios
- Parameter validation and clamping are demonstrated
- Cross-provider compatibility is highlighted throughout