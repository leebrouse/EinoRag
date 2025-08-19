package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/leebrouse/eino/internal/config"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
)

func main() {
	// Get the environment variables
	addr := viper.GetString("milvus.addr")
	username := viper.GetString("milvus.username")
	password := viper.GetString("milvus.password")
	collection := viper.GetString("milvus.collection")

	// Create a client
	ctx := context.Background()
	cli, err := client.NewClient(ctx, client.Config{
		Address:  addr,
		Username: username,
		Password: password,
		DBName:   "default",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}
	defer cli.Close()

	// Create an embedding model
	emb, err := gemini.NewEmbedder()

	// Create a retriever
	retriever, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      cli,
		Collection:  collection,
		OutputFields: []string{
			"id",
			"content",
			"metadata",
		},
		TopK:              0,
		ScoreThreshold:    5,
		Embedding:         emb,
	})
	if err != nil {
		log.Fatalf("Failed to create retriever: %v", err)
		return
	}

	// Retrieve documents
	documents, err := retriever.Retrieve(ctx, "milvus")
	if err != nil {
		log.Fatalf("Failed to retrieve: %v", err)
		return
	}

	// Print the documents
	for i, doc := range documents {
		fmt.Printf("Document %d:\n", i)
		fmt.Printf("title: %s\n", doc.ID)
		fmt.Printf("content: %s\n", doc.Content)
		fmt.Printf("metadata: %v\n", doc.MetaData)
	}
}
