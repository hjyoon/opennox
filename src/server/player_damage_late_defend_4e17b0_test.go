package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func playerDamageLateEffectFixture4E17B0() *ModifierEff {
	return &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
}

// 004E17B0 reaches the default tail only after armor/GodMode/Quest. In that
// tail, 004E0F77 calls 004E1320 after BuffOff, not an entry-time effect plan.
func TestPlayerDamageNative4E17B0LateDefendAfterEarlierCallbacks(t *testing.T) {
	for _, stage := range []string{"armor report", "quest scale", "buff off"} {
		t.Run(stage, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			target.UpdateDataPlayer().Field57 = math.Float32bits(0.5)
			staleEffect, liveEffect := playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0()
			stale := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, staleEffect}})}
			// The helper's equipped gate does not require a weapon/armor class,
			// health, update data, or an item damage callback.
			live := &Object{ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, liveEffect}})}
			armor := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
				InitData: unsafe.Pointer(&ModifierInitData{}), UpdateData: unsafe.Pointer(&WeaponArmorUpdateData{}),
				HealthData: &HealthData{Cur: 20, Max: 20}, Damage: unsafe.Pointer(new(byte)), InvNextItem: stale}
			target.InvFirstItem = armor
			var events []string
			r := playerDamageRuntime4E17B0(t, sound, new([]int32))
			r.ItemArmorValue = func(*Object) float32 { return 0.5 }
			r.CanDamageArmor = func(item *Object) bool { return item == armor }
			r.DamageArmor = func(item, attacker, weapon *Object, d int32, typ object.DamageType) bool {
				if item != armor || attacker != source || weapon != source || d != 4 || typ != object.DamageBite {
					t.Fatal("armor call arguments")
				}
				events = append(events, "armor")
				item.HealthData.Cur -= uint16(d)
				return true
			}
			install := func() { target.InvFirstItem = live }
			r.ReportArmorHealth = func(*Object, *Object, uint16, uint16) {
				events = append(events, "report")
				if stage == "armor report" {
					install()
				}
			}
			r.QuestMode = func() bool { return true }
			r.QuestDamageScale = func() float32 {
				events = append(events, "quest")
				if stage == "quest scale" {
					install()
				}
				return 1
			}
			r.BuffOff = func(*Object, EnchantID) {
				events = append(events, "buff")
				if stage == "buff off" {
					install()
				}
			}
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(m *ModifierEff, item, owner, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
				if m != liveEffect || item != live || owner != target || weapon != source || attacker != source || d != 4 || typ != object.DamageBite || target.UpdateDataPlayer().Field76 != 2 || target.Pos132 != source.PrevPos {
					t.Fatal("late Defend used a stale entry-time record or preceded the hit prefix")
				}
				events = append(events, "live defend")
				return -7
			}
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != -7 {
					t.Fatalf("final signed damage=%d, want -7", d)
				}
				events = append(events, "HP")
			}
			if h, result := PlayerDamageNative4E17B0(target, source, source, 8, object.DamageBite, r); !h || !result {
				t.Fatalf("handled/result=%t/%t", h, result)
			}
			if want := []string{"armor", "report", "quest", "buff", "live defend", "HP"}; !slices.Equal(events, want) {
				t.Fatalf("events=%v, want %v", events, want)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0LateDefendLiveSlotsLinksAndFlags(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	first, stale, third, nextEffect := playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0()
	base := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, stale}}
	item := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(base)}
	next := &Object{InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nextEffect}})}
	detached := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, stale}})}
	item.InvNextItem, target.InvFirstItem = detached, item
	var calls []*ModifierEff
	r := playerDamageRuntime4E17B0(t, sound, new([]int32))
	r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
	r.ApplyLateDefend = func(m *ModifierEff, gotItem, owner, weapon, attacker *Object, d int32, typ object.DamageType) int32 {
		if owner != target || weapon != source || attacker != source || typ != object.DamageBite {
			t.Fatal("late Defend argument order")
		}
		calls = append(calls, m)
		switch m {
		case first:
			if gotItem != item || d != 3 {
				t.Fatal("first late Defend arguments")
			}
			base.Modifiers[3] = third
			item.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, stale, stale}})
			item.InvNextItem = next
			next.ObjFlags = object.FlagEquipped
			target.InvFirstItem = nil // Do not recapture the head within 004E1320.
			return 0
		case third:
			if gotItem != item || d != 0 {
				t.Fatal("slot three was snapshotted or damage was clamped")
			}
			return -7
		case nextEffect:
			if gotItem != next || d != -7 {
				t.Fatal("next link/flags were snapshotted or negative damage was skipped")
			}
			return math.MinInt32
		default:
			t.Fatal("detached/replaced modifier was called")
			return d
		}
	}
	r.DamageClear = func(owner *Object, d int32) {
		if owner != target || d != math.MinInt32 {
			t.Fatalf("final raw DWORD damage=%#x", uint32(d))
		}
	}
	if h, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, r); !h || !result {
		t.Fatalf("handled/result=%t/%t", h, result)
	}
	if !slices.Equal(calls, []*ModifierEff{first, third, nextEffect}) {
		t.Fatalf("live calls=%v", calls)
	}
}

