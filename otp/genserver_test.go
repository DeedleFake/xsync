package otp_test

import (
	"testing"

	"deedles.dev/xsync/otp"
)

func TestGenServer(t *testing.T) {
	p := otp.StartGenServer(nil)
	r := otp.Call(p, 3)
	if r != 6 {
		t.Fatal(r)
	}
}
