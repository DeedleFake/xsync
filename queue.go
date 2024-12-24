package xsync

import (
	"iter"
	"runtime"
	"sync"

	"deedles.dev/xsync/internal/list"
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
	all chan iter.Seq[T]
}

func (q *Queue[T]) init() {
	q.start.Do(func() {
		q.add = make(chan T)
		q.get = make(chan T)
		q.all = make(chan iter.Seq[T])

		done := make(chan struct{})
		var stop sync.Once
		stopfunc := func() { stop.Do(func() { close(done) }) }
		q.stop = stopfunc

		runner := queueRunner[T]{
			add: q.add,
			get: q.get,
			all: q.all,
		}
		go runner.run(done)

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

// Push returns a channel that enqueues values sent to it. Closing
// this channel will cause the channel returned by Pop to be closed
// once the Queue's contents are emptied, similar to how a regular
// channel works.
func (q *Queue[T]) Push() chan<- T {
	q.init()
	return q.add
}

// Pop returns a channel that yields values from the queue when they
// are available. The channel will be closed when the Queue is
// stopped.
func (q *Queue[T]) Pop() <-chan T {
	q.init()
	return q.get
}

// All returns an iterator over all values currently in the queue.
// Receiving from the channel will empty the queue. The channel will
// block until there is at least one value in the queue.
//
// Like the channel returned by [Pop], it will be closed when the
// Queue is stopped.
func (q *Queue[T]) All() <-chan iter.Seq[T] {
	q.init()
	return q.all
}

type queueRunner[T any] struct {
	add chan T
	get chan T
	all chan iter.Seq[T]
}

func (q *queueRunner[T]) run(done <-chan struct{}) {
	add := q.add
	var get chan T
	var all chan iter.Seq[T]

	defer func() {
		close(q.get)
		close(q.all)
		if add != nil {
			// Ensure that future attempts to send to the queue will fail.
			close(add)
		}
	}()

	var s list.Single[T]
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
			all = q.all

		case get <- s.Peek():
			ok := s.Pop()
			if !ok {
				if add == nil {
					return
				}

				get = nil
			}

		case all <- s.All():
			s = list.Single[T]{}
			all = nil
			get = nil
		}
	}
}
