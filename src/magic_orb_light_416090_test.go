//go:build !server

package opennox

import (
	"fmt"
	"image"
	"math"
	"math/big"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/prand"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Independent big.Float instruction model: the original binary32 scale,
// retained Chop53 multiplies, then the caller's Chop24 FSTPS. No production
// RNG precision/spill helper participates in the expected result.
func magicOrbReferenceIntensity416090(word int) float32 {
	scaled := new(big.Float).SetPrec(53).SetMode(big.ToZero)
	scaled.Mul(new(big.Float).SetFloat64(float64(word)), new(big.Float).SetFloat64(float64(math.Float32frombits(0x38000100))))
	retained := new(big.Float).SetPrec(53).SetMode(big.ToZero)
	retained.Mul(scaled, new(big.Float).SetFloat64(100))
	stored := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(retained)
	value, _ := stored.Float32()
	if value > 63 {
		value = 63
	}
	return value
}

func magicOrbFixture416090(t *testing.T, glow bool) (*Client, *client.Drawable, *noxrender.Viewport, *noximage.Image16) {
	t.Helper()
	r := noxrender.NewRender(nil, nil)
	t.Cleanup(r.Part.Free)
	data, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(data)
	r.Part.RenderGlow = glow
	data.SetClip(true)
	data.SetClipRect(image.Rect(2, 3, 62, 63))
	data.SetClipRect2(image.Rect(2, 3, 61, 62))
	data.SetRect3(image.Rect(2, 3, 62, 63))
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	dr, freeDrawable := alloc.New(client.Drawable{})
	t.Cleanup(freeDrawable)
	*dr = client.Drawable{PosVec: image.Pt(132, 241), ZVal: 12, ZVal2: 4}
	vp, freeViewport := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeViewport)
	*vp = noxrender.Viewport{Screen: image.Rect(2, 3, 62, 63), World: image.Rect(100, 200, 160, 260), Size: image.Pt(60, 60), Field10: 7, Field11: 9, Jiggle12: -3}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{"drawable": dr.C(), "render-data": data.C(), "viewport": vp.C()} {
			if uintptr(pointer) <= uintptr(^uint32(0)) {
				t.Fatalf("%s is not a real high-address allocation: %p", name, pointer)
			}
		}
	}
	if got := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -16)); got != image.Pt(34, 28) {
		t.Fatalf("fixture failed to initialize real native drawable/viewport: point=%v", got)
	}
	world := new(server.Server)
	world.Rand.Other, world.Rand.Logic = prand.New(0), prand.New(93)
	return &Client{srv: &Server{Server: world}, r: NewNoxRender(r)}, dr, vp, pix
}

func magicOrbCallback416090(missile bool) unsafe.Pointer {
	if missile {
		return legacy.Get_nox_thing_magic_missle_draw()
	}
	return legacy.Get_nox_thing_magic_draw()
}

func TestMagicOrb416090ActualCallbackWholeTable(t *testing.T) {
	for _, missile := range []bool{false, true} {
		for _, glow := range []bool{false, true} {
			t.Run(fmt.Sprintf("missile=%t/glow=%t", missile, glow), func(t *testing.T) {
				c, dr, vp, pix := magicOrbFixture416090(t, glow)
				dr.DrawFuncPtr = magicOrbCallback416090(missile)
				if dr.DrawFuncPtr == nil {
					t.Fatal("real legacy draw callback is absent")
				}
				originalView := *vp
				clamped := 0
				for index := 0; index < 4096; index++ {
					clear(pix.Pix)
					c.srv.Rand.Other.Reset(index)
					table := prand.New(index)
					radius := 1 + table.Int(0, 0x7fff)%4
					word := table.Int(0, 0x7fff)
					wantIntensity := magicOrbReferenceIntensity416090(word)
					if wantIntensity == 63 {
						clamped++
					}
					before := *dr
					result, handled := c.callMagicDrawableDraw4B98A0(dr, vp)
					if result != 1 || !handled || *vp != originalView {
						t.Fatalf("actual callback return/viewport: %d/%t view=%+v", result, handled, *vp)
					}
					if math.Float32bits(dr.LightIntensity) != math.Float32bits(wantIntensity) {
						t.Fatalf("start=%d light word=%d: brightness=%08x want original FSTPS=%08x", index, word, math.Float32bits(dr.LightIntensity), math.Float32bits(wantIntensity))
					}
					if c.srv.Rand.Other.Index() != (index+2)%4096 || c.srv.Rand.Logic.Index() != 93 {
						t.Fatalf("start=%d: radius then light must consume exactly two Other words, no Logic: %d/%d", index, c.srv.Rand.Other.Index(), c.srv.Rand.Logic.Index())
					}
					before.LightFlags = 2
					before.LightColor = noxrender.RGB{R: 200, G: 200, B: 255}
					if missile {
						before.LightColor = noxrender.RGB{R: 255, G: 180, B: 50}
					}
					before.LightIntensity = wantIntensity
					before.LightIntensityU16 = uint32(wantIntensity*65536 + 0.5)
					// The unchanged light-radius implementation has its own tests;
					// this comparison checks only that the actual callback applies it.
					before.LightIntensityRad = uint32(client.LightRadius(wantIntensity))
					if !reflect.DeepEqual(*dr, before) {
						t.Fatalf("start=%d: callback changed unrelated drawable fields", index)
					}
					white := uint16(noxcolor.RGB5551Color(255, 255, 255))
					for y := 28 - radius/2; y < 28-radius/2+radius; y++ {
						for x := 34 - radius/2; x < 34-radius/2+radius; x++ {
							if pix.Pix[pix.PixOffset(x, y)] != white {
								t.Fatalf("start=%d: original radius %d white rectangle missing at %d,%d", index, radius, x, y)
							}
						}
					}
				}
				if clamped == 0 {
					t.Fatal("whole-table test never exercised original >63 light clamp")
				}
			})
		}
	}
}

