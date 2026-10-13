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

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const registeredArrowChild4B7920 = "OPENNOX_TEST_REGISTERED_ARROW_CALLBACK_R167"

var registeredArrowStyles4B7920 = []string{"short", "threshold", "long", "reverse", "negative-prev", "negative-z", "clip-corner", "edge-right", "outside", "frame-wrap", "nil-data", "invalid-frame", "pool-full"}
var registeredArrowTailStyles4B7920 = []string{"horizontal", "vertical", "reverse", "zero-length", "negative-prev", "negative-z", "clip-corner", "edge-right", "outside", "frame-wrap", "half-life", "last-positive", "expired-deadline", "clamp-life"}

// Exercise all four real registered C entries, including the actual SlaveDraw
// image callback and spawned tails. The colored image bank, palette and type
// registry are explicit native service fixtures, not a Windows pixel oracle.
// Separate processes retain native pointer faults without crashing the suite.
func TestRegisteredArrowCallbackNative4B7920(t *testing.T) {
	if name := os.Getenv(registeredArrowChild4B7920); name != "" {
		testRegisteredArrowChild4B7920(t, name)
		return
	}
	for _, kind := range []string{"arrow", "weak-arrow", "tail", "weak-tail"} {
		styles := registeredArrowStyles4B7920
		if strings.HasSuffix(kind, "tail") {
			styles = registeredArrowTailStyles4B7920
		}
		for _, style := range styles {
			name := kind + "/" + style
			t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredArrowCallbackNative4B7920$", "-test.v")
				cmd.Env = append(os.Environ(), registeredArrowChild4B7920+"="+name)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("registered native arrow callback failed: %v\n%s", err, out)
				}
				if !bytes.Contains(out, []byte("REGISTERED ARROW CALLBACK PASS: "+name)) {
					t.Fatalf("missing completed image/tail/raster/upload receipt: %s", out)
				}
				t.Logf("REGISTERED ARROW CALLBACK PASS: %s", name)
			})
		}
	}
}

func registeredArrowTypes4B7920(t *testing.T, c *Client) {
	t.Helper()
	for _, row := range []struct {
		name, draw string
		callback   unsafe.Pointer
		images     bool
	}{
		{"Arrow", "ArrowDraw", legacy.Get_nox_thing_arrow_draw(), true},
		{"WeakArrow", "WeakArrowDraw", legacy.Get_nox_thing_weak_arrow_draw(), true},
		{"ArrowTailLink", "ArrowTailLinkDraw", legacy.Get_nox_thing_arrow_tail_link_draw(), false},
		{"WeakArrowTailLink", "WeakArrowTailLinkDraw", legacy.Get_nox_thing_weak_arrow_tail_link_draw(), false},
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
		if row.images {
			raw = append(raw, 3)
			for _, id := range []uint32{0, 1, 2} {
				raw = binary.LittleEndian.AppendUint32(raw, id)
			}
		}
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
		if typ == nil || typ.DrawFunc != row.callback || !client.IsRegisteredDrawFunc(row.callback) {
			t.Fatal("actual registered arrow callback was not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
		if !row.images {
			if typ.DrawData != nil || client.DrawableDataKind(typ.DrawFunc) != 0 {
				t.Fatal("actual tail parser produced unexpected draw data")
			}
			continue
		}
		if typ.DrawData == nil || client.DrawableDataKind(typ.DrawFunc) != 2 {
			t.Fatal("actual SlaveDraw parser lost image data")
		}
		data := (*nativeStaticRandomDrawData)(typ.DrawData)
		if data.size != uint32(unsafe.Sizeof(*data)) || data.count != 3 || data.images == nil {
			t.Fatal("actual native SlaveDraw data layout differs")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(typ.DrawData) <= 0xffffffff || uintptr(unsafe.Pointer(data.images)) <= 0xffffffff) {
			t.Fatal("actual C parser/image array must be above 4 GiB")
		}
		for i, handle := range unsafe.Slice(data.images, int(data.count)) {
			if handle != c.r.GetBag().ImageByIndex(i).C() {
				t.Fatal("actual parser lost native image-handle identity")
			}
		}
		t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, 2) })
	}
}

