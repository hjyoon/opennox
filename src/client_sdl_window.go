//go:build !server

package opennox

import (
	"image"

	"github.com/opennox/libs/client/seat"
)

// SDL mouse positions use logical window units; rendering uses drawable pixels.
// Ask the owning seat so go-sdl2 remains the only package linking SDL.
func sdlLogicalWindowSize(st seat.Seat, fallback image.Point) image.Point {
	if win, ok := st.(interface{ LogicalSize() image.Point }); ok {
		if sz := win.LogicalSize(); sz.X > 0 && sz.Y > 0 {
			return sz
		}
	}
	return fallback
}
