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

// These are entry/callback contract tests. The C-owned integration tests use
// the actual DefaultDamage/protection/HP services; this fixture records the
// signed amount forwarded to that separate tail without inventing a clamp.
func normalElectricPrefixFixture4E17B0(t *testing.T, playerSource, selfWeapon bool, events *[]string, damages *[]int32) (target, source, weapon *Object, cached *PlayerUpdateData, r PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, weapon = possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
	cached = target.UpdateDataPlayer()
	cached.Player.Field3680, cached.Player.CameraFollowObj = 0, nil
	cached.Field21 = math.Float32bits(0.5)
	r = playerDamageRuntime4E17B0(t, nil, damages)
	// A normal target does not require or invoke ObserveClear.
	r.ObserveClear = nil
	r.ElectricArmorScale = func(v *Object) float32 {
		if v != target || cached.Field76 != 0 {
			t.Fatal("electric scale preceded the cached marker reset")
		}
		*events = append(*events, "scale")
		return 0.5
	}
	return
}

func TestPlayerDamageNormalElectricPrefix4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, raw := range []int32{5, 0, -5} {
			for _, defense := range []string{"none", "shield", "great sword", "staff", "Reflect front", "Reflect rear", "excluded"} {
				t.Run(fmt.Sprintf("raw-%d/%s", raw, defense), func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, r := normalElectricPrefixFixture4E17B0(t, playerSource, selfWeapon, &events, &damages)
					switch defense {
					case "shield":
						cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
					case "great sword":
						cached.State, cached.Player.WeaponEquip = PlayerState13, 0x400
					case "staff":
						cached.State, cached.Player.WeaponEquip = PlayerState13, 0x8000
					case "Reflect front", "Reflect rear":
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
					blocked := defense == "Reflect front" && typ == object.DamageAirborneElectric
					previous := source.PrevPos
					exclusion := func(v *Object, hasWeapon bool) bool {
						if v != source || hasWeapon != selfWeapon || blocked || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("normal electric exclusion skipped the clear marker or used the wrong 4/6 set")
						}
						events = append(events, "exclude")
						// 004E1A49/004E1ABE already saved this previous position.
						source.PrevPos, source.TypeInd = types.Ptf(99, 77), 777
						return defense == "excluded"
					}
					r.BlockSourceExcluded = func(v *Object) bool { return exclusion(v, true) }
					r.BlockSourceOnlyExcluded = func(v *Object) bool { return exclusion(v, false) }
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("normal electric direction preceded the cached marker reset")
						}
						if pos == source.PosVec {
							if typ != object.DamageAirborneElectric || (defense != "Reflect front" && defense != "Reflect rear") || len(events) != 0 {
								t.Fatal("normal electric Reflect must use current position before exclusions")
							}
							events = append(events, "Reflect")
							return blocked
						}
						if blocked || defense == "excluded" || pos != previous || events[len(events)-1] != "exclude" {
							t.Fatal("normal electric facing lost the previous-position snapshot")
						}
						events = append(events, "direction")
						return true
					}
					r.Audio = func(id int, v *Object) {
						if !blocked || id != 122 || v != target {
							t.Fatal("ordinary equipment blocked an electric hit")
						}
						events = append(events, "audio")
					}
					amount := int32(0)
					if raw > 0 {
						amount = 3
					} else if raw < 0 {
						amount = -2
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						if blocked || v != target || a != source || w != weapon || d != amount || gotType != typ || cached.Field76 != 2 || cached.Field75 != uint32(typ) {
							t.Fatal("normal electric prefix changed the signed switch tail or attributed a self weapon")
						}
						events = append(events, "default")
						r.DamageClear(v, d)
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, typ, r)
					var want []string
					if typ == object.DamageAirborneElectric && (defense == "Reflect front" || defense == "Reflect rear") {
						want = append(want, "Reflect")
					}
					marker, markerType, residual, hp := uint32(2), uint32(typ), float32(0), uint16(60-amount)
					if blocked {
						want = append(want, "audio")
						marker, markerType, residual, hp = 0, 77, 0.5, 60
					} else {
						want = append(want, "exclude")
						if defense != "excluded" {
							want = append(want, "direction")
						}
						want = append(want, "scale", "default")
						if raw == 0 {
							residual = 0.5
						}
					}
					if !h || result == blocked || !slices.Equal(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(residual) || target.HealthData.Cur != hp {
						t.Fatalf("normal prefix=%t/%t events=%v want=%v marker=%d/%d carry=%g HP=%d", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21), target.HealthData.Cur)
					}
					if blocked && len(damages) != 0 || !blocked && !slices.Equal(damages, []int32{amount}) {
						t.Fatalf("normal prefix HP tail=%v", damages)
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalElectricPrefixLive4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, stage := range []string{"exclude", "direction"} {
			for _, marker := range []uint32{0, 42, 99} {
				t.Run(fmt.Sprintf("%s/marker-%d", stage, marker), func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, r := normalElectricPrefixFixture4E17B0(t, playerSource, selfWeapon, &events, &damages)
					live := &PlayerUpdateData{State: PlayerState16, Field21: math.Float32bits(-0.5), Field76: 31, Field75: 33}
					if marker != 99 {
						// A replacement observer must not restart an already consumed prefix.
						live.Player = &Player{Field3680: 3, CameraFollowObj: source, ArmorEquip: 0x1000000, WeaponEquip: 0x400}
					}
					previous := source.PrevPos
					swap := func() {
						target.UpdateData = unsafe.Pointer(live)
						cached.Field76, cached.Field75 = marker, 777
					}
					exclusion := func(v *Object) bool {
						if v != source || cached.Field76 != 0 {
							t.Fatal("normal live exclusion did not consume the entry reset")
						}
						events = append(events, "exclude")
						source.PrevPos = types.Ptf(99, 77)
						if stage == "exclude" {
							swap()
						}
						return false
					}
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || pos != previous {
							t.Fatal("normal live facing lost the entry snapshot")
						}
						events = append(events, "direction")
						if stage == "direction" {
							swap()
						}
						return false
					}
					r.ObserveClear = func(*Object) { t.Fatal("normal prefix re-queried the replacement observer") }
					r.ElectricArmorScale = func(v *Object) float32 {
						if v != target || target.UpdateDataPlayer() != live || cached.Field76 != marker || cached.Field75 != 777 {
							t.Fatal("normal electric tail reset the marker or ignored the replacement record")
						}
						events = append(events, "scale")
						return 0.5
					}
					wantMarker, wantType := marker, uint32(777)
					if marker == 0 {
						wantMarker, wantType = 2, uint32(typ)
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
						if v != target || a != source || w != weapon || d != 2 || gotType != typ || cached.Field76 != wantMarker || cached.Field75 != wantType {
							t.Fatal("normal live tail lost cached-marker/live-carry separation")
						}
						events = append(events, "default")
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r)
					if !h || !result || !slices.Equal(events, []string{"exclude", "direction", "scale", "default"}) || cached.Field21 != math.Float32bits(0.5) || live.Field21 != 0 || live.Field76 != 31 || live.Field75 != 33 || target.HealthData.Cur != 60 {
						t.Fatalf("normal live prefix=%t/%t events=%v cached/live carry=%g/%g", h, result, events, math.Float32frombits(cached.Field21), math.Float32frombits(live.Field21))
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalElectricPrefixMissing4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, missing := range []string{"exclusion", "direction", "scale", "default", "Quest", "armor"} {
			t.Run(missing, func(t *testing.T) {
				var events []string
				var damages []int32
				target, source, weapon, cached, r := normalElectricPrefixFixture4E17B0(t, playerSource, selfWeapon, &events, &damages)
				exclusion := func(*Object) bool { events = append(events, "exclude"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
				r.BlockDirection = func(*Object, types.Pointf) bool { events = append(events, "direction"); return false }
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("missing service reached default")
					return false
				}
				switch missing {
				case "exclusion":
					if selfWeapon {
						r.BlockSourceExcluded = nil
					} else {
						r.BlockSourceOnlyExcluded = nil
					}
				case "direction":
					r.BlockDirection = nil
				case "scale":
					r.ElectricArmorScale = nil
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
						t.Fatal("missing service report lost the signed original arguments")
					}
					reason = why
				}
				before, beforeCached, beforePlayer, beforeSource, beforeHealth := *target, *cached, *cached.Player, *source, *target.HealthData
				h, result := PlayerDamageNative4E17B0(target, source, weapon, -5, typ, r)
				if h || result || reason == "" || len(events) != 0 || len(damages) != 0 || *target != before || *cached != beforeCached || *cached.Player != beforePlayer || *source != beforeSource || *target.HealthData != beforeHealth {
					t.Fatalf("missing normal prefix service=%t/%t reason=%q events=%v marker=%d", h, result, reason, events, cached.Field76)
				}
			})
		}
	})
}

func TestPlayerDamageNormalElectricPrefixEarly4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, gate := range []string{"NoUpdate", "dead", "observer", "Coop", "invulnerable"} {
			t.Run(gate, func(t *testing.T) {
				var events []string
				var damages []int32
				target, source, weapon, cached, _ := normalElectricPrefixFixture4E17B0(t, playerSource, selfWeapon, &events, &damages)
				r := PlayerDamageRuntime4E17B0{Frame: func() uint32 { return 1400 }}
				switch gate {
				case "NoUpdate":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "observer":
					cached.Player.Field3680 = 1
				case "Coop":
					r.CoopMode, source.ObjOwner = func() bool { return true }, target
					if playerSource {
						// Owner-chain resolution stops at the first Player, so a
						// Player source is a self hit only when it IS the target.
						source = target
						if selfWeapon {
							weapon = target
						}
					}
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				}
				before, beforeCached, beforePlayer, beforeSource := *target, *cached, *cached.Player, *source
				h, result := PlayerDamageNative4E17B0(target, source, weapon, -5, typ, r)
				if !h || result != (gate == "invulnerable") || *target != before || *cached != beforeCached || *cached.Player != beforePlayer || *source != beforeSource {
					t.Fatalf("normal electric early gate=%t/%t marker=%d", h, result, cached.Field76)
				}
			})
		}
	})
}

