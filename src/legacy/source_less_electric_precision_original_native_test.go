package legacy

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type sourceLessElectricOriginalItem struct {
	Class       object.Class `json:"class_bits"`
	Flags       object.Flags `json:"flags"`
	Subclass    uint32       `json:"subclass"`
	Material    uint16       `json:"material"`
	Coefficient uint32       `json:"coefficient_bits32"`
	Material7   bool         `json:"material7"`
	Missing     bool         `json:"missing_definition"`
}

type sourceLessElectricOriginalCase struct {
	Name           string                           `json:"name"`
	Player         bool                             `json:"player"`
	Type           object.DamageType                `json:"damage_type"`
	Raw            int32                            `json:"raw"`
	Carry          uint32                           `json:"carry_bits32"`
	Quest          bool                             `json:"quest"`
	Scale          uint32                           `json:"quest_scale_bits32"`
	Items          []sourceLessElectricOriginalItem `json:"items"`
	ExpectedScale  uint32                           `json:"expected_scale_bits32"`
	ExpectedCarry  uint32                           `json:"expected_carry_bits32"`
	ExpectedDamage []uint32                         `json:"expected_damage_bits32"`
}

func sourceLessElectricOriginalFrame(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), int(n))
}

func observeSourceLessElectricOriginalNative(t *testing.T, s *server.Server, c sourceLessElectricOriginalCase) map[string]any {
	t.Helper()
	var free []func()
	defer func() {
		for i := len(free) - 1; i >= 0; i-- {
			free[i]()
		}
	}()
	v, f := alloc.New(server.Object{})
	free = append(free, f)
	hp, f := alloc.New(server.HealthData{})
	free = append(free, f)
	*hp = server.HealthData{Cur: 200, Max: 200, Field2: 200}
	*v = server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, HealthData: hp}
	var carry, marker, kind *uint32
	var update []byte
	var offsets []uintptr
	var frames [][]byte
	if c.Player {
		u, f := alloc.New(server.PlayerUpdateData{})
		free = append(free, f)
		info, f := alloc.New(server.Player{})
		free = append(free, f)
		*info = server.Player{}
		*u = server.PlayerUpdateData{Player: info, State: server.PlayerState13, Field21: c.Carry, Field76: 99, Field75: 77}
		v.ObjClass, v.ObjSubClass, v.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(u)
		carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
		update = sourceLessElectricOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
		offsets = []uintptr{unsafe.Offsetof(u.Field21), unsafe.Offsetof(u.Field76), unsafe.Offsetof(u.Field75)}
		frames = append(frames, sourceLessElectricOriginalFrame(unsafe.Pointer(info), unsafe.Sizeof(*info)))
	} else {
		u, f := alloc.New(server.MonsterUpdateData{})
		free = append(free, f)
		*u = server.MonsterUpdateData{Field1: c.Carry, Field547: 99, Field546: 77}
		v.UpdateData = unsafe.Pointer(u)
		carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
		update = sourceLessElectricOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
		offsets = []uintptr{unsafe.Offsetof(u.Field1), unsafe.Offsetof(u.Field547), unsafe.Offsetof(u.Field546)}
	}
	name, freeName := alloc.CString("Material7")
	free = append(free, freeName)
	effect, f := alloc.New(server.ModifierEff{})
	free = append(free, f)
	*effect = server.ModifierEff{}
	// ModifierEff's canonical first member is a native name pointer. All
	// fixture memory and the string are C-owned, including this private field.
	*(*unsafe.Pointer)(effect.C()) = unsafe.Pointer(name)
	var lastItem *server.Object
	var firstDef, lastDef *server.Modifier
	for i, in := range c.Items {
		item, f := alloc.New(server.Object{})
		free = append(free, f)
		init, f := alloc.New(server.ModifierInitData{})
		free = append(free, f)
		*init = server.ModifierInitData{}
		*item = server.Object{TypeInd: uint16(100 + i), ObjClass: in.Class, ObjSubClass: object.SubClass(in.Subclass), ObjFlags: in.Flags, Material: in.Material, InitData: unsafe.Pointer(init)}
		if in.Material7 {
			init.Modifiers[1] = effect
		}
		if lastItem == nil {
			v.InvFirstItem = item
		} else {
			lastItem.InvNextItem = item
		}
		lastItem = item
		def, f := alloc.New(server.Modifier{})
		free = append(free, f)
		*def = server.Modifier{TypeInd: uint32(item.TypeInd), DamageCoeffOrArmor64: math.Float32frombits(in.Coefficient)}
		if in.Missing {
			def.TypeInd += 900
		}
		if firstDef == nil {
			firstDef = def
		} else {
			lastDef.Next80 = def
		}
		lastDef = def
		frames = append(frames, sourceLessElectricOriginalFrame(unsafe.Pointer(item), unsafe.Sizeof(*item)), sourceLessElectricOriginalFrame(unsafe.Pointer(init), unsafe.Sizeof(*init)), sourceLessElectricOriginalFrame(unsafe.Pointer(def), unsafe.Sizeof(*def)))
	}
	oldDefs := s.Modif.Dword_5d4594_251608
	s.Modif.Dword_5d4594_251608 = firstDef
	defer func() { s.Modif.Dword_5d4594_251608 = oldDefs }()
	if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(unsafe.Pointer(v)) <= 0xffffffff || uintptr(v.UpdateData) <= 0xffffffff {
		t.Fatal("expected C-owned high native pointers")
	}
	frames = append(frames, sourceLessElectricOriginalFrame(unsafe.Pointer(v), unsafe.Sizeof(*v)), sourceLessElectricOriginalFrame(unsafe.Pointer(hp), unsafe.Sizeof(*hp)), sourceLessElectricOriginalFrame(effect.C(), unsafe.Sizeof(*effect)))
	for _, frame := range frames {
		if len(frame) == 0 || uintptr(unsafe.Pointer(&frame[0])) <= 0xffffffff {
			t.Fatal("expected full C-owned high native frame")
		}
	}
	before := make([][]byte, len(frames))
	for i, b := range frames {
		before[i] = append([]byte(nil), b...)
	}
	updateBefore := append([]byte(nil), update...)
	gotScale := math.Float32bits(playerDamageElectricArmorScale4E2220(s, v))
	forward := []uint32{}
	unsupported := []string{}
	r := server.PlayerDamageRuntime4E17B0{
		CoopMode: func() bool { return false }, GodMode: func() bool { return false }, QuestMode: func() bool { return c.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(c.Scale) },
		ItemArmorValue: func(item *server.Object) float32 {
			if def := s.Modif.Nox_xxx_equipClothFindDefByTT413270(int(item.TypeInd)); def != nil {
				return def.DamageCoeffOrArmor64
			}
			return 0
		},
		ElectricArmorScale: func(target *server.Object) float32 {
			if target != v {
				t.Fatal("changed armor target")
			}
			return playerDamageElectricArmorScale4E2220(s, target)
		},
		DefaultDamage: func(target, source, weapon *server.Object, amount int32, typ object.DamageType) bool {
			if target != v || source != nil || weapon != nil || typ != c.Type {
				t.Fatal("changed native identities")
			}
			forward = append(forward, uint32(amount))
			return true
		},
		Unsupported: func(reason string, _, _, _ *server.Object, _ int32, _ object.DamageType) {
			unsupported = append(unsupported, reason)
		},
	}
	handled, result := server.PlayerDamageNative4E17B0(v, nil, nil, c.Raw, c.Type, r)
	for _, off := range offsets {
		copy(updateBefore[off:off+4], update[off:off+4])
	}
	preserved := bytes.Equal(updateBefore, update)
	for i, b := range frames {
		preserved = preserved && bytes.Equal(before[i], b)
	}
	match := handled && result && len(unsupported) == 0 && gotScale == c.ExpectedScale && *carry == c.ExpectedCarry && *marker == 2 && *kind == uint32(c.Type) && reflect.DeepEqual(forward, c.ExpectedDamage) && preserved
	return map[string]any{"name": c.Name, "handled": handled, "result": result, "scale_bits32": gotScale, "carry_bits32": *carry, "damage_bits32": forward, "marker": *marker, "marker_type": *kind, "unsupported": unsupported, "native_nonoutput_frames_preserved": preserved, "matches_original": match}
}

func TestSourceLessElectricPrecisionOriginal4E17B0Native(t *testing.T) {
	s := npcReflectLegacyServer4E17B0(t)
	b, err := os.ReadFile("testdata/source_less_electric_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases              []sourceLessElectricOriginalCase `json:"cases"`
		SpecialCoefficient uint32                           `json:"special_armor_coefficient_bits32"`
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 5376 {
		t.Fatal("incomplete independent original electric inputs")
	}
	if f.SpecialCoefficient != 0x3e19999a {
		t.Fatal("missing originalspecialarmor constant input")
	}
	coefficient := memmap.PtrUint32(0x587000, 201108)
	oldCoefficient := *coefficient
	*coefficient = f.SpecialCoefficient
	defer func() { *coefficient = oldCoefficient }()
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := observeSourceLessElectricOriginalNative(t, s, c)
			if !got["matches_original"].(bool) {
				t.Fatalf("actual native electric mismatch=%v wantScale=%08x wantCarry=%08x wantDamage=%v", got, c.ExpectedScale, c.ExpectedCarry, c.ExpectedDamage)
			}
		})
	}
}
