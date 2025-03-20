package otp

import (
	"sync"

	"deedles.dev/xsync/internal/list"
)

// Sender is an interface wrapping the Send method. Its primary
// implementation is [Mailbox].
type Sender interface {
	Send(msg any)
}

// Mailbox is an OTP process mailbox. It works similarly to a channel
// but is not tied to a specific type and features a dynamic buffer.
// Sends to a mailbox are always asynchronous. A zero-value Mailbox is
// ready to use.
type Mailbox struct {
	once sync.Once

	m sync.Mutex
	c sync.Cond

	queue list.Double[any]
}

func (mb *Mailbox) init() {
	mb.once.Do(func() {
		mb.c.L = &mb.m
	})
}

// Send delivers a message to the Mailbox. If there are any blocked
// receives, they will check the new message to see if it is what
// they're waiting for after this function returns.
func (mb *Mailbox) Send(msg any) {
	mb.init()

	mb.m.Lock()
	defer mb.m.Unlock()

	mb.queue.Push(msg)
	mb.c.Broadcast()
}

func find[T any](mb *Mailbox, match func(T) bool) (v T, ok bool) {
	for n := range mb.queue.Nodes() {
		v, ok := n.Val.(T)
		if ok && match(v) {
			mb.queue.Remove(n)
			return v, true
		}
	}

	return v, false
}

// Recv checks mb to see if any messages that have been sent to it are
// matched by the given function. A message is considered to be a
// match if it both can be type asserted to T and the match function
// returns true. If there is such a message, it is removed from the
// Mailbox and returned. If there is no such message, Recv blocks
// until such a message arrives.
//
// If match is nil, all messages will be considered to be matching.
//
// For a non-blocking variant that returns immediately whether or not
// a matching message is present, see [TryRecv].
func Recv[T any](mb *Mailbox, match func(T) bool) T {
	mb.init()

	if match == nil {
		match = func(T) bool { return true }
	}

	mb.m.Lock()
	defer mb.m.Unlock()

	for {
		msg, ok := find(mb, match)
		if ok {
			return msg
		}

		mb.c.Wait()
	}
}

// TryRecv is like [Recv] but doesn't block, instead returning
// immediately whether or not a matching message is present in the
// Mailbox. If no message matches, it returns false as the second
// return.
func TryRecv[T any](mb *Mailbox, match func(T) bool) (msg T, ok bool) {
	mb.init()

	if match == nil {
		match = func(T) bool { return true }
	}

	mb.m.Lock()
	defer mb.m.Unlock()

	return find(mb, match)
}
