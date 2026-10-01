package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDefaultDamageWorld4E0B30MonsterElectricPlayer(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			target, source, _ := playerDamageFixture4E17B0(t)
			target.DamageSound = nil
			update := target.UpdateDataPlayer()
			update.Field40_0, update.Field40_1 = 0x1122, 0xabcd
			update.Field76, update.Field75 = 2, math.Float32bits(float32(typ))
			var events []string
			r := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 700 },
				GameplayFlag1: func() bool { return true },
				IsEnemy:       func(a, b *Object) bool { return a == target && b == source },
				ElectricProtection: func(got *Object) float64 {
					if got != target {
						t.Fatal("electric protection target")
					}
					events = append(events, "protection")
					return 0.25
				},
				Audio: func(id int, got *Object) {
					if id != 108 || got != target {
						t.Fatal("protection sound")
					}
					events = append(events, "protection-sound")
				},
				BuffOff: func(got *Object, enchant EnchantID) {
					if got != target || enchant != 0 || update.Field40_0 != 2 {
						t.Fatal("invisibility/electric state order")
					}
					events = append(events, "buff-off")
				},
				MonsterHasHitSound: func(got *Object) bool {
					if got != source {
						t.Fatal("monster sound lookup")
					}
					events = append(events, "hit-sound")
					return false
				},
				DefaultDamageSound: func(a, b *Object) {
					if a != target || b != source || a.Obj130 != source {
						t.Fatal("default sound/attribution order")
					}
					events = append(events, "damage-sound")
				},
				PlayerSetState: func(*Object, PlayerState) bool { t.Fatal("small hit caused hurt state"); return false },
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 6 || source.UpdateDataMonster().Field130 != 700 {
						t.Fatal("damage/monster attack latch")
					}
					events = append(events, "damage")
					got.HealthData.Cur -= uint16(damage)
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("electric player hit rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, typ, r) {
				t.Fatal("electric player hit returned false")
			}
			want := []string{"protection", "protection-sound", "buff-off", "hit-sound", "damage-sound", "damage"}
			if !reflect.DeepEqual(events, want) || target.HealthData.Cur != 14 ||
				target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 700 ||
				update.Field40_0 != 2 || update.Field40_1 != 0xabcd || update.Field76 != 2 || update.Field75 != math.Float32bits(float32(typ)) {
				t.Fatalf("events=%v HP=%d source=%p position=%+v electric=%#x/%#x marker=%#x/%d", events, target.HealthData.Cur, target.Obj130, target.Pos132, update.Field40_0, update.Field40_1, update.Field75, update.Field76)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ElectricPlayerDefenseTail(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	target.HealthData.Cur = 2000
	oldUpdate := target.UpdateDataPlayer()
	newUpdate := &PlayerUpdateData{Player: oldUpdate.Player, State: PlayerState13, Field40_0: 0x1111}
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier, nil}})}
	target.InvFirstItem = armor
	target.Field129 = &Object{TypeInd: 77, ObjOwner: target}
	target.Buffs = 1 << defaultDamageShieldEnchant4E0B30
	var events []string
	r := DefaultDamageWorldRuntime4E0B30{
		Frame: func() uint32 { return 700 }, GameplayFlag1: func() bool { return true },
		IsEnemy:            func(*Object, *Object) bool { return true },
		ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0.25 },
		BuffOff:            func(*Object, EnchantID) { events = append(events, "buff-off") },
		CanApplyLateDefend: func(got *ModifierEff) bool { return got == modifier },
		ApplyLateDefend: func(m *ModifierEff, item, victim, weapon, attacker *Object, damage int32, typ object.DamageType) int32 {
			if m != modifier || item != armor || victim != target || weapon != nil || attacker != source || damage != 30 || typ != object.DamageAirborneElectric || oldUpdate.Field40_0 != 2 {
				t.Fatal("late defend order/arguments")
			}
			events = append(events, "late-defend")
			target.UpdateData = unsafe.Pointer(newUpdate)
			return damage + 2
		},
		MonsterHasHitSound: func(*Object) bool { events = append(events, "hit-sound"); return false },
		PlayerDamageSoundC: sound,
		PlayerDamageSound: func(a, b *Object) {
			if a != target || b != source || a.Obj130 != source {
				t.Fatal("native player damage sound")
			}
			events = append(events, "player-sound")
		},
		GameBallType: 77,
		GameBallOnDamage: func(a, v *Object, damage int32) {
			if a != source || v != target || damage != 32 || v.HealthData.Cur != 2000 {
				t.Fatal("GameBall before Shield")
			}
			events = append(events, "ball")
		},
		PlayerSetState: func(got *Object, state PlayerState) bool {
			if got != target || state != PlayerState30 || got.UpdateDataPlayer() != newUpdate {
				t.Fatal("hurt state/live update reload")
			}
			events = append(events, "hurt")
			newUpdate.State = state
			return true
		},
		ShieldReduce: func(v *Object, damage *int32, typ object.DamageType, attacker *Object) {
			if v != target || *damage != 32 || typ != object.DamageAirborneElectric || attacker != source || newUpdate.State != PlayerState30 || source.UpdateDataMonster().Field130 != 700 {
				t.Fatal("Shield tail order/arguments")
			}
			events = append(events, "shield")
			*damage = 0
		},
		DamageClear: func(*Object, int32) { t.Fatal("Shield absorbed hit reached HP") },
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("defense tail rejected: %s", reason)
		},
	}
	if DefaultDamageWorld4E0B30(target, source, nil, 40, object.DamageAirborneElectric, r) {
		t.Fatal("absorbed electric hit returned true")
	}
	want := []string{"protection", "buff-off", "late-defend", "hit-sound", "player-sound", "ball", "hurt", "shield"}
	if !reflect.DeepEqual(events, want) || target.HealthData.Cur != 2000 || newUpdate.Field40_0 != 0x1111 {
		t.Fatalf("events=%v HP=%d new electric word=%#x", events, target.HealthData.Cur, newUpdate.Field40_0)
	}
}

