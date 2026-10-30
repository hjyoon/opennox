package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestFightSchedulersDefaultSummonAdmission531B40(t *testing.T) {
	for _, accept := range []bool{false, true} {
		name := "reject"
		if accept {
			name = "accept"
		}
		t.Run(name, func(t *testing.T) {
			f := schedNewFixture531B40(t)
			rows := schedLoad531B40(t)
			f.reset(&rows[0])
			unit, target := &f.units[0].Value, &f.units[1].Value
			unit.ObjClass = object.ClassMonster
			unit.ObjFlags = object.FlagEnabled
			unit.Buffs = 0
			update := (*MonsterUpdateData)(unit.UpdateData)
			update.StatusFlags = object.MonStatusCanCastSpells
			update.MonsterDef.MissileName148[0] = 0
			update.AIStackInd = 0
			update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
			update.AIStack[0].Field5 = 0
			monsterFightSetSpellFlag540B90(update, spell.SPELL_SUMMON_BAT, monsterFightRelatedSpellMask540D90)
			f.server.Rand.Logic = prand.New(531)
			f.server.Spells.byID = map[spell.ID]*SpellDef{
				spell.SPELL_SUMMON_BAT: {ID: spell.SPELL_SUMMON_BAT, Def: things.Spell{Flags: things.SpellMobsCanCast | things.SpellTargeted}},
			}
			calls := 0
			result := f.server.MonsterFightScheduler531B40(0x531b40, unit, target, MonsterFightSchedulerRuntime531B40{
				CanSummon: func(got *Object, guideIndex int) bool {
					calls++
					if got != unit || guideIndex != int(spell.SPELL_SUMMON_BAT)-74 {
						t.Fatalf("summon admission args=%p/%d", got, guideIndex)
					}
					return accept
				},
			})
			if result != 1 || calls != 1 {
				t.Fatalf("ordinary scheduler return/calls=%#x/%d", result, calls)
			}
			casts := 0
			for _, item := range update.AIStack[:int(update.AIStackInd)+1] {
				if item.Type() == ai.ACTION_CAST_SPELL_ON_OBJECT {
					casts++
					if item.ArgU32(0) != uint32(spell.SPELL_SUMMON_BAT) || item.Args[2] != uintptr(unsafe.Pointer(target)) {
						t.Fatalf("production summon stack args=%#x", item.Args)
					}
				}
			}
			if accept && (casts != 1 || update.Field369 != f.server.Frame()) || !accept && (casts != 0 || update.AIStackInd != 0 || update.Field369 != 0) {
				t.Fatalf("summon admission=%t casts/head/deadline=%d/%d/%d", accept, casts, update.AIStackInd, update.Field369)
			}
		})
	}
}
