package openai

import (
	"fmt"
	"strings"

	"github.com/nexora/sdk/base"
)

// Model represents an OpenAI model
type Model struct {
	id      string
	object  string
	created int64
	ownedBy string
}

func (m Model) GetID() string {
	return m.id
}

func (m Model) GetName() string {
	// OpenAI doesn't provide names, use the ID as name
	return m.id
}

func (m Model) GetProvider() string {
	return "openai"
}

func (m Model) GetCapabilities() base.ModelCapabilities {
	cap := base.ModelCapabilities{
		SupportsChat:       true,
		SupportsFIM:        false,
		SupportsEmbeddings: strings.Contains(strings.ToLower(m.id), "embedding"),
		SupportsFineTuning: strings.Contains(strings.ToLower(m.id), "ft:"),
		SupportsAgents:     false,
		SupportsFileUpload: true,
		SupportsStreaming:  true,
		SupportsJSONMode:   true,
		SupportsVision:     strings.Contains(strings.ToLower(m.id), "vision") || strings.Contains(strings.ToLower(m.id), "dall-e"),
		SupportsAudio:      strings.Contains(strings.ToLower(m.id), "tts") || strings.Contains(strings.ToLower(m.id), "whisper"),
		SupportsTools:      true,
		CanReason:          strings.Contains(strings.ToLower(m.id), "o1-preview") || strings.Contains(strings.ToLower(m.id), "gpt-4"),
	}

	// Add specific capabilities
	if cap.SupportsTools {
		cap.SupportedParameters = []string{
			"model", "messages", "temperature", "top_p", "max_tokens",
			"stream", "stop", "response_format", "tools", "tool_choice",
			"presence_penalty", "frequency_penalty", "n", "seed",
		}
		cap.SecurityFeatures = []string{"content_filtering"}
	}
	if cap.SupportsVision {
		cap.SupportedParameters = append(cap.SupportedParameters, "modalities")
	}
	if cap.SupportsAudio {
		cap.SupportedParameters = append(cap.SupportedParameters, "voice", "speed")
	}

	return cap
}

func (m Model) GetPricing() base.ModelPricing {
	// Default pricing (you can customize)
	return base.ModelPricing{
		CostPer1MIn:  0.005, // Example price
		CostPer1MOut: 0.015,
		Currency:     "USD",
		Unit:         "1M tokens",
	}
}

func (m Model) GetContextWindow() int {
	// Return context windows based on known models
	switch {
	case strings.Contains(m.id, "gpt-4"):
		return 128000
	case strings.Contains(m.id, "gpt-3.5"):
		return 4096
	case strings.Contains(m.id, "gpt-3"):
		return 2048
	default:
		return 4096 // Default for unknown models
	}
}

// ModelListResponse represents OpenAI's model list response
type ModelListResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}
