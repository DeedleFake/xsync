package list

import "iter"

// Single is a singly-linked list that also contains a reference to
// the last node for quick inserts and removals at the head and tail.
type Single[T any] struct {
	head, tail *SingleNode[T]
}

// Enqueue adds v as a new node at the tail of the list.
func (ls *Single[T]) Enqueue(v T) {
	n := ls.tail.insert()
	n.Val = v
	ls.tail = n

	if ls.head == nil {
		ls.head = n
	}
}

// Peek returns the value of the head node or the zero value if the
// list is empty.
func (ls *Single[T]) Peek() (v T) {
	if ls.head == nil {
		return v
	}
	return ls.head.Val
}

// Pop removes the current head node from the list. It returns false
// if the list was already empty.
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

// All returns an iterator over the elements of the list.
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

// SingleNode is a node of a [Single].
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
