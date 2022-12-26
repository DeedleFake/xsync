package xsync

import (
	"runtime"
	"sync"
)

// A Queue concurrently collects values and returns them in FIFO
// order. A zero value Queue is ready to use.
//
// A Queue is stopped when it is garbage collected. Therefore, a
// reference to the Queue must be kept alive during its use or its
// behavior will become undefined. Because of that, it is recommended
// to access the Queue's channels via the methods every time instead
// of storing a copy somewhere.
//
// A Queue is initialized by calling any of its methods, so a copy of
// a Queue made before those methods are called is a completely
// independent Queue, while a copy made afterwards is the same Queue.
//
// If a Queue's contents contain any references to the Queue itself,
// it can cause garbage collection to fail. For example, given a
// Queue[func()], if the Queue contains any closures which reference
// the actual instance of the Queue, the Queue's finalizer will not
// run until those elements have been removed from the Queue. Because
// of this, such a Queue will need to be manually stopped with a call
// to Queue.Stop.
type Queue[T any] struct {
	start sync.Once
	stop  func()
	block *byte

	add chan T
	get chan T
}

func (q *Queue[T]) init() {
	q.start.Do(func() {
		q.add = make(chan T)
		q.get = make(chan T)

		done := make(chan struct{})
		var stop sync.Once
		stopfunc := func() { stop.Do(func() { close(done) }) }
		q.stop = stopfunc

		go q.run(done)

		// SetFinalizer can only be called on the beginning of an
		// allocated block. If a Queue value, not a pointer, is present as
		// a field inside of another struct or in an array or something,
		// it won't be the beginning of the block. By tying the finalizer
		// to a field in the Queue that is allocated separately here, it
		// guarantees that it'll work. The field must be a type with a
		// size because the runtime doesn't actually allocate zero-sized
		// types, so this uses a byte instead of a struct{}.
		q.block = new(byte)
		runtime.SetFinalizer(q.block, func(*byte) { stopfunc() })
	})
}

func (q *Queue[T]) Stop() {
	q.init()
	q.stop()
}

// Add returns a channel that enqueues values sent to it. Closing this
// channel will cause the channel returned by Get to be closed once
// the Queue's contents are emptied, similar to how a regular channel
// works.
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

func (q Queue[T]) run(done <-chan struct{}) {
	add := q.add
	var get chan T

	defer func() {
		close(q.get)
		if add != nil {
			// Ensure that future attempts to send to the queue will fail.
			close(add)
		}
	}()

	var s list[T]
	for {
		select {
		case <-done:
			return

		case v, ok := <-add:
			if !ok {
				add = nil
				continue
			}

			s.Enqueue(v)
			get = q.get

		case get <- s.Peek():
			ok := s.Pop()
			if !ok {
				if add == nil {
					return
				}

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
