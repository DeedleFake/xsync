package list

import "iter"

// Double is a doubly-linked list. Unlike [Single], nodes of a Double
// can be removed from the middle of the list rather than only the
// ends.
type Double[T any] struct {
	head, tail *DoubleNode[T]
}

// Push adds a new node containing v to the tail of the list.
func (ls *Double[T]) Push(v T) {
	n := DoubleNode[T]{Val: v, prev: ls.tail}
	if ls.head == nil {
		ls.head = &n
		ls.tail = &n
		return
	}

	ls.tail.next = &n
	ls.tail = &n
}

// Remove removes the given node from the list.
func (ls *Double[T]) Remove(n *DoubleNode[T]) {
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

// Nodes returns an iterator over the nodes of the list. It is safe to
// remove the currently-yielded node from the list during iteration.
func (ls Double[T]) Nodes() iter.Seq[*DoubleNode[T]] {
	return func(yield func(*DoubleNode[T]) bool) {
		cur := ls.head
		for cur != nil {
			if !yield(cur) {
				return
			}
			cur = cur.next
		}
	}
}

// DoubleNode is a node of a [Double].
type DoubleNode[T any] struct {
	Val        T
	prev, next *DoubleNode[T]
}
