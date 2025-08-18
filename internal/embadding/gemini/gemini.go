package gemini

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	_ "github.com/leebrouse/eino/internal/config"
	"google.golang.org/genai"
)

// GeminiEmbedder 实现 Embedder 接口
type GeminiEmbedder struct {
	Client   *genai.Client
	Embedder string
}

// NewGeminiEmbedder 初始化 Gemini Embedder
func NewEmbedder(client *genai.Client, Embedder string) (embedding.Embedder, error) {

	return &GeminiEmbedder{
		Client:   client,
		Embedder: Embedder,
	}, nil
}

// EmbedStrings 将多条文本转换成向量
func (e *GeminiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var contents []*genai.Content
	for _, text := range texts {
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	result, err := e.Client.Models.EmbedContent(ctx,
		e.Embedder,
		contents,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("embedding error: %w", err)
	}

	// 转换成 [][]float64
	embeddings := make([][]float64, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vec := make([]float64, len(emb.Values))
		for j, v := range emb.Values {
			vec[j] = float64(v)
		}
		embeddings[i] = vec
	}

	return embeddings, nil
}
