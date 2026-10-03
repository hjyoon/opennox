package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func playerDamageNPCPierceFixture4E17B0(t *testing.T) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0, *[]int32) {
	t.Helper()
	_, source, arrow, tail := defaultDamagePierceFixture4E0B30(t)
	target := monsterActionTestObject50A910(t)
	target.ObjClass, target.ObjSubClass = object.ClassMonster, 0x11012
	target.HealthData = &HealthData{Cur: 20, Max: 20}
	update := target.UpdateDataMonster()
	update.Field518, update.Field547, update.Field546 = math.Float32bits(0.25), 99, 77
	var damages []int32
	tail.DamageClear = func(v *Object, d int32) {
		damages = append(damages, d)
		v.HealthData.Cur -= uint16(d)
	}
	tail.FireProtection = func(*Object) float64 { t.Fatal("NPC PIERCE used fire protection"); return 0 }
	tail.ElectricProtection = func(*Object) float64 { t.Fatal("NPC PIERCE used electric protection"); return 0 }
	tail.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("NPC PIERCE used player state"); return false }
	r := PlayerDamageRuntime4E17B0{
		DefaultDamage: func(v, a, w *Object, d int32, typ object.DamageType) bool {
			return DefaultDamageWorld4E0B30(v, a, w, d, typ, tail)
		},
		GodMode:             func() bool { t.Fatal("NPC PIERCE used player GodMode"); return false },
		ElectricArmorScale:  func(*Object) float32 { t.Fatal("NPC PIERCE used electric armor"); return 0 },
		BlockSourceExcluded: func(*Object) bool { return false },
		BlockDirection:      func(*Object, types.Pointf) bool { return false },
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("NPC PIERCE rejected: %s", reason)
		},
	}
	return target, source, arrow, r, &damages
}

func playerDamageNPCPierceSourceFixture4E17B0(t *testing.T, playerSource bool) (*Object, *Object, *Object, PlayerDamageRuntime4E17B0, *[]int32) {
	t.Helper()
	if !playerSource {
		return playerDamageNPCPierceFixture4E17B0(t)
	}
	target, _, _, r, damages := playerDamageNPCPierceFixture4E17B0(t)
	_, source, arrow, tail := defaultDamagePlayerPierceFixture4E0B30(t)
	arrow.ObjOwner = source
	// Use the real shared tail contract, with a correctly typed player
	// update and no monster-only hit-sound dependency.
	tail.DamageClear = func(v *Object, d int32) {
		*damages = append(*damages, d)
		v.HealthData.Cur -= uint16(d)
	}
	r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, d, typ, tail)
	}
	return target, source, arrow, r, damages
}

