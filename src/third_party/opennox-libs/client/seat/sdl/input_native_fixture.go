//go:build opennox_sdl_input_test

package sdl

/*
#cgo pkg-config: sdl2
#include <SDL.h>

// SDL synchronously copies its complete event union. Construct the union on
// the C stack instead of casting a smaller Go event allocation in PushEvent.
static int nox_input_push_key(Uint32 id, int scancode, int pressed) {
	SDL_Event ev = {0};
	ev.type = pressed ? SDL_KEYDOWN : SDL_KEYUP;
	ev.key.windowID = id;
	ev.key.state = pressed ? SDL_PRESSED : SDL_RELEASED;
	ev.key.keysym.scancode = (SDL_Scancode)scancode;
	return SDL_PushEvent(&ev);
}
*/
import "C"

func inputNativePushKey(id uint32, scancode int, pressed bool) int {
	return int(C.nox_input_push_key(C.Uint32(id), C.int(scancode), C.int(inputNativeBool(pressed))))
}

func inputNativeBool(v bool) int {
	if v {
		return 1
	}
	return 0
}
