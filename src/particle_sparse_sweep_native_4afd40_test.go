package opennox

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func particleSparseGap4AFD40(count int) {
	for i := 0; i < count; i++ {
		_ = handles.New()
	}
}

func TestParticleSparseSweepOriginal4AFD40Default(t *testing.T) {
	for _, count := range []int{17, 64} {
		t.Run(fmt.Sprintf("records-%d", count), func(t *testing.T) {
			owner := particleSweepOwner4AFD40(t)
			particles := make([]*particleFx, count)
			for i := range particles {
				particles[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 64, color.White)
				particleSparseGap4AFD40(1024)
			}
			for frame := 0; frame < 16; frame++ {
				owner.Draw()
				for i, p := range particles {
					if p.cnt40 != frame+1 || p.ticksLeft != 63-frame || owner.byHandle[p.chnd] != p {
						t.Fatalf("original default update/lifetime, frame=%d slot=%d ticks=%d left=%d", frame, i, p.cnt40, p.ticksLeft)
					}
				}
			}
		})
	}
}

func TestParticleSparseSweepOriginal4AFD40Dynamic(t *testing.T) {
	cases := []struct {
		name   string
		first  []string
		second []int
	}{
		{"append-from-draw", particleSweepLiteralTrace4AFD40(0, 1, 2, 3, 4), []int{0, 1, 2, 3, 4}},
		{"append-from-update", particleSweepLiteralTrace4AFD40(0, 1, 2, 3, 4), []int{0, 1, 2, 3, 4}},
		{"reuse-before-current", particleSweepLiteralTrace4AFD40(0, 1, 2, 3), []int{4, 1, 2, 3}},
		{"activate-future-hole", particleSweepLiteralTrace4AFD40(0, 1, 4, 3), []int{0, 1, 4, 3}},
		{"replace-future", particleSweepLiteralTrace4AFD40(0, 1, 2, 4), []int{0, 1, 2, 4}},
		{"remove-future", particleSweepLiteralTrace4AFD40(0, 1, 2), []int{0, 1, 2}},
		{"replace-current", []string{"D0", "U0", "D1", "U4", "D2", "U2", "D3", "U3"}, []int{0, 4, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner := particleSweepOwner4AFD40(t)
			var trace []string
			var slots [4]*particleFx
			mutated := false
			change := func(phase string, label int) {
				when := "D"
				if tc.name == "append-from-update" {
					when = "U"
				}
				if phase != when || label != 1 || mutated {
					return
				}
				mutated = true
				switch tc.name {
				case "reuse-before-current":
					slots[0].free()
				case "replace-future", "remove-future":
					slots[3].free()
				case "replace-current":
					slots[1].free()
				}
				if tc.name != "remove-future" {
					particleSparseGap4AFD40(4096)
					particleSweepTrace4AFD40(owner, 4, &trace, nil)
				}
			}
			for i := range slots {
				slots[i] = particleSweepTrace4AFD40(owner, i, &trace, change)
				particleSparseGap4AFD40(4096)
			}
			if tc.name == "activate-future-hole" {
				slots[2].free()
			}
			owner.Draw()
			if !mutated || !slices.Equal(trace, tc.first) {
				t.Fatalf("sparse callback barrier: got=%v original=%v changed=%t", trace, tc.first, mutated)
			}
			trace = nil
			owner.Draw()
			if want := particleSweepLiteralTrace4AFD40(tc.second...); !slices.Equal(trace, want) {
				t.Fatalf("sparse next-frame cursor: got=%v original=%v", trace, want)
			}
		})
	}
}

func TestParticleSparseSweepOriginal4AFD40DenseCallbackWidensArena(t *testing.T) {
	owner := particleSweepOwner4AFD40(t)
	var trace []string
	created := false
	particleSweepTrace4AFD40(owner, 0, &trace, func(phase string, label int) {
		if phase == "U" && !created {
			created = true
			particleSparseGap4AFD40(16384)
			particleSweepTrace4AFD40(owner, 2, &trace, nil)
		}
	})
	particleSweepTrace4AFD40(owner, 1, &trace, nil)
	owner.Draw()
	if want := particleSweepLiteralTrace4AFD40(0, 1, 2); !created || !slices.Equal(trace, want) {
		t.Fatalf("dense-to-wide callback frontier: got=%v original=%v", trace, want)
	}
}
