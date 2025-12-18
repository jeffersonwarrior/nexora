# Mistral AI API Schema Documentation

This document contains the complete schema specification for Mistral AI's API, based on their official OpenAPI specification.

## Overview

Mistral AI provides a comprehensive API for language models with the following main endpoints:
- Chat Completions
- FIM (Fill-In-the-Middle) Completions  
- Agents
- Embeddings
- Models Management
- Files Management
- Fine-tuning

## Base URL

```
https://api.mistral.ai
```

## Authentication

Authentication is performed via Bearer token:

```
Authorization: Bearer YOUR_API_KEY
```

## API Endpoints

### 1. Chat Completions

#### Endpoint: `POST /v1/chat/completions`

Creates a model response for the given chat conversation.

**Request Schema:**
```yaml
ChatCompletionRequest:
  type: object
  required:
    - messages
    - model
  properties:
    model:
      type: string
      description: ID of the model to use
      examples:
        - mistral-small-latest
    messages:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/SystemMessage'
          - $ref: '#/components/schemas/UserMessage'
          - $ref: '#/components/schemas/AssistantMessage'
          - $ref: '#/components/schemas/ToolMessage'
      description: The prompt(s) to generate completions for
    temperature:
      type: number
      minimum: 0
      maximum: 1.5
      default: 0.7
      description: Sampling temperature between 0.0 and 1.0
    top_p:
      type: number
      minimum: 0
      maximum: 1
      default: 1
      description: Nucleus sampling probability mass
    max_tokens:
      type: integer
      minimum: 0
      description: Maximum number of tokens to generate
    min_tokens:
      type: integer
      minimum: 0
      description: Minimum number of tokens to generate
    stream:
      type: boolean
      default: false
      description: Whether to stream back partial progress
    stop:
      oneOf:
        - type: string
        - type: array
          items:
            type: string
      description: Stop generation if this token is detected
    random_seed:
      type: integer
      minimum: 0
      description: Seed for deterministic results
    response_format:
      $ref: '#/components/schemas/ResponseFormat'
    tools:
      type: array
      items:
        $ref: '#/components/schemas/Tool'
    tool_choice:
      $ref: '#/components/schemas/ToolChoice'
      default: auto
    safe_prompt:
      type: boolean
      default: false
      description: Whether to inject a safety prompt
```

**Response Schema:**
```yaml
ChatCompletionResponse:
  allOf:
    - $ref: '#/components/schemas/ChatCompletionResponseBase'
    - type: object
      properties:
        choices:
          type: array
          items:
            $ref: '#/components/schemas/ChatCompletionChoice'
```

**Message Types:**

SystemMessage:
```yaml
type: object
required:
  - content
properties:
  content:
    oneOf:
      - type: string
      - type: array
        items:
          $ref: '#/components/schemas/ContentChunk'
  role:
    type: string
    enum: [system]
    default: system
```

UserMessage:
```yaml
type: object
required:
  - content
properties:
  content:
    oneOf:
      - type: string
      - type: array
        items:
          $ref: '#/components/schemas/TextChunk'
  role:
    type: string
    enum: [user]
    default: user
```

AssistantMessage:
```yaml
type: object
required: []
properties:
  content:
    oneOf:
      - type: string
      - type: 'null'
  tool_calls:
    type: array
    items:
      $ref: '#/components/schemas/ToolCall'
  prefix:
    type: boolean
    default: false
    description: Set to true when adding as prefix to condition model response
  role:
    type: string
    enum: [assistant]
    default: assistant
```

ToolMessage:
```yaml
type: object
required:
  - content
properties:
  content:
    type: string
  tool_call_id:
    oneOf:
      - type: string
      - type: 'null'
  name:
    oneOf:
      - type: string
      - type: 'null'
  role:
    type: string
    enum: [tool]
    default: tool
```

### 2. FIM (Fill-In-the-Middle) Completions

#### Endpoint: `POST /v1/fim/completions`

FIM completion for code generation.

**Request Schema:**
```yaml
FIMCompletionRequest:
  type: object
  required:
    - prompt
    - model
  properties:
    model:
      type: string
      default: codestral-2405
      description: ID of the model to use (compatible with codestral-2405, codestral-latest)
    prompt:
      type: string
      description: The text/code to complete
    suffix:
      type: string
      default: ''
      description: Optional text/code that adds context after the prompt
    temperature:
      type: number
      minimum: 0
      maximum: 1.5
      default: 0.7
    top_p:
      type: number
      minimum: 0
      maximum: 1
      default: 1
    max_tokens:
      type: integer
      minimum: 0
    min_tokens:
      type: integer
      minimum: 0
    stream:
      type: boolean
      default: false
    stop:
      oneOf:
        - type: string
        - type: array
          items:
            type: string
    random_seed:
      type: integer
      minimum: 0
```

### 3. Agents API

#### Endpoint: `POST /v1/agents/completions`

Agents completion API.

**Request Schema:**
```yaml
AgentsCompletionRequest:
  type: object
  required:
    - messages
    - agent_id
  properties:
    agent_id:
      type: string
      description: The ID of the agent to use
    messages:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/UserMessage'
          - $ref: '#/components/schemas/AssistantMessage'
          - $ref: '#/components/schemas/ToolMessage'
    max_tokens:
      type: integer
      minimum: 0
    min_tokens:
      type: integer
      minimum: 0
    stream:
      type: boolean
      default: false
    stop:
      oneOf:
        - type: string
        - type: array
          items:
            type: string
    random_seed:
      type: integer
      minimum: 0
    response_format:
      $ref: '#/components/schemas/ResponseFormat'
    tools:
      type: array
      items:
        $ref: '#/components/schemas/Tool'
    tool_choice:
      $ref: '#/components/schemas/ToolChoice'
      default: auto
```

#### Beta Agents API

##### List Agents: `GET /v1/agents`
Retrieves a list of agent entities sorted by creation time.

##### Create Agent: `POST /v1/agents`
Creates a new agent with instructions, tools, and description.

Request Schema:
```yaml
AgentCreateRequest:
  type: object
  required:
    - model
    - name
  properties:
    model:
      type: string
    name:
      type: string
    description:
      type: string
    instructions:
      type: string
      description: Instruction prompt the model will follow
    tools:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/FunctionTool'
          - $ref: '#/components/schemas/WebSearchTool'
          - $ref: '#/components/schemas/WebSearchPremiumTool'
          - $ref: '#/components/schemas/CodeInterpreterTool'
          - $ref: '#/components/schemas/ImageGenerationTool'
          - $ref: '#/components/schemas/DocumentLibraryTool'
    completion_args:
      $ref: '#/components/schemas/CompletionArgs'
    handoffs:
      type: array
      items:
        type: string
```

##### Get Agent: `GET /v1/agents/{agent_id}`
Retrieves an agent entity by ID.

##### Update Agent: `PATCH /v1/agents/{agent_id}`
Updates an agent entity and creates a new version.

##### Delete Agent: `DELETE /v1/agents/{agent_id}`
Deletes an agent entity.

### 4. Embeddings API

#### Endpoint: `POST /v1/embeddings`

Creates embeddings for the provided input texts.

**Request Schema:**
```yaml
EmbeddingRequest:
  type: object
  required:
    - input
    - model
  properties:
    model:
      type: string
      description: ID of the model to be used for embedding
    input:
      oneOf:
        - type: string
        - type: array
          items:
            type: string
      description: Text to embed
    encoding_format:
      type: string
      enum: [float, base64]
      default: float
    output_dimension:
      type: integer
      description: Dimension of output embeddings when feature available
    output_dtype:
      type: string
      enum: [float, int8, uint8, binary, ubinary]
```

**Response Schema:**
```yaml
EmbeddingResponse:
  type: object
  required:
    - data
    - model
    - object
    - usage
  properties:
    data:
      type: array
      items:
        $ref: '#/components/schemas/EmbeddingResponseData'
    model:
      type: string
    object:
      type: string
    usage:
      $ref: '#/components/schemas/UsageInfo'
```

