package otp_test

import (
	"context"
	"testing"

	"deedles.dev/xsync/otp"
)

func TestProcessMonitor(t *testing.T) {
	var mb otp.Mailbox
	p := otp.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})
	p.Monitor(&mb)

	v, ok := otp.TryRecv[any](&mb, nil)
	if ok {
		t.Fatal(v)
	}

	p.Stop()
	v = otp.Recv[any](&mb, nil)
	if v != (otp.MonitoredProcessExited{Proc: p}) {
		t.Fatal(v)
	}
	v, ok = otp.TryRecv[any](&mb, nil)
	if ok {
		t.Fatal(v)
	}

	p.Monitor(&mb)
	v = otp.Recv[any](&mb, nil)
	if v != (otp.MonitoredProcessExited{Proc: p}) {
		t.Fatal(v)
	}
}
