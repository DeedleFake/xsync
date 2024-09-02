package xsync_test

import (
	"testing"

	"deedles.dev/xsync"
)

func TestFuture(t *testing.T) {
	f, complete := xsync.NewFuture[int]()
	go func() {
		complete(3)
	}()

	val := f.Get()
	if val2 := f.Get(); val != val2 {
		t.Fatalf("%v != %v", val, val2)
	}
	if val != 3 {
		t.Fatal(val)
	}
}

func BenchmarkFuture(b *testing.B) {
	for range b.N {
		f, complete := xsync.NewFuture[int]()
		complete(3)
		f.Get()
	}
}
