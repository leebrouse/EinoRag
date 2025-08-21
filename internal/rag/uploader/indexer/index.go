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
	"github.com/leebrouse/eino/internal/rag/uploader/indexer/field"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/spf13/viper"
)

// Indexer wraps a Milvus client and an embedding engine (e.g., Gemini) for indexing documents
type Indexer struct {
	client   milvusClient.Client // Native Milvus client
	embedder embedding.Embedder  // Gemini embedder
}

// NewIndexer creates a new Indexer instance
func NewIndexer() (indexer.Indexer, error) {
	ctx := context.Background()

	// Connect to Milvus server
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  viper.GetString("milvus.addr"),
		Username: viper.GetString("milvus.username"),
		Password: viper.GetString("milvus.password"),
	})
	if err != nil {
		return nil, fmt.Errorf("milvus connect: %w", err)
	}

	// Create Gemini embedder
	embedder, err := gemini.NewEmbedder()
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}

	return &Indexer{
		client:   cli,
		embedder: embedder,
	}, nil
}

// Store stores documents into Milvus index
func (i *Indexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	// Delegate to the actual implementation
	return i.doStore(ctx, docs)
}

// doStore handles the actual storage process
func (i *Indexer) doStore(ctx context.Context, docs []*schema.Document) (ids []string, err error) {
	// Create a Milvus Indexer instance
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:            i.client,
		Embedding:         i.embedder,
		Collection:        viper.GetString("milvus.collection"),
		MetricType:        "COSINE",               // Cosine similarity metric
		DocumentConverter: floatDocumentConverter, // Converter for float64 -> float32
		Fields:            field.NewFields(nil),   // Define Milvus fields
	})
	if err != nil {
		log.Fatalf("Failed to create indexer: %v", err)
	}

	log.Printf("Indexer created successfully")

	// Store documents
	ids, err = indexer.Store(ctx, docs)
	if err != nil {
		log.Fatalf("Failed to store documents: %v", err)
	}
	log.Printf("Documents stored successfully, ids: %v", ids)

	return ids, nil
}

// floatDocumentConverter converts document embeddings from float64 -> float32
// and prepares rows in the format expected by Milvus
func floatDocumentConverter(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
	rows := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		// Convert vector from float64 to float32
		float32Vec := make([]float32, len(vectors[i]))
		for j, v := range vectors[i] {
			float32Vec[j] = float32(v)
		}

		// Prepare a row map for Milvus
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
