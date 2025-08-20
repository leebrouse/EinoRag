package transformer

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/leebrouse/eino/internal/embadding/gemini"
	workerpool "github.com/leebrouse/eino/pkg/wokerpool"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// Transformer is responsible for splitting documents into chunks and embedding them.
type Transformer struct {
	embedder     embedding.Embedder // Embedding engine (e.g., Gemini)
	bufferSize   int                // Size of the buffer for chunking
	minChunkSize int                // Minimum chunk size
	percentile   float64            // Percentile threshold for chunking
}

// NewTransformer creates a new Transformer with configuration from viper
func NewTransformer() (document.Transformer, error) {
	// Load configuration values
	bufferSize := viper.GetInt("rag.transformer.bufferSize")
	minChunkSize := viper.GetInt("rag.transformer.minChunkSize")
	percentile := viper.GetFloat64("rag.transformer.percentile")

	// Initialize the embedder
	emb, err := gemini.NewEmbedder()
	if err != nil {
		return nil, fmt.Errorf("create embedder: %w", err)
	}

	// Validate configuration parameters
	if bufferSize <= 0 {
		return nil, fmt.Errorf("invalid chunk_size: %d, must be positive", bufferSize)
	}
	if minChunkSize < 0 {
		return nil, fmt.Errorf("invalid overlap: %d, must be non-negative", minChunkSize)
	}
	if percentile <= 0 {
		return nil, fmt.Errorf("invalid min_chunk_length: %f, must be positive", percentile)
	}

	// Return the initialized Transformer
	return &Transformer{
		embedder:     emb,
		bufferSize:   bufferSize,
		minChunkSize: minChunkSize,
		percentile:   percentile,
	}, nil
}

// Transform splits documents into chunks, embeds them, and returns the processed documents
func (t *Transformer) Transform(ctx context.Context, src []*schema.Document, opts ...document.TransformerOption) ([]*schema.Document, error) {
	// Initialize a semantic splitter with embedding and chunking configuration
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    t.embedder,
		BufferSize:   t.bufferSize,
		MinChunkSize: t.minChunkSize,
		Percentile:   t.percentile,
		Separators:   []string{"\n", ".", "?", "!", "。", "！", "？"}, // sentence separators
	})
	if err != nil {
		return nil, fmt.Errorf("fail to init splitter: %w", err)
	}

	// Initialize a rate limiter to prevent API exhaustion
	limiter := rate.NewLimiter(rate.Every(2*time.Second), 1) // 1 request every 2 seconds

	// Create a worker pool to process documents concurrently
	pool := workerpool.NewWorkerPool(splitter, limiter)

	// Generate tasks for the worker pool based on the input documents
	pool.GenerateTasks(src)

	// Run the worker pool to perform chunking and embedding
	pool.Run(ctx)

	// Collect and return all processed chunks from the worker pool
	return pool.AssembleChunks(), nil
}
