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

func playerDamagePossessionMagicCases4E17B0(t *testing.T, fn func(*testing.T, object.DamageType, bool, bool)) {
	t.Helper()
	for _, typ := range []object.DamageType{object.DamageFlame, object.DamageExplosion} {
		for _, playerSource := range []bool{false, true} {
			for _, splash := range []bool{false, true} {
				t.Run(fmt.Sprintf("type-%d/player-source-%t/splash-%t", typ, playerSource, splash), func(t *testing.T) {
					fn(t, typ, playerSource, splash)
				})
			}
		}
	}
}

func TestPlayerDamagePossessionMagic4E17B0(t *testing.T) {
	playerDamagePossessionMagicCases4E17B0(t, func(t *testing.T, typ object.DamageType, playerSource, splash bool) {
		for _, defense := range []string{"none", "Reflect Shield", "ordinary shield", "GreatSword", "excluded shield", "rear Reflect Shield"} {
			t.Run(defense, func(t *testing.T) {
				target, source, weapon, missile := damageFlameFixture4E17B0(t, playerSource, true, splash)
				missile.ObjSubClass = 0x10 // Ordinary shield absorbs; GreatSword still reflects.
				cached := target.UpdateDataPlayer()
				cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
				cached.Field57, cached.Field21 = math.Float32bits(0.25), math.Float32bits(0.125)
				cached.Field76, cached.Field75 = 88, 77
				if defense == "ordinary shield" || defense == "excluded shield" {
					cached.State, cached.Player.ArmorEquip = PlayerState16, 0x1000000
				}
				if defense == "GreatSword" {
					cached.Player.WeaponEquip = 0x400
				}
				live := &PlayerUpdateData{Player: &Player{Field3680: 2, CameraFollowObj: missile}, State: PlayerState13, Field57: math.Float32bits(0.5), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				previousPosition, missileType := missile.PrevPos, uint32(missile.TypeInd)
				prefixMarker, prefixType := uint32(1), missileType
				if splash {
					prefixMarker, prefixType = 0, 77
				}
				reflectShield := defense == "Reflect Shield" || defense == "rear Reflect Shield"
				var events []string
				r := damageFlameRuntime4E17B0(t, 0)
				r.ObserveClear = func(v *Object) {
					if v != target || target.UpdateDataPlayer() != cached || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("ObserveClear must follow only the cached marker reset")
					}
					events = append(events, "observe")
					cached.Field57 = math.Float32bits(0.9)
					cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0, 0
					target.UpdateData = unsafe.Pointer(live)
					if reflectShield {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
				}
				exclusion := func(w *Object) bool {
					before := []string{"observe"}
					if defense == "rear Reflect Shield" {
						before = append(before, "direction")
					}
					if w != missile || !slices.Equal(events, before) || cached.Field76 != prefixMarker || cached.Field75 != prefixType {
						t.Fatal("exclusion preceded ObserveClear/attribution or ran twice")
					}
					events = append(events, "exclude")
					missile.TypeInd++
					missile.PrevPos = types.Ptf(-72, 87)
					return defense == "excluded shield"
				}
				wrongExclusion := func(*Object) bool { t.Fatal("wrong source-only/weapon exclusion set"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclusion, wrongExclusion
				if splash {
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = wrongExclusion, exclusion
				}
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || target.UpdateDataPlayer() != live {
						t.Fatal("facing preceded live update replacement")
					}
					if reflectShield && len(events) == 1 {
						if pos != missile.PosVec || cached.Field76 != 0 {
							t.Fatal("Reflect must use current position before attribution")
						}
					} else {
						before := []string{"observe", "exclude"}
						if defense == "rear Reflect Shield" {
							before = []string{"observe", "direction", "exclude"}
						}
						if !slices.Equal(events, before) || pos != previousPosition {
							t.Fatal("ordinary facing lost its snapshot or ran twice")
						}
					}
					events = append(events, "direction")
					return defense == "Reflect Shield" || defense == "ordinary shield" || defense == "GreatSword"
				}
				r.ProjectileReflect = func(w, v *Object) {
					if w != missile || v != target {
						t.Fatal("reflection arguments")
					}
					events = append(events, "reflect")
				}
				r.ClearOwner = func(w *Object) {
					if w != missile {
						t.Fatal("clear owner")
					}
					events = append(events, "clear")
				}
				r.SetOwner = func(v, w *Object) {
					if v != target || w != missile {
						t.Fatal("set owner")
					}
					events = append(events, "owner")
				}
				r.Audio = func(id int, v *Object) {
					want := map[string]int{"Reflect Shield": 122, "ordinary shield": 878, "GreatSword": 890}[defense]
					if v != target || id != want {
						t.Fatal("defense audio")
					}
					events = append(events, "audio")
				}
				r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
				r.PlayerSetState = func(v *Object, state PlayerState) bool {
					if v != target || state != PlayerState19 {
						t.Fatal("GreatSword state")
					}
					events = append(events, "state")
					live.State = state
					return true
				}
				r.Melee.RandomInt = func(min, max int) int {
					if min != 18 || max != 20 {
						t.Fatal("GreatSword RNG")
					}
					events = append(events, "rng")
					return 19
				}
				blockItem := &Object{HealthData: &HealthData{Cur: 30}, ObjFlags: object.FlagEquipped}
				if defense == "ordinary shield" {
					blockItem.ObjClass, blockItem.ObjSubClass, target.InvFirstItem = object.ClassArmor, 2, blockItem
				} else if defense == "GreatSword" {
					blockItem.ObjClass, blockItem.ObjSubClass, target.InvFirstItem = object.ClassWeapon, 0x400, blockItem
				}
				r.CanDamageBlockItem = func(item *Object) bool { return item == blockItem }
				r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
				wear := func(item, v, a, w *Object, amount float32, gotType object.DamageType) bool {
					if item != blockItem || v != target || a != source || w != missile || amount != 2.5 || gotType != typ || cached.Field76 != prefixMarker || cached.Field75 != prefixType {
						t.Fatal("block wear lost cached prefix or source")
					}
					events = append(events, "wear")
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
				tail := r.DefaultDamage
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					events = append(events, "default")
					return tail(v, a, w, d, gotType)
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r)
				want := []string{"observe", "exclude", "direction", "default"}
				switch defense {
				case "Reflect Shield":
					want = []string{"observe", "direction", "reflect", "clear", "owner", "audio"}
				case "ordinary shield":
					want = []string{"observe", "exclude", "direction", "audio", "balance", "wear"}
				case "GreatSword":
					want = []string{"observe", "exclude", "direction", "reflect", "clear", "owner", "audio", "rng", "state", "balance", "wear"}
				case "excluded shield":
					want = []string{"observe", "exclude", "default"}
				case "rear Reflect Shield":
					want = []string{"observe", "direction", "exclude", "direction", "default"}
				}
				wantDamage := defense == "none" || defense == "excluded shield" || defense == "rear Reflect Shield"
				wantHP, wantCarry := uint16(200), float32(0.5)
				if wantDamage {
					wantHP = 195
					if typ == object.DamageExplosion {
						wantHP, wantCarry = 196, 0.25
					}
					if splash {
						prefixMarker, prefixType = 2, uint32(typ)
					}
				} else if defense == "Reflect Shield" {
					prefixMarker, prefixType = 0, 77
				}
				if !h || result != wantDamage || !slices.Equal(events, want) || target.HealthData.Cur != wantHP || live.Field21 != math.Float32bits(wantCarry) || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != prefixMarker || cached.Field75 != prefixType || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("magic possession=%t/%t HP=%d marker=%d/%d events=%v want=%v", h, result, target.HealthData.Cur, cached.Field76, cached.Field75, events, want)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionMagicAdmission4E17B0(t *testing.T) {
	playerDamagePossessionMagicCases4E17B0(t, func(t *testing.T, typ object.DamageType, playerSource, splash bool) {
		for _, service := range []string{"observe", "direction", "exclusion", "default", "Quest scale"} {
			t.Run(service, func(t *testing.T) {
				target, source, weapon, missile := damageFlameFixture4E17B0(t, playerSource, true, splash)
				ud := target.UpdateDataPlayer()
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, missile
				ud.Field76, ud.Field75 = 88, 77
				r := damageFlameRuntime4E17B0(t, 0)
				r.ObserveClear = func(*Object) { t.Fatal("missing service reached ObserveClear") }
				r.BlockDirection = func(*Object, types.Pointf) bool { return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = func(*Object) bool { return false }, func(*Object) bool { return false }
				switch service {
				case "observe":
					r.ObserveClear = nil
				case "direction":
					r.BlockDirection = nil
				case "exclusion":
					if splash {
						r.BlockSourceOnlyExcluded = nil
					} else {
						r.BlockSourceExcluded = nil
					}
				case "default":
					r.DefaultDamage = nil
				case "Quest scale":
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = nil
				}
				var reason string
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r); h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("magic admission=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionMagicEarlyGates4E17B0(t *testing.T) {
	playerDamagePossessionMagicCases4E17B0(t, func(t *testing.T, typ object.DamageType, playerSource, splash bool) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon, missile := damageFlameFixture4E17B0(t, playerSource, true, splash)
				ud := target.UpdateDataPlayer()
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, missile
				ud.Field76, ud.Field75 = 88, 77
				r := damageFlameRuntime4E17B0(t, 0)
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
					missile.ObjOwner = target
					if !splash {
						source = target
					}
					r.CoopMode = func() bool { return true }
				}
				r.ObserveClear = func(*Object) { t.Fatal("early gate cleared possession") }
				r.BlockSourceExcluded = func(*Object) bool { t.Fatal("early gate reached exclusion"); return false }
				r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate reached facing"); return false }
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || id != 71 || v != target {
						t.Fatal("unexpected early audio")
					}
				}
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r); !h || result != (gate == "invulnerable") || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("magic early gate=%t/%t", h, result)
				}
			})
		}
	})
}
