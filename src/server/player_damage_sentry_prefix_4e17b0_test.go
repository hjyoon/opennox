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

// Independent GAME.EXE service-order contracts for positive player-owned
// SentryGlobe ZAP_RAY, not synthetic HP writes or stock-map/GUI evidence.
// 004E18C4/004E18EB precede Reflect, 004E1A49 snapshots PrevPos, 004E1AA2
// attributes the live distinct weapon, and case 16 bypasses armor/carry.
func sentryPrefixFixture4E17B0(t *testing.T, observe bool) (*Object, *Object, *Object, *PlayerUpdateData, PlayerDamageRuntime4E17B0) {
	t.Helper()
	target, source, cached, r := chargePrefixFixture4E17B0(t, observe)
	weapon := &Object{ObjClass: object.ClassSimple | object.ClassImmobile, TypeInd: 71, PrevPos: types.Ptf(20, 0), PosVec: types.Ptf(-20, 0)}
	r.SentryGlobeType = weapon.TypeInd
	r.ItemArmorValue = func(*Object) float32 { t.Fatal("ZAP_RAY entered armor wear"); return 0 }
	r.ElectricArmorScale = func(*Object) float32 { t.Fatal("ZAP_RAY entered electric armor scale"); return 0 }
	r.PlayerDamageSound = func(*Object, *Object) {}
	return target, source, weapon, cached, r
}

func TestPlayerDamageSentryPrefixOrder4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, shielded := range []bool{false, true} {
			for _, facing := range []string{"front", "rear", "excluded"} {
				for _, raw := range []int32{1, 19, 20} {
					t.Run(fmt.Sprintf("shield-%t/%s/raw-%d", shielded, facing, raw), func(t *testing.T) {
						target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
						shield := &Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2, InitData: unsafe.Pointer(new(ModifierInitData))}
						if shielded {
							cached.Player.ArmorEquip, target.InvFirstItem = 0x1000000, shield
						}
						var events []string
						wantPos := weapon.PrevPos
						if observe {
							wantPos = types.Ptf(30, 7)
						}
						clear := r.ObserveClear
						r.ObserveClear = func(obj *Object) { clear(obj); events = append(events, "observe"); weapon.PrevPos = wantPos }
						r.BlockSourceExcluded = func(obj *Object) bool {
							if obj != weapon || cached.Field76 != 0 || cached.Field75 != 77 {
								t.Fatal("distinct Sentry marker was not cleared before the six exclusions")
							}
							events = append(events, "excluded")
							weapon.PrevPos, weapon.TypeInd = types.Ptf(-100, 99), 888
							cached.Player.ArmorEquip, cached.Player.WeaponEquip, cached.State = 0, 0x400, PlayerState16
							return facing == "excluded"
						}
						r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
							if obj != target || pos != wantPos || cached.Field76 != 1 || cached.Field75 != 888 {
								t.Fatal("facing must use the position snapshot and post-exclusion Sentry type")
							}
							events = append(events, "direction")
							return facing == "front"
						}
						r.CanDamageBlockItem = func(item *Object) bool {
							if item != shield {
								t.Fatal("wrong live shield")
							}
							events = append(events, "block-admission")
							return true
						}
						r.Audio = func(id int, obj *Object) {
							if id != 878 || obj != target {
								t.Fatal("wrong shield sound")
							}
							events = append(events, "block-audio")
						}
						r.BlockDamagePercent = func() float64 { events = append(events, "block-percent"); return 0.25 }
						r.DamageBlockItem = func(item, owner, attacker, attack *Object, amount float32, typ object.DamageType) bool {
							if item != shield || owner != target || attacker != source || attack != weapon || amount != float32(raw)*0.25 || typ != object.DamageZapRay {
								t.Fatal("ZAP_RAY shield arguments")
							}
							events = append(events, "block-wear")
							return true
						}
						r.GodMode = func() bool { events = append(events, "god"); return false }
						r.QuestMode = func() bool { events = append(events, "quest"); return false }
						r.GameplayFlag1 = func() bool { events = append(events, "gameplay"); return true }
						r.BuffOff = func(obj *Object, id EnchantID) {
							if obj != target || id != playerDamageInvisibleEnchant4E17B0 {
								t.Fatal("BuffOff arguments")
							}
							events = append(events, "buff")
						}
						r.PlayerDamageSound = func(obj, attack *Object) {
							if obj != target || attack != weapon {
								t.Fatal("sound arguments")
							}
							events = append(events, "sound")
						}
						r.GameBallOnDamage = func(attacker, obj *Object, amount int32) {
							if attacker != source || obj != target || amount != raw {
								t.Fatal("GameBall arguments")
							}
							events = append(events, "ball")
						}
						r.PlayerSetState = func(obj *Object, state PlayerState) bool {
							if obj != target || state != PlayerState30 {
								t.Fatal("hurt arguments")
							}
							events = append(events, "hurt")
							return true
						}
						r.DamageClear = func(obj *Object, amount int32) {
							if obj != target || amount != raw {
								t.Fatalf("HP argument %d, want %d", amount, raw)
							}
							events = append(events, "hp")
						}
						blocked := shielded && facing == "front"
						if blocked {
							// None of these unexecuted HP-tail services may be a
							// prerequisite for an intact shield's early return.
							r.GameplayFlag1, r.IsEnemy, r.DefaultDamage, r.DamageClear, r.BuffOff, r.QuestDamageScale = nil, nil, nil, nil, nil, nil
						} else {
							bindPlayerZapRayDefault4E17B0(&r)
						}
						h, result := PlayerDamageNative4E17B0(target, source, weapon, raw, object.DamageZapRay, r)
						want := []string{"excluded"}
						if observe {
							want = append([]string{"observe"}, want...)
						}
						if facing != "excluded" {
							want = append(want, "direction")
						}
						if blocked {
							want = append(want, "block-audio", "block-percent", "block-admission", "block-wear")
						} else {
							want = append(want, "god", "quest", "gameplay", "buff", "sound", "ball")
							if raw >= 20 {
								want = append(want, "hurt")
							}
							want = append(want, "hp")
						}
						if !h || result == blocked || !reflect.DeepEqual(events, want) || cached.Field76 != 1 || cached.Field75 != 888 || cached.Field21 != math.Float32bits(0.125) || *target.HealthData != (HealthData{Cur: 200, Field2: 200, Max: 200}) {
							t.Fatalf("Sentry order=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, want, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21))
						}
					})
				}
			}
		}
	})
}

