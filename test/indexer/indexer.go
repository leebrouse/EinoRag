package main

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

func main() {
	addr := viper.GetString("milvus.addr")
	username := "minioadmin"
	password := "minioadmin"

	ctx := context.Background()
	cli, err := client.NewClient(ctx, client.Config{
		Address:  addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	emb, err := gemini.NewEmbedder()
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:     cli,
		Embedding:  emb,
		Collection: "rag_docs",
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
				DataType:   entity.FieldTypeBinaryVector,
				TypeParams: map[string]string{"dim": "3072"},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}
	log.Printf("Indexer created success")

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

	ids, err := indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("Failed to store: %v", err)
	}
	log.Printf("Store success, ids: %v", ids)
}
