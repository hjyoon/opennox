//go:build opennox_sdl_resize_test

package sdl

/*
#cgo pkg-config: sdl2
#include <SDL.h>

// SDL copies the complete union synchronously. Keep it on the C stack instead
// of go-sdl2 PushEvent's cast from a smaller Go event allocation.
static int nox_resize_push_window(Uint32 id, Uint8 subtype) {
	SDL_Event ev = {0};
	ev.type = SDL_WINDOWEVENT;
	ev.window.windowID = id;
	ev.window.event = subtype;
	ev.window.data1 = 7;
	ev.window.data2 = 9;
	return SDL_PushEvent(&ev);
}
static int nox_resize_push_mouse(Uint32 id) {
	SDL_Event ev = {0};
	ev.type = SDL_MOUSEMOTION;
	ev.motion.windowID = id;
	ev.motion.x = 400;
	ev.motion.y = 225;
	return SDL_PushEvent(&ev);
}
static int nox_resize_push_quit(void) {
	SDL_Event ev = {0};
	ev.type = SDL_QUIT;
	return SDL_PushEvent(&ev);
}
*/
import "C"

func resizeNativePushWindow(id uint32, subtype uint8) int {
	return int(C.nox_resize_push_window(C.Uint32(id), C.Uint8(subtype)))
}

func resizeNativePushMouse(id uint32) int {
	return int(C.nox_resize_push_mouse(C.Uint32(id)))
}

func resizeNativePushQuit() int { return int(C.nox_resize_push_quit()) }
