// Package cq implements simple concurrent queues.
package cq

import "sync"

// A BulkQueue queues up values of type T and yields a collection of
// them all at once as a value of type Q.
type BulkQueue[T, Q any] struct {
	done  chan struct{}
	close sync.Once

	add  chan T
	get  chan Q
	wrap func([]T) Q
}

// Simple creates a simple Queue that yields its values as a slice of
// T.
func Simple[T any]() *BulkQueue[T, []T] {
	return New(func(q []T) []T { return q })
}

// New creates a Queue that converts queued up values into a Q by
// calling collect.
func New[T, Q any](collect func([]T) Q) *BulkQueue[T, Q] {
	q := BulkQueue[T, Q]{
		done: make(chan struct{}),
		add:  make(chan T),
		get:  make(chan Q),
		wrap: collect,
	}
	go q.run()

	return &q
}

// Stop stops the queue.
func (q *BulkQueue[T, Q]) Stop() {
	q.close.Do(func() {
		close(q.done)
	})
}

// Add returns a channel that enqueues values sent to it. This channel
// must not be closed.
func (q *BulkQueue[T, Q]) Add() chan<- T {
	return q.add
}

// Get returns a channel that yields all values enqueued since the
// last time it was recieved from as a Q. In other words, if ints are
// being queued
//
//	q.Add() <- 1
//	v = <-q.Get() // [1]
//	q.Add() <- 2
//	q.Add() <- 3
//	v = <-q.Get() // [2, 3]
func (q *BulkQueue[T, Q]) Get() <-chan Q {
	return q.get
}

func (q *BulkQueue[T, Q]) run() {
	defer func() {
		close(q.get)
	}()

	var s []T
	var get chan Q

	for {
		select {
		case <-q.done:
			return

		case v := <-q.add:
			s = append(s, v)
			get = q.get

		case get <- q.wrap(s):
			s = nil
			get = nil
		}
	}
}