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
	"github.com/opennox/libs/noxnet/netmsg"
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

const registeredShieldChild4B8020 = "OPENNOX_TEST_REGISTERED_SPHERICAL_SHIELD_R168"

var registeredShieldCases4B8020 = func() []string {
	var out []string
	for _, name := range sphericalShieldNames48EA70 {
		if name == "" {
			continue
		}
		for _, kind := range []string{"dynamic", "static"} {
			out = append(out, name+"/"+kind+"/visible")
		}
	}
	for _, kind := range []string{"dynamic", "static"} {
		for _, style := range []string{"moved", "edge-right", "clip-corner", "outside", "negative-z", "frame-wrap", "animation-end", "nil-data", "missing-target", "duplicate"} {
			out = append(out, "SphericalShieldNW/"+kind+"/"+style)
		}
	}
	for _, style := range []string{"center", "disconnected", "invalid-direction", "short-packet"} {
		out = append(out, "SphericalShieldNW/dynamic/"+style)
	}
	return out
}()

// The real packet consumer, DRAW parser, high-address target lookup and C
// AnimateDraw image callback must agree with the native world route. The image
// bank/type registry are explicit services, not a Windows raster/transport
// oracle. Child processes retain any pre-fix native pointer faults.
func TestRegisteredSphericalShieldCallbackNative4B8020(t *testing.T) {
	if name := os.Getenv(registeredShieldChild4B8020); name != "" {
		testRegisteredShieldChild4B8020(t, name)
		return
	}
	for _, name := range registeredShieldCases4B8020 {
		t.Run(strings.ReplaceAll(name, "/", "-"), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredSphericalShieldCallbackNative4B8020$", "-test.v")
			cmd.Env = append(os.Environ(), registeredShieldChild4B8020+"="+name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native spherical shield failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED SPHERICAL SHIELD CALLBACK PASS: "+name)) {
				t.Fatalf("missing completed packet/target/image/raster/upload receipt: %s", out)
			}
			t.Logf("REGISTERED SPHERICAL SHIELD CALLBACK PASS: %s", name)
		})
	}
}

func registeredShieldTypes4B8020(t *testing.T, c *Client, ended bool) {
	t.Helper()
	names := append([]string{"ShieldTarget", "ShieldImagePrimer"}, sphericalShieldNames48EA70[:]...)
	for _, name := range names {
		if name == "" {
			continue
		}
		drawName, callback := "SphericalShieldDraw", legacy.Get_nox_thing_spherical_shield_draw()
		if name == "ShieldTarget" {
			drawName, callback = "NoDraw", nil
		} else if name == "ShieldImagePrimer" {
			drawName, callback = "AnimateDraw", legacy.Get_nox_thing_animate_draw()
		}
		var raw []byte
		for _, value := range []string{name, "DRAW", drawName} {
			raw = append(raw, byte(len(value)))
			raw = append(raw, value...)
		}
		for len(raw)%8 != 0 {
			raw = append(raw, 0)
		}
		raw = append(raw, make([]byte, 8)...)
		animationKind, kindName := uint32(2), "Loop"
		if ended && name != "ShieldImagePrimer" {
			animationKind, kindName = 1, "OneShotRemove"
		}
		if name != "ShieldTarget" {
			raw = append(raw, 3, 1, byte(len(kindName)))
			raw = append(raw, kindName...)
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
		typ := c.Things.TypeByID(name)
		if typ == nil || typ.DrawFunc != callback || !client.IsRegisteredDrawFunc(callback) {
			t.Fatal("actual registered shield/target/image callback not installed")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(typ.C()) <= 0xffffffff {
			t.Fatal("actual native type must be above 4 GiB")
		}
		if name == "ShieldTarget" {
			if typ.DrawData != nil {
				t.Fatal("NoDraw target must not have image data")
			}
			continue
		}
		if typ.DrawData == nil || client.DrawableDataKind(callback) != 3 {
			t.Fatal("actual AnimateDraw parser lost image data")
		}
		data := (*summonAnimateDrawData4B7D00)(typ.DrawData)
		if data.size != uint32(unsafe.Sizeof(*data)) || data.frameCount != 3 || data.frameDelay != 1 || data.animationKind != animationKind || data.images == nil {
			t.Fatal("actual native AnimateDraw data layout differs")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(typ.DrawData) <= 0xffffffff || uintptr(data.images) <= 0xffffffff) {
			t.Fatal("actual C parser/image array must be above 4 GiB")
		}
		for i, handle := range unsafe.Slice((*noxrender.ImageHandle)(data.images), int(data.frameCount)) {
			if handle != c.r.GetBag().ImageByIndex(i).C() {
				t.Fatal("actual image handle identity differs")
			}
		}
		t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, 3) })
	}
}

