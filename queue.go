package xsync

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

type list[T any] struct {
	head, tail *node[T]
}

func (ls *list[T]) Enqueue(v T) {
	n := ls.tail.insert()
	n.Val = v
	ls.tail = n

	if ls.head == nil {
		ls.head = n
	}
}

func (ls *list[T]) Peek() (v T) {
	if ls.head == nil {
		return v
	}
	return ls.head.Val
}

func (ls *list[T]) Pop() bool {
	if ls.head == nil {
		return false
	}

	n := ls.head
	ls.head = n.next
	if ls.head == nil {
		ls.tail = nil
	}

	return ls.head != nil
}

type node[T any] struct {
	Val  T
	next *node[T]
}

func (n *node[T]) insert() *node[T] {
	if n == nil {
		return new(node[T])
	}

	n.next = &node[T]{next: n.next}
	return n.next
}