func TestPlayerDamageNPCPierce4E17B0FractionalCarry(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			for _, pure := range []bool{false, true} {
				t.Run(map[bool]string{false: "stock MISSILE WEAPON", true: "pure missile"}[pure], func(t *testing.T) {
					target, source, arrow, r, damages := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
					if pure {
						arrow.ObjClass = object.ClassMissile
					}
					beforeSource := *source
					var beforePlayer PlayerUpdateData
					if playerSource {
						beforePlayer = *source.UpdateDataPlayer()
					}
					update := target.UpdateDataMonster()
					update.Field523_2 = 0x34
					for hit, carry := range []float32{0.25, 0.5, -0.25} {
						if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result {
							t.Fatalf("hit %d = %t/%t", hit, h, result)
						}
						if math.Float32frombits(update.Field1) != carry || update.Field547 != 1 || update.Field546 != uint32(arrow.TypeInd) {
							t.Fatalf("hit %d carry=%g marker=%d/%d", hit, math.Float32frombits(update.Field1), update.Field547, update.Field546)
						}
					}
					if playerSource && (*source != beforeSource || *source.UpdateDataPlayer() != beforePlayer) {
						t.Fatal("player source or update was mutated/read as a monster")
					}
					wantPos := source.PrevPos
					if pure {
						wantPos = arrow.PrevPos
					}
					if !reflect.DeepEqual(*damages, []int32{2, 2, 3}) || target.HealthData.Cur != 13 ||
						update.Field523_2 != 0x34 || !update.StatusFlags.Has(object.MonStatusInjured) ||
						target.Obj130 != arrow || target.Field131 != 3 || target.Frame134 != 700 || target.Pos132 != wantPos || (!playerSource && source.UpdateDataMonster().Field130 != 700) {
						t.Fatalf("damage=%v HP=%d position=%v", *damages, target.HealthData.Cur, target.Pos132)
					}
				})
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0ArmorBeforeQuestAndMinimum(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			for _, tc := range []struct {
				name                string
				armor, carry, scale float32
				damage, hp, wear    int32
				wantCarry           float32
			}{
				{"ordinary", 0.25, 0.25, 1, 3, 2, 1, 0.5},
				{"Quest after wear", 0.25, 0.25, 0.5, 8, 3, 2, 0.25},
				{"minimum after armor", 1, 0, 1, 1, 1, 1, 0},
				{"Quest minimum", 1, 0, 0, 1, 1, 1, 0},
				{"zero", 0.25, 0, 1, 0, 0, 0, 0},
				{"negative", 0.25, 0, 1, -3, -2, -1, -0.25},
				{"negative tie", 0.25, -0.25, 1, -3, -2, -1, -0.5},
			} {
				t.Run(tc.name, func(t *testing.T) {
					target, source, arrow, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
					ud := target.UpdateDataMonster()
					ud.Field518, ud.Field1 = math.Float32bits(tc.armor), math.Float32bits(tc.carry)
					itemCarry := float32(0.25)
					item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
						HealthData: &HealthData{Cur: 100, Max: 100}, UpdateData: unsafe.Pointer(&itemCarry), InitData: unsafe.Pointer(&ModifierInitData{})}
					target.InvFirstItem = item
					var events []string
					r.ItemArmorValue = func(*Object) float32 { return tc.armor }
					r.CanDamageArmor = func(v *Object) bool { return v == item }
					r.DamageArmor = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if v != item || a != source || w != arrow || d != tc.wear || typ != object.DamageImpale || ud.Field547 != 1 || ud.Field546 != uint32(arrow.TypeInd) {
							t.Fatalf("armor damage/order=%d marker=%d/%d", d, ud.Field547, ud.Field546)
						}
						events = append(events, "armor")
						v.HealthData.Cur -= uint16(d)
						return true
					}
					r.ReportArmorHealth = func(v, i *Object, before, after uint16) {
						if v != target || i != item || before != 100 || after != uint16(100-tc.wear) {
							t.Fatal("armor report")
						}
						events = append(events, "report")
					}
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = func() float32 { events = append(events, "quest"); return tc.scale }
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if v != target || a != source || w != arrow || d != tc.hp || typ != object.DamageImpale {
							t.Fatalf("HP damage=%d, want %d", d, tc.hp)
						}
						events = append(events, "HP")
						return false
					}
					if h, result := PlayerDamageNative4E17B0(target, source, arrow, tc.damage, object.DamageImpale, r); !h || result {
						t.Fatalf("result propagation=%t/%t", h, result)
					}
					wantEvents := []string{"quest", "HP"}
					if tc.wear > 0 {
						wantEvents = []string{"armor", "report", "quest", "HP"}
					}
					if !reflect.DeepEqual(events, wantEvents) || math.Float32frombits(ud.Field1) != tc.wantCarry || itemCarry != 0.25 {
						t.Fatalf("events=%v HP carry=%g armor carry=%g", events, math.Float32frombits(ud.Field1), itemCarry)
					}
				})
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0CachedArmorAndLiveCarry(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			target, source, arrow, r, damages := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
			old := target.UpdateDataMonster()
			old.ArmorEquipFlags, old.AIStackInd = 0x1000000, 0
			old.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
			live := &MonsterUpdateData{Field518: math.Float32bits(0.5), Field1: math.Float32bits(0.25)}
			r.BlockDirection = func(v *Object, p types.Pointf) bool {
				if v != target || p != arrow.PrevPos {
					t.Fatal("ordinary shield must face previous position")
				}
				old.Field518 = math.Float32bits(0.75)
				target.UpdateData = unsafe.Pointer(live)
				return false
			}
			itemCarry := float32(0)
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 50},
				UpdateData: unsafe.Pointer(&itemCarry), InitData: unsafe.Pointer(&ModifierInitData{})}
			target.InvFirstItem = item
			r.ItemArmorValue = func(*Object) float32 { return 0.25 }
			r.CanDamageArmor = func(v *Object) bool { return v == item }
			r.DamageArmor = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				if v != item || a != source || w != arrow || d != 1 || typ != object.DamageImpale || old.Field547 != 1 || old.Field546 != uint32(arrow.TypeInd) || live.Field1 != math.Float32bits(0.25) {
					t.Fatalf("cached absorption/live durability: wear=%d marker=%d/%d", d, old.Field547, old.Field546)
				}
				v.HealthData.Cur--
				return true
			}
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); !h || !result ||
				!reflect.DeepEqual(*damages, []int32{6}) || target.HealthData.Cur != 14 || item.HealthData.Cur != 49 || itemCarry != 0 || old.Field1 != 0 || live.Field1 != math.Float32bits(0.25) || live.Field547 != 1 {
				t.Fatalf("cached/live result=%t/%t HP=%d armor=%d damages=%v", h, result, target.HealthData.Cur, item.HealthData.Cur, *damages)
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0AbsorptionBinary32Boundary(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			for _, tc := range []struct {
				armor  float32
				damage int32
				hp     uint16
				carry  uint32
			}{
				{0.1, 3, 17, 0xbe999998},
				{0.2, 9, 13, 0x3e4cccc0},
				{0.4, 3, 18, 0xbe4cccd0},
			} {
				target, source, arrow, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
				ud := target.UpdateDataMonster()
				ud.Field518 = math.Float32bits(tc.armor)
				// 004E1F84 subtracts/multiplies in the wider x87 format, then
				// spills to binary32 before 004E20F0. Rounding 1-armor to binary32
				// prematurely gives a different fractional carry for all three.
				if h, result := PlayerDamageNative4E17B0(target, source, arrow, tc.damage, object.DamageImpale, r); !h || !result ||
					target.HealthData.Cur != tc.hp || ud.Field1 != tc.carry {
					t.Fatalf("armor=%g damage=%d result=%t/%t HP=%d carry=%08x, want HP=%d carry=%08x", tc.armor, tc.damage, h, result, target.HealthData.Cur, ud.Field1, tc.hp, tc.carry)
				}
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0WearClearsCachedMarker(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			target, source, arrow, r, damages := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
			cached := target.UpdateDataMonster()
			live := &MonsterUpdateData{Field1: math.Float32bits(-0.125), Field518: math.Float32bits(0.875)}
			itemCarry := float32(0)
			item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 50},
				UpdateData: unsafe.Pointer(&itemCarry), InitData: unsafe.Pointer(&ModifierInitData{})}
			target.InvFirstItem = item
			r.ItemArmorValue = func(*Object) float32 { return 0.25 }
			r.CanDamageArmor = func(v *Object) bool { return v == item }
			r.DamageArmor = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				if v != item || a != source || w != arrow || d != 1 || typ != object.DamageImpale ||
					cached.Field547 != 1 || cached.Field546 != uint32(arrow.TypeInd) || cached.Field1 != math.Float32bits(0.25) {
					t.Fatal("durability occurred before the missile marker/carry")
				}
				cached.Field547, cached.Field546 = 0, 99
				target.UpdateData = unsafe.Pointer(live)
				v.HealthData.Cur--
				return true
			}
			tail := r.DefaultDamage
			r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
				if cached.Field547 != 2 || cached.Field546 != uint32(object.DamageImpale) || live.Field547 != 0 || d != 2 {
					t.Fatal("cached marker fallback must precede the live DefaultDamage tail")
				}
				return tail(v, a, w, d, typ)
			}
			if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result ||
				!reflect.DeepEqual(*damages, []int32{2}) || target.HealthData.Cur != 18 || item.HealthData.Cur != 49 ||
				cached.Field547 != 2 || cached.Field546 != 3 || cached.Field1 != math.Float32bits(0.25) || itemCarry != 0 ||
				live.Field547 != 1 || live.Field546 != uint32(arrow.TypeInd) || live.Field1 != math.Float32bits(-0.125) {
				t.Fatalf("wear/live marker result=%t/%t cached=%d/%d live=%d/%d damage=%v", h, result, cached.Field547, cached.Field546, live.Field547, live.Field546, *damages)
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0AdmissionBeforeMutation(t *testing.T) {
	for _, playerSource := range []bool{false, true} {
		t.Run(fmt.Sprintf("player-source-%t", playerSource), func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				change func(*Object, *Object, *Object, *PlayerDamageRuntime4E17B0)
			}{
				{"ordinary monster", func(v, _, _ *Object, _ *PlayerDamageRuntime4E17B0) { v.ObjSubClass = 0 }},
				{"missing target update", func(v, _, _ *Object, _ *PlayerDamageRuntime4E17B0) { v.UpdateData = nil }},
				{"missing source update", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.UpdateData = nil }},
				{"non-unit source", func(_, a, _ *Object, _ *PlayerDamageRuntime4E17B0) { a.ObjClass = object.ClassSimple }},
				{"qualifying melee weapon", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjSubClass = 0 }},
				{"missile wand", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjClass |= object.ClassWand }},
				{"missile unit", func(_, _, w *Object, _ *PlayerDamageRuntime4E17B0) { w.ObjClass |= object.ClassMonster }},
				{"default service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) { r.DefaultDamage = nil }},
				{"Quest service", func(_, _, _ *Object, r *PlayerDamageRuntime4E17B0) {
					r.QuestMode = func() bool { return true }
					r.QuestDamageScale = nil
				}},
				{"armor damage service", func(v, _, _ *Object, r *PlayerDamageRuntime4E17B0) {
					carry := float32(0.25)
					v.InvFirstItem = &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
						HealthData: &HealthData{Cur: 50}, UpdateData: unsafe.Pointer(&carry), InitData: unsafe.Pointer(&ModifierInitData{})}
					r.ItemArmorValue = func(*Object) float32 { return 0.25 }
					r.CanDamageArmor = func(*Object) bool { return false }
				}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					target, source, arrow, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, playerSource)
					originalUpdate := target.UpdateDataMonster()
					tc.change(target, source, arrow, &r)
					beforeTarget, beforeSource, beforeArrow, beforeUpdate := *target, *source, *arrow, *originalUpdate
					rejected := 0
					r.Unsupported = func(string, *Object, *Object, *Object, int32, object.DamageType) { rejected++ }
					if h, result := PlayerDamageNative4E17B0(target, source, arrow, 8, object.DamageImpale, r); h || result || rejected != 1 ||
						*target != beforeTarget || *source != beforeSource || *arrow != beforeArrow || *originalUpdate != beforeUpdate || target.HealthData.Cur != 20 ||
						(target.InvFirstItem != nil && (*(*float32)(target.InvFirstItem.UpdateData) != 0.25 || target.InvFirstItem.HealthData.Cur != 50)) {
						t.Fatalf("unsupported shape mutated state: handled=%t result=%t reports=%d", h, result, rejected)
					}
				})
			}
		})
	}
}

