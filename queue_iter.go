//go:build go1.23

package xsync

import (
	"context"
	"iter"
)

// Values returns an iterator that yields values from the queue until
// either the queue is closed or the context is canceled.
func (q *Queue[T]) Values(ctx context.Context) iter.Seq[T] {
	q.init()
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-q.get:
				if !ok || !yield(v) {
					return
				}
			}
		}
	}
}
