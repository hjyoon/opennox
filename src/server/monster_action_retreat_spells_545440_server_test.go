package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterRetreatSpellFixture545440(t *testing.T, duration bool) (*Server, *Object, *MonsterUpdateData) {
	t.Helper()
	s, unit, update := monsterScriptHitFixture515A30(t)
	s.SetFrame(100)
	s.SetTickRate(30)
	s.Rand.Logic = prand.New(545440)
	flags := things.SpellMobsCanCast
	if duration {
		flags |= things.SpellDuration
	}
	s.Spells.byID = map[spell.ID]*SpellDef{
		spell.SPELL_HASTE: {ID: spell.SPELL_HASTE, Def: things.Spell{Flags: flags}},
	}
	unit.ObjFlags = object.FlagEnabled
	unit.HealthData = &HealthData{Cur: 10, Max: 100}
	update.ResumeLevel = 0.5
	update.StatusFlags = object.MonStatusCanCastSpells
	update.CurrentEnemy = &Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(-100, 200)}
	update.Field370_0, update.Field370_2 = 9, 9
	update.Field371 = 100 // Equality is ready in the original unsigned frame gate.
	update.Field365, update.Field367, update.Field369 = 17, 19, 23
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_RETREAT), Args: [4]uintptr{1, 2, 3, 4}}
	monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFleeRelatedSpellMask541050)
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(update.CurrentEnemy)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("retreat fixture pointer %p is not above 4 GiB", pointer)
		}
	}
	return s, unit, update
}

func TestMonsterRetreatServer545440SchedulesRealSelfCast(t *testing.T) {
	for _, npc := range []bool{false, true} {
		for _, duration := range []bool{false, true} {
			name := "monster/instant"
			if npc {
				name = "NPC/instant"
			}
			if duration {
				name = name[:len(name)-len("instant")] + "duration"
			}
			t.Run(name, func(t *testing.T) {
				s, unit, update := monsterRetreatSpellFixture545440(t, duration)
				if npc {
					unit.ObjSubClass = object.SubClass(object.MonsterNPC)
				}
				retreat, enemy := update.AIStack[0], update.CurrentEnemy
				s.MonsterActionRetreat545440(unit)
				wantAction, wantIndex := ai.ACTION_CAST_SPELL_ON_OBJECT, int8(2)
				if duration {
					wantAction, wantIndex = ai.ACTION_CAST_DURATION_SPELL, 3
					deadline := update.AIStack[2].ArgU32(0)
					if update.AIStack[2].Type() != ai.DEPENDENCY_TIME || deadline < 115 || deadline > 160 {
						t.Fatalf("real duration dependency = %+v", update.AIStack[2])
					}
				}
				head := update.AIStackHead()
				if update.AIStackInd != wantIndex || head.Type() != wantAction ||
					head.ArgU32(0) != uint32(spell.SPELL_HASTE) || head.ArgU32(1) != 0 ||
					head.Args[2] != uintptr(unsafe.Pointer(unit)) || head.ArgObj(2) != unit ||
					update.AIStack[1].Type() != ai.DEPENDENCY_UNINTERRUPTABLE || !s.AI.StackChanged {
					t.Fatalf("real retreat self cast = %+v, index %d, changed %t", head, update.AIStackInd, s.AI.StackChanged)
				}
				if update.AIStack[0] != retreat || update.CurrentEnemy != enemy ||
					update.Field371 != 109 || update.Field365 != 17 || update.Field367 != 19 || update.Field369 != 23 ||
					unit.HealthData.Cur != 10 || unit.Buffs != 0 {
					t.Fatalf("retreat/cooldown/health/enchant changed outside scheduling: %+v", update.GetAIStack())
				}
				runtime.KeepAlive(unit)
				runtime.KeepAlive(enemy)
			})
		}
	}
}

