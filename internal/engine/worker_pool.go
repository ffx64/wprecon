package engine

import (
	"sync"
)

type Task interface{}

type WorkerPool struct {
	workers    int
	taskChan   chan Task
	resultChan chan interface{}
	wg         sync.WaitGroup
	done       chan struct{}
}

type WorkerFunc func(task Task) interface{}

func NewWorkerPool(workerCount int, bufferSize int) *WorkerPool {
	return &WorkerPool{
		workers:    workerCount,
		taskChan:   make(chan Task, bufferSize),
		resultChan: make(chan interface{}, bufferSize),
		done:       make(chan struct{}),
	}
}

func (wp *WorkerPool) Start(fn WorkerFunc) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(fn)
	}
}

func (wp *WorkerPool) worker(fn WorkerFunc) {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.taskChan:
			if !ok {
				return
			}
			result := fn(task)
			wp.resultChan <- result
		case <-wp.done:
			return
		}
	}
}

func (wp *WorkerPool) Submit(task Task) {
	wp.taskChan <- task
}

func (wp *WorkerPool) Close() {
	close(wp.taskChan)
	wp.wg.Wait()
	close(wp.resultChan)
}

func (wp *WorkerPool) CloseInput() {
	close(wp.taskChan)

	go func() {
		wp.wg.Wait()
		close(wp.resultChan)
	}()
}

func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
}

func (wp *WorkerPool) Results() <-chan interface{} {
	return wp.resultChan
}
