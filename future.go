package syncutil

import "sync"

// A Future holds a value that might not be available yet.
type Future[T any] struct {
	done chan struct{}
	val  T
}

// NewFuture returns a new future and a function that completes that
// future with the given value. The returned complete function becomes
// a no-op after the first usage.
func NewFuture[T any]() (f *Future[T], complete func(T)) {
	var once sync.Once
	f = &Future[T]{done: make(chan struct{})}
	return f, func(val T) {
		once.Do(func() {
			f.val = val
			close(f.done)
		})
	}
}

// Done returns a channel that is closed when the future completes.
func (f *Future[T]) Done() <-chan struct{} {
	return f.done
}

// Get blocks, if necessary, until the future is completed and then
// returns its value.
func (f *Future[T]) Get() T {
	<-f.done
	return f.val
}
