package server

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func playerDamageUnarmedPlayerArmorRuntime4E17B0(t *testing.T, target, source, armor *Object, modifier *ModifierEff, typ object.DamageType, events *[]string, damages *[]int32) PlayerDamageRuntime4E17B0 {
	t.Helper()
	r := playerDamageElectricSelfArmorRuntime4E17B0(t, target, source, armor, modifier, typ, events, damages)
	r.ApplyArmorDefend = func(m *ModifierEff, item, victim, weapon, attacker *Object, portion *float32) bool {
		if m != modifier || item != armor || victim != target || weapon != nil || attacker != source || *portion != 5 {
			t.Fatal("armor defense lost nil weapon or original damage")
		}
		*events = append(*events, "armor-defend")
		*portion += 0.5
		return true
	}
	r.DamageArmor = func(item, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
		marker, _, _ := playerDamageElectricSelfMetadata4E17B0(target)
		if item != armor || attacker != source || weapon != nil || damage != 6 || gotType != typ || marker != 0 || *(*float32)(item.UpdateData) != -0.25 {
			t.Fatal("armor callback lost nil weapon, damage, carry or prefix order")
		}
		*events = append(*events, "armor-damage")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	return r
}

func TestPlayerDamageNative4E17B0UnarmedPlayerElectricArmorAndDefault(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			t.Run(fmt.Sprintf("player-target-%t/%s", playerTarget, typ), func(t *testing.T) {
				target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, playerTarget, true)
				var events []string
				var damages []int32
				r := playerDamageUnarmedPlayerArmorRuntime4E17B0(t, target, source, armor, modifier, typ, &events, &damages)
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = func() float32 { events = append(events, "quest"); return 0.5 }
				r.DefaultDamage = func(victim, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
					marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
					if victim != target || attacker != source || weapon != nil || damage != 2 || gotType != typ || marker != 2 || rawType != uint32(typ) || residual != 0 || armor.HealthData.Cur != 94 || target.HealthData.Cur != 60 {
						t.Fatal("electric default entry lost nil weapon or ordered metadata")
					}
					events = append(events, "default")
					return DefaultDamageWorld4E0B30(victim, attacker, weapon, damage, gotType, DefaultDamageWorldRuntime4E0B30{
						Frame: r.Frame, GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
						ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0.25 },
						MonsterHasHitSound: func(*Object) bool { t.Fatal("player read as monster"); return false },
						PlayerSetState:     func(*Object, PlayerState) bool { t.Fatal("small electric hit hurt state"); return false },
						DamageClear:        r.DamageClear, Unsupported: r.Unsupported,
					})
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, nil, 5, typ, r); !handled || !result {
					t.Fatalf("unarmed player electric=%t/%t", handled, result)
				}
				want := []string{"scale", "armor-defend", "armor-damage", "report", "quest", "default", "protection"}
				if !slices.Equal(events, want) || !slices.Equal(damages, []int32{2}) || target.HealthData.Cur != 58 || *carry != -0.25 || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) {
					t.Fatalf("events=%v damage=%v HP=%d armor-carry=%g source=%p", events, damages, target.HealthData.Cur, *carry, target.Obj130)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0UnarmedPlayerElectricReflect(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		for _, front := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/front-%t", typ, front), func(t *testing.T) {
				target, source := defaultDamageElectricSelfFixture4E0B30(t, true, true)
				target.Buffs = 1 << playerDamageReflectEnchant4E17B0
				source.PosVec.X = 37
				var damages []int32
				r := playerDamageRuntime4E17B0(t, nil, &damages)
				r.ElectricArmorScale = func(*Object) float32 { return 1 }
				reflected, calls := false, 0
				r.BlockDirection = func(got *Object, position types.Pointf) bool {
					if typ != object.DamageAirborneElectric || got != target || position != source.PosVec {
						t.Fatal("nil-weapon Reflect facing")
					}
					return front
				}
				r.Audio = func(id int, got *Object) {
					if id != 122 || got != target || reflected {
						t.Fatal("nil-weapon reflection audio")
					}
					reflected = true
				}
				r.DefaultDamage = func(victim, attacker, weapon *Object, damage int32, gotType object.DamageType) bool {
					if victim != target || attacker != source || weapon != nil || damage != 8 || gotType != typ {
						t.Fatal("reflection default arguments")
					}
					calls++
					return true
				}
				wantReflect := front && typ == object.DamageAirborneElectric
				if handled, result := PlayerDamageNative4E17B0(target, source, nil, 8, typ, r); !handled || result == wantReflect || reflected != wantReflect {
					t.Fatalf("reflection=%t/%t audio=%t want=%t", handled, result, reflected, wantReflect)
				}
				marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
				if wantReflect {
					if calls != 0 || marker != 0 || residual != 0 {
						t.Fatal("reflected hit entered electric default")
					}
				} else if calls != 1 || marker != 2 || rawType != uint32(typ) || residual != 0 {
					t.Fatal("unreflected electric attribution")
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0UnarmedPlayerElectricMissingService(t *testing.T) {
	for _, playerTarget := range []bool{false, true} {
		for _, missing := range []string{"scale", "default", "quest", "armor defend", "armor callback"} {
			t.Run(fmt.Sprintf("player-target-%t/%s", playerTarget, missing), func(t *testing.T) {
				target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, playerTarget, true)
				var events []string
				var damages []int32
				r := playerDamageUnarmedPlayerArmorRuntime4E17B0(t, target, source, armor, modifier, object.DamageElectric, &events, &damages)
				r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
					t.Fatal("missing service entered default")
					return false
				}
				reason := ""
				r.Unsupported = func(why string, _, _, weapon *Object, _ int32, _ object.DamageType) {
					if weapon != nil {
						t.Fatal("unsupported report lost nil weapon")
					}
					reason = why
				}
				switch missing {
				case "scale":
					r.ElectricArmorScale = nil
				case "default":
					r.DefaultDamage = nil
				case "quest":
					r.QuestMode, r.QuestDamageScale = func() bool { return true }, nil
				case "armor defend":
					r.ApplyArmorDefend = nil
				case "armor callback":
					r.CanDamageArmor = nil
				}
				if handled, result := PlayerDamageNative4E17B0(target, source, nil, 5, object.DamageElectric, r); handled || result || reason == "" {
					t.Fatalf("missing service=%t/%t reason=%q", handled, result, reason)
				}
				marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
				if marker != 99 || rawType != 77 || residual != math.Float32bits(0.25) || *carry != 0.25 || armor.HealthData.Cur != 100 || target.HealthData.Cur != 60 || target.Obj130 != nil || len(damages) != 0 {
					t.Fatal("missing service mutated target")
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0UnarmedPlayerElectricArmorBeforeGodMode(t *testing.T) {
	target, source, armor, carry, modifier := playerDamageElectricSelfArmorFixture4E17B0(t, true, true)
	var events []string
	var damages []int32
	r := playerDamageUnarmedPlayerArmorRuntime4E17B0(t, target, source, armor, modifier, object.DamageElectric, &events, &damages)
	r.GodMode = func() bool { events = append(events, "god"); return true }
	r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("GodMode entered default")
		return false
	}
	if handled, result := PlayerDamageNative4E17B0(target, source, nil, 5, object.DamageElectric, r); !handled || !result {
		t.Fatalf("GodMode=%t/%t", handled, result)
	}
	marker, rawType, residual := playerDamageElectricSelfMetadata4E17B0(target)
	if !slices.Equal(events, []string{"scale", "armor-defend", "armor-damage", "report", "god"}) || marker != 2 || rawType != 9 || residual != 0 || armor.HealthData.Cur != 94 || *carry != -0.25 || len(damages) != 0 || target.HealthData.Cur != 60 {
		t.Fatalf("events=%v marker=%d/%d HP=%d", events, marker, rawType, target.HealthData.Cur)
	}
}
