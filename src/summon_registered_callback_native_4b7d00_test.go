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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

const registeredSummonChild4B7D00 = "OPENNOX_TEST_REGISTERED_SUMMON_R178"

var registeredSummonStyles4B7D00 = []string{"first", "half", "last", "expired", "zero-life", "frame-wrap", "clip", "outside"}

// The registered C entry must perform the same native image rendering and
// lifecycle as the existing world dispatcher. Neither route's raster, pool,
// packet, parser or deletion services are replaced. These colored fixtures
// are a native dispatch regression, not a Windows framebuffer oracle.
func TestRegisteredSummonCallbackNative4B7D00(t *testing.T) {
	if style := os.Getenv(registeredSummonChild4B7D00); style != "" {
		testRegisteredSummonChild4B7D00(t, style)
		return
	}
	for _, style := range registeredSummonStyles4B7D00 {
		t.Run(style, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestRegisteredSummonCallbackNative4B7D00$", "-test.v")
			cmd.Env = append(os.Environ(), registeredSummonChild4B7D00+"="+style)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered summon callback failed: %v\n%s", err, out)
			}
			if !bytes.Contains(out, []byte("REGISTERED SUMMON CALLBACK PASS: "+style)) {
				t.Fatalf("missing completed native raster/lifecycle receipt: %s", out)
			}
			t.Logf("REGISTERED SUMMON CALLBACK PASS: %s", style)
		})
	}
}

func registeredSummonType4B7D00(t *testing.T, c *Client, name, draw string) *client.ObjectType {
	t.Helper()
	var raw []byte
	for _, text := range []string{name, "DRAW", draw} {
		raw = append(raw, byte(len(text)))
		raw = append(raw, text...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	switch draw {
	case "SummonEffectDraw":
		raw = append(raw, 3, 1, 4)
		raw = append(raw, "Loop"...)
		for _, id := range []uint32{0, 1, 2} {
			raw = binary.LittleEndian.AppendUint32(raw, id)
		}
	case "StaticDraw":
		raw = binary.LittleEndian.AppendUint32(raw, 2)
	}
	raw = append(raw, 0)
	buf, _ := alloc.CloneSlice(raw)
	f := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	defer f.Free()
	scratch, free := alloc.Make([]byte{}, 256)
	defer free()
	if err := c.Things.ReadType(f, scratch); err != nil || len(f.Data()) != 0 {
		t.Fatalf("actual DRAW parser failed: %v unread=%d", err, len(f.Data()))
	}
	typ := c.Things.TypeByID(name)
	if typ == nil || !client.IsRegisteredDrawFunc(typ.DrawFunc) {
		t.Fatal("actual registered draw callback missing")
	}
	if typ.DrawData != nil {
		t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, client.DrawableDataKind(typ.DrawFunc)) })
	}
	return typ
}

// Only pointers to live graph nodes are canonicalized, including the
// unlinked child. Scalar, image/type/callback and padding bytes stay exact.
func registeredSummonGraph4B7D00(t *testing.T, c *Client, parent, child *client.Drawable) [][]byte {
	t.Helper()
	nodes := append(c.Objs.AllList1(), child)
	if len(nodes) != c.Objs.Count {
		t.Fatal("summon graph/count differs")
	}
	ids := make(map[uintptr]uint64, len(nodes))
	for i, dr := range nodes {
		ids[uintptr(dr.C())] = uint64(i + 1)
	}
	var out [][]byte
	for _, dr := range nodes {
		raw := registeredEffectBytes4B7540(dr)
		canonicalize := func(at uintptr) {
			var v uint64
			if unsafe.Sizeof(uintptr(0)) == 8 {
				v = binary.NativeEndian.Uint64(raw[at:])
			} else {
				v = uint64(binary.NativeEndian.Uint32(raw[at:]))
			}
			if id, ok := ids[uintptr(v)]; ok {
				if unsafe.Sizeof(uintptr(0)) == 8 {
					binary.NativeEndian.PutUint64(raw[at:], id)
				} else {
					binary.NativeEndian.PutUint32(raw[at:], uint32(id))
				}
			}
		}
		typ := reflect.TypeOf(*dr)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Type.Kind() == reflect.Pointer || f.Type.Kind() == reflect.UnsafePointer {
				canonicalize(f.Offset)
			}
		}
		if dr == parent {
			canonicalize(unsafe.Offsetof(dr.Union))
		}
		out = append(out, raw)
	}
	return out
}