func TestPlayerDamageNPCPierce4E17B0PlayerArmorDefendArguments(t *testing.T) {
	target, source, arrow, r, damages := playerDamageNPCPierceSourceFixture4E17B0(t, true)
	ud := target.UpdateDataMonster()
	carry := float32(0.25)
	modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		HealthData: &HealthData{Cur: 50}, UpdateData: unsafe.Pointer(&carry),
		InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, modifier, nil, nil}})}
	target.InvFirstItem = item
	beforeSource, beforePlayer := *source, *source.UpdateDataPlayer()
	var events []string
	r.ItemArmorValue = func(got *Object) float32 {
		if got != item {
			t.Fatal("armor definition identity")
		}
		return 0.25
	}
	r.ApplyArmorDefend = func(m *ModifierEff, armor, owner, weapon, attacker *Object, amount *float32) bool {
		if m != modifier || armor != item || owner != target || weapon != arrow || attacker != source || *amount != 1 {
			t.Fatal("armor modifier lost the actual player/missile arguments")
		}
		events = append(events, "defend")
		*amount = 0.5
		return true
	}
	r.CanDamageArmor = func(got *Object) bool { return got == item }
	r.DamageArmor = func(got, attacker, weapon *Object, amount int32, typ object.DamageType) bool {
		if got != item || attacker != source || weapon != arrow || amount != 1 || typ != object.DamageImpale ||
			ud.Field547 != 1 || ud.Field546 != uint32(arrow.TypeInd) || ud.Field1 != math.Float32bits(0.25) || carry != -0.25 {
			t.Fatal("durability must receive the player source after missile marker/carry")
		}
		events = append(events, "armor")
		item.HealthData.Cur--
		return true
	}
	r.ReportArmorHealth = func(owner, armor *Object, before, after uint16) {
		if owner != target || armor != item || before != 50 || after != 49 {
			t.Fatal("armor health report identity/order")
		}
		events = append(events, "report")
	}
	tail := r.DefaultDamage
	r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		events = append(events, "default")
		return tail(v, a, w, d, typ)
	}
	if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result ||
		!reflect.DeepEqual(events, []string{"defend", "armor", "report", "default"}) || !reflect.DeepEqual(*damages, []int32{2}) ||
		target.HealthData.Cur != 18 || item.HealthData.Cur != 49 || *source != beforeSource || *source.UpdateDataPlayer() != beforePlayer {
		t.Fatalf("result=%t/%t events=%v HP=%d armor=%d", h, result, events, target.HealthData.Cur, item.HealthData.Cur)
	}
}

