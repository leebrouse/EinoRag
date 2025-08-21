package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/leebrouse/eino/internal/embadding/gemini"
)

func main() {
	start := time.Now()
	ctx := context.Background()

	// --- 1. 解析 PDF ---
	// 初始化 PDF 解析器
	p, err := pdf.NewPDFParser(ctx, &pdf.Config{ToPages: true})
	if err != nil {
		panic(fmt.Errorf("初始化 PDF 解析器失败: %w", err))
	}

	// 打开 PDF 文件
	file, err := os.Open("document.pdf")
	if err != nil {
		panic(fmt.Errorf("打开 'document.pdf' 文件失败: %w", err))
	}
	defer file.Close()

	// 执行解析
	docs, err := p.Parse(ctx, file,
		parser.WithURI("document.pdf"),
		parser.WithExtraMeta(map[string]any{"source": "./document.pdf"}),
	)
	if err != nil {
		panic(fmt.Errorf("解析 PDF 失败: %w", err))
	}
	fmt.Printf("PDF 解析完成，共得到 %d 个文档页面。\n", len(docs))

	// --- 2. 初始化 Embedder 和 Splitter ---
	// 初始化 Gemini embedder
	embedder, err := gemini.NewEmbedder()
	if err != nil {
		panic(fmt.Errorf("初始化 Gemini embedder 失败: %w", err))
	}

	// 初始化语义分割器
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,
		BufferSize:   2,
		MinChunkSize: 100,
		Separators:   []string{"\n", ".", "?", "!", "。", "！", "？"},
		Percentile:   0.9,
	})
	if err != nil {
		panic(fmt.Errorf("初始化语义分割器失败: %w", err))
	}

	result, _ := splitter.Transform(ctx, docs)
	fmt.Println(result)
	fmt.Printf("Total elapsed time: %v\n", time.Since(start))
}
