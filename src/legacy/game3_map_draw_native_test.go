package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func game3TestDrawable(t *testing.T) *client.Drawable {
	t.Helper()
	p := game3TestAlloc(unsafe.Sizeof(client.Drawable{}))
	if p == nil {
		t.Fatal("C allocation failed")
	}
	t.Cleanup(func() {
		if !game3TestGuards(p, unsafe.Sizeof(client.Drawable{})) {
			t.Error("native drawable canary changed")
		}
		game3TestFree(p)
	})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
		t.Fatalf("C drawable must exercise a full native address: %p", p)
	}
	return (*client.Drawable)(p)
}
func game3TestSnapshot(d *client.Drawable) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(d.C()), int(unsafe.Sizeof(*d)))...)
}
func game3TestMapIO(t *testing.T, raw []byte, d *client.Drawable, fail bool) (game3MapServices, *int) {
	t.Helper()
	cursor := new(int)
	h := game3MapServices{
		rw: func(p []byte) {
			if *cursor+len(p) > len(raw) {
				panic("bounded input exhausted")
			}
			copy(p, raw[*cursor:*cursor+len(p)])
			*cursor += len(p)
		},
		seek: func(n int32) {
			next := *cursor + int(n)
			if next < 0 || next > len(raw) {
				panic("bounded seek outside literal input")
			}
			*cursor = next
		},
		sprite: func(id int32, p image.Point) *client.Drawable {
			if fail {
				return nil
			}
			d.TypeIDVal = uint32(id)
			d.PosVec = p
			return d
		},
		typeID: func(string) uint32 { return 999 }, team: func(*client.Drawable) {}, floatInt: game3MapFloatInt,
		readOnly: func() bool { return false }, failed: func() {},
	}
	return h, cursor
}
func game3TestMapRecord(version uint16, withFlags bool) []byte {
	var b bytes.Buffer
	put := func(v any) {
		if err := binary.Write(&b, binary.LittleEndian, v); err != nil {
			panic(err)
		}
	}
	put(version)
	put(uint32(0xe1234567))
	put(uint32(0x01020408))
	put(float32(-4))
	put(float32(22))
	if version < 61 {
		put(byte(0))
		put(byte(7))
		put(byte(0xfe))
		put(uint32(0xabababab))
		put(uint16(2))
		put([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
		put(uint32(0x8000005e))
		return b.Bytes()
	}
	if !withFlags {
		put(byte(0))
		return b.Bytes()
	}
	put(byte(1))
	put(uint32(0x01020408))
	put(byte(3))
	put([3]byte{'a', 0, 'z'})
	put(byte(7))
	put(byte(0xfe))
	put(uint16(2))
	put([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	put(uint32(0x8000005e))
	if version >= 63 {
		put(uint16(0x1020))
		put(uint32(3))
		put([3]byte{8, 9, 10})
		put(uint32(0x40404040))
	}
	if version >= 64 {
		put(uint32(0x50505050))
	}
	return b.Bytes()
}

func TestGame3MapSpriteCABIAndRetainedCurrentC(t *testing.T) {
	layout := game3TestLayouts()
	align := unsafe.Alignof(client.Drawable{})
	prefix := (unsafe.Offsetof(client.Drawable{}.Field_127) + 4 + align - 1) &^ (align - 1)
	// Go appends a private client handle after the exact native C prefix.
	if layout[0] != prefix || unsafe.Sizeof(client.Drawable{}) != prefix+unsafe.Sizeof(uintptr(0)) || layout[1] != unsafe.Sizeof(server.EquipmentData{}) || layout[2] != unsafe.Offsetof(server.EquipmentData{}.Field4) || layout[3] != unsafe.Offsetof(client.Drawable{}.Union) {
		t.Fatalf("actual native C/Go layout mismatch: %v", layout)
	}
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(1)
	defer func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) }()
	oldFactory := game3MapFactory
	defer func() { game3MapFactory = oldFactory }()
	for _, version := range []uint16{5, 61, 62, 63, 64} {
		for _, withFlags := range []bool{false, true} {
			if version == 5 && !withFlags {
				continue
			}
			for _, fail := range []bool{false, true} {
				t.Run(fmt.Sprintf("%d/flags%t/nil%t", version, withFlags, fail), func(t *testing.T) {
					reference, actual := game3TestDrawable(t), game3TestDrawable(t)
					for _, d := range []*client.Drawable{reference, actual} {
						d.ObjFlags = object.Flags(0xfedcba98)
						d.Flags70Val = 0x102030ff
						d.TeamVal.Field0 = 0xaabbccdd
						d.TeamVal.ID = 11
					}
					raw := game3TestMapRecord(version, withFlags)
					before := append([]byte(nil), raw...)
					ref, refCount, refCursor := game3TestReference(raw, reference, 17, 40, fail)
					hooks, cursor := game3TestMapIO(t, raw, actual, fail)
					game3MapFactory = func() game3MapServices { return hooks }
					got, count := game3TestPublicMap(17, 40)
					if (got == nil) != (ref == nil) || count != refCount || *cursor != refCursor || !bytes.Equal(game3TestSnapshot(actual), game3TestSnapshot(reference)) {
						t.Fatalf("native public entry differs from preserved C: count=%d/%d cursor=%d/%d", count, refCount, *cursor, refCursor)
					}
					if !bytes.Equal(raw, before) {
						t.Fatal("wire input changed")
					}
					if !fail && (got != actual || got.NetCode32 != 0xe1234567 || got.PosVec != image.Pt(-4, 22)) {
						t.Fatalf("full pointer/WORD-DWORD coordinates: %p %+v", got, actual.PosVec)
					}
				})
			}
		}
	}
	runtime.GC()
}

func TestGame3MapLightAndFullModifierPointers(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(1)
	defer func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) }()
	oldFactory := game3MapFactory
	defer func() { game3MapFactory = oldFactory }()
	t.Run("native-four-modifier-slots", func(t *testing.T) {
		d := game3TestDrawable(t)
		raw := append([]byte{42, 0}, game3TestMapRecord(61, false)...)
		for _, s := range []string{"one", "two\x00ignored", "three", "four"} {
			raw = append(raw, byte(len(s)))
			raw = append(raw, []byte(s)...)
		}
		owners := []*server.ModifierEff{{}, {}, {}, {}}
		var pins runtime.Pinner
		for _, p := range owners {
			pins.Pin(p)
		}
		data := (*client.DrawableUnionItem)(unsafe.Pointer(&d.Union))
		defer func() {
			data.Field_108 = nil
			data.Field_109 = nil
			data.Field_110 = nil
			data.Field_111 = nil
			pins.Unpin()
		}()
		h, cursor := game3TestMapIO(t, raw, d, false)
		names := []string{}
		h.modifier = func(name string) unsafe.Pointer {
			names = append(names, name)
			return unsafe.Pointer(owners[len(names)-1])
		}
		game3MapFactory = func() game3MapServices { return h }
		got := game3TestPublicLoad(17, 2)
		if int(got) != len(raw) || *cursor != len(raw) || !reflect.DeepEqual(names, []string{"one", "two", "three", "four"}) {
			t.Fatalf("exact stream/C-name boundary %d/%d %v", got, *cursor, names)
		}
		runtime.GC()
		for i, p := range owners {
			if game3TestItemModifier(d, i) != unsafe.Pointer(p) {
				t.Fatalf("modifier %d truncated/lost", i)
			}
		}
		if data.Field_112_0 != -1 || data.Field_112_2 != -1 {
			t.Fatal("native trailing WORD defaults")
		}
	})
	t.Run("fixed140-light-native-RGB", func(t *testing.T) {
		d := game3TestDrawable(t)
		raw := append([]byte{42, 0}, game3TestMapRecord(61, false)...)
		var light [140]byte
		for i := range light {
			light[i] = byte(13*i + 5)
		}
		binary.LittleEndian.PutUint32(light[16:], 0xffffffff)
		binary.LittleEndian.PutUint32(light[20:], 0x7fffffff)
		binary.LittleEndian.PutUint32(light[24:], 0x80000000)
		spans := [][2]int{{0, 4}, {4, 4}, {8, 4}, {12, 4}, {16, 12}, {28, 2}, {30, 2}, {32, 4}, {40, 2}, {42, 48}, {90, 16}, {106, 16}, {122, 2}, {124, 2}, {126, 2}, {128, 4}, {134, 2}, {136, 2}, {138, 1}, {36, 4}}
		for _, s := range spans {
			raw = append(raw, light[s[0]:s[0]+s[1]]...)
		}
		h, cursor := game3TestMapIO(t, raw, d, false)
		game3MapFactory = func() game3MapServices { return h }
		n := game3TestPublicLoad(17, 1)
		if int(n) != len(raw) || *cursor != len(raw) || d.LightColor.R != -1 || d.LightColor.G != 2147483647 || int64(d.LightColor.B) != -2147483648 {
			t.Fatalf("fixed-wire light projection/count: %d/%d RGB=%+v", n, len(raw), d.LightColor)
		}
		got := d.LightXferData()
		for _, s := range spans {
			if !bytes.Equal(got[s[0]:s[0]+s[1]], light[s[0]:s[0]+s[1]]) {
				t.Fatalf("light span %v shifted at native width", s)
			}
		}
	})
}

