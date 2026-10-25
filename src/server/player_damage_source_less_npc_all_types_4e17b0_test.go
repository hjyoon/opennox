package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// These original NPC prefix rows stop at declared wear/DefaultDamage services.
// The separate native callback matrix checks real HP and downstream state.
func TestSourceLessNPCAllTypesMatchesGAMEEXE4E17B0(t *testing.T) {
	data, err := os.ReadFile("testdata/player_damage_source_less_npc_all_types_4e17b0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Type    object.DamageType `json:"damage_type"`
			Raw     int32             `json:"raw"`
			Forward uint32            `json:"forward_damage"`
			Marker  uint32            `json:"marker"`
			Kind    uint32            `json:"marker_type"`
			Carry   uint32            `json:"carry_bits"`
			Wear    []uint32          `json:"wear"`
			Return  int               `json:"return"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 192 {
		t.Fatal("unexpected original all-type source-less fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(fmt.Sprintf("type-%d/raw-%d", row.Type, row.Raw), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .25)
			ud := v.UpdateDataMonster()
			ud.Field547, ud.Field546 = 99, 77
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.CoopMode = func() bool { return false }
			r.QuestMode = func() bool { return false }
			r.GodMode = func() bool { return false }
			r.GameplayFlag1 = func() bool { return true }
			r.FireProtection = func(*Object) float64 { return 0 }
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.BuffOff = func(*Object, EnchantID) {}
			r.PlayerSetState = func(*Object, PlayerState) bool { return true }
			r.BlockDirection, r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = nil, nil, nil
			calls, forward, wear := 0, int32(123456), int32(0)
			r.DefaultDamage = func(target, a, w *Object, d int32, k object.DamageType) bool {
				if target != v || a != nil || w != nil || k != row.Type {
					t.Fatal("source-less DefaultDamage arguments")
				}
				calls++
				forward = d
				return true
			}
			r.DamageClear = func(target *Object, d int32) {
				if target != v {
					t.Fatal("wrong HP target")
				}
				calls++
				forward = d
			}
			r.DamageArmor = func(item, a, w *Object, d int32, k object.DamageType) bool {
				if item != armor || a != nil || w != nil || k != row.Type {
					t.Fatal("source-less armor arguments")
				}
				wear += d
				return true
			}
			reason := ""
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			h, ok := PlayerDamageNative4E17B0(v, nil, nil, row.Raw, row.Type, r)
			wantWear := int32(0)
			if len(row.Wear) != 0 {
				wantWear = max(int32(row.Wear[0]), 0)
			}
			if !h || ok != (row.Return != 0) || calls != 1 || uint32(forward) != row.Forward || ud.Field547 != row.Marker || ud.Field546 != row.Kind || ud.Field1 != row.Carry || wear != wantWear || reason != "" {
				t.Fatalf("handled/result=%t/%t forward=%d/%d calls=%d marker=%d/%d carry=%g/%g wear=%d/%d reason=%s", h, ok, forward, int32(row.Forward), calls, ud.Field547, ud.Field546, math.Float32frombits(ud.Field1), math.Float32frombits(row.Carry), wear, wantWear, reason)
			}
		})
	}
}

func TestSourceLessNPCElectricReloadsCarryAfterScale4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(fmt.Sprint(int32(typ)), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(v, .5, .125)
			cached := v.UpdateDataMonster()
			live := &MonsterUpdateData{Field518: math.Float32bits(.75), Field1: math.Float32bits(.25), Field547: 81, Field546: 83}
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, .5)
			r.ElectricArmorScale = func(*Object) float32 {
				if v.UpdateDataMonster() != cached || cached.Field547 != 0 {
					t.Fatal("electric scale preceded cached marker reset")
				}
				v.UpdateData = unsafe.Pointer(live)
				return 1.5
			}
			wear, forward := int32(0), int32(0)
			r.DamageArmor = func(_, a, w *Object, d int32, k object.DamageType) bool {
				if a != nil || w != nil || k != typ || live.Field1 != math.Float32bits(.75) {
					t.Fatal("electric wear arguments/order")
				}
				wear += d
				return true
			}
			r.DefaultDamage = func(_, a, w *Object, d int32, k object.DamageType) bool {
				if a != nil || w != nil || k != typ {
					t.Fatal("electric DefaultDamage arguments")
				}
				forward = d
				return true
			}
			if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, typ, r); !h || !ok || forward != 13 || wear != 6 ||
				cached.Field547 != 2 || cached.Field546 != uint32(typ) || cached.Field1 != math.Float32bits(.125) ||
				live.Field547 != 81 || live.Field546 != 83 || live.Field1 != math.Float32bits(.75) {
				t.Fatalf("NPC cached/live records=%t/%t forward=%d wear=%d", h, ok, forward, wear)
			}
		})
	}
}

func TestSourceLessNPCRawLateDefendShield4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageDrain, object.DamageDeathMagic} {
		t.Run(fmt.Sprint(int32(typ)), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			v.Buffs = 1 << defaultDamageShieldEnchant4E0B30
			mod := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			init := &ModifierInitData{}
			init.Modifiers[2] = mod
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
			v.InvFirstItem = item // Raw cases have no durability record.
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.ItemArmorValue = func(*Object) float32 { t.Fatal("raw NPC hit read armor"); return 0 }
			world := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			world.CanApplyLateDefend = func(m *ModifierEff) bool { return m == mod }
			world.ApplyLateDefend = func(m *ModifierEff, it, target, w, a *Object, d int32, k object.DamageType) int32 {
				if m != mod || it != item || target != v || w != nil || a != nil || k != typ {
					t.Fatal("NPC raw late Defend arguments")
				}
				events = append(events, "defend")
				return d * 2
			}
			world.ShieldReduce = func(target *Object, d *int32, k object.DamageType, a *Object) {
				if target != v || k != typ || a != nil {
					t.Fatal("NPC raw Shield arguments")
				}
				events = append(events, "shield")
				*d -= 3
			}
			r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
				return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
			}
			if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 9, typ, r); !h || !ok || v.HealthData.Cur != 185 || fmt.Sprint(events) != "[defend shield]" || v.UpdateDataMonster().Field547 != 2 || v.UpdateDataMonster().Field546 != uint32(typ) {
				t.Fatalf("NPC raw effects=%t/%t HP=%d events=%v", h, ok, v.HealthData.Cur, events)
			}
		})
	}
}

func TestSourceLessNPCQuestMinimum4E17B0(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageDrain, object.DamageElectric, object.DamageDeathMagic} {
		t.Run(fmt.Sprint(int32(typ)), func(t *testing.T) {
			v := damageMeleeUnitFixture4E17B0(t, false)
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.ElectricArmorScale = func(*Object) float32 { return 0 }
			r.QuestMode = func() bool { return true }
			r.QuestDamageScale = func() float32 { return 0 }
			r.GodMode = func() bool { return true } // GodMode applies to players.
			forward := int32(0)
			r.DefaultDamage = func(_, a, w *Object, d int32, k object.DamageType) bool {
				if a != nil || w != nil || k != typ {
					t.Fatal("NPC Quest DefaultDamage arguments")
				}
				forward = d
				return true
			}
			if h, ok := PlayerDamageNative4E17B0(v, nil, nil, 1, typ, r); !h || !ok || forward != 1 {
				t.Fatalf("NPC positive minimum=%t/%t forward=%d", h, ok, forward)
			}
		})
	}
}
