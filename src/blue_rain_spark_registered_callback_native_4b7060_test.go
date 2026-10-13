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

const registeredBlueRainSparkChild4B7060 = "OPENNOX_TEST_REGISTERED_BLUE_RAIN_SPARK_CALLBACK_R165"

var registeredBlueRainSparkCases4B7060 = []string{"slow", "convert", "odd-below", "odd-convert", "negative-velocity", "bounce-convert", "bounce-rest", "outside", "left-edge", "frame-wrap", "last-positive", "expired", "zero-life", "pool-full"}

// The actual registered C entry must match the existing native world route
// through real movement, bounce, replacement allocation/activation and deletion.
// Types are a declared native service fixture, not a Windows pixel oracle.
// Child isolation retains old C pointer faults without killing the runner.
func TestRegisteredBlueRainSparkCallbackNative4B7060(t *testing.T) {
	if name := os.Getenv(registeredBlueRainSparkChild4B7060); name != "" {
		testRegisteredBlueRainSparkChild4B7060(t, name)
		return
	}
	for _, name := range registeredBlueRainSparkCases4B7060 {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredBlueRainSparkCallbackNative4B7060$", "-test.v")
			cmd.Env = append(os.Environ(), registeredBlueRainSparkChild4B7060+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native blue-rain spark callback failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED BLUE RAIN SPARK CALLBACK PASS: "+name)) {
				t.Fatalf("missing completed movement/spawn/RNG/raster/upload receipt: %s", out)
			}
			t.Logf("REGISTERED BLUE RAIN SPARK CALLBACK PASS: %s", name)
		})
	}
}

