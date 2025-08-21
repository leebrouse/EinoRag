package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/rag/generator/generating"
	customRetriever "github.com/leebrouse/eino/internal/rag/generator/retriever"
	"github.com/spf13/viper"
	"google.golang.org/genai"
)

// `generator` 基于向量数据库检索结果，智能精炼并重组上下文，为 Chatbox 实时生成高匹配提示词。
type Generator struct {
	client    *genai.Client
	retriever retriever.Retriever
	model     string
}

func NewGenerator() (generating.Generator, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  viper.GetString("gemini.apikey"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// 这里需要创建一个 retriever 实例，并返回 Generator 实例
	r, err := customRetriever.NewRetriever()
	if err != nil {
		return nil, fmt.Errorf("failed to create retriever: %w", err)
	}

	return &Generator{
		client:    client,
		retriever: r,
		model:     viper.GetString("gemini.model"),
	}, nil
}

func (g *Generator) Generate(ctx context.Context, query string) (string, error) {
	// 1. 调用 Retriever 获取候选文档
	researchResults, err := g.retriever.Retrieve(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve: %w", err)
	}

	// 2. 整理文档结果
	chunks, err := assembleResults(researchResults)
	if err != nil {
		return "", fmt.Errorf("failed to assemble results: %w", err)
	}

	// 3. 拼接 Prompt 给 Gemini
	newPrompt := fmt.Sprintf(`
You are an intelligent assistant. 
User query: %s

Here are retrieved context documents:
%s

Please refine and reorganize the context into a concise, high-quality response.
`, query, strings.Join(chunks, "\n---\n"))

	// 4. 调用 Gemini 模型生成
	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(newPrompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to call gemini: %w", err)
	}

	// 5. 提取 Gemini 输出
	if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
		return "", fmt.Errorf("empty response from gemini")
	}
	output := result.Text()

	return output, nil
}

func assembleResults(results []*schema.Document) ([]string, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to assemble")
	}

	chunks := make([]string, 0, len(results))
	for i, doc := range results {
		if doc == nil {
			continue
		}
		// 假设 schema.Document 有 Content
		text := fmt.Sprintf("Result %d:\n%s", i+1, doc.Content)
		if doc.Content != "" && len(doc.Content) > 0 {
			text = fmt.Sprintf("%s\nContent: %+v", text, doc.Content)
		}
		chunks = append(chunks, text)
	}

	return chunks, nil
}
