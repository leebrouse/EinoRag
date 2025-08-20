package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
	"github.com/leebrouse/eino/internal/embadding/gemini" // 假设这是你的 gemini embedder 包

	"golang.org/x/time/rate"
)

func main() {
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

	// --- 3. 使用工作池和限速器处理文档 ---
	// 配置参数
	batchSize := 10  // 每批处理的文档数量。保持为 1 可以更精细地控制速率。
	numWorkers := 10 // 并发 worker 的数量。可以适当增加，但速率由限速器控制。
	maxRetries := 5 // 最大重试次数

	// 创建任务和结果通道
	tasks := make(chan []*schema.Document, len(docs)/batchSize+1)
	results := make(chan []*schema.Document, len(docs)/batchSize+1)
	var wg sync.WaitGroup

	// **核心改进：调整限速器**
	// Gemini API 免费版通常有每分钟 60 次的限制。
	// 设置为每 2 秒 1 个请求（即每分钟 30 次），以提供足够的安全边际。
	limiter := rate.NewLimiter(rate.Every(2*time.Second), 1)

	// 启动 worker goroutines
	for i := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			fmt.Printf("Worker %d 启动\n", workerID)
			for batch := range tasks {
				// 在处理每个批次前，等待限速器的许可
				if err := limiter.Wait(ctx); err != nil {
					fmt.Printf("Worker %d 限速器等待失败: %v\n", workerID, err)
					return // 或者 continue，取决于你希望如何处理上下文取消
				}

				fmt.Printf("Worker %d 正在处理一个批次...\n", workerID)
				// 调用带有智能重试逻辑的 transform 函数
				splitted, err := retryTransform(ctx, splitter, batch, maxRetries)
				if err != nil {
					// 如果重试后仍然失败，记录错误并继续处理下一个任务
					fmt.Printf("Worker %d 处理批次失败（已达最大重试次数）: %v\n", workerID, err)
					continue
				}
				results <- splitted
			}
			fmt.Printf("Worker %d 结束\n", workerID)
		}(i)
	}

	// 启动 goroutine 以投递任务
	go func() {
		for i := 0; i < len(docs); i += batchSize {
			end := min(i+batchSize, len(docs))
			tasks <- docs[i:end]
		}
		close(tasks) // 所有任务投递完毕后，关闭 tasks 通道
	}()

	// 启动 goroutine 以等待所有 worker 完成，然后关闭 results 通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// --- 4. 收集并聚合所有结果 ---
	var allChunks []*schema.Document
	for batchResult := range results {
		allChunks = append(allChunks, batchResult...)
	}

	// --- 5. 打印最终结果 ---
	fmt.Println("--- 处理完成 ---")
	if len(allChunks) > 0 {
		fmt.Println("示例 chunk 内容:", allChunks[0].Content)
		fmt.Printf("第一个 chunk 的元数据: %+v\n", allChunks)
	}
	fmt.Printf("所有批次处理成功，总共生成了 %d 个 chunks。\n", len(allChunks))
}

// retryTransform 带有指数退避和智能重试逻辑的 Transform 函数
// 它只在遇到 429 (Too Many Requests) 错误时才进行重试
func retryTransform(ctx context.Context, splitter document.Transformer, batch []*schema.Document, maxRetries int) ([]*schema.Document, error) {
	var result []*schema.Document
	var err error
	backoff := 2 * time.Second // 初始退避时间

	for i := range maxRetries {
		// 调用 splitter 的 Transform 方法
		result, err = splitter.Transform(ctx, batch)
		if err == nil {
			// 成功，直接返回结果
			return result, nil
		}

		// **核心改进：智能错误检查**
		// 检查错误是否为 429 错误。如果不是，则不应重试。
		if strings.Contains(err.Error(), "429") {
			fmt.Printf("检测到 429 错误，将在 %v 后进行第 %d 次重试... 错误: %v\n", backoff, i+1, err)
			time.Sleep(backoff)
			backoff *= 2 // 指数增加退避时间
			continue     // 继续下一次循环以重试
		} else {
			// 如果是其他类型的错误（如 API key 无效，网络问题等），立即返回错误
			return nil, fmt.Errorf("发生不可重试的错误: %w", err)
		}
	}

	// 如果所有重试都失败了，返回最后一次的错误
	return nil, fmt.Errorf("经过 %d 次重试后仍然失败: %w", maxRetries, err)
}
