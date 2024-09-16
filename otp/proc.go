package otp

import "context"

// Proc is an OTP process.
type Proc struct {
	cancel func()
	done   chan struct{}
	err    error
	mb     Mailbox
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
		defer func() {
			switch err := recover().(type) {
			case error:
				p.err = err
			case nil:
				return
			default:
				panic(err)
			}
		}()

		defer cancel()
		defer close(p.done)

		p.err = f(ctx)
	}()

	return &p
}

// Mailbox returns the process's Mailbox.
func (p *Proc) Mailbox() *Mailbox {
	return &p.mb
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

type selfKey struct{}

// Self returns the current process from the context, or nil if there
// is none.
func Self(ctx context.Context) *Proc {
	p, _ := ctx.Value(selfKey{}).(*Proc)
	return p
}
