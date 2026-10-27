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

// The oracle executes unchanged PlayerDamage, carry and integer conversion
// instructions after the original startup sets CW=0xe3f. Armor wear, final HP,
// facing and electric armor scale are declared services, not original Windows
// HP or equipment callbacks. Native nonoutput bytes remain checked separately.
func TestPlayerDamageArmedExplosionPrecisionOriginal4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/armed_explosion_startup_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Startup     uint16 `json:"original_startup_control_word"`
		Rows        []struct {
			Input struct {
				Name        string            `json:"name"`
				Player      bool              `json:"player"`
				Shape       string            `json:"shape"`
				WeaponClass object.Class      `json:"weapon_class"`
				Type        object.DamageType `json:"damage_type"`
				Raw         int32             `json:"raw"`
				Armor       uint32            `json:"absorption_bits32"`
				Carry       uint32            `json:"carry_bits32"`
				Quest       bool              `json:"quest"`
				Scale       uint32            `json:"quest_scale_bits32"`
				Electric    uint32            `json:"electric_scale_bits32"`
			} `json:"input"`
			Events []struct {
				Service string `json:"service"`
				Damage  uint32 `json:"damage_bits32"`
			} `json:"events"`
			Carry    uint32 `json:"carry_bits32"`
			Marker   uint32 `json:"marker"`
			Kind     uint32 `json:"marker_type"`
			Returned uint32 `json:"returned"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Startup != 0xe3f || len(fixture.Rows) != 80 {
		t.Fatal("not the complete original-startup armed EXPLOSION precision fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			input := row.Input
			unit, freeUnit := alloc.New(Object{})
			defer freeUnit()
			info, freeInfo := alloc.New(Player{})
			defer freeInfo()
			*info = Player{}
			var carry, marker, kind *uint32
			var update []byte
			var outputOffsets []uintptr
			if input.Player {
				ud, freeUD := alloc.New(PlayerUpdateData{})
				defer freeUD()
				*ud = PlayerUpdateData{Player: info, Field57: input.Armor, Field21: input.Carry, Field76: 99, Field75: 77}
				*unit = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud)}
				carry, marker, kind = &ud.Field21, &ud.Field76, &ud.Field75
				update = unsafe.Slice((*byte)(unsafe.Pointer(ud)), int(unsafe.Sizeof(*ud)))
				outputOffsets = []uintptr{unsafe.Offsetof(ud.Field21), unsafe.Offsetof(ud.Field76), unsafe.Offsetof(ud.Field75)}
			} else {
				ud, freeUD := alloc.New(MonsterUpdateData{})
				defer freeUD()
				*ud = MonsterUpdateData{Field518: input.Armor, Field1: input.Carry, Field547: 99, Field546: 77}
				*unit = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(ud)}
				carry, marker, kind = &ud.Field1, &ud.Field547, &ud.Field546
				update = unsafe.Slice((*byte)(unsafe.Pointer(ud)), int(unsafe.Sizeof(*ud)))
				outputOffsets = []uintptr{unsafe.Offsetof(ud.Field1), unsafe.Offsetof(ud.Field547), unsafe.Offsetof(ud.Field546)}
			}
			attack, freeAttack := alloc.New(Object{})
			defer freeAttack()
			owner, freeOwner := alloc.New(Object{})
			defer freeOwner()
			ownerUD, freeOwnerUD := alloc.New(PlayerUpdateData{})
			defer freeOwnerUD()
			ownerInfo, freeOwnerInfo := alloc.New(Player{})
			defer freeOwnerInfo()
			attackUD, freeAttackUD := alloc.New(MonsterUpdateData{})
			defer freeAttackUD()
			*owner = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ownerUD)}
			*ownerInfo = Player{PlayerUnit: owner}
			*ownerUD = PlayerUpdateData{Player: ownerInfo}
			*attackUD = MonsterUpdateData{}
			*attack = Object{TypeInd: 73}
			var source, weapon *Object
			source, weapon = owner, attack
			init, freeInit := alloc.New(ModifierInitData{})
			defer freeInit()
			*init = ModifierInitData{}
			attack.ObjClass, attack.InitData = input.WeaponClass, unsafe.Pointer(init)
			switch input.Shape {
			case "armed-weapon-player", "armed-weapon-monster":
				attack.ObjSubClass = 0x100
			case "armed-wand-player", "armed-wand-monster":
				use, freeUse := alloc.New(WandUseData{})
				defer freeUse()
				*use = WandUseData{Flags: 2}
				attack.ObjSubClass, attack.UseData.Ptr = 0x10000, unsafe.Pointer(use)
			default:
				t.Fatal("unknown original armed source shape")
			}
			if input.Shape == "armed-weapon-monster" || input.Shape == "armed-wand-monster" {
				*owner = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(attackUD)}
			}
			if input.WeaponClass != weapon.ObjClass || !playerDamageArmedExplosionShape4E17B0(source, weapon, input.Type) {
				t.Fatal("fixture must use the original armed EXPLOSION admission")
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(attack.CObj()) <= 0xffffffff) {
				t.Fatal("expected real C-owned pointers above 4 GiB")
			}
			unitBefore, attackBefore, ownerBefore, infoBefore := *unit, *attack, *owner, *info
			updateBefore := bytes.Clone(update)
			var got, want []uint32
			for _, event := range row.Events {
				if event.Service == "default" {
					want = append(want, event.Damage)
				}
			}
			r := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool { return false }, GodMode: func() bool { return false },
				QuestMode: func() bool { return input.Quest }, QuestDamageScale: func() float32 { return math.Float32frombits(input.Scale) },
				BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection: func(*Object, types.Pointf) bool { return false },
				ElectricArmorScale: func(target *Object) float32 {
					if target != unit {
						t.Fatal("electric scale changed target")
					}
					return math.Float32frombits(input.Electric)
				},
				DefaultDamage: func(target, a, w *Object, amount int32, typ object.DamageType) bool {
					if target != unit || a != source || w != weapon || typ != input.Type {
						t.Fatal("changed native damage identity or type")
					}
					got = append(got, uint32(amount))
					return true
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatal(reason)
				},
			}
			handled, result := PlayerDamageNative4E17B0(unit, source, weapon, input.Raw, input.Type, r)
			for _, offset := range outputOffsets {
				copy(updateBefore[offset:offset+4], update[offset:offset+4])
			}
			if !handled || result != (row.Returned != 0) || !reflect.DeepEqual(got, want) || *carry != row.Carry || *marker != row.Marker || *kind != row.Kind {
				t.Fatalf("original startup mismatch: handled/result=%t/%t damage=%v/%v carry=%08x/%08x marker=%d/%d kind=%d/%d", handled, result, got, want, *carry, row.Carry, *marker, row.Marker, *kind, row.Kind)
			}
			if *unit != unitBefore || *attack != attackBefore || *owner != ownerBefore || *info != infoBefore || !bytes.Equal(updateBefore, update) {
				t.Fatal("native nonoutput bytes changed")
			}
		})
	}
}
