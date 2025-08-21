package retriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"

	"github.com/leebrouse/eino/internal/embadding/gemini"
)

// Retriever wraps a Milvus client and an embedder (Gemini)
// to perform vector similarity search.
type Retriever struct {
	cli        milvusClient.Client // Native Milvus client
	embedder   embedding.Embedder  // Gemini embedder
	collection string              // Milvus collection name (index)
	topK       int                 // Default top K results
}

// NewRetriever reads config from viper and creates a new Retriever
func NewRetriever() (retriever.Retriever, error) {
	// 1. Connect to Milvus
	ctx := context.Background()
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  viper.GetString("milvus.addr"),
		Username: viper.GetString("milvus.username"),
		Password: viper.GetString("milvus.password"),
	})
	if err != nil {
		return nil, fmt.Errorf("milvus connect: %w", err)
	}

	// 2. Create Gemini embedder
	emb, err := gemini.NewEmbedder()
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}

	return &Retriever{
		cli:        cli,
		embedder:   emb,
		collection: viper.GetString("milvus.collection"),
		topK:       viper.GetInt("rag.retriever.topk"),
	}, nil
}

// Retrieve implements retriever.Retriever interface
// It embeds the query and searches top-K similar documents in Milvus.
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// Merge options (defaults + user-provided options)
	options := &retriever.Options{
		TopK: &r.topK,
	}

	// Delegate to internal method
	return r.doRetrieve(ctx, []string{query}, options)
}

// doRetrieve does the actual retrieval work
func (r *Retriever) doRetrieve(ctx context.Context, query []string, opt *retriever.Options) ([]*schema.Document, error) {
	// 1. Convert text query -> vector
	vec, err := r.embedder.EmbedStrings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}

	// Convert []float64 -> []float32 (Milvus requires float32 vectors)
	floatVec := make([]float32, len(vec[0]))
	for i, v := range vec[0] {
		floatVec[i] = float32(v)
	}

	// 2. Determine topK (default or user override)
	topK := r.topK
	if opt.TopK != nil && *opt.TopK > 0 {
		topK = *opt.TopK
	}

	// 3. Execute Milvus search
	sp, _ := entity.NewIndexFlatSearchParam() // flat index search param
	searchRes, err := r.cli.Search(
		ctx,
		r.collection,                          // collection name
		[]string{},                            // partition names (empty = all)
		"",                                    // filter expression (none here)
		[]string{"id", "content", "metadata"}, // fields to return
		[]entity.Vector{entity.FloatVector(floatVec)}, // query vector
		"vector",      // vector field name
		entity.COSINE, // similarity metric
		topK,          // number of results
		sp,            // search parameters
	)
	if err != nil {
		return nil, fmt.Errorf("milvus search: %w", err)
	}

	// 4. Convert search results to []*schema.Document
	if len(searchRes) == 0 {
		return nil, nil
	}

	// Only handle first vector (single query case)
	res := searchRes[0]
	docs := make([]*schema.Document, 0, res.ResultCount)

	for i := 0; i < res.ResultCount; i++ {
		id, _ := res.Fields.GetColumn("id").GetAsInt64(i)
		content, _ := res.Fields.GetColumn("content").GetAsString(i)
		metaRaw, _ := res.Fields.GetColumn("metadata").Get(i)
		metadata, _ := metaRaw.(map[string]any)

		docs = append(docs, &schema.Document{
			ID:       fmt.Sprintf("%d", id),
			Content:  content,
			MetaData: metadata,
		})
	}
	return docs, nil
}
