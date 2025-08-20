package test

import (
	"context"
	"testing"

	_ "github.com/leebrouse/eino/internal/config" // 仅用于加载全局配置
	retriever "github.com/leebrouse/eino/internal/rag/retriever"
)

// TestRetriever_Real 针对 Retriever 进行端到端集成测试：
// 1. 通过 NewRetriever 读取 viper 配置并创建实例；
// 2. 向已存在的 Milvus collection 发起一次真实检索；
// 3. 断言检索成功并打印结果，方便本地调试。
func TestRetriever_Real(t *testing.T) {
	// 1. 构造 Retriever
	r, err := retriever.NewRetriever()
	if err != nil {
		t.Fatalf("failed to create retriever: %v", err)
	}

	// 2. 执行查询
	ctx := context.Background()
	docs, err := r.Retrieve(ctx, "Milvus")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	// 3. 输出结果（测试日志）
	t.Logf("retrieved docs: %+v", docs)
}
