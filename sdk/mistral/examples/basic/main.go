package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nexora/go-mistral"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set MISTRAL_API_KEY environment variable")
	}

	// Create client
	client := mistral.NewClient(apiKey)

	ctx := context.Background()

	// Create a simple chat completion
	req := mistral.CreateChatCompletionRequestWithDefaults(
		"mistral-small-latest",
		[]interface{}{
			map[string]string{
				"role":    "user",
				"content": "What is the capital of France? Say it in one word.",
			},
		},
	)

	// Enable JSON mode for structured output
	req.EnableJSONMode()

	// Set temperature for creativity
	temp := 0.3
	req.Temperature = &temp

	// Make the request
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	// Print the response
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Usage: %+v\n", resp.Usage)
	if resp.Choices[0].Message.Content != nil {
		fmt.Printf("Response: %s\n", *resp.Choices[0].Message.Content)
	}
}
