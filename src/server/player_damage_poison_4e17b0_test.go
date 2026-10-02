package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestPlayerDamageNative4E17B0MonsterPoisonSkipsEquipment(t *testing.T) {
	for _, tc := range []struct {
		amount int32
		quest  bool
		scale  float32
		want   int32
	}{{-3, false, 0, -3}, {0, false, 0, 0}, {1, false, 0, 1}, {9, false, 0, 9},
		{-3, true, 0.25, -1}, {0, true, 0.25, 0}, {1, true, 0.25, 1}, {9, true, 0.5, 4}} {
		t.Run(fmt.Sprintf("damage-%d/quest-%t", tc.amount, tc.quest), func(t *testing.T) {
			target := defaultDamagePoisonFixture4E0B30(t, 0x11012)
			update := target.UpdateDataMonster()
			update.Field518 = math.Float32bits(0.9)
			update.ArmorEquipFlags = 0x3000000
			update.AIStack[0].Action = 21
			target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			armor, freeArmor := alloc.New(Object{})
			defer freeArmor()
			carry, freeCarry := alloc.New(float32(0))
			defer freeCarry()
			*carry = 0.4
			health, freeHealth := alloc.New(HealthData{})
			defer freeHealth()
			*health = HealthData{Cur: 10, Max: 10}
			*armor = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
				UpdateData: unsafe.Pointer(carry), HealthData: health}
			target.InvFirstItem = armor
			var events []string
			r := PlayerDamageRuntime4E17B0{
				QuestMode: func() bool { return tc.quest },
				QuestDamageScale: func() float32 {
					if update.Field547 != 2 || update.Field546 != 5 {
						t.Fatal("quest scaling preceded the raw type-5 marker")
					}
					events = append(events, "quest")
					return tc.scale
				},
				GodMode:            func() bool { t.Fatal("NPC poison tested player GodMode"); return false },
				ItemArmorValue:     func(*Object) float32 { t.Fatal("poison queried armor"); return 0 },
				ElectricArmorScale: func(*Object) float32 { t.Fatal("poison used electric armor"); return 0 },
				DamageArmor: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("poison damaged armor")
					return false
				},
				DefaultDamage: func(got, source, weapon *Object, damage int32, typ object.DamageType) bool {
					if got != target || source != nil || weapon != nil || damage != tc.want || typ != object.DamagePoison || update.Field547 != 2 || update.Field546 != 5 {
						t.Fatalf("poison default args damage=%d marker=%#x/%d", damage, update.Field546, update.Field547)
					}
					events = append(events, "default")
					return true
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("NPC poison rejected: %s", reason)
				},
			}
			if h, result := PlayerDamageNative4E17B0(target, nil, nil, tc.amount, object.DamagePoison, r); !h || !result {
				t.Fatalf("NPC poison result=%t/%t", h, result)
			}
			wantEvents := []string{"default"}
			if tc.quest {
				wantEvents = []string{"quest", "default"}
			}
			if !slices.Equal(events, wantEvents) || update.Field1 != math.Float32bits(0.25) || update.Field518 != math.Float32bits(0.9) || *carry != 0.4 || armor.HealthData.Cur != 10 || target.HealthData.Cur != 20 {
				t.Fatalf("poison affected armor/carry or order: %v", events)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0MonsterPoisonEarlyGates(t *testing.T) {
	for _, gate := range []string{"non-npc", "no-update", "dead", "invulnerable"} {
		t.Run(gate, func(t *testing.T) {
			target := defaultDamagePoisonFixture4E0B30(t, 0x10)
			update := target.UpdateDataMonster()
			switch gate {
			case "non-npc":
				target.ObjSubClass = 1
			case "no-update":
				target.ObjFlags |= object.FlagNoUpdate
			case "dead":
				target.ObjFlags |= object.FlagDead
			case "invulnerable":
				target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			}
			audio, reason := 0, ""
			r := PlayerDamageRuntime4E17B0{
				Frame: func() uint32 { return 1400 },
				Audio: func(id int, got *Object) {
					if gate != "invulnerable" || id != 71 || got != target {
						t.Fatal("early poison audio")
					}
					audio++
				},
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("early gate reached default")
					return false
				},
				Unsupported: func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s },
			}
			h, result := PlayerDamageNative4E17B0(target, nil, nil, 1, object.DamagePoison, r)
			if h != (gate != "non-npc") || result != (gate == "invulnerable") || audio != map[bool]int{true: 1, false: 0}[gate == "invulnerable"] || update.Field547 != 99 || update.Field546 != 91 {
				t.Fatalf("gate=%s handled/result/audio=%t/%t/%d marker=%d/%d reason=%q", gate, h, result, audio, update.Field546, update.Field547, reason)
			}
			if gate == "non-npc" && reason != "unsupported monster target" {
				t.Fatal(reason)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0MonsterPoisonPreflightAndResult(t *testing.T) {
	for _, missing := range []string{"default", "quest-scale", "none"} {
		t.Run(missing, func(t *testing.T) {
			target := defaultDamagePoisonFixture4E0B30(t, 0x10)
			reason := ""
			r := PlayerDamageRuntime4E17B0{
				QuestMode:   func() bool { return missing == "quest-scale" },
				Unsupported: func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s },
			}
			if missing != "default" {
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool { return false }
			}
			h, result := PlayerDamageNative4E17B0(target, nil, nil, 1, object.DamagePoison, r)
			if result || h != (missing == "none") {
				t.Fatalf("result=%t/%t", h, result)
			}
			if missing != "none" {
				want := "missing default damage service"
				if missing == "quest-scale" {
					want = "missing quest damage service"
				}
				if reason != want || target.UpdateDataMonster().Field547 != 99 || target.UpdateDataMonster().Field546 != 91 {
					t.Fatalf("preflight reason=%q", reason)
				}
			} else if reason != "" {
				t.Fatal(reason)
			}
		})
	}
}
