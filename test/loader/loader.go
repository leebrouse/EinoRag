package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"github.com/leebrouse/eino/internal/embadding/gemini"

	"golang.org/x/time/rate"
)

func main() {
	ctx := context.Background()

	// 解析 PDF
	p, err := pdf.NewPDFParser(ctx, &pdf.Config{ToPages: true})
	if err != nil {
		panic(err)
	}
	file, err := os.Open("document.pdf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	docs, err := p.Parse(ctx, file,
		parser.WithURI("document.pdf"),
		parser.WithExtraMeta(map[string]any{"source": "./document.pdf"}))
	if err != nil {
		panic(err)
	}

	// 初始化 embedder + splitter
	embedder, err := gemini.NewEmbedder()
	if err != nil {
		panic(err)
	}
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,
		BufferSize:   2,
		MinChunkSize: 100,
		Separators:   []string{"\n", ".", "?", "!", "。", "！", "？"},
		Percentile:   0.9,
	})
	if err != nil {
		panic(err)
	}

	// --- 使用工作池对 docs 分批处理 ---
	batchSize := 1  // 每批处理多少个 doc（调小，降低请求量）
	numWorkers := 1 // 并发 worker 数量（调小，避免 API 打爆）

	tasks := make(chan []*schema.Document)
	results := make(chan []*schema.Document)
	var wg sync.WaitGroup

	// 限速器：每秒 1 个请求
	limiter := rate.NewLimiter(1, 1)

	// worker
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range tasks {
				// 等待限速许可
				if err := limiter.Wait(ctx); err != nil {
					panic(err)
				}

				// 调用带重试的 transform
				splitted, err := retryTransform(ctx, splitter, batch)
				if err != nil {
					fmt.Printf("batch failed after retries: %v\n", err)
					continue
				}
				results <- splitted
			}
		}()
	}

	// 投递任务
	go func() {
		for i := 0; i < len(docs); i += batchSize {
			end := min(i + batchSize, len(docs))
			tasks <- docs[i:end]
		}
		close(tasks)
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(results)
	}()

	// 聚合所有 chunk
	var allChunks []*schema.Document
	for batchResult := range results {
		allChunks = append(allChunks, batchResult...)
	}

	// 打印示例
	if len(allChunks) > 0 {
		fmt.Println("示例 chunk:", allChunks[0].Content)
		fmt.Println("content:", allChunks)
	}
	fmt.Printf("All batches processed successfully, total chunks: %d\n", len(allChunks))
}

// retryTransform 带重试的 Transform（处理 429 错误）
func retryTransform(ctx context.Context, splitter document.Transformer, batch []*schema.Document) ([]*schema.Document, error) {
	var result []*schema.Document
	var err error
	backoff := time.Second
	for i := range 5 { // 最多重试 5 次
		result, err = splitter.Transform(ctx, batch)
		if err == nil {
			return result, nil
		}
		fmt.Printf("retry %d due to error: %v\n", i+1, err)
		time.Sleep(backoff)
		backoff *= 2 // 指数退避
	}
	return nil, err
}