func TestPlayerDamageSentryPrefixReflect4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
		target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
		var events []string
		clear := r.ObserveClear
		r.ObserveClear = func(obj *Object) { clear(obj); events = append(events, "observe") }
		r.BlockDirection = func(obj *Object, pos types.Pointf) bool {
			if obj != target || pos != weapon.PosVec || cached.Field76 != 0 || cached.Field75 != 77 || cached.Player.ObserveTarget() != nil {
				t.Fatal("Reflect precedes cleared/returned player prefix")
			}
			events = append(events, "reflect-direction")
			return true
		}
		r.BlockSourceExcluded = func(*Object) bool { t.Fatal("Reflect called exclusions"); return false }
		r.PointFX = func(id int, pos types.Pointf) {
			if id != 132 || pos != target.PosVec {
				t.Fatal("Reflect point FX")
			}
			events = append(events, "fx")
		}
		r.Audio = func(id int, obj *Object) {
			if id != 122 || obj != target {
				t.Fatal("Reflect sound")
			}
			events = append(events, "audio")
		}
		r.GameplayFlag1, r.IsEnemy, r.PlayerSetState, r.QuestDamageScale = nil, nil, nil, nil
		r.QuestMode = func() bool { t.Fatal("blocked Reflect queried Quest"); return false }
		h, result := PlayerDamageNative4E17B0(target, source, weapon, 500, object.DamageZapRay, r)
		want := []string{"reflect-direction", "fx", "audio"}
		if observe {
			want = append([]string{"observe"}, want...)
		}
		if !h || result || !reflect.DeepEqual(events, want) || cached.Field76 != 0 || cached.Field75 != 77 {
			t.Fatalf("Reflect=%t/%t events=%v marker=%d/%d", h, result, events, cached.Field76, cached.Field75)
		}
	})
}

