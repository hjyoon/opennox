package opennox

import (
	"math"

	"github.com/opennox/libs/types"
)

// Keep the unowned Glyph nearer without requiring a large empty arena.
// Moving it opposite the dominant cardinal axis keeps it outside the
// selected Glyph's original +/-100 chain rectangle even in diagonal lanes.
// This only calculates fixture placements; it supplies no game outcome.
func e2eTriggerGlyphLayout(origin, direction types.Pointf) (selected, foreign types.Pointf) {
	foreign = origin
	if math.Abs(float64(direction.X)) >= math.Abs(float64(direction.Y)) {
		foreign.X -= float32(math.Copysign(40, float64(direction.X)))
	} else {
		foreign.Y -= float32(math.Copysign(40, float64(direction.Y)))
	}
	return origin.Add(direction.Mul(160)), foreign
}
