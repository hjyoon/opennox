//go:build !server

package opennox

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"reflect"
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
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const registeredPixieChild4B6E80 = "OPENNOX_TEST_REGISTERED_PIXIE_CALLBACK_R164"

var registeredPixieStyles4B6E80 = []string{"short-tail", "long-tail", "squared-400", "squared-401", "negative-prev", "z-min", "z-max", "negative-z", "z-over-max", "left-edge", "top-edge", "right-edge", "bottom-edge", "outside", "frame-wrap", "expired-deadline"}

// The actual registered C callback must match the existing native world route,
// including signed altitude mutation, tail geometry, RNG, clip and full raster.
// This is a native service fixture, not an independent Windows pixel oracle.
// Children keep the original native C pointer faults observable.
func TestRegisteredPixieCallbackNative4B6E80(t *testing.T) {
	if name := os.Getenv(registeredPixieChild4B6E80); name != "" {
		testRegisteredPixieChild4B6E80(t, name)
		return
	}
	for _, roll := range []string{"rise", "fall"} {
		for _, style := range registeredPixieStyles4B6E80 {
			name := roll + "/" + style
			t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredPixieCallbackNative4B6E80$", "-test.v")
				cmd.Env = append(os.Environ(), registeredPixieChild4B6E80+"="+name)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("registered native pixie callback failed: %v\n%s", err, out)
				}
				if !bytes.Contains(out, []byte("REGISTERED PIXIE CALLBACK PASS: "+name)) {
					t.Fatalf("missing completed altitude/RNG/raster/upload receipt: %s", out)
				}
				t.Logf("REGISTERED PIXIE CALLBACK PASS: %s", name)
			})
		}
	}
}

func registeredPixieTypeLoad4B6E80(t *testing.T, c *Client) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, value := range []string{"Pixie", "DRAW", "PixieDraw"} {
		raw = append(raw, byte(len(value)))
		raw = append(raw, value...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	raw = append(raw, 0)
	buf, _ := alloc.CloneSlice(raw)
	f := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	scratch, freeScratch := alloc.Make([]byte{}, 256)
	err := c.Things.ReadType(f, scratch)
	remaining := len(f.Data())
	freeScratch()
	f.Free()
	if err != nil || remaining != 0 {
		t.Fatalf("actual DRAW parser failed: %v unread=%d", err, remaining)
	}
	typ := c.Things.TypeByID("Pixie")
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_pixie_draw() || !client.IsRegisteredDrawFunc(typ.DrawFunc) || typ.DrawData != nil {
		t.Fatal("actual kind0 PixieDraw callback was not installed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
		t.Fatal("actual native type must be above 4 GiB")
	}
	return typ
}

func testRegisteredPixieChild4B6E80(t *testing.T, name string) {
	roll, style, ok := strings.Cut(name, "/")
	if !ok || !slices.Contains([]string{"rise", "fall"}, roll) || !slices.Contains(registeredPixieStyles4B6E80, style) {
		t.Fatal("unknown frozen pixie case")
	}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	frame := uint32(100)
	if style == "frame-wrap" {
		frame = 0xffffffff
	}
	s.SetFrame(frame)
	s.Rand.Logic.Reset(93)
	// The unchanged Nox table at indices 0/1 yields 53/38 in [0,100].
	seed := 0
	if roll == "fall" {
		seed = 1
	}
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
	t.Cleanup(cl.Objs.Free)
	cl.Things.Nox_things_free_44C580_B()
	t.Cleanup(cl.Things.Nox_things_free_44C580_B)
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
	typ := registeredPixieTypeLoad4B6E80(t, c)
	beforeType := registeredEffectBytes4B7540(typ)
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	z := int16(8)
	switch style {
	case "z-min":
		z = 0
	case "z-max":
		z = 35
	case "negative-z":
		z = -2
	case "z-over-max":
		z = 36
	}
	wantZ := z
	if roll == "rise" {
		if wantZ < 35 {
			wantZ++
		}
	} else if wantZ > 0 {
		wantZ--
	}

	type observation struct {
		Result, Count, Other int
		Graph                [][]byte
		Draw                 []byte
		Pixels               []uint16
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
		pos := image.Pt(164, 280)
		z2 := int16(3)
		switch style {
		case "left-edge":
			pos.X = 110
		case "top-edge":
			pos.Y = 210 + int(z2) + int(wantZ)
		case "right-edge":
			pos.X = 218
		case "bottom-edge":
			pos.Y = 318 + int(z2) + int(wantZ)
		case "outside":
			pos.X = 400
		case "negative-z":
			z2 = -3
		}
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("Pixie"), pos)
		if dr == nil {
			t.Fatal("actual native allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		dr.ZVal, dr.ZVal2 = uint16(z), uint16(z2)
		previous := pos.Sub(image.Pt(12, 8))
		switch style {
		case "long-tail":
			previous = pos.Sub(image.Pt(50, 40))
		case "squared-400":
			previous = pos.Sub(image.Pt(20, int(z2)))
		case "squared-401":
			previous = pos.Sub(image.Pt(20, int(z2)+1))
		case "negative-prev":
			previous = image.Pt(-5, -10)
		}
		dr.Field_8, dr.Field_9 = uint32(int32(previous.X)), uint32(int32(previous.Y))
		dr.Deadline = frame + 30
		if style == "expired-deadline" {
			// Original PixieDraw does not own the update-time deadline removal.
			dr.Deadline = frame - 1
		}
		dr.UnionEffect().Field_108, dr.UnionEffect().Field_111 = 0xaabbccdd, 0xabcddcba
		beforeDR := registeredEffectBytes4B7540(dr)
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(dr, vp)
		}
		if result != 1 || cl.Objs.Count != 1 || cl.Objs.List1 != dr || s.Rand.Other.Index() != seed+1 || dr.ZVal != uint16(wantZ) {
			t.Fatalf("real altitude/return/lifetime/RNG differs: Z=%d result=%d count=%d Other=%d want=%d/1/1/%d", int16(dr.ZVal), result, cl.Objs.Count, s.Rand.Other.Index(), wantZ, seed+1)
		}
		afterDR := registeredEffectBytes4B7540(dr)
		zOffset, zSize := int(unsafe.Offsetof(dr.ZVal)), int(unsafe.Sizeof(dr.ZVal))
		for i, v := range beforeDR {
			if (i < zOffset || i >= zOffset+zSize) && afterDR[i] != v {
				t.Fatalf("pixie draw mutated unrelated native drawable byte %d", i)
			}
		}
		got := observation{Result: result, Count: cl.Objs.Count, Other: s.Rand.Other.Index(), Graph: registeredRainGraph4B7310(t, c), Draw: registeredEffectBytes4B7540(d), Pixels: slices.Clone(pix.Pix)}
		wantPixels := style != "right-edge" && style != "bottom-edge" && style != "outside"
		changed := slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 })
		if changed != wantPixels {
			t.Fatalf("natural pixie raster visibility differs: %t want %t", changed, wantPixels)
		}
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 {
			t.Fatal("callback mutated viewport/type/frame/tick-rate/Logic RNG")
		}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("actual pool/list cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered pixie differs from native world route: graph=%t draw=%t pixels=%t RNG=%d/%d", reflect.DeepEqual(got.Graph, want.Graph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels), got.Other, want.Other)
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED PIXIE CALLBACK PASS: %s actual-C-entry/high-native-types-and-records/signed-altitude/clipped-tail/RNG/full-raster/headless-upload", name))
}
