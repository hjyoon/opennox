//go:build !windows

package memguard

func newPageHigh(size int) ([]byte, func()) { return newPage(size) }
