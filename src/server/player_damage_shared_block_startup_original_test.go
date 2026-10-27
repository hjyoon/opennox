package server

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func sharedBlockOriginalWearBits4E17B0(t *testing.T, percent float64, raw int32) uint32 {
	t.Helper()
	blob, err := os.ReadFile("testdata/shared_block_compatibility_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA  string `json:"original_sha256"`
		CW   uint16 `json:"original_startup_control_word"`
		Rows []struct {
			Percent float64 `json:"percent"`
			Raw     int32   `json:"raw"`
			Wear    uint32  `json:"wear_bits32"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.CW != 0xe3f {
		t.Fatal("not original startup block observations")
	}
	for _, row := range fixture.Rows {
		if row.Percent == percent && row.Raw == raw {
			return row.Wear
		}
	}
	t.Fatalf("missing original block amount percent=%g raw=%d", percent, raw)
	return 0
}

// Unchanged original PlayerDamage and its real inventory/readiness loops
// supply expected block wear. Final item wear and HP are declared services.
func TestPlayerDamageSharedBlockStartupOriginal4E17B0(t *testing.T) {
	blob, err := os.ReadFile("testdata/shared_block_startup_original.json")
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
				Name        string            `json:"name"`
				Profile     string            `json:"profile"`
				SourceClass object.Class      `json:"source_class"`
				WeaponClass object.Class      `json:"weapon_class"`
				Subclass    object.SubClass   `json:"weapon_subclass"`
				Type        object.DamageType `json:"type"`
				Form        string            `json:"form"`
				Player      bool              `json:"player"`
				Defense     string            `json:"defense"`
				Percent     float64           `json:"percent"`
				Raw         int32             `json:"raw"`
			} `json:"input"`
			Wear         uint32 `json:"wear_bits32"`
			Calls        int    `json:"wear_calls"`
			Service      string `json:"wear_service"`
			WeaponSource bool   `json:"wear_weapon_equals_source"`
			WeaponInput  bool   `json:"wear_weapon_equals_input_weapon"`
			DefaultCalls int    `json:"default_calls"`
			Returned     uint32 `json:"returned"`
			Marker       uint32 `json:"marker"`
			Kind         uint32 `json:"marker_type"`
			Carry        uint32 `json:"carry_bits32"`
			State        uint32 `json:"state"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Faults) != 0 || len(fixture.Rows) != 1260 {
		t.Fatal("incomplete original shared block observations")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			if row.Calls != 1 || row.DefaultCalls != 0 || row.Returned != 0 {
				t.Fatal("fixture must be an original blocking path")
			}
			if (in.Profile == "caster-impact" || in.Profile == "spell-missile-impact" || in.Profile == "monster-self-impact" || in.Profile == "player-self-impact") && in.Type != object.DamageImpact {
				t.Fatal("IMPACT fixture used a different original type")
			}
			makeUnit := func(class object.Class) *Object {
				u := selfWeaponElectricAllocOriginal(t, Object{TypeInd: 713, ObjClass: class})
				if class.Has(object.ClassPlayer) {
					info := selfWeaponElectricAllocOriginal(t, Player{PlayerUnit: u})
					ud := selfWeaponElectricAllocOriginal(t, PlayerUpdateData{Player: info})
					u.UpdateData = unsafe.Pointer(ud)
				} else if class.Has(object.ClassMonster) {
					u.TypeInd, u.ObjSubClass = 1343, 0x11012
					u.UpdateData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, MonsterUpdateData{}))
				}
				return u
			}
			class := object.ClassMonster
			if in.Player {
				class = object.ClassPlayer
			}
			target, source := makeUnit(class), makeUnit(in.SourceClass)
			target.HealthData = selfWeaponElectricAllocOriginal(t, HealthData{Cur: 200, Max: 200})
			var weapon *Object
			switch in.Form {
			case "self":
				weapon, source.ObjSubClass = source, in.Subclass
			case "distinct":
				weapon = makeUnit(in.WeaponClass)
				weapon.TypeInd, weapon.ObjSubClass, weapon.ObjOwner = 73, in.Subclass, source
				if weapon.Class().Has(object.ClassWeapon) {
					weapon.InitData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, ModifierInitData{}))
				}
			case "nil":
			default:
				t.Fatal("unknown original source form")
			}
			attack := weapon
			if attack == nil {
				attack = source
			}
			attack.PosVec, attack.PrevPos = types.Ptf(19, 11), types.Ptf(17, 9)
			armorFlags, weaponFlags, subclass := uint32(0), uint32(0x400), object.SubClass(0x400)
			itemClass := object.ClassWeapon
			if in.Defense == "shield" {
				armorFlags, weaponFlags, subclass, itemClass = 0x1000000, 0, 2, object.ClassArmor
			} else if in.Defense == "staff" {
				weaponFlags, subclass = 0x8000, 0x8000
			}
			var carry, marker, kind *uint32
			if in.Player {
				u := target.UpdateDataPlayer()
				u.Field21, u.Field57, u.Field76, u.Field75 = math.Float32bits(.25), math.Float32bits(.5), 99, 77
				u.Player.ArmorEquip, u.Player.WeaponEquip = armorFlags, weaponFlags
				u.State = PlayerState13
				if in.Defense == "shield" {
					u.State = PlayerState16
				}
				carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
			} else {
				u := target.UpdateDataMonster()
				u.Field1, u.Field518, u.Field547, u.Field546 = math.Float32bits(.25), math.Float32bits(.5), 99, 77
				u.ArmorEquipFlags, u.WeaponEquipFlags = armorFlags, weaponFlags
				if in.Defense == "shield" {
					u.AIStack[0].Action = 21
				}
				carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
			}
			item := selfWeaponElectricAllocOriginal(t, Object{ObjClass: itemClass, ObjSubClass: subclass, ObjFlags: object.FlagEquipped})
			item.HealthData = selfWeaponElectricAllocOriginal(t, HealthData{Cur: 100, Max: 100})
			item.InitData = unsafe.Pointer(selfWeaponElectricAllocOriginal(t, ModifierInitData{}))
			target.InvFirstItem = item
			if source.Class() != in.SourceClass || (weapon != nil && weapon.Class() != in.WeaponClass) {
				t.Fatal("source/weapon original class changed")
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32 || uintptr(item.CObj()) <= math.MaxUint32) {
				t.Fatal("expected native-width C-owned records above 4 GiB")
			}
			calls, percentCalls := 0, 0
			wear := func(service string) func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool {
				return func(got, owner, attacker, effective *Object, amount float32, typ object.DamageType) bool {
					if service != row.Service || got != item || owner != target || attacker != source || typ != in.Type || (effective == source) != row.WeaponSource || (effective == weapon) != row.WeaponInput {
						t.Errorf("original wear identity: service=%s/%s effective source/input=%t/%t want=%t/%t", service, row.Service, effective == source, effective == weapon, row.WeaponSource, row.WeaponInput)
					}
					if math.Float32bits(amount) != row.Wear {
						t.Errorf("original block wear bits=%08x want=%08x", math.Float32bits(amount), row.Wear)
					}
					calls++
					return true
				}
			}
			r := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool { return false }, GodMode: func() bool { return false }, QuestMode: func() bool { return false },
				BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection:     func(*Object, types.Pointf) bool { return true },
				BlockDamagePercent: func() float64 { percentCalls++; return in.Percent },
				ProjectileReflect:  func(*Object, *Object) {}, ClearOwner: func(v *Object) { v.ObjOwner = nil },
				SetOwner:           func(owner, v *Object) { v.ObjOwner = owner },
				Audio:              func(int, *Object) {},
				PlayerSetState:     func(v *Object, state PlayerState) bool { v.UpdateDataPlayer().State = state; return true },
				CanDamageBlockItem: func(v *Object) bool { return v == item }, DamageBlockItem: wear("0x4e16d0"),
				ItemArmorValue: func(*Object) float32 { return .5 }, CanDamageArmor: func(*Object) bool { return true },
				DamageArmor: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("blocked hit reached armor tail")
					return false
				},
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("blocked hit reached default HP")
					return false
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			r.Melee = PlayerDamageMeleeRuntime4E17B0{
				RandomInt: func(lo, hi int) int {
					if lo != 18 || hi != 20 {
						t.Fatal("original block RNG bounds")
					}
					return 19
				},
				MonsterBlockAction:    func(v *Object) { v.UpdateDataMonster().AIStack[0].Action = 23 },
				MonsterPopBlockAction: func(*Object) { t.Fatal("intact original block item popped action") },
				CanDamageBlockWeapon:  func(v *Object) bool { return v == item }, DamageBlockWeapon: wear("0x4e1560"),
			}
			h, result := PlayerDamageNative4E17B0(target, source, weapon, in.Raw, in.Type, r)
			var state uint32
			if in.Player {
				state = uint32(target.UpdateDataPlayer().State)
			} else {
				state = target.UpdateDataMonster().AIStack[0].Action
			}
			if !h || result || calls != 1 || percentCalls != 1 || *marker != row.Marker || *kind != row.Kind || *carry != row.Carry || state != row.State || target.HealthData.Cur != 200 || item.HealthData.Cur != 100 {
				t.Fatalf("original block prefix h/result=%t/%t calls=%d percent=%d marker=%d/%d want=%d/%d carry=%08x/%08x state=%d/%d", h, result, calls, percentCalls, *marker, *kind, row.Marker, row.Kind, *carry, row.Carry, state, row.State)
			}
		})
	}
}