func TestPlayerDamageNative4E17B0LateDefendRejectsLiveUnknownAndContinues(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	target.UpdateDataPlayer().Field57 = 0
	first, stale, unknown, nextEffect := playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0(), playerDamageLateEffectFixture4E17B0()
	base := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, first, stale}}
	item := &Object{ObjClass: object.ClassWeapon, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(base)}
	next := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, nextEffect}})}
	target.InvFirstItem = item
	var events []string
	r := playerDamageRuntime4E17B0(t, sound, new([]int32))
	r.CanApplyLateDefend = func(m *ModifierEff) bool { return m != unknown }
	r.ApplyLateDefend = func(m *ModifierEff, _, _, _, _ *Object, d int32, _ object.DamageType) int32 {
		switch m {
		case first:
			base.Modifiers[3], item.InvNextItem = unknown, next
			events = append(events, "first")
			return -7
		case nextEffect:
			if d != -7 {
				t.Fatal("rejected effect changed the live damage word")
			}
			events = append(events, "next")
			return 1
		default:
			t.Fatal("unported or stale callback was invoked")
			return d
		}
	}
	r.Unsupported = func(reason string, owner, attacker, weapon *Object, d int32, typ object.DamageType) {
		if reason != "unsupported live late equipped-item defend effect" || owner != target || attacker != source || weapon != source || d != -7 || typ != object.DamageBite {
			t.Fatalf("live rejection=%q damage=%d type=%d", reason, d, typ)
		}
		events = append(events, "reject")
	}
	r.DamageClear = func(owner *Object, d int32) {
		if owner != target || d != 1 {
			t.Fatalf("supported continuation damage=%d, want 1", d)
		}
		events = append(events, "HP")
	}
	if h, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, r); !h || !result {
		t.Fatalf("handled/result=%t/%t", h, result)
	}
	if want := []string{"first", "reject", "next", "HP"}; !slices.Equal(events, want) {
		t.Fatalf("events=%v, want %v", events, want)
	}
}

func TestPlayerDamageNative4E17B0LateDefendFlagsOnlyAdmission(t *testing.T) {
	target, source, sound := playerDamageFixture4E17B0(t)
	target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, playerDamageLateEffectFixture4E17B0()}})}
	before, beforeUD := *target, *target.UpdateDataPlayer()
	r := playerDamageRuntime4E17B0(t, sound, new([]int32))
	var reason string
	r.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
	if h, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, r); h || result || reason != "late equipped-item defend effect" {
		t.Fatalf("flags-only admission=%t/%t %q", h, result, reason)
	}
	if *target != before || *target.UpdateDataPlayer() != beforeUD {
		t.Fatal("unsupported entry-time effect mutated the hit prefix")
	}
}

