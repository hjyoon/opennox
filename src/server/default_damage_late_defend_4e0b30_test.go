package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultDamageLateEffectFixture4E0B30() *ModifierEff {
	return &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
}

// 004E0F5D reloads target class/subclass after protection/position/BuffOff;
// 004E0F77 then calls the live inventory helper before attribution stores.
func TestDefaultDamageWorld4E0B30LateDefendAfterEarlierCallbacks(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, stage := range []string{"protection", "buff off"} {
			t.Run(map[bool]string{false: "NPC/", true: "player/"}[player]+stage, func(t *testing.T) {
				target, source := damageMeleeUnitFixture4E17B0(t, player), damageMeleeUnitFixture4E17B0(t, true)
				staleEffect, liveEffect := defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30()
				stale := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped,
					InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, staleEffect}})}
				// No class, HP, update or damage callback is required by 004E1320.
				live := &Object{ObjFlags: object.FlagEquipped,
					InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, liveEffect}})}
				target.InvFirstItem = stale
				target.Obj130, target.Field131, target.Frame134 = stale, 99, 77
				var events []string
				r := damageMeleeWorldRuntime4E0B30(t)
				r.ElectricProtection = func(*Object) float64 {
					events = append(events, "protection")
					if stage == "protection" {
						target.InvFirstItem = live
					}
					return 0.5
				}
				r.BuffOff = func(unit *Object, id EnchantID) {
					if unit != target || id != defaultDamageInvisibleEnchant4E0B30 {
						t.Fatal("BuffOff arguments")
					}
					events = append(events, "buff")
					if stage == "buff off" {
						target.InvFirstItem = live
					}
				}
				r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
				r.ApplyLateDefend = func(m *ModifierEff, item, owner, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
					if m != liveEffect || item != live || owner != target || weapon != nil || attacker != source || d != 4 || typ != object.DamageAirborneElectric ||
						target.Pos132 != source.PrevPos || target.Obj130 != stale || target.Field131 != 99 || target.Frame134 != 77 {
						t.Fatal("late Defend used stale records or ran outside the original hit prefix")
					}
					events = append(events, "live defend")
					return -7
				}
				r.DamageClear = func(owner *Object, d int32) {
					if owner != target || d != -7 || owner.Obj130 != source || owner.Field131 != uint32(object.DamageAirborneElectric) || owner.Frame134 != 1400 {
						t.Fatal("signed damage or attribution order")
					}
					events = append(events, "HP")
				}
				if !DefaultDamageWorld4E0B30(target, source, nil, 8, object.DamageAirborneElectric, r) {
					t.Fatal("supported electric tail returned false")
				}
				if want := []string{"protection", "buff", "live defend", "HP"}; !slices.Equal(events, want) {
					t.Fatalf("events=%v, want %v", events, want)
				}
			})
		}
	}
}

