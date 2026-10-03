package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageNative4E17B0MonsterElectricPlayer(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			update := target.UpdateDataPlayer()
			update.Field21 = math.Float32bits(0.25)
			update.Field57 = math.Float32bits(0.9)
			var damages []int32
			r := playerDamageRuntime4E17B0(t, sound, &damages)
			r.ElectricArmorScale = func(got *Object) float32 {
				if got != target {
					t.Fatal("electric armor target")
				}
				return 0.75
			}
			r.QuestMode = func() bool { return true }
			r.QuestDamageScale = func() float32 { return 0.5 }
			r.DefaultDamage = func(v, a, weapon *Object, damage int32, gotType object.DamageType) bool {
				if v != target || a != source || weapon != nil || damage != 2 || gotType != typ || update.Field76 != 2 || update.Field75 != uint32(typ) {
					t.Fatalf("default entry: damage=%d marker=%#x/%d", damage, update.Field75, update.Field76)
				}
				return DefaultDamageWorld4E0B30(v, a, weapon, damage, gotType, DefaultDamageWorldRuntime4E0B30{
					Frame: r.Frame, GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
					ElectricProtection: func(*Object) float64 { return 0.25 },
					MonsterHasHitSound: func(*Object) bool { return false },
					PlayerSetState:     func(*Object, PlayerState) bool { return true },
					PlayerDamageSoundC: sound, PlayerDamageSound: func(*Object, *Object) {},
					DamageClear: r.DamageClear, Unsupported: r.Unsupported,
				})
			}
			if h, result := PlayerDamageNative4E17B0(target, source, nil, 5, typ, r); !h || !result {
				t.Fatalf("electric player hit = %t/%t", h, result)
			}
			if !reflect.DeepEqual(damages, []int32{2}) || target.HealthData.Cur != 18 || update.Field21 != 0 || update.Field40_0 != 2 || target.Obj130 != source || target.Pos132 != source.PrevPos {
				t.Fatalf("damage=%v HP=%d carry=%g source=%p", damages, target.HealthData.Cur, math.Float32frombits(update.Field21), target.Obj130)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0ElectricFractionalCarry(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	var damages, forwarded []int32
	r := playerDamageRuntime4E17B0(t, sound, &damages)
	r.ElectricArmorScale = func(*Object) float32 { return 0.75 }
	r.DefaultDamage = func(_, _, _ *Object, d int32, _ object.DamageType) bool {
		forwarded = append(forwarded, d)
		return true
	}
	for i, want := range []float32{-0.25, -0.5, 0.25} {
		if h, result := PlayerDamageNative4E17B0(target, source, nil, 5, object.DamageAirborneElectric, r); !h || !result {
			t.Fatalf("electric tick %d = %t/%t", i, h, result)
		}
		if got := math.Float32frombits(target.UpdateDataPlayer().Field21); got != want {
			t.Fatalf("carry tick %d=%g want=%g", i, got, want)
		}
	}
	if !reflect.DeepEqual(forwarded, []int32{4, 4, 3}) || len(damages) != 0 {
		t.Fatalf("forwarded=%v inline=%v", forwarded, damages)
	}
}

func TestPlayerDamageNative4E17B0ElectricArmorBeforeGodMode(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	update := target.UpdateDataPlayer()
	update.Field57 = math.Float32bits(0.4)
	carry := damageArmorCarryFixture4E17B0(0.25)
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, Damage: unsafe.Pointer(new(byte)), HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(&ModifierInitData{})}
	target.InvFirstItem = armor
	var damages []int32
	var events []string
	r := playerDamageRuntime4E17B0(t, sound, &damages)
	r.ElectricArmorScale = func(*Object) float32 { events = append(events, "electric-armor"); return 0.5 }
	r.ItemArmorValue = func(*Object) float32 { return 0.4 }
	r.CanDamageArmor = func(got *Object) bool { return got == armor }
	r.DamageArmor = func(item, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
		if item != armor || attacker != source || weapon != nil || damage != 8 || typ != object.DamageAirborneElectric || update.Field76 != 0 {
			t.Fatal("electric distributes RAW damage to equipped armor")
		}
		events = append(events, "armor")
		item.HealthData.Cur -= uint16(damage)
		return true
	}
	r.ReportArmorHealth = func(owner, item *Object, before, after uint16) {
		if owner != target || item != armor || before != 100 || after != 92 {
			t.Fatal("armor report")
		}
		events = append(events, "report")
	}
	r.GodMode = func() bool { events = append(events, "god"); return true }
	r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
		t.Fatal("GodMode reached default damage")
		return false
	}
	if h, result := PlayerDamageNative4E17B0(target, source, nil, 8, object.DamageAirborneElectric, r); !h || !result {
		t.Fatalf("GodMode electric hit=%t/%t", h, result)
	}
	if !reflect.DeepEqual(events, []string{"electric-armor", "armor", "report", "god"}) || armor.HealthData.Cur != 92 || *carry != 0.25 || update.Field76 != 2 || update.Field75 != 17 || target.HealthData.Cur != 20 || len(damages) != 0 {
		t.Fatalf("events=%v armor=%d carry=%g marker=%#x/%d", events, armor.HealthData.Cur, *carry, update.Field75, update.Field76)
	}
}

