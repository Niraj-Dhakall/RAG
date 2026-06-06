package main

import (
	"context"

	"rag-project/internal/embeddings"
)

func embeddingGenerator(ctx context.Context, text string) ([]float32, error) {
	return embeddings.Generate(ctx, text)
}
