package myretriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	_ "github.com/leebrouse/eino/internal/config"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"

	"github.com/leebrouse/eino/internal/embadding/gemini"
)

type Retriever struct {
	cli        client.Client      // milvus 原生 client
	embedder   embedding.Embedder // gemini
	collection string             // 对应 index
	topK       int
}

// NewRetriever 读取 viper 配置，创建 Retriever
func NewRetriever() (retriever.Retriever, error) {
	// 1. 连接 milvus
	ctx := context.Background()
	cli, err := client.NewClient(ctx, client.Config{
		Address:  viper.GetString("milvus.addr"),
		Username: viper.GetString("milvus.username"),
		Password: viper.GetString("milvus.password"),
	})
	if err != nil {
		return nil, fmt.Errorf("milvus connect: %w", err)
	}

	// 2. 创建 embedder
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

// Retrieve 实现 retriever.Retriever 接口
func (r *Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// 组装最终配置（默认 + 调用者 Option）
	options := &retriever.Options{
		TopK: &r.topK,
	}

	return r.doRetrieve(ctx, []string{query}, options)
}

// doRetrieve 真正干活的地方
func (r *Retriever) doRetrieve(ctx context.Context, query []string, opt *retriever.Options) ([]*schema.Document, error) {
	// 1. 文本 -> 向量
	vec, err := r.embedder.EmbedStrings(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}

	// 将 []float64 转换为 []float32
	floatVec := make([]float32, len(vec[0]))
	for i, v := range vec[0] {
		floatVec[i] = float32(v)
	}

	// 2. 构造搜索参数
	topK := r.topK
	if opt.TopK != nil && *opt.TopK > 0 {
		topK = *opt.TopK
	}

	// 3. 执行搜索
	sp, _ := entity.NewIndexFlatSearchParam()
	searchRes, err := r.cli.Search(
		ctx,
		r.collection,
		[]string{},
		"",
		[]string{"id", "content", "metadata"},
		[]entity.Vector{entity.FloatVector(floatVec)},
		"vector",
		entity.COSINE,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("milvus search: %w", err)
	}

	// 4. 把 SDK 结果转成 []*schema.Document
	if len(searchRes) == 0 {
		return nil, nil
	}

	// 只处理第一个向量（单 query）
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
