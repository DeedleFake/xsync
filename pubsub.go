package xsync

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"weak"
)

// Pub is a publisher in a PubSub system. It tracks subscriptions and
// can broadcast values of type T to them.
type Pub[T any] struct {
	once   sync.Once
	subs   sync.Map
	nextID uint64
}

func (p *Pub[T]) init() {
	p.once.Do(func() {
		runtime.AddCleanup(p, func(m *sync.Map) {
			for _, w := range m.Range {
				sub := w.(weak.Pointer[Sub[T]]).Value()
				if sub != nil {
					close(sub.recv)
				}
			}
		}, &p.subs)
	})
}

// Sub returns a new subscription to p. See [Sub] for more
// information.
func (p *Pub[T]) Sub() *Sub[T] {
	p.init()

	id := atomic.AddUint64(&p.nextID, 1)
	recv := make(chan T)
	sub := Sub[T]{
		recv: recv,
		stop: sync.OnceFunc(func() {
			p.subs.Delete(id)
		}),
	}
	runtime.AddCleanup(&sub, func(stop func()) { stop() }, sub.stop)

	p.subs.Store(id, weak.Make(&sub))
	return &sub
}

// Send publishes v to all of p's subscribers. If the context is
// canceled before v is sent, the context's cause is returned.
func (p *Pub[T]) Send(ctx context.Context, v T) error {
	p.init()

	for k, w := range p.subs.Range {
		sub := w.(weak.Pointer[Sub[T]]).Value()
		if sub == nil {
			p.subs.Delete(k)
			continue
		}

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case sub.recv <- v:
			continue
		}
	}

	return nil
}

// Sub is a subscription to a [Pub]. A Sub must not be copied after
// first use.
type Sub[T any] struct {
	_ noCopy

	stop func()
	recv chan T
}

// Recv returns a channel that yields values published by the
// corresponding [Pub].
//
// Note that the returned channel is not closed when the Sub is
// unsubscribed.
func (s *Sub[T]) Recv() <-chan T {
	return s.recv
}

// Stop unsubscribes from the publisher.
func (s *Sub[T]) Stop() {
	s.stop()
}