func testRegisteredShieldChild4B8020(t *testing.T, name string) {
	if !slices.Contains(registeredShieldCases4B8020, name) {
		t.Fatal("unknown frozen spherical shield case")
	}
	parts := strings.Split(name, "/")
	typeName, kind, style := parts[0], parts[1], parts[2]
	noSpawn := slices.Contains([]string{"center", "disconnected", "invalid-direction", "short-packet"}, style)
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
	cl.Objs.Init(8)
	t.Cleanup(cl.Objs.Free)
	cl.Things.Nox_things_free_44C580_B()
	t.Cleanup(cl.Things.Nox_things_free_44C580_B)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	d, freeData := noxrender.NewRenderData()
	t.Cleanup(freeData)
	r.SetData(d)
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
	t.Cleanup(func() { *connected = oldConnected })
	fade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(fade) })
	registeredShieldTypes4B8020(t, c, style == "animation-end")
	beforeTypes, beforeData, beforeImages := make(map[*client.ObjectType][]byte), make(map[*client.ObjectType][]byte), make(map[*client.ObjectType][]noxrender.ImageHandle)
	c.Things.Each(func(typ *client.ObjectType) {
		beforeTypes[typ] = registeredEffectBytes4B7540(typ)
		if typ.DrawData != nil {
			data := (*summonAnimateDrawData4B7D00)(typ.DrawData)
			beforeData[typ] = registeredEffectBytes4B7540(data)
			beforeImages[typ] = slices.Clone(unsafe.Slice((*noxrender.ImageHandle)(data.images), int(data.frameCount)))
		}
	})
	vp, freeVP := alloc.New(noxrender.Viewport{})
	t.Cleanup(freeVP)
	vp.World, vp.Screen = image.Rect(100, 200, 228, 328), image.Rect(8, 9, 136, 137)
	vp.Size = pix.Size()
	beforeVP := registeredEffectBytes4B7540(vp)

	type observation struct {
		Consumed, Returned, Count int
		ProducerGraph, FinalGraph [][]byte
		Draw                      []byte
		Pixels                    []uint16
	}
	run := func(direct bool) observation {
		// Prime the real C previous-drawable/height cache identically before
		// either route. This is a real clipped image draw, not a renderer hook.
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		primer := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("ShieldImagePrimer"), image.Pt(400, 280))
		if primer == nil || legacy.Nox_thing_animate_draw(vp, primer) != 1 {
			t.Fatal("actual image cache primer failed")
		}
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(primer)
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		*d = noxrender.RenderData{}
		d.SetAlpha(64)
		d.SetAlphaEnabled(false)
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		*connected = 1
		if style == "disconnected" {
			*connected = 0
		}
		pos := image.Pt(164, 280)
		switch style {
		case "clip-corner":
			pos = vp.World.Min.Add(image.Pt(2, 0))
		case "edge-right":
			pos.X = 227
		case "outside":
			pos.X = 400
		}
		target := c.Nox_xxx_spriteLoadAdd_45A360_drawable(c.Things.IndByID("ShieldTarget"), pos)
		if target == nil {
			t.Fatal("actual native target allocation failed")
		}
		code := uint16(0x1234)
		target.NetCode32 = uint32(code)
		if kind == "static" {
			target.ObjClass = 0x20400000
			code |= 0x8000
		}
		if cl.Objs.ByNetCode(code) != target {
			t.Fatal("actual native static/dynamic target lookup failed")
		}
		direction := slices.Index(sphericalShieldNames48EA70[:], typeName)
		if style == "center" {
			direction = 4
		} else if style == "invalid-direction" {
			direction = 9
		}
		packet := []byte{byte(netmsg.MSG_FX_SHIELD), byte(code), byte(code >> 8), byte(direction)}
		if style == "short-packet" {
			packet = packet[:3]
		}
		beforePacket := slices.Clone(packet)
		consumed := c.handleShieldFXPacketNative48EA70(packet)
		if !bytes.Equal(packet, beforePacket) || consumed != 4 && style != "short-packet" || style == "short-packet" && consumed != -1 {
			t.Fatal("actual packet consumer mutated bytes/width")
		}
		var shield *client.Drawable
		if noSpawn {
			if cl.Objs.Count != 1 || cl.Objs.List1 != target {
				t.Fatal("actual packet guard spawned an effect")
			}
		} else {
			shield = cl.Objs.List1
			if shield == nil || shield == target || cl.Objs.Count != 2 || shield.NextPtr != target || shield.TypeIDVal != uint32(c.Things.IndByID(typeName)) || shield.DrawFuncPtr != legacy.Get_nox_thing_spherical_shield_draw() || shield.DrawData != c.Things.TypeByID(typeName).DrawData || shield.PosVec != pos.Add(image.Pt(0, 3)) || shield.UnionEffect().Field_108 != uint32(code) {
				t.Fatal("actual 8-direction shield packet/type/native endpoint differs")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(shield.C()) <= 0xffffffff || uintptr(target.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
				t.Fatal("actual native records must be above 4 GiB")
			}
			if style == "duplicate" && (c.handleShieldFXPacketNative48EA70(packet) != 4 || cl.Objs.Count != 2 || cl.Objs.List1 != shield || !bytes.Equal(packet, beforePacket)) {
				t.Fatal("actual duplicate packet did not preserve effect identity")
			}
			c.Objs.TransparentDecay(shield, 50)
			if style == "moved" {
				c.Nox_xxx_updateSpritePosition_49AA90(target, 218, 220)
			}
			if style == "negative-z" {
				shield.ZVal, shield.ZVal2 = 0xfffe, 0xfffd
			}
			if style == "animation-end" {
				shield.AnimStart = frame - 10
			}
			if style == "nil-data" {
				shield.DrawData = nil
			}
			if style == "missing-target" {
				c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(target)
				target = nil
			}
		}
		got := observation{Consumed: consumed, ProducerGraph: registeredRainGraph4B7310(t, c)}
		if shield != nil {
			if direct {
				got.Returned = shield.CallDraw(vp)
			} else {
				got.Returned = c.callDrawableDraw4B6B80(shield, vp)
			}
			removed := style == "missing-target" || style == "animation-end"
			if removed {
				wantCount := 1
				if target == nil {
					wantCount = 0
				}
				if got.Returned != 0 || cl.Objs.Count != wantCount || cl.Objs.List1 != target || cl.Objs.DeadlineList != nil {
					t.Fatalf("actual target/animation-end removal differs: direct=%t", direct)
				}
			} else if got.Returned != 1 || cl.Objs.Count != 2 || cl.Objs.List1 != shield || shield.PosVec != target.PosVec.Add(image.Pt(0, 3)) || shield.Deadline != frame+50 || cl.Objs.DeadlineList != shield || shield.UnionEffect().Field_108 != uint32(code) || shield.DrawFuncPtr != legacy.Get_nox_thing_spherical_shield_draw() {
				t.Fatalf("actual native target position/image callback/lifetime differs: direct=%t", direct)
			}
		}
		got.Count, got.FinalGraph, got.Draw, got.Pixels = cl.Objs.Count, registeredRainGraph4B7310(t, c), registeredEffectBytes4B7540(d), slices.Clone(pix.Pix)
		wantPixels := !noSpawn && !slices.Contains([]string{"outside", "negative-z", "animation-end", "nil-data", "missing-target"}, style)
		if slices.ContainsFunc(got.Pixels, func(v uint16) bool { return v != 0x2104 }) != wantPixels {
			t.Fatalf("actual AnimateDraw image visibility differs: direct=%t", direct)
		}
		if !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != frame || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 || s.Rand.Other.Index() != 17 {
			t.Fatal("callback mutated viewport/frame/tick-rate/Logic/Other RNG")
		}
		for typ, raw := range beforeTypes {
			if !bytes.Equal(raw, registeredEffectBytes4B7540(typ)) {
				t.Fatal("callback mutated shared type metadata")
			}
			if typ.DrawData != nil {
				data := (*summonAnimateDrawData4B7D00)(typ.DrawData)
				if !bytes.Equal(beforeData[typ], registeredEffectBytes4B7540(data)) || !slices.Equal(beforeImages[typ], unsafe.Slice((*noxrender.ImageHandle)(data.images), int(data.frameCount))) {
					t.Fatal("callback mutated shared native image data/handles")
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
		t.Fatalf("registered shield differs from native world route: producer=%t graph=%t draw=%t pixels=%t", reflect.DeepEqual(got.ProducerGraph, want.ProducerGraph), reflect.DeepEqual(got.FinalGraph, want.FinalGraph), bytes.Equal(got.Draw, want.Draw), slices.Equal(got.Pixels, want.Pixels))
	}
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	presentParticleRegistration4B01B0(t, pix, output, sc, want.Pixels, 0)
	t.Log(fmt.Sprintf("REGISTERED SPHERICAL SHIELD CALLBACK PASS: %s actual-packet-parser-C-entry/high-native-types-image-arrays-and-records/static-dynamic-target-lookup-position-or-natural-delete/real-AnimateDraw-image/full-graph-render-state-raster/headless-upload", name))
}
