package opennox

import (
	"image"

	"github.com/opennox/libs/types"
)

type e2eTelekinesisCursorInput struct {
	canvas, previousWorld image.Point
}

// The ordinary camera can scroll between queuing a canvas event and sending
// MSG_MOUSE. Check the requested input and the client's actual world packet,
// not a pre-input viewport snapshot. Neither value is supplied to the server.
func e2eTelekinesisCursorMatches(input e2eTelekinesisCursorInput, canvas, sent, received image.Point, hand types.Pointf) bool {
	return canvas == input.canvas && sent != input.previousWorld && received == sent && hand == types.Ptf(float32(sent.X), float32(sent.Y))
}
