package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func damageWeaponlessExplosionSource4E17B0(t *testing.T, target *Object, kind string) *Object {
	t.Helper()
	switch kind {
	case "none":
		return nil
	case "self":
		return target
	default:
		return damageMeleeUnitFixture4E17B0(t, kind == "player")
	}
}

func TestPlayerDamageNative4E17B0NPCWeaponlessExplosionArmor(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, immune := range []bool{false, true} {
			for _, protected := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/immune-%t/protected-%t", kind, immune, protected), func(t *testing.T) {
					target := damageMeleeUnitFixture4E17B0(t, false)
					if immune {
						target.ObjSubClass |= 0x400
					}
					source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
					armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
					protection, want := float64(0), int32(5)
					if immune {
						want = 2
					}
					if protected {
						protection = 0.5
						want = int32(math.RoundToEven(float64(want) * 0.5))
					}
					r := damageFlameRuntime4E17B0(t, protection)
					damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
					r.GodMode = func() bool { t.Fatal("NPC queried player GodMode"); return true }
					if h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r); !h || !result {
						t.Fatalf("handled=%t result=%t", h, result)
					}
					marker, typ, carry := damageMeleeMarker4E17B0(target)
					position := types.Pointf{}
					if source != nil {
						position = source.PrevPos
					}
					if target.HealthData.Cur != uint16(200-want) || armor.HealthData.Cur != 21 || marker != 2 || typ != 7 ||
						math.Abs(float64(carry)+0.1) > 1e-6 || target.Obj130 != source || target.Pos132 != position ||
						target.Field131 != 7 || target.Frame134 != 1400 ||
						!target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) {
						t.Fatalf("HP=%d want=%d armor=%d marker=%d/%d carry=%g source=%p", target.HealthData.Cur, 200-want,
							armor.HealthData.Cur, marker, typ, carry, target.Obj130)
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0NPCWeaponlessExplosionRounding(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, tc := range []struct {
			damage, effective int32
			armor, carry      float32
			wantCarry         uint32
		}{
			{9, 4, 0.5, 0, math.Float32bits(0.5)}, {11, 6, 0.5, 0, math.Float32bits(-0.5)},
			{-9, -4, 0.5, 0, math.Float32bits(-0.5)}, {1, 1, 1, 0, 0}, {0, 0, 0.5, 0, 0},
			{3, 3, 0.1, 0, 0xbe999998}, {9, 7, 0.2, 0, 0x3e4cccc0}, {3, 2, 0.4, 0, 0xbe4cccd0},
		} {
			t.Run(fmt.Sprintf("%s/%d/%g", kind, tc.damage, tc.armor), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, false)
				source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
				armor := damageMeleeArmorFixture4E17B0(target, tc.armor, tc.carry)
				r := damageFlameRuntime4E17B0(t, 0)
				damageMeleeArmorRuntime4E17B0(&r, armor, tc.armor)
				got := int32(math.MaxInt32)
				r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
					if v != target || s != source || w != nil || typ != object.DamageExplosion {
						t.Fatal("weapon-less explosion identity changed")
					}
					got = amount
					return true
				}
				if h, result := PlayerDamageNative4E17B0(target, source, nil, tc.damage, object.DamageExplosion, r); !h || !result {
					t.Fatalf("handled=%t result=%t", h, result)
				}
				marker, typ, carry := damageMeleeMarker4E17B0(target)
				if got != tc.effective || marker != 2 || typ != 7 || math.Float32bits(carry) != tc.wantCarry {
					t.Fatalf("damage=%d marker=%d/%d carry=%#x", got, marker, typ, math.Float32bits(carry))
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0NPCWeaponlessExplosionCachedLiveOrder(t *testing.T) {
	target := damageMeleeUnitFixture4E17B0(t, false)
	source := damageMeleeUnitFixture4E17B0(t, true)
	armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
	cached := target.UpdateDataMonster()
	cached.Field547, cached.Field546 = 99, 77
	live := &MonsterUpdateData{Field518: math.Float32bits(0.25), Field1: math.Float32bits(0.25)}
	r := damageFlameRuntime4E17B0(t, 0)
	damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
	var events []string
	position := source.PrevPos
	r.BlockSourceExcluded = func(*Object) bool { t.Fatal("a3 == nil used six-type exclusions"); return true }
	r.BlockSourceOnlyExcluded = func(v *Object) bool {
		if v != source || cached.Field547 != 0 {
			t.Fatal("exclusion preceded cached marker clear")
		}
		events = append(events, "exclude")
		source.PrevPos = types.Pointf{X: 999}
		target.UpdateData = unsafe.Pointer(live)
		return false
	}
	r.BlockDirection = func(v *Object, pos types.Pointf) bool {
		if v != target || pos != position || cached.Field547 != 0 {
			t.Fatal("facing lost pre-exclusion position")
		}
		events = append(events, "facing")
		cached.Field547, cached.Field546 = 77, 88
		return false
	}
	quest := false
	r.DamageArmor = func(item, s, w *Object, amount int32, typ object.DamageType) bool {
		if item != armor || s != source || w != nil || amount != 4 || typ != object.DamageExplosion ||
			cached.Field547 != 77 || cached.Field546 != 88 || cached.Field1 != math.Float32bits(0.4) || live.Field1 != math.Float32bits(-0.25) {
			t.Fatal("cached absorption/marker or live carry/wear order changed")
		}
		events = append(events, "armor")
		cached.Field547 = 0
		quest = true
		return true
	}
	r.QuestMode = func() bool {
		if !quest || cached.Field547 != 2 || cached.Field546 != 7 {
			t.Fatal("Quest query preceded armor/marker fallback")
		}
		events = append(events, "quest")
		return quest
	}
	r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
	r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
		if v != target || s != source || w != nil || amount != 2 || typ != object.DamageExplosion {
			t.Fatal("late Quest/default identity")
		}
		events = append(events, "default")
		return false
	}
	h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
	if !h || result || target.HealthData.Cur != 200 || !slices.Equal(events, []string{"exclude", "facing", "armor", "quest", "scale", "default"}) {
		t.Fatalf("handled=%t result=%t events=%v", h, result, events)
	}
}

func TestPlayerDamageNative4E17B0NPCWeaponlessExplosionShield(t *testing.T) {
	for _, reflected := range []string{"unit", "missile-keeps-owner", "missile-transfers-owner"} {
		t.Run(reflected, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, false)
			source := damageMeleeUnitFixture4E17B0(t, true)
			cached := target.UpdateDataMonster()
			cached.ArmorEquipFlags, cached.Field547 = 0x3000000, 99
			cached.AIStackInd = 0
			cached.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
			r := damageFlameRuntime4E17B0(t, 0)
			var events []string
			r.BlockSourceOnlyExcluded = func(v *Object) bool {
				if v != source || cached.Field547 != 0 {
					t.Fatal("shield exclusion preceded cached marker reset")
				}
				events = append(events, "exclude")
				cached.ArmorEquipFlags = 0 // Original equipment was cached before this callback.
				return false
			}
			r.BlockDirection = func(v *Object, pos types.Pointf) bool { events = append(events, "facing"); return true }
			r.Audio = func(id int, v *Object) {
				if id != 878 || v != target || cached.Field547 != 0 {
					t.Fatal("shield audio/marker changed")
				}
				events = append(events, "audio")
				if reflected != "unit" {
					source.ObjClass = object.ClassMissile
					if reflected == "missile-keeps-owner" {
						source.ObjSubClass = 2
					}
				}
			}
			r.ProjectileReflect = func(v, by *Object) {
				if v != source || by != target {
					t.Fatal("reflection identity")
				}
				events = append(events, "reflect")
			}
			r.ClearOwner = func(v *Object) {
				if v != source {
					t.Fatal("clear identity")
				}
				events = append(events, "clear")
			}
			r.SetOwner = func(by, v *Object) {
				if by != target || v != source {
					t.Fatal("owner identity")
				}
				events = append(events, "owner")
			}
			r.BlockDamagePercent = func() float64 { events = append(events, "balance"); target.InvFirstItem = shield; return 0.25 }
			r.CanDamageBlockItem = func(v *Object) bool { return v == shield }
			r.DamageBlockItem = func(v, by, s, w *Object, amount float32, typ object.DamageType) bool {
				if v != shield || by != target || s != source || w != nil || amount != 2.25 || typ != object.DamageExplosion {
					t.Fatal("live shield wear identity/order")
				}
				events = append(events, "wear")
				return true
			}
			r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("blocked explosion reached health")
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			want := []string{"exclude", "facing", "audio"}
			if reflected != "unit" {
				want = append(want, "reflect")
			}
			if reflected == "missile-transfers-owner" {
				want = append(want, "clear", "owner")
			}
			want = append(want, "balance", "wear")
			if !h || result || target.HealthData.Cur != 200 || cached.Field547 != 0 || !slices.Equal(events, want) {
				t.Fatalf("handled=%t result=%t marker=%d events=%v want=%v", h, result, cached.Field547, events, want)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0NPCWeaponlessExplosionAdmission(t *testing.T) {
	for _, missing := range []string{"default", "armor", "quest", "exclude", "direction"} {
		t.Run(missing, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, false)
			source := damageMeleeUnitFixture4E17B0(t, true)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			r := damageFlameRuntime4E17B0(t, 0)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			switch missing {
			case "default":
				r.DefaultDamage = nil
			case "armor":
				r.DamageArmor = nil
			case "quest":
				r.QuestMode = func() bool { return true }
			case "exclude":
				r.BlockSourceOnlyExcluded = nil
			case "direction":
				r.BlockDirection = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			if h || result || reason == "" || target.HealthData.Cur != 200 || armor.HealthData.Cur != 25 || target.UpdateDataMonster().Field1 != math.Float32bits(0.4) {
				t.Fatalf("failed admission wrote health/carry: handled=%t result=%t reason=%s", h, result, reason)
			}
		})
	}
}
