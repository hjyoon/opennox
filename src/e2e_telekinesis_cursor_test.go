package opennox

import (
	"image"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2ETelekinesisCursorIndependentInputAndPacket(t *testing.T) {
	input := e2eTelekinesisCursorInput{canvas: image.Pt(557, 429), previousWorld: image.Pt(3120, 2200)}
	sent := image.Pt(3159, 2291)
	for _, tc := range []struct {
		name                   string
		input                  e2eTelekinesisCursorInput
		canvas, sent, received image.Point
		hand                   types.Pointf
		want                   bool
	}{
		{"camera moved before packet", input, input.canvas, sent, sent, types.Ptf(3159, 2291), true},
		{"mouse event not consumed", input, image.Pt(556, 429), sent, sent, types.Ptf(3159, 2291), false},
		{"unchanged cursor is not movement", e2eTelekinesisCursorInput{canvas: input.canvas, previousWorld: sent}, input.canvas, sent, sent, types.Ptf(3159, 2291), false},
		{"server did not receive packet", input, input.canvas, sent, image.Pt(3158, 2291), types.Ptf(3159, 2291), false},
		{"hand missed packet", input, input.canvas, sent, sent, types.Ptf(3159, 2290), false},
		{"signed cursor coordinates", input, input.canvas, image.Pt(-1, -2), image.Pt(-1, -2), types.Ptf(-1, -2), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eTelekinesisCursorMatches(tc.input, tc.canvas, tc.sent, tc.received, tc.hand); got != tc.want {
				t.Fatalf("cursor match=%t want=%t", got, tc.want)
			}
		})
	}
}
