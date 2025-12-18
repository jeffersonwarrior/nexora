package zai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nexora/cli/internal/config"
)

const (
	serverName = "zai-vision"
)

// Config holds the Z.ai-specific configuration
type Config struct {
	APIKey string `json:"api_key"`
}

// Client wraps the MCP client session for Z.ai vision tools
type Client struct {
	session *mcp.ClientSession
	config  Config
	logger  *slog.Logger
}

// NewClient creates a new Z.ai MCP client
func NewClient(cfg config.Config) (*Client, error) {
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ZAI_API_KEY environment variable is required")
	}

	config := Config{
		APIKey: apiKey,
	}

	return &Client{
		config: config,
		logger: slog.Default().With("mcp", "zai-vision"),
	}, nil
}

// Connect establishes connection to Z.ai MCP server
func (c *Client) Connect(ctx context.Context) error {
	c.logger.Info("connecting to Z.ai vision MCP server")

	// Mock session creation for now - in real implementation this would create
	// an actual connection to a Z.ai MCP server
	c.session = &mcp.ClientSession{}

	return nil
}

// Disconnect closes the connection to Z.ai MCP server
func (c *Client) Disconnect(ctx context.Context) error {
	c.logger.Info("disconnecting from Z.ai vision MCP server")

	if c.session != nil {
		// Close session in real implementation
		c.session = nil
	}

	return nil
}

// GetTools returns the list of available Z.ai vision tools
func (c *Client) GetTools() []*mcp.Tool {
	return []*mcp.Tool{
		{
			Name:        "mcp_vision_analyze_data_visualization",
			Description: "Analyze data visualizations, charts, graphs, and dashboards to extract insights and trends. Use this tool ONLY when the user has a data visualization image and wants to understand the data patterns or metrics.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"analysis_focus": map[string]any{
						"type":        "string",
						"description": "Optional: specify what to focus on (e.g., 'trends', 'anomalies', 'comparisons', 'performance metrics'). Leave empty for comprehensive analysis.",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "What insights or information you want to extract from this visualization.",
					},
				},
				"required": []string{"image_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_analyze_image",
			Description: "General-purpose image analysis for scenarios not covered by specialized tools. Use this tool as a FALLBACK when none of the other specialized tools (ui_to_artifact, extract_text_from_screenshot, diagnose_error_screenshot, understand_technical_diagram, analyze_data_visualization) fit the user's need. This tool provides flexible image understanding for any visual content.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Detailed description of what you want to analyze, extract, or understand from the image. Be specific about your requirements.",
					},
				},
				"required": []string{"image_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_extract_text_from_screenshot",
			Description: "Extract and recognize text from screenshots using advanced OCR capabilities. Use this tool ONLY when the user has a screenshot containing text and wants to extract it. This tool specializes in OCR for code, terminal output, documentation, and general text extraction.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Instructions for text extraction. Specify what type of text to extract and any formatting requirements.",
					},
					"programming_language": map[string]any{
						"type":        "string",
						"description": "Optional: specify the programming language if the screenshot contains code (e.g., 'python', 'javascript', 'java'). Leave empty for auto-detection or non-code text.",
					},
				},
				"required": []string{"image_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_analyze_video",
			Description: "Analyze video content using advanced AI vision models. Use this tool when the user wants to: Understand what happens in a video, Extract key moments or actions from video, Analyze video content, scenes, or sequences, Get descriptions of video footage, Identify objects, people, or activities in video. Supports both local files and remote URL. Maximum file size: 8MB. Supports MP4, MOV, M4V formats.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"video_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the video (supports MP4, MOV, M4V)",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Detailed text prompt describing what to analyze, extract, or understand from the video",
					},
				},
				"required": []string{"video_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_ui_to_artifact",
			Description: "Convert UI screenshots into various artifacts: code, prompts, design specifications, or descriptions. Use this tool ONLY when the user wants to: - Generate frontend code from UI design (output_type='code') - Create AI prompts for UI generation (output_type='prompt') - Extract design specifications (output_type='spec') - Get natural language description of the UI (output_type='description'). Do NOT use for: screenshots containing text/code to extract, error messages, diagrams, or data visualizations.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"output_type": map[string]any{
						"type":        "string",
						"description": "Type of output to generate. Options: 'code' (generate frontend code), 'prompt' (generate AI prompt for recreating this UI), 'spec' (generate design specification document), 'description' (natural language description of the UI).",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Detailed instructions describing what to generate from this UI image. Should clearly state the desired output and any specific requirements.",
					},
				},
				"required": []string{"image_source", "output_type", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_diagnose_error_screenshot",
			Description: "Diagnose and analyze error messages, stack traces, and exception screenshots. Use this tool ONLY when the user has an error screenshot and needs help understanding or fixing it. This tool specializes in error analysis and provides actionable solutions.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"context": map[string]any{
						"type":        "string",
						"description": "Optional: additional context about when the error occurred (e.g., 'during npm install', 'when running the app', 'after deployment'). Helps with more accurate diagnosis.",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Description of what you need help with regarding this error. Include any relevant context about when it occurred.",
					},
				},
				"required": []string{"image_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_understand_technical_diagram",
			Description: "Analyze and explain technical diagrams including architecture diagrams, flowcharts, UML, ER diagrams, and system design diagrams. Use this tool ONLY when the user has a technical diagram and wants to understand its structure or components. This tool specializes in interpreting visual technical documentation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the image",
					},
					"diagram_type": map[string]any{
						"type":        "string",
						"description": "Optional: specify the diagram type if known (e.g., 'architecture', 'flowchart', 'uml', 'er-diagram', 'sequence'). Leave empty for auto-detection.",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "What you want to understand or extract from this diagram.",
					},
				},
				"required": []string{"image_source", "prompt"},
			},
		},
		{
			Name:        "mcp_vision_ui_diff_check",
			Description: "Compare two UI screenshots to identify visual differences and implementation discrepancies. Use this tool ONLY when the user wants to compare an expected/reference UI with an actual implementation. This tool is specialized for UI quality assurance and design-to-implementation verification.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"expected_image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the expected/reference image",
					},
					"actual_image_source": map[string]any{
						"type":        "string",
						"description": "Local file path or remote URL to the actual implementation image",
					},
					"prompt": map[string]any{
						"type":        "string",
						"description": "Instructions for the comparison. Specify what aspects to focus on or what level of detail is needed.",
					},
				},
				"required": []string{"expected_image_source", "actual_image_source", "prompt"},
			},
		},
	}
}