func TestPlayerDamageNormalElectricPrefixInvalid4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, stage := range []string{"exclude", "direction"} {
			for _, invalid := range []string{"class", "nil update"} {
				t.Run(stage+"/"+invalid, func(t *testing.T) {
					var events []string
					var damages []int32
					target, source, weapon, cached, r := normalElectricPrefixFixture4E17B0(t, playerSource, selfWeapon, &events, &damages)
					invalidate := func() {
						if invalid == "class" {
							target.ObjClass = object.ClassMonster
						} else {
							target.UpdateData = nil
						}
					}
					exclusion := func(*Object) bool {
						if cached.Field76 != 0 {
							t.Fatal("normal invalidation preceded the entry reset")
						}
						events = append(events, "exclude")
						if stage == "exclude" {
							invalidate()
						}
						return false
					}
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, exclusion
					r.BlockDirection = func(*Object, types.Pointf) bool {
						events = append(events, "direction")
						if stage == "direction" {
							invalidate()
						}
						return false
					}
					r.ElectricArmorScale = func(*Object) float32 { t.Fatal("invalid live normal record reached scale"); return 1 }
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("invalid live normal record reached default")
						return false
					}
					reason := ""
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, weapon, -5, typ, r)
					if h || result || reason != "unsupported live player electric record" || !slices.Equal(events, []string{"exclude", "direction"}) || cached.Field76 != 0 || cached.Field75 != 77 || cached.Field21 != math.Float32bits(0.5) || target.HealthData.Cur != 60 {
						t.Fatalf("normal invalid record=%t/%t reason=%q events=%v marker=%d", h, result, reason, events, cached.Field76)
					}
				})
			}
		}
	})
}