func TestPlayerDamageNPCPierce4E17B0PlayerFriendlyCampaignStillWearsArmor(t *testing.T) {
	target, source, arrow, r, _ := playerDamageNPCPierceSourceFixture4E17B0(t, true)
	_, _, _, tail := defaultDamagePlayerPierceFixture4E0B30(t)
	tail.GameplayFlag1 = func() bool { return false }
	tail.IsEnemy = func(v, owner *Object) bool {
		if v != target || owner != source {
			t.Fatal("friendly owner predicate")
		}
		return false
	}
	tail.DamageClear = func(*Object, int32) { t.Fatal("friendly campaign HP changed") }
	r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		return DefaultDamageWorld4E0B30(v, a, w, d, typ, tail)
	}
	carry := float32(0)
	item := &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped,
		HealthData: &HealthData{Cur: 50}, UpdateData: unsafe.Pointer(&carry),
		InitData: unsafe.Pointer(&ModifierInitData{})}
	target.InvFirstItem = item
	r.ItemArmorValue = func(*Object) float32 { return 0.25 }
	r.CanDamageArmor = func(*Object) bool { return true }
	wear := 0
	r.DamageArmor = func(v, a, w *Object, d int32, typ object.DamageType) bool {
		if v != item || a != source || w != arrow || d != 1 || typ != object.DamageImpale {
			t.Fatal("friendly hit armor arguments")
		}
		wear++
		item.HealthData.Cur--
		return true
	}
	// 004E2180 armor wear precedes DefaultDamage's campaign-friendly HP
	// gate, whose original prefix clears only the NPC hit latch. Do not
	// incorrectly suppress all PlayerDamage side effects.
	if h, result := PlayerDamageNative4E17B0(target, source, arrow, 3, object.DamageImpale, r); !h || !result ||
		target.HealthData.Cur != 20 || item.HealthData.Cur != 49 || wear != 1 ||
		target.UpdateDataMonster().Field1 != math.Float32bits(0.25) || target.UpdateDataMonster().Field547 != 0 ||
		target.UpdateDataMonster().Field546 != uint32(arrow.TypeInd) ||
		target.Obj130 != nil || target.UpdateDataMonster().StatusFlags.Has(object.MonStatusInjured) {
		t.Fatalf("friendly result=%t/%t HP=%d armor=%d wear=%d", h, result, target.HealthData.Cur, item.HealthData.Cur, wear)
	}
}
