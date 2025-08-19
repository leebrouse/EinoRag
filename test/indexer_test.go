package test

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/eino-ext/components/indexer/milvus" // 换成你的包路径
	"github.com/cloudwego/eino/schema"                        // Document 定义
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

// ---------- 测试：Indexer 端到端写入 ----------
func TestIndexer_Store(t *testing.T) {
	ctx := context.Background()

	// 1. 创建 Milvus 客户端
	addr := viper.GetString("milvus.addr")
	cli, err := client.NewClient(ctx, client.Config{
		Address:  addr,
		Username: "minioadmin",
		Password: "minioadmin",
	})
	require.NoError(t, err)
	defer cli.Close()

	// 2. 创建 embedder
	embedder, err := gemini.NewEmbedder()
	require.NoError(t, err)

	collection := "test_rag_docs"
	// 3. 初始化 indexer
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            cli,
		Embedding:         embedder,
		Collection:        collection, // 独立测试集合，避免污染线上
		MetricType:        "COSINE",
		DocumentConverter: floatDocumentConverter,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{"max_length": "128"},
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "4096"},
			},
			{
				Name:     "metadata",
				DataType: entity.FieldTypeJSON,
			},
			{
				Name:       "vector",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": "3072"},
			},
		},
	})
	require.NoError(t, err)

	// 删除collection
	defer func() {
		_ = cli.DropCollection(ctx, collection)
	}()

	// 4. 构造测试文档
	docs := []*schema.Document{
		{
			ID:      "test-doc-1",
			Content: "Milvus is an open-source vector database.",
			MetaData: map[string]any{
				"source": "test",
			},
		},
		{
			ID:      "test-doc-2",
			Content: "Milvus supports distributed deployment.",
		},
	}

	// 5. 写入并断言
	ids, err := indexer.Store(ctx, docs)
	require.NoError(t, err)
	require.Len(t, ids, 2)
	log.Printf("Stored docs, ids=%v", ids)
}

// ---------- 复用你已有的 converter ----------
func floatDocumentConverter(
	ctx context.Context,
	docs []*schema.Document,
	vectors [][]float64,
) ([]interface{}, error) {
	rows := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		vec32 := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			vec32[j] = float32(v)
		}
		rows = append(rows, map[string]interface{}{
			"id":       doc.ID,
			"content":  doc.Content,
			"vector":   vec32,
			"metadata": doc.MetaData,
		})
	}
	return rows, nil
}