// CallTool executes a Z.ai vision tool with the given parameters
func (c *Client) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	c.logger.Info("calling Z.ai vision tool", "tool", params.Name)

	// Mock implementations for all tools - in real implementation these would
	// make actual API calls to Z.ai vision services

	switch params.Name {
	case "mcp_vision_analyze_data_visualization":
		return c.analyzeDataVisualization(ctx, params.Arguments)
	case "mcp_vision_analyze_image":
		return c.analyzeImage(ctx, params.Arguments)
	case "mcp_vision_extract_text_from_screenshot":
		return c.extractTextFromScreenshot(ctx, params.Arguments)
	case "mcp_vision_analyze_video":
		return c.analyzeVideo(ctx, params.Arguments)
	case "mcp_vision_ui_to_artifact":
		return c.uiToArtifact(ctx, params.Arguments)
	case "mcp_vision_diagnose_error_screenshot":
		return c.diagnoseErrorScreenshot(ctx, params.Arguments)
	case "mcp_vision_understand_technical_diagram":
		return c.understandTechnicalDiagram(ctx, params.Arguments)
	case "mcp_vision_ui_diff_check":
		return c.uiDiffCheck(ctx, params.Arguments)
	default:
		return nil, fmt.Errorf("unknown Z.ai vision tool: %s", params.Name)
	}
}

