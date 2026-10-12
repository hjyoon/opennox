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

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const registeredMagicChild4B98A0 = "OPENNOX_TEST_REGISTERED_MAGIC_CALLBACK_R160"

var registeredMagicKinds4B98A0 = []string{"orb", "missile", "tail", "missile-tail"}
var registeredMagicStyles4B98A0 = []string{"center", "boundary", "negative-z", "outside", "expired", "wrap"}

// Real registered C callbacks must match the existing production world route.
// This compares native routing/state/raster/RNG, not an independent Windows
// pixel oracle. Process isolation keeps the old C pointer fault observable.
func TestRegisteredMagicCallbackNative4B98A0(t *testing.T) {
	if name := os.Getenv(registeredMagicChild4B98A0); name != "" {
		testRegisteredMagicChild4B98A0(t, name)
		return
	}
	for _, kind := range registeredMagicKinds4B98A0 {
		for _, style := range registeredMagicStyles4B98A0 {
			name := kind + "/" + style
			t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredMagicCallbackNative4B98A0$", "-test.v")
				cmd.Env = append(os.Environ(), registeredMagicChild4B98A0+"="+name)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("registered native magic callback failed: %v\n%s", err, out)
				}
				if !bytes.Contains(out, []byte("REGISTERED MAGIC CALLBACK PASS: "+name)) {
					t.Fatalf("missing completed callback/raster/upload receipt: %s", out)
				}
				t.Logf("REGISTERED MAGIC CALLBACK PASS: %s", name)
			})
		}
	}
}

