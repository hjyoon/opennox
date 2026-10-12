//go:build !server

package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
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

type destroyOnceInput48EA70 struct {
	Opcode    byte   `json:"opcode"`
	Static    bool   `json:"static"`
	Connected bool   `json:"connected"`
	Shape     string `json:"shape"`
	Kind      uint32 `json:"kind"`
	Delay     byte   `json:"delay"`
	Start     uint32 `json:"start"`
}

type destroyOncePacket48EA70 struct {
	Input    destroyOnceInput48EA70 `json:"input"`
	Returned int                    `json:"returned"`
	Actions  [][]json.RawMessage    `json:"actions"`
	Flags    uint32                 `json:"flags"`
	Counters [3]uint32              `json:"counters"`
}

type destroyOnceTick48EA70 struct {
	Elapsed  int                 `json:"elapsed"`
	Frame    uint32              `json:"frame"`
	Returned int                 `json:"returned"`
	Actions  [][]json.RawMessage `json:"actions"`
}

// This legal uncompressed bag exercises real image handles and the real C
// parser. It is a bounded colored fixture, not a stock Windows pixel oracle.
func destroyOnceBag48EA70(t *testing.T) {
	t.Helper()
	var data []byte
	idx := make([]byte, 24)
	binary.LittleEndian.PutUint32(idx, 0xfaedbcea)
	binary.LittleEndian.PutUint32(idx[8:], 1)
	for _, v := range []uint32{0, 75, 3} {
		idx = binary.LittleEndian.AppendUint32(idx, v)
	}
	for i, color := range []uint16{0x7c00, 0x03e0, 0x001f} {
		name := fmt.Sprintf("once-%d", i)
		idx = append(idx, byte(len(name)))
		idx = append(idx, name...)
		idx = append(idx, 2)
		idx = binary.LittleEndian.AppendUint32(idx, 25)
		idx = binary.LittleEndian.AppendUint32(idx, 25)
		payload := make([]byte, 17)
		binary.LittleEndian.PutUint32(payload, 2)
		binary.LittleEndian.PutUint32(payload[4:], 2)
		for j := 0; j < 4; j++ {
			payload = binary.LittleEndian.AppendUint16(payload, color)
		}
		data = append(data, payload...)
	}
	for name, payload := range map[string][]byte{"video.bag": data, "video.idx": idx} {
		if err := os.WriteFile(name, payload, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func destroyOnceType48EA70(t *testing.T, c *Client, kind uint32, delay byte) *client.ObjectType {
	t.Helper()
	kinds := [...]string{"OneShot", "OneShotRemove", "Loop", "LoopAndFade", "Random", "Slave"}
	name := fmt.Sprintf("NativeOnce-%d-%d-%s", kind, delay, t.Name())
	if typ := c.Things.TypeByID(name); typ != nil {
		return typ
	}
	var raw []byte
	for _, s := range []string{name, "DRAW", "AnimateDraw"} {
		raw = append(raw, byte(len(s)))
		raw = append(raw, s...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	raw = append(raw, 3, delay, byte(len(kinds[kind])))
	raw = append(raw, kinds[kind]...)
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
		t.Fatalf("real registered AnimateDraw parser failed: %v unread=%d", err, len(f.Data()))
	}
	typ := c.Things.TypeByID(name)
	if typ == nil || typ.DrawFunc != legacy.Get_nox_thing_animate_draw() || typ.DrawData == nil {
		t.Fatal("actual parser did not install AnimateDraw data/callback")
	}
	anim := (*summonAnimateDrawData4B7D00)(typ.DrawData)
	if anim.size != uint32(unsafe.Sizeof(*anim)) || anim.frameCount != 3 || anim.frameDelay != delay || anim.animationKind != kind {
		t.Fatalf("actual C parser/native layout mismatch: %+v", anim)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(typ.C()) <= 0xffffffff || uintptr(typ.DrawData) <= 0xffffffff || uintptr(anim.images) <= 0xffffffff || unsafe.Offsetof(anim.animationKind) != 20) {
		t.Fatal("actual native parser allocations must be high-address with kind at +20")
	}
	for i, handle := range unsafe.Slice((*noxrender.ImageHandle)(anim.images), 3) {
		if handle != c.r.GetBag().ImageByIndex(i).C() {
			t.Fatal("actual parser lost native image-pointer array")
		}
	}
	t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, 3) })
	return typ
}

func TestDestroyOnceNative48EA70HeadlessV5(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "destroy_once_48ea70.json"))
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		PacketRows []destroyOncePacket48EA70 `json:"packet_rows"`
		Timelines  []struct {
			Input  destroyOnceInput48EA70  `json:"input"`
			Packet destroyOncePacket48EA70 `json:"packet"`
			Ticks  []destroyOnceTick48EA70 `json:"ticks"`
		} `json:"timelines"`
	}
	if err := json.Unmarshal(data, &original); err != nil || len(original.PacketRows) != 108 || len(original.Timelines) != 24 {
		t.Fatalf("frozen original transcript invalid: %v", err)
	}
	t.Chdir(t.TempDir())
	destroyOnceBag48EA70(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(8)
	t.Cleanup(cl.Objs.Free)
	cl.Things.Nox_things_free_44C580_B()
	t.Cleanup(cl.Things.Nox_things_free_44C580_B)
	r := cl.Render()
	t.Cleanup(r.Part.Free)
	draw, freeDraw := noxrender.NewRenderData()
	t.Cleanup(freeDraw)
	r.SetData(draw)
	pix := noximage.NewImage16(image.Rect(0, 0, 64, 64))
	r.SetPixBuffer(pix)
	if err := r.GetBag().ReadVideoBag(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.GetBag().Free)
	c := &Client{Client: cl, srv: &Server{Server: s}, r: NewNoxRender(r)}
	// The real image callback interpolates ambient lighting into C-owned
	// storage. Initialize the missing renderer context, not packet flags,
	// frame inputs or expected pixels.
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
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = pix.Rect, pix.Rect, pix.Size()
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}

	packet := func(t *testing.T, row destroyOncePacket48EA70) (*client.Drawable, bool) {
		t.Helper()
		p := row.Input
		typ := destroyOnceType48EA70(t, c, p.Kind, p.Delay)
		dr := cl.Objs.New()
		if dr == nil {
			t.Fatal("actual drawable pool exhausted")
		}
		cl.Objs.Count, cl.Objs.List1 = 1, dr
		t.Cleanup(func() {
			if cl.Objs.List1 == dr {
				c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			}
		})
		dr.NetCode32, dr.TypeIDVal = 0x1234, uint32(typ.Index())
		dr.ObjFlags = 0x01000004
		if p.Static {
			dr.ObjClass = 0x20400000
		}
		dr.DrawData, dr.DrawFuncPtr = typ.DrawData, typ.DrawFunc
		dr.PosVec = image.Pt(24, 24)
		dr.AnimStart = p.Start
		if p.Shape == "nil-data" {
			dr.DrawData = nil
		}
		if p.Shape == "foreign" {
			dr.DrawFuncPtr = legacy.Get_nox_thing_glyph_draw()
		}
		if p.Shape == "missing" {
			dr.NetCode32 = 0x1245
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("actual drawable/viewport are not native high-address records")
		}
		code := uint16(0x1234)
		if p.Static {
			code |= 0x8000
		}
		payload := []byte{p.Opcode, byte(code), byte(code >> 8)}
		beforePacket, beforeData, beforeType := slices.Clone(payload), bytes.Clone(unsafe.Slice((*byte)(typ.DrawData), int(unsafe.Sizeof(summonAnimateDrawData4B7D00{})))), registeredEffectBytes4B7540(typ)
		beforeDR, beforeVP := registeredEffectBytes4B7540(dr), registeredEffectBytes4B7540(vp)
		*connected = 0
		if p.Connected {
			*connected = 1
		}
		s.SetFrame(p.Start)
		logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
		if got := c.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, payload); got != row.Returned {
			t.Fatalf("actual whole packet result=%d original=%d", got, row.Returned)
		}
		deleted := len(row.Actions) != 0 && !p.Static
		if deleted {
			if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.ByNetCode(code) != nil {
				t.Fatal("original dynamic delete did not clean actual pool/list")
			}
		} else {
			if cl.Objs.Count != 1 || cl.Objs.List1 != dr {
				t.Fatalf("original packet lifetime mismatch: count=%d list=%p; freed records are not read", cl.Objs.Count, cl.Objs.List1)
			}
			if uint32(dr.ObjFlags) != row.Flags || [3]uint32{dr.Field_120, dr.Field_121, dr.Field_122} != row.Counters {
				t.Fatalf("original packet lifetime/flags/counters mismatch: count=%d list=%p flags=%x want=%x counters=%v want=%v", cl.Objs.Count, cl.Objs.List1, uint32(dr.ObjFlags), row.Flags, [3]uint32{dr.Field_120, dr.Field_121, dr.Field_122}, row.Counters)
			}
			after := registeredEffectBytes4B7540(dr)
			for i := range beforeDR {
				flag, counter := int(unsafe.Offsetof(dr.ObjFlags)), int(unsafe.Offsetof(dr.Field_120))
				if (i >= flag && i < flag+4) || (i >= counter && i < counter+12) {
					continue
				}
				if after[i] != beforeDR[i] {
					t.Fatalf("packet changed unrelated drawable byte %d", i)
				}
			}
		}
		if !bytes.Equal(payload, beforePacket) || !bytes.Equal(beforeData, unsafe.Slice((*byte)(typ.DrawData), len(beforeData))) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != p.Start || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
			t.Fatal("actual packet changed input/parser data/type/viewport/frame/RNG")
		}
		return dr, !deleted
	}

	for i, row := range original.PacketRows {
		t.Run(fmt.Sprintf("packet-%03d", i), func(t *testing.T) {
			dr, alive := packet(t, row)
			if alive {
				c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			}
		})
	}
	for i, timeline := range original.Timelines {
		t.Run(fmt.Sprintf("lifetime-%02d", i), func(t *testing.T) {
			dr, alive := packet(t, timeline.Packet)
			if !alive {
				t.Fatal("one-shot was deleted before its first frame")
			}
			beforeData := bytes.Clone(unsafe.Slice((*byte)(dr.DrawData), int(unsafe.Sizeof(summonAnimateDrawData4B7D00{}))))
			for _, tick := range timeline.Ticks {
				clear(pix.Pix)
				*draw = noxrender.RenderData{}
				draw.SetClip(true)
				draw.SetClipRect(pix.Rect)
				draw.SetClipRect2(pix.Rect)
				s.SetFrame(tick.Frame)
				if got := dr.CallDraw(vp); got != tick.Returned {
					t.Fatalf("actual C animation result at %d = %d original=%d", tick.Elapsed, got, tick.Returned)
				}
				want := make([]uint16, len(pix.Pix))
				if tick.Returned != 0 {
					var frame int
					if err := json.Unmarshal(tick.Actions[0][1], &frame); err != nil {
						t.Fatal(err)
					}
					color := [...]uint16{0x7c00, 0x03e0, 0x001f}[frame]
					for y := 24; y < 26; y++ {
						for x := 24; x < 26; x++ {
							want[y*64+x] = color
						}
					}
					if cl.Objs.Count != 1 || cl.Objs.List1 != dr {
						t.Fatal("actual one-shot expired before original boundary")
					}
				} else if cl.Objs.Count != 0 || cl.Objs.List1 != nil || cl.Objs.DeadlineList != nil {
					t.Fatal("natural expiry did not clean actual drawable pool/list")
				}
				presentParticleRegistration4B01B0(t, pix, output, sc, want, tick.Elapsed)
				t.Logf("ONCE ORIGINAL FRAME: opcode=%d static=%t delay=%d start=%08x elapsed=%d original-C-selected-frame/full-raster/headless-upload", timeline.Input.Opcode, timeline.Input.Static, timeline.Input.Delay, timeline.Input.Start, tick.Elapsed)
			}
			if !bytes.Equal(beforeData, unsafe.Slice((*byte)(destroyOnceType48EA70(t, c, 1, timeline.Input.Delay).DrawData), len(beforeData))) {
				t.Fatal("actual C lifecycle changed shared parser animation data")
			}
		})
	}
	if sc.PresentCount() != 168 {
		t.Fatalf("headless lifetime frames=%d original=168", sc.PresentCount())
	}
}
