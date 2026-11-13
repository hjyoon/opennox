package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func serverNativeNew[T comparable](high bool, value T) (*T, func()) {
	if high {
		return alloc.NewHigh(value)
	}
	return alloc.New(value)
}

func serverNativeMake[T comparable](high bool, src []T, count int) ([]T, func()) {
	if high {
		return alloc.MakeHigh(src, count)
	}
	return alloc.Make(src, count)
}

func serverNativeMalloc(high bool, size uintptr) (unsafe.Pointer, func()) {
	if high {
		return alloc.MallocHigh(size)
	}
	return alloc.Malloc(size)
}
