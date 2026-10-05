package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterFleeTrySpell544760AggressionGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		aggression float32
		want       bool
	}{
		{"passive", 0, false},
		{"below passive", math.Nextafter32(0.08, 0), false},
		{"passive boundary", 0.08, true},
		{"above passive", math.Nextafter32(0.08, 1), false},
		{"below active", math.Nextafter32(0.33, 0), false},
		{"active boundary", 0.33, true},
		{"above active", math.Nextafter32(0.33, 1), true},
		{"hostile", 0.8, true},
		{"NaN", float32(math.NaN()), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterFightSpellTestUnit540B90(t)
			update.Aggression = tc.aggression
			calls := 0
			monsterFleeTrySpell544760(unit, update, monsterFleeSpellHooks544760{
				buffSelf:    func(*Object) bool { calls++; return true },
				castRelated: func(*Object) bool { t.Fatal("wrong selector"); return false },
				canInteract: func(*Object, *Object, int) bool { t.Fatal("interaction after self success"); return false },
			})
			if (calls == 1) != tc.want {
				t.Fatalf("self selector calls = %d, want enabled %t", calls, tc.want)
			}
		})
	}
	for _, tc := range []string{"cannot cast", "anti magic"} {
		t.Run(tc, func(t *testing.T) {
			unit, update := monsterFightSpellTestUnit540B90(t)
			update.Aggression = 0.8
			if tc == "cannot cast" {
				update.StatusFlags &^= object.MonStatusCanCastSpells
			} else {
				unit.Buffs = uint32(1) << ENCHANT_ANTI_MAGIC
			}
			monsterFleeTrySpell544760(unit, update, monsterFleeSpellHooks544760{
				buffSelf:    func(*Object) bool { t.Fatal("ineligible self selector"); return true },
				canInteract: func(*Object, *Object, int) bool { t.Fatal("ineligible interaction"); return true },
			})
		})
	}
}

func TestMonsterFleeTrySpell544760WholeStackAndCachedEnemyReloads(t *testing.T) {
	for _, dependency := range []bool{false, true} {
		for _, success := range []bool{false, true} {
			t.Run(fmt.Sprintf("dependency=%t/success=%t", dependency, success), func(t *testing.T) {
				unit, update := monsterFightSpellTestUnit540B90(t)
				update.Aggression = 0.8
				update.AIStackInd = 1
				update.AIStack[0].Action = uint32(ai.ACTION_WAIT)
				if dependency {
					update.AIStack[0].Action = uint32(ai.DEPENDENCY_UNINTERRUPTABLE)
				}
				update.AIStack[1].Action = uint32(ai.ACTION_FLEE)
				first, second, third := new(Object), new(Object), new(Object)
				update.CurrentEnemy = first
				var calls []string
				selectSpell := func(kind string) bool {
					calls = append(calls, kind)
					update.CurrentEnemy = second
					unit.UpdateData = unsafe.Pointer(new(MonsterUpdateData))
					return success
				}
				monsterFleeTrySpell544760(unit, update, monsterFleeSpellHooks544760{
					buffSelf:    func(*Object) bool { return selectSpell("self") },
					castRelated: func(*Object) bool { return selectSpell("related") },
					canInteract: func(got, target *Object, mode int) bool {
						if got != unit || target != second || mode != 0 {
							t.Fatal("interaction used stale/live replacement target or wrong mode")
						}
						calls = append(calls, "interact")
						update.CurrentEnemy = third
						return true
					},
					castOffensive: func(got, target *Object) bool {
						if got != unit || target != third {
							t.Fatal("offensive did not reload cached enemy")
						}
						calls = append(calls, "offensive")
						return true
					},
				})
				unit.UpdateData = unsafe.Pointer(update)
				kind := "self"
				if dependency {
					kind = "related"
				}
				want := []string{kind}
				if !success {
					want = append(want, "interact", "offensive")
				}
				if !reflect.DeepEqual(calls, want) {
					t.Fatalf("calls = %v, want %v", calls, want)
				}
			})
		}
	}
}

