package opennox

import (
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2EMeteorShowerImpactFireMinimum(t *testing.T) {
	for _, tc := range []struct {
		name       string
		raw        int32
		distance   float32
		unoccluded bool
		want       int32
	}{
		{"center", 85, 0, true, 85},
		{"inner-edge", 85, 30, true, 85},
		{"half-falloff", 85, 55, true, 42},
		{"fractional-positive-edge", 85, 79.9, true, 1},
		{"exact-outer-edge", 85, 80, true, 1},
		{"outside", 85, 81, true, 0},
		{"occluded-center", 85, 0, false, 0},
		{"occluded-edge", 85, 79.9, false, 0},
		{"admitted-zero", 0, 0, true, 1},
		{"signed-truncation", -85, 55, true, -42},
		{"signed-zero-edge", -85, 79.9, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := e2eMeteorShowerImpactDamage(tc.raw, types.Ptf(100, 200), types.Ptf(100+tc.distance, 200), tc.unoccluded)
			if got != tc.want {
				t.Fatalf("admitted radial/fire damage=%d want=%d", got, tc.want)
			}
		})
	}
}
