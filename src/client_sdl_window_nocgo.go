//go:build !server && !cgo

package opennox

import "image"

func sdlLogicalWindowSize(fallback image.Point) image.Point {
	return fallback
}