func TestMonsterActionFlee544760SpellStillMovesCachedAction(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{PosVec: types.Ptf(150, 220)}
	update.Field70 = 20
	update.AIStackHead().SetArgs(types.Ptf(0, 0), uint32(123))
	replacement := new(MonsterUpdateData)
	var calls []string
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame: func() uint32 { return 100 }, tickRate: func() uint32 { return 30 },
		pop: func() int { t.Fatal("unexpected pop"); return 0 },
		trySpell: func(got *Object, cached *MonsterUpdateData) {
			if got != unit || cached != update || cached.AIStackHead().ArgPos(0) != cached.CurrentEnemy.PosVec {
				t.Fatal("spell before target refresh")
			}
			calls = append(calls, "spell")
			unit.UpdateData = unsafe.Pointer(replacement)
		},
		generatePath: func(path []types.Pointf, got *Object, target *types.Pointf) int {
			if &path[0] != &update.Path[0] || got != unit || *target != update.CurrentEnemy.PosVec {
				t.Fatal("path not cached after spell")
			}
			calls = append(calls, "path")
			return 2
		},
		actuallyMove: func(*Object) bool { calls = append(calls, "move"); return true },
		moveAudio:    func(*Object) { calls = append(calls, "audio") },
	})
	unit.UpdateData = unsafe.Pointer(update)
	if !reflect.DeepEqual(calls, []string{"spell", "path", "move", "audio"}) || update.Field2 != 0 || update.Field70 != 100 || update.AIStackHead().ArgU32(2) != 123 || replacement.Field70 != 0 {
		t.Fatalf("flee calls/state = %v/%d/%d/%d/%d", calls, update.Field2, update.Field70, update.AIStackHead().ArgU32(2), replacement.Field70)
	}
}

func TestMonsterFleeCastRelated541050FullSpanAndExactCooldownOrder(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	monsterFightSetSpellFlag540B90(update, 1, monsterFleeRelatedSpellMask541050)
	monsterFightSetSpellFlag540B90(update, 136, monsterFleeRelatedSpellMask541050)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFightRelatedSpellMask540D90)
	update.Field371 = 100
	var calls []string
	frames := 0
	replacement := new(MonsterUpdateData)
	rng := 0
	if !monsterFleeCastRelated541050(unit, monsterFightSpellHooks540B90{
		frame: func() uint32 {
			frames++
			calls = append(calls, "frame")
			if frames == 1 {
				return 100
			}
			if rng != 2 {
				t.Fatal("post-cast frame read before cooldown RNG")
			}
			return math.MaxUint32 - 4
		},
		spellAllowed: func(id spell.ID) bool { calls = append(calls, fmt.Sprintf("allowed:%d", id)); return true },
		random: func(min, max int) int {
			rng++
			if rng == 1 {
				calls = append(calls, "select")
				if min != 0 || max != 1 {
					t.Fatalf("candidate bounds = %d/%d", min, max)
				}
				return 1
			}
			calls = append(calls, "cooldown")
			if min != 7 || max != 9 {
				t.Fatalf("post-cast bounds = %d/%d", min, max)
			}
			return 8
		},
		cast: func(got *Object, id spell.ID, target *Object) {
			if got != unit || id != 136 || target != unit {
				t.Fatalf("cast = %p/%d/%p", got, id, target)
			}
			calls = append(calls, "cast")
			update.Field370_0, update.Field370_2 = 7, 9
			unit.UpdateData = unsafe.Pointer(replacement)
		},
	}) {
		t.Fatal("related selector did not cast")
	}
	unit.UpdateData = unsafe.Pointer(update)
	want := []string{"frame", "allowed:1", "allowed:136", "select", "cast", "cooldown", "frame"}
	if !reflect.DeepEqual(calls, want) || update.Field371 != 3 || replacement.Field371 != 0 || update.Field369 != 0 || update.Field365 != 0 {
		t.Fatalf("calls/cooldown = %v/%d/%d", calls, update.Field371, replacement.Field371)
	}
}

func TestMonsterFleeCastRelated541050ExactGates(t *testing.T) {
	for _, gate := range []string{"disabled", "cannot cast", "before deadline", "cast object", "cast position", "cast duration", "no candidates", "disallowed", "active enchant"} {
		t.Run(gate, func(t *testing.T) {
			unit, update := monsterFightSpellTestUnit540B90(t)
			monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFleeRelatedSpellMask541050)
			switch gate {
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "cannot cast":
				update.StatusFlags &^= object.MonStatusCanCastSpells
			case "before deadline":
				update.Field371 = 101
			case "cast object":
				update.AIStackHead().Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			case "cast position":
				update.AIStackHead().Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
			case "cast duration":
				update.AIStackHead().Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
			case "no candidates":
				monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFightSelfSpellMask540B90)
			case "active enchant":
				unit.Buffs = uint32(1) << ENCHANT_HASTED
			}
			if monsterFleeCastRelated541050(unit, monsterFightSpellHooks540B90{
				frame:        func() uint32 { return 100 },
				spellAllowed: func(spell.ID) bool { return gate != "disallowed" },
				random:       func(int, int) int { t.Fatal("ineligible selector consumed RNG"); return 0 },
				cast:         func(*Object, spell.ID, *Object) { t.Fatal("ineligible cast") },
			}) {
				t.Fatal("ineligible selector succeeded")
			}
		})
	}
}
