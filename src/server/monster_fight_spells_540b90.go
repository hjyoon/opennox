package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterFightSelfSpellMask540B90      = uint32(0x10000000)
	monsterFightOffensiveSpellMask540F20 = uint32(0x20000000)
	monsterFightRelatedSpellMask540D90   = uint32(0x40000000)
	monsterFightFirstSummonSpell540D20   = spell.ID(75)
	monsterFightLastSummonSpell540D20    = spell.ID(114)
)

const monsterFightSpellFlagSpan540B90 = uintptr(SpellsMax-2) * unsafe.Sizeof(uint32(0))

var _ = [1]struct{}{}[unsafe.Offsetof(MonsterUpdateData{}.Field508)-unsafe.Offsetof(MonsterUpdateData{}.Field373)-monsterFightSpellFlagSpan540B90]
var _ = [1]struct{}{}[unsafe.Offsetof(MonsterUpdateData{}.Field373)+monsterFightSpellFlagSpan540B90-unsafe.Offsetof(MonsterUpdateData{}.Field508)]

type monsterFightSpellHooks540B90 struct {
	frame        func() uint32
	spellAllowed func(spell.ID) bool
	random       func(int, int) int
	cast         func(*Object, spell.ID, *Object)
	activeSummon func(*Object) bool
	canSummon    func(*Object, int) bool
}

func monsterFightSpellFlag540B90(update *MonsterUpdateData, id spell.ID) uint32 {
	if update == nil || id <= 0 || id >= spell.ID(SpellsMax) {
		return 0
	}
	return unsafe.Slice(&update.Field373, SpellsMax-1)[int(id)-1]
}

func monsterFightSpellIsSummon540D20(id spell.ID) bool {
	return id >= monsterFightFirstSummonSpell540D20 && id <= monsterFightLastSummonSpell540D20
}

func monsterFightHasSpellEnchant540CE0(unit *Object, id spell.ID) bool {
	for enchant := EnchantID(0); enchant < 32; enchant++ {
		if enchant.Spell() == id {
			return unit.HasEnchant(enchant)
		}
	}
	return false
}

func monsterFightSpellReady540B90(unit *Object, update *MonsterUpdateData, deadline uint32, frame func() uint32) bool {
	if unit == nil || update == nil || frame == nil ||
		!unit.Flags().Has(object.FlagEnabled) ||
		!update.StatusFlags.Has(object.MonStatusCanCastSpells) ||
		frame() < deadline {
		return false
	}
	head := update.AIStackHead()
	if head == nil {
		return false
	}
	action := head.Type()
	return action < ai.ACTION_CAST_SPELL_ON_OBJECT || action > ai.ACTION_CAST_DURATION_SPELL
}

func monsterFightSpellCandidates540B90(update *MonsterUpdateData, mask uint32, spellAllowed func(spell.ID) bool) []spell.ID {
	var candidates [SpellsMax - 1]spell.ID
	count := 0
	for id := spell.ID(1); id < spell.ID(SpellsMax); id++ {
		if monsterFightSpellFlag540B90(update, id)&mask != 0 && spellAllowed(id) {
			candidates[count] = id
			count++
		}
	}
	return candidates[:count]
}

// monsterFightBuffSelf540B90 preserves GAME.EXE 00540B90. UpdateData is
// cached across casting, while the post-cast frame and cooldown RNG remain
// live calls. Any candidate whose first matching enchant is already active
// suppresses the whole selector before RNG is consumed.
func monsterFightBuffSelf540B90(unit *Object, hooks monsterFightSpellHooks540B90) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.spellAllowed == nil || hooks.random == nil || hooks.cast == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if !monsterFightSpellReady540B90(unit, update, update.Field365, hooks.frame) {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterFightSelfSpellMask540B90, hooks.spellAllowed)
	if len(candidates) == 0 {
		return false
	}
	for _, id := range candidates {
		if monsterFightHasSpellEnchant540CE0(unit, id) {
			return false
		}
	}
	id := candidates[hooks.random(0, len(candidates)-1)]
	hooks.cast(unit, id, unit)
	update.Field365 = hooks.frame() + uint32(hooks.random(int(update.Field364_0), int(update.Field364_2)))
	return true
}

// monsterFightCastOffensive540F20 preserves GAME.EXE 00540F20.
func monsterFightCastOffensive540F20(unit, target *Object, hooks monsterFightSpellHooks540B90) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.spellAllowed == nil || hooks.random == nil || hooks.cast == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if !monsterFightSpellReady540B90(unit, update, update.Field367, hooks.frame) {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterFightOffensiveSpellMask540F20, hooks.spellAllowed)
	if len(candidates) == 0 {
		return false
	}
	id := candidates[hooks.random(0, len(candidates)-1)]
	hooks.cast(unit, id, target)
	update.Field367 = hooks.frame() + uint32(hooks.random(int(update.Field366_0), int(update.Field366_2)))
	return true
}

// monsterFightCastRelated540D90 preserves GAME.EXE 00540D90. A rejected
// summon is retried when any non-summon candidate exists. An all-summon set
// stops immediately when an active summon or the creature limit rejects the
// selected spell.
func monsterFightCastRelated540D90(unit, target *Object, hooks monsterFightSpellHooks540B90) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) ||
		hooks.spellAllowed == nil || hooks.random == nil || hooks.cast == nil ||
		hooks.activeSummon == nil || hooks.canSummon == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if !monsterFightSpellReady540B90(unit, update, update.Field369, hooks.frame) {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterFightRelatedSpellMask540D90, hooks.spellAllowed)
	if len(candidates) == 0 {
		return false
	}
	allSummons := true
	for _, id := range candidates {
		if !monsterFightSpellIsSummon540D20(id) {
			allSummons = false
			break
		}
	}
	for {
		id := candidates[hooks.random(0, len(candidates)-1)]
		if monsterFightSpellIsSummon540D20(id) {
			if hooks.activeSummon(unit) {
				if allSummons {
					return false
				}
				continue
			}
			if !hooks.canSummon(unit, int(id)-74) {
				if allSummons {
					return false
				}
				continue
			}
		}
		hooks.cast(unit, id, target)
		update.Field369 = hooks.frame() + uint32(hooks.random(int(update.Field368_0), int(update.Field368_2)))
		return true
	}
}
