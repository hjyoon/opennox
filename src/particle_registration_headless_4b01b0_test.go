//go:build !server

package opennox

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func newParticleRegistrationHeadless4B01B0(t *testing.T) (*partFXes, *noximage.Image16, *render.Renderer, *headless.Seat) {
	t.Helper()
	handles.Init()
	core := noxrender.NewRender(nil, nil)
	t.Cleanup(core.Part.Free)
	data, free := noxrender.NewRenderData()
	t.Cleanup(free)
	core.SetData(data)
	pix := noximage.NewImage16(image.Rect(0, 0, 256, 160))
	core.SetPixBuffer(pix)
	data.SetClip(true)
	data.SetClipRect(pix.Rect)
	data.SetClipRect2(image.Rect(0, 0, 255, 159))
	r := NewNoxRender(core)
	r.partfx.Init(r)
	t.Cleanup(r.partfx.Free)
	previousOffset := partViewportOff
	partViewportOff = image.Pt(7, 11)
	t.Cleanup(func() { partViewportOff = previousOffset })
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	win, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	return &r.partfx, pix, win, sc
}

func particleRegistrationWhiteSquare4B01B0(want []uint16, rect image.Rectangle, pos image.Point) {
	for y := pos.Y; y < pos.Y+2; y++ {
		for x := pos.X; x < pos.X+2; x++ {
			if image.Pt(x, y).In(rect) {
				want[(y-rect.Min.Y)*rect.Dx()+x-rect.Min.X] = 0x7fff
			}
		}
	}
}

func presentParticleRegistration4B01B0(t *testing.T, pix *noximage.Image16, win *render.Renderer, sc *headless.Seat, want []uint16, tick int) {
	t.Helper()
	if !slices.Equal(pix.Pix, want) {
		for i, value := range want {
			if pix.Pix[i] != value {
				t.Fatalf("actual particle raster at tick=%d pixel=(%d,%d): got=%04x original=%04x", tick, i%pix.Rect.Dx(), i/pix.Rect.Dx(), pix.Pix[i], value)
			}
		}
	}
	win.CopyBuffer(pix)
	shown, view := sc.Snapshot()
	if shown == nil || view != pix.Rect || !slices.Equal(shown.Pix, want) {
		t.Fatalf("actual headless upload did not preserve particle frame %d", tick)
	}
}

// Original default drawing is before updating/expiring. Real factory, draw,
// default update and free run without replacing callbacks or calling p.C().
// Damping/bounce and overlapping different-color ordering are separate work.
func TestParticleRegistrationOriginal4B01B0HeadlessDefaultLifecycle(t *testing.T) {
	for _, duration := range []int{0, 1, 3} {
		for _, hidden := range []bool{false, true} {
			t.Run(fmt.Sprintf("duration-%d/hidden-%t", duration, hidden), func(t *testing.T) {
				owner, pix, win, sc := newParticleRegistrationHeadless4B01B0(t)
				p := owner.nox_xxx_particleFxNew_4AF990(image.Pt(71, 101), 1, duration, color.White)
				assertParticleRegistration4B01B0(t, owner, p)
				handle := p.chnd
				p.pointSize, p.xrel16, p.yrel16 = 2, 0x8000, -0x4000
				if hidden {
					p.flags = 8
				}
				want := make([]uint16, len(pix.Pix))
				for tick := 0; tick <= duration+1; tick++ {
					clear(pix.Pix)
					clear(want)
					if tick <= duration && !hidden {
						pos := image.Pt(((71<<16)+tick*0x8000)>>16, (((101<<16)-tick*0x4000)>>16)-1).Sub(image.Pt(7, 11))
						particleRegistrationWhiteSquare4B01B0(want, pix.Rect, pos)
					}
					owner.Draw()
					presentParticleRegistration4B01B0(t, pix, win, sc, want, tick)
					if tick < duration {
						if owner.asParticlefx(handle) != p || p.ticksLeft != duration-tick-1 || p.cnt40 != tick+1 || p.x16 != (71<<16)+(tick+1)*0x8000 || p.y16 != (101<<16)-(tick+1)*0x4000 {
							t.Fatal("real default update differs from pre-expiry original state")
						}
					} else if owner.asParticlefx(handle) != nil || len(owner.byHandle) != 0 || p.cnt40 != duration+1 || p.ticksLeft != -1 || p.x16 != (71<<16)+duration*0x8000 || p.y16 != (101<<16)-duration*0x4000 {
						t.Fatal("expiry did not remove the real slot at the original boundary")
					}
				}
				if sc.PresentCount() != uint64(duration+2) {
					t.Fatal("missing actual headless particle lifecycle frames")
				}
			})
		}
	}
}

type particleType1Reference4B01B0 struct {
	horizontal, delay, vertical     int32
	accel, phase, x, y, left, ticks int32
}