func TestMonsterRetreatServer545440OriginalSpellGatesStillFlee(t *testing.T) {
	for _, gate := range []string{"disabled", "cannot cast", "before deadline", "wrong mask", "disallowed", "active enchant", "any active candidate", "anti magic", "non-caster anti magic", "cast object", "cast location", "cast duration"} {
		t.Run(gate, func(t *testing.T) {
			s, unit, update := monsterRetreatSpellFixture545440(t, false)
			switch gate {
			case "disabled":
				unit.ObjFlags &^= object.FlagEnabled
			case "cannot cast":
				update.StatusFlags &^= object.MonStatusCanCastSpells
			case "before deadline":
				update.Field371 = 101
			case "wrong mask":
				monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFightSelfSpellMask540B90)
			case "disallowed":
				s.Spells.DefByInd(spell.SPELL_HASTE).Def.Flags &^= things.SpellMobsCanCast
			case "active enchant":
				unit.Buffs = 1 << ENCHANT_HASTED
			case "any active candidate":
				unit.Buffs = 1 << ENCHANT_HASTED
				s.Spells.byID[spell.SPELL_FIREBALL] = &SpellDef{ID: spell.SPELL_FIREBALL, Def: things.Spell{Flags: things.SpellMobsCanCast}}
				monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFleeRelatedSpellMask541050)
			case "anti magic", "non-caster anti magic":
				unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
				if gate == "non-caster anti magic" {
					update.StatusFlags &^= object.MonStatusCanCastSpells
				}
			case "cast object":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
			case "cast location":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
			case "cast duration":
				update.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
			}
			before, originalHead := update.Field371, update.AIStack[0]
			s.MonsterActionRetreat545440(unit)
			if update.AIStackInd != 2 || update.AIStack[0] != originalHead ||
				update.AIStack[1].Type() != ai.DEPENDENCY_TIME ||
				update.AIStack[1].ArgU32(0) < 220 || update.AIStack[1].ArgU32(0) > 280 ||
				update.AIStack[2].Type() != ai.ACTION_FLEE || update.AIStack[2].ArgPos(0) != update.CurrentEnemy.PosVec ||
				update.AIStack[2].ArgU32(2) != 0 || update.Field371 != before || !s.AI.StackChanged {
				t.Fatalf("ineligible related spell changed original retreat fallback: %+v cooldown %d/%d", update.GetAIStack(), before, update.Field371)
			}
			runtime.KeepAlive(unit)
		})
	}
}

func TestMonsterRetreatServer545440RejectedSelfCastKeepsSelectorCooldown(t *testing.T) {
	for _, remaining := range []int{0, 1} {
		name := "full"
		if remaining == 1 {
			name = "one slot"
		}
		t.Run(name, func(t *testing.T) {
			s, unit, update := monsterRetreatSpellFixture545440(t, false)
			update.AIStackInd = int8(len(update.AIStack) - 1 - remaining)
			update.AIStackHead().Action = uint32(ai.ACTION_RETREAT)
			s.MonsterActionRetreat545440(unit)
			wantHead := ai.ACTION_RETREAT
			if remaining == 1 {
				wantHead = ai.DEPENDENCY_UNINTERRUPTABLE
			}
			if update.AIStackInd != int8(len(update.AIStack)-1) || update.AIStackHead().Type() != wantHead ||
				update.Field371 != 109 || s.AI.StackChanged != (remaining == 1) {
				t.Fatalf("rejected self cast must still consume selector cooldown without FLEE: head=%+v cooldown=%d changed=%t", update.AIStackHead(), update.Field371, s.AI.StackChanged)
			}
			runtime.KeepAlive(unit)
		})
	}
}

func TestMonsterRetreatServer545440RecoveredDoesNotCast(t *testing.T) {
	s, unit, update := monsterRetreatSpellFixture545440(t, false)
	unit.HealthData.Cur = 50
	s.MonsterActionRetreat545440(unit)
	if update.AIStackInd != 0 || update.AIStackHead().Type() != ai.ACTION_IDLE || update.Field371 != 100 || !s.AI.StackChanged {
		t.Fatalf("recovered retreat must pop without casting: %+v cooldown %d", update.GetAIStack(), update.Field371)
	}
	runtime.KeepAlive(unit)
}
