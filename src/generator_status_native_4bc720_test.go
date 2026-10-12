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

type generatorInput4BC720 struct {
	Count     byte   `json:"count"`
	Delay     byte   `json:"delay"`
	Static    bool   `json:"static"`
	Connected bool   `json:"connected"`
	Generator bool   `json:"generator"`
	Missing   bool   `json:"missing"`
	Old       uint32 `json:"old"`
	New       uint32 `json:"new"`
	Kind      uint32 `json:"kind"`
}

type generatorResult4BC720 struct {
	Input    generatorInput4BC720 `json:"input"`
	Returned int                  `json:"returned"`
	Flags    uint32               `json:"flags"`
	Timer    uint32               `json:"timer"`
	Class    uint32               `json:"cls"`
	ObjFlags uint32               `json:"obj_flags"`
	Frame    uint32               `json:"frame"`
	Images   []int                `json:"images"`
}

// This schema is independently checked against the real C parser, not a
// packed PE32 fixture. Existing packed branch tests remain unchanged.
type generatorParserData4BC720 struct {
	Size   uint32
	Images [5]*noxrender.ImageHandle
	Count  [5]byte
	Delay  [5]byte
	Kind   [5]uint32
}

var generatorColors4BC720 = [...]uint16{0x7c00, 0x03e0, 0x001f, 0x7fff, 0x7fe0, 0x7c1f, 0x4210}