### 5. Models API

#### List Models: `GET /v1/models`
Lists all models available to the user.

**Response Schema:**
```yaml
ModelList:
  type: object
  properties:
    object:
      type: string
      default: list
    data:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/BaseModelCard'
          - $ref: '#/components/schemas/FTModelCard'
```

#### Get Model: `GET /v1/models/{model_id}`
Retrieves information about a specific model.

**Response Schema:**
```yaml
ModelCard:
  type: object
  properties:
    id:
      type: string
    object:
      type: string
      default: model
    created:
      type: integer
    owned_by:
      type: string
      default: mistralai
    root:
      oneOf:
        - type: string
        - type: 'null'
    archived:
      type: boolean
      default: false
    name:
      oneOf:
        - type: string
        - type: 'null'
    description:
      oneOf:
        - type: string
        - type: 'null'
    capabilities:
      $ref: '#/components/schemas/ModelCapabilities'
    max_context_length:
      type: integer
      default: 32768
    aliases:
      type: array
      items:
        type: string
      default: []
    deprecation:
      oneOf:
        - type: string
          format: date-time
        - type: 'null'
```

#### Delete Model: `DELETE /v1/models/{model_id}`
Deletes a fine-tuned model.

### 6. Files API

#### Upload File: `POST /v1/files`
Uploads a file that can be used across various endpoints.

**Request Schema (multipart/form-data):**
```yaml
FileUpload:
  type: object
  required:
    - file
  properties:
    file:
      type: string
      format: binary
      description: The File object to be uploaded
    purpose:
      type: string
      enum: [fine-tune]
      default: fine-tune
```

#### List Files: `GET /v1/files`
Returns a list of files that belong to the user's organization.

#### Retrieve File: `GET /v1/files/{file_id}`
Returns information about a specific file.

#### Delete File: `DELETE /v1/files/{file_id}`
Deletes a file.

### 7. Fine-tuning API

#### Create Fine-tuning Job: `POST /v1/fine_tuning/jobs`
Creates a new fine-tuning job.

**Request Schema:**
```yaml
JobIn:
  type: object
  required:
    - model
    - hyperparameters
  properties:
    model:
      $ref: '#/components/schemas/FineTuneableModel'
    training_files:
      type: array
      items:
        $ref: '#/components/schemas/TrainingFile'
      default: []
    validation_files:
      type: array
      items:
        type: string
        format: uuid
    hyperparameters:
      $ref: '#/components/schemas/TrainingParametersIn'
    suffix:
      type: string
      maxLength: 18
    integrations:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/WandbIntegration'
    repositories:
      type: array
      items:
        oneOf:
          - $ref: '#/components/schemas/GithubRepositoryIn'
      default: []
    auto_start:
      type: boolean
```

#### Get Fine-tuning Jobs: `GET /v1/fine_tuning/jobs`
Gets a list of fine-tuning jobs.

#### Get Fine-tuning Job: `GET /v1/fine_tuning/jobs/{job_id}`
Gets details of a specific fine-tuning job.

#### Cancel Fine-tuning Job: `POST /v1/fine_tuning/jobs/{job_id}/cancel`
Requests cancellation of a fine-tuning job.

#### Start Fine-tuning Job: `POST /v1/fine_tuning/jobs/{job_id}/start`
Requests start of a validated fine-tuning job.

## Common Schema Components

### Tool Schemas

Tool:
```yaml
type: object
required:
  - function
properties:
  type:
    $ref: '#/components/schemas/ToolTypes'
    default: function
  function:
    $ref: '#/components/schemas/Function'
```

ToolCall:
```yaml
type: object
required:
  - function
properties:
  id:
    type: string
    default: 'null'
  type:
    $ref: '#/components/schemas/ToolTypes'
    default: function
  function:
    $ref: '#/components/schemas/FunctionCall'
```

