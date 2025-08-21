package einorag

import (
	"context"
	"fmt"

	"github.com/leebrouse/eino/internal/rag/generator"
	"github.com/leebrouse/eino/internal/rag/generator/generating"
	"github.com/leebrouse/eino/internal/rag/uploader"
	"github.com/leebrouse/eino/internal/rag/uploader/uploading"
)

type EinoRag struct {
	generator generating.Generator
	uploader  uploading.Uploader
}

func NewRagClient() (RAG, error) {
	// 创建 generator
	gen, err := generator.NewGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	// 创建 uploader
	up, err := uploader.NewUploader()
	if err != nil {
		return nil, fmt.Errorf("failed to create uploader: %w", err)
	}

	// 返回 RagClient 实例
	return &EinoRag{
		generator: gen,
		uploader:  up,
	}, nil
}

// Query 调用 generator 生成答案
func (e *EinoRag) Query(ctx context.Context, prompt string) (string, error) {
	resp, err := e.generator.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}
	return resp, nil
}

// Upload 调用 uploader 进行文档上传 + 索引
func (e *EinoRag) Upload(ctx context.Context, fileUrl string) ([]string, error) {
	ids, err := e.uploader.Upload(ctx, fileUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	return ids, nil
}
