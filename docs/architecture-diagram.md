# AI Provider Wrapper - Architecture Diagram

This document contains class interaction diagrams that illustrate the architecture and code flow of the AI Provider Wrapper.

## High-Level Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        User[User Application]
        Client[Client Interface]
        Factory[ClientFactory]
    end
    
    subgraph "Core Layer"
        MainClient[client struct]
        Config[Config]
        Types[Types Package]
        Errors[Error Handling]
    end
    
    subgraph "Adapter Layer"
        ProviderAdapter[ProviderAdapter Interface]
        OpenAIAdapter[OpenAI Adapter]
        AnthropicAdapter[Anthropic Adapter]
        GoogleAdapter[Google Adapter]
    end
    
    subgraph "Utility Layer"
        Utils[Validation Utils]
        HTTPClient[HTTP Client]
        ParamMapper[Parameter Mapper]
    end
    
    subgraph "External APIs"
        OpenAIAPI[OpenAI API]
        AnthropicAPI[Anthropic API]
        GoogleAPI[Google AI API]
    end
    
    User --> Client
    User --> Factory
    Factory --> MainClient
    Client --> MainClient
    MainClient --> Config
    MainClient --> ProviderAdapter
    MainClient --> Utils
    MainClient --> Errors
    
    ProviderAdapter --> OpenAIAdapter
    ProviderAdapter --> AnthropicAdapter
    ProviderAdapter --> GoogleAdapter
    
    OpenAIAdapter --> HTTPClient
    AnthropicAdapter --> HTTPClient
    GoogleAdapter --> HTTPClient
    
    OpenAIAdapter --> OpenAIAPI
    AnthropicAdapter --> AnthropicAPI
    GoogleAdapter --> GoogleAPI
    
    Utils --> ParamMapper
    Types --> Config
    Types --> Errors
```

## Detailed Class Diagram

```mermaid
classDiagram
    class Client {
        <<interface>>
        +Complete(ctx, req) CompletionResponse
        +ChatComplete(ctx, req) ChatResponse
        +Close() error
    }
    
    class ClientFactory {
        <<interface>>
        +CreateClient(provider, config) Client
        +SupportedProviders() []ProviderType
    }
    
    class client {
        -adapter ProviderAdapter
        -provider ProviderType
        -config Config
        +Complete(ctx, req) CompletionResponse
        +ChatComplete(ctx, req) ChatResponse
        +Close() error
        -validateAndNormalizeCompletionRequest(req) CompletionRequest
        -validateAndNormalizeChatRequest(req) ChatRequest
        -validateConversationStructure(messages) error
    }
    
    class defaultClientFactory {
        +CreateClient(provider, config) Client
        +SupportedProviders() []ProviderType
    }
    
    class ProviderAdapter {
        <<interface>>
        +Complete(ctx, req) CompletionResponse
        +ChatComplete(ctx, req) ChatResponse
        +ValidateConfig(config) error
        +Name() string
        +SupportedFeatures() []string
    }
    
    class OpenAIAdapter {
        -httpClient HTTPClient
        -config AdapterConfig
        -baseURL string
        -apiKey string
        +Complete(ctx, req) CompletionResponse
        +ChatComplete(ctx, req) ChatResponse
        +ValidateConfig(config) error
        +Name() string
        +SupportedFeatures() []string
        -makeRequest(ctx, endpoint, body) Response
        -parseErrorResponse(resp) error
        -mapCompletionRequest(req) OpenAICompletionRequest
        -normalizeCompletionResponse(resp) CompletionResponse
    }
    
    class AnthropicAdapter {
        -httpClient HTTPClient
        -config AdapterConfig
        -baseURL string
        -apiKey string
        +Complete(ctx, req) CompletionResponse
        +ChatComplete(ctx, req) ChatResponse
        +ValidateConfig(config) error
        +Name() string
        +SupportedFeatures() []string
        -makeRequest(ctx, endpoint, body) Response
        -parseErrorResponse(resp) error
        -mapCompletionRequest(req) AnthropicCompletionRequest
        -mapChatRequest(req) AnthropicChatRequest
        -normalizeCompletionResponse(resp) CompletionResponse
        -normalizeChatResponse(resp) ChatResponse
    }
    
    class Config {
        +APIKey string
        +BaseURL string
        +Timeout Duration
        +MaxRetries int
        +Temperature *float64
        +MaxTokens *int
        +Validate(provider) error
        +WithAPIKey(key) Config
        +WithTimeout(timeout) Config
        +WithMaxRetries(retries) Config
        +WithTemperature(temp) Config
        +WithMaxTokens(tokens) Config
    }
    
    class Error {
        +Type ErrorType
        +Message string
        +Code string
        +Provider string
        +Wrapped error
        +RetryAfter *int
        +TokenCount *int
        +Error() string
        +Unwrap() error
        +Is(target) bool
        +IsRetryable() bool
    }
    
    class HTTPClient {
        -httpClient HTTPClient
        -timeout Duration
        -maxRetries int
        +Post(ctx, url, headers, body) Response
        +Get(ctx, url, headers) Response
        -doWithRetry(req) Response
        -shouldRetryError(err) bool
        -shouldRetryStatus(code) bool
        -waitBeforeRetry(attempt)
    }
    
    class ValidationUtils {
        +ValidateCompletionRequest(req) error
        +ValidateChatRequest(req) error
        +ValidateMessage(msg, index) error
        +ClampParameters(req, provider) interface
        +GetProviderTokenLimit(provider) int
        +GetProviderMaxTemperature(provider) float64
        +GetProviderMaxStopSequences(provider) int
    }
    
    class ParameterMapper {
        -sourceProvider ProviderType
        -targetProvider ProviderType
        +MapTemperature(temp) float64
        +MapMaxTokens(tokens) int
        +MapStopSequences(sequences) []string
    }
    
    Client <|-- client
    ClientFactory <|-- defaultClientFactory
    ProviderAdapter <|-- OpenAIAdapter
    ProviderAdapter <|-- AnthropicAdapter
    
    client --> ProviderAdapter
    client --> Config
    client --> ValidationUtils
    
    OpenAIAdapter --> HTTPClient
    AnthropicAdapter --> HTTPClient
    
    defaultClientFactory --> client
    
    ValidationUtils --> ParameterMapper
