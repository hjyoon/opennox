//go:build !server

package opennox

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const registeredEffectChild4B7540 = "OPENNOX_TEST_REGISTERED_EFFECT_CALLBACK_R156"

var registeredEffectCases4B7540 = []string{
	"bubble-grow", "bubble-shrink", "bubble-deadline", "bubble-remove-before-draw", "bubble-remove-after-draw", "bubble-negative-altitude",
	"vortex-front", "vortex-back", "vortex-wrap", "vortex-expiry", "vortex-outside", "vortex-boundary",
}

// Compare the real registered C callback/Drawable.CallDraw to the existing
// production world dispatch. This proves routing, state, raster and lifetime
// equivalence, not an independent original Windows pixel oracle. Children keep
// the old C pointer-truncation crash observable without killing the test runner.
func TestRegisteredEffectCallbackNative4B7540(t *testing.T) {
	if name := os.Getenv(registeredEffectChild4B7540); name != "" {
		testRegisteredEffectChild4B7540(t, name)
		return
	}
	for _, name := range registeredEffectCases4B7540 {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredEffectCallbackNative4B7540$", "-test.v")
			cmd.Env = append(os.Environ(), registeredEffectChild4B7540+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native callback failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED EFFECT CALLBACK PASS: "+name)) {
				t.Fatalf("missing completed real callback/raster/upload receipt: %s", out)
			}
			t.Logf("REGISTERED EFFECT CALLBACK PASS: %s", name)
		})
	}
}

func registeredEffectBytes4B7540[T any](p *T) []byte {
	return slices.Clone(unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))))
}

func testRegisteredEffectChild4B7540(t *testing.T, name string) {
	if !slices.Contains(registeredEffectCases4B7540, name) {
		t.Fatal("unknown frozen effect case")
	}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(1)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	pix := noximage.NewImage16(image.Rect(0, 0, 128, 128))
	r.SetPixBuffer(pix)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	oldClient := noxClient
	noxClient = c
	t.Cleanup(func() { noxClient = oldClient })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World = image.Rect(100, 200, 228, 328)
	vp.Screen = pix.Rect
	vp.Size = pix.Size()
	initialVP := registeredEffectBytes4B7540(vp)

	type observation struct {
		result, count      int
		drawable, drawData []byte
		pixels             []uint16
	}
	run := func(direct bool) observation {
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetAlpha(128)
		d.SetAlphaEnabled(false)
		d.SetClip(true)
		d.SetClipRect(pix.Rect)
		d.SetClipRect2(pix.Rect)
		dr := cl.Objs.New()
		if dr == nil {
			t.Fatal("real native drawable pool exhausted")
		}
		cl.Objs.Count = 1
		cl.Objs.List1 = dr
		dr.ObjFlags = 0x1000000
		dr.PosVec = image.Pt(164, 264)
		dr.ZVal = 8
		effect := dr.UnionEffect()
		effect.Field_108, effect.Field_109 = 0x03e0, 0x001f
		if strings.HasPrefix(name, "bubble-") {
			dr.DrawFuncPtr = legacy.Get_nox_thing_bubble_draw()
			effect.Field_110 = 4 | 1<<8 | 1<<16 | 7<<24
			effect.Field_111 = 3 | 1<<8 | uint32(byte(0xfe))<<16
			switch name {
			case "bubble-shrink":
				effect.Field_110 = 1 | 2<<8 | 1<<16 | 6<<24
			case "bubble-deadline":
				dr.Deadline = 1
			case "bubble-remove-before-draw":
				effect.Field_110 = 3<<8 | 4<<16 | 4<<24
			case "bubble-remove-after-draw":
				effect.Field_110 = 1 | 3<<8 | 1<<16 | 4<<24
			case "bubble-negative-altitude":
				dr.ZVal = 1
			}
		} else {
			dr.DrawFuncPtr = legacy.Get_nox_thing_vortex_draw()
			effect.Field_110, effect.Field_111 = 164, 264
			effect.Field_112 = 0 | 2<<8 | 10<<16 | 1<<24
			switch name {
			case "vortex-back":
				effect.Field_112 |= 192
			case "vortex-wrap":
				effect.Field_112 = 1 | uint32(byte(0xfe))<<8 | 10<<16 | 1<<24
			case "vortex-expiry":
				dr.ZVal = 399
				// Keep the lifted point inside the actual viewport.
				effect.Field_111 = 664
			case "vortex-outside":
				effect.Field_110 = 300
			case "vortex-boundary":
				effect.Field_110 = 90
			}
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("native records must be above 4 GiB")
		}
		logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			var handled bool
			result, handled = c.callBubbleVortexDraw4B7540(dr, vp)
			if !handled {
				t.Fatal("existing world dispatch rejected supported effect")
			}
		}
		got := observation{result: result, count: cl.Objs.Count, drawData: registeredEffectBytes4B7540(d), pixels: slices.Clone(pix.Pix)}
		if cl.Objs.Count == 1 {
			got.drawable = registeredEffectBytes4B7540(dr)
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		}
		if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("real drawable lifetime/list cleanup incomplete")
		}
		if !bytes.Equal(initialVP, registeredEffectBytes4B7540(vp)) || s.Frame() != 1 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
			t.Fatal("callback mutated viewport/frame/RNG")
		}
		return got
	}
	want := run(false)
	got := run(true)
	if got.result != want.result || got.count != want.count || !bytes.Equal(got.drawable, want.drawable) || !bytes.Equal(got.drawData, want.drawData) || !slices.Equal(got.pixels, want.pixels) {
		t.Fatalf("registered callback differs from production world route: return/count=%d/%d want=%d/%d drawable=%t draw-data=%t pixels=%t", got.result, got.count, want.result, want.count, bytes.Equal(got.drawable, want.drawable), bytes.Equal(got.drawData, want.drawData), slices.Equal(got.pixels, want.pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED EFFECT CALLBACK PASS: %s return=%d active=%d actual-C-dispatch/high-native-records/full-raster/headless-upload", name, got.result, got.count))
}
