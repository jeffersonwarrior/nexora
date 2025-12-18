package indexer

import (
	"context"
)

// EmbeddingGeneratorAdapter wraps EmbeddingProvider to implement EmbeddingGenerator
type EmbeddingGeneratorAdapter struct {
	provider EmbeddingProvider
}

// NewEmbeddingGeneratorAdapter creates an adapter that wraps an EmbeddingProvider
func NewEmbeddingGeneratorAdapter(provider EmbeddingProvider) *EmbeddingGeneratorAdapter {
	return &EmbeddingGeneratorAdapter{provider: provider}
}

func (a *EmbeddingGeneratorAdapter) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return a.provider.GenerateEmbedding(ctx, text)
}

func (a *EmbeddingGeneratorAdapter) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// Simple batch implementation - call individual method for each text
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := a.provider.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

func (a *EmbeddingGeneratorAdapter) GenerateSymbolEmbeddings(ctx context.Context, symbols []Symbol) ([]Embedding, error) {
	// This would require access to the embedding engine, so for now
	// we'll delegate to a simple implementation
	embeddings := make([]Embedding, 0, len(symbols))

	for _, symbol := range symbols {
		text := createSimpleEmbeddingText(symbol)
		if text == "" {
			continue
		}

		vector, err := a.provider.GenerateEmbedding(ctx, text)
		if err != nil {
			continue
		}

		embedding := Embedding{
			ID:     symbol.Name,
			Type:   symbol.Type,
			Text:   text,
			Vector: vector,
			Metadata: MetaData{
				Package:    symbol.Package,
				File:       symbol.File,
				Line:       symbol.Line,
				Complexity: 1,
				Public:     symbol.Public,
			},
		}
		embeddings = append(embeddings, embedding)
	}

	return embeddings, nil
}

func (a *EmbeddingGeneratorAdapter) Name() string {
	return a.provider.Name()
}

func (a *EmbeddingGeneratorAdapter) ValidateAPIKey(ctx context.Context) bool {
	// For providers that support this (like Mistral), we can validate
	if mistralProv, ok := a.provider.(*MistralProvider); ok {
		return mistralProv.ValidateAPIKey(ctx)
	}
	return true // Assume valid for other providers
}

// Helper function for simple embedding text creation
func createSimpleEmbeddingText(symbol Symbol) string {
	if symbol.Signature != "" {
		return symbol.Signature
	}
	return symbol.Name
}