func generatorBag4BC720(t *testing.T) {
	t.Helper()
	idx := make([]byte, 24)
	binary.LittleEndian.PutUint32(idx, 0xfaedbcea)
	binary.LittleEndian.PutUint32(idx[8:], 1)
	for _, v := range []uint32{0, 175, 7} {
		idx = binary.LittleEndian.AppendUint32(idx, v)
	}
	var data []byte
	for i, color := range generatorColors4BC720 {
		name := fmt.Sprintf("generator-%d", i)
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

func generatorType4BC720(t *testing.T, c *Client, p generatorInput4BC720) *client.ObjectType {
	t.Helper()
	name := fmt.Sprintf("NativeGenerator-%d-%d-%d-%s", p.Count, p.Delay, p.Kind, t.Name())
	var raw []byte
	for _, s := range []string{name, "DRAW", "MonsterGeneratorDraw"} {
		raw = append(raw, byte(len(s)))
		raw = append(raw, s...)
	}
	for len(raw)%8 != 0 {
		raw = append(raw, 0)
	}
	raw = append(raw, make([]byte, 8)...)
	raw = append(raw, 5)
	kind := "OneShot"
	if p.Kind == 2 {
		kind = "Loop"
	}
	for state := 0; state < 5; state++ {
		raw = append(raw, p.Count, p.Delay, byte(len(kind)))
		raw = append(raw, kind...)
		for i := 0; i < int(p.Count); i++ {
			raw = binary.LittleEndian.AppendUint32(raw, uint32(i%len(generatorColors4BC720)))
		}
	}
	raw = append(raw, 0)
	buf, _ := alloc.CloneSlice(raw)
	f := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	defer f.Free()
	scratch, freeScratch := alloc.Make([]byte{}, 256)
	defer freeScratch()
	if err := c.Things.ReadType(f, scratch); err != nil || len(f.Data()) != 0 {
		t.Fatalf("actual registered parser failed: %v unread=%d", err, len(f.Data()))
	}
	typ := c.Things.TypeByID(name)
	if typ == nil || typ.DrawData == nil || typ.DrawFunc != legacy.Get_nox_thing_monster_gen_draw() {
		t.Fatal("real conditional parser/registry did not install generator data/callback")
	}
	d := (*generatorParserData4BC720)(typ.DrawData)
	t.Cleanup(func() { legacy.Nox_xxx_draw_44C650_free_kind(typ.DrawData, 4) })
	if d.Size != uint32(unsafe.Sizeof(*d)) {
		t.Fatalf("actual parser size=%d native=%d", d.Size, unsafe.Sizeof(*d))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(typ.C()) <= 0xffffffff || uintptr(typ.DrawData) <= 0xffffffff || unsafe.Offsetof(d.Count) != 48 || unsafe.Offsetof(d.Delay) != 53) {
		t.Fatal("actual parser must be high-address native layout")
	}
	for state := 0; state < 5; state++ {
		if d.Count[state] != p.Count || d.Delay[state] != p.Delay || d.Kind[state] != p.Kind {
			t.Fatalf("actual C state%d metadata differs: %+v", state, d)
		}
		if p.Count != 0 && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(d.Images[state])) <= 0xffffffff {
			t.Fatal("C pointer table must be above4GiB")
		}
		for i, h := range unsafe.Slice(d.Images[state], int(p.Count)) {
			if h != c.r.GetBag().ImageByIndex(i%7).C() {
				t.Fatal("actual C parser lost native image pointers")
			}
		}
	}
	return typ
}

func TestGeneratorStatusNative4BC720Headless(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "generator_status_4bc720.json"))
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		PacketRows []generatorResult4BC720 `json:"packet_rows"`
		Timelines  []struct {
			Input  generatorInput4BC720    `json:"input"`
			Packet generatorResult4BC720   `json:"packet"`
			Ticks  []generatorResult4BC720 `json:"ticks"`
		} `json:"timelines"`
	}
	if err := json.Unmarshal(raw, &original); err != nil || len(original.PacketRows) != 152 || len(original.Timelines) != 16 {
		t.Fatalf("frozen original transcript invalid: %v", err)
	}
	t.Chdir(t.TempDir())
	generatorBag4BC720(t)
	handles.Init()
	t.Cleanup(handles.Release)
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	cl := client.NewClient(nil, nil, s)
	t.Cleanup(cl.Close)
	t.Cleanup(func() { alloc.FreePtr(cl.Viewport().C()) })
	cl.Objs.Init(4)
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
	vp := cl.Viewport()
	vp.Screen, vp.World, vp.Size = pix.Rect, pix.Rect, pix.Size()
	sc := headless.New(pix.Size())
	t.Cleanup(func() { _ = sc.Close() })
	output, err := render.New(sc)
	if err != nil {
		t.Fatal(err)
	}

	packet := func(t *testing.T, row generatorResult4BC720) (*client.Drawable, *client.ObjectType) {
		t.Helper()
		p := row.Input
		typ := generatorType4BC720(t, c, p)
		dr := cl.Objs.New()
		if dr == nil {
			t.Fatal("real drawable pool exhausted")
		}
		cl.Objs.Count, cl.Objs.List1 = 1, dr
		t.Cleanup(func() {
			if cl.Objs.List1 == dr {
				c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			}
		})
		dr.NetCode32, dr.TypeIDVal = 0x1234, uint32(typ.Index())
		if p.Missing {
			dr.NetCode32 = 0x1245
		}
		dr.ObjClass = 0x80000
		if p.Generator {
			dr.ObjClass |= 0x20000
		}
		if p.Static {
			dr.ObjClass |= 0x20400000
		}
		dr.ObjFlags = 0x21000004
		dr.Flags70Val = p.Old
		dr.UnionEffect().Field_108 = 777
		dr.DrawData, dr.DrawFuncPtr = typ.DrawData, typ.DrawFunc
		dr.PosVec = image.Pt(24, 24)
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(dr.C()) <= 0xffffffff || uintptr(vp.C()) <= 0xffffffff) {
			t.Fatal("real native drawable/viewport must be high-address")
		}
		code := uint16(0x1234)
		if p.Static {
			code |= 0x8000
		}
		payload := []byte{101, byte(code), byte(code >> 8)}
		payload = binary.LittleEndian.AppendUint32(payload, p.New)
		beforePacket, beforeDR, beforeType, beforeVP := slices.Clone(payload), registeredEffectBytes4B7540(dr), registeredEffectBytes4B7540(typ), registeredEffectBytes4B7540(vp)
		beforeData := bytes.Clone(unsafe.Slice((*byte)(typ.DrawData), int(unsafe.Sizeof(generatorParserData4BC720{}))))
		*connected = 0
		if p.Connected {
			*connected = 1
		}
		s.SetFrame(100)
		logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
		if got := c.nox_xxx_netOnPacketRecvCli48EA70(server.HostPlayerIndex, payload); got != row.Returned {
			t.Fatalf("actual whole packet result=%d original=%d", got, row.Returned)
		}
		if dr.Flags70Val != row.Flags || dr.UnionEffect().Field_108 != row.Timer || uint32(dr.ObjClass) != row.Class || uint32(dr.ObjFlags) != row.ObjFlags {
			t.Fatalf("actual status/startup state differs: flags=%x/%x timer=%d/%d class=%x/%x obj_flags=%x/%x", dr.Flags70Val, row.Flags, dr.UnionEffect().Field_108, row.Timer, uint32(dr.ObjClass), row.Class, uint32(dr.ObjFlags), row.ObjFlags)
		}
		after := registeredEffectBytes4B7540(dr)
		mutable := []struct{ at, size int }{{int(unsafe.Offsetof(dr.Flags70Val)), 4}, {int(unsafe.Offsetof(dr.ObjClass)), 4}, {int(unsafe.Offsetof(dr.ObjFlags)), 4}, {int(unsafe.Offsetof(dr.Union)), 4}}
		for i, v := range beforeDR {
			allowed := false
			for _, span := range mutable {
				allowed = allowed || (i >= span.at && i < span.at+span.size)
			}
			if !allowed && after[i] != v {
				t.Fatalf("actual packet changed unrelated drawable byte%d", i)
			}
		}
		if !bytes.Equal(beforePacket, payload) || !bytes.Equal(beforeType, registeredEffectBytes4B7540(typ)) || !bytes.Equal(beforeData, unsafe.Slice((*byte)(typ.DrawData), len(beforeData))) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != 100 || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || cl.Objs.Count != 1 || cl.Objs.List1 != dr {
			t.Fatal("actual packet mutated input/parser/type/viewport/frame/RNG/pool")
		}
		return dr, typ
	}
	for i, row := range original.PacketRows {
		t.Run(fmt.Sprintf("packet-%03d", i), func(t *testing.T) { packet(t, row) })
	}
	for i, timeline := range original.Timelines {
		t.Run(fmt.Sprintf("lifetime-%02d", i), func(t *testing.T) {
			dr, typ := packet(t, timeline.Packet)
			beforeData := bytes.Clone(unsafe.Slice((*byte)(typ.DrawData), int(unsafe.Sizeof(generatorParserData4BC720{}))))
			beforeVP := registeredEffectBytes4B7540(vp)
			logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
			for _, tick := range timeline.Ticks {
				clear(pix.Pix)
				*draw = noxrender.RenderData{}
				draw.SetClip(true)
				draw.SetClipRect(pix.Rect)
				draw.SetClipRect2(pix.Rect)
				s.SetFrame(tick.Frame)
				if got := dr.CallDraw(vp); got != tick.Returned {
					t.Fatalf("actual registered generator callback result=%d original=%d", got, tick.Returned)
				}
				if dr.Flags70Val != tick.Flags || dr.UnionEffect().Field_108 != tick.Timer || uint32(dr.ObjClass) != tick.Class || uint32(dr.ObjFlags) != tick.ObjFlags {
					t.Fatalf("generator timer/terminal state differs at frame%d: flags=%x/%x timer=%d/%d class=%x/%x obj_flags=%x/%x", tick.Frame, dr.Flags70Val, tick.Flags, dr.UnionEffect().Field_108, tick.Timer, uint32(dr.ObjClass), tick.Class, uint32(dr.ObjFlags), tick.ObjFlags)
				}
				if len(tick.Images) != 1 {
					t.Fatal("frozen original image service count differs")
				}
				want := make([]uint16, len(pix.Pix))
				color := generatorColors4BC720[tick.Images[0]]
				for y := 24; y < 26; y++ {
					for x := 24; x < 26; x++ {
						want[y*64+x] = color
					}
				}
				presentParticleRegistration4B01B0(t, pix, output, sc, want, int(tick.Frame))
				if !bytes.Equal(beforeData, unsafe.Slice((*byte)(typ.DrawData), len(beforeData))) || !bytes.Equal(beforeVP, registeredEffectBytes4B7540(vp)) || s.Frame() != tick.Frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || cl.Objs.Count != 1 || cl.Objs.List1 != dr {
					t.Fatal("actual generator callback changed shared data/viewport/frame/RNG/pool")
				}
				t.Logf("GENERATOR ORIGINAL FRAME: case=%d frame=%d timer=%d actual-C-entry/native-parser/full-packet/full-raster/headless-upload", i, tick.Frame, tick.Timer)
			}
		})
	}
	if sc.PresentCount() != 192 {
		t.Fatalf("actual headless frames=%d original=192", sc.PresentCount())
	}
}