func TestDefaultDamageWorld4E0B30LateDefendLiveSlotsLinksAndFlags(t *testing.T) {
	for _, player := range []bool{false, true} {
		t.Run(map[bool]string{false: "NPC", true: "player"}[player], func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, player), damageMeleeUnitFixture4E17B0(t, true)
			first, stale, third, nextEffect := defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30()
			base := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, stale}}
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(base)}
			next := &Object{InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nextEffect}})}
			detached := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, stale}})}
			item.InvNextItem, target.InvFirstItem = detached, item
			var calls []*ModifierEff
			r := damageMeleeWorldRuntime4E0B30(t)
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(m *ModifierEff, gotItem, owner, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
				if owner != target || weapon != nil || attacker != source || typ != object.DamageClaw {
					t.Fatal("late Defend native argument order")
				}
				calls = append(calls, m)
				switch m {
				case first:
					if gotItem != item || d != 3 {
						t.Fatal("first late Defend arguments")
					}
					base.Modifiers[3] = third
					item.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, stale, stale}})
					item.InvNextItem, next.ObjFlags = next, object.FlagEquipped
					target.InvFirstItem = nil // The helper must not recapture its head.
					return 0
				case third:
					if gotItem != item || d != 0 {
						t.Fatal("cached init base/live slot three or zero copyback")
					}
					return -7
				case nextEffect:
					if gotItem != next || d != -7 {
						t.Fatal("live next/equipped flags or negative copyback")
					}
					return math.MinInt32
				default:
					t.Fatal("detached or replaced modifier was called")
					return d
				}
			}
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != math.MinInt32 {
					t.Fatalf("final raw damage DWORD=%#x", uint32(d))
				}
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) || !slices.Equal(calls, []*ModifierEff{first, third, nextEffect}) {
				t.Fatalf("ordered live effects=%v", calls)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LateDefendLiveTargetGate(t *testing.T) {
	for _, test := range []struct {
		name                                             string
		player, entryNPC, liveMonster, liveNPC, wantCall bool
	}{
		{"NPC loses subclass", false, true, true, false, false},
		{"ordinary monster becomes NPC", false, false, true, true, true},
		{"NPC becomes non-unit", false, true, false, true, false},
		{"player becomes ordinary monster", true, false, true, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, test.player), damageMeleeUnitFixture4E17B0(t, true)
			if !test.entryNPC {
				target.ObjSubClass = 0
			}
			m := defaultDamageLateEffectFixture4E0B30()
			target.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, m}})}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.BuffOff = func(*Object, EnchantID) {
				target.ObjClass, target.ObjSubClass = 0, 0
				if test.liveMonster {
					target.ObjClass = object.ClassMonster
					if test.player {
						// A class change needs a valid new native update layout.
						target.UpdateData = unsafe.Pointer(&MonsterUpdateData{})
					}
				}
				if test.liveNPC {
					target.ObjSubClass = 0x10
				}
			}
			calls := 0
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(got *ModifierEff, _, _, _, _ *Object, d int32, typ object.DamageType) int32 {
				if got != m || d != 3 || typ != object.DamageClaw || !test.wantCall {
					t.Fatal("stale target-class gate")
				}
				calls++
				return 9
			}
			r.DamageClear = func(*Object, int32) {}
			if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) || calls != map[bool]int{false: 0, true: 1}[test.wantCall] {
				t.Fatalf("live class/subclass calls=%d, want call=%t", calls, test.wantCall)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LateDefendAdmissionAndLiveRejection(t *testing.T) {
	t.Run("flags-only admission", func(t *testing.T) {
		target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
		target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped,
			InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, defaultDamageLateEffectFixture4E0B30()}})}
		before := *target
		r := damageMeleeWorldRuntime4E0B30(t)
		var reason string
		r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
		r.BuffOff = func(*Object, EnchantID) { t.Fatal("unported admission reached the late prefix") }
		r.DamageClear = func(*Object, int32) { t.Fatal("unported admission reached HP") }
		if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) || reason != "monster defense callbacks" || *target != before {
			t.Fatalf("flags-only admission=%q", reason)
		}
	})
	t.Run("introduced unknown slot continues", func(t *testing.T) {
		target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
		first, stale, unknown, nextEffect := defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30(), defaultDamageLateEffectFixture4E0B30()
		base := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, stale}}
		item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(base)}
		next := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nextEffect}})}
		target.InvFirstItem = item
		var events []string
		r := damageMeleeWorldRuntime4E0B30(t)
		r.CanApplyLateDefend = func(m *ModifierEff) bool { return m != unknown }
		r.ApplyLateDefend = func(m *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 {
			switch m {
			case first:
				base.Modifiers[3], item.InvNextItem = unknown, next
				events = append(events, "first")
				return -7
			case nextEffect:
				if d != -7 {
					t.Fatal("unknown effect changed damage")
				}
				events = append(events, "next")
				return 1
			default:
				t.Fatal("unported/stale callback entered")
				return d
			}
		}
		r.Unsupported = func(reason string, owner, attacker, weapon *Object, d int32, typ object.DamageType) {
			if reason != "unsupported live late equipped-item defend effect" || owner != target || attacker != source || weapon != nil || d != -7 || typ != object.DamageClaw {
				t.Fatalf("live rejection=%q damage=%d type=%d", reason, d, typ)
			}
			events = append(events, "reject")
		}
		r.DamageClear = func(owner *Object, d int32) {
			if owner != target || d != 1 {
				t.Fatal("supported continuation damage")
			}
			events = append(events, "HP")
		}
		if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) || !slices.Equal(events, []string{"first", "reject", "next", "HP"}) {
			t.Fatalf("live continuation=%v", events)
		}
	})
}

