package list

import "iter"

type Double[T any] struct {
	head, tail *DoubleNode[T]
}

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

type DoubleNode[T any] struct {
	Val        T
	prev, next *DoubleNode[T]
}
