package main

import (
	"fmt"
	"sync"
	"time"
)

type Executor struct {
	numWorkers int       // worker 数量
	numTasks   int       // 任务总数
	taskch     chan int  // 任务通道
	resultch   chan int  // 结果通道
	wg         sync.WaitGroup
}

// NewExecutor 创建一个执行器
func NewExecutor(numWorkers, numTasks int) *Executor {
	return &Executor{
		numWorkers: numWorkers,
		numTasks:   numTasks,
		taskch:     make(chan int),
		resultch:   make(chan int),
	}
}

// Run 执行任务并计时
func (e *Executor) Run() {
	start := time.Now()

	// 生产任务
	go func() {
		for i := 0; i < e.numTasks; i++ {
			e.taskch <- i
		}
		close(e.taskch)
	}()

	// 启动 worker
	for i := 0; i < e.numWorkers; i++ {
		e.wg.Add(1)
		go e.worker()
	}

	// 收集结果
	done := make(chan struct{})
	go func() {
		for num := range e.resultch {
			fmt.Println(num)
		}
		close(done)
	}()

	// 等待 worker 完成
	e.wg.Wait()
	close(e.resultch)
	<-done

	fmt.Printf("Total elapsed time: %v\n", time.Since(start))
}

func (e *Executor) worker() {
	defer e.wg.Done()
	for num := range e.taskch {
		result := process(num)
		e.resultch <- result
	}
}

func process(num int) int {
	time.Sleep(2 * time.Second)
	return num * 2
}

func main() {
	exec := NewExecutor(100, 1000) // 100 个 worker，1000 个任务
	exec.Run()
}
