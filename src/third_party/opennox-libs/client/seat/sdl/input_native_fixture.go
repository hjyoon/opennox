//go:build opennox_sdl_input_test

package sdl

import native "github.com/veandco/go-sdl2/sdl"

func inputNativePushKey(id uint32, scancode int, pressed bool) int {
	event := &native.KeyboardEvent{
		Type: native.KEYUP, WindowID: id, State: native.RELEASED,
		Keysym: native.Keysym{Scancode: native.Scancode(scancode)},
	}
	if pressed {
		event.Type, event.State = native.KEYDOWN, native.PRESSED
	}
	return pushNativeFixtureEvent(event)
}
