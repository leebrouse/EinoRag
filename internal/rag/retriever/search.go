package retriever

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"
)

type Retriever struct {
	client   *milvus.Client
	embedder embedding.Embedder
	index    string
	topK     int
}

func NewRetriever() (retriever.Retriever, error) {

	// gemini.NewEmbedder()

	// return &MyRetriever{
	// 	embedder: ,
	// 	index:    config.Index,
	// 	topK:     config.DefaultTopK,
	// }, nil

	return nil, nil
}

func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 1. 处理选项
	// options := &retriever.Options{
	// 	Index:     &r.index,
	// 	TopK:      &r.topK,
	// 	Embedding: r.embedder,
	// }

	return nil, nil
}

func (r *Retriever) doRetrieve(ctx context.Context, query string, opts *retriever.Options) ([]*schema.Document, error) {
	return nil, nil
}