func TestPlayerDamageNative4E17B0LateDefendNilInitFaultAtUse(t *testing.T) {
	for _, godMode := range []bool{false, true} {
		t.Run(map[bool]string{false: "default tail", true: "GodMode early return"}[godMode], func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			target.UpdateDataPlayer().Field57 = 0
			// A malformed equipped record has no modifier base. Admission
			// cannot inspect it; only the original default-tail helper faults.
			target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped}
			r := playerDamageRuntime4E17B0(t, sound, new([]int32))
			r.GodMode = func() bool { return godMode }
			buffCalls := 0
			r.BuffOff = func(*Object, EnchantID) { buffCalls++ }
			r.DamageClear = func(*Object, int32) { t.Fatal("malformed late Defend reached HP") }
			var caught any
			func() {
				defer func() { caught = recover() }()
				if h, result := PlayerDamageNative4E17B0(target, source, source, 3, object.DamageBite, r); !h || !result {
					t.Fatalf("GodMode early return=%t/%t", h, result)
				}
			}()
			wantBuff := 1
			if godMode {
				wantBuff = 0
			}
			if (caught != nil) == godMode || buffCalls != wantBuff || target.UpdateDataPlayer().Field76 != 2 || target.HealthData.Cur != 20 {
				t.Fatalf("fault=%v GodMode=%t BuffOff=%d marker=%d HP=%d", caught, godMode, buffCalls, target.UpdateDataPlayer().Field76, target.HealthData.Cur)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0LateDefendLiveGameBallGuard(t *testing.T) {
	for _, test := range []struct {
		name       string
		ballType   uint16
		drop       bool
		sourceLess bool
		reason     string
	}{
		{name: "missing type", reason: "unsupported live GameBall type"},
		{name: "missing drop", ballType: 77, reason: "unsupported live GameBall drop"},
		{name: "source-less poison", ballType: 77, drop: true, sourceLess: true, reason: "unsupported live source-less GameBall drop"},
		{name: "supported drop before HP", ballType: 77, drop: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			target, source, sound := playerDamageFixture4E17B0(t)
			target.UpdateDataPlayer().Field57 = 0
			target.Field129 = &Object{TypeInd: 77}
			typ := object.DamageBite
			if test.sourceLess {
				source, typ = nil, object.DamagePoison
			}
			m := playerDamageLateEffectFixture4E17B0()
			r := playerDamageRuntime4E17B0(t, sound, new([]int32))
			r.GameBallType = test.ballType
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32 { return 30 }
			// No entry-time Defend exists. The earlier callback installs one
			// on the live inventory, including Poison's source-less path.
			r.GodMode = func() bool {
				target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped,
					InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, m}})}
				return false
			}
			var reason string
			var events []string
			r.Unsupported = func(s string, _, _, _ *Object, d int32, _ object.DamageType) {
				reason = s
				if d != 30 {
					t.Fatal("live service guard did not use late Defend damage")
				}
			}
			if test.drop {
				r.GameBallOnDamage = func(attacker, owner *Object, d int32) {
					if test.reason != "" || attacker != source || owner != target || d != 30 {
						t.Fatal("unsupported or incorrect live GameBall drop")
					}
					events = append(events, "drop")
				}
			}
			r.DamageClear = func(owner *Object, d int32) {
				if test.reason != "" || owner != target || d != 30 {
					t.Fatal("unsupported or incorrect live GameBall HP tail")
				}
				events = append(events, "HP")
			}
			h, result := PlayerDamageNative4E17B0(target, source, source, 3, typ, r)
			wantHandled := test.reason == ""
			if h != wantHandled || result != wantHandled || reason != test.reason {
				t.Fatalf("live GameBall guard=%t/%t %q, want supported=%t %q", h, result, reason, wantHandled, test.reason)
			}
			var wantEvents []string
			if wantHandled {
				wantEvents = []string{"drop", "HP"}
			}
			if !slices.Equal(events, wantEvents) || target.HealthData.Cur != 20 || target.UpdateDataPlayer().Field76 != 2 {
				t.Fatalf("live tail lost order/prefix: events=%v marker=%d HP=%d", events, target.UpdateDataPlayer().Field76, target.HealthData.Cur)
			}
		})
	}
}
