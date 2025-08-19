package storage

import (
	"context"
	"fmt"

	milvus"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type Milvus struct {
	Client milvus.Client
}

// connect
// NewMilvus 连接到 Milvus，支持可选用户名密码（留空则不启用鉴权）
func NewMilvus(ctx context.Context, address string, username string, password string) (*Milvus, error) {
	cfg := milvus.Config{
		Address: address,
	}
	if username != "" {
		cfg.Username = username
		cfg.Password = password
	}
	c, err := milvus.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create milvus client failed: %w", err)
	}
	return &Milvus{Client: c}, nil
}

// Close 关闭连接
func (m *Milvus) Close() error {
	if m == nil || m.Client == nil {
		return nil
	}
	return m.Client.Close()
}

// create
// CreateCollectionWithIndex 创建集合（含主键与向量字段）并建立 FLAT 索引，随后加载集合
// - collName: 集合名
// - dim: 向量维度
// - metric: 度量（例如 entity.L2 或 entity.IP）
func (m *Milvus) CreateCollectionWithIndex(ctx context.Context, collName string, dim int, metric entity.MetricType) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("milvus client is nil")
	}

	has, err := m.Client.HasCollection(ctx, collName)
	if err != nil {
		return fmt.Errorf("has collection failed: %w", err)
	}
	if has {
		// 已存在则直接加载
		return m.Client.LoadCollection(ctx, collName, false)
	}

	schema := entity.NewSchema().
		WithName(collName).
		WithDescription("eino demo collection").
		WithAutoID(true)

	pkField := entity.NewField().
		WithName("id").
		WithDataType(entity.FieldTypeInt64).
		WithIsPrimaryKey(true)

	vecField := entity.NewField().
		WithName("vector").
		WithDataType(entity.FieldTypeFloatVector).
		WithDim(int64(dim))

	schema.WithField(pkField).WithField(vecField)

	if err := m.Client.CreateCollection(ctx, schema, 2); err != nil { // 2 个分片，按需调整
		return fmt.Errorf("create collection failed: %w", err)
	}

	// 创建 FLAT 索引（简单直观，适合示例）
	idx, err := entity.NewIndexFlat(metric)
	if err != nil {
		return fmt.Errorf("new flat index failed: %w", err)
	}
	if err := m.Client.CreateIndex(ctx, collName, "vector", idx, false); err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	// 加载集合以便查询/搜索
	if err := m.Client.LoadCollection(ctx, collName, false); err != nil {
		return fmt.Errorf("load collection failed: %w", err)
	}
	return nil
}

// InsertVectors 插入向量，返回系统生成的主键列（AutoID=true 时生效）
// - vectors: 形如 [][]float32，长度需与 dim 一致
func (m *Milvus) InsertVectors(ctx context.Context, collName string, dim int, vectors [][]float32) (entity.Column, error) {
	if m == nil || m.Client == nil {
		return nil, fmt.Errorf("milvus client is nil")
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("vectors is empty")
	}

	colVec := entity.NewColumnFloatVector("vector", dim, vectors)
	ids, err := m.Client.Insert(ctx, collName, "", colVec)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}
	// 刷盘以确保可查询
	if err := m.Client.Flush(ctx, collName, false); err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}
	return ids, nil
}

// research
// SearchTopK 对给定向量执行 TopK 搜索，返回原始搜索结果
func (m *Milvus) SearchTopK(ctx context.Context, collName string, queryVectors [][]float32, vectorField string, metric entity.MetricType, topK int) ([]milvus.SearchResult, error) {
	if m == nil || m.Client == nil {
		return nil, fmt.Errorf("milvus client is nil")
	}
	if len(queryVectors) == 0 {
		return nil, fmt.Errorf("query vectors is empty")
	}

	// 构造 []entity.Vector
	var vs []entity.Vector
	for _, v := range queryVectors {
		vs = append(vs, entity.FloatVector(v))
	}

	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, fmt.Errorf("new search param failed: %w", err)
	}

	results, err := m.Client.Search(
		ctx,
		collName,
		nil,            // partitions
		"",             // expr
		[]string{"id"}, // output fields
		vs,
		vectorField,
		metric,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return results, nil
}

// delete
// DeleteByExpr 通过表达式删除，例如：expr = "id in [1,2,3]"
func (m *Milvus) DeleteByExpr(ctx context.Context, collName string, expr string) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("milvus client is nil")
	}
	return m.Client.Delete(ctx, collName, "", expr)
}

// delete by primary keys
func (m *Milvus) DeleteByPKs(ctx context.Context, collName string, ids []int64) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("milvus client is nil")
	}
	if len(ids) == 0 {
		return nil
	}
	idCol := entity.NewColumnInt64("id", ids)
	return m.Client.DeleteByPks(ctx, collName, "", idCol)
}

// update
// UpsertVectors 根据主键进行 upsert，若不存在则插入
func (m *Milvus) UpsertVectors(ctx context.Context, collName string, dim int, ids []int64, vectors [][]float32) (entity.Column, error) {
	if m == nil || m.Client == nil {
		return nil, fmt.Errorf("milvus client is nil")
	}
	if len(ids) == 0 || len(vectors) == 0 || len(ids) != len(vectors) {
		return nil, fmt.Errorf("ids and vectors must be non-empty and aligned")
	}
	idCol := entity.NewColumnInt64("id", ids)
	vecCol := entity.NewColumnFloatVector("vector", dim, vectors)
	pkCol, err := m.Client.Upsert(ctx, collName, "", idCol, vecCol)
	if err != nil {
		return nil, fmt.Errorf("upsert failed: %w", err)
	}
	if err := m.Client.Flush(ctx, collName, false); err != nil {
		return nil, fmt.Errorf("flush failed: %w", err)
	}
	return pkCol, nil
}