func TestMagicOrb416090PixelsAndRadiusOrder(t *testing.T) {
	for _, missile := range []bool{false, true} {
		for _, glow := range []bool{false, true} {
			for radius := 1; radius <= 4; radius++ {
				t.Run(fmt.Sprintf("missile=%t/glow=%t/radius=%d", missile, glow, radius), func(t *testing.T) {
					c, dr, vp, pix := magicOrbFixture416090(t, glow)
					ref, _, _, refPix := magicOrbFixture416090(t, glow)
					index := 0
					for ; index < 4096; index++ {
						if 1+prand.New(index).Int(0, 0x7fff)%4 == radius {
							break
						}
					}
					if index == 4096 {
						t.Fatal("original table has no requested radius")
					}
					c.srv.Rand.Other.Reset(index)
					dr.DrawFuncPtr = magicOrbCallback416090(missile)
					if result, handled := c.callMagicDrawableDraw4B98A0(dr, vp); result != 1 || !handled {
						t.Fatalf("actual callback=%d/%t", result, handled)
					}
					// Literal original arguments, independent of all magic draw
					// helpers. This preserves the native pixel path; it is not a
					// claim of a complete Windows-renderer screenshot golden.
					color := noxcolor.RGB5551Color(0, 200, 255)
					if missile {
						color = noxcolor.RGB5551Color(255, 255, 50)
					}
					ref.r.DrawGlow(image.Pt(34, 28), color, 2*radius+1, radius/2+3)
					white := noxcolor.RGB5551Color(255, 255, 255)
					ref.r.Data().SetColor2(white)
					ref.r.DrawRectFilledOpaque(34-radius/2, 28-radius/2, radius, radius, white)
					if !slices.Equal(pix.Pix, refPix.Pix) || c.srv.Rand.Other.Index() != (index+2)%4096 || c.srv.Rand.Logic.Index() != 93 {
						t.Fatal("glow/white-rectangle pixels or radius/light stream order changed")
					}
				})
			}
		}
	}
}

func TestMagicOrb416090ClipBoundariesNoStep(t *testing.T) {
	for _, missile := range []bool{false, true} {
		for _, test := range []struct {
			name    string
			point   image.Point
			visible bool
		}{
			{"left-outside", image.Pt(11, 28), false}, {"left-equal", image.Pt(12, 28), true},
			{"top-outside", image.Pt(34, 12), false}, {"top-equal", image.Pt(34, 13), true},
			{"right-equal", image.Pt(52, 28), false}, {"right-inside", image.Pt(51, 28), true},
			{"bottom-equal", image.Pt(34, 53), false}, {"bottom-inside", image.Pt(34, 52), true},
		} {
			t.Run(fmt.Sprintf("missile=%t/%s", missile, test.name), func(t *testing.T) {
				c, dr, vp, pix := magicOrbFixture416090(t, true)
				dr.DrawFuncPtr = magicOrbCallback416090(missile)
				dr.PosVec = test.point.Sub(image.Pt(2, 3)).Add(image.Pt(100, 216))
				c.srv.Rand.Other.Reset(4095)
				before, view, data := *dr, *vp, *c.r.Data()
				pixels := slices.Clone(pix.Pix)
				if result, handled := c.callMagicDrawableDraw4B98A0(dr, vp); result != 1 || !handled || *vp != view {
					t.Fatalf("callback/view=%d/%t/%+v", result, handled, *vp)
				}
				if test.visible {
					if c.srv.Rand.Other.Index() != 1 || dr.LightFlags != 2 || slices.Equal(pix.Pix, pixels) {
						t.Fatal("visible boundary did not draw/consume exactly two words through wrap")
					}
				} else if !reflect.DeepEqual(*dr, before) || !reflect.DeepEqual(*c.r.Data(), data) || !slices.Equal(pix.Pix, pixels) || c.srv.Rand.Other.Index() != 4095 {
					t.Fatal("clipped callback mutated drawable/render state/pixels or advanced Other")
				}
				if c.srv.Rand.Logic.Index() != 93 {
					t.Fatal("magic renderer advanced Logic stream")
				}
			})
		}
	}
}

func TestMagicOrb416090SpillOriginalToZero(t *testing.T) {
	// Literal IEEE binary32 ToZero boundaries, independent of Go's nearest
	// float32 conversion (big.Float.Float32 also always rounds to nearest).
	for _, test := range []struct {
		value float64
		bits  uint32
	}{
		{0, 0}, {math.Copysign(0, -1), 0x80000000}, {1, 0x3f800000}, {-1, 0xbf800000},
		{1 + math.Ldexp(3, -25), 0x3f800000}, {-1 - math.Ldexp(3, -25), 0xbf800000},
		{math.Pi, 0x40490fda}, {-math.Pi, 0xc0490fda},
		{62.9999999, 0x427bffff}, {63.0000001, 0x427c0000},
		{99.9999999, 0x42c7ffff}, {-99.9999999, 0xc2c7ffff},
		{math.Ldexp(3, -151), 0}, {-math.Ldexp(3, -151), 0x80000000},
		{float64(math.MaxFloat32) * 2, 0x7f7fffff}, {-float64(math.MaxFloat32) * 2, 0xff7fffff},
	} {
		t.Run(fmt.Sprintf("%016x", math.Float64bits(test.value)), func(t *testing.T) {
			got := magicDrawSpill32_4B98A0(test.value)
			if math.Float32bits(got) != test.bits {
				t.Fatalf("spill=%08x want original ToZero=%08x", math.Float32bits(got), test.bits)
			}
		})
	}
}
