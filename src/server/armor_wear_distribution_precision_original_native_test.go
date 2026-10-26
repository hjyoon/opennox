package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type armorWearDistributionOriginalCase struct {
	Name        string            `json:"name"`
	Class       object.Class      `json:"class_bits"`
	Subclass    object.SubClass   `json:"subclass"`
	Armor       uint32            `json:"armor_bits32"`
	Coefficient uint32            `json:"coefficient_bits32"`
	Raw         int32             `json:"raw"`
	Type        object.DamageType `json:"damage_type"`
	ItemClass   object.Class      `json:"item_class"`
	ItemFlags   object.Flags      `json:"item_flags"`
	Empty       bool              `json:"empty"`
	Missing     bool              `json:"missing"`
	NullUpdate  bool              `json:"null_update"`
	Expected    []uint32          `json:"expected_portions_bits32"`
}

func armorWearDistributionOriginalFrame(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), int(n))
}

func observeArmorWearDistributionOriginalNative(t *testing.T, s *Server, c armorWearDistributionOriginalCase) map[string]any {
	t.Helper()
	var frees []func()
	defer func() {
		for i := len(frees) - 1; i >= 0; i-- {
			frees[i]()
		}
	}()
	v, f := alloc.New(Object{})
	frees = append(frees, f)
	*v = Object{ObjClass: c.Class, ObjSubClass: c.Subclass}
	source, f := alloc.New(Object{})
	frees = append(frees, f)
	*source = Object{}
	effective, f := alloc.New(Object{})
	frees = append(frees, f)
	*effective = Object{}
	item, f := alloc.New(Object{})
	frees = append(frees, f)
	init, f := alloc.New(ModifierInitData{})
	frees = append(frees, f)
	*init = ModifierInitData{}
	*item = Object{TypeInd: 300, ObjClass: c.ItemClass, ObjFlags: c.ItemFlags, InitData: unsafe.Pointer(init)}
	def, f := alloc.New(Modifier{})
	frees = append(frees, f)
	*def = Modifier{TypeInd: 300, DamageCoeffOrArmor64: math.Float32frombits(c.Coefficient)}
	if c.Missing {
		def.TypeInd = 301
	}
	old := s.Modif.Dword_5d4594_251608
	s.Modif.Dword_5d4594_251608 = def
	defer func() { s.Modif.Dword_5d4594_251608 = old }()
	var update []byte
	if c.Class.Has(object.ClassPlayer) {
		u, f := alloc.New(PlayerUpdateData{})
		frees = append(frees, f)
		*u = PlayerUpdateData{Field57: c.Armor}
		v.UpdateData = unsafe.Pointer(u)
		update = armorWearDistributionOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
	} else {
		u, f := alloc.New(MonsterUpdateData{})
		frees = append(frees, f)
		*u = MonsterUpdateData{Field518: c.Armor}
		v.UpdateData = unsafe.Pointer(u)
		update = armorWearDistributionOriginalFrame(unsafe.Pointer(u), unsafe.Sizeof(*u))
	}
	if c.NullUpdate {
		v.UpdateData = nil
	}
	if !c.Empty {
		v.InvFirstItem = item
	}
	frames := [][]byte{update, armorWearDistributionOriginalFrame(unsafe.Pointer(v), unsafe.Sizeof(*v)), armorWearDistributionOriginalFrame(unsafe.Pointer(source), unsafe.Sizeof(*source)), armorWearDistributionOriginalFrame(unsafe.Pointer(effective), unsafe.Sizeof(*effective)), armorWearDistributionOriginalFrame(unsafe.Pointer(item), unsafe.Sizeof(*item)), armorWearDistributionOriginalFrame(unsafe.Pointer(init), unsafe.Sizeof(*init)), armorWearDistributionOriginalFrame(unsafe.Pointer(def), unsafe.Sizeof(*def))}
	before := make([][]byte, len(frames))
	for i, frame := range frames {
		if len(frame) == 0 || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&frame[0])) <= 0xffffffff {
			t.Fatal("expected C-owned native frames above4GiB")
		}
		before[i] = append([]byte(nil), frame...)
	}
	portions := []uint32{}
	lookups := 0
	PlayerDamageItems4E2180(v, source, effective, c.Raw, c.Type, PlayerDamageItemsRuntime4E2180{
		ItemArmorValue: func(got *Object) float64 {
			if got != item {
				t.Fatal("changed coefficient lookup identity")
			}
			lookups++
			// Canonical linked definition lookup and exact binary32 field,
			// matching the original no-slot-zero-modifier getter contexts.
			if d := s.Modif.Nox_xxx_equipClothFindDefByTT413270(int(got.TypeInd)); d != nil {
				return float64(d.DamageCoeffOrArmor64)
			}
			return 0
		},
		EquipDamage: func(got, owner, src, eff *Object, amount float32, typ object.DamageType) {
			if got != item || owner != v || src != source || eff != effective || typ != c.Type {
				t.Fatal("changed native EquipDamage arguments")
			}
			portions = append(portions, math.Float32bits(amount))
		},
	})
	preserved := true
	for i, frame := range frames {
		preserved = preserved && bytes.Equal(before[i], frame)
	}
	match := reflect.DeepEqual(portions, c.Expected) && lookups == len(c.Expected) && preserved
	return map[string]any{"name": c.Name, "portions_bits32": portions, "lookups": lookups, "all_native_nonoutput_frames_preserved": preserved, "matches_original": match}
}

func TestPlayerDamageItemsPrecisionOriginal4E2180Native(t *testing.T) {
	s := New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	b, err := os.ReadFile("testdata/armor_wear_distribution_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []armorWearDistributionOriginalCase `json:"cases"`
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 1518 {
		t.Fatal("incomplete independently executed original armor distribution contexts")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := observeArmorWearDistributionOriginalNative(t, s, c)
			if !got["matches_original"].(bool) {
				t.Fatalf("actualnativearmor distribution mismatch=%v original=%v", got, c.Expected)
			}
		})
	}
}
