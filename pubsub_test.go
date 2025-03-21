package xsync_test

import (
	"testing"

	"deedles.dev/xsync"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	var pub xsync.Pub[string]
	err := pub.Send(t.Context(), "no subs")
	require.Nil(t, err)

	sub1 := pub.Sub()
	go pub.Send(t.Context(), "one sub")
	require.Equal(t, <-sub1.Recv(), "one sub")

	sub2 := pub.Sub()
	go pub.Send(t.Context(), "two subs")
	for range 2 {
		select {
		case v := <-sub1.Recv():
			require.Equal(t, v, "two subs")
		case v := <-sub2.Recv():
			require.Equal(t, v, "two subs")
		}
	}

	sub1.Stop()
	go pub.Send(t.Context(), "one sub")
	require.Equal(t, <-sub2.Recv(), "one sub")
}
