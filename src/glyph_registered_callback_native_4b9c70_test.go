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
	"testing"
	"unsafe"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const registeredGlyphChild4B9C70 = "OPENNOX_TEST_REGISTERED_GLYPH_R180"

var registeredGlyphStyles4B9C70 = []string{
	"near", "middle", "edge", "far", "forced", "infravision",
	"forced-infravision", "negative-z", "clip", "outside",
	"frame-wrap", "nil-animation", "animation-fade", "no-local", "no-client",
}

// Compare the actual registered C GlyphDraw, native C DRAW/image parser,
// local-player pointer slot, AnimateDraw/raster and headless upload with the
// existing game route. The colored bag is a declared image service, not an
// independent original Windows framebuffer or a physical SDL/OpenGL oracle.
func TestRegisteredGlyphCallbackNative4B9C70(t *testing.T) {
	if style := os.Getenv(registeredGlyphChild4B9C70); style != "" {
		testRegisteredGlyphChild4B9C70(t, style)
		return
	}
	for _, style := range registeredGlyphStyles4B9C70 {
		t.Run(style, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredGlyphCallbackNative4B9C70$", "-test.v")
			cmd.Env = append(os.Environ(), registeredGlyphChild4B9C70+"="+style)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered native glyph failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED GLYPH CALLBACK PASS: "+style)) {
				t.Fatalf("missing completed parser/local-player/raster/upload receipt: %s", out)
			}
			t.Logf("REGISTERED GLYPH CALLBACK PASS: %s", style)
		})
	}
}

func registeredGlyphType4B9C70(t *testing.T, c *Client) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, text := range []string{"Glyph", "DRAW", "GlyphDraw"} {
		raw = append(raw, byte(len(text)))
		raw = append(raw, text...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	raw = append(raw, 3, 1, 4)
	raw = append(raw, "Loop"...)
	for _, id := range []uint32{0, 1, 2} {
		raw = binary.LittleEndian.AppendUint32(raw, id)
	}
	raw = append(raw, 0)
	buffer, _ := alloc.CloneSlice(raw)
	f := binfile.NewMemFile(unsafe.Pointer(&buffer[0]), len(buffer))
	defer f.Free()
	scratch, free := alloc.Make([]byte{}, 256)
	defer free()
	if err := c.Things.ReadType(f, scratch); err != nil || len(f.Data()) != 0 {
		t.Fatalf("actual GlyphDraw parser failed: %v unread=%d", err, len(f.Data()))
	}
	typ := c.Things.TypeByID("Glyph")
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_glyph_draw() ||
		!client.IsRegisteredDrawFunc(typ.DrawFunc) || client.DrawableDataKind(typ.DrawFunc) != 3 || typ.DrawData == nil {
		t.Fatal("native parser lost registered GlyphDraw kind/data")
	}
	anim := (*summonAnimateDrawData4B7D00)(typ.DrawData)
	if anim.size != uint32(unsafe.Sizeof(*anim)) || anim.frameCount != 3 ||
		anim.frameDelay != 1 || anim.animationKind != 2 || anim.images == nil {
		t.Fatal("native GlyphDraw animation layout differs")
	}
	for i, h := range unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3) {
		if h != c.r.GetBag().ImageByIndex(i).C() {
			t.Fatal("native C parser lost image handle identity")
		}
	}
	t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, 3) })
	return typ
}

