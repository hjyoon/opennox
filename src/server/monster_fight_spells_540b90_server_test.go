package server

import (
	"testing"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func TestMonsterFightSpellHooks540B90BindNativeServerState(t *testing.T) {
	const seed = 540
	s := new(Server)
	s.frame = 1234
	s.Rand.Logic = prand.New(seed)
	s.Spells.byID = map[spell.ID]*SpellDef{
		spell.SPELL_FIREBALL: {
			ID:  spell.SPELL_FIREBALL,
			Def: things.Spell{Flags: things.SpellMobsCanCast},
		},
		spell.SPELL_HASTE: {
			ID:  spell.SPELL_HASTE,
			Def: things.Spell{},
		},
	}
	caster := new(Object)
	other := new(Object)
	second := &DurSpell{Spell: uint32(spell.SPELL_SUMMON_BAT), Caster16: other}
	first := &DurSpell{Spell: uint32(spell.SPELL_FIREBALL), Caster16: caster, Next: second}
	s.Spells.Dur.List = first
	limitCalls := 0
	hooks := s.monsterFightSpellHooks540B90(func(got *Object, size int) bool {
		if got != caster || size != 17 {
			t.Fatalf("summon limit args = %p/%d", got, size)
		}
		limitCalls++
		return true
	})

	if got := hooks.frame(); got != 1234 {
		t.Fatalf("frame = %d, want 1234", got)
	}
	if !hooks.spellAllowed(spell.SPELL_FIREBALL) || hooks.spellAllowed(spell.SPELL_HASTE) {
		t.Fatalf("registry flags = fireball:%t haste:%t", hooks.spellAllowed(spell.SPELL_FIREBALL), hooks.spellAllowed(spell.SPELL_HASTE))
	}
	wantRNG := prand.New(seed)
	if got, want := hooks.random(3, 19), wantRNG.IntClamp(3, 19); got != want {
		t.Fatalf("logic RNG = %d, want %d", got, want)
	}
	if hooks.activeSummon(caster) {
		t.Fatal("another caster's summon matched")
	}
	second.Caster16 = caster
	if !hooks.activeSummon(caster) {
		t.Fatal("native duration summon was not found")
	}
	if !hooks.canSummon(caster, 17) || limitCalls != 1 {
		t.Fatalf("summon limit result/calls = false/%d", limitCalls)
	}
}
