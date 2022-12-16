// Package cq implements simple concurrent queues.
package cq

import "sync"

// A Queue concurrently collects values and returns them in FIFO
// order. A zero value Queue is ready to use.
type Queue[T any] struct {
	start sync.Once

	done  chan struct{}
	close sync.Once

	add chan T
	get chan T
}

func (q *Queue[T]) init() {
	q.start.Do(func() {
		q.done = make(chan struct{})
		q.add = make(chan T)
		q.get = make(chan T)

		go q.run()
	})
}

// Stop stops the queue.
func (q *Queue[T]) Stop() {
	q.close.Do(func() {
		close(q.done)
	})
}

// Add returns a channel that enqueues values sent to it. This channel
// must not be closed.
func (q *Queue[T]) Add() chan<- T {
	q.init()
	return q.add
}

// Get returns a channel that yields values from the queue when they
// are available. The channel will be closed when the Queue is
// stopped.
func (q *Queue[T]) Get() <-chan T {
	q.init()
	return q.get
}

func (q *Queue[T]) run() {
	defer func() {
		close(q.get)
	}()

	var s list[T]
	var get chan T

	for {
		select {
		case <-q.done:
			return

		case v := <-q.add:
			s.Enqueue(v)
			get = q.get

		case get <- s.Peek():
			ok := s.Pop()
			if !ok {
				get = nil
			}
		}
	}
}
