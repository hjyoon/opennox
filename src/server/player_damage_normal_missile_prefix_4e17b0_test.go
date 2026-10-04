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

// Entry/callback contracts only: DefaultDamage records its signed input here.
// The separate C-owned tests exercise the real protection/HP services.
func normalMissilePrefixCases4E17B0(t *testing.T, check func(*testing.T, bool, string, object.DamageType)) {
	t.Helper()
	for _, playerSource := range []bool{false, true} {
		for _, kind := range []string{"PIERCE stock", "PIERCE pure", "FLAME direct", "FLAME splash", "EXPLOSION direct", "EXPLOSION splash"} {
			t.Run(fmt.Sprintf("player-source-%t/%s", playerSource, kind), func(t *testing.T) {
				typ := object.DamageImpale
				if kind == "FLAME direct" || kind == "FLAME splash" {
					typ = object.DamageFlame
				} else if kind == "EXPLOSION direct" || kind == "EXPLOSION splash" {
					typ = object.DamageExplosion
				}
				check(t, playerSource, kind, typ)
			})
		}
	}
}

func normalMissilePrefixFixture4E17B0(t *testing.T, playerSource bool, kind string) (target, source, weapon, missile *Object, cached *PlayerUpdateData, r PlayerDamageRuntime4E17B0) {
	t.Helper()
	if kind == "PIERCE stock" || kind == "PIERCE pure" {
		target, source, weapon, r, _ = playerDamagePierceFixture4E17B0(t, playerSource)
		missile = weapon
		if kind == "PIERCE pure" {
			missile.ObjClass = object.ClassMissile
		}
	} else {
		splash := kind == "FLAME splash" || kind == "EXPLOSION splash"
		target, source, weapon, missile = damageFlameFixture4E17B0(t, playerSource, true, splash)
		r = playerDamageRuntime4E17B0(t, nil, new([]int32))
	}
	cached = target.UpdateDataPlayer()
	cached.Player.Field3680, cached.Player.CameraFollowObj = 0, nil
	cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0, 0
	cached.State = PlayerState13
	cached.Field57, cached.Field21 = math.Float32bits(0.25), math.Float32bits(0.5)
	cached.Field76, cached.Field75 = 88, 77
	missile.PosVec, missile.PrevPos = types.Ptf(11, 22), types.Ptf(33, 44)
	r.ObserveClear = nil // Neither required nor invoked for an ordinary target.
	r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool { return true }
	return
}

