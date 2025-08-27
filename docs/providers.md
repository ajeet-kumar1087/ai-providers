# Provider-specific Documentation

This document contains detailed information about each supported AI provider, including their specific capabilities, limitations, and configuration options.

## OpenAI

### Overview
OpenAI provides access to GPT models through their API. The AI Provider Wrapper supports all major OpenAI models including GPT-3.5 and GPT-4 variants.

### Configuration

#### API Key Format
- **Format**: `sk-` followed by 48 characters
- **Example**: `sk-1234567890abcdef1234567890abcdef12345678`
- **Environment Variable**: `OPENAI_API_KEY`

#### Base URL
- **Default**: `https://api.openai.com/v1`
- **Custom**: Set via `OPENAI_BASE_URL` environment variable
- **Use Case**: Custom deployments, Azure OpenAI Service

#### Example Configuration
```go
config := wrapper.Config{
    APIKey:  "sk-your-openai-api-key",
    BaseURL: "https://api.openai.com/v1", // Optional
    Timeout: 30 * time.Second,
}
```

### Supported Models
| Model | Context Length | Best For |
|-------|----------------|----------|
| GPT-3.5-turbo | 4,096 tokens | Fast, cost-effective tasks |
| GPT-4 | 8,192 tokens | Complex reasoning, high quality |
| GPT-4-turbo | 128,000 tokens | Long documents, complex tasks |

### Parameter Mapping

#### Temperature
- **Range**: 0.0 - 2.0
- **Default**: 1.0
- **Behavior**: Higher values increase randomness

#### Max Tokens
- **Range**: 1 - 4,096 (varies by model)
- **Default**: Model-specific
- **Note**: Includes both prompt and completion tokens

#### Stop Sequences
- **Maximum**: 4 stop sequences
- **Format**: Array of strings
- **Behavior**: Generation stops when any sequence is encountered

### Rate Limits
- **Requests per minute**: Varies by plan and model
- **Tokens per minute**: Varies by plan and model
- **Retry Strategy**: Exponential backoff recommended
- **Headers**: Rate limit info in response headers

### Error Handling

#### Common Error Codes
- `invalid_api_key`: API key is invalid or missing
- `insufficient_quota`: Account has exceeded quota
- `rate_limit_exceeded`: Too many requests
- `model_not_found`: Specified model doesn't exist
- `context_length_exceeded`: Input too long for model

#### Example Error Response
```json
{
  "error": {
    "message": "You exceeded your current quota",
    "type": "insufficient_quota",
    "code": "insufficient_quota"
  }
}
```

### Best Practices
1. **Model Selection**: Use GPT-3.5-turbo for most tasks, GPT-4 for complex reasoning
2. **Token Management**: Monitor token usage to control costs
3. **Rate Limiting**: Implement exponential backoff for rate limit errors
4. **Prompt Engineering**: Use clear, specific prompts for better results

---

## Anthropic

### Overview
Anthropic provides access to Claude models, known for their safety features and large context windows. The AI Provider Wrapper supports Claude-3 and Claude-2 model families.

### Configuration

#### API Key Format
- **Format**: `sk-ant-` followed by additional characters
- **Example**: `sk-ant-api03-1234567890abcdef1234567890abcdef12345678`
- **Environment Variable**: `ANTHROPIC_API_KEY`

#### Base URL
- **Default**: `https://api.anthropic.com`
- **Custom**: Set via `ANTHROPIC_BASE_URL` environment variable
- **Use Case**: Regional endpoints, custom deployments

#### Example Configuration
```go
config := wrapper.Config{
    APIKey:  "sk-ant-your-anthropic-api-key",
    BaseURL: "https://api.anthropic.com", // Optional
    Timeout: 30 * time.Second,
}
```

### Supported Models
| Model | Context Length | Best For |
|-------|----------------|----------|
| Claude-3 Haiku | 200,000 tokens | Fast, lightweight tasks |
| Claude-3 Sonnet | 200,000 tokens | Balanced performance |
| Claude-3 Opus | 200,000 tokens | Most capable, complex tasks |
| Claude-2 | 100,000 tokens | General purpose |

### Parameter Mapping

#### Temperature
- **Range**: 0.0 - 1.0 (different from OpenAI)
- **Default**: 1.0
- **Behavior**: Higher values increase randomness
- **Note**: Automatically clamped from OpenAI's 0-2 range

#### Max Tokens
- **Range**: 1 - 100,000
- **Default**: Model-specific
- **Note**: Much higher limits than OpenAI

#### Stop Sequences
- **Maximum**: 10 stop sequences (more than OpenAI)
- **Format**: Array of strings
- **Behavior**: Generation stops when any sequence is encountered

### Rate Limits
- **Requests per minute**: Varies by plan
- **Tokens per minute**: Varies by plan
- **Retry Strategy**: Exponential backoff with longer delays
- **Headers**: Rate limit info in response headers

### Error Handling

