package embeddings

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"google.golang.org/genai"
)

func Generate(ctx context.Context, text string) ([]float32, error) {
	switch provider := os.Getenv("EMBEDDING_PROVIDER"); provider {
	case "openai":
		client := openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
		result, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
			Model: openai.EmbeddingModelTextEmbedding3Small,
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(text),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("openai embed: %w", err)
		}
		return toFloat32(result.Data[0].Embedding), nil

	case "gemini":
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  os.Getenv("GEMINI_API_KEY"),
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("gemini client: %w", err)
		}
		result, err := client.Models.EmbedContent(ctx, "gemini-embedding-001", []*genai.Content{
			genai.NewContentFromText(text, genai.RoleUser),
		}, &genai.EmbedContentConfig{
			OutputDimensionality: genai.Ptr(int32(1536)),
		})
		if err != nil {
			return nil, fmt.Errorf("gemini embed: %w", err)
		}
		return result.Embeddings[0].Values, nil

	default:
		return nil, fmt.Errorf("unknown EMBEDDING_PROVIDER: %q", provider)
	}
}

func toFloat32(v []float64) []float32 {
	result := make([]float32, len(v))
	for i, x := range v {
		result[i] = float32(x)
	}
	return result
}
