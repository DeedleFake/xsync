package list

import "iter"

type Single[T any] struct {
	head, tail *SingleNode[T]
}

func (ls *Single[T]) Enqueue(v T) {
	n := ls.tail.insert()
	n.Val = v
	ls.tail = n

	if ls.head == nil {
		ls.head = n
	}
}

func (ls *Single[T]) Peek() (v T) {
	if ls.head == nil {
		return v
	}
	return ls.head.Val
}

func (ls *Single[T]) Pop() bool {
	if ls.head == nil {
		return false
	}

	n := ls.head
	ls.head = n.next
	if ls.head == nil {
		ls.tail = nil
	}

	return ls.head != nil
}

func (ls *Single[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		cur := ls.head
		for cur != nil {
			if !yield(cur.Val) {
				return
			}
			cur = cur.next
		}
	}
}

type SingleNode[T any] struct {
	Val  T
	next *SingleNode[T]
}

func (n *SingleNode[T]) insert() *SingleNode[T] {
	if n == nil {
		return new(SingleNode[T])
	}

	n.next = &SingleNode[T]{next: n.next}
	return n.next
}
