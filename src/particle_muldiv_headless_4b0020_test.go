//go:build !server

package opennox

import (
	"fmt"
	"image"
	"image/color"
	"testing"
)

// Actual factory -> default draw -> default update with all three damping
// axes and ground bounce -> free -> headless upload. Existing R146 renderer
// fixtures are reused unchanged; no production callbacks/results are hooked.
func TestParticleMulDivOriginal4B0020HeadlessDampingBounce(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		for _, bounce := range []bool{false, true} {
			t.Run(fmt.Sprintf("mask-%d/bounce-%t", mask, bounce), func(t *testing.T) {
				defer func() {
					if value := recover(); value != nil {
						t.Fatalf("real headless damping/bounce panicked: %v", value)
					}
				}()
				owner, pix, win, sc := newParticleRegistrationHeadless4B01B0(t)
				r := particleMulDivInput4B0020(mask, bounce)
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(135, 131), 1, 4, color.White)
				assertParticleRegistration4B01B0(t, owner, p)
				setParticleMulDivInput4B0020(p, r)
				p.pointSize = 2
				want := make([]uint16, len(pix.Pix))
				for tick := 0; tick < 8; tick++ {
					clear(pix.Pix)
					clear(want)
					if !r.dead {
						particleRegistrationWhiteSquare4B01B0(want, pix.Rect, image.Pt(int(r.x>>16)-7, int((r.y>>16)-(r.z>>16))-11))
					}
					r.step()
					owner.Draw()
					assertParticleMulDivState4B0020(t, owner, p, r, tick)
					presentParticleRegistration4B01B0(t, pix, win, sc, want, tick)
				}
				if sc.PresentCount() != 8 || !r.dead || len(owner.byHandle) != 0 {
					t.Fatal("actual headless lifecycle frames/free incomplete")
				}
				t.Log("PARTICLE MULDIV HEADLESS: real factory/draw/damping/bounce/free and 8 exact raster/upload frames completed; callback/result/registration hooks=0")
			})
		}
	}
}
