package workerpool

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"

	"golang.org/x/time/rate"
)

// --- Worker Pool Structure Definition ---

// WorkerPool encapsulates all components needed for concurrent task processing
type WorkerPool struct {
	workers    int                     // Number of concurrent workers
	maxRetries int                     // Maximum retries per task
	batchSize  int                     // Number of documents per batch
	tasks      chan []*schema.Document // Channel for tasks
	results    chan []*schema.Document // Channel for results
	wg         sync.WaitGroup          // WaitGroup for workers
	limiter    *rate.Limiter           // Rate limiter to avoid API exhaustion
	splitter   document.Transformer    // Transformer that splits and embeds documents
}

// NewWorkerPool creates and initializes a new WorkerPool
func NewWorkerPool(splitter document.Transformer, limiter *rate.Limiter) *WorkerPool {
	workers := viper.GetInt("workerPool.workers")
	batchSize := viper.GetInt("workerPool.batchSize")
	retry := viper.GetInt("workerPool.retry")

	return &WorkerPool{
		workers:    workers,
		batchSize:  batchSize,
		maxRetries: retry,
		tasks:      make(chan []*schema.Document),
		results:    make(chan []*schema.Document),
		limiter:    limiter,
		splitter:   splitter,
	}
}

// worker is the core function executed by each goroutine
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	fmt.Printf("Worker %d started\n", workerID)
	for batch := range wp.tasks {
		// Wait for the rate limiter before processing
		if err := wp.limiter.Wait(ctx); err != nil {
			fmt.Printf("Worker %d rate limiter wait failed: %v\n", workerID, err)
			return
		}

		fmt.Printf("Worker %d processing a batch (%d documents)...\n", workerID, len(batch))
		splitted, err := retryTransform(ctx, wp.splitter, batch, wp.maxRetries)
		if err != nil {
			fmt.Printf("Worker %d failed to process batch (max retries reached): %v\n", workerID, err)
			continue // Continue with the next task
		}
		wp.results <- splitted
	}
	fmt.Printf("Worker %d finished\n", workerID)
}

// Run starts all workers in the WorkerPool
func (wp *WorkerPool) Run(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}

	// Start a goroutine to close results channel after all workers complete
	go func() {
		wp.wg.Wait()
		close(wp.results)
	}()
}

// GenerateTasks splits documents into batches and sends them to the task channel
func (wp *WorkerPool) GenerateTasks(docs []*schema.Document) {
	go func() {
		for i := 0; i < len(docs); i += wp.batchSize {
			end := min(i+wp.batchSize, len(docs))
			wp.tasks <- docs[i:end]
		}
		close(wp.tasks) // Close the task channel after all tasks are sent
	}()
}

// AssembleChunks collects all processed document chunks from the results channel
func (wp *WorkerPool) AssembleChunks() []*schema.Document {
	var allChunks []*schema.Document
	for batchResult := range wp.results {
		allChunks = append(allChunks, batchResult...)
	}

	return allChunks
}

// retryTransform is a transform function with exponential backoff and intelligent retries
func retryTransform(ctx context.Context, splitter document.Transformer, batch []*schema.Document, maxRetries int) ([]*schema.Document, error) {
	var result []*schema.Document
	var err error
	backoff := 2 * time.Second // Initial backoff duration

	for i := range maxRetries {
		result, err = splitter.Transform(ctx, batch)
		if err == nil {
			return result, nil // Success, return immediately
		}

		// Intelligent error check: retry only on 429 errors
		if strings.Contains(err.Error(), "429") {
			fmt.Printf("Detected 429 error, retrying in %v (attempt %d)... Error: %v\n", backoff, i+1, err)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			continue
		} else {
			// Other errors (e.g., invalid API key) are non-retryable
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

