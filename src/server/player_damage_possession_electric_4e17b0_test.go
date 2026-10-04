package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func possessionElectricFixture4E17B0(t *testing.T, playerSource, selfWeapon bool) (target, source, weapon *Object) {
	t.Helper()
	target, source = defaultDamageElectricSelfFixture4E0B30(t, true, playerSource)
	if selfWeapon {
		weapon = source
	}
	source.PosVec = types.Ptf(20, 10)
	ud := target.UpdateDataPlayer()
	ud.Player.Field3680, ud.Player.CameraFollowObj = 2, source
	ud.Field76, ud.Field75, ud.Field21 = 99, 77, math.Float32bits(0.125)
	return
}

func TestPlayerDamagePossessionElectric4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, defense := range []string{"none", "ordinary equipment", "Reflect front", "Reflect rear", "excluded", "Reflect removed"} {
			t.Run(defense, func(t *testing.T) {
				target, source, weapon := possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
				cached := target.UpdateDataPlayer()
				cached.State = PlayerState13
				if defense == "ordinary equipment" || defense == "excluded" {
					cached.State, cached.Player.ArmorEquip, cached.Player.WeaponEquip = PlayerState16, 0x1000000, 0x400
				}
				if defense == "Reflect removed" {
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
				}
				live := &PlayerUpdateData{Player: &Player{Field3680: 3, CameraFollowObj: source, ArmorEquip: 0x1000000, WeaponEquip: 0x8000}, State: PlayerState21, Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
				var events []string
				var damages []int32
				r := playerDamageRuntime4E17B0(t, nil, &damages)
				r.ObserveClear = func(v *Object) {
					if v != target || target.UpdateDataPlayer() != cached || cached.Field76 != 0 || cached.Field75 != 77 || len(events) != 0 {
						t.Fatal("electric ObserveClear must follow only the cached marker reset")
					}
					events = append(events, "observe")
					target.UpdateData = unsafe.Pointer(live)
					if defense == "Reflect front" || defense == "Reflect rear" {
						target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					}
					if defense == "Reflect removed" {
						target.Buffs &^= 1 << playerDamageReflectEnchant4E17B0
					}
				}
				reflectPresent := typ == object.DamageAirborneElectric && (defense == "Reflect front" || defense == "Reflect rear")
				reflected := reflectPresent && defense == "Reflect front"
				previous := source.PrevPos
				exclude := func(w *Object) bool {
					before := []string{"observe"}
					if reflectPresent {
						before = append(before, "Reflect direction")
					}
					if w != source || !slices.Equal(events, before) || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatal("electric exclusion preceded ObserveClear/Reflect or attributed a self weapon")
					}
					events = append(events, "exclude")
					source.TypeInd++
					source.PrevPos = types.Ptf(-72, 87)
					return defense == "excluded"
				}
				wrong := func(*Object) bool { t.Fatal("wrong four/six electric exclusion set"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, wrong
				if !selfWeapon {
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = wrong, exclude
				}
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || target.UpdateDataPlayer() != live || cached.Field76 != 0 {
						t.Fatal("electric facing preceded ObserveClear or changed self-weapon marker")
					}
					if reflectPresent && len(events) == 1 {
						if pos != source.PosVec {
							t.Fatal("Reflect did not use current source position")
						}
						events = append(events, "Reflect direction")
						return reflected
					}
					if pos != previous {
						t.Fatal("electric facing lost previous-position snapshot")
					}
					events = append(events, "direction")
					return true // ordinary equipment must never block type 9/17.
				}
				r.Audio = func(id int, v *Object) {
					if !reflected || id != 122 || v != target {
						t.Fatal("unexpected electric block audio")
					}
					events = append(events, "audio")
				}
				r.ElectricArmorScale = func(v *Object) float32 {
					if v != target || target.UpdateDataPlayer() != live || cached.Field76 != 0 {
						t.Fatal("electric scale preceded the consumed prefix")
					}
					events = append(events, "electric scale")
					return 0.5
				}
				r.DefaultDamage = func(v, a, w *Object, d int32, gotType object.DamageType) bool {
					if v != target || a != source || w != weapon || d != 3 || gotType != typ || cached.Field76 != 2 || cached.Field75 != uint32(typ) {
						t.Fatal("electric entry lost live carry or cached marker fallback")
					}
					events = append(events, "default")
					return DefaultDamageWorld4E0B30(v, a, w, d, gotType, DefaultDamageWorldRuntime4E0B30{
						Frame: r.Frame, GameplayFlag1: func() bool { return true }, IsEnemy: r.IsEnemy,
						ElectricProtection: func(*Object) float64 { events = append(events, "protection"); return 0.25 },
						MonsterHasHitSound: func(*Object) bool { return false }, PlayerSetState: func(*Object, PlayerState) bool { return true },
						DamageClear: r.DamageClear, Unsupported: r.Unsupported,
					})
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r)
				want := []string{"observe"}
				if reflectPresent {
					want = append(want, "Reflect direction")
				}
				marker, markerType, carry, hp := uint32(2), uint32(typ), float32(0), uint16(58)
				if reflected {
					want = append(want, "audio")
					marker, markerType, carry, hp = 0, 77, 0.5, 60
				} else {
					want = append(want, "exclude")
					if defense != "excluded" {
						want = append(want, "direction")
					}
					want = append(want, "electric scale", "default", "protection")
				}
				if !h || result == reflected || !slices.Equal(events, want) || target.HealthData.Cur != hp || live.Field21 != math.Float32bits(carry) || cached.Field21 != math.Float32bits(0.125) || cached.Field76 != marker || cached.Field75 != markerType || live.Field76 != 31 || live.Field75 != 33 {
					t.Fatalf("electric possession=%t/%t HP=%d marker=%d/%d events=%v want=%v", h, result, target.HealthData.Cur, cached.Field76, cached.Field75, events, want)
				}
				if !reflected && (!slices.Equal(damages, []int32{2}) || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ)) {
					t.Fatal("electric default damage/attribution")
				}
			})
		}
	})
}

