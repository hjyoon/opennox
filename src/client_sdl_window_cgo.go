//go:build !server && cgo

package opennox

/*
#cgo pkg-config: sdl2
#include <SDL.h>

static int nox_sdl_logical_window_size(int *w, int *h) {
	SDL_Window *win = SDL_GL_GetCurrentWindow();
	if (win == NULL) {
		return 0;
	}
	SDL_GetWindowSize(win, w, h);
	return *w > 0 && *h > 0;
}
*/
import "C"

import "image"

// sdlLogicalWindowSize returns SDL's logical client-area size. Mouse events
// use this coordinate space, while GLGetDrawableSize returns physical pixels
// on high-DPI displays.
func sdlLogicalWindowSize(fallback image.Point) image.Point {
	var w, h C.int
	if C.nox_sdl_logical_window_size(&w, &h) == 0 {
		return fallback
	}
	return image.Pt(int(w), int(h))
}
