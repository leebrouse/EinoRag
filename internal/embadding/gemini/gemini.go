package gemini

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/spf13/viper"
	"google.golang.org/genai"
)

// GeminiEmbedder 实现 Embedder 接口
type GeminiEmbedder struct {
	client   *genai.Client
	embedder string
}

// NewGeminiEmbedder 初始化 Gemini Embedder
func NewEmbedder() (embedding.Embedder, error) {

	apikey := viper.GetString("gemini.apikey")
	embedder := viper.GetString("gemini.embedder")
	if embedder == "" {
		return nil, fmt.Errorf("gemini embedder not configured")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apikey,
		Backend: genai.BackendGeminiAPI,
	}) 
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &GeminiEmbedder{
		client:   client,
		embedder: embedder,
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

// doEmbed 将 genai.EmbedContentResponse 转换为 [][]float64 格式
func (e *GeminiEmbedder) doEmbed(result *genai.EmbedContentResponse) [][]float64 {
	if result == nil || len(result.Embeddings) == 0 {
		return [][]float64{}
	}

	embeddings := make([][]float64, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		if emb == nil || len(emb.Values) == 0 {
			embeddings[i] = []float64{}
			continue
		}

		vec := make([]float64, len(emb.Values))
		for j, v := range emb.Values {
			vec[j] = float64(v)
		}
		embeddings[i] = vec
	}

	return embeddings
}
