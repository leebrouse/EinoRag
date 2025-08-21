package main

import (
	"fmt"

	einorag "github.com/leebrouse/eino/Eino-rag"
	_ "github.com/leebrouse/eino/internal/config"
	"golang.org/x/net/context"
)

func main() {

	ctx := context.Background()
	client, err := einorag.NewRagClient()
	if err != nil {
		panic("error")
	}

	ids, err := client.Upload(ctx, "/root/Eino/data/document.pdf")
	if err != nil {
		panic("error")
	}
	fmt.Println("\nIds:", ids)

	result, err := client.Query(ctx, "what is Milvus")
	if err != nil {
		panic("error")
	}

	fmt.Println("\nResult:", result)

}