Function:
```yaml
type: object
required:
  - name
  - parameters
properties:
  name:
    type: string
  description:
    type: string
    default: ''
  parameters:
    type: object
    additionalProperties: true
```

### Response Formats

ResponseFormat:
```yaml
type: object
properties:
  type:
    $ref: '#/components/schemas/ResponseFormats'
    default: text
```

ResponseFormats:
```yaml
type: string
enum:
  - text
  - json_object
description: >-
    Setting to { "type": "json_object" } enables JSON mode, which guarantees
    the message the model generates is in JSON
```

### Usage Information

UsageInfo:
```yaml
type: object
required:
  - prompt_tokens
  - completion_tokens
  - total_tokens
properties:
  prompt_tokens:
    type: integer
    example: 16
  completion_tokens:
    type: integer
    example: 34
  total_tokens:
    type: integer
    example: 50
```

### Model Capabilities

ModelCapabilities:
```yaml
type: object
properties:
  completion_chat:
    type: boolean
    default: true
  completion_fim:
    type: boolean
    default: false
  function_calling:
    type: boolean
    default: true
  fine_tuning:
    type: boolean
    default: false
```

## Tool Types

### Function Tools
Standard function calling tools with name, description, and parameters.

### Web Search Tools
- WebSearchTool: Basic web search capabilities
- WebSearchPremiumTool: Premium web search with additional features

### Code Interpreter Tool
Allows the model to execute code in a sandboxed environment.

### Image Generation Tool
Enables the model to generate images.

### Document Library Tool
Provides access to document libraries for reference.

## Streaming Responses

For endpoints that support streaming (chat completions, FIM completions), the response will be sent as server-sent events with the following format:

```
data: {json_chunk}

...
data: [DONE]
```

Each chunk contains partial completion information with delta updates.

## Error Handling

The API returns standard HTTP status codes and error responses:

```yaml
HTTPValidationError:
  type: object
  properties:
    detail:
      type: array
      items:
        $ref: '#/components/schemas/ValidationError'
```

ValidationError:
```yaml
type: object
required:
  - loc
  - msg
  - type
properties:
  loc:
    type: array
    items:
      oneOf:
        - type: string
        - type: integer
  msg:
    type: string
  type:
    type: string
```

## Fine-tuning Training Parameters

TrainingParametersIn:
```yaml
type: object
properties:
  training_steps:
    type: integer
    minimum: 1
    description: Number of training steps to perform
  learning_rate:
    type: number
    default: 0.0001
    minimum: 1e-8
    maximum: 1
  weight_decay:
    type: number
    default: 0.1
    minimum: 0
    maximum: 1
  warmup_fraction:
    type: number
    default: 0.05
    minimum: 0
    maximum: 1
  epochs:
    type: number
    minimum: 0.01
  fim_ratio:
    type: number
    default: 0.9
    minimum: 0
    maximum: 1
```

## SDK Support

Mistral AI provides official SDKs for:
- Python
- TypeScript/JavaScript

Example Python usage:
```python
from mistralai import Mistral

client = Mistral(api_key="your-api-key")

response = client.chat.complete(
    model="mistral-small-latest",
    messages=[{"content": "Hello!", "role": "user"}]
)
```

Example TypeScript usage:
```typescript
import { Mistral } from "@mistralai/mistralai";

const client = new Mistral({ apiKey: "your-api-key" });

const response = await client.chat.complete({
  model: "mistral-small-latest",
  messages: [{ content: "Hello!", role: "user" }]
});
```

## Rate Limits

The API has rate limits based on your subscription tier:
- Free tier: Limited requests per minute
- Paid tiers: Higher limits
- Enterprise: Custom limits

Contact support for rate limit increases or enterprise pricing.

## Additional Resources

- [Official Documentation](https://docs.mistral.ai)
- [Mistral Console](https://console.mistral.ai)
- [API Reference](https://docs.mistral.ai/api)
- [GitHub Repository](https://github.com/mistralai)