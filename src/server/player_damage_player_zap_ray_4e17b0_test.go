package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Entry contracts are independent of the HP implementation. The C-owned
// companion calls the registered PlayerDamage and real DefaultDamage/HP tail.
func playerZapRayFixture4E17B0(t *testing.T, owner string, observe bool) (*Object, *Object, *Object, *PlayerUpdateData, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, unit, cached, r := chargePrefixFixture4E17B0(t, observe)
	ray := &Object{TypeInd: 801, ObjClass: object.ClassLight | object.ClassSimple | object.ClassImmobile | object.ClassVisibleEnable, ObjFlags: object.FlagAirborne, PrevPos: types.Ptf(20, 7), PosVec: types.Ptf(-20, 3)}
	source := ray
	switch owner {
	case "Player":
		source, ray.ObjOwner = unit, unit
	case "NPC":
		unit.ObjClass, unit.ObjSubClass, unit.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(new(MonsterUpdateData))
		source, ray.ObjOwner = unit, unit
	case "world":
	default:
		t.Fatal("unknown ray owner")
	}
	r.SentryGlobeType = 999 // Admission is the stock class, not a cached type ID.
	r.Frame = func() uint32 { t.Fatal("unexecuted invulnerability timestamp read"); return 0 }
	r.ItemArmorValue = func(*Object) float32 { t.Fatal("ray entered armor wear"); return 0 }
	r.ElectricArmorScale = func(*Object) float32 { t.Fatal("ray entered electric armor scale"); return 0 }
	r.BerserkShieldBlock = func(*Object) bool { t.Fatal("type 16 entered the melee berserker block"); return false }
	return target, source, ray, cached, r
}

// Bind the historical positive Sentry contracts to the same production
// DefaultDamage body used by the native adapter. These service callbacks do
// not bypass defense, attribution, sound, GameBall, hurt, Shield or HP ordering.
func bindPlayerZapRayDefault4E17B0(r *PlayerDamageRuntime4E17B0) {
	r.DefaultDamage = func(target, source, weapon *Object, damage int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(target, source, weapon, damage, typ, DefaultDamageWorldRuntime4E0B30{
			Frame: r.Frame, GameplayFlag1: r.GameplayFlag1, QuestMode: r.QuestMode, IsEnemy: r.IsEnemy,
			Audio: r.Audio, BuffOff: r.BuffOff, MonsterHasHitSound: func(*Object) bool { return false },
			PlayerDamageSound: r.PlayerDamageSound, PlayerDamageSoundC: r.PlayerDamageSoundC,
			GameBallType: r.GameBallType, GameBallOnDamage: r.GameBallOnDamage,
			BalanceFloatInd: r.BalanceFloatInd, PlayerSetState: r.PlayerSetState, AdjustHP: r.AdjustHP,
			VampirismFX: r.VampirismFX, ShieldReduce: r.ShieldReduce,
			CanApplyLateDefend: r.CanApplyLateDefend, ApplyLateDefend: r.ApplyLateDefend,
			DamageClear: r.DamageClear, Unsupported: r.Unsupported,
		})
	}
}

func TestPlayerDamagePlayerZapRaySignedQuest4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, observe := range []bool{false, true} {
			for _, raw := range []int32{-7, 0, 1, 19, 20, 500} {
				for _, scale := range []float32{-1, 0, 0.5, 1.25} {
					for _, god := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/observe-%t/raw-%d/scale-%g/god-%t", owner, observe, raw, scale, god), func(t *testing.T) {
							target, source, ray, cached, r := playerZapRayFixture4E17B0(t, owner, observe)
							var events []string
							clear := r.ObserveClear
							r.ObserveClear = func(o *Object) { clear(o); events = append(events, "observe") }
							r.BlockSourceExcluded = func(o *Object) bool {
								if o != ray || cached.Field76 != 0 || cached.Field75 != 77 {
									t.Fatal("marker before exclusions")
								}
								events = append(events, "excluded")
								return false
							}
							r.BlockDirection = func(o *Object, pos types.Pointf) bool {
								if o != target || pos != ray.PrevPos {
									t.Fatal("ray facing arguments")
								}
								events = append(events, "direction")
								return false
							}
							r.GodMode = func() bool { events = append(events, "god"); return god }
							r.QuestMode = func() bool { events = append(events, "quest"); return scale != -1 }
							r.QuestDamageScale = func() float32 { events = append(events, "scale"); return scale }
							wantDamage := raw
							if scale != -1 {
								// GAME.EXE spills the product to binary32, then uses FISTP.
								wantDamage = int32(math.RoundToEven(float64(float32(float64(scale) * float64(raw)))))
								if raw > 0 && wantDamage < 1 {
									wantDamage = 1
								}
							}
							r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
								if v != target || a != source || w != ray || d != wantDamage || typ != object.DamageZapRay {
									t.Fatalf("DefaultDamage ray arguments: %d want %d", d, wantDamage)
								}
								events = append(events, "default")
								return false
							}
							beforeHP, beforeSource, beforeRay := *target.HealthData, *source, *ray
							h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
							want := []string{"excluded", "direction", "god"}
							if observe {
								want = append([]string{"observe"}, want...)
							}
							if !god {
								want = append(want, "quest")
								if scale != -1 {
									want = append(want, "scale")
								}
								want = append(want, "default")
							}
							marker, markerType := uint32(1), uint32(801)
							if source == ray {
								marker, markerType = 2, 16
							}
							if !h || result != god || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) || cached.Field57 != math.Float32bits(0.25) || *target.HealthData != beforeHP || *source != beforeSource || *ray != beforeRay {
								t.Fatalf("ray=%t/%t order=%v want=%v marker=%d/%d", h, result, events, want, cached.Field76, cached.Field75)
							}
							if observe && (cached.Player.Field3680 != 0x20 || cached.Player.ObserveTarget() != nil) {
								t.Fatal("possession not returned by the executed prefix")
							}
						})
					}
				}
			}
		}
	}
}

func TestPlayerDamagePlayerZapRayCachedLive4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, observe := range []bool{false, true} {
			for _, stage := range []string{"observe", "reflect", "excluded", "direction", "scale"} {
				t.Run(fmt.Sprintf("%s/observe-%t/%s", owner, observe, stage), func(t *testing.T) {
					target, source, ray, cached, r := playerZapRayFixture4E17B0(t, owner, observe)
					live := &PlayerUpdateData{Player: &Player{}, State: PlayerState1, Field76: 31, Field75: 33, Field21: math.Float32bits(-0.5)}
					mutate := func() { target.UpdateData = unsafe.Pointer(live) }
					clear := r.ObserveClear
					r.ObserveClear = func(o *Object) {
						clear(o)
						if stage == "observe" {
							mutate()
						}
						ray.PrevPos = types.Ptf(30, 7)
					}
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					pos := ray.PrevPos
					if observe {
						pos = types.Ptf(30, 7)
					}
					directions := 0
					r.BlockDirection = func(_ *Object, got types.Pointf) bool {
						directions++
						if directions == 1 {
							if got != ray.PosVec || cached.Field76 != 0 {
								t.Fatal("Reflect position/prefix")
							}
							if stage == "reflect" {
								mutate()
							}
						} else {
							if got != pos {
								t.Fatal("facing lost pre-exclusion position snapshot")
							}
							if stage == "direction" {
								mutate()
							}
						}
						return false
					}
					r.BlockSourceExcluded = func(*Object) bool {
						if stage == "excluded" {
							mutate()
						}
						ray.TypeInd, ray.PrevPos = 888, types.Ptf(-100, 99)
						cached.Player.ArmorEquip, cached.Player.WeaponEquip = 0x3000000, 0x400
						return false
					}
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = func() float32 {
						if stage == "scale" {
							mutate()
						}
						return 0.5
					}
					calls := 0
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						calls++
						if v != target || a != source || w != ray || d != 10 || typ != object.DamageZapRay {
							t.Fatal("cached/live default arguments")
						}
						return true
					}
					h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
					marker, markerType := uint32(1), uint32(888)
					if source == ray {
						marker, markerType = 2, 16
					}
					if !h || !result || calls != 1 || directions != 2 || cached.Field76 != marker || cached.Field75 != markerType || live.Field76 != 31 || live.Field75 != 33 || live.Field21 != math.Float32bits(-0.5) || cached.Field21 != math.Float32bits(0.125) {
						t.Fatalf("cached/live=%t/%t default=%d facing=%d marker=%d/%d", h, result, calls, directions, cached.Field76, cached.Field75)
					}
					if (stage != "observe" || observe) && target.UpdateDataPlayer() != live {
						t.Fatal("live update replacement lost")
					}
				})
			}
		}
	}
}

func TestPlayerDamagePlayerZapRayReflectEarly4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, observe := range []bool{false, true} {
			for _, raw := range []int32{-7, 0, 500} {
				t.Run(fmt.Sprintf("%s/observe-%t/raw-%d", owner, observe, raw), func(t *testing.T) {
					target, source, ray, cached, r := playerZapRayFixture4E17B0(t, owner, observe)
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					var events []string
					r.BlockDirection = func(_ *Object, pos types.Pointf) bool {
						if pos != ray.PosVec || cached.Field76 != 0 {
							t.Fatal("Reflect branch snapshot")
						}
						ray.ObjClass = object.ClassMissile // Must not retroactively choose the missile branch.
						events = append(events, "direction")
						return true
					}
					r.PointFX = func(id int, pos types.Pointf) {
						if id != 132 || pos != target.PosVec {
							t.Fatal("Reflect FX")
						}
						events = append(events, "fx")
					}
					r.Audio = func(id int, o *Object) {
						if id != 122 || o != target {
							t.Fatal("Reflect audio")
						}
						events = append(events, "audio")
					}
					r.BlockSourceExcluded, r.DefaultDamage, r.QuestMode, r.GodMode = nil, nil, nil, nil
					h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
					if !h || result || !reflect.DeepEqual(events, []string{"direction", "fx", "audio"}) || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatalf("Reflect=%t/%t events=%v marker=%d/%d", h, result, events, cached.Field76, cached.Field75)
					}
				})
			}
		}
	}
}

func TestPlayerDamagePlayerZapRayShieldCachedFlagsLiveItem4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, observe := range []bool{false, true} {
			for _, raw := range []int32{-7, 0, 19} {
				for _, kind := range []string{"nil", "intact", "broken"} {
					t.Run(fmt.Sprintf("%s/observe-%t/raw-%d/%s", owner, observe, raw, kind), func(t *testing.T) {
						target, source, ray, cached, r := playerZapRayFixture4E17B0(t, owner, observe)
						cached.Player.ArmorEquip = 0x1000000
						shield := &Object{ObjSubClass: 2, ObjFlags: object.FlagEquipped}
						var events []string
						r.BlockDirection = func(*Object, types.Pointf) bool {
							cached.State = PlayerState16
							cached.Player.ArmorEquip = 0
							return true
						}
						r.Audio = func(id int, o *Object) {
							if id != 878 || o != target {
								t.Fatal("block audio")
							}
							events = append(events, "audio")
							ray.ObjClass = object.ClassMissile
						}
						r.ProjectileReflect = func(w, v *Object) {
							if w != ray || v != target {
								t.Fatal("live missile reflection")
							}
							events = append(events, "reflect")
						}
						r.ClearOwner = func(w *Object) {
							if w != ray {
								t.Fatal("owner clear")
							}
							events = append(events, "clear")
						}
						r.SetOwner = func(v, w *Object) {
							if v != target || w != ray {
								t.Fatal("owner set")
							}
							events = append(events, "owner")
						}
						r.BlockDamagePercent = func() float64 {
							events = append(events, "balance")
							if kind != "nil" {
								target.InvFirstItem = shield
							}
							return 0.25
						}
						r.CanDamageBlockItem = func(o *Object) bool {
							if o != shield {
								t.Fatal("stale shield")
							}
							events = append(events, "admit")
							return true
						}
						r.DamageBlockItem = func(item, v, a, w *Object, amount float32, typ object.DamageType) bool {
							want := shield
							if kind == "nil" {
								want = nil
							}
							if item != want || v != target || a != source || w != ray || amount != float32(raw)*0.25 || typ != object.DamageZapRay {
								t.Fatal("block wear arguments")
							}
							events = append(events, "wear")
							if kind == "broken" {
								shield.ObjFlags |= object.FlagDestroyed
							}
							return true
						}
						r.PlayerSetState = func(v *Object, state PlayerState) bool {
							if v != target || state != PlayerState13 {
								t.Fatal("broken shield state")
							}
							events = append(events, "state")
							return true
						}
						r.DefaultDamage, r.GodMode, r.QuestMode = nil, nil, nil
						h, result := PlayerDamageNative4E17B0(target, source, ray, raw, object.DamageZapRay, r)
						want := []string{"audio", "reflect", "clear", "owner", "balance"}
						if kind != "nil" {
							want = append(want, "admit")
						}
						want = append(want, "wear")
						if kind == "broken" {
							want = append(want, "state")
						}
						marker, markerType := uint32(1), uint32(801)
						if source == ray {
							marker, markerType = 0, 77
						}
						if !h || result || !reflect.DeepEqual(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != math.Float32bits(0.125) {
							t.Fatalf("shield=%t/%t events=%v want=%v marker=%d/%d", h, result, events, want, cached.Field76, cached.Field75)
						}
					})
				}
			}
		}
	}
}

func TestPlayerDamagePlayerZapRayAtUseFailures4E17B0(t *testing.T) {
	for _, observe := range []bool{false, true} {
		for _, fault := range []string{"observe", "excluded", "direction", "quest", "default", "live-update", "live-class", "live-ray"} {
			t.Run(fmt.Sprintf("observe-%t/%s", observe, fault), func(t *testing.T) {
				target, source, ray, cached, r := playerZapRayFixture4E17B0(t, "Player", observe)
				marker, markerType := uint32(1), uint32(801)
				switch fault {
				case "observe":
					r.ObserveClear = nil
					marker, markerType = 0, 77
				case "excluded":
					r.BlockSourceExcluded = nil
					marker, markerType = 0, 77
				case "direction":
					r.BlockDirection = nil
				case "quest":
					r.QuestMode, r.QuestDamageScale = func() bool { return true }, nil
				case "default":
					r.DefaultDamage = nil
				case "live-update":
					r.BlockDirection = func(*Object, types.Pointf) bool { target.UpdateData = nil; return false }
				case "live-class":
					r.BlockDirection = func(*Object, types.Pointf) bool {
						target.ObjClass, target.UpdateData = object.ClassSimple, nil
						return false
					}
				case "live-ray":
					r.BlockSourceExcluded = func(*Object) bool { ray.ObjClass = object.ClassWeapon; return false }
				}
				if r.DefaultDamage == nil && fault != "default" {
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("fault entered HP tail")
						return false
					}
				}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				if fault == "observe" && !observe {
					r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool { return true }
					marker, markerType = 1, 801
				}
				h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
				unused := fault == "observe" && !observe
				if h != unused || result != unused || (reason == "") != unused || cached.Field76 != marker || cached.Field75 != markerType || target.HealthData.Cur != 200 || target.Obj130 != nil {
					t.Fatalf("fault=%t/%t reason=%s marker=%d/%d", h, result, reason, cached.Field76, cached.Field75)
				}
			})
		}
	}
}

func TestPlayerDamagePlayerZapRayEntryGates4E17B0(t *testing.T) {
	for _, owner := range []string{"world", "Player", "NPC"} {
		for _, gate := range []string{"dead", "no-update", "invulnerable", "player-disabled", "coop-self"} {
			t.Run(owner+"/"+gate, func(t *testing.T) {
				target, source, ray, cached, r := playerZapRayFixture4E17B0(t, owner, true)
				calls, sounds := 0, 0
				r.Frame = func() uint32 { calls++; return 1400 }
				r.Audio = func(id int, o *Object) {
					if id != 71 || o != target {
						t.Fatal("invulnerability sound")
					}
					sounds++
				}
				switch gate {
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "no-update":
					target.ObjFlags |= object.FlagNoUpdate
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "player-disabled":
					cached.Player.Field3680 |= 1
				case "coop-self":
					if owner == "Player" {
						// A player is already the terminal player owner; its
						// ObjOwner does not replace it in this lookup.
						source = target
					} else {
						source.ObjOwner = target
					}
					if source.FindOwnerChainPlayer() != target {
						t.Fatal("fixture is not a cooperative self-hit")
					}
					r.CoopMode = func() bool { return true }
				}
				beforeTarget, beforeCached, beforePlayer := *target, *cached, *cached.Player
				h, result := PlayerDamageNative4E17B0(target, source, ray, -7, object.DamageZapRay, r)
				want := gate == "invulnerable"
				if !h || result != want || (calls == 1) != want || (sounds == 1) != want || *target != beforeTarget || *cached != beforeCached || *cached.Player != beforePlayer {
					t.Fatalf("gate=%t/%t clock=%d sound=%d", h, result, calls, sounds)
				}
			})
		}
	}
}

func TestPlayerDamagePlayerZapRayAdmission4E17B0(t *testing.T) {
	for _, shape := range []string{"nil-source", "nil-weapon", "no-simple", "no-immobile", "ray-player", "ray-npc", "ray-weapon", "ray-wand", "ray-missile", "source-simple", "source-nil-update", "source-weapon", "source-missile"} {
		t.Run(shape, func(t *testing.T) {
			target, source, ray, cached, r := playerZapRayFixture4E17B0(t, "Player", false)
			r.Frame = func() uint32 { return 1400 }
			r.SentryGlobeType = ray.TypeInd // Type equality must not admit a non-stock class.
			switch shape {
			case "nil-source":
				source = nil
			case "nil-weapon":
				ray = nil
			case "no-simple":
				ray.ObjClass &^= object.ClassSimple
			case "no-immobile":
				ray.ObjClass &^= object.ClassImmobile
			case "ray-player":
				ray.ObjClass |= object.ClassPlayer
			case "ray-npc":
				ray.ObjClass |= object.ClassMonster
			case "ray-weapon":
				ray.ObjClass |= object.ClassWeapon
			case "ray-wand":
				ray.ObjClass |= object.ClassWand
			case "ray-missile":
				ray.ObjClass |= object.ClassMissile
			case "source-simple":
				source.ObjClass = object.ClassSimple
			case "source-nil-update":
				source.UpdateData = nil
			case "source-weapon":
				source.ObjClass |= object.ClassWeapon
			case "source-missile":
				source.ObjClass |= object.ClassMissile
			}
			beforeTarget, beforeCached, beforePlayer := *target, *cached, *cached.Player
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("unsupported shape entered default tail")
				return false
			}
			h, result := PlayerDamageNative4E17B0(target, source, ray, 19, object.DamageZapRay, r)
			if h || result || reason == "" || *target != beforeTarget || *cached != beforeCached || *cached.Player != beforePlayer {
				t.Fatalf("admission=%t/%t reason=%s", h, result, reason)
			}
		})
	}
}
