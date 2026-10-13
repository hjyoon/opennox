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

const registeredSimpleProjectileChild4B9D70 = "OPENNOX_TEST_REGISTERED_SIMPLE_PROJECTILE_CALLBACK_R166"

var registeredSimpleProjectileStyles4B9D70 = []string{"horizontal", "vertical", "equal", "reverse", "zero-length", "negative-prev", "negative-z", "clip-corner", "edge-right", "outside", "frame-wrap", "expired-deadline"}

// Actual registered C entries must match the existing native world geometry,
// full renderer state/raster and immutable drawable/type/viewport/RNG state.
// Types are a declared service fixture, not an independent Windows pixel oracle.
// Child processes retain the old native C pointer faults as evidence.
func TestRegisteredSimpleProjectileCallbackNative4B9D70(t *testing.T) {
	if name := os.Getenv(registeredSimpleProjectileChild4B9D70); name != "" {
		testRegisteredSimpleProjectileChild4B9D70(t, name)
		return
	}
	for _, kind := range []string{"spit", "powder"} {
		for _, style := range registeredSimpleProjectileStyles4B9D70 {
			name := kind + "/" + style
			t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredSimpleProjectileCallbackNative4B9D70$", "-test.v")
				cmd.Env = append(os.Environ(), registeredSimpleProjectileChild4B9D70+"="+name)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("registered native simple projectile callback failed: %v\n%s", err, out)
				}
				if !bytes.Contains(out, []byte("REGISTERED SIMPLE PROJECTILE CALLBACK PASS: "+name)) {
					t.Fatalf("missing completed geometry/raster/upload receipt: %s", out)
				}
				t.Logf("REGISTERED SIMPLE PROJECTILE CALLBACK PASS: %s", name)
			})
		}
	}
}

func registeredSimpleProjectileTypes4B9D70(t *testing.T, c *Client) {
	t.Helper()
	for _, row := range []struct {
		name, draw string
		callback   unsafe.Pointer
	}{
		{"SpiderSpit", "SpiderSpitDraw", legacy.Get_nox_thing_spider_spit_draw()},
		{"BlackPowder", "BlackPowderDraw", legacy.Get_nox_thing_black_powder_draw()},
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
			t.Fatal("actual kind0 projectile callback was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
	}
}

func testRegisteredSimpleProjectileChild4B9D70(t *testing.T, name string) {
	kind, style, ok := strings.Cut(name, "/")
	if !ok || !slices.Contains([]string{"spit", "powder"}, kind) || !slices.Contains(registeredSimpleProjectileStyles4B9D70, style) {
		t.Fatal("unknown frozen simple projectile case")
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
	s.Rand.Other.Reset(17)
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
	registeredSimpleProjectileTypes4B9D70(t, c)
	beforeTypes := make(map[*client.ObjectType][]byte)
	c.Things.Each(func(typ *client.ObjectType) { beforeTypes[typ] = registeredEffectBytes4B7540(typ) })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	type observation struct {
		Result, Count int
		Graph         [][]byte
		Draw          []byte
		Pixels        []uint16
	}
	run := func(direct bool) observation {
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
		case "clip-corner":
			pos = vp.World.Min
		case "edge-right":
			pos.X = 227
		case "outside":
			pos.X = 400
		}
		typeName := "SpiderSpit"
		if kind == "powder" {
			typeName = "BlackPowder"
		}
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID(typeName), pos)
		if dr == nil {
			t.Fatal("actual native allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		dr.ZVal, dr.ZVal2 = 8, 3
		previous := pos.Sub(image.Pt(12, 8))
		switch style {
		case "vertical":
			previous = pos.Sub(image.Pt(3, 20))
		case "equal":
			previous = pos.Sub(image.Pt(12, 12))
		case "reverse":
			previous = pos.Add(image.Pt(12, 20))
		case "zero-length":
			previous = pos
		case "negative-prev":
			previous = image.Pt(-5, -10)
		case "negative-z":
			dr.ZVal, dr.ZVal2 = 0xfffe, 0xfffd
		case "clip-corner":
			dr.ZVal, dr.ZVal2 = 0, 0
		}
		dr.Field_8, dr.Field_9 = uint32(int32(previous.X)), uint32(int32(previous.Y))
		dr.Deadline = frame + 30
		if style == "expired-deadline" {
			// Like the original, these draw callbacks do not own deadline removal.
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
		if result != 1 || cl.Objs.Count != 1 || cl.Objs.List1 != dr || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil || !bytes.Equal(beforeDR, registeredEffectBytes4B7540(dr)) {
			t.Fatal("draw callback mutated native drawable/lifetime/pool/list or return")
		}
		got := observation{Result: result, Count: cl.Objs.Count, Graph: registeredRainGraph4B7310(t, c), Draw: registeredEffectBytes4B7540(d), Pixels: slices.Clone(pix.Pix)}
		wantPixels := style != "outside"
		changed := slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 })
		if changed != wantPixels {
			t.Fatalf("natural projectile raster visibility differs: %t want %t", changed, wantPixels)
		}
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 || s.Rand.Other.Index() != 17 {
			t.Fatal("callback mutated viewport/frame/tick-rate/Logic/Other RNG")
		}
		for typ, raw := range beforeTypes {
			if !bytes.Equal(raw, registeredEffectBytes4B7540(typ)) {
				t.Fatal("callback mutated shared type metadata")
			}
		}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("actual pool/list cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered simple projectile differs from native world route: graph=%t draw=%t pixels=%t", reflect.DeepEqual(got.Graph, want.Graph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED SIMPLE PROJECTILE CALLBACK PASS: %s actual-C-entry/high-native-types-and-records/geometry-gray-outline-white-core-powder-marks/full-raster/headless-upload", name))
}
