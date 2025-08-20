package indexer

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/leebrouse/eino/internal/rag/indexer/field"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

type Indexer struct {
	client   milvusClient.Client // milvus 原生 client
	embedder embedding.Embedder  // gemini
}

func NewIndexer() (indexer.Indexer, error) {
	ctx := context.Background()
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  viper.GetString("milvus.addr"),
		Username: viper.GetString("milvus.username"),
		Password: viper.GetString("milvus.password"),
	})
	if err != nil {
		return nil, fmt.Errorf("milvus connect: %w", err)
	}

	// 2. 创建 embedder
	embedder, err := gemini.NewEmbedder()
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}

	return &Indexer{
		client:   cli,
		embedder: embedder,
	}, nil
}

// Store 用于存储文档。
func (i *Indexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	// TODO: 实现文档存储逻辑
	return i.doStore(ctx, docs)
}

func (i *Indexer) doStore(ctx context.Context, docs []*schema.Document) (ids []string, err error) {

	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            i.client,
		Embedding:         i.embedder,
		Collection:        viper.GetString("milvus.collection"),
		MetricType:        "COSINE",
		DocumentConverter: floatDocumentConverter,
		Fields:            field.NewFields(nil),
	})
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}

	log.Printf("Indexer created success")

	ids, err = indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("Failed to store: %v", err)
	}
	log.Printf("Store success, ids: %v", ids)

	return ids, nil
}

// DocumentConverter: float64 -> float32
func floatDocumentConverter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	rows := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		// float64 -> float32
		float32Vec := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			float32Vec[j] = float32(v)
		}
		row := map[string]interface{}{
			"id":       doc.ID,
			"content":  doc.Content,
			"vector":   float32Vec,
			"metadata": doc.MetaData,
		}
		rows = append(rows, row)
	}
	return rows, nil
}
