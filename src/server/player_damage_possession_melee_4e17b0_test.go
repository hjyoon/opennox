package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type possessionMeleeAttack4E17B0 struct {
	name     string
	class    object.Class
	subclass object.SubClass
	typ      object.DamageType
}

func possessionMeleeCases4E17B0(t *testing.T, fn func(*testing.T, bool, possessionMeleeAttack4E17B0)) {
	t.Helper()
	for _, playerSource := range []bool{false, true} {
		for _, attack := range []possessionMeleeAttack4E17B0{
			{"Sword", object.ClassWeapon, object.SubClass(object.WeaponSword), object.DamageBlade},
			{"MorningStar", object.ClassWeapon, object.SubClass(object.WeaponMace), object.DamageCrush},
			{"WarHammer", object.ClassWeapon, object.SubClass(object.WeaponHammer), object.DamageCrush},
			{"WoodenStaff", object.ClassWand, 0, object.DamageBlade},
			{"UnarmedClaw", 0, 0, object.DamageClaw},
			{"UnarmedCrush", 0, 0, object.DamageCrush},
			{"SimpleFist", object.ClassSimple, 0, object.DamageCrush},
		} {
			t.Run(fmt.Sprintf("player-source-%t/%s", playerSource, attack.name), func(t *testing.T) { fn(t, playerSource, attack) })
		}
	}
}

func possessionMeleeFixture4E17B0(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) (target, source, weapon, attack *Object) {
	t.Helper()
	target, source = damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, playerSource)
	source.TypeInd = 88
	if kind.class != 0 {
		weapon = &Object{TypeInd: 777, ObjClass: kind.class, ObjSubClass: kind.subclass, InitData: unsafe.Pointer(&ModifierInitData{}), PrevPos: types.Ptf(20, 0)}
	}
	attack = weapon
	if attack == nil {
		attack = source
	}
	ud := target.UpdateDataPlayer()
	ud.Player.Field3680, ud.Player.CameraFollowObj = 2, source
	ud.Field57, ud.Field21 = math.Float32bits(0.25), math.Float32bits(0.125)
	ud.Field76, ud.Field75 = 88, 77
	return
}

func TestPlayerDamagePossessionMelee4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		defenses := []string{"none", "Reflect Shield", "shield", "excluded shield", "rear shield"}
		if kind.typ == object.DamageBlade {
			defenses = append(defenses, "GreatSword", "staff")
		}
		for _, defense := range defenses {
			t.Run(defense, func(t *testing.T) {
				target, source, weapon, attack := possessionMeleeFixture4E17B0(t, playerSource, kind)
				cached := target.UpdateDataPlayer()
				if defense == "shield" || defense == "excluded shield" || defense == "rear shield" {
					cached.Player.ArmorEquip = 0x1000000
				} else if defense == "GreatSword" {
					cached.Player.WeaponEquip = 0x400
				} else if defense == "staff" {
					cached.Player.WeaponEquip = 0x8000
				}
				live := &PlayerUpdateData{Player: &Player{Field3680: 3, CameraFollowObj: attack}, State: PlayerState30, Field57: math.Float32bits(0.9), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				var events []string
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.ObserveClear = func(v *Object) {
					if v != target || target.UpdateDataPlayer() != cached || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("ObserveClear must follow only the cached marker reset")
					}
					events = append(events, "observe")
					cached.Field57, cached.State = math.Float32bits(0.8), PlayerState1
					cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0, 0
					target.UpdateData = unsafe.Pointer(live)
					if defense == "Reflect Shield" {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
				}
				position := attack.PrevPos
				exclude := func(w *Object) bool {
					if w != attack || !slices.Equal(events, []string{"observe"}) || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("exclusion must precede attribution and execute once")
					}
					events = append(events, "exclude")
					attack.TypeInd++
					attack.PrevPos = types.Ptf(-72, 87)
					return defense == "excluded shield"
				}
				wrong := func(*Object) bool { t.Fatal("wrong four/six exclusion set"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, wrong
				if weapon == nil {
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = wrong, exclude
				}
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || pos != position || !slices.Equal(events, []string{"observe", "exclude"}) || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) {
						t.Fatal("facing lost the position snapshot or attribution order")
					}
					events = append(events, "direction")
					cached.State = PlayerState13
					if defense == "shield" || defense == "rear shield" {
						cached.State = PlayerState16
					} else if defense == "staff" {
						cached.State = PlayerState21
					}
					return defense != "rear shield"
				}
				blockItem := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
				sound, blocked := 0, defense == "shield" || defense == "GreatSword" || defense == "staff"
				if defense == "shield" {
					blockItem.ObjSubClass, sound = 2, 878
				} else if defense == "GreatSword" {
					blockItem.ObjSubClass, sound = 0x400, 890
				} else if defense == "staff" {
					blockItem.ObjSubClass, sound = 0x8000, 894
				}
				if blocked {
					target.InvFirstItem = blockItem
				}
				r.Audio = func(id int, v *Object) {
					if v != target || id != sound {
						t.Fatal("block audio")
					}
					events = append(events, "audio")
				}
				r.Melee.RandomInt = func(min, max int) int {
					if min != 18 || max != 20 {
						t.Fatal("GreatSword RNG")
					}
					events = append(events, "rng")
					return 19
				}
				r.PlayerSetState = func(v *Object, state PlayerState) bool {
					want := PlayerState19
					if defense == "staff" {
						want = PlayerState21
					}
					if v != target || state != want {
						t.Fatal("block live state")
					}
					events = append(events, "state")
					live.State = state
					return true
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
				r.CanDamageBlockItem = func(item *Object) bool { return item == blockItem }
				r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
				wear := func(item, v, a, w *Object, d float32, typ object.DamageType) bool {
					if item != blockItem || v != target || a != source || w != attack || d != 2.5 || typ != kind.typ || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) {
						t.Fatal("block lost cached marker/source")
					}
					events = append(events, "wear")
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
				tail := r.DefaultDamage
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					events = append(events, "default")
					return tail(v, a, w, d, typ)
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
				wantEvents := []string{"observe", "exclude"}
				if defense != "excluded shield" {
					wantEvents = append(wantEvents, "direction")
				}
				wantHP, wantCarry := uint16(196), float32(0.25)
				if kind.typ == object.DamageCrush {
					wantHP, wantCarry = 195, -0.125
				}
				if blocked {
					wantHP, wantCarry = 200, 0.5
					wantEvents = append(wantEvents, "audio")
					if defense == "GreatSword" {
						wantEvents = append(wantEvents, "rng")
					}
					if defense != "shield" {
						wantEvents = append(wantEvents, "state")
					}
					wantEvents = append(wantEvents, "balance", "wear")
				} else {
					wantEvents = append(wantEvents, "default")
				}
				if !h || result == blocked || !slices.Equal(events, wantEvents) || target.HealthData.Cur != wantHP || live.Field21 != math.Float32bits(wantCarry) || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("possession melee=%t/%t HP=%d marker=%d/%d carry=%g events=%v want=%v", h, result, target.HealthData.Cur, cached.Field76, cached.Field75, math.Float32frombits(live.Field21), events, wantEvents)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionMeleeTail4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, tc := range []struct {
			name                    string
			damage                  int32
			clearMarker, quest, god bool
			dropPlayer              string
		}{
			{"retain callback marker", 5, false, false, false, ""},
			{"fallback cleared marker", 5, true, false, false, ""},
			{"live Quest after wear", 5, false, true, false, ""},
			{"God after wear", 5, false, true, true, ""},
			{"God uses live class after wear", 5, false, true, true, "wear"},
			{"God flag precedes live class", 5, false, true, true, "God"},
			{"positive minimum", 1, false, false, false, ""},
			{"zero", 0, false, false, false, ""},
			{"signed negative", -5, false, false, false, ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				target, source, weapon, _ := possessionMeleeFixture4E17B0(t, playerSource, kind)
				cached := target.UpdateDataPlayer()
				armor := damageMeleeArmorFixture4E17B0(target, 0.25, 0.125)
				modifier := &ModifierEff{}
				modifier.Defend76.Fnc = unsafe.Pointer(new(byte))
				armor.InitDataModifier().Modifiers[1] = modifier
				live := &PlayerUpdateData{Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				if tc.damage == 1 {
					live.Field21 = math.Float32bits(-0.5)
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				r.ObserveClear = func(*Object) { target.UpdateData = unsafe.Pointer(live); cached.Field57 = math.Float32bits(0.9) }
				r.BlockSourceExcluded = func(*Object) bool { return false }
				r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
				r.BlockDirection = func(*Object, types.Pointf) bool { cached.Field76, cached.Field75 = 9, 321; return false }
				absorption := float64(0.25)
				if kind.typ == object.DamageCrush {
					absorption *= 0.5
				}
				accumulated := float32((1-absorption)*float64(tc.damage)) + math.Float32frombits(live.Field21)
				effective := playerDamageRound4E17B0(accumulated)
				carry := accumulated - float32(effective)
				wearAmount := float32(tc.damage - effective) // live item/total armor = 0.5/0.5
				itemTotal := wearAmount + float32(0.4)
				itemDamage := itemDurabilityRound4E1560(itemTotal)
				var events []string
				liveQuest := false
				r.ItemArmorValue = func(item *Object) float32 {
					if item != armor || live.Field21 != math.Float32bits(carry) || cached.Field76 != 9 || cached.Field75 != 321 {
						t.Fatal("armor lookup preceded live carry or lost callback marker")
					}
					events = append(events, "lookup")
					return 0.5
				}
				r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
				r.ApplyArmorDefend = func(m *ModifierEff, item, owner, w, a *Object, amount *float32) bool {
					// Defend runs for every signed/zero amount; the subsequent
					// Damage callback runs only when rounded item wear is positive.
					if m != modifier || item != armor || owner != target || a != source || w != weapon || *amount != wearAmount || cached.Field76 != 9 || cached.Field75 != 321 || live.Field21 != math.Float32bits(carry) {
						t.Fatal("wear lost live carry or overwrote callback marker")
					}
					events = append(events, "wear")
					if tc.clearMarker {
						cached.Field76, cached.Field75 = 0, 99
					}
					liveQuest = tc.quest
					if tc.dropPlayer == "wear" {
						target.ObjClass = object.ClassSimple
					}
					return true
				}
				r.DamageArmor = func(item, a, w *Object, d int32, typ object.DamageType) bool {
					if item != armor || a != source || w != weapon || d != itemDamage || d <= 0 || typ != kind.typ {
						t.Fatal("armor damage callback arguments")
					}
					events = append(events, "armor damage")
					return true
				}
				r.GodMode = func() bool {
					events = append(events, "God")
					if tc.dropPlayer == "God" {
						target.ObjClass = object.ClassSimple
					}
					return tc.god
				}
				r.QuestMode = func() bool { events = append(events, "Quest"); return liveQuest }
				r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
				if tc.damage > 0 && effective == 0 {
					effective = 1
				}
				god := tc.god && tc.dropPlayer == ""
				if tc.quest && !god {
					before := effective
					effective = playerDamageRound4E17B0(float32(float64(effective) * 0.5))
					if before > 0 && effective < 1 {
						effective = 1
					}
				}
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					if v != target || a != source || w != weapon || d != effective || typ != kind.typ {
						t.Fatal("damage tail arguments")
					}
					events = append(events, "default")
					return true
				}
				wantEvents := []string{"lookup", "wear"}
				if itemDamage > 0 {
					wantEvents = append(wantEvents, "armor damage")
				}
				wantEvents = append(wantEvents, "God")
				if !god {
					wantEvents = append(wantEvents, "Quest")
					if tc.quest {
						wantEvents = append(wantEvents, "scale")
					}
					wantEvents = append(wantEvents, "default")
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, kind.typ, r)
				marker, markerType := uint32(9), uint32(321)
				if tc.clearMarker {
					marker, markerType = 2, uint32(kind.typ)
				}
				if !h || !result || !slices.Equal(events, wantEvents) || live.Field21 != math.Float32bits(carry) || live.Field76 != 31 || live.Field75 != 33 || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != marker || cached.Field75 != markerType || armor.UpdateDataWeaponArmor().Field0 != math.Float32bits(itemTotal-float32(itemDamage)) {
					t.Fatalf("tail=%t/%t marker=%d/%d events=%v want=%v", h, result, cached.Field76, cached.Field75, events, wantEvents)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionMeleeLiveRecordGuard4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, stage := range []string{"ObserveClear", "direction"} {
			for _, change := range []string{"class", "nil update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, weapon, attack := possessionMeleeFixture4E17B0(t, playerSource, kind)
					cached := target.UpdateDataPlayer()
					beforeCarry := cached.Field21
					invalidate := func() {
						if change == "class" {
							target.ObjClass = object.ClassMonster
						} else {
							target.UpdateData = nil
						}
					}
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.ObserveClear = func(*Object) {
						if cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("live-record guard lost cached prefix reset")
						}
						if stage == "ObserveClear" {
							invalidate()
						}
					}
					r.BlockSourceExcluded = func(*Object) bool { return false }
					r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
					r.BlockDirection = func(*Object, types.Pointf) bool { invalidate(); return false }
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("invalid live record reached damage tail")
						return false
					}
					reason := ""
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
					marker, markerType := uint32(0), uint32(77)
					if stage == "direction" {
						marker, markerType = 1, uint32(attack.TypeInd)
					}
					if h || result || reason == "" || target.HealthData.Cur != 200 || cached.Field21 != beforeCarry || cached.Field76 != marker || cached.Field75 != markerType {
						t.Fatalf("live guard=%t/%t reason=%q marker=%d/%d HP=%d", h, result, reason, cached.Field76, cached.Field75, target.HealthData.Cur)
					}
				})
			}
		}
	})
}

func TestPlayerDamagePossessionMeleeAdmission4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, service := range []string{"observe", "exclusion", "direction", "default", "Quest scale"} {
			t.Run(service, func(t *testing.T) {
				target, source, weapon, _ := possessionMeleeFixture4E17B0(t, playerSource, kind)
				ud := target.UpdateDataPlayer()
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.ObserveClear = func(*Object) { t.Fatal("missing service reached ObserveClear") }
				r.BlockSourceExcluded = func(*Object) bool { return false }
				r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
				r.BlockDirection = func(*Object, types.Pointf) bool { return false }
				switch service {
				case "observe":
					r.ObserveClear = nil
				case "exclusion":
					if weapon == nil {
						r.BlockSourceOnlyExcluded = nil
					} else {
						r.BlockSourceExcluded = nil
					}
				case "direction":
					r.BlockDirection = nil
				case "default":
					r.DefaultDamage = nil
				case "Quest scale":
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = nil
				}
				var reason string
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r); h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("admission=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionMeleeEarlyGates4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon, _ := possessionMeleeFixture4E17B0(t, playerSource, kind)
				ud := target.UpdateDataPlayer()
				r := damageMeleeRuntimeFixture4E17B0(t)
				switch gate {
				case "no update":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "observer":
					ud.Player.Field3680 |= 1
				case "Coop self":
					source = target
					r.CoopMode = func() bool { return true }
				}
				r.ObserveClear = func(*Object) { t.Fatal("early gate cleared possession") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate reached exclusion"); return false }
				r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate reached direction"); return false }
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || id != 71 || v != target {
						t.Fatal("early audio")
					}
				}
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r); !h || result != (gate == "invulnerable") || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("early gate=%t/%t", h, result)
				}
			})
		}
	})
}
