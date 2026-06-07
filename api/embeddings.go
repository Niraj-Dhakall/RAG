package main

import (
	"context"

	"rag-project/internal/embeddings"
)

// Generates embeddings for a given text. Returns an array of float32 and any errors
func embeddingGenerator(ctx context.Context, text string) ([]float32, error) {
	return embeddings.Generate(ctx, text)
}