// DefaultDamage 004E1136/004E1147 reloads live class/update/state after
// BuffOff, late Defend, damage sound and GameBall, not the entry state.
func TestPlayerDamageSentryPrefixLiveHurt4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, stage := range []string{"buff", "defend", "sound", "ball"} {
			for _, state := range []PlayerState{PlayerState1, PlayerState15, PlayerState13} {
				t.Run(fmt.Sprintf("%s/state-%d", stage, state), func(t *testing.T) {
					target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
					live := &PlayerUpdateData{Player: &Player{}, State: state, Field21: math.Float32bits(-0.5), Field76: 31, Field75: 33}
					mutate := func() {
						cached.State = PlayerState13
						if state == PlayerState13 {
							cached.State = PlayerState1
						}
						target.UpdateData = unsafe.Pointer(live)
					}
					r.BuffOff = func(*Object, EnchantID) {
						if stage == "buff" {
							mutate()
						}
					}
					modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
					target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, modifier}})}
					r.CanApplyLateDefend = func(m *ModifierEff) bool { return m == modifier }
					r.ApplyLateDefend = func(_ *ModifierEff, _, _, _, _ *Object, amount int32, typ object.DamageType) int32 {
						if stage == "defend" {
							mutate()
						}
						return amount
					}
					r.PlayerDamageSound = func(*Object, *Object) {
						if stage == "sound" {
							mutate()
						}
					}
					r.GameBallOnDamage = func(_, _ *Object, _ int32) {
						if stage == "ball" {
							mutate()
						}
					}
					hurt, hp := 0, 0
					r.PlayerSetState = nil
					if state == PlayerState13 {
						r.PlayerSetState = func(obj *Object, got PlayerState) bool {
							if obj != target || got != PlayerState30 || obj.UpdateDataPlayer() != live {
								t.Fatal("hurt used cached player update")
							}
							hurt++
							return true
						}
					}
					r.DamageClear = func(obj *Object, amount int32) {
						if obj != target || amount != 20 {
							t.Fatal("HP arguments")
						}
						hp++
					}
					bindPlayerZapRayDefault4E17B0(&r)
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 20, object.DamageZapRay, r)
					wantHurt := 0
					if state == PlayerState13 {
						wantHurt = 1
					}
					if !h || !result || hurt != wantHurt || hp != 1 || cached.Field76 != 1 || cached.Field75 != 71 || live.Field76 != 31 || live.Field75 != 33 || live.Field21 != math.Float32bits(-0.5) || cached.Field21 != math.Float32bits(0.125) {
						t.Fatalf("live hurt=%t/%t hurt=%d want=%d hp=%d cached=%d/%d live=%d/%d", h, result, hurt, wantHurt, hp, cached.Field76, cached.Field75, live.Field76, live.Field75)
					}
				})
			}
		}
	})
}

func TestPlayerDamageSentryPrefixLiveModes4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, mode := range []string{"quest", "god", "non-player-god", "friendly"} {
			t.Run(mode, func(t *testing.T) {
				target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
				changed, hp, scaleCalls := false, int32(-1), 0
				wantMarker, wantType := uint32(1), uint32(71)
				if mode == "non-player-god" {
					wantMarker, wantType = 0, 77
				}
				r.BlockSourceExcluded = func(*Object) bool {
					changed = true
					if mode == "non-player-god" {
						target.ObjClass, target.UpdateData = object.ClassSimple, nil
					}
					return false
				}
				r.GodMode = func() bool {
					if !changed || cached.Field76 != wantMarker {
						t.Fatal("GodMode before the distinct marker")
					}
					return mode == "god" || mode == "non-player-god"
				}
				r.QuestMode = func() bool {
					if !changed || cached.Field76 != wantMarker {
						t.Fatal("Quest read before the executed prefix")
					}
					return mode == "quest"
				}
				r.QuestDamageScale = func() float32 { scaleCalls++; return 0.25 }
				r.GameplayFlag1 = func() bool { return mode != "friendly" }
				r.IsEnemy = func(*Object, *Object) bool { return false }
				r.DamageClear = func(_ *Object, amount int32) { hp = amount }
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				bindPlayerZapRayDefault4E17B0(&r)
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 19, object.DamageZapRay, r)
				if mode == "non-player-god" {
					// The bounded stock ray slice cannot interpret a callback's
					// replacement non-player record as PlayerUpdateData.
					if h || result || reason != "unsupported live player ZAP_RAY tail record" || hp != -1 || scaleCalls != 0 || cached.Field76 != 0 || cached.Field75 != 77 {
						t.Fatalf("live boundary=%t/%t reason=%s HP=%d marker=%d/%d", h, result, reason, hp, cached.Field76, cached.Field75)
					}
					return
				}
				wantHP, wantScale := int32(19), 0
				if mode == "quest" {
					wantHP, wantScale = 5, 1
				}
				if mode == "god" || mode == "friendly" {
					wantHP = -1
				}
				if !h || !result || hp != wantHP || scaleCalls != wantScale || cached.Field76 != wantMarker || cached.Field75 != wantType || cached.Field21 != math.Float32bits(0.125) {
					t.Fatalf("live modes=%t/%t HP=%d want=%d scale=%d marker=%d/%d", h, result, hp, wantHP, scaleCalls, cached.Field76, cached.Field75)
				}
			})
		}
	})
}

