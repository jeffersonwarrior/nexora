package zai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// VisionResponse represents a response from a vision analysis
type VisionResponse struct {
	Content   string `json:"content"`
	Data      []byte `json:"data,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Type      string `json:"type"`
}

// AnalyzeDataVisualization performs data visualization analysis
func (c *Client) AnalyzeDataVisualization(ctx context.Context, imageSource, analysisFocus, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source":     imageSource,
		"analysis_focus":   analysisFocus,
		"prompt":           prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_analyze_data_visualization", arguments)
}

// AnalyzeImage performs general-purpose image analysis
func (c *Client) AnalyzeImage(ctx context.Context, imageSource, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source": imageSource,
		"prompt":       prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_analyze_image", arguments)
}

// ExtractTextFromScreenshot performs OCR on screenshots
func (c *Client) ExtractTextFromScreenshot(ctx context.Context, imageSource, programmingLanguage, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source":         imageSource,
		"programming_language": programmingLanguage,
		"prompt":               prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_extract_text_from_screenshot", arguments)
}

// UIToArtifact converts UI screenshots to various artifacts
func (c *Client) UIToArtifact(ctx context.Context, imageSource, outputType, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source": imageSource,
		"output_type":  outputType,
		"prompt":       prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_ui_to_artifact", arguments)
}

// DiagnoseErrorScreenshot analyzes error screenshots
func (c *Client) DiagnoseErrorScreenshot(ctx context.Context, imageSource, context, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source": imageSource,
		"context":      context,
		"prompt":       prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_diagnose_error_screenshot", arguments)
}

// UnderstandTechnicalDiagram analyzes technical diagrams
func (c *Client) UnderstandTechnicalDiagram(ctx context.Context, imageSource, diagramType, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"image_source":  imageSource,
		"diagram_type":  diagramType,
		"prompt":        prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_understand_technical_diagram", arguments)
}

// UIDiffCheck compares two UI screenshots
func (c *Client) UIDiffCheck(ctx context.Context, expectedImageSource, actualImageSource, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"expected_image_source": expectedImageSource,
		"actual_image_source":   actualImageSource,
		"prompt":                prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_ui_diff_check", arguments)
}

// AnalyzeVideo performs video content analysis
func (c *Client) AnalyzeVideo(ctx context.Context, videoSource, prompt string) (*VisionResponse, error) {
	arguments := map[string]interface{}{
		"video_source": videoSource,
		"prompt":       prompt,
	}
	
	return c.executeVisionTool(ctx, "mcp_vision_analyze_video", arguments)
}

// executeVisionTool is a helper method to execute vision tools and convert responses
func (c *Client) executeVisionTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*VisionResponse, error) {
	c.logger.Info("executing Z.ai vision tool", "tool", toolName)
	
	result, err := c.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call vision tool %s: %w", toolName, err)
	}
	
	return c.convertMCPResult(result)
}

// convertMCPResult converts MCP CallToolResult to VisionResponse
func (c *Client) convertMCPResult(result *mcp.CallToolResult) (*VisionResponse, error) {
	if len(result.Content) == 0 {
		return &VisionResponse{
			Type:    "text",
			Content: "",
		}, nil
	}
	
	var textContent string
	var data []byte
	var mediaType string
	
	for _, content := range result.Content {
		switch c := content.(type) {
		case *mcp.TextContent:
			textContent += c.Text
		case *mcp.ImageContent:
			data = c.Data
			mediaType = c.MIMEType
		case *mcp.AudioContent:
			data = c.Data
			mediaType = c.MIMEType
		}
	}
	
	response := &VisionResponse{
		Content: textContent,
		Data:    data,
	}
	
	if mediaType != "" {
		response.MediaType = mediaType
		if mediaType == "image/png" || mediaType == "image/jpeg" || mediaType == "image/webp" {
			response.Type = "image"
		} else if mediaType == "audio/mp3" || mediaType == "audio/wav" || mediaType == "audio/ogg" {
			response.Type = "audio"
		} else {
			response.Type = "media"
		}
	} else {
		response.Type = "text"
	}
	
	return response, nil
}

// VisionInput represents input parameters for vision tools
type VisionInput struct {
	ImageSource         string `json:"image_source"`
	VideoSource         string `json:"video_source,omitempty"`
	AnalysisFocus       string `json:"analysis_focus,omitempty"`
	DiagramType         string `json:"diagram_type,omitempty"`
	ProgrammingLanguage string `json:"programming_language,omitempty"`
	OutputType          string `json:"output_type,omitempty"`
	Context             string `json:"context,omitempty"`
	Prompt              string `json:"prompt"`
	ExpectedImageSource string `json:"expected_image_source,omitempty"`
	ActualImageSource   string `json:"actual_image_source,omitempty"`
}

// ParseVisionInput parses JSON input into VisionInput struct
func ParseVisionInput(jsonInput string) (*VisionInput, error) {
	var input VisionInput
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		return nil, fmt.Errorf("failed to parse vision input: %w", err)
	}
	return &input, nil
}

// ValidateVisionInput validates the input for vision tools
func ValidateVisionInput(input *VisionInput, requiredType string) error {
	switch requiredType {
	case "image":
		if input.ImageSource == "" {
			return fmt.Errorf("image_source is required")
		}
	case "video":
		if input.VideoSource == "" {
			return fmt.Errorf("video_source is required")
		}
	case "diff":
		if input.ExpectedImageSource == "" || input.ActualImageSource == "" {
			return fmt.Errorf("expected_image_source and actual_image_source are required for UI diff comparison")
		}
	}
	
	if input.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	
	return nil
}

// GetRequiredInputType returns the required input type for a given tool
func GetRequiredInputType(toolName string) string {
	switch toolName {
	case "mcp_vision_analyze_video":
		return "video"
	case "mcp_vision_ui_diff_check":
		return "diff"
	default:
		return "image"
	}
}