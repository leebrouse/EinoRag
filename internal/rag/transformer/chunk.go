package transformer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

type Transformer struct {
	client         *milvus.Client
	embedder       embedding.Embedder
	chunkSize      int
	overlap        int
	minChunkLength int
}

func NewTransformer() (document.Transformer, error) {
	chunkSize := viper.GetInt("rag.transformer.chunk_size")
	overlap := viper.GetInt("rag.transformer.overlap")
	minChunkLength := viper.GetInt("rag.transformer.min_chunk_length")

	// 检查配置参数的有效性
	if chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunk_size: %d, must be positive", chunkSize)
	}
	if overlap < 0 {
		return nil, fmt.Errorf("invalid overlap: %d, must be non-negative", overlap)
	}
	if minChunkLength <= 0 {
		return nil, fmt.Errorf("invalid min_chunk_length: %d, must be positive", minChunkLength)
	}
	if overlap >= chunkSize {
		return nil, fmt.Errorf("overlap (%d) must be less than chunk_size (%d)", overlap, chunkSize)
	}

	return &Transformer{
		chunkSize:      chunkSize,
		overlap:        overlap,
		minChunkLength: minChunkLength,
	}, nil
}

func (t *Transformer) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	// 1. 处理 Option
	//return docs, nil
	return nil, nil
}

func (t *Transformer) doTransform(ctx context.Context, src []*schema.Document) ([]*schema.Document, error) {
	// 实现文档转换逻辑
	return nil, nil
}