func registeredBlueRainSparkTypes4B7060(t *testing.T, c *Client) {
	t.Helper()
	for _, row := range []struct {
		name, draw string
		callback   unsafe.Pointer
	}{
		{"BlueRainSpark", "BlueRainSparkDraw", legacy.Get_nox_thing_blue_rain_spark_draw()},
		{"WhiteSpark", "WhiteSparkDraw", legacy.Get_nox_thing_white_spark_draw()},
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
			t.Fatal("actual kind0 spark callback was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
	}
}

func testRegisteredBlueRainSparkChild4B7060(t *testing.T, name string) {
	if !slices.Contains(registeredBlueRainSparkCases4B7060, name) {
		t.Fatal("unknown frozen blue-rain spark case")
	}
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	frame := uint32(100)
	if name == "odd-below" || name == "odd-convert" {
		frame = 101
	} else if name == "frame-wrap" {
		frame = 0xffffffff
	}
	s.SetFrame(frame)
	s.Rand.Logic.Reset(93)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	capacity := 4
	if name == "pool-full" {
		capacity = 1
	}
	cl.Objs.Init(capacity)
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
	registeredBlueRainSparkTypes4B7060(t, c)
	beforeTypes := make(map[*client.ObjectType][]byte)
	c.Things.Each(func(typ *client.ObjectType) { beforeTypes[typ] = registeredEffectBytes4B7540(typ) })
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	converted := name != "slow" && name != "odd-below" && name != "bounce-rest" && name != "expired" && name != "zero-life"
	replacement := converted && name != "pool-full"
	type observation struct {
		Result, Count, Other int
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
		d.SetAlpha(64)
		d.SetAlphaEnabled(false)
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		pos := image.Pt(164, 280)
		if name == "outside" {
			pos.X = 400
		} else if name == "left-edge" {
			pos.X = 110
		}
		source := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("BlueRainSpark"), pos)
		if source == nil {
			t.Fatal("actual native source allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(source.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native records must be above 4 GiB")
		}
		source.ZVal, source.ZVal2, source.VelZ, source.Field_74_4 = 20, 3, 5, 0
		effect := source.UnionEffect()
		effect.Field_108, effect.Field_109, effect.Field_110 = uint32(pos.X)<<12, uint32(pos.Y)<<12, 256
		effect.Field_111, effect.Field_112 = frame-10, frame+20
		switch name {
		case "slow":
			source.VelZ = 4
		case "odd-convert", "frame-wrap":
			source.VelZ = 6
		case "negative-velocity":
			source.VelZ = -1
		case "bounce-convert":
			source.ZVal, source.VelZ = 1, -8
		case "bounce-rest":
			source.ZVal, source.VelZ = 0, -1
		case "last-positive":
			effect.Field_111, effect.Field_112 = frame-19, frame+1
		case "expired":
			effect.Field_111, effect.Field_112 = frame-20, frame
		case "zero-life":
			effect.Field_111, effect.Field_112 = frame, frame
		}
		var result int
		if direct {
			result = source.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(source, vp)
		}
		wantResult, wantCount, steps := 1, 1, 0
		if name == "expired" || name == "zero-life" || name == "pool-full" {
			wantResult, wantCount = 0, 0
		} else if replacement {
			wantResult, steps = 0, 5
		}
		if result != wantResult || cl.Objs.Count != wantCount || s.Rand.Other.Index() != 17+steps {
			t.Fatalf("real return/spawn/lifetime/RNG differs: result=%d count=%d Other=%d want=%d/%d/%d", result, cl.Objs.Count, s.Rand.Other.Index(), wantResult, wantCount, 17+steps)
		}
		if result == 0 && slices.Contains(cl.Objs.AllList1(), source) {
			t.Fatal("naturally deleted source remains in actual live list")
		}
		if replacement {
			spark := cl.Objs.List1
			if spark == nil || int(spark.TypeIDVal) != c.Things.IndByID("WhiteSpark") || spark.DrawFuncPtr != legacy.Get_nox_thing_white_spark_draw() || cl.Objs.List4 != spark || spark.ObjFlags&0x400000 == 0 || spark.PosVec != pos.Add(image.Pt(1, 0)) {
				t.Fatal("replacement type/callback/activation/native movement differs")
			}
			data := spark.UnionEffect()
			if data.Field_108 != uint32(spark.PosVec.X)<<12 || data.Field_109 != uint32(spark.PosVec.Y)<<12 || data.Field_111 != frame || int32(data.Field_112-frame) < 10 || int32(data.Field_112-frame) > 96 || data.Field_110 < 1 || data.Field_110 > 1611 || spark.ZVal < 5 || spark.ZVal > 15 || spark.ZVal2 != 0 || spark.VelZ < 0 || spark.VelZ > 8 {
				t.Fatal("real replacement native union/altitude initialization differs")
			}
		} else if wantCount != 0 && (cl.Objs.List1 != source || source.PosVec != pos.Add(image.Pt(1, 0)) || cl.Objs.List4 != nil) {
			t.Fatal("unconverted source movement/list changed")
		}
		got := observation{Result: result, Count: cl.Objs.Count, Initial: registeredRainGraph4B7310(t, c)}
		// Render the actual replacement through the existing native world route
		// and snapshot again after its natural movement/altitude update.
		if replacement {
			if c.callDrawableDraw4B6B80(cl.Objs.List1, vp) != 1 || cl.Objs.Count != 1 || s.Rand.Other.Index() != 22 {
				t.Fatal("real emitted WhiteSpark did not render without extra RNG")
			}
		}
		got.Final, got.Other = registeredRainGraph4B7310(t, c), s.Rand.Other.Index()
		got.Draw, got.Pixels = registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		wantPixels := name != "outside" && name != "expired" && name != "zero-life"
		changed := slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 })
		if changed != wantPixels {
			t.Fatalf("natural spark raster visibility differs: %t want %t", changed, wantPixels)
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
		t.Fatalf("registered blue-rain spark differs from native world route: result/count/Other=%d/%d/%d want=%d/%d/%d initial=%t final=%t draw=%t pixels=%t", got.Result, got.Count, got.Other, want.Result, want.Count, want.Other, reflect.DeepEqual(got.Initial, want.Initial), reflect.DeepEqual(got.Final, want.Final), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED BLUE RAIN SPARK CALLBACK PASS: %s actual-C-entry/high-native-types-and-records/movement-bounce/replacement-activation-natural-deletion/RNG/full-raster/headless-upload", name))
}
