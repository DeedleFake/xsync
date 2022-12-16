package cq

type list[T any] struct {
	head, tail *node[T]
}

func (ls *list[T]) Enqueue(v T) {
	n := ls.tail.insert()
	n.Val = v
	ls.tail = n

	if ls.head == nil {
		ls.head = n
	}
}

func (ls *list[T]) Peek() (v T) {
	if ls.head == nil {
		return v
	}
	return ls.head.Val
}

func (ls *list[T]) Pop() bool {
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

type node[T any] struct {
	Val  T
	next *node[T]
}

func (n *node[T]) insert() *node[T] {
	if n == nil {
		return new(node[T])
	}

	n.next = &node[T]{next: n.next}
	return n.next
}
