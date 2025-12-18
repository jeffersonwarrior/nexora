package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexora/sdk/base"
	"github.com/nexora/sdk/openai"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENAI_API_KEY environment variable")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	ctx := context.Background()

	// Create a simple chat completion request
	req := openai.CreateChatCompletionRequest(
		"gpt-4",
		[]base.Message{
			openai.UserMessage{
				Content: "What is the capital of France? Answer in one word.",
				Role:    "user",
			},
		},
	)

	// Set temperature for focused responses
	req.SetTemperature(0.3)

	// Enable streaming
	req.EnableStreaming()

	// Make the request
	chunkCh, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	fmt.Println("Response:")
	for chunk := range chunkCh {
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != nil {
			fmt.Print(*chunk.Choices[0].Delta.Content)
		}
	}
	fmt.Println()

	// Example with JSON mode
	req2 := openai.CreateChatCompletionRequest(
		"gpt-4",
		[]base.Message{
			openai.UserMessage{
				Content: "List the first 5 elements as a JSON array",
				Role:    "user",
			},
		},
	)

	req2.EnableJSONSchemaMode(
		"capitals",
		"Array of capital cities",
		map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"minItems": 1,
			"maxItems": 5,
		},
		true,
	)

	resp2, err := client.CreateChatCompletion(ctx, req2)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	fmt.Printf("JSON Response: %s\n", resp2.Choices[0].Message.Content)
	fmt.Printf("Usage: %+v\n", resp2.Usage)
}

// Advanced example with tool calling
func exampleWithTools() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return
	}

	client := openai.NewClient(apiKey)

	// Define a calculator tool
	calculatorTool := base.Tool{
		Type: "function",
		Function: base.Function{
			Name:        "calculator",
			Description: "Perform a calculation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type": "string",
						"enum": []string{"add", "subtract", "multiply", "divide"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
					"required": []string{"operation", "a", "b"},
				},
			},
		},
	}

	// Create request with tool
	req := openai.CreateChatCompletionRequest(
		"gpt-4",
		[]base.Message{
			openai.UserMessage{
				Content: "What is 25 * 17?",
				Role:    "user",
			},
		},
	)
	req.AddTool(calculatorTool)
	req.SetToolChoice("auto")

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)

	// Check if tool was called
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		fmt.Printf("Tool called: %s\n", resp.Choices[0].Message.ToolCalls[0].Function.Name)
		fmt.Printf("Arguments: %s\n", resp.Choices[0].Message.ToolCalls[0].Function.Arguments)
	}
}
