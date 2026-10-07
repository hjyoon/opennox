package opennox

import (
	"image"

	"github.com/opennox/libs/client/seat"
)

// Five original sixty-second Lock lifetimes outlast the online game's
// 2700-input-tick idle-observer threshold. Keep the test active through the
// ordinary canvas-input queue, without buttons, movement, casts, clock resets
// or any write to player/door state. The unsigned cadence also spans wrap.
type e2eLockIdleInput struct {
	lastFrame uint32
}

func (input *e2eLockIdleInput) observe(frame uint32, cursor image.Point, queue func(...seat.InputEvent)) bool {
	if frame-input.lastFrame < 300 {
		return false
	}
	queue(&seat.MouseMoveEvent{Pos: cursor})
	input.lastFrame = frame
	return true
}