func TestDefaultDamageWorld4E0B30ElectricPlayerRoundingAndHurt(t *testing.T) {
	for _, tc := range []struct {
		name           string
		damage, want   int32
		protection     float64
		state          PlayerState
		hurt, hitSound bool
	}{
		{"fully protected minimum", 8, 1, 1, PlayerState13, false, false},
		{"ties to even", 5, 2, 0.5, PlayerState13, false, false},
		{"below hurt threshold", 30, 18, 0.4, PlayerState13, false, false},
		{"at hurt threshold", 40, 20, 0.5, PlayerState13, true, false},
		{"attack state excluded", 40, 40, 0, PlayerState1, false, false},
		{"special state excluded", 40, 40, 0, PlayerState15, false, false},
		{"monster hit sound suppresses", 8, 8, 0, PlayerState13, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			target.UpdateDataPlayer().State = tc.state
			hurt, played, damageCalls := false, false, 0
			r := DefaultDamageWorldRuntime4E0B30{
				GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return true },
				ElectricProtection: func(*Object) float64 { return tc.protection },
				MonsterHasHitSound: func(*Object) bool { return tc.hitSound },
				PlayerDamageSoundC: sound, PlayerDamageSound: func(*Object, *Object) { played = true },
				PlayerSetState: func(*Object, PlayerState) bool { hurt = true; return true },
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != tc.want {
						t.Fatalf("damage=%d want=%d", damage, tc.want)
					}
					damageCalls++
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("rounding/hurt rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, tc.damage, object.DamageAirborneElectric, r) || hurt != tc.hurt || played == tc.hitSound || damageCalls != 1 {
				t.Fatalf("hurt=%t played=%t damage calls=%d", hurt, played, damageCalls)
			}
		})
	}
}
