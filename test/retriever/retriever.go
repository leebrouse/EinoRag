package main

import (
	"context"
	"fmt"
	"log"

	myretriever "github.com/leebrouse/eino/internal/rag/generator/retriever"
)

func main() {
	// Get the environment variables

	// Create a client
	// 2. 创建 retriever
	r, err := myretriever.NewRetriever()
	if err != nil {
		log.Fatalf("new retriever: %v", err)
	}

	// 3. 查询
	docs, err := r.Retrieve(context.Background(), "milvus")
	if err != nil {
		log.Fatalf("retrieve: %v", err)
	}

	// 4. 打印结果
	for _, d := range docs {
		fmt.Printf("\ncontent=%q\n", d.Content)
	}
}
