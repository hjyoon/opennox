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

func TestFightSchedulersDefaultRelatedSpell531B40(t *testing.T) {
	f := schedNewFixture531B40(t)
	rows := schedLoad531B40(t)
	f.reset(&rows[0])
	unit, target := &f.units[0].Value, &f.units[1].Value
	unit.ObjClass = object.ClassMonster
	unit.ObjFlags = object.FlagEnabled
	unit.Buffs = 0
	update := (*MonsterUpdateData)(unit.UpdateData)
	update.StatusFlags = object.MonStatusCanCastSpells
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	update.AIStack[0].Field5 = 0
	monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFightRelatedSpellMask540D90)
	f.server.Rand.Logic = prand.New(531)
	f.server.Spells.byID = map[spell.ID]*SpellDef{
		spell.SPELL_FIREBALL: {ID: spell.SPELL_FIREBALL, Def: things.Spell{Flags: things.SpellMobsCanCast | things.SpellTargeted}},
	}
	if got := f.server.MonsterFightScheduler531B40(0x531b40, unit, target, MonsterFightSchedulerRuntime531B40{}); got != 1 {
		t.Errorf("ordinary related spell return=%#x, want integer1", got)
	}
	var casts int
	for _, item := range update.AIStack[:int(update.AIStackInd)+1] {
		if item.Type() == ai.ACTION_CAST_SPELL_ON_OBJECT {
			casts++
			if item.ArgU32(0) != uint32(spell.SPELL_FIREBALL) || item.Args[2] != uintptr(unsafe.Pointer(target)) {
				t.Errorf("ordinary production spell stack args=%#x", item.Args)
			}
		}
	}
	if casts != 1 || update.Field369 != f.server.Frame() {
		t.Errorf("ordinary production related cast count=%d deadline=%d, want1/%d", casts, update.Field369, f.server.Frame())
	}
}