func testRegisteredGlyphChild4B9C70(t *testing.T, style string) {
	if !slices.Contains(registeredGlyphStyles4B9C70, style) {
		t.Fatal("unknown frozen glyph case")
	}
	t.Chdir(t.TempDir())
	destroyOnceBag48EA70(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	s.SetFrame(100)
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
	playerSlot := memmap.PtrPtr(0x852978, 8)
	oldPlayer := *playerSlot
	t.Cleanup(func() { *playerSlot = oldPlayer })
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	if style != "no-client" {
		noxflags.SetGame(noxflags.GameClient)
	}
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	oldGreen := legacy.Get_dword_8531A0_2572()
	legacy.Set_dword_8531A0_2572(nox_color_green.Color32())
	t.Cleanup(func() { legacy.Set_dword_8531A0_2572(oldGreen) })
	connected := memmap.PtrUint32(0x5D4594, 815764)
	oldConnected := *connected
	*connected = 1
	t.Cleanup(func() { *connected = oldConnected })
	fade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(fade) })
	typ := registeredGlyphType4B9C70(t, c)
	anim := (*summonAnimateDrawData4B7D00)(typ.DrawData)
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = image.Rect(8, 9, 136, 137), image.Rect(100, 200, 228, 328), pix.Size()
	dr := c.Nox_xxx_spriteLoadAdd_45A360_drawable(typ.Index(), image.Pt(164, 280))
	if dr == nil {
		t.Fatal("actual native glyph creation failed")
	}
	local, freeLocal := alloc.New(client.Drawable{})
	local.PosVec = dr.PosVec
	t.Cleanup(freeLocal)
	dr.NetCode32, dr.AnimStart = 1, 96
	switch style {
	case "middle":
		local.PosVec.X -= 75
	case "edge":
		local.PosVec.X -= 149
	case "far", "forced", "forced-infravision":
		local.PosVec.X -= 150
	}
	if style == "forced" || style == "forced-infravision" {
		dr.ObjFlags |= 0x40000000
	}
	if style == "infravision" || style == "forced-infravision" {
		local.Buffs |= 1 << server.ENCHANT_INFRAVISION
	}
	if style == "negative-z" {
		dr.ZVal, dr.ZVal2 = uint16(0xfffd), 2
	}
	if style == "clip" {
		dr.PosVec, local.PosVec = image.Pt(99, 201), image.Pt(99, 201)
	}
	if style == "outside" {
		dr.PosVec, local.PosVec = image.Pt(1000, 1000), image.Pt(1000, 1000)
	}
	if style == "frame-wrap" {
		dr.AnimStart = 0xfffffffc
		s.SetFrame(1)
	}
	if style == "nil-animation" {
		dr.DrawData = nil
	}
	if style == "animation-fade" {
		anim.animationKind = 3
	}
	*playerSlot = local.C()
	if style == "no-local" {
		*playerSlot = nil
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff ||
		uintptr(local.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff ||
		uintptr(typ.DrawData) <= 0xffffffff || uintptr(anim.images) <= 0xffffffff) {
		t.Fatal("native glyph/local player/viewport/animation/image array must exceed 4 GiB")
	}
	beforeDR, beforeLocal, beforeVP := registeredEffectBytes4B7540(dr), registeredEffectBytes4B7540(local), registeredEffectBytes4B7540(vp)
	beforeType, beforeAnim := registeredEffectBytes4B7540(typ), registeredEffectBytes4B7540(anim)
	beforeImages := slices.Clone(unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3))
	primer, freePrimer := alloc.New(client.Drawable{})
	t.Cleanup(freePrimer)
	*primer = *dr
	primer.PosVec = image.Pt(400, 280)
	primer.ZVal, primer.ZVal2 = 0, 0
	primer.DrawData = typ.DrawData
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(output.Close)
	type observation struct {
		Returned           int
		Draw, Drawable     []byte
		Pixels             []uint16
		Count              int
		OtherRNG, LogicRNG uint32
	}
	run := func(direct bool) observation {
		// Prime the real C previous-drawable/negative-height clip cache
		// identically for each route with an actual off-screen image draw.
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		if legacy.Nox_thing_animate_draw(vp, primer) != 1 {
			t.Fatal("actual image cache primer failed")
		}
		// Reuse the same C-owned drawable address, preserving the original
		// low-pointer-byte fallback alpha in the no-client/no-local cases.
		copy(unsafe.Slice((*byte)(dr.C()), len(beforeDR)), beforeDR)
		*d = noxrender.RenderData{}
		d.SetAlpha(33)
		d.SetAlphaEnabled(true)
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		var returned int
		if direct {
			returned = dr.CallDraw(vp)
		} else {
			returned = c.callDrawableDraw4B6B80(dr, vp)
		}
		got := observation{returned, registeredEffectBytes4B7540(d), registeredEffectBytes4B7540(dr),
			slices.Clone(pix.Pix), cl.Objs.Count, uint32(s.Rand.Other.Index()), uint32(s.Rand.Logic.Index())}
		if returned != 1 || cl.Objs.Count != 1 || !bytes.Equal(beforeLocal, registeredEffectBytes4B7540(local)) ||
			!bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) ||
			!bytes.Equal(beforeAnim, registeredEffectBytes4B7540(anim)) ||
			!slices.Equal(beforeImages, unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3)) ||
			s.Rand.Other.Index() != 17 || s.Rand.Logic.Index() != 93 {
			t.Fatal("native glyph mutated return/pool/local player/viewport/parser/images/RNG")
		}
		if style == "far" {
			if d.Alpha() != 33 || !d.IsAlphaEnabled() {
				t.Fatal("out-of-range glyph altered incoming alpha state")
			}
		} else if d.IsAlphaEnabled() {
			t.Fatalf("visible glyph did not retire alpha after AnimateDraw: direct=%t alpha=%d glyph=%v local=%v slot=%p game=%x flags=%x", direct, d.Alpha(), dr.PosVec, local.PosVec, c.ClientPlayerUnit(), noxflags.GetGame(), dr.ObjFlags)
		}
		output.CopyBuffer(pix)
		if shown, _ := sc.Snapshot(); shown == nil || !slices.Equal(shown.Pix, pix.Pix) {
			t.Fatal("actual headless upload differs from native glyph raster")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("registered glyph differs from game route: return=%d/%d render=%t drawable=%t pixels=%t",
			got.Returned, want.Returned, bytes.Equal(got.Draw, want.Draw),
			bytes.Equal(got.Drawable, want.Drawable), slices.Equal(got.Pixels, want.Pixels))
	}
	if style == "near" && slices.Equal(got.Pixels, slices.Repeat([]uint16{0x2104}, len(got.Pixels))) {
		t.Fatal("visible glyph image was not actually rasterized")
	}
	c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
	if cl.Objs.Count != 0 {
		t.Fatal("native glyph cleanup incomplete")
	}
	fmt.Println("REGISTERED GLYPH CALLBACK PASS: " + style)
}