func TestPlayerDamageNative4E17B0ElectricMinimumAndReflect(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		typ                   object.DamageType
		reflect, front, quest bool
		wantDamage            int32
		wantAudio             bool
	}{
		{"minimum", object.DamageElectric, false, false, false, 1, false},
		{"quest minimum", object.DamageElectric, false, false, true, 1, false},
		{"ordinary electric ignores Reflect Shield", object.DamageElectric, true, true, false, 1, false},
		{"airborne front reflects", object.DamageAirborneElectric, true, true, false, 0, true},
		{"airborne behind damages", object.DamageAirborneElectric, true, false, false, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			update := target.UpdateDataPlayer()
			update.State = PlayerState16
			update.Player.ArmorEquip = 0x3000000
			if tc.reflect {
				target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			}
			var damages, forwarded []int32
			var audio bool
			r := playerDamageRuntime4E17B0(t, sound, &damages)
			r.ElectricArmorScale = func(*Object) float32 { return 0.25 }
			r.QuestMode = func() bool { return tc.quest }
			r.QuestDamageScale = func() float32 { return 0 }
			r.BlockDirection = func(*Object, types.Pointf) bool { return tc.front }
			r.BlockSourceExcluded = func(*Object) bool { t.Fatal("ordinary shield block handled electric damage"); return false }
			r.Audio = func(id int, got *Object) {
				if id != 122 || got != target || audio {
					t.Fatal("Reflect Shield audio")
				}
				audio = true
			}
			r.DefaultDamage = func(_, _, _ *Object, d int32, _ object.DamageType) bool {
				forwarded = append(forwarded, d)
				return true
			}
			if h, result := PlayerDamageNative4E17B0(target, source, nil, 1, tc.typ, r); !h || result == tc.wantAudio {
				t.Fatalf("electric boundary=%t/%t", h, result)
			}
			if audio != tc.wantAudio || len(damages) != 0 {
				t.Fatalf("audio=%t inline damage=%v", audio, damages)
			}
			if tc.wantAudio {
				if len(forwarded) != 0 || update.Field21 != 0 || update.Field76 != 0 {
					t.Fatal("reflected hit reached electric carry/default")
				}
			} else if !reflect.DeepEqual(forwarded, []int32{tc.wantDamage}) || math.Float32frombits(update.Field21) != 0.25 {
				t.Fatalf("forwarded=%v carry=%g", forwarded, math.Float32frombits(update.Field21))
			}
		})
	}
}

func TestPlayerDamageNative4E17B0ElectricMissingServicesDoesNotMutate(t *testing.T) {
	for _, missing := range []string{"scale", "default", "quest"} {
		t.Run(missing, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			update := target.UpdateDataPlayer()
			update.Field21, update.Field75, update.Field76 = math.Float32bits(0.25), 41, 7
			var damages []int32
			r := playerDamageRuntime4E17B0(t, sound, &damages)
			r.ElectricArmorScale = func(*Object) float32 { return 1 }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("unsupported electric reached default")
				return true
			}
			var reason string
			r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			switch missing {
			case "scale":
				r.ElectricArmorScale = nil
			case "default":
				r.DefaultDamage = nil
			case "quest":
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = nil
			}
			if h, result := PlayerDamageNative4E17B0(target, source, nil, 2, object.DamageAirborneElectric, r); h || result || reason == "" {
				t.Fatalf("unsupported electric=%t/%t reason=%q", h, result, reason)
			}
			if update.Field21 != math.Float32bits(0.25) || update.Field75 != 41 || update.Field76 != 7 || len(damages) != 0 || target.HealthData.Cur != 20 {
				t.Fatal("unsupported electric mutated player")
			}
		})
	}
}
