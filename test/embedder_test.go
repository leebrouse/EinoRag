/*
集成测试说明

前置条件：
 1. 在 internal/config/global.yaml 中设置 Embedding 模型，例如：
    gemini:
    embedder: text-embedding-004
 2. 导出 Google Gemini API Key（genai.NewClient 会自动读取）：
    export GOOGLE_API_KEY="<你的_API_KEY>"

运行命令：

	go test ./test -run TestGeminiEmbedder_Real -v
*/
package test

import (
	"context"
	"log"
	"testing"

	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/spf13/viper"
	"google.golang.org/genai"
)

func TestGeminiEmbedder_Real(t *testing.T) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, nil) // 会自动读取环境变量 API KEY
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	log.Println("Gemini api_key:", viper.GetString("gemini.apikey"))
	log.Println("Gemini embedder:", viper.GetString("gemini.embedder"))

	embedder, err := gemini.NewEmbedder(client, viper.GetString("gemini.embedder"))
	if err != nil {
		t.Fatalf("failed to create embedder: %v", err)
	}

	texts := []string{"hello world"}
	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedStrings failed: %v", err)
	}

	t.Logf("embedding result: %v", vectors)
}