func TestPlayerDamageSentryPrefixMissingServices4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, missing := range []string{"observe", "exclude", "direction"} {
			t.Run(missing, func(t *testing.T) {
				target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
				before, beforeCached, beforePlayer := *target, *cached, *cached.Player
				switch missing {
				case "observe":
					r.ObserveClear = nil
				case "exclude":
					r.BlockSourceExcluded = nil
				case "direction":
					r.BlockDirection = nil
				}
				hp := 0
				r.DamageClear = func(*Object, int32) { hp++ }
				var reason string
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				bindPlayerZapRayDefault4E17B0(&r)
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 19, object.DamageZapRay, r)
				if missing == "observe" && !observe {
					if !h || !result || hp != 1 || reason != "" {
						t.Fatalf("unexecuted ObserveClear was required: %t/%t/%s", h, result, reason)
					}
				} else {
					// Availability is checked at the executed service, not before
					// the original marker/ObserveClear prefix.
					beforeCached.Field76 = 0
					wantReason := "missing player ZAP_RAY observe service"
					if missing == "exclude" {
						wantReason = "missing player ZAP_RAY exclusion service"
					}
					if missing == "direction" {
						wantReason = "missing player ZAP_RAY block direction service"
						beforeCached.Field76, beforeCached.Field75 = 1, 71
					}
					if observe && missing != "observe" {
						beforePlayer.Field3680 &^= 2
						beforePlayer.CameraFollowObj = nil
					}
					if h || result || hp != 0 || reason != wantReason || *target != before || *cached != beforeCached || *cached.Player != beforePlayer {
						t.Fatalf("missing prefix=%t/%t reason=%s HP calls=%d", h, result, reason, hp)
					}
				}
			})
		}
	})
}

func TestPlayerDamageSentryPrefixLiveHurtFault4E17B0(t *testing.T) {
	for _, fault := range []string{"nil-update", "missing-service", "non-player"} {
		t.Run(fault, func(t *testing.T) {
			target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, false)
			r.GameBallOnDamage = func(_, _ *Object, _ int32) {
				switch fault {
				case "nil-update":
					target.UpdateData = nil
				case "non-player":
					target.ObjClass, target.UpdateData = object.ClassSimple, nil
				}
			}
			// A function value in the by-value runtime cannot be removed by a
			// callback; select the absent service on entry and reach its use.
			if fault == "missing-service" {
				r.PlayerSetState = nil
			}
			hp, reason := 0, ""
			r.DamageClear = func(*Object, int32) { hp++ }
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			bindPlayerZapRayDefault4E17B0(&r)
			h, result := PlayerDamageNative4E17B0(target, source, weapon, 20, object.DamageZapRay, r)
			if fault == "non-player" {
				if !h || !result || reason != "" || hp != 1 {
					t.Fatalf("non-player hurt=%t/%t reason=%s hp=%d", h, result, reason, hp)
				}
			} else {
				want := "unsupported live player hurt update"
				if fault == "missing-service" {
					want = "unsupported live player hurt-state service"
				}
				if !h || !result || reason != want || hp != 0 || cached.Field76 != 1 || cached.Field75 != 71 || target.Obj130 != weapon || target.Field131 != uint32(object.DamageZapRay) || target.Frame134 != 700 {
					t.Fatalf("live fault=%t/%t reason=%s marker=%d/%d attribution=%p hp=%d", h, result, reason, cached.Field76, cached.Field75, target.Obj130, hp)
				}
			}
		})
	}
}

func TestPlayerDamageSentryPrefixLiveClassBoundary4E17B0(t *testing.T) {
	bitePrefixCases4E17B0(t, func(t *testing.T, observe bool) {
		for _, class := range []object.Class{object.ClassMonster, object.ClassWeapon, object.ClassWand} {
			t.Run(fmt.Sprint(class), func(t *testing.T) {
				target, source, weapon, cached, r := sentryPrefixFixture4E17B0(t, observe)
				r.BlockSourceExcluded = func(*Object) bool { weapon.ObjClass = class; return false }
				hp, reason := 0, ""
				r.DamageClear = func(*Object, int32) { hp++ }
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 19, object.DamageZapRay, r)
				if h || result || reason != "unsupported live player ZAP_RAY tail record" || hp != 0 || cached.Field76 != 1 || cached.Field75 != 71 || cached.Field21 != math.Float32bits(0.125) || cached.Player.ObserveTarget() != nil || target.Obj130 != nil || *target.HealthData != (HealthData{Cur: 200, Field2: 200, Max: 200}) {
					t.Fatalf("live class=%v handled=%t/%t reason=%s hp=%d marker=%d/%d", class, h, result, reason, hp, cached.Field76, cached.Field75)
				}
			})
		}
	})
}
