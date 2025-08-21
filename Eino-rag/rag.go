package einorag

import "context"

type RAG interface {
	// Research relative answers from vector database
	// Step: 1. reason the meaning from the input promopt (llm such gemini)
	//  	 2. embedding the prompt and research the document from the vector database (retriever)
	// 		 3. get the results (gemini client api)
	Query(ctx context.Context, prompt string) (string, error)
	// Upload file such as "pdf,markdown,txt....." and embed them to vector [][]float64
	// Step: 1. upload file (loader)
	// 		 2. extract and chunk it (transformer)
	//  	 3. embedding the file and insert to the vector database (indexer)
	Upload(ctx context.Context, fileUrl string) ([]string, error)
}