func TestPlayerDamageNormalMissilePrefix4E17B0(t *testing.T) {
	normalMissilePrefixCases4E17B0(t, func(t *testing.T, playerSource bool, kind string, typ object.DamageType) {
		for _, raw := range []int32{5, 0, -5} {
			for _, defenseName := range []string{"none", "Reflect front", "Reflect rear", "shield", "GreatSword", "GreatSword without default", "excluded shield"} {
				t.Run(fmt.Sprintf("raw-%d/%s", raw, defenseName), func(t *testing.T) {
					defense := defenseName
					if defenseName == "GreatSword without default" {
						defense = "GreatSword"
					}
					target, source, weapon, missile, cached, r := normalMissilePrefixFixture4E17B0(t, playerSource, kind)
					beforeHealth := *target.HealthData
					var events []string
					var item *Object
					switch defense {
					case "Reflect front", "Reflect rear":
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					case "shield", "excluded shield":
						cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
						// A no-health shield is a supported EquipDamage wear no-op.
						item = &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
						r.ItemArmorValue = func(*Object) float32 { return 1 }
					case "GreatSword":
						cached.Player.WeaponEquip = 0x400
						item = &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 30, Max: 30}}
					}
					target.InvFirstItem = item
					blocked := defense == "Reflect front" || defense == "shield" || defense == "GreatSword"
					previous := missile.PrevPos
					exclusion := func(v *Object, hasWeapon bool) bool {
						if v != missile || hasWeapon != (weapon != nil) || defense == "Reflect front" || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("normal missile exclusion skipped the reset or selected the wrong 4/6 set")
						}
						events = append(events, "exclude")
						missile.PrevPos, missile.TypeInd = types.Ptf(99, 77), 777
						// These values were cached before the external prefix calls.
						cached.Field57 = math.Float32bits(0.9)
						cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0, 0
						return defense == "excluded shield"
					}
					r.BlockSourceExcluded = func(v *Object) bool { return exclusion(v, true) }
					r.BlockSourceOnlyExcluded = func(v *Object) bool { return exclusion(v, false) }
					prefixMarker := func() (uint32, uint32) {
						if weapon != nil {
							return 1, 777
						}
						return 0, 77
					}
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target {
							t.Fatal("wrong direction target")
						}
						if pos == missile.PosVec {
							if (defense != "Reflect front" && defense != "Reflect rear") || len(events) != 0 || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("Reflect direction preceded the clear marker or followed exclusion")
							}
							events = append(events, "Reflect")
							return defense == "Reflect front"
						}
						marker, markerType := prefixMarker()
						if pos != previous || len(events) == 0 || events[len(events)-1] != "exclude" || cached.Field76 != marker || cached.Field75 != markerType {
							t.Fatal("normal missile facing lost snapshot, attribution, or callback order")
						}
						events = append(events, "direction")
						return defense == "shield" || defense == "GreatSword"
					}
					r.ProjectileReflect = func(v, owner *Object) {
						if v != missile || owner != target || !blocked {
							t.Fatal("bad reflection arguments")
						}
						events = append(events, "reflect")
					}
					r.ClearOwner = func(v *Object) {
						if v != missile {
							t.Fatal("bad owner clear")
						}
						events = append(events, "clear")
					}
					r.SetOwner = func(owner, v *Object) {
						if owner != target || v != missile {
							t.Fatal("bad owner set")
						}
						events = append(events, "owner")
					}
					r.Audio = func(id int, v *Object) {
						wantID := map[string]int{"Reflect front": 122, "shield": 878, "GreatSword": 890}[defense]
						if !blocked || id != wantID || v != target {
							t.Fatal("bad block audio")
						}
						events = append(events, "audio")
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
					r.CanDamageBlockItem, r.Melee.CanDamageBlockWeapon = func(v *Object) bool { return v == item }, func(v *Object) bool { return v == item }
					wear := func(v, owner, a, effective *Object, d float32, gotType object.DamageType) bool {
						marker, markerType := prefixMarker()
						if v != item || owner != target || a != source || effective != missile || d != float32(raw)*0.5 || gotType != typ || cached.Field76 != marker || cached.Field75 != markerType {
							t.Fatal("block wear changed signed input or cached attribution")
						}
						events = append(events, "wear")
						return true
					}
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
					r.Melee.RandomInt = func(lo, hi int) int {
						if lo != 18 || hi != 20 {
							t.Fatal("bad GreatSword state range")
						}
						events = append(events, "rng")
						return 19
					}
					r.PlayerSetState = func(v *Object, state PlayerState) bool {
						if defense != "GreatSword" || v != target || state != PlayerState19 {
							t.Fatal("bad GreatSword state")
						}
						events = append(events, "state")
						cached.State = state
						return true
					}
					amount, residual := raw, float32(0.5)
					if typ != object.DamageFlame {
						switch raw {
						case 5:
							amount, residual = 4, 0.25
						case -5:
							amount, residual = -3, -0.25
						}
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						marker, markerType := prefixMarker()
						if marker == 0 {
							marker, markerType = 2, uint32(typ)
						}
						if blocked || v != target || a != source || w != weapon || d != amount || gotType != typ || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(residual) {
							t.Fatal("normal missile tail changed signed amount, cached armor, live carry, or marker")
						}
						events = append(events, "default")
						return true
					}
					if defenseName == "GreatSword without default" {
						r.DefaultDamage = nil
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, typ, r)
					var want []string
					if defense == "Reflect front" || defense == "Reflect rear" {
						want = append(want, "Reflect")
					}
					if defense != "Reflect front" {
						want = append(want, "exclude")
						if defense != "excluded shield" {
							want = append(want, "direction")
						}
					}
					switch defense {
					case "Reflect front":
						want = append(want, "reflect", "clear", "owner", "audio")
					case "shield":
						want = append(want, "audio")
						if uint32(missile.SubClass())&0x70 == 0 {
							want = append(want, "reflect", "clear", "owner")
						}
						want = append(want, "balance", "wear")
					case "GreatSword":
						want = append(want, "reflect", "clear", "owner", "audio", "rng", "state", "balance", "wear")
					default:
						want = append(want, "default")
					}
					marker, markerType := prefixMarker()
					if defense == "Reflect front" {
						marker, markerType = 0, 77
					} else if !blocked && marker == 0 {
						marker, markerType = 2, uint32(typ)
					}
					if blocked {
						residual = 0.5
					}
					if !h || result == blocked || !slices.Equal(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(residual) || *target.HealthData != beforeHealth {
						t.Fatalf("normal missile=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMissilePrefixLive4E17B0(t *testing.T) {
	normalMissilePrefixCases4E17B0(t, func(t *testing.T, playerSource bool, kind string, typ object.DamageType) {
		for _, stage := range []string{"exclude", "direction"} {
			for _, marker := range []uint32{0, 42, 99} {
				t.Run(fmt.Sprintf("%s/marker-%d", stage, marker), func(t *testing.T) {
					target, source, weapon, missile, cached, r := normalMissilePrefixFixture4E17B0(t, playerSource, kind)
					cached.Field21 = math.Float32bits(0.125)
					live := &PlayerUpdateData{State: PlayerState16, Field57: math.Float32bits(0.9), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
					if marker != 99 {
						live.Player = &Player{Field3680: 3, CameraFollowObj: source, ArmorEquip: 0x1000000, WeaponEquip: 0x400}
					}
					var events []string
					previous := missile.PrevPos
					swap := func() { target.UpdateData = unsafe.Pointer(live); cached.Field76, cached.Field75 = marker, 777 }
					exclusion := func(v *Object) bool {
						if v != missile || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("live exclusion skipped entry reset")
						}
						events = append(events, "exclude")
						missile.PrevPos, missile.TypeInd = types.Ptf(99, 77), 777
						if stage == "exclude" {
							swap()
						}
						return false
					}
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || pos != previous {
							t.Fatal("live facing lost previous-position snapshot")
						}
						events = append(events, "direction")
						if stage == "direction" {
							swap()
						}
						return false
					}
					r.ObserveClear = func(*Object) { t.Fatal("normal prefix re-queried replacement observer") }
					wantMarker, wantType := marker, uint32(777)
					if stage == "exclude" && weapon != nil {
						wantMarker = 1
					}
					if wantMarker == 0 {
						wantMarker, wantType = 2, uint32(typ)
					}
					amount, residual := int32(4), float32(0.25)
					if typ == object.DamageFlame {
						amount, residual = 5, 0.5
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						if v != target || a != source || w != weapon || d != amount || gotType != typ || cached.Field76 != wantMarker || cached.Field75 != wantType || live.Field21 != math.Float32bits(residual) {
							t.Fatal("live tail lost cached-marker/live-carry split")
						}
						events = append(events, "default")
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r)
					if !h || !result || !slices.Equal(events, []string{"exclude", "direction", "default"}) || cached.Field21 != math.Float32bits(0.125) || live.Field76 != 31 || live.Field75 != 33 {
						t.Fatalf("live prefix=%t/%t events=%v", h, result, events)
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMissilePrefixMissing4E17B0(t *testing.T) {
	normalMissilePrefixCases4E17B0(t, func(t *testing.T, playerSource bool, kind string, typ object.DamageType) {
		for _, missing := range []string{"exclusion", "direction", "default", "Quest", "armor"} {
			t.Run(missing, func(t *testing.T) {
				target, source, weapon, missile, cached, r := normalMissilePrefixFixture4E17B0(t, playerSource, kind)
				var events []string
				exclusion := func(*Object) bool { events = append(events, "exclude"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
				r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return false }
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("missing service reached default")
					return false
				}
				switch missing {
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
				case "Quest":
					r.QuestMode, r.QuestDamageScale = func() bool { return true }, nil
				case "armor":
					target.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 100}}
				}
				reason := ""
				r.Unsupported = func(why string, v, a, w *Object, d int32, gotType object.DamageType) {
					if v != target || a != source || w != weapon || d != -5 || gotType != typ {
						t.Fatal("missing service report lost original signed arguments")
					}
					reason = why
				}
				before, beforeCached, beforePlayer, beforeSource, beforeMissile, beforeHealth := *target, *cached, *cached.Player, *source, *missile, *target.HealthData
				h, result := PlayerDamageNative4E17B0(target, source, weapon, -5, typ, r)
				if h || result || reason == "" || len(events) != 0 || *target != before || *cached != beforeCached || *cached.Player != beforePlayer || *source != beforeSource || *missile != beforeMissile || *target.HealthData != beforeHealth {
					t.Fatalf("missing normal service=%t/%t reason=%q events=%v", h, result, reason, events)
				}
			})
		}
	})
}

func TestPlayerDamageNormalMissilePrefixEarly4E17B0(t *testing.T) {
	normalMissilePrefixCases4E17B0(t, func(t *testing.T, playerSource bool, kind string, typ object.DamageType) {
		for _, gate := range []string{"NoUpdate", "dead", "observer", "Coop", "invulnerable"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon, _, cached, _ := normalMissilePrefixFixture4E17B0(t, playerSource, kind)
				r := PlayerDamageRuntime4E17B0{Frame: func() uint32 { return 1400 }}
				switch gate {
				case "NoUpdate":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "observer":
					cached.Player.Field3680 = 1
				case "Coop":
					r.CoopMode = func() bool { return true }
					if weapon != nil && playerSource {
						source = target
					} else {
						source.ObjOwner = target
					}
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				}
				before, beforeCached, beforePlayer, beforeSource := *target, *cached, *cached.Player, *source
				h, result := PlayerDamageNative4E17B0(target, source, weapon, -5, typ, r)
				if !h || result != (gate == "invulnerable") || *target != before || *cached != beforeCached || *cached.Player != beforePlayer || *source != beforeSource {
					t.Fatalf("normal missile early=%t/%t marker=%d", h, result, cached.Field76)
				}
			})
		}
	})
}
