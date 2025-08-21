package uploader

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/indexer"
	customIndexer "github.com/leebrouse/eino/internal/rag/uploader/indexer"
	"github.com/leebrouse/eino/internal/rag/uploader/loader"
	"github.com/leebrouse/eino/internal/rag/uploader/transformer"
	"github.com/leebrouse/eino/internal/rag/uploader/uploading"
)

type Uploader struct {
	loader      document.Loader
	transformer document.Transformer
	indexer     indexer.Indexer
}

func NewUploader() (uploading.Uploader, error) {
	// 创建 loader
	loader, err := loader.NewLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create loader: %w", err)
	}

	// 创建 transformer
	transformer, err := transformer.NewTransformer()
	if err != nil {
		return nil, fmt.Errorf("failed to create transformer: %w", err)
	}

	// 创建 indexer
	indexer, err := customIndexer.NewIndexer()
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	// 返回 Uploader 实例
	return &Uploader{
		loader:      loader,
		transformer: transformer,
		indexer:     indexer,
	}, nil
}

func (u *Uploader) Upload(ctx context.Context, fileUrl string) ([]string, error) {
	// 1. loader: 从文件加载文档
	docs, err := u.loader.Load(ctx, document.Source{URI: fileUrl})
	if err != nil {
		return nil, fmt.Errorf("failed to load document from %s: %w", fileUrl, err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents loaded from %s", fileUrl)
	}

	// 2. transformer: 对文档进行分块 / 转换
	chunkDocs, err := u.transformer.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to transform documents: %w", err)
	}
	if len(chunkDocs) == 0 {
		return nil, fmt.Errorf("transformer returned empty chunks")
	}

	// 3. indexer: 将分块文档存储到向量数据库
	ids, err := u.indexer.Store(ctx, chunkDocs)
	if err != nil {
		return nil, fmt.Errorf("failed to index documents: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("indexer did not return any IDs")
	}

	return ids, nil
}