func testRegisteredMagicChild4B98A0(t *testing.T, name string) {
	kind, style, ok := strings.Cut(name, "/")
	if !ok || !slices.Contains(registeredMagicKinds4B98A0, kind) || !slices.Contains(registeredMagicStyles4B98A0, style) {
		t.Fatal("unknown frozen magic case")
	}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	frame := uint32(100)
	if style == "wrap" {
		frame = 0xffffffff
	}
	s.SetFrame(frame)
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
	pix := noximage.NewImage16(image.Rect(0, 0, 144, 144))
	r.SetPixBuffer(pix)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	initialVP := registeredEffectBytes4B7540(vp)

	// Both native routes consume the same explicit fade-table service. Restore
	// all table bytes afterwards; no original Windows color/raster claim is made.
	for _, base := range []uintptr{1312500, 1312756} {
		table := unsafe.Slice(memmap.PtrUint32(0x5D4594, base), 64)
		before := slices.Clone(table)
		t.Cleanup(func() { copy(table, before) })
		for i := range table {
			table[i] = uint32(noxcolor.RGB5551Color(byte(40+i*3), byte(240-i*3), byte(64+i*2)))
		}
	}
	seed := 17 + 92*slices.Index(registeredMagicStyles4B98A0, style)
	s.Rand.Logic.Reset(93)

	type observation struct {
		result, logic, other int
		drawable, drawData   []byte
		pixels               []uint16
	}
	run := func(direct bool) observation {
		s.Rand.Other.Reset(seed)
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetAlpha(64)
		d.SetAlphaEnabled(false)
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		dr := cl.Objs.New()
		if dr == nil {
			t.Fatal("real native drawable pool exhausted")
		}
		cl.Objs.Count, cl.Objs.List1 = 1, dr
		dr.ObjFlags = 0x1000000
		dr.PosVec = image.Pt(164, 280)
		dr.ZVal, dr.ZVal2 = 8, 3
		dr.Deadline = frame + 30
		dr.LightFlags, dr.LightColor = 0x80, noxrender.RGB{R: 11, G: 22, B: 33}
		dr.LightIntensity, dr.LightIntensityRad, dr.LightIntensityU16 = 7, 123, 456
		dr.UnionEffect().Field_108, dr.UnionEffect().Field_109 = 120, 250
		switch kind {
		case "orb":
			dr.DrawFuncPtr = legacy.Get_nox_thing_magic_draw()
		case "missile":
			dr.DrawFuncPtr = legacy.Get_nox_thing_magic_missle_draw()
		case "tail":
			dr.DrawFuncPtr = legacy.Get_nox_thing_magic_tail_link_draw()
		case "missile-tail":
			dr.DrawFuncPtr = legacy.Get_nox_thing_magic_missle_tail_link_draw()
		}
		switch style {
		case "boundary":
			dr.PosVec.X = 110
			dr.Deadline = frame + 1
		case "negative-z":
			dr.ZVal, dr.ZVal2 = 0xfffe, 0xfffd
			dr.Deadline = frame + 200
		case "outside":
			dr.PosVec.X = 400
		case "expired":
			dr.Deadline = frame - 1
		case "wrap":
			dr.Deadline = frame + 10
		}
		if !client.IsRegisteredDrawFunc(dr.DrawFuncPtr) {
			t.Fatal("actual DRAW registry does not contain magic callback")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		before := registeredEffectBytes4B7540(dr)
		logic := s.Rand.Logic.Index()
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			var handled bool
			result, handled = c.callMagicDrawableDraw4B98A0(dr, vp)
			if !handled {
				t.Fatal("existing world dispatch rejected supported magic effect")
			}
		}
		after := registeredEffectBytes4B7540(dr)
		mutable := []struct{ at, size int }{
			{int(unsafe.Offsetof(dr.LightFlags)), int(unsafe.Sizeof(dr.LightFlags))},
			{int(unsafe.Offsetof(dr.LightColor)), int(unsafe.Sizeof(dr.LightColor))},
			{int(unsafe.Offsetof(dr.LightIntensity)), int(unsafe.Sizeof(dr.LightIntensity))},
			{int(unsafe.Offsetof(dr.LightIntensityRad)), int(unsafe.Sizeof(dr.LightIntensityRad))},
			{int(unsafe.Offsetof(dr.LightIntensityU16)), int(unsafe.Sizeof(dr.LightIntensityU16))},
		}
		for i, v := range before {
			allowed := false
			for _, span := range mutable {
				allowed = allowed || i >= span.at && i < span.at+span.size
			}
			if !allowed && after[i] != v {
				t.Fatalf("magic callback changed unrelated drawable byte%d", i)
			}
		}
		if result != 1 || cl.Objs.Count != 1 || cl.Objs.List1 != dr || cl.Objs.DeadlineList != nil || !bytes.Equal(initialVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != logic {
			t.Fatal("callback mutated viewport/frame/tick-rate/Logic-RNG/pool/lifetime or return")
		}
		wantOther := seed
		if (kind == "orb" || kind == "missile") && style != "outside" {
			wantOther += 2
		}
		if s.Rand.Other.Index() != wantOther {
			t.Fatalf("callback consumed incorrect RNG steps: %d, want %d", s.Rand.Other.Index(), wantOther)
		}
		got := observation{result: result, logic: s.Rand.Logic.Index(), other: s.Rand.Other.Index(), drawable: after, drawData: registeredEffectBytes4B7540(d), pixels: slices.Clone(pix.Pix)}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("real native pool cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if got.result != want.result || got.logic != want.logic || got.other != want.other || !bytes.Equal(got.drawable, want.drawable) || !bytes.Equal(got.drawData, want.drawData) || !slices.Equal(got.pixels, want.pixels) {
		t.Fatalf("registered callback differs from production world route: RNG=%d/%d drawable=%t draw-data=%t pixels=%t", got.other, want.other, bytes.Equal(got.drawable, want.drawable), bytes.Equal(got.drawData, want.drawData), slices.Equal(got.pixels, want.pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED MAGIC CALLBACK PASS: %s return=%d Other-index=%d actual-C-dispatch/high-native-records/full-raster/headless-upload", name, got.result, got.other))
}
