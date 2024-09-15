package otp_test

import (
	"testing"

	"deedles.dev/xsync/otp"
)

func TestMailbox(t *testing.T) {
	t.Run("sequential", func(t *testing.T) {
		var mb otp.Mailbox
		mb.Send(1)
		mb.Send(2)
		mb.Send(3)
		v := otp.Recv(&mb, func(v int) bool { return v%2 == 0 })
		if v != 2 {
			t.Fatal(v)
		}
		v = otp.Recv[int](&mb, nil)
		if v != 1 {
			t.Fatal(v)
		}
		v = otp.Recv[int](&mb, nil)
		if v != 3 {
			t.Fatal(v)
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		var mb otp.Mailbox
		c := make(chan int)
		go func() {
			c <- otp.Recv(&mb, func(v int) bool { return v%2 == 0 })
		}()

		go func() {
			mb.Send(1)
			mb.Send(2)
			mb.Send(3)
		}()
		v := <-c
		if v != 2 {
			t.Fatal(v)
		}

		match := func(v int) bool { return v%2 != 0 }
		v = otp.Recv(&mb, match)
		if v != 1 {
			t.Fatal(v)
		}
		v = otp.Recv(&mb, match)
		if v != 3 {
			t.Fatal(v)
		}
	})
}