// analyzeDataVisualization is a mock implementation of data visualization analysis
func (c *Client) analyzeDataVisualization(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	// In real implementation, this would analyze the data visualization
	// using Z.ai's computer vision capabilities

	response := map[string]any{
		"insights": []string{
			"Data shows an upward trend over time",
			"Peak values observed in Q3",
			"Anomaly detected in May 2023",
		},
		"metrics": map[string]any{
			"total_points": 150,
			"trend":        "positive",
			"correlation":  0.87,
		},
		"analysis": "The visualization shows clear patterns indicating growth with seasonal variations",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// analyzeImage is a mock implementation of general image analysis
func (c *Client) analyzeImage(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"description": "Image contains a modern interface with clean layout",
		"elements": []string{
			"Navigation bar at top",
			"Central content area",
			"Action buttons",
		},
		"style": "Minimalist design with good contrast",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// extractTextFromScreenshot is a mock implementation of OCR
func (c *Client) extractTextFromScreenshot(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"extracted_text": "This is the extracted text from the screenshot",
		"confidence":     0.95,
		"format":         "plain_text",
		"language":       "en",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// analyzeVideo is a mock implementation of video analysis
func (c *Client) analyzeVideo(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"summary":  "Video shows a demonstration of a software application",
		"duration": "2:30",
		"key_frames": []string{
			"0:00 - Opening screen with logo",
			"0:15 - Main interface walkthrough",
			"1:45 - Feature demonstration",
			"2:15 - Closing screen",
		},
		"content": "The video presents a step-by-step guide to using the application",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// uiToArtifact is a mock implementation of UI to artifact conversion
func (c *Client) uiToArtifact(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"artifact_type":  "code",
		"generated_code": "<div class=\"container\">\n  <h1>Generated UI</h1>\n  <button>Click me</button>\n</div>",
		"framework":      "HTML/CSS",
		"description":    "Generated HTML/CSS code matching the UI design",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// diagnoseErrorScreenshot is a mock implementation of error diagnosis
func (c *Client) diagnoseErrorScreenshot(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"error_type":     "SyntaxError",
		"error_location": "line 42, column 15",
		"diagnosis":      "Missing closing parenthesis in function call",
		"solution":       "Add closing parenthesis: functionName(arg1, arg2)",
		"confidence":     0.92,
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// understandTechnicalDiagram is a mock implementation of diagram analysis
func (c *Client) understandTechnicalDiagram(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"diagram_type": "architecture_diagram",
		"components": []string{
			"API Gateway",
			"Microservice A",
			"Microservice B",
			"Database",
			"Cache layer",
		},
		"relationships": []string{
			"API Gateway -> Microservice A",
			"API Gateway -> Microservice B",
			"Microservice A -> Database",
			"Microservice B -> Cache",
		},
		"summary": "The diagram shows a microservices architecture with API gateway pattern",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// uiDiffCheck is a mock implementation of UI comparison
func (c *Client) uiDiffCheck(ctx context.Context, args any) (*mcp.CallToolResult, error) {
	response := map[string]any{
		"differences": []string{
			"Button color: expected #007bff, actual #0056b3",
			"Font size: expected 16px, actual 14px",
			"Missing border on input field",
			"Incorrect spacing between elements",
		},
		"similarity_score": 0.87,
		"critical_issues": []string{
			"Missing border affects usability",
			"Font size may impact accessibility",
		},
		"recommendation": "Adjust button color, font size, and add missing border to match design specifications",
	}

	responseBytes, _ := json.Marshal(response)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(responseBytes),
			},
		},
	}, nil
}

// Close is a compatibility method that calls Disconnect
func (c *Client) Close() error {
	return c.Disconnect(context.Background())
}

// ServerInfo returns information about the Z.ai MCP server
func (c *Client) ServerInfo() mcp.Implementation {
	return mcp.Implementation{
		Name:    serverName,
		Version: "1.0.0",
	}
}

// IsVisionTool checks if a tool name is one of the Z.ai vision tools
func IsVisionTool(toolName string) bool {
	visionTools := []string{
		"mcp_vision_analyze_data_visualization",
		"mcp_vision_analyze_image",
		"mcp_vision_extract_text_from_screenshot",
		"mcp_vision_ui_to_artifact",
		"mcp_vision_diagnose_error_screenshot",
		"mcp_vision_understand_technical_diagram",
		"mcp_vision_ui_diff_check",
		"mcp_vision_analyze_video",
	}

	for _, tool := range visionTools {
		if toolName == tool {
			return true
		}
	}
	return false
}

// GetToolDescription returns a user-friendly description for a vision tool
func GetToolDescription(toolName string) string {
	descriptions := map[string]string{
		"mcp_vision_analyze_data_visualization":   "Analyze charts, graphs, and data visualizations to extract insights",
		"mcp_vision_analyze_image":                "General-purpose image analysis for any visual content",
		"mcp_vision_extract_text_from_screenshot": "Extract and recognize text from screenshots using OCR",
		"mcp_vision_ui_to_artifact":               "Convert UI screenshots to code, prompts, or specifications",
		"mcp_vision_diagnose_error_screenshot":    "Diagnose and analyze error messages and stack traces",
		"mcp_vision_understand_technical_diagram": "Analyze technical diagrams, flowcharts, and architecture diagrams",
		"mcp_vision_ui_diff_check":                "Compare two UI screenshots to identify differences",
		"mcp_vision_analyze_video":                "Analyze video content and extract key information",
	}

	if desc, ok := descriptions[toolName]; ok {
		return desc
	}
	return "Z.ai vision analysis tool"
}