func testRegisteredSummonChild4B7D00(t *testing.T, style string) {
	if !slices.Contains(registeredSummonStyles4B7D00, style) {
		t.Fatal("unknown summon fixture")
	}
	t.Chdir(t.TempDir())
	destroyOnceBag48EA70(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetTickRate(60)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(128)
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
	*connected = 1
	t.Cleanup(func() { *connected = oldConnected })
	fade := legacy.Get_nox_client_fadeObjects_80836()
	legacy.Set_nox_client_fadeObjects_80836(0)
	t.Cleanup(func() { legacy.Set_nox_client_fadeObjects_80836(fade) })
	parentType := registeredSummonType4B7D00(t, c, "SummonEffect", "SummonEffectDraw")
	childType := registeredSummonType4B7D00(t, c, "SummonChild", "StaticDraw")
	registeredSummonType4B7D00(t, c, "BlueSpark", "BlueSparkDraw")
	if parentType.DrawFunc != legacy.Get_nox_thing_summon_effect_draw() {
		t.Fatal("SummonEffect parser selected a different callback")
	}
	anim := (*summonAnimateDrawData4B7D00)(parentType.DrawData)
	if anim.size != uint32(unsafe.Sizeof(*anim)) || anim.frameCount != 3 || anim.frameDelay != 1 || anim.animationKind != 2 {
		t.Fatal("actual native animation layout differs")
	}
	for i, h := range unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3) {
		if h != c.r.GetBag().ImageByIndex(i).C() {
			t.Fatal("native C parser lost image handle identity")
		}
	}
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = pix.Rect, pix.Rect, pix.Size()
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(output.Close)
	start, elapsed, lifetime := uint32(100), uint32(0), uint16(10)
	pos := image.Pt(72, 72)
	switch style {
	case "half":
		elapsed = 5
	case "last":
		elapsed = 9
	case "expired":
		elapsed = 10
	case "zero-life":
		lifetime = 0
	case "frame-wrap":
		start, elapsed = 0xfffffffd, 5
	case "clip":
		pos = image.Pt(0, 0)
	case "outside":
		pos = image.Pt(1000, 1000)
	}
	packet := []byte{0x7e} // MSG_FX_SUMMON
	packet = binary.LittleEndian.AppendUint16(packet, uint16(pos.X))
	packet = binary.LittleEndian.AppendUint16(packet, uint16(pos.Y))
	packet = binary.LittleEndian.AppendUint16(packet, 19)
	packet = binary.LittleEndian.AppendUint16(packet, uint16(childType.Index()))
	packet = append(packet, 0)
	packet = binary.LittleEndian.AppendUint16(packet, lifetime)
	type observation struct {
		Returned, Count, ListLength int
		Data, Draw                  []byte
		Graph                       [][]byte
		Pixels                      []uint16
		Alpha                       byte
		OtherRNG                    int
	}
	run := func(direct bool) observation {
		d.Reset()
		d.SetClip(true)
		d.SetClipRect(vp.Screen)
		d.SetClipRect2(vp.Screen)
		for i := range pix.Pix {
			pix.Pix[i] = 0x2104
		}
		s.SetFrame(start)
		s.Rand.Logic.Reset(93)
		s.Rand.Other.Reset(17)
		beforePacket := bytes.Clone(packet)
		if got := c.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, packet); got != 1 {
			t.Fatalf("real summon packet result=%d", got)
		}
		parent := cl.Objs.List1
		if parent == nil || parent.TypeIDVal != uint32(parentType.Index()) || cl.Objs.Count != 2 {
			t.Fatal("real summon parent/child creation failed")
		}
		child := parent.UnionSummon().Child
		if child == nil || child.TypeIDVal != uint32(childType.Index()) || child.PosVec != pos || parent.UnionSummon().Lifetime != lifetime || parent.AnimStart != start {
			t.Fatal("real summon packet lost child/lifetime/start")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(parent.C()) <= 0xffffffff || uintptr(child.C()) <= 0xffffffff || uintptr(parent.DrawData) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("native parent/child/animation/viewport must be above 4 GiB")
		}
		parent.AnimFrameSlave = 17
		beforeVP := registeredEffectBytes4B7540(vp)
		beforeData := registeredEffectBytes4B7540(anim)
		beforeImages := slices.Clone(unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3))
		s.SetFrame(start + elapsed)
		var result int
		if direct {
			result = parent.CallDraw(vp)
		} else {
			result = c.callDrawableDraw4B6B80(parent, vp)
		}
		got := observation{Returned: result, Count: cl.Objs.Count, ListLength: len(cl.Objs.AllList1()), Pixels: slices.Clone(pix.Pix), Alpha: d.Alpha(), Data: registeredEffectBytes4B7540(anim), Draw: registeredEffectBytes4B7540(d), OtherRNG: int(s.Rand.Other.Index())}
		live := elapsed < uint32(lifetime)
		if live {
			if result != 1 || parent.PosVec != pos || parent.AnimFrameSlave != 17 || anim.animationKind != 2 || parent.UnionSummon().Child != child || d.IsAlphaEnabled() || d.Alpha() != byte(elapsed*255/uint32(lifetime)) {
				t.Fatal("live summon lost restored fields/child/alpha")
			}
			got.Graph = registeredSummonGraph4B7D00(t, c, parent, child)
			wantCount := 2
			if elapsed == uint32(lifetime)-1 {
				wantCount += 50
			}
			if got.Count != wantCount || got.ListLength != wantCount-1 {
				t.Fatal("live summon spark/pool count differs")
			}
		} else if result != 0 || got.Count != 0 || got.ListLength != 0 {
			t.Fatal("expired summon did not delete both native objects")
		}
		if !bytes.Equal(beforePacket, packet) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || !bytes.Equal(beforeData, got.Data) || !slices.Equal(beforeImages, unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3)) || s.Frame() != start+elapsed || s.TickRate() != 60 || s.Rand.Logic.Index() != 93 {
			t.Fatal("callback mutated frozen packet/parser/viewport/frame/Logic RNG")
		}
		output.CopyBuffer(pix)
		if snap, _ := sc.Snapshot(); snap == nil || !slices.Equal(snap.Pix, pix.Pix) {
			t.Fatal("actual headless surface upload differs from raster")
		}
		if live {
			// Use the real expiry branch to release the otherwise unlinked child.
			s.SetFrame(start + uint32(lifetime))
			if c.drawSummonEffect4B7D00(parent, vp) != 0 {
				t.Fatal("real summon expiry cleanup failed")
			}
		}
		for cl.Objs.List1 != nil {
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(cl.Objs.List1)
		}
		if cl.Objs.Count != 0 {
			t.Fatal("native drawable cleanup incomplete")
		}
		return got
	}
	want, got := run(false), run(true)
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("registered summon differs from existing world rendering: return=%d/%d count=%d/%d pixels=%t graph=%t render=%t", got.Returned, want.Returned, got.Count, want.Count, slices.Equal(got.Pixels, want.Pixels), reflect.DeepEqual(got.Graph, want.Graph), bytes.Equal(got.Draw, want.Draw))
	}
	fmt.Println("REGISTERED SUMMON CALLBACK PASS: " + style)
}
