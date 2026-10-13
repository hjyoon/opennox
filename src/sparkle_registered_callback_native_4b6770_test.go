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

const registeredSparkleChild4B6770 = "OPENNOX_TEST_REGISTERED_SPARKLE_CALLBACK_R163"

var registeredSparkleStyles4B6770 = []string{"spawn", "half", "last-positive", "expired", "zero-life", "negative-life", "wrap", "negative-z", "left-edge", "right-edge", "right-inner", "outside"}

// The actual registered C entries must match the existing native world route.
// This freezes real lifetime/RNG/pool/raster behavior, not an independent
// Windows pixel oracle. Children preserve old C pointer faults as evidence.
func TestRegisteredSparkleCallbackNative4B6770(t *testing.T) {
	if name := os.Getenv(registeredSparkleChild4B6770); name != "" {
		testRegisteredSparkleChild4B6770(t, name)
		return
	}
	for _, kind := range []string{"magic", "pixie"} {
		for _, palette := range []string{"white", "colored"} {
			for _, style := range registeredSparkleStyles4B6770 {
				name := kind + "/" + palette + "/" + style
				t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
					cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredSparkleCallbackNative4B6770$", "-test.v")
					cmd.Env = append(os.Environ(), registeredSparkleChild4B6770+"="+name)
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("registered native sparkle callback failed: %v\n%s", err, out)
					}
					if !bytes.Contains(out, []byte("REGISTERED SPARKLE CALLBACK PASS: "+name)) {
						t.Fatalf("missing completed lifetime/RNG/raster/upload receipt: %s", out)
					}
					t.Logf("REGISTERED SPARKLE CALLBACK PASS: %s", name)
				})
			}
		}
	}
}

func registeredSparkleTypesLoad4B6770(t *testing.T, c *Client) {
	t.Helper()
	for _, row := range []struct {
		name, draw string
		callback   unsafe.Pointer
	}{
		{"MagicSparkle", "MagicSparkleDraw", legacy.Get_nox_thing_magic_sparkle_draw()},
		{"PixieDust", "PixieDustDraw", legacy.Get_nox_thing_pixie_dust_draw()},
	} {
		var raw []byte
		for _, value := range []string{row.name, "DRAW", row.draw} {
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
		typ := c.Things.TypeByID(row.name)
		if typ == nil || typ.DrawFunc != row.callback || !client.IsRegisteredDrawFunc(row.callback) || typ.DrawData != nil {
			t.Fatal("actual kind0 sparkle callback was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
	}
}

func testRegisteredSparkleChild4B6770(t *testing.T, name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 3 || !slices.Contains([]string{"magic", "pixie"}, parts[0]) || !slices.Contains([]string{"white", "colored"}, parts[1]) || !slices.Contains(registeredSparkleStyles4B6770, parts[2]) {
		t.Fatal("unknown frozen sparkle case")
	}
	kind, palette, style := parts[0], parts[1], parts[2]
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	frame := uint32(100)
	if style == "wrap" {
		frame = 0xffffffff
	}
	s.SetFrame(frame)
	s.Rand.Logic.Reset(93)
	// Table indices 0 and 1 select the two palette branches, and their next
	// words both yield radius roll 4. They are identical for both routes.
	seed := 0
	if palette == "colored" {
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
	registeredSparkleTypesLoad4B6770(t, c)
	beforeTypes := make(map[*client.ObjectType][]byte)
	c.Things.Each(func(typ *client.ObjectType) { beforeTypes[typ] = registeredEffectBytes4B7540(typ) })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

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
		switch style {
		case "left-edge":
			pos.X = 110
		case "right-edge":
			pos.X = 218
		case "right-inner":
			pos.X = 217
		case "outside":
			pos.X = 400
		}
		typeName := "MagicSparkle"
		if kind == "pixie" {
			typeName = "PixieDust"
		}
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID(typeName), pos)
		if dr == nil {
			t.Fatal("actual native allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		dr.ZVal, dr.ZVal2 = 8, 3
		effect := dr.UnionEffect()
		effect.Field_108, effect.Field_109, effect.Field_110 = 0xaabbccdd, 0xabcddcba, 0x01020304
		effect.Field_111, effect.Field_112 = frame-10, frame+10
		switch style {
		case "spawn":
			effect.Field_111, effect.Field_112 = frame, frame+20
		case "last-positive":
			effect.Field_111, effect.Field_112 = frame-19, frame+1
		case "expired":
			effect.Field_111, effect.Field_112 = frame-20, frame
		case "zero-life":
			effect.Field_111, effect.Field_112 = frame, frame
		case "negative-life":
			effect.Field_111, effect.Field_112 = frame+20, frame+5
		case "negative-z":
			dr.ZVal, dr.ZVal2 = 0xfffe, 0xfffd
		}
		beforeDR := registeredEffectBytes4B7540(dr)
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(dr, vp)
		}
		wantResult, wantCount, steps := 1, 1, 2
		if style == "expired" || style == "zero-life" || style == "negative-life" {
			wantResult, wantCount, steps = 0, 0, 1
		} else if style == "outside" || style == "right-edge" {
			steps = 1
		}
		if result != wantResult || cl.Objs.Count != wantCount || s.Rand.Other.Index() != seed+steps {
			t.Fatalf("real return/lifetime/RNG differs: result=%d count=%d Other=%d want=%d/%d/%d", result, cl.Objs.Count, s.Rand.Other.Index(), wantResult, wantCount, seed+steps)
		}
		if result == 0 && slices.Contains(cl.Objs.AllList1(), dr) {
			t.Fatal("naturally deleted sparkle remains in actual live list")
		}
		if result != 0 && !bytes.Equal(beforeDR, registeredEffectBytes4B7540(dr)) {
			t.Fatal("sparkle draw mutated a live drawable")
		}
		got := observation{Result: result, Count: cl.Objs.Count, Other: s.Rand.Other.Index(), Graph: registeredRainGraph4B7310(t, c), Draw: registeredEffectBytes4B7540(d), Pixels: slices.Clone(pix.Pix)}
		wantPixels := steps == 2 && style != "last-positive"
		changed := slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 })
		if changed != wantPixels {
			t.Fatalf("natural sparkle raster visibility differs: %t want %t", changed, wantPixels)
		}
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 {
			t.Fatal("callback mutated viewport/frame/tick-rate/Logic RNG")
		}
		for typ, raw := range beforeTypes {
			if !bytes.Equal(raw, registeredEffectBytes4B7540(typ)) {
				t.Fatal("callback mutated shared type metadata")
			}
		}
		for cl.Objs.List1 != nil {
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(cl.Objs.List1)
		}
		if cl.Objs.Count != 0 || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("actual pool/list cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered sparkle differs from existing world route: return/count=%d/%d want=%d/%d graph=%t draw=%t pixels=%t RNG=%d/%d", got.Result, got.Count, want.Result, want.Count, reflect.DeepEqual(got.Graph, want.Graph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels), got.Other, want.Other)
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED SPARKLE CALLBACK PASS: %s active=%d actual-C-entry/high-native-types-and-records/both-palettes/RNG/natural-deletion/full-raster/headless-upload", name, got.Count))
}