```

## Request Flow Sequence Diagram

```mermaid
sequenceDiagram
    participant User as User Application
    participant Client as Client Interface
    participant MainClient as client struct
    participant Utils as Validation Utils
    participant Adapter as Provider Adapter
    participant HTTP as HTTP Client
    participant API as External API
    
    User->>Client: NewClient(provider, config)
    Client->>MainClient: create instance
    MainClient->>Utils: validate config
    Utils-->>MainClient: validation result
    MainClient->>Adapter: CreateAdapter(provider, config)
    Adapter-->>MainClient: adapter instance
    MainClient-->>Client: client instance
    Client-->>User: configured client
    
    User->>Client: Complete(ctx, request)
    Client->>MainClient: Complete(ctx, request)
    MainClient->>Utils: validateAndNormalizeCompletionRequest(req)
    Utils-->>MainClient: normalized request
    MainClient->>Adapter: Complete(ctx, normalized_req)
    
    Adapter->>Adapter: mapCompletionRequest(req)
    Adapter->>HTTP: Post(ctx, url, headers, body)
    HTTP->>API: HTTP POST request
    API-->>HTTP: HTTP response
    HTTP-->>Adapter: response
    
    alt Success Response
        Adapter->>Adapter: normalizeCompletionResponse(resp)
        Adapter-->>MainClient: CompletionResponse
        MainClient-->>Client: CompletionResponse
        Client-->>User: CompletionResponse
    else Error Response
        Adapter->>Adapter: parseErrorResponse(resp)
        Adapter-->>MainClient: Error
        MainClient-->>Client: Error
        Client-->>User: Error
    end
```

## Error Handling Flow

```mermaid
flowchart TD
    A[Request Initiated] --> B[Client Validation]
    B --> C{Validation OK?}
    C -->|No| D[Return Validation Error]
    C -->|Yes| E[Normalize Parameters]
    E --> F[Send to Adapter]
    F --> G[Adapter Processing]
    G --> H[HTTP Request]
    H --> I{HTTP Success?}
    
    I -->|No| J[Parse HTTP Error]
    J --> K{Retryable?}
    K -->|Yes| L{Max Retries?}
    L -->|No| M[Wait & Retry]
    M --> H
    L -->|Yes| N[Return Network Error]
    K -->|No| O[Return Provider Error]
    
    I -->|Yes| P[Parse Response]
    P --> Q{Parse OK?}
    Q -->|No| R[Return Parse Error]
    Q -->|Yes| S[Normalize Response]
    S --> T[Return Success]
    
    D --> U[Standardized Error]
    N --> U
    O --> U
    R --> U
    T --> V[Success Response]
    U --> W[Error with Type & Provider Info]
