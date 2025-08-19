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
	client   *genai.Client
	embedder string
}

// NewGeminiEmbedder 初始化 Gemini Embedder
func NewEmbedder(client *genai.Client, Embedder string) (embedding.Embedder, error) {

	return &GeminiEmbedder{
		client:   client,
		embedder: Embedder,
	}, nil
}

// EmbedStrings 将多条文本转换成向量
func (e *GeminiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var contents []*genai.Content
	for _, text := range texts {
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	result, err := e.client.Models.EmbedContent(ctx,
		e.embedder,
		contents,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("embedding error: %w", err)
	}

	// 转换成 [][]float64
	embeddings := e.doEmbed(result)

	return embeddings, nil
}

// change to [][]float64
func (e *GeminiEmbedder) doEmbed(result *genai.EmbedContentResponse) [][]float64 {
	embeddings := make([][]float64, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		vec := make([]float64, len(emb.Values))
		for j, v := range emb.Values {
			vec[j] = float64(v)
		}
		embeddings[i] = vec
	}

	return embeddings
}
