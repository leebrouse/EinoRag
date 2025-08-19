package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var collection = "rag_doc"

var fields = []*entity.Field{
	{
		Name:     "id",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "255",
		},
		PrimaryKey: true,
	},
	{
		Name:     "vector", // 确保字段名匹配
		DataType: entity.FieldTypeFloatVector,
		TypeParams: map[string]string{
			"dim": "3072",
		},
	},
	{
		Name:     "content",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length": "8192",
		},
	},
	{
		Name:     "metadata",
		DataType: entity.FieldTypeJSON,
	},
}

func IndexerRAG(docs []*schema.Document) {
	ctx := context.Background()
	// 初始化嵌入器
	embedder, err := gemini.NewEmbedder()
	if err != nil {
		panic(err)
	}

	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            MilvusCli,
		Collection:        collection,
		Fields:            fields,
		Embedding:         embedder,
		MetricType: "COSINE",
		DocumentConverter: floatDocumentConverter,
	})
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}
	for _, doc := range docs {
		storeDoc := []*schema.Document{
			{
				ID:       doc.ID,
				Content:  doc.Content,
				MetaData: doc.MetaData,
			},
		}
		ids, err := indexer.Store(ctx, storeDoc)
		if err != nil {
			log.Fatalf("Failed to store documents: %v", err)
		}
		println("Stored documents with IDs: %v", ids)
	}
}

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

func main() {

	docs := []*schema.Document{
		{
			ID:      "milvus-1",
			Content: "milvus is an open-source vector database",
			MetaData: map[string]any{
				"h1": "milvus",
				"h2": "open-source",
				"h3": "vector database",
			},
		},
		{
			ID:       "milvus-2",
			Content:  "milvus is a distributed vector database",
			MetaData: map[string]any{}, // 保持 map 类型即可
		},
	}

	IndexerRAG(docs)
}