func TestGame3BounceCABIOriginalSignedWidths(t *testing.T) {
	old := game3EffectsFactory
	defer func() { game3EffectsFactory = old }()
	actual, ref := game3TestDrawable(t), game3TestDrawable(t)
	for _, z := range []int16{-32768, -1, 0, 1, 32767} {
		for _, v := range []int8{-128, -3, -1, 0, 1, 4, 127} {
			for _, frame := range []uint32{0, 1} {
				actual.ZVal = uint16(z)
				ref.ZVal = uint16(z)
				actual.VelZ = v
				ref.VelZ = v
				game3EffectsFactory = func() game3EffectServices { return game3EffectServices{frame: func() uint32 { return frame }} }
				want := game3TestReferenceBounce(ref, frame)
				got := game3TestPublicBounce(actual)
				if got != want || !bytes.Equal(game3TestSnapshot(actual), game3TestSnapshot(ref)) {
					t.Fatalf("WORD=%d BYTE=%d frame=%d return=%d/%d", z, v, frame, got, want)
				}
			}
		}
	}
	runtime.GC()
}

func TestGame3SummonFullChildPointerAndFreeOrder(t *testing.T) {
	oldBlue := game3TestOwner(0)
	defer game3TestSetOwner(0, oldBlue)
	parent, child := game3TestDrawable(t), game3TestDrawable(t)
	old := game3EffectsFactory
	defer func() { game3EffectsFactory = old }()
	clock := uint32(0xfedcba98)
	calls := []string{}
	h := game3EffectServices{typeID: func(s string) uint32 { calls = append(calls, s); return 17 }, load: func(id uint32, p image.Point) *client.Drawable {
		if id != 17 || p != image.Pt(65535, 32768) {
			t.Error("literal packet coordinates/thing")
		}
		parent.PosVec = p
		parent.TypeIDVal = id
		return parent
	}, newChild: func(id uint16) *client.Drawable {
		if id != 0xaffe {
			t.Error("WORD thing id")
		}
		return child
	}, direction: func(v byte) byte { return v ^ 0x80 }, frame: func() uint32 { v := clock; clock++; return v }}
	game3EffectsFactory = func() game3EffectServices { return h }
	got := game3TestPublicSummon(-123, [2]uint16{65535, 32768}, 0xaffe, 255, -4)
	runtime.GC()
	if got != 0xfedcba98 || parent.AnimStart != 0xfedcba99 || game3TestSummonChild(parent) != child.C() || parent.UnionSummon().ID != uint16(0xff85) || parent.UnionSummon().Lifetime != 65532 || child.AnimDir != 127 || child.AnimInd != 8 {
		t.Fatalf("native summon child/token fields: %#x child=%p", got, game3TestSummonChild(parent))
	}
	p := memmap.PtrUint32(0x5D4594, 1313744)
	save := *p
	*p = 17
	defer func() { *p = save }()
	h.first = func() *client.Drawable { return parent }
	h.pointFX = func(uint32, image.Point) { calls = append(calls, "spark") }
	h.deleteChild = func(d *client.Drawable) {
		if d != child {
			t.Error("truncated child passed to deletion")
		}
		calls = append(calls, "child")
	}
	h.deleteSelf = func(d *client.Drawable) {
		if d != parent {
			t.Error("parent identity")
		}
		calls = append(calls, "parent")
	}
	game3EffectsFactory = func() game3EffectServices { return h }
	game3TestPublicRemoveSummon(-123)
	if !reflect.DeepEqual(calls[len(calls)-3:], []string{"spark", "child", "parent"}) {
		t.Fatalf("delete order %v", calls)
	}
}