func TestDefaultDamageWorld4E0B30LateDefendNilInitFaultAtUse(t *testing.T) {
	for _, npc := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary monster skips helper", true: "NPC faults at helper"}[npc], func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			if !npc {
				target.ObjSubClass = 0
			}
			target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped}
			target.Field131, target.Frame134 = 99, 77
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.BuffOff = func(*Object, EnchantID) { events = append(events, "buff") }
			r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
			var caught any
			func() {
				defer func() { caught = recover() }()
				if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) {
					t.Fatal("tail result")
				}
			}()
			if npc {
				if caught == nil || !slices.Equal(events, []string{"buff"}) || target.Field131 != 99 || target.Frame134 != 77 || target.Pos132 != source.PrevPos {
					t.Fatalf("original fault prefix: caught=%v events=%v", caught, events)
				}
			} else if caught != nil || !slices.Equal(events, []string{"buff", "HP"}) {
				t.Fatalf("ordinary monster unexpectedly entered helper: caught=%v events=%v", caught, events)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30LateDefendLiveGameBallGuard(t *testing.T) {
	for _, test := range []struct {
		name     string
		ballType uint16
		drop     bool
		reason   string
	}{
		{"missing live type", 0, false, "unsupported live GameBall type"},
		{"missing live drop", 77, false, "unsupported live GameBall drop"},
		{"supported drop before HP", 77, true, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, true)
			m := defaultDamageLateEffectFixture4E0B30()
			r := damageMeleeWorldRuntime4E0B30(t)
			r.GameBallType = test.ballType
			r.BuffOff = func(*Object, EnchantID) {
				target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped,
					InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, m}})}
			}
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32 {
				target.Field129 = &Object{TypeInd: 77}
				return 30
			}
			var reason string
			var events []string
			r.Unsupported = func(s string, _, _, _ *Object, d int32, _ object.DamageType) {
				if d != 30 {
					t.Fatal("live guard lost post-Defend damage")
				}
				reason = s
			}
			if test.drop {
				r.GameBallOnDamage = func(attacker, owner *Object, d int32) {
					if attacker != source || owner != target || d != 30 || test.reason != "" {
						t.Fatal("GameBall drop args")
					}
					events = append(events, "drop")
				}
			}
			r.PlayerSetState = func(owner *Object, state PlayerState) bool {
				if owner != target || state != PlayerState30 || test.reason != "" {
					t.Fatal("hurt tail args")
				}
				events = append(events, "hurt")
				return true
			}
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != 30 || test.reason != "" {
					t.Fatal("unsupported or incorrect HP tail")
				}
				events = append(events, "HP")
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 3, object.DamageClaw, r) || reason != test.reason {
				t.Fatalf("live guard=%q, want %q", reason, test.reason)
			}
			var want []string
			if test.reason == "" {
				want = []string{"drop", "hurt", "HP"}
			}
			if !slices.Equal(events, want) || target.Obj130 != source || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageClaw) || target.HealthData.Cur != 200 {
				t.Fatalf("late guard order/prefix: events=%v attribution=%p frame=%d HP=%d", events, target.Obj130, target.Frame134, target.HealthData.Cur)
			}
		})
	}
}
