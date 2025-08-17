# AI Provider Wrapper

A Go package that provides a unified interface for interacting with multiple AI providers (OpenAI, Anthropic, Google). This package abstracts away provider-specific implementations and offers a consistent API for developers to integrate AI capabilities into their applications without being locked into a single provider.

## Features

- **Unified Interface**: Single API for multiple AI providers
- **Provider Abstraction**: Switch between providers without code changes
- **Consistent Error Handling**: Standardized error types across all providers
- **Parameter Mapping**: Automatic translation between generic and provider-specific parameters
- **Minimal Dependencies**: Lightweight design using standard library components
- **Easy Installation**: Simple `go get` installation with Go modules support

## Supported Providers

- OpenAI (GPT models)
- Anthropic (Claude models)
- Google AI (Gemini models) - Coming soon

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ai-provider-wrapper/ai-provider-wrapper"
)

func main() {
    // Initialize client with OpenAI
    client, err := aiprovider.NewClient(aiprovider.ProviderOpenAI, aiprovider.Config{
        APIKey: "your-openai-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send completion request
    resp, err := client.Complete(context.Background(), aiprovider.CompletionRequest{
        Prompt: "Hello, world!",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(resp.Text)
}
```

## Installation

```bash
go get github.com/ai-provider-wrapper/ai-provider-wrapper
```

## Documentation

- [API Documentation](https://pkg.go.dev/github.com/ai-provider-wrapper/ai-provider-wrapper)
- [Provider-specific Documentation](docs/providers.md)
- [Troubleshooting Guide](docs/troubleshooting.md)

## Examples

See the [examples](examples/) directory for comprehensive usage examples:

- [Basic Usage](examples/basic/) - Simple completion requests
- [Chat Conversations](examples/chat/) - Multi-turn conversations
- [Configuration](examples/config/) - Advanced configuration options
- [Error Handling](examples/errors/) - Robust error handling patterns

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests for any improvements.

## License

This project is licensed under the MIT License - see the LICENSE file for details.