// Package xsync provides extra synchronization primitives to
// supplement the standard library.
package xsync

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
