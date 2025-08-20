package main

import (
	"fmt"
	"log"
	"math"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/net/context"
	"google.golang.org/genai"
)

type GeminiEmbedder struct {
	Client *genai.Client
}

func NewGeminiEmbedder() embedding.Embedder {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return &GeminiEmbedder{
		Client: client,
	}
}

// cosineSimilarity calculates the similarity between two vectors.
func cosineSimilarity(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same length")
	}

	var dotProduct, aMagnitude, bMagnitude float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		aMagnitude += float64(a[i] * a[i])
		bMagnitude += float64(b[i] * b[i])
	}

	if aMagnitude == 0 || bMagnitude == 0 {
		return 0, nil
	}

	return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude)), nil
}

// EmbedStrings 将多条文本转换成向量
func (e *GeminiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	var contents []*genai.Content
	for _, text := range texts {
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	result, err := e.Client.Models.EmbedContent(ctx,
		"text-embedding-004",
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

	// 如果你还想计算相似度，可以放在这里
	for i := 0; i < len(texts); i++ {
		for j := i + 1; j < len(texts); j++ {
			similarity, _ := cosineSimilarity(
				result.Embeddings[i].Values,
				result.Embeddings[j].Values,
			)
			fmt.Printf("Similarity between '%s' and '%s': %.4f\n",
				texts[i], texts[j], similarity)
		}
	}

	return embeddings, nil
}

func main() {
	ctx := context.Background()

	// 初始化嵌入器（示例使用）
	embedder := NewGeminiEmbedder()

	//embedder.EmbedStrings()
	// 初始化分割器
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,
		BufferSize:   2,
		MinChunkSize: 100,
		Separators:   []string{"\n", "。", ".","?", "!"},
		Percentile:   0.9,
	})
	if err != nil {
		panic(err)
	}

	// 准备要分割的文档
	docs := []*schema.Document{
		{
			ID:      "doc1",
			Content: `雅思阅读考点词 库—引用于刘洪波《剑桥雅思阅读考点词真经》538 个雅思阅读考点词，分成了 3 类，进行了重要性排序。`,
		},
	}

	// 执行分割
	results, err := splitter.Transform(ctx, docs)
	if err != nil {
		panic(err)
	}

	// log.Println(results)

	// 处理分割结果
	for i, doc := range results {
		println("片段", i+1, ":", doc.Content)
	}
}
