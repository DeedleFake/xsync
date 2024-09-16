package otp

import (
	"context"
	"fmt"
	"log/slog"
)

func StartGenServer(handler any) *Proc {
	return Go(func(ctx context.Context) error {
		hcall := CallHandler(defaultHandler{})
		if h, ok := handler.(CallHandler); ok {
			hcall = h
		}

		hcast := CastHandler(defaultHandler{})
		if h, ok := handler.(CastHandler); ok {
			hcast = h
		}

		hinfo := InfoHandler(defaultHandler{})
		if h, ok := handler.(InfoHandler); ok {
			hinfo = h
		}

		mb := Self(ctx).Mailbox()
		for {
			switch v := Recv[any](mb, nil).(type) {
			case call:
				hcall.HandleCall(func(result any) {
					v.from.Send(callResult{result})
				}, v.msg)

			case cast:
				hcast.HandleCast(v.msg)

			default:
				hinfo.HandleInfo(v)
			}
		}
	})
}

type call struct {
	from Sender
	msg  any
}

type callResult struct {
	result any
}

func Call(s Sender, msg any) any {
	var mb Mailbox
	return CallFrom(&mb, s, msg)
}

func CallFrom(self *Mailbox, to Sender, msg any) any {
	to.Send(call{from: self, msg: msg})
	return Recv[callResult](self, nil).result
}

type cast struct {
	msg any
}

func Cast(s Sender, msg any) {
	s.Send(cast{msg: msg})
}

type CallHandler interface {
	HandleCall(reply func(result any), msg any)
}

type CastHandler interface {
	HandleCast(msg any)
}

type InfoHandler interface {
	HandleInfo(msg any)
}

type defaultHandler struct{}

func (defaultHandler) HandleCall(reply func(result any), msg any) {
	panic(fmt.Errorf("call handling not implemented but got call with message %#v", msg))
}

func (defaultHandler) HandleCast(msg any) {
	panic(fmt.Errorf("call handling not implemented but got call with message %#v", msg))
}

func (defaultHandler) HandleInfo(msg any) {
	slog.Warn("info handling not implemented but got message", "message", msg)
}