#### Common Error Codes
- `authentication_error`: API key is invalid
- `permission_error`: Insufficient permissions
- `rate_limit_error`: Rate limit exceeded
- `invalid_request_error`: Malformed request
- `overloaded_error`: Service temporarily overloaded

#### Example Error Response
```json
{
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded"
  }
}
```

### Special Features

#### System Messages
- **Support**: Excellent system message support
- **Behavior**: System messages provide context and instructions
- **Best Practice**: Use system messages for role definition

#### Constitutional AI
- **Feature**: Built-in safety and helpfulness training
- **Behavior**: Automatically avoids harmful content
- **Benefit**: Reduced need for content filtering

### Best Practices
1. **Context Windows**: Take advantage of large context windows for document analysis
2. **System Messages**: Use detailed system messages for better control
3. **Safety**: Rely on built-in safety features but still validate outputs
4. **Token Efficiency**: Claude is efficient with tokens, suitable for long conversations

---

## Google AI

### Overview
Google AI provides access to Gemini models through their API. Support is currently in development and will be available in a future release.

### Planned Features
- **Models**: Gemini Pro, Gemini Pro Vision
- **Context Length**: Up to 32,000 tokens
- **Multimodal**: Text and image inputs (future)
- **Integration**: Full compatibility with existing wrapper interface

### Configuration (Planned)

#### API Key Format
- **Format**: Standard Google API key (39 characters)
- **Environment Variable**: `GOOGLE_API_KEY`

#### Base URL
- **Default**: `https://generativelanguage.googleapis.com`
- **Custom**: Set via `GOOGLE_BASE_URL` environment variable

### Development Status
- **Status**: 🚧 In Development
- **Expected**: Q2 2024
- **Features**: Text completion, chat completion
- **Multimodal**: Planned for later release

### Getting Updates
Follow the project repository for updates on Google AI integration progress.

---

## Provider Comparison

### Feature Matrix
| Feature | OpenAI | Anthropic | Google AI |
|---------|--------|-----------|-----------|
| Text Completion | ✅ | ✅ | 🚧 |
| Chat Completion | ✅ | ✅ | 🚧 |
| Max Context | 128K | 200K | 32K |
| Temperature Range | 0-2 | 0-1 | 0-1 |
| Stop Sequences | 4 | 10 | TBD |
| Streaming | 🚧 | 🚧 | 🚧 |
| Function Calling | 🚧 | ❌ | 🚧 |

### Performance Characteristics
| Provider | Speed | Cost | Quality | Context |
|----------|-------|------|---------|---------|
| OpenAI | Fast | Medium | High | Medium |
| Anthropic | Medium | Medium | High | Large |
| Google AI | TBD | TBD | TBD | Medium |

### Use Case Recommendations

#### Choose OpenAI When:
- You need fast response times
- Working with shorter contexts
- Cost efficiency is important
- You need the latest model features

#### Choose Anthropic When:
- You need large context windows
- Safety and alignment are priorities
- Working with long documents
- You prefer constitutional AI approach

#### Choose Google AI When:
- Available in future releases
- You need Google ecosystem integration
- Multimodal capabilities are required

---

## Migration Between Providers

### Switching Providers
The wrapper makes it easy to switch between providers:

```go
// Original OpenAI client
client1, _ := wrapper.NewClient(wrapper.ProviderOpenAI, config)

// Switch to Anthropic with same interface
client2, _ := wrapper.NewClient(wrapper.ProviderAnthropic, config)

// Same request works with both
request := wrapper.CompletionRequest{
    Prompt: "Hello, world!",
    Temperature: &[]float64{0.7}[0],
}

resp1, _ := client1.Complete(ctx, request)
resp2, _ := client2.Complete(ctx, request)
```

### Parameter Considerations
- **Temperature**: Automatically clamped to provider ranges
- **Max Tokens**: Validated against provider limits
- **Stop Sequences**: Limited to provider maximums
- **Context Length**: Validated against model limits

### Cost Optimization
1. **Model Selection**: Choose appropriate model for task complexity
2. **Token Management**: Monitor usage across providers
3. **Caching**: Implement response caching for repeated requests
4. **Batch Processing**: Group similar requests when possible

---

## Getting Help

### Provider-Specific Issues
1. **OpenAI**: Check [OpenAI Status](https://status.openai.com/)
2. **Anthropic**: Check [Anthropic Status](https://status.anthropic.com/)
3. **Google AI**: Check Google Cloud Status (when available)

### Documentation Links
- **OpenAI**: [API Documentation](https://platform.openai.com/docs)
- **Anthropic**: [API Documentation](https://docs.anthropic.com/)
- **Google AI**: [AI Platform Documentation](https://cloud.google.com/ai-platform/docs)

### Support Channels
- **Wrapper Issues**: GitHub Issues
- **Provider Issues**: Contact provider support directly
- **Integration Help**: Check examples and troubleshooting guide