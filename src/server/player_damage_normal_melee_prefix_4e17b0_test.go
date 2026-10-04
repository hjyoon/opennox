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

// This slice has no entry shield, GreatSword or staff block equipment. Keep
// those normal-player block branches and the possessed branch independently
// covered; a late callback must not restart equipment admission here.
func normalMeleePrefixFixture4E17B0(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) (target, source, weapon, attack *Object) {
	t.Helper()
	target, source, weapon, attack = possessionMeleeFixture4E17B0(t, playerSource, kind)
	ud := target.UpdateDataPlayer()
	ud.Player.Field3680, ud.Player.CameraFollowObj = 0, nil
	attack.PrevPos, attack.PosVec = types.Ptf(20, 0), types.Ptf(-20, 0)
	return
}

func TestPlayerDamageNormalMeleePrefix4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, damage := range []int32{5, 0, -5} {
			for _, defense := range []string{"front", "rear", "excluded", "Reflect Shield"} {
				t.Run(fmt.Sprintf("damage-%d/%s", damage, defense), func(t *testing.T) {
					target, source, weapon, attack := normalMeleePrefixFixture4E17B0(t, playerSource, kind)
					ud, position := target.UpdateDataPlayer(), attack.PrevPos
					if defense == "Reflect Shield" {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
					var events []string
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.ObserveClear = func(*Object) { t.Fatal("normal player must not clear possession") }
					exclude := func(v *Object) bool {
						if v != attack || len(events) != 0 || ud.Field76 != 0 || ud.Field75 != 77 {
							t.Fatal("exclusion must follow only the cached marker reset")
						}
						events = append(events, "exclude")
						attack.TypeInd++
						attack.PrevPos = types.Ptf(-72, 87)
						return defense == "excluded"
					}
					wrong := func(*Object) bool { t.Fatal("wrong four/six exclusion set"); return false }
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, wrong
					if weapon == nil {
						r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = wrong, exclude
					}
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || pos != position || !slices.Equal(events, []string{"exclude"}) || ud.Field76 != 1 || ud.Field75 != uint32(attack.TypeInd) {
							t.Fatal("facing lost the previous-position snapshot or attribution order")
						}
						events = append(events, "direction")
						ud.Field76, ud.Field75 = 9, 321
						// New cached equipment/armor cannot restart the prefix.
						ud.Field57, ud.State = math.Float32bits(0.9), PlayerState16
						ud.Player.WeaponEquip, ud.Player.ArmorEquip = 0x400, 0x1000000
						return defense != "rear"
					}
					r.ProjectileReflect = func(*Object, *Object) { t.Fatal("ordinary melee/SIMPLE is not reflected") }
					absorption := float64(0.25)
					if kind.typ == object.DamageCrush {
						absorption *= 0.5
					}
					accumulated := float32((1-absorption)*float64(damage)) + 0.125
					effective := playerDamageRound4E17B0(accumulated)
					wantCarry := math.Float32bits(accumulated - float32(effective))
					if damage > 0 && effective == 0 {
						effective = 1
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if v != target || a != source || w != weapon || d != effective || typ != kind.typ {
							t.Fatal("normal melee tail arguments")
						}
						events = append(events, "default")
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, damage, kind.typ, r)
					want := []string{"exclude"}
					marker, markerType := uint32(1), uint32(attack.TypeInd)
					if defense != "excluded" {
						want = append(want, "direction")
						marker, markerType = 9, 321
					}
					want = append(want, "default")
					if !h || !result || !slices.Equal(events, want) || ud.Field21 != wantCarry || ud.Field76 != marker || ud.Field75 != markerType || target.HealthData.Cur != 200 {
						t.Fatalf("prefix=%t/%t carry=%g marker=%d/%d events=%v want=%v", h, result, math.Float32frombits(ud.Field21), ud.Field76, ud.Field75, events, want)
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMeleePrefixLive4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, stage := range []string{"exclusion", "direction"} {
			for _, change := range []string{"replacement", "class", "nil update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, weapon, attack := normalMeleePrefixFixture4E17B0(t, playerSource, kind)
					cached := target.UpdateDataPlayer()
					live := &PlayerUpdateData{Player: &Player{Field3680: 3, CameraFollowObj: source}, Field57: math.Float32bits(0.9), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
					changeRecord := func() {
						switch change {
						case "replacement":
							target.UpdateData = unsafe.Pointer(live)
						case "class":
							target.ObjClass = object.ClassSimple
						case "nil update":
							target.UpdateData = nil
						}
						cached.Field57 = math.Float32bits(0.8)
						cached.Player.WeaponEquip, cached.Player.ArmorEquip = 0x400, 0x1000000
					}
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.ObserveClear = func(*Object) { t.Fatal("new live observer must not restart prefix") }
					exclude := func(*Object) bool {
						if stage == "exclusion" {
							changeRecord()
						}
						return false
					}
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
					r.BlockDirection = func(*Object, types.Pointf) bool {
						if stage == "direction" {
							changeRecord()
						}
						return false
					}
					absorption := float64(0.25)
					if kind.typ == object.DamageCrush {
						absorption *= 0.5
					}
					accumulated := float32((1-absorption)*5) + 0.5
					effective := playerDamageRound4E17B0(accumulated)
					calls, reason := 0, ""
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if change != "replacement" || v != target || a != source || w != weapon || d != effective || typ != kind.typ {
							t.Fatal("invalid live record reached the tail, or cached armor/live carry was lost")
						}
						calls++
						return true
					}
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
					if change == "replacement" {
						if !h || !result || calls != 1 || reason != "" || live.Field21 != math.Float32bits(accumulated-float32(effective)) || live.Field76 != 31 || live.Field75 != 33 {
							t.Fatalf("live replacement=%t/%t calls=%d reason=%q carry=%g", h, result, calls, reason, math.Float32frombits(live.Field21))
						}
					} else if h || result || calls != 0 || reason == "" || target.HealthData.Cur != 200 {
						t.Fatalf("live guard=%t/%t calls=%d reason=%q", h, result, calls, reason)
					}
					if cached.Field21 != math.Float32bits(0.125) || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) {
						t.Fatal("live carry/guard overwrote the cached prefix base")
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMeleePrefixTail4E17B0(t *testing.T) {
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
				target, source, weapon, _ := normalMeleePrefixFixture4E17B0(t, playerSource, kind)
				ud := target.UpdateDataPlayer()
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.125)
				modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				armor.InitDataModifier().Modifiers[1] = modifier
				if tc.damage == 1 {
					ud.Field21 = math.Float32bits(-0.5)
				}
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = func(*Object) bool { return false }, func(*Object) bool { return false }
				r.BlockDirection = func(*Object, types.Pointf) bool { ud.Field76, ud.Field75 = 9, 321; return false }
				absorption := float64(0.5)
				if kind.typ == object.DamageCrush {
					absorption *= 0.5
				}
				accumulated := float32((1-absorption)*float64(tc.damage)) + math.Float32frombits(ud.Field21)
				effective := playerDamageRound4E17B0(accumulated)
				carry := math.Float32bits(accumulated - float32(effective))
				wear := float32(tc.damage - effective)
				var events []string
				liveQuest := false
				r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
				r.ApplyArmorDefend = func(m *ModifierEff, item, v, w, a *Object, d *float32) bool {
					if m != modifier || item != armor || v != target || w != weapon || a != source || *d != wear || ud.Field21 != carry || ud.Field76 != 9 || ud.Field75 != 321 {
						t.Fatal("wear lost cached attribution/live carry or signed arguments")
					}
					events = append(events, "wear")
					if tc.clearMarker {
						ud.Field76, ud.Field75 = 0, 99
					}
					liveQuest = tc.quest
					if tc.dropPlayer == "wear" {
						target.ObjClass = object.ClassSimple
					}
					return true
				}
				r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool { return true }
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
						t.Fatal("late Quest/God tail arguments")
					}
					events = append(events, "default")
					return true
				}
				want := []string{"wear", "God"}
				if !god {
					want = append(want, "Quest")
					if tc.quest {
						want = append(want, "scale")
					}
					want = append(want, "default")
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, kind.typ, r)
				marker, markerType := uint32(9), uint32(321)
				if tc.clearMarker {
					marker, markerType = 2, uint32(kind.typ)
				}
				if !h || !result || !slices.Equal(events, want) || ud.Field21 != carry || ud.Field76 != marker || ud.Field75 != markerType {
					t.Fatalf("tail=%t/%t marker=%d/%d events=%v want=%v", h, result, ud.Field76, ud.Field75, events, want)
				}
			})
		}
	})
}

