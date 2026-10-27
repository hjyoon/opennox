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
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func selfWeaponElectricAllocOriginal[T comparable](t *testing.T, value T) *T {
	t.Helper()
	p, free := alloc.New(value)
	*p = value
	t.Cleanup(free)
	return p
}

// Original PlayerDamage/carry/conversion bytes supply expected observations.
// Electric scale, armor wear and final HP are declared service boundaries.
func TestPlayerDamageSelfWeaponElectricPrecisionOriginal4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/self_weapon_electric_startup_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"original_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Name          string            `json:"name"`
				Player        bool              `json:"player"`
				PlayerSource  bool              `json:"player_source"`
				Type          object.DamageType `json:"damage_type"`
				Raw           int32             `json:"raw"`
				LiveMode      int               `json:"live_mode"`
				Carry         uint32            `json:"carry_bits32"`
				LiveCarry     uint32            `json:"live_carry_bits32"`
				CallbackCarry uint32            `json:"callback_carry_bits32"`
				Quest         bool              `json:"quest"`
				QuestScale    uint32            `json:"quest_scale_bits32"`
				ElectricScale uint32            `json:"electric_scale_bits32"`
			} `json:"input"`
			Events []struct {
				Service string `json:"service"`
				Damage  uint32 `json:"damage_bits32"`
			} `json:"events"`
			Carry      uint32 `json:"entry_carry_bits32"`
			LiveCarry  uint32 `json:"live_carry_bits32"`
			Marker     uint32 `json:"marker"`
			Kind       uint32 `json:"marker_type"`
			LiveMarker uint32 `json:"live_marker"`
			LiveKind   uint32 `json:"live_marker_type"`
			Returned   uint32 `json:"returned"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Faults) != 0 || len(fixture.Rows) != 456 {
		t.Fatal("incomplete unchanged-original unit self-weapon electric fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			input := row.Input
			unit := selfWeaponElectricAllocOriginal(t, Object{})
			info := selfWeaponElectricAllocOriginal(t, Player{})
			var entry, live unsafe.Pointer
			var carry, liveCarry, marker, kind, liveMarker, liveKind *uint32
			var entryBytes, liveBytes []byte
			var entryOffsets []uintptr
			var liveOffset uintptr
			if input.Player {
				u := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: info, Field57: math.Float32bits(.5), Field21: input.Carry, Field76: 99, Field75: 77})
				v := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: info, Field57: math.Float32bits(.5), Field21: input.LiveCarry, Field76: 33, Field75: 44})
				*unit = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(u)}
				entry, live = unsafe.Pointer(u), unsafe.Pointer(v)
				carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
				liveCarry, liveMarker, liveKind = &v.Field21, &v.Field76, &v.Field75
				entryBytes, liveBytes = unsafe.Slice((*byte)(entry), int(unsafe.Sizeof(*u))), unsafe.Slice((*byte)(live), int(unsafe.Sizeof(*v)))
				entryOffsets, liveOffset = []uintptr{unsafe.Offsetof(u.Field21), unsafe.Offsetof(u.Field76), unsafe.Offsetof(u.Field75)}, unsafe.Offsetof(v.Field21)
			} else {
				u := selfWeaponElectricAllocOriginal(t, MonsterUpdateData{Field518: math.Float32bits(.5), Field1: input.Carry, Field547: 99, Field546: 77})
				v := selfWeaponElectricAllocOriginal(t, MonsterUpdateData{Field518: math.Float32bits(.5), Field1: input.LiveCarry, Field547: 33, Field546: 44})
				*unit = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(u)}
				entry, live = unsafe.Pointer(u), unsafe.Pointer(v)
				carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
				liveCarry, liveMarker, liveKind = &v.Field1, &v.Field547, &v.Field546
				entryBytes, liveBytes = unsafe.Slice((*byte)(entry), int(unsafe.Sizeof(*u))), unsafe.Slice((*byte)(live), int(unsafe.Sizeof(*v)))
				entryOffsets, liveOffset = []uintptr{unsafe.Offsetof(u.Field1), unsafe.Offsetof(u.Field547), unsafe.Offsetof(u.Field546)}, unsafe.Offsetof(v.Field1)
			}
			source := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012})
			if input.PlayerSource {
				si := selfWeaponElectricAllocOriginal(t, Player{PlayerUnit: source})
				su := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: si})
				source.TypeInd, source.ObjClass, source.ObjSubClass, source.UpdateData = 713, object.ClassPlayer, 0, unsafe.Pointer(su)
			} else {
				source.UpdateData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, MonsterUpdateData{}))
			}
			if !playerDamageElectricShape4E17B0(source, source, input.Type) || source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile) {
				t.Fatal("fixture must select unit self-weapon electric")
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(entry) <= 0xffffffff || uintptr(live) <= 0xffffffff || uintptr(source.CObj()) <= 0xffffffff) {
				t.Fatal("expected native-width C-owned pointers above 4 GiB")
			}
			unitBefore, sourceBefore, infoBefore := *unit, *source, *info
			entryBefore, liveBefore := bytes.Clone(entryBytes), bytes.Clone(liveBytes)
			var got, want []uint32
			for _, event := range row.Events {
				if event.Service == "default" {
					want = append(want, event.Damage)
				}
			}
			scaleCalls := 0
			r := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool { return false }, GodMode: func() bool { return false },
				QuestMode: func() bool { return input.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(input.QuestScale) },
				BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection: func(*Object, types.Pointf) bool { return false },
				ElectricArmorScale: func(target *Object) float32 {
					if target != unit || *marker != 0 {
						t.Fatal("electric scale identity or cached prefix order changed")
					}
					scaleCalls++
					switch input.LiveMode {
					case 1:
						unit.UpdateData = live
					case 2:
						*carry = input.CallbackCarry
					}
					return math.Float32frombits(input.ElectricScale)
				},
				DefaultDamage: func(target, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
					if target != unit || attacker != source || weapon != source || typ != input.Type {
						t.Fatal("changed original self-weapon identity or type")
					}
					got = append(got, uint32(damage))
					return true
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			handled, result := PlayerDamageNative4E17B0(unit, source, source, input.Raw, input.Type, r)
			if input.LiveMode == 1 {
				unitBefore.UpdateData = live
			}
			for _, offset := range entryOffsets {
				copy(entryBefore[offset:offset+4], entryBytes[offset:offset+4])
			}
			copy(liveBefore[liveOffset:liveOffset+4], liveBytes[liveOffset:liveOffset+4])
			if !handled || result != (row.Returned != 0) || scaleCalls != 1 || !reflect.DeepEqual(got, want) || *carry != row.Carry || *liveCarry != row.LiveCarry || *marker != row.Marker || *kind != row.Kind || *liveMarker != row.LiveMarker || *liveKind != row.LiveKind {
				t.Fatalf("original startup mismatch: handled/result=%t/%t damage=%v/%v carry=%08x/%08x live=%08x/%08x marker=%d/%d kind=%d/%d", handled, result, got, want, *carry, row.Carry, *liveCarry, row.LiveCarry, *marker, row.Marker, *kind, row.Kind)
			}
			if *unit != unitBefore || *source != sourceBefore || *info != infoBefore || !bytes.Equal(entryBefore, entryBytes) || !bytes.Equal(liveBefore, liveBytes) {
				t.Fatal("native nonoutput bytes changed")
			}
		})
	}
}
