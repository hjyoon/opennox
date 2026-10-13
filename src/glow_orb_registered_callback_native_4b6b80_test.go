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

const registeredGlowOrbChild4B6B80 = "OPENNOX_TEST_REGISTERED_GLOW_ORB_CALLBACK_R162"

var registeredGlowOrbTypes4B6B80 = []string{"HealOrb", "DrainManaOrb", "CharmOrb", "WhiteOrb", "ManaBombOrb", "WhiteMoveOrb", "BlueMoveOrb", "CustomGlowOrb"}
var registeredGlowOrbStyles4B6B80 = []string{"visible", "outside", "fade-last", "arrival", "wrap"}

// The real C callbacks must match the existing native world route. The type
// registry is an explicit service fixture, not a Windows pixel oracle. Process
// isolation retains the old C pointer faults without killing the runner.
func TestRegisteredGlowOrbCallbackNative4B6B80(t *testing.T) {
	if name := os.Getenv(registeredGlowOrbChild4B6B80); name != "" {
		testRegisteredGlowOrbChild4B6B80(t, name)
		return
	}
	for _, mode := range []string{"still", "move"} {
		for _, typ := range registeredGlowOrbTypes4B6B80 {
			for _, style := range registeredGlowOrbStyles4B6B80 {
				name := mode + "/" + typ + "/" + style
				t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
					cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredGlowOrbCallbackNative4B6B80$", "-test.v")
					cmd.Env = append(os.Environ(), registeredGlowOrbChild4B6B80+"="+name)
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("registered native glow-orb callback failed: %v\n%s", err, out)
					}
					if !bytes.Contains(out, []byte("REGISTERED GLOW ORB CALLBACK PASS: "+name)) {
						t.Fatalf("missing completed movement/lifetime/raster/upload receipt: %s", out)
					}
					t.Logf("REGISTERED GLOW ORB CALLBACK PASS: %s", name)
				})
			}
		}
	}
}

func registeredGlowOrbTypesLoad4B6B80(t *testing.T, c *Client, moving bool) {
	t.Helper()
	drawName, callback := "GlowOrbDraw", legacy.Get_nox_thing_glow_orb_draw()
	if moving {
		drawName, callback = "GlowOrbMoveDraw", legacy.Get_nox_thing_glow_orb_move_draw()
	}
	for _, name := range registeredGlowOrbTypes4B6B80 {
		var raw []byte
		for _, value := range []string{name, "DRAW", drawName} {
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
		typ := c.Things.TypeByID(name)
		if typ == nil || typ.DrawFunc != callback || !client.IsRegisteredDrawFunc(callback) || typ.DrawData != nil {
			t.Fatal("actual kind0 registered callback was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
	}
}

func testRegisteredGlowOrbChild4B6B80(t *testing.T, name string) {
	parts := strings.Split(name, "/")
	if len(parts) != 3 || !slices.Contains([]string{"still", "move"}, parts[0]) || !slices.Contains(registeredGlowOrbTypes4B6B80, parts[1]) || !slices.Contains(registeredGlowOrbStyles4B6B80, parts[2]) {
		t.Fatal("unknown frozen glow-orb case")
	}
	moving, typeName, style := parts[0] == "move", parts[1], parts[2]
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	frame := uint32(100)
	if style == "wrap" {
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
	registeredGlowOrbTypesLoad4B6B80(t, c, moving)
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
		if style == "outside" {
			pos.X = 400
		}
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID(typeName), pos)
		if dr == nil {
			t.Fatal("actual native allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		effect := dr.UnionEffect()
		effect.Field_108 = uint32(204) | uint32(310)<<16
		effect.Field_110 = 8 << 24
		effect.Field_111 = 6 | 2<<8 | 2<<16 | 0xa5<<24
		if style == "fade-last" {
			effect.Field_111 = 1 | 1<<8 | 1<<16 | 0xa5<<24
		}
		if style == "arrival" {
			effect.Field_108 = uint32(pos.X+3) | uint32(pos.Y+4)<<16
		}
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(dr, vp)
		}
		wantResult, wantCount := 1, 1
		if style == "fade-last" || moving && style == "arrival" {
			wantResult, wantCount = 0, 0
		}
		if result != wantResult || cl.Objs.Count != wantCount {
			t.Fatalf("real return/lifetime differs: result=%d count=%d want=%d/%d", result, cl.Objs.Count, wantResult, wantCount)
		}
		if result == 0 && slices.Contains(cl.Objs.AllList1(), dr) {
			t.Fatal("naturally deleted orb remains in actual live list")
		}
		got := observation{Result: result, Count: cl.Objs.Count, Graph: registeredRainGraph4B7310(t, c), Draw: registeredEffectBytes4B7540(d), Pixels: slices.Clone(pix.Pix)}
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.Rand.Logic.Index() != 93 || s.Rand.Other.Index() != 17 {
			t.Fatal("callback mutated viewport/frame/RNG")
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
		t.Fatalf("registered glow orb differs from existing world route: return/count=%d/%d want=%d/%d graph=%t draw=%t pixels=%t", got.Result, got.Count, want.Result, want.Count, reflect.DeepEqual(got.Graph, want.Graph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED GLOW ORB CALLBACK PASS: %s active=%d actual-C-entry/high-native-types-and-records/real-movement-natural-deletion/full-raster/headless-upload", name, got.Count))
}
