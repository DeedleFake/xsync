package xsync

import "sync"

// A Stopper provides a simple way to handle a done channel for
// internal coordination. For coordination across API boundaries, it
// is generally better to use context.Context.
//
// The zero value of a Stopper is ready to use.
type Stopper struct {
	done  chan struct{}
	start sync.Once
	close sync.Once
}

func (s *Stopper) init() {
	s.start.Do(func() {
		s.done = make(chan struct{})
	})
}

// Stop closes the Stoppers Done channel. It is safe to call more than
// once.
func (s *Stopper) Stop() {
	s.init()
	s.close.Do(func() {
		close(s.done)
	})
}

// Done returns a channel that is closed when the Stop method is
// called.
func (s *Stopper) Done() <-chan struct{} {
	s.init()
	return s.done
}
