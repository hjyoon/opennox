//go:build !server

package opennox

import (
	"fmt"
	"image"
	"testing"

	noxcolor "github.com/opennox/libs/color"
)

// Full different-color overlap pixels use the actual factory/default
// draw/update/free and renderer/upload/seat, without callback replacements.
func TestParticleSweepOriginal4AFD40HeadlessOverlap(t *testing.T) {
	palette := [7]uint16{0x7c00, 0x03e0, 0x001f, 0x7fe0, 0x7c1f, 0x03ff, 0x7fff}
	for _, mode := range []string{"low-hole-reuse", "all-reuse", "expiry"} {
		t.Run(mode, func(t *testing.T) {
			for rotation := 0; rotation < len(palette); rotation++ {
				t.Run(fmt.Sprintf("rotation-%d", rotation), func(t *testing.T) {
					owner, pix, win, seat := newParticleRegistrationHeadless4B01B0(t)
					var slots [7]*particleFx
					var colors [7]uint16
					for i := range slots {
						colors[i] = palette[(i+rotation)%len(palette)]
						duration := 64
						if mode == "expiry" {
							duration = 2*i + 1
						}
						slots[i] = owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, duration, noxcolor.RGBA5551(colors[i]))
						slots[i].pointSize = 2
						assertParticleRegistration4B01B0(t, owner, slots[i])
					}
					want := make([]uint16, len(pix.Pix))
					for frame := 0; frame < 16; frame++ {
						if frame == 8 && mode != "expiry" {
							indices := []int{5, 0}
							if mode == "all-reuse" {
								indices = []int{6, 5, 4, 3, 2, 1, 0}
							}
							for _, i := range indices {
								slots[i].free()
							}
							lowSlots := []int{0, 5}
							if mode == "all-reuse" {
								lowSlots = []int{0, 1, 2, 3, 4, 5, 6}
							}
							for _, i := range lowSlots {
								colors[i] = palette[(6-i+rotation)%len(palette)]
								p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 64, noxcolor.RGBA5551(colors[i]))
								if p != slots[i] {
									t.Fatalf("visible replacement did not reuse original low slot %d", i)
								}
								p.pointSize = 2
							}
						}
						clear(pix.Pix)
						clear(want)
						highest := -1
						for i := range colors {
							if mode != "expiry" || frame <= 2*i+1 {
								highest = i
							}
						}
						if highest >= 0 {
							for y := 89; y < 91; y++ {
								for x := 64; x < 66; x++ {
									want[y*pix.Rect.Dx()+x] = colors[highest]
								}
							}
						}
						owner.Draw()
						presentParticleRegistration4B01B0(t, pix, win, seat, want, frame)
					}
					if seat.PresentCount() != 16 || mode == "expiry" && len(owner.byHandle) != 0 {
						t.Fatal("real overlap/default lifetime/headless upload incomplete")
					}
					t.Log("PARTICLE SWEEP OVERLAP: 16 actual different-color full raster/upload frames; factory/draw/update/free replacements=0")
				})
			}
		})
	}
}
