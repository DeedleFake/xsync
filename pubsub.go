package xsync

import (
	"context"
	"runtime"
	"sync"
	"weak"
)

type Pub[T any] struct {
	once sync.Once
	stop func()

	pub   chan T
	sub   chan *Sub[T]
	unsub chan uint64
}

func (p *Pub[T]) init() {
	p.once.Do(func() {
		done := make(chan struct{})
		p.stop = sync.OnceFunc(func() { close(done) })

		p.pub = make(chan T)
		p.sub = make(chan *Sub[T])
		p.unsub = make(chan uint64)

		runner := pubRunner[T]{
			done:  done,
			pub:   p.pub,
			sub:   p.sub,
			unsub: p.unsub,
		}
		go runner.run()

		runtime.AddCleanup(p, func(stop func()) { stop() }, p.stop)
	})
}

func (p *Pub[T]) Sub() <-chan *Sub[T] {
	p.init()
	return p.sub
}

type pubRunner[T any] struct {
	done  chan struct{}
	pub   chan T
	sub   chan *Sub[T]
	unsub chan uint64

	subs map[uint64]weak.Pointer[Sub[T]]
}

func (p *pubRunner[T]) run() {
	p.subs = make(map[uint64]weak.Pointer[Sub[T]])
	next := p.next(0)

	for {
		select {
		case <-p.done:
			p.unsubAll()
			return

		case p.sub <- next:
			p.subs[next.id] = weak.Make(next)
			next = p.next(next.id + 1)

		case id := <-p.unsub:
			delete(p.subs, id)

		case v := <-p.pub:
			for id, w := range p.subs {
				done := p.send(id, w, v)
				if done {
					p.unsubAll()
					return
				}
			}
		}
	}
}

func (p *pubRunner[T]) unsubAll() {
	for _, w := range p.subs {
		sub := w.Value()
		if sub == nil {
			continue
		}
		close(sub.get)
	}
}

func (p *pubRunner[T]) send(id uint64, w weak.Pointer[Sub[T]], v T) bool {
	sub := w.Value()
	if sub == nil {
		delete(p.subs, id)
		return true
	}

	for {
		select {
		case <-p.done:
			return false

		case unsubID := <-p.unsub:
			delete(p.subs, unsubID)
			if id == unsubID {
				return true
			}

		case sub.get <- v:
		}
	}
}

func (p *pubRunner[T]) next(id uint64) *Sub[T] {
	sub := Sub[T]{
		id:  id,
		get: make(chan T),
	}

	runtime.AddCleanup(&sub, func(p *pubRunner[T]) {
		select {
		case <-p.done:
		case p.unsub <- id:
		}
	}, p)

	return &sub
}

type Sub[T any] struct {
	_ noCopy

	id  uint64
	get chan T
}

func (s *Sub[T]) Get() <-chan T {
	return s.get
}

func (s *Sub[T]) Stop(ctx context.Context) {
	// TODO
}