func TestGame3ShieldTypedPointerAndOpaqueSentinel(t *testing.T) {
	oldFound := game3TestOwner(1)
	defer game3TestSetOwner(1, oldFound)
	offsets := []uintptr{1313748, 1313752, 1313756, 1313760, 1313764, 1313768, 1313772, 1313776, 1313780, 1313784}
	saved := make([]uint32, len(offsets))
	for i, off := range offsets {
		saved[i] = memmap.Uint32(0x5D4594, off)
		*memmap.PtrUint32(0x5D4594, off) = uint32(1000 + i)
	}
	*memmap.PtrUint32(0x5D4594, 1313784) = 1
	defer func() {
		for i, off := range offsets {
			*memmap.PtrUint32(0x5D4594, off) = saved[i]
		}
	}()
	owner, existing, created := game3TestDrawable(t), game3TestDrawable(t), game3TestDrawable(t)
	owner.PosVec = image.Pt(50, 70)
	existing.TypeIDVal = 1001
	existing.UnionEffect().Field_108 = 0x80001234
	old := game3EffectsFactory
	defer func() { game3EffectsFactory = old }()
	seen := []string{}
	loadCalls := 0
	h := game3EffectServices{typeID: func(s string) uint32 { seen = append(seen, s); return 77 }, byCode: func(code uint32) *client.Drawable {
		if code != 0x80001234 {
			t.Error("DWORD network id changed")
		}
		return owner
	}, each: func(r image.Rectangle, f func(*client.Drawable)) {
		if r != image.Rect(40, 60, 60, 80) {
			t.Error("native scan rectangle")
		}
		f(existing)
	}, load: func(id uint32, p image.Point) *client.Drawable {
		loadCalls++
		if id != 1001 || p != image.Pt(50, 73) {
			t.Error("typed shield location/id")
		}
		return created
	}}
	game3EffectsFactory = func() game3EffectServices { return h }
	if got := game3TestPublicShield(0x80001234, 1); got != 1 || loadCalls != 0 {
		t.Fatalf("opaque found sentinel %#x loads=%d", got, loadCalls)
	}
	h.each = func(image.Rectangle, func(*client.Drawable)) {}
	game3EffectsFactory = func() game3EffectServices { return h }
	if got := game3TestPublicShield(0x80001234, 1); got != uintptr(created.C()) || created.UnionEffect().Field_108 != 0x80001234 {
		t.Fatalf("full native shield %#x", got)
	}
	if !reflect.DeepEqual(seen, []string{"ShpericalShieldN", "ShpericalShieldN"}) {
		t.Fatalf("preserved lookup spelling/order %v", seen)
	}
}

