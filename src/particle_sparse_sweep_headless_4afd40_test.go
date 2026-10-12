//go:build !server

package opennox

import (
	"fmt"
	"image"
	"testing"

	noxcolor "github.com/opennox/libs/color"
)

func TestParticleSparseSweepOriginal4AFD40HeadlessOverlap(t *testing.T) {
	palette := [7]uint16{0x7c00, 0x03e0, 0x001f, 0x7fe0, 0x7c1f, 0x03ff, 0x7fff}
	for rotation := 0; rotation < 7; rotation++ {
		t.Run(fmt.Sprintf("rotation-%d", rotation), func(t *testing.T) {
			owner, pix, win, seat := newParticleRegistrationHeadless4B01B0(t)
			for i := 0; i < 7; i++ {
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 32, noxcolor.RGBA5551(palette[(i+rotation)%7]))
				p.pointSize = 2
				particleSparseGap4AFD40(4096)
			}
			want := make([]uint16, len(pix.Pix))
			for y := 89; y < 91; y++ {
				for x := 64; x < 66; x++ {
					want[y*pix.Rect.Dx()+x] = palette[(6+rotation)%7]
				}
			}
			for frame := 0; frame < 16; frame++ {
				clear(pix.Pix)
				owner.Draw()
				presentParticleRegistration4B01B0(t, pix, win, seat, want, frame)
			}
			t.Log("PARTICLE SPARSE PIXELS: 16 actual different-color raster/upload frames with 4096 foreign keys between records")
		})
	}
}

type particleSparseCallbackColor4AFD40 struct{ call func() }

func (c particleSparseCallbackColor4AFD40) RGBA() (r, g, b, a uint32) {
	c.call()
	return 0xffff, 0xffff, 0xffff, 0xffff
}

// A color.Color implementation is itself a callback. Default point drawing
// must not be assumed side-effect-free for an arbitrary implementation.
func TestParticleSparseSweepOriginal4AFD40HeadlessColorCallback(t *testing.T) {
	owner, pix, win, seat := newParticleRegistrationHeadless4B01B0(t)
	created := false
	p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 32, particleSparseCallbackColor4AFD40{call: func() {
		if !created {
			created = true
			particleSparseGap4AFD40(16384)
			child := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 32, noxcolor.RGBA5551(0x001f))
			child.pointSize = 2
		}
	}})
	p.pointSize = 2
	particleSparseGap4AFD40(4096)
	// Ensure traversal is already sparse before the custom color is invoked.
	other := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, 32, noxcolor.RGBA5551(0x03e0))
	other.pointSize = 2
	want := make([]uint16, len(pix.Pix))
	for y := 89; y < 91; y++ {
		for x := 64; x < 66; x++ {
			want[y*pix.Rect.Dx()+x] = 0x001f
		}
	}
	owner.Draw()
	presentParticleRegistration4B01B0(t, pix, win, seat, want, 0)
	if !created || len(owner.byHandle) != 3 {
		t.Fatal("real color callback did not append the current-frame tail")
	}
	t.Log("PARTICLE SPARSE COLOR CALLBACK: actual default point drawing emits a same-frame higher slot; blue tail wins full raster/upload")
}
