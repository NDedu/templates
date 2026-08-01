package main

import (
	"fmt"
	"sync"
)

type WorkQueue[T any] struct {
	ch     chan T
	mu     sync.Mutex
	closed bool
}

func NewWorkQueue[T any](bufferSize int) *WorkQueue[T] {

	return &WorkQueue[T]{
		ch: make(chan T, bufferSize),
	}
}

func (q *WorkQueue[T]) Send(item T) error {

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is shut down")
	}

	select {
	case q.ch <- item:
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

func (q *WorkQueue[T]) Close() {

	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {

		return
	}
	q.closed = true
	close(q.ch)
}

func (q *WorkQueue[T]) Listen(handler func(T), wg *sync.WaitGroup) {

	for item := range q.ch {
		handler(item)
	}
	wg.Done()
}
