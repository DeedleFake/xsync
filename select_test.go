package xsync_test

import (
	"testing"

	"deedles.dev/xsync"
)

func TestSelect(t *testing.T) {
	c1 := make(chan int)
	c2 := make(chan string)

	go func() {
		c1 <- 3
	}()

	var got any
	xsync.Select(
		xsync.Recv(c1, func(v int) {
			got = v
		}),
		xsync.Recv(c2, func(v string) {
			got = v
		}),
	)
	if got != 3 {
		t.Fatalf("expected 3 but got %v", got)
	}
}