func TestPlayerDamagePossessionElectricAdmission4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, service := range []string{"observe", "exclusion", "direction", "default", "electric scale", "Quest scale", "armor"} {
			t.Run(service, func(t *testing.T) {
				target, source, weapon := possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
				ud := target.UpdateDataPlayer()
				r := PlayerDamageRuntime4E17B0{
					ObserveClear:        func(*Object) { t.Fatal("missing electric service cleared possession") },
					BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
					BlockDirection: func(*Object, types.Pointf) bool { return false },
					DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("missing electric service entered damage tail")
						return false
					},
					ElectricArmorScale: func(*Object) float32 { t.Fatal("missing electric service ran scale"); return 1 },
					QuestDamageScale:   func() float32 { return 1 },
				}
				switch service {
				case "observe":
					r.ObserveClear = nil
				case "exclusion":
					if selfWeapon {
						r.BlockSourceExcluded = nil
					} else {
						r.BlockSourceOnlyExcluded = nil
					}
				case "direction":
					r.BlockDirection = nil
				case "default":
					r.DefaultDamage = nil
				case "electric scale":
					r.ElectricArmorScale = nil
				case "Quest scale":
					r.QuestMode, r.QuestDamageScale = func() bool { return true }, nil
				case "armor":
					target.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped}
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r); h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("electric admission=%t/%t reason=%q", h, result, reason)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionElectricEarlyGates4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon := possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
				ud := target.UpdateDataPlayer()
				r := PlayerDamageRuntime4E17B0{Frame: func() uint32 { return 1400 }, ObserveClear: func(*Object) { t.Fatal("electric early gate cleared possession") }, ElectricArmorScale: func(*Object) float32 { t.Fatal("electric early gate ran scale"); return 1 }}
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
					if selfWeapon {
						weapon = target
					}
					r.CoopMode = func() bool { return true }
				}
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || id != 71 || v != target {
						t.Fatal("electric early audio")
					}
				}
				before, beforeUD, beforePlayer := *target, *ud, *ud.Player
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r); !h || result != (gate == "invulnerable") || *target != before || *ud != beforeUD || *ud.Player != beforePlayer {
					t.Fatalf("electric early gate=%t/%t", h, result)
				}
			})
		}
	})
}

func TestPlayerDamagePossessionElectricLiveRecordGuard4E17B0(t *testing.T) {
	electricPlayerCases4E17B0(t, func(t *testing.T, playerSource, selfWeapon bool, typ object.DamageType) {
		for _, stage := range []string{"ObserveClear", "direction"} {
			for _, change := range []string{"class", "nil update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, weapon := possessionElectricFixture4E17B0(t, playerSource, selfWeapon)
					cached := target.UpdateDataPlayer()
					invalidate := func() {
						if change == "class" {
							target.ObjClass = object.ClassMonster
						} else {
							target.UpdateData = nil
						}
					}
					reason := ""
					r := PlayerDamageRuntime4E17B0{
						ObserveClear: func(*Object) {
							if stage == "ObserveClear" {
								invalidate()
							}
						},
						BlockSourceExcluded: func(*Object) bool { return false }, BlockSourceOnlyExcluded: func(*Object) bool { return false },
						BlockDirection:     func(*Object, types.Pointf) bool { invalidate(); return false },
						ElectricArmorScale: func(*Object) float32 { t.Fatal("invalid record reached electric scale"); return 1 },
						DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
							t.Fatal("invalid record reached electric tail")
							return false
						},
						Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why },
					}
					if h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, typ, r); h || result || reason == "" || cached.Field76 != 0 || cached.Field75 != 77 || cached.Field21 != math.Float32bits(0.125) || target.HealthData.Cur != 60 {
						t.Fatalf("electric live guard=%t/%t reason=%q", h, result, reason)
					}
				})
			}
		}
	})
}
