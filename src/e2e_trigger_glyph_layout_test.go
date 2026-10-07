package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2ETriggerGlyphLayout(t *testing.T) {
	for _, tc := range []struct {
		name   string
		offset types.Pointf
	}{
		{"east", types.Ptf(160, 0)}, {"west", types.Ptf(-160, 0)},
		{"south", types.Ptf(0, 160)}, {"north", types.Ptf(0, -160)},
		{"southeast", types.Ptf(114, 114)}, {"southwest", types.Ptf(-114, 114)},
		{"northeast", types.Ptf(114, -114)}, {"northwest", types.Ptf(-114, -114)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			direction := tc.offset.Normalize()
			for _, origin := range []types.Pointf{
				{}, types.Ptf(3116, 2166), types.Ptf(-3116, -2166),
			} {
				selected, foreign := e2eTriggerGlyphLayout(origin, direction)
				if selected != origin.Add(direction.Mul(160)) {
					t.Fatalf("selected %v left the stock 160-unit lane", selected)
				}
				near := foreign.Sub(origin)
				if math.Abs(float64(near.X))+math.Abs(float64(near.Y)) != 40 ||
					near.X != 0 && near.Y != 0 || near.X*direction.X+near.Y*direction.Y >= 0 {
					t.Fatalf("foreign %v is not on the opposite 40-unit cardinal axis", foreign)
				}
				far := selected.Sub(origin)
				if near.X*near.X+near.Y*near.Y >= far.X*far.X+far.Y*far.Y {
					t.Fatal("unowned Glyph is not nearer to the caster")
				}
				delta := selected.Sub(foreign)
				if math.Abs(float64(delta.X)) <= 100 && math.Abs(float64(delta.Y)) <= 100 {
					t.Fatalf("Glyph placements %v and %v enter the original chain rectangle", selected, foreign)
				}
			}
		})
	}
}