// Literal original 48C5E0/48C610 streams feed 004AEF80; this independent
// DWORD reference follows 004AF060's delay, acceleration, SAR2 and expiry.
// Nothing is read back from the production particle to manufacture a result.
func (r *particleType1Reference4B01B0) step() {
	if r.left < 0 {
		return
	}
	r.ticks++
	if r.delay != 0 {
		r.delay--
		return
	}
	if r.horizontal < 0 {
		r.accel += 4
	} else {
		r.accel -= 4
	}
	r.horizontal += r.accel
	x := (int32(135) << 16) + ((r.horizontal * r.phase) >> 2)
	r.phase += 256
	r.left--
	if r.left >= 0 {
		r.x = x
		r.y += r.vertical
	}
}

func TestParticleRegistrationOriginal4B01B0HeadlessRealType1(t *testing.T) {
	for _, seed := range []uint16{0x1352, 0x8000} {
		t.Run(fmt.Sprintf("seed-%04x", seed), func(t *testing.T) {
			for _, count := range []int{1, 3, 7} {
				t.Run(fmt.Sprintf("count-%d", count), func(t *testing.T) {
					owner, pix, win, sc := newParticleRegistrationHeadless4B01B0(t)
					dr, free := alloc.New(client.Drawable{})
					t.Cleanup(free)
					dr.PosVec = image.Pt(135, 131)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= 0xffffffff {
						t.Fatal("real type-1 drawable is not native high-address C-owned memory")
					}
					before := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))))
					owner.rand.u16 = seed
					initial := [7][3]int32{{48, 60, -27609}, {107, 1, -25261}, {-164, 95, -29331}, {-25, 49, -17187}, {-101, 58, -23672}, {-121, 9, -18250}, {-251, 67, -29577}}
					if seed == 0x8000 {
						initial = [7][3]int32{{48, 60, -24585}, {107, 1, -24669}, {-164, 95, -25091}, {-25, 49, -30230}, {-101, 58, -27270}, {-121, 9, -27571}, {-251, 67, -24897}}
					}
					owner.newPartfxT1(&packetParticleFx{dr: dr, cnt: count, max: 128})
					if len(owner.byHandle) != count {
						t.Fatalf("actual type-1 emitter created %d active slots, original count=%d; no forced C() registration", len(owner.byHandle), count)
					}
					refs := make([]particleType1Reference4B01B0, count)
					particles := make([]*particleFx, count)
					for i := range refs {
						v := initial[i]
						refs[i] = particleType1Reference4B01B0{horizontal: v[0], delay: v[1], vertical: v[2], x: 135 << 16, y: 131 << 16, left: 128}
						for _, p := range owner.byHandle {
							if p.field_68 == int(v[0]) && p.remains44 == int(v[1]) && p.yrel16 == int(v[2]) {
								particles[i] = p
							}
						}
						assertParticleRegistration4B01B0(t, owner, particles[i])
					}
					want := make([]uint16, len(pix.Pix))
					for tick := 0; tick < 260; tick++ {
						clear(pix.Pix)
						clear(want)
						active := 0
						for i := range refs {
							r := &refs[i]
							if r.left >= 0 {
								particleRegistrationWhiteSquare4B01B0(want, pix.Rect, image.Pt(int(r.x>>16)-7, int(r.y>>16)-11))
							}
							r.step()
							if r.left >= 0 {
								active++
							}
						}
						owner.Draw()
						presentParticleRegistration4B01B0(t, pix, win, sc, want, tick)
						if len(owner.byHandle) != active {
							t.Fatalf("real type-1 active count at tick %d: got=%d original=%d", tick, len(owner.byHandle), active)
						}
						for i, p := range particles {
							r := refs[i]
							if p.cnt40 != int(r.ticks) || p.ticksLeft != int(r.left) || p.remains44 != int(r.delay) || p.x16 != int(r.x) || p.y16 != int(r.y) || p.field_68 != int(r.horizontal) || p.field_72 != int(r.accel) || p.field_76 != int(r.phase) || (owner.asParticlefx(p.chnd) == p) != (r.left >= 0) {
								t.Fatalf("actual type-1 DWORD motion/lifetime/free differs: tick=%d particle=%d", tick, i)
							}
						}
						if !bytes.Equal(before, unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))) {
							t.Fatal("real type-1 rendering changed source drawable bytes")
						}
					}
					if sc.PresentCount() != 260 || len(owner.byHandle) != 0 {
						t.Fatal("real emitter did not complete its immutable headless lifecycle")
					}
					t.Logf("PARTICLE HEADLESS: seed=%04x count=%d actual frames=260 complete raster/DWORD state/source bytes verified, factory/render/update/free hooks=0, forced C() registration=0", seed, count)
				})
			}
		})
	}
}
