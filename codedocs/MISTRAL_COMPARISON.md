# Mistral API Implementation Comparison

This document compares our NEXORA implementations of Mistral, Devstral, and Codestral providers against the official Mistral AI API schema.

## Summary

Our implementations use OpenAI-compatible endpoints to access Mistral's services, but there are several gaps and inconsistencies compared to the official API specification.

## Key Differences

### 1. API Approach

**Official Mistral API:**
- Native REST endpoints with distinct schemas
- Multiple dedicated endpoints (chat, fim, agents, embeddings, models, files, fine-tuning)
- Specific request/response structures for each endpoint
- Tool calling with native function schema support

**Our Implementation:**
- Uses OpenAI-compatible endpoint format (`/v1/chat/completions`)
- Relies on OpenAI compatibility layer instead of native APIs
- Single endpoint approach for all models
- May not expose all Mistral-specific features

### 2. Model Implementations

#### Mistral Native Provider (`mistral_native.go`)

**Models Supported:**
- `devstral-2512` (Devstral 2, 123B)
- `devstral-small-2512` (Devstral Small 2, 24B)

**Issues:**
- ❌ Uses incorrect model naming convention (should be `devstral-latest` according to docs)
- ❌ Missing general purpose models (Mistral Small/Large/Medium, Ministral)
- ❌ Not using native API type (uses `catwalk.TypeOpenAICompat`)

#### Mistral Devstral Provider (`mistral_devstral.go`)

**Models Supported:**
- `devstral-2512` 
- `devstral-small-2512`

**Issues:**
- ❌ Same naming issues as native provider
- ❌ Costs marked as FREE during beta (should verify current pricing)
- ❌ Missing integration with Agents API for agentic tasks

#### Mistral Codestral Provider (`mistral_codestral.go`)

**Models Supported:**
- `codestral-25-08`
- `codestral-embed-25-05`

**Issues:**
- ❌ Using incorrect model IDs (should be `codestral-latest` for chat, `codestral-2405` or `codestral-latest` for FIM)
- ❌ Not utilizing dedicated FIM endpoint (`/v1/fim/completions`)
- ❌ Missing code interpreter tool support
- ❌ Embed model context window may be incorrect (spec shows 8192, we match)

#### Mistral General Provider (`mistral_general.go`)

**Models Supported:**
- `mistral-large-3-25-12`
- `mistral-medium-3-1-25-08`
- `mistral-small-3-2-25-06`
- `ministral-3-14b-25-12`
- `ministral-3-8b-25-12`
- `ministral-3-3b-25-12`
- `mistral-embed`

**Issues:**
- ❌ Model names include date identifiers not in official naming convention
- ❌ Should use consistent naming like `mistral-large-latest`, `mistral-small-latest`, etc.
- ❌ Ministral models show image support (correct)

### 3. Missing Features

#### Completely Missing from Our Implementation:

1. **Agents API** (`/v1/agents/`)
   - No support for creating, listing, or managing agents
   - Missing agent tools (WebSearch, CodeInterpreter, ImageGeneration)
   - No agent completion endpoint support

2. **Fine-tuning API** (`/v1/fine_tuning/jobs`)
   - No fine-tuning job creation or management
   - Missing training parameters configuration
   - No integration with file uploads for training data

3. **Files API** (`/v1/files`)
   - No file upload support
   - Cannot manage training/validation files

4. **Models Management API** (`/v1/models`)
   - No dynamic model listing capability
   - Hardcoded model definitions instead of dynamic

5. **Native Streaming**
   - May use OpenAI's streaming format instead of Mistral's native format

6. **Mistral-Specific Parameters**
   - `safe_prompt` parameter not exposed
   - `response_format` with JSON schema mode not implemented
   - `prompt_mode` for reasoning models missing
   - `parallel_tool_calls` parameter not exposed

### 4. Parameter Mapping Issues

**Chat Completion Parameters:**

| Parameter | Official API | Our Implementation | Status |
|-----------|-------------|-------------------|---------|
| `model` | Required | Required | ✅ Implemented |
| `messages` | Required | Required | ✅ Implemented (via OpenAI compat) |
| `temperature` | 0.0-1.5 | ✓ | ✅ Implemented |
| `top_p` | 0-1 | ✓ | ✅ Implemented |
| `max_tokens` | Optional | ✓ | ✅ Implemented |
| `min_tokens` | Optional | ❌ | ❌ Missing |
| `stream` | Boolean | ✓ | ✅ Implemented |
| `stop` | String/array | ✓ | ✅ Implemented |
| `random_seed` | Integer | ❌ | ❌ Missing |
| `response_format` | JSON Schema | ❌ | ❌ Missing |
| `tools` | Array | ✓ (via OpenAI) | ⚠️ May have schema differences |
| `tool_choice` | auto/none/any/required | ✓ (via OpenAI) | ⚠️ May not support all options |
| `safe_prompt` | Boolean | ❌ | ❌ Missing |

### 5. Security and Authentication

**Requirements:**
- ✅ Both use Bearer token authentication
- ✅ Same API key format (`$MISTRAL_API_KEY`)

**Differences:**
- ❌ Our implementation may not validate token scopes
- ❌ No support for API key rotation or multiple keys

## Recommendations for Alignment

### High Priority:

1. **Update Model Naming Convention**
   ```go
   // Current
   "devstral-2512"
   "mistral-large-3-25-12"
   
   // Should be
   "devstral-latest"
   "mistral-large-latest"
   ```

2. **Implement Native FIM Endpoint**
   - Use `/v1/fim/completions` for Codestral
   - Support `prompt` and `suffix` parameters
   - Enable code-specific optimizations

3. **Add Missing Parameters**
   - `safe_prompt` for conversation safety
   - `response_format` for structured output
   - `random_seed` for deterministic results
   - `min_tokens` for response length control

### Medium Priority:

4. **Implement Agents API**
   - Create agent management endpoints
   - Support agent tools (WebSearch, CodeInterpreter)
   - Enable agent-based conversations

5. **Add Fine-tuning Support**
   - File upload capabilities
   - Job creation and monitoring
   - Model fine-tuning workflows

### Low Priority:

6. **Add Models Management**
   - Dynamic model listing
   - Model capability detection
   - Deprecation handling

7. **Implement Files API**
   - Training data management
   - File organization

## Implementation Strategy

### Phase 1: Model and Parameter Updates
1. Update all model IDs to match official naming
2. Add missing chat parameters
3. Implement native FIM endpoint

### Phase 2: Advanced Features
1. Implement Agents API infrastructure
2. Add fine-tuning job support
3. Integrate file management

### Phase 3: Full Migration
1. Create Mistral-native provider (not OpenAI-compatible)
2. Migrate existing providers to use native APIs where available
3. Maintain backward compatibility during transition

## Conclusion

While our current implementations provide functional access to Mistral models through OpenAI-compatible endpoints, there are significant gaps in feature support and API compliance. The most critical issues are:

1. Incorrect model naming conventions that may break with API updates
2. Missing native features like Agents and Fine-tuning
3. Absence of Mistral-specific parameters that enhance functionality

Implementing the recommended changes will improve reliability, feature parity, and future-proof our Mistral integrations.