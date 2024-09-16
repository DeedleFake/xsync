package otp

import (
	"context"
	"sync"
)

// Proc is an OTP process.
type Proc struct {
	cancel func()
	done   chan struct{}
	err    error
	mb     Mailbox
	mon    sync.Map
}

// Go runs f as an OTP process. The passed context will be canceled
// when the process is requested to exit, as well as if the process
// exits on its own. If f panics with an error, it will be recovered
// from and treated as though that error had been returned. Non-error
// panic values will be repanicked.
//
// The current process can be retrieved from the provided context
// using the [Self] function.
func Go(f func(ctx context.Context) error) *Proc {
	p := Proc{done: make(chan struct{})}
	ctx := context.WithValue(context.Background(), selfKey{}, &p)
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	go func() {
		defer p.notify()
		defer p.catch()
		defer cancel()
		defer close(p.done)

		p.err = f(ctx)
	}()

	return &p
}

func (p *Proc) notify() {
	for s := range p.mon.Range {
		s.(Sender).Send(MonitoredProcessExited{Proc: p})
	}
}

func (p *Proc) catch() {
	switch err := recover().(type) {
	case error:
		p.err = err
	case nil:
		return
	default:
		panic(err)
	}
}

// Mailbox returns the process's Mailbox.
func (p *Proc) Mailbox() *Mailbox {
	return &p.mb
}

// Send sends a message to the process's mailbox.
func (p *Proc) Send(msg any) {
	p.mb.Send(msg)
}

// Done returns a channel that will be closed once the process has
// fully exited.
func (p *Proc) Done() <-chan struct{} {
	return p.done
}

// Wait will block until the process has exited and then return any
// error that it exited with.
func (p *Proc) Wait() error {
	<-p.done
	return p.err
}

// Stop signals to the process that it should exit by canceling its
// root context.
func (p *Proc) Stop() {
	p.cancel()
}

// Monitor registers s as monitoring p. When p exits, all monitoring
// processes will be sent a [MonitoredProcessExited] message.
//
// If p has already exited when Monitor is called, the message will be
// sent to s immediately.
func (p *Proc) Monitor(s Sender) {
	select {
	case <-p.done:
		s.Send(MonitoredProcessExited{Proc: p})
	default:
		p.mon.Store(s, struct{}{})
	}
}

// Unmonitor unregisters s from monitoring p. If s is not monitoring
// p, this function is a no-op.
func (p *Proc) Unmonitor(s Sender) {
	p.mon.Delete(s)
}

type selfKey struct{}

// Self returns the current process from the context, or nil if there
// is none.
func Self(ctx context.Context) *Proc {
	p, _ := ctx.Value(selfKey{}).(*Proc)
	return p
}

// MonitoredProcessExited is a message sent to processes that are
// monitoring another process when the monitored process exits. See
// [Process.Monitor].
type MonitoredProcessExited struct {
	Proc *Proc
}