func TestGame3EquipmentNativeStrideModifiersAndGC(t *testing.T) {
	n := unsafe.Sizeof([server.PlayerWeaponCnt]server.EquipmentData{})
	p := game3TestAlloc(n)
	if p == nil {
		t.Fatal("C allocation")
	}
	defer game3TestFree(p)
	entries := (*[server.PlayerWeaponCnt]server.EquipmentData)(p)
	index := 26
	entries[index].Field0 = 0x80000000
	owners := []*server.ModifierEff{{}, {}, {}, {}}
	var pins runtime.Pinner
	for _, v := range owners {
		pins.Pin(v)
	}
	defer func() { clear(entries[index].Field4[:]); pins.Unpin() }()
	for i, v := range owners {
		v.Color24.R = byte(100 + i)
		v.Color24.G = byte(150 + i)
		v.Color24.B = byte(200 + i)
		entries[index].Field4[i] = unsafe.Pointer(v)
	}
	def := new(server.Modifier)
	def.Effectiveness36 = 9
	def.Material40 = 10
	def.PriEnchant44 = 11
	def.SecEnchant48 = 12
	for i := 1; i <= 6; i++ {
		def.Colors12[i].R = byte(i)
		def.Colors12[i].G = byte(2 * i)
		def.Colors12[i].B = byte(3 * i)
	}
	calls := [][4]int{}
	old := game3MaterialFactory
	defer func() { game3MaterialFactory = old }()
	game3MaterialFactory = func() game3MaterialServices {
		return game3MaterialServices{definition: func(bit uint32) *server.Modifier {
			if bit != 0x80000000 {
				t.Error("wire bit narrowed")
			}
			return def
		}, set: func(i, r, g, b int) { calls = append(calls, [4]int{i, r, g, b}) }}
	}
	runtime.GC()
	got := game3TestPublicMaterial(p, 0x80000000)
	if got != uintptr(unsafe.Pointer(owners[3])) || len(calls) != 10 {
		t.Fatalf("native modifier return/call count %#x %v", got, calls)
	}
	for i := 1; i <= 6; i++ {
		if calls[i-1] != [4]int{i, i, 2 * i, 3 * i} {
			t.Fatalf("base material %d", i)
		}
	}
	for i := 0; i < 4; i++ {
		if calls[6+i] != [4]int{9 + i, 100 + i, 150 + i, 200 + i} {
			t.Fatalf("modifier order %d", i)
		}
	}
	if !game3TestGuards(p, n) {
		t.Fatal("native equipment array canary changed")
	}
	if got := game3TestPublicMaterial(p, 0x12345678); got != uintptr(unsafe.Add(p, n)) {
		t.Fatalf("native one-past no-match return %#x", got)
	}
}
