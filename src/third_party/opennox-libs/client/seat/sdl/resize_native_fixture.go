//go:build opennox_sdl_resize_test

package sdl

import native "github.com/veandco/go-sdl2/sdl"

func resizeNativePushWindow(id uint32, subtype uint8) int {
	return pushNativeFixtureEvent(&native.WindowEvent{
		Type: native.WINDOWEVENT, WindowID: id,
		Event: native.WindowEventID(subtype), Data1: 7, Data2: 9,
	})
}

func resizeNativePushMouse(id uint32) int {
	return pushNativeFixtureEvent(&native.MouseMotionEvent{
		Type: native.MOUSEMOTION, WindowID: id, X: 400, Y: 225,
	})
}

func resizeNativePushQuit() int { return pushNativeFixtureEvent(&native.QuitEvent{Type: native.QUIT}) }