func testRegisteredArrowChild4B7920(t *testing.T, name string) {
	kind, style, ok := strings.Cut(name, "/")
	if !ok || !slices.Contains([]string{"arrow", "weak-arrow", "tail", "weak-tail"}, kind) {
		t.Fatal("unknown frozen arrow callback")
	}
	isTail := strings.HasSuffix(kind, "tail")
	styles := registeredArrowStyles4B7920
	if isTail {
		styles = registeredArrowTailStyles4B7920
	}
	if !slices.Contains(styles, style) {
		t.Fatal("unknown frozen arrow condition")
	}
	t.Chdir(t.TempDir())
	destroyOnceBag48EA70(t)
	handles.Init()
	t.Cleanup(handles.Release)
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
	poolSize := 4
	if style == "pool-full" {
		poolSize = 1
	}
	cl.Objs.Init(poolSize)
	t.Cleanup(cl.Objs.Free)
	cl.Things.Nox_things_free_44C580_B()
	t.Cleanup(cl.Things.Nox_things_free_44C580_B)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
	// Real image clipping saves/restores this C renderer pointer, independently
	// of the Go Render.SetData slot. Initialize both to the same C-owned data.
	oldRenderDataC := *legacy.VideoModeRenderDataC
	legacy.Set_nox_draw_curDrawData_3799572(d)
	t.Cleanup(func() { *legacy.VideoModeRenderDataC = oldRenderDataC })
	pix := noximage.NewImage16(image.Rect(0, 0, 144, 144))
	r.SetPixBuffer(pix)
	if err := r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.GetBag().Free)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	// Supply the same real ambient-light context as the image callback tests.
	c.tiles.lightsOutBuf, _ = alloc.Make([]uint32{}, 6)
	t.Cleanup(func() { alloc.FreeSlice(c.tiles.lightsOutBuf) })
	for i := 0; i < 3; i++ {
		c.tiles.lightsOutBuf[i] = 255
	}
	for x := range c.tiles.nox_arr2_853BC0 {
		for y := range c.tiles.nox_arr2_853BC0[x] {
			c.tiles.nox_arr2_853BC0[x][y] = noxrender.RGB{R: 255 << 16, G: 255 << 16, B: 255 << 16}
		}
	}
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = c, c.srv
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	connected := memmap.PtrUint32(0x5D4594, 815764)
	oldConnected := *connected
	*connected = 0
	t.Cleanup(func() { *connected = oldConnected })
	fade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(fade) })
	for n, base := range []uintptr{1313012, 1313268} {
		table := unsafe.Slice(memmap.PtrUint32(0x5D4594, base), 64)
		before := slices.Clone(table)
		t.Cleanup(func() { copy(table, before) })
		for i := range table {
			table[i] = uint32(noxcolor.RGB5551Color(byte(40+i*3), byte(240-i*3), byte(64+i*2+n*32)))
		}
	}
	registeredArrowTypes4B7920(t, c)
	beforeTypes, beforeData, beforeImages := make(map[*client.ObjectType][]byte), make(map[*client.ObjectType][]byte), make(map[*client.ObjectType][]noxrender.ImageHandle)
	c.Things.Each(func(typ *client.ObjectType) {
		beforeTypes[typ] = registeredEffectBytes4B7540(typ)
		if typ.DrawData != nil {
			data := (*nativeStaticRandomDrawData)(typ.DrawData)
			beforeData[typ] = registeredEffectBytes4B7540(data)
			beforeImages[typ] = slices.Clone(unsafe.Slice(data.images, int(data.count)))
		}
	})
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	type observation struct {
		Result, Count       int
		Graph, FinalGraph   [][]byte
		Draw, FinalDraw     []byte
		Pixels, FinalPixels []uint16
	}
	run := func(direct bool) observation {
		// The real C image renderer retains previous_drawable/previous_z even
		// after pool deletion. Prime that cache with the same real zero-Z,
		// clipped image draw before each independent route comparison.
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		primer := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("Arrow"), image.Pt(400, 280))
		if primer == nil {
			t.Fatal("actual renderer-cache primer allocation failed")
		}
		if legacy.Nox_thing_slave_draw(vp, primer) != 1 {
			t.Fatal("actual renderer-cache primer image draw failed")
		}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(primer)
		if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("actual renderer-cache primer cleanup incomplete")
		}
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
			pos = vp.World.Min.Add(image.Pt(2, 5))
		case "edge-right":
			pos.X = 227
		case "outside":
			pos.X = 400
		}
		typeName := map[string]string{"arrow": "Arrow", "weak-arrow": "WeakArrow", "tail": "ArrowTailLink", "weak-tail": "WeakArrowTailLink"}[kind]
		dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID(typeName), pos)
		if dr == nil {
			t.Fatal("actual native drawable allocation failed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual native drawable/viewport must be above 4 GiB")
		}
		dr.ZVal, dr.ZVal2 = 8, 3
		previous := pos.Sub(image.Pt(15, 8))
		switch style {
		case "short":
			previous = pos.Sub(image.Pt(14, 0))
		case "threshold":
			previous = pos.Sub(image.Pt(10, 10))
		case "horizontal":
			previous = pos.Sub(image.Pt(20, 0))
		case "vertical":
			previous = pos.Sub(image.Pt(0, 20))
		case "reverse":
			previous = pos.Add(image.Pt(20, 10))
		case "zero-length":
			previous = pos
		case "negative-prev":
			previous = image.Pt(-5, -10)
		case "negative-z":
			dr.ZVal, dr.ZVal2 = 0xfffe, 0xfffd
		case "clip-corner":
			dr.ZVal, dr.ZVal2 = 0, 0
		}
		dr.Field_81, dr.Field_82 = uint32(int32(previous.X)), uint32(int32(previous.Y))
		dr.AnimFrameSlave = 1
		dr.UnionEffect().Field_108, dr.UnionEffect().Field_109 = uint32(int32(previous.X)), uint32(int32(previous.Y))
		dr.UnionEffect().Field_111 = 0xabcddcba
		if isTail {
			lifetime := 20
			switch style {
			case "half-life":
				lifetime = 10
			case "last-positive":
				lifetime = 1
			case "expired-deadline":
				lifetime = -1
			case "clamp-life":
				lifetime = 30
			}
			c.Objs.List34Add(dr)
			c.Objs.TransparentDecay(dr, lifetime)
		} else {
			if style == "nil-data" {
				dr.DrawData = nil
			}
			if style == "invalid-frame" {
				dr.AnimFrameSlave = 3
			}
		}
		expectedDR := *dr
		spawned := !isTail && style != "short" && style != "threshold" && style != "pool-full"
		if spawned {
			expectedDR.Field_81, expectedDR.Field_82 = uint32(pos.X), uint32(pos.Y)
		}
		var result int
		if direct {
			result = dr.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(dr, vp)
		}
		var tail *client.Drawable
		if spawned {
			tail = cl.Objs.List1
			if tail == dr || tail == nil || tail.NextPtr != dr || cl.Objs.Count != 2 || cl.Objs.List4 != tail || cl.Objs.DeadlineList != tail || tail.Deadline != frame+20 || tail.PosVec != previous || tail.UnionEffect().Field_108 != uint32(pos.X) || tail.UnionEffect().Field_109 != uint32(pos.Y) {
				t.Fatalf("actual tail spawn/position/activation/decay/native endpoint differs: direct=%t", direct)
			}
			tailName := "ArrowTailLink"
			if kind == "weak-arrow" {
				tailName = "WeakArrowTailLink"
			}
			if tail.TypeIDVal != uint32(c.Things.IndByID(tailName)) || tail.DrawFuncPtr != c.Things.TypeByID(tailName).DrawFunc {
				t.Fatal("actual spawned tail type/callback differs")
			}
			// Only the real list insertion and two position-cache fields may change.
			expectedDR.Field_93 = tail
			if previous.X/client.Nox_drawable_2d_div == pos.X/client.Nox_drawable_2d_div && previous.Y/client.Nox_drawable_2d_div == pos.Y/client.Nox_drawable_2d_div {
				expectedDR.Field_101 = tail
			}
		} else if cl.Objs.Count != 1 || cl.Objs.List1 != dr {
			t.Fatalf("callback mutated native pool/list count: direct=%t", direct)
		}
		if !isTail && dr.DrawData != nil && dr.AnimFrameSlave < uint32((*nativeStaticRandomDrawData)(dr.DrawData).count) {
			// Real image rendering records its image identity and clipping cache.
			// Account for that documented 4C4770 service, not a no-op substitute.
			expectedDR.Field_0 = expectedDR.Field_0&0x0000ffff | 0x00010000
			expectedDR.Field_1 = uint32(memmap.Uint16(0x973F18, 88)) | uint32(memmap.Uint16(0x973F18, 76))<<16
			expectedDR.Field_2 = unsafe.Pointer(c.r.GetBag().ImageByIndex(int(dr.AnimFrameSlave)).C())
		}
		if result != 1 || !bytes.Equal(registeredEffectBytes4B7540(&expectedDR), registeredEffectBytes4B7540(dr)) {
			t.Fatalf("callback mutated unrelated drawable fields/callback or return: direct=%t", direct)
		}
		got := observation{Result: result, Count: cl.Objs.Count, Graph: registeredRainGraph4B7310(t, c), Draw: registeredEffectBytes4B7540(d), Pixels: slices.Clone(pix.Pix)}
		wantPixels := style != "outside" && style != "nil-data" && style != "invalid-frame" && style != "expired-deadline" && style != "zero-length"
		if !isTail && style == "negative-z" {
			// The real 2-pixel image is entirely cropped by draw_z == 2.
			wantPixels = false
		}
		changed := slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 })
		if changed != wantPixels {
			t.Fatalf("actual image/tail visibility differs: %t want %t direct=%t", changed, wantPixels, direct)
		}
		if tail != nil {
			beforeTail := registeredEffectBytes4B7540(tail)
			var returned int
			if direct {
				returned = tail.CallDraw(vp)
			} else {
				returned = c.callDrawableDraw4B6B80(tail, vp)
			}
			if returned != 1 || !bytes.Equal(beforeTail, registeredEffectBytes4B7540(tail)) {
				t.Fatal("actual spawned tail draw mutated lifetime/native record")
			}
		}
		got.FinalGraph, got.FinalDraw, got.FinalPixels = registeredRainGraph4B7310(t, c), registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 || s.Rand.Other.Index() != 17 {
			t.Fatal("callback mutated viewport/frame/tick-rate/Logic/Other RNG")
		}
		for typ, raw := range beforeTypes {
			if !bytes.Equal(raw, registeredEffectBytes4B7540(typ)) {
				t.Fatal("callback mutated shared type metadata")
			}
			if typ.DrawData != nil {
				data := (*nativeStaticRandomDrawData)(typ.DrawData)
				if !bytes.Equal(beforeData[typ], registeredEffectBytes4B7540(data)) || !slices.Equal(beforeImages[typ], unsafe.Slice(data.images, int(data.count))) {
					t.Fatal("callback mutated shared image data/handles")
				}
			}
		}
		for cl.Objs.List1 != nil {
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(cl.Objs.List1)
		}
		if cl.Objs.Count != 0 || cl.Objs.List4 != nil || cl.Objs.DeadlineList != nil {
			t.Fatal("actual pool/activation/deadline cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registered arrow differs from native world route: graph=%t image-draw=%t image-pixels=%t tail-draw=%t tail-pixels=%t", reflect.DeepEqual(got.Graph, want.Graph) && reflect.DeepEqual(got.FinalGraph, want.FinalGraph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels), bytes.Equal(got.FinalDraw, want.FinalDraw), slices.Equal(got.FinalPixels, want.FinalPixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.FinalPixels, 0)
	t.Log(fmt.Sprintf("REGISTERED ARROW CALLBACK PASS: %s actual-C-entry/high-native-types-image-arrays-and-records/real-SlaveDraw-image/tail-spawn-activation-decay/full-graph-render-state-raster/headless-upload", name))
}
