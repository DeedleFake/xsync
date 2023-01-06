package xsync

import "reflect"

// SelectCase represents either a send or receive on a channel, or a
// default case with no channel associated.
type SelectCase interface {
	Dir() reflect.SelectDir

	rcase() reflect.SelectCase
	do(reflect.Value, bool)
}

// Select performs a select operation on the provided cases.
func Select(cases ...SelectCase) {
	rcases := make([]reflect.SelectCase, 0, len(cases))
	for _, c := range cases {
		rcases = append(rcases, c.rcase())
	}

	i, v, ok := reflect.Select(rcases)
	cases[i].do(v, ok)
}

type recvCase[T any] struct {
	c reflect.Value
	f func(T)
}

// Recv returns a SelectCase representing single-value receive from
// the channel c. If f is not nil, it will be called with the result
// of the receive if the receive is selected.
func Recv[T any](c <-chan T, f func(T)) SelectCase {
	return recvCase[T]{c: reflect.ValueOf(c), f: f}
}

func (c recvCase[T]) Dir() reflect.SelectDir {
	return reflect.SelectRecv
}

func (c recvCase[T]) rcase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir:  c.Dir(),
		Chan: c.c,
	}
}

func (c recvCase[T]) do(v reflect.Value, ok bool) {
	if c.f != nil {
		c.f(v.Interface().(T))
	}
}

type recvOKCase[T any] struct {
	c reflect.Value
	f func(T, bool)
}

// RecvOK returns a SelectCase representing a two-value receive from
// the channel c. If f is not nil, it will be called with the result
// of the receive if the receive is selected.
func RecvOK[T any](c <-chan T, f func(T, bool)) SelectCase {
	return recvOKCase[T]{c: reflect.ValueOf(c), f: f}
}

func (c recvOKCase[T]) Dir() reflect.SelectDir {
	return reflect.SelectRecv
}

func (c recvOKCase[T]) rcase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir:  c.Dir(),
		Chan: c.c,
	}
}

func (c recvOKCase[T]) do(v reflect.Value, ok bool) {
	if c.f != nil {
		c.f(v.Interface().(T), ok)
	}
}

type sendCase[T any] struct {
	c reflect.Value
	v reflect.Value
	f func()
}

// Send returns a SelectCase representing a send of v to the channel
// c. If f is not nil, it will be called if the send is selected.
func Send[T any](c chan<- T, v T, f func()) SelectCase {
	return sendCase[T]{c: reflect.ValueOf(c), v: reflect.ValueOf(v), f: f}
}

func (c sendCase[T]) Dir() reflect.SelectDir {
	return reflect.SelectSend
}

func (c sendCase[T]) rcase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir:  c.Dir(),
		Chan: c.c,
		Send: c.v,
	}
}

func (c sendCase[T]) do(v reflect.Value, ok bool) {
	if c.f != nil {
		c.f()
	}
}

type defaultCase struct {
	f func()
}

// Default returns a SelectCase that represents a default case. If f
// is not nil, it will be called if the case is selected.
func Default(f func()) SelectCase {
	return defaultCase{f: f}
}

func (c defaultCase) Dir() reflect.SelectDir {
	return reflect.SelectDefault
}

func (c defaultCase) rcase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir: c.Dir(),
	}
}

func (c defaultCase) do(reflect.Value, bool) {
	if c.f != nil {
		c.f()
	}
}