```

## Provider Adapter Pattern

```mermaid
graph LR
    subgraph "Generic Interface"
        GReq[Generic Request]
        GResp[Generic Response]
    end
    
    subgraph "OpenAI Adapter"
        OAIMap[Map to OpenAI Format]
        OAIReq[OpenAI Request]
        OAIAPI[OpenAI API Call]
        OAIResp[OpenAI Response]
        OAINorm[Normalize Response]
    end
    
    subgraph "Anthropic Adapter"
        AntMap[Map to Anthropic Format]
        AntReq[Anthropic Request]
        AntAPI[Anthropic API Call]
        AntResp[Anthropic Response]
        AntNorm[Normalize Response]
    end
    
    GReq --> OAIMap
    OAIMap --> OAIReq
    OAIReq --> OAIAPI
    OAIAPI --> OAIResp
    OAIResp --> OAINorm
    OAINorm --> GResp
    
    GReq --> AntMap
    AntMap --> AntReq
    AntReq --> AntAPI
    AntAPI --> AntResp
    AntResp --> AntNorm
    AntNorm --> GResp
```

## Configuration and Factory Pattern

```mermaid
graph TB
    subgraph "Configuration Sources"
        EnvVars[Environment Variables]
        Defaults[Default Values]
        UserConfig[User Configuration]
    end
    
    subgraph "Configuration Builder"
        ConfigBuilder[Config Builder]
        Validation[Config Validation]
    end
    
    subgraph "Client Factory"
        Factory[ClientFactory]
        AdapterFactory[Adapter Factory]
    end
    
    subgraph "Client Creation"
        ClientInstance[Client Instance]
        AdapterInstance[Adapter Instance]
    end
    
    EnvVars --> ConfigBuilder
    Defaults --> ConfigBuilder
    UserConfig --> ConfigBuilder
    
    ConfigBuilder --> Validation
    Validation --> Factory
    
    Factory --> AdapterFactory
    AdapterFactory --> AdapterInstance
    Factory --> ClientInstance
    
    ClientInstance --> AdapterInstance
```

## Parameter Mapping and Validation

```mermaid
flowchart TD
    A[Raw Request Parameters] --> B[Basic Validation]
    B --> C{Valid?}
    C -->|No| D[Return Validation Error]
    C -->|Yes| E[Parameter Mapping]
    
    E --> F[Temperature Mapping]
    E --> G[Token Limit Mapping]
    E --> H[Stop Sequences Mapping]
    
    F --> I{Provider Limits}
    G --> I
    H --> I
    
    I --> J[Clamp to Provider Range]
    J --> K[Apply Default Values]
    K --> L[Final Normalized Request]
    
    subgraph "Provider-Specific Limits"
        M[OpenAI: temp 0-2, tokens 4K]
        N[Anthropic: temp 0-1, tokens 100K]
        O[Google: temp 0-1, tokens 8K]
    end
    
    I --> M
    I --> N
    I --> O
```

## Key Design Patterns

### 1. Adapter Pattern
- **Purpose**: Allows different AI providers to work with a unified interface
- **Implementation**: Each provider has its own adapter that implements `ProviderAdapter`
- **Benefits**: Easy to add new providers, consistent API across providers

### 2. Factory Pattern
- **Purpose**: Creates clients and adapters based on provider type
- **Implementation**: `ClientFactory` and adapter creation functions
- **Benefits**: Centralized creation logic, easy testing and dependency injection

### 3. Strategy Pattern
- **Purpose**: Different validation and parameter mapping strategies per provider
- **Implementation**: Provider-specific validation and parameter clamping
- **Benefits**: Flexible parameter handling, provider-specific optimizations

### 4. Decorator Pattern
- **Purpose**: Adds validation, normalization, and error handling around core functionality
- **Implementation**: Client wraps adapters with additional functionality
- **Benefits**: Separation of concerns, consistent behavior across providers

## Code Flow Summary

1. **Client Creation**:
   - User calls `NewClient()` or factory methods
   - Configuration is validated and normalized
   - Appropriate adapter is created based on provider type
   - Client instance wraps the adapter

2. **Request Processing**:
   - User calls `Complete()` or `ChatComplete()`
   - Client validates and normalizes request parameters
   - Request is passed to provider-specific adapter
   - Adapter maps generic request to provider format

3. **API Communication**:
   - Adapter makes HTTP request using shared HTTP client
   - HTTP client handles retries and error conditions
   - Response is parsed and error-checked

4. **Response Processing**:
   - Adapter normalizes provider response to generic format
   - Response is returned through the client interface
   - Errors are standardized and categorized

5. **Error Handling**:
   - All errors are wrapped in standardized `Error` type
   - Error types enable consistent retry logic
   - Provider-specific error codes are preserved

This architecture provides a clean separation of concerns, making it easy to add new providers while maintaining a consistent API for users.