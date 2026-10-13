//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
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
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const registeredRainChild4B7310 = "OPENNOX_TEST_REGISTERED_RAIN_CALLBACK_R161"

var registeredRainCases4B7310 = []string{
	"blue/normal", "blue/disabled", "blue/wrap",
	"level/normal", "level/below", "level/wrap",
	"oblivion/normal", "oblivion/below", "oblivion/wrap",
	"white/falling", "white/impact", "white/negative", "white/outside", "white/tiny", "white/wrap",
	"orb-blue/falling", "orb-blue/impact", "orb-blue/negative", "orb-blue/outside", "orb-blue/tiny", "orb-blue/wrap",
}

// Real C entries must match the existing native world implementation. The
// private type registry is a service fixture, not an original Windows pixel
// oracle. Child processes retain the old pointer faults without hiding them.
func TestRegisteredRainCallbackNative4B7310(t *testing.T) {
	if name := os.Getenv(registeredRainChild4B7310); name != "" {
		testRegisteredRainChild4B7310(t, name)
		return
	}
	for _, name := range registeredRainCases4B7310 {
		t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredRainCallbackNative4B7310$", "-test.v")
			cmd.Env = append(os.Environ(), registeredRainChild4B7310+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native rain callback failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED RAIN CALLBACK PASS: "+name)) {
				t.Fatalf("missing completed spawn/lifetime/raster/upload receipt: %s", out)
			}
			t.Logf("REGISTERED RAIN CALLBACK PASS: %s", name)
		})
	}
}

func registeredRainTypes4B7310(t *testing.T, c *Client) {
	t.Helper()
	for _, def := range [][2]string{
		{"RainSourceBlue", "BlueRainDraw"}, {"RainSourceLevel", "LevelUpDraw"}, {"RainSourceOblivion", "OblivionUpDraw"},
		{"BlueRainSpark", "BlueRainSparkDraw"}, {"RainOrbWhite", "RainOrbDraw"}, {"RainOrbBlue", "RainOrbDraw"},
		{"WhiteMoveOrb", "GlowOrbMoveDraw"}, {"BlueMoveOrb", "GlowOrbMoveDraw"},
	} {
		var raw []byte
		for _, value := range []string{def[0], "DRAW", def[1]} {
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
			t.Fatalf("actual DRAW type parser failed: %v unread=%d", err, remaining)
		}
		typ := c.Things.TypeByID(def[0])
		if typ == nil || typ.DrawFunc == nil || !client.IsRegisteredDrawFunc(typ.DrawFunc) || typ.DrawData != nil {
			t.Fatal("actual kind0 DRAW registry was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("real native type must be above 4 GiB")
		}
	}
}

// Only native pointers to live graph nodes are canonicalized. All scalar,
// union, callback/type pointer and padding bytes are compared unchanged.
func registeredRainGraph4B7310(t *testing.T, c *Client) [][]byte {
	t.Helper()
	nodes := c.Objs.AllList1()
	if len(nodes) != c.Objs.Count {
		t.Fatal("real drawable list/count differs")
	}
	ids := make(map[uintptr]uint64, len(nodes))
	for i, dr := range nodes {
		ids[uintptr(dr.C())] = uint64(i + 1)
	}
	var out [][]byte
	for _, dr := range nodes {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(dr.C()) <= 0xffffffff {
			t.Fatal("spawned native drawable must be above 4 GiB")
		}
		raw := registeredEffectBytes4B7540(dr)
		typ := reflect.TypeOf(*dr)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Type.Kind() != reflect.Pointer && f.Type.Kind() != reflect.UnsafePointer {
				continue
			}
			var value uint64
			if f.Type.Size() == 8 {
				value = binary.NativeEndian.Uint64(raw[f.Offset:])
			} else {
				value = uint64(binary.NativeEndian.Uint32(raw[f.Offset:]))
			}
			if id, ok := ids[uintptr(value)]; ok {
				if f.Type.Size() == 8 {
					binary.NativeEndian.PutUint64(raw[f.Offset:], id)
				} else {
					binary.NativeEndian.PutUint32(raw[f.Offset:], uint32(id))
				}
			}
		}
		out = append(out, raw)
	}
	return out
}

func testRegisteredRainChild4B7310(t *testing.T, name string) {
	if !slices.Contains(registeredRainCases4B7310, name) {
		t.Fatal("unknown frozen rain case")
	}
	kind, style, _ := strings.Cut(name, "/")
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	frame := uint32(100)
	if style == "wrap" {
		frame = 0xffffffff
	}
	s.SetFrame(frame)
	s.Rand.Logic.Reset(93)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(16)
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
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	if style == "disabled" {
		noxflags.SetGame(noxflags.GameFlag22)
	}
	flags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	registeredRainTypes4B7310(t, c)
	beforeTypes := make(map[*client.ObjectType][]byte)
	c.Things.Each(func(p *client.ObjectType) { beforeTypes[p] = registeredEffectBytes4B7540(p) })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	type observation struct {
		Result, Other, Count int
		Initial, Final       [][]byte
		Draw                 []byte
		Pixels               []uint16
	}
	run := func(direct bool) observation {
		s.Rand.Other.Reset(17)
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		typeName := map[string]string{"blue": "RainSourceBlue", "level": "RainSourceLevel", "oblivion": "RainSourceOblivion", "white": "RainOrbWhite", "orb-blue": "RainOrbBlue"}[kind]
		pos := image.Pt(164, 280)
		if style == "below" {
			pos.Y = 180
		}
		source := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID(typeName), pos)
		wantCallback := map[string]unsafe.Pointer{
			"blue": legacy.Get_nox_thing_blue_rain_draw(), "level": legacy.Get_nox_thing_levelup_draw(),
			"oblivion": legacy.Get_nox_thing_oblivion_up_draw(), "white": legacy.Get_nox_thing_rain_orb_draw(), "orb-blue": legacy.Get_nox_thing_rain_orb_draw(),
		}[kind]
		if source == nil || source.DrawFuncPtr != wantCallback || !client.IsRegisteredDrawFunc(source.DrawFuncPtr) {
			t.Fatal("real source allocation/type/registration failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(source.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native source/viewport must be above 4 GiB")
		}
		spawner := kind == "blue" || kind == "level" || kind == "oblivion"
		if !spawner {
			source.ZVal, source.VelZ = 20, -8
			source.UnionEffect().Field_108, source.UnionEffect().Field_109 = 150, 260
			source.UnionEffect().Field_110 = 30 | 6<<16
			switch style {
			case "impact":
				source.ZVal = 0
			case "negative":
				source.ZVal = 0xfffe
			case "outside":
				source.PosVec.X = 400
			case "tiny":
				source.ZVal = 1
			}
		}
		var result int
		if direct {
			result = source.CallDraw(vp)
		} else {
			var handled bool
			result, handled = c.callRainDrawableDraw4B7310(source, vp)
			if !handled {
				t.Fatal("existing world route rejected supported rain")
			}
		}
		count, steps := 1, 0
		wantResult := 1
		if spawner && style != "disabled" {
			count, steps = 3, 8
			if kind == "blue" {
				steps = 6
			}
		} else if style == "impact" || style == "negative" {
			steps, wantResult = 2, 0
		}
		if result != wantResult || cl.Objs.Count != count || s.Rand.Other.Index() != 17+steps {
			t.Fatalf("actual callback spawn/lifetime/RNG differs: result=%d count=%d Other=%d want=%d/%d/%d", result, cl.Objs.Count, s.Rand.Other.Index(), wantResult, count, 17+steps)
		}
		if result == 0 && slices.Contains(cl.Objs.AllList1(), source) {
			t.Fatal("naturally landed source remains in real live list")
		}
		got := observation{Result: result, Count: cl.Objs.Count, Initial: registeredRainGraph4B7310(t, c)}
		// Render the real emitted particles via their existing native world
		// callbacks. Snapshot before any landing can free a spawned record.
		if spawner {
			for _, dr := range cl.Objs.AllList1() {
				if dr != source {
					c.callDrawableDraw4B6B80(dr, vp)
				}
			}
		}
		got.Final, got.Other = registeredRainGraph4B7310(t, c), s.Rand.Other.Index()
		got.Draw, got.Pixels = registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.Rand.Logic.Index() != 93 || noxflags.GetGame() != flags {
			t.Fatal("actual callback changed viewport/frame/Logic-RNG/game flags")
		}
		for typ, raw := range beforeTypes {
			if !bytes.Equal(raw, registeredEffectBytes4B7540(typ)) {
				t.Fatal("actual callback mutated shared type metadata")
			}
		}
		for cl.Objs.List1 != nil {
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(cl.Objs.List1)
		}
		if cl.Objs.Count != 0 || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("real native pool/list/lifetime cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("direct registered rain differs from world route: return/count/Other=%d/%d/%d want=%d/%d/%d initial=%t final=%t draw=%t pixels=%t", got.Result, got.Count, got.Other, want.Result, want.Count, want.Other, reflect.DeepEqual(got.Initial, want.Initial), reflect.DeepEqual(got.Final, want.Final), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED RAIN CALLBACK PASS: %s active=%d Other=%d actual-C-entry/high-native-types-and-graph/real-spawn-landing/full-raster/headless-upload", name, got.Count, got.Other))
}
