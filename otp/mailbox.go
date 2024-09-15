package otp

import (
	"iter"
	"sync"
)

// Mailbox is an OTP process mailbox. It works similarly to a channel
// but is not tied to a specific type and features a dynamic buffer.
// Sends to a mailbox are always asynchronous. A zero-value Mailbox is
// ready to use.
type Mailbox struct {
	once sync.Once

	m sync.Mutex
	c sync.Cond

	queue list[any]
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
	mb.m.Lock()
	defer mb.m.Unlock()

	mb.queue.push(msg)
	mb.c.Broadcast()
}

func find[T any](mb *Mailbox, match func(T) bool) (v T, ok bool) {
	for n := range mb.queue.nodes() {
		v, ok := n.val.(T)
		if ok && match(v) {
			mb.queue.remove(n)
			return v, true
		}
	}

	return v, false
}

// Recv checks mb to see if any messages that have been sent to it are
// matched by the given function. A message is considered to be a
// match if it both can be type asserted to T and the match function
// returns true. If there is such a message, it is removes from the
// Mailbox and returned. If there is no such message, Recv blocks
// until such a message arrives.
//
// For a non-blocking variant that returns immediately whether or not
// a matching message is present, see [TryRecv].
func Recv[T any](mb *Mailbox, match func(T) bool) T {
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

// TryRecv is like [Recv] but doesn't block, returning immediately
// whether not a matching message is present in the Mailbox. If no
// message matches, it returns false as the second return.
func TryRecv[T any](mb *Mailbox, match func(T) bool) (msg T, ok bool) {
	mb.m.Lock()
	defer mb.m.Unlock()

	return find(mb, match)
}

type list[T any] struct {
	head, tail *node[T]
}

func (ls *list[T]) push(v T) {
	n := node[T]{val: v, prev: ls.tail}
	ls.tail.next = &n
	ls.tail = &n
}

func (ls *list[T]) remove(n *node[T]) {
	if ls.head == ls.tail {
		ls.head = nil
		ls.tail = nil
		return
	}

	switch n {
	case ls.head:
		ls.head = n.next
		n.next.prev = nil
	case ls.tail:
		ls.tail = n.prev
		n.prev.next = nil
	default:
		n.next.prev = n.prev
		n.prev.next = n.next
	}
}

func (ls list[T]) nodes() iter.Seq[*node[T]] {
	return func(yield func(*node[T]) bool) {
		cur := ls.head
		for cur != nil {
			if !yield(cur) {
				return
			}
			cur = cur.next
		}
	}
}

type node[T any] struct {
	val        T
	prev, next *node[T]
}
