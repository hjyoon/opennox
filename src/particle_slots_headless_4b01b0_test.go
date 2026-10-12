//go:build !server

package opennox

import (
	"image"
	"image/color"
	"testing"
)

// Real, nonoverlapping particle pixels certify actual factory/free/reuse and
// default draw/update ownership, NOT the still-random map sweep order.
func TestParticleSlotsOriginal4B01B0HeadlessReuse(t *testing.T) {
	tests := []struct {
		name       string
		free, want []int
	}{
		{"low-first", []int{0, 1}, []int{0, 1}},
		{"high-first", []int{1, 0}, []int{0, 1}},
		{"all-mixed", []int{2, 0, 1}, []int{0, 1, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner, pix, win, sc := newParticleRegistrationHeadless4B01B0(t)
			var slots [3]*particleFx
			for i := range slots {
				slots[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(71+i*20, 101), 1, 4, color.White)
				assertParticleRegistration4B01B0(t, owner, slots[i])
				slots[i].pointSize = 2
			}
			want := make([]uint16, len(pix.Pix))
			for i := range slots {
				particleRegistrationWhiteSquare4B01B0(want, pix.Rect, image.Pt(64+i*20, 89))
			}
			owner.Draw()
			presentParticleRegistration4B01B0(t, pix, win, sc, want, 0)
			for _, index := range tc.free {
				slots[index].free()
			}
			for i, index := range tc.want {
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71+index*20, 101), 1, 3, color.White)
				assertParticleRegistration4B01B0(t, owner, p)
				if p != slots[index] {
					t.Fatalf("actual visible factory allocation %d did not choose original low slot %d", i, index)
				}
				p.pointSize = 2
			}
			// Retained particles and newly reused ones now all have 3 ticks.
			for tick := 1; tick <= 4; tick++ {
				clear(pix.Pix)
				owner.Draw()
				presentParticleRegistration4B01B0(t, pix, win, sc, want, tick)
				if tick < 4 && len(owner.byHandle) != 3 {
					t.Fatal("actual reuse/default update lost a live slot")
				}
			}
			clear(pix.Pix)
			clear(want)
			owner.Draw()
			presentParticleRegistration4B01B0(t, pix, win, sc, want, 5)
			if sc.PresentCount() != 6 || len(owner.byHandle) != 0 {
				t.Fatal("actual headless reuse/expiry/free lifecycle incomplete")
			}
			t.Log("PARTICLE SLOT HEADLESS: 6 real nonoverlapping raster/upload frames, factory/free/low-slot reuse/default lifetime verified; allocator/result/draw hooks=0; overlap sweep order NOT certified")
		})
	}
}