func TestPlayerDamageNormalMeleePrefixAdmission4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, service := range []string{"exclusion", "direction", "default", "Quest scale", "armor lookup", "armor init", "armor damage", "armor modifier"} {
			t.Run(service, func(t *testing.T) {
				target, source, weapon, _ := normalMeleePrefixFixture4E17B0(t, playerSource, kind)
				ud := target.UpdateDataPlayer()
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.125)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				exclude := func(*Object) bool { t.Fatal("missing service consumed prefix"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("missing service reached facing"); return false }
				switch service {
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
				case "armor lookup":
					r.ItemArmorValue = nil
				case "armor init":
					armor.InitData = nil
				case "armor damage":
					r.DamageArmor = nil
				case "armor modifier":
					armor.InitDataModifier().Modifiers[1] = &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				before, beforeUD, beforePlayer, beforeArmor := *target, *ud, *ud.Player, *armor
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
				if h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer || *armor != beforeArmor || armor.HealthData.Cur != 25 {
					t.Fatalf("admission=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamageNormalMeleePrefixEarly4E17B0(t *testing.T) {
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon, _ := normalMeleePrefixFixture4E17B0(t, playerSource, kind)
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
				exclude := func(*Object) bool { t.Fatal("early gate consumed prefix"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate reached facing"); return false }
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || id != 71 || v != target {
						t.Fatal("early audio")
					}
				}
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
				if !h || result != (gate == "invulnerable") || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("early gate=%t/%t", h, result)
				}
			})
		}
	})
}
