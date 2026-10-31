//go:build opennox_sdl_input_test || opennox_sdl_resize_test

package sdl

import (
	"runtime"
	"unsafe"

	native "github.com/veandco/go-sdl2/sdl"
)

// go-sdl2's PushEvent reads the complete SDL union, even for smaller event
// structs. Give it a zeroed, aligned allocation large enough for that union.
// SDL copies these pointer-free events before PushEvent returns.
func pushNativeFixtureEvent(event native.Event) int {
	var storage [(unsafe.Sizeof(native.CEvent{}) + 7) / 8]uint64
	var full native.Event
	switch event := event.(type) {
	case *native.KeyboardEvent:
		var _ [unsafe.Sizeof(native.CEvent{}) - unsafe.Sizeof(*event)]byte
		dst := (*native.KeyboardEvent)(unsafe.Pointer(&storage))
		*dst = *event
		full = dst
	case *native.WindowEvent:
		var _ [unsafe.Sizeof(native.CEvent{}) - unsafe.Sizeof(*event)]byte
		dst := (*native.WindowEvent)(unsafe.Pointer(&storage))
		*dst = *event
		full = dst
	case *native.MouseMotionEvent:
		var _ [unsafe.Sizeof(native.CEvent{}) - unsafe.Sizeof(*event)]byte
		dst := (*native.MouseMotionEvent)(unsafe.Pointer(&storage))
		*dst = *event
		full = dst
	case *native.QuitEvent:
		var _ [unsafe.Sizeof(native.CEvent{}) - unsafe.Sizeof(*event)]byte
		dst := (*native.QuitEvent)(unsafe.Pointer(&storage))
		*dst = *event
		full = dst
	default:
		panic("unsupported native SDL fixture event")
	}
	filtered, err := native.PushEvent(full)
	runtime.KeepAlive(&storage)
	if err != nil {
		return -1
	}
	if filtered {
		return 0
	}
	return 1
}
