package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterInversionSpellMask5408D0 = uint32(0x08000000)

type monsterCastInversionHooks5408D0 struct {
	frame        func() uint32
	balance      func(string) float64
	eachMissile  func(types.Pointf, float32, func(*Object) bool)
	loadThreat   func() uint32
	storeThreat  func(uint32)
	spellAllowed func(spell.ID) bool
	random       func(int, int) int
	cast         func(*Object, spell.ID, *Object)
}

// monsterUnitIsMagicMissile540B60 preserves GAME.EXE 00540B60, including
// its otherwise-unused native object return. Non-missiles and non-magic
// missiles do not read update data. Every targeted match stores the scan
// DWORD; a match does not stop the surrounding missile enumeration.
func monsterUnitIsMagicMissile540B60(missile, target *Object, storeThreat func(uint32)) *Object {
	if !missile.ObjClass.Has(object.ClassMissile) || !missile.ObjSubClass.AsMissile().Has(object.MissileMagic) {
		return missile
	}
	if missile.UpdateDataMissile().Target == target {
		storeThreat(1)
	}
	return target
}

// monsterCastInversion5408D0 restores GAME.EXE 005408D0. The update-data
// pointer is cached before the enable gate and retained across casting. Spell
// flags are scanned in ascending ID order over exactly 1..136. The cooldown
// bounds are read after the cast, its RNG is called before the live frame is
// loaded, and the final addition wraps as a DWORD. ANTI_MAGIC belongs to the
// caller's gate at 0054742D, not to this selector.
func monsterCastInversion5408D0(unit *Object, h monsterCastInversionHooks5408D0) bool {
	update := unit.UpdateDataMonster()
	if !unit.ObjFlags.Has(object.FlagEnabled) || !update.StatusFlags.Has(object.MonStatusCanCastSpells) {
		return false
	}
	if h.frame() < update.Field363 {
		return false
	}
	if action := update.AIStackHead().Type(); action >= ai.ACTION_CAST_SPELL_ON_OBJECT && action <= ai.ACTION_CAST_DURATION_SPELL {
		return false
	}
	h.storeThreat(0)
	radius := float32(h.balance("InversionRange") * 0.5)
	h.eachMissile(unit.PosVec, radius, func(missile *Object) bool {
		monsterUnitIsMagicMissile540B60(missile, unit, h.storeThreat)
		return true
	})
	if h.loadThreat() == 0 {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterInversionSpellMask5408D0, h.spellAllowed)
	if len(candidates) == 0 {
		return false
	}
	id := candidates[h.random(0, len(candidates)-1)]
	h.cast(unit, id, unit)
	cooldown := h.random(int(update.Field362_0), int(update.Field362_2))
	update.Field363 = h.frame() + uint32(cooldown)
	return true
}

// MonsterCastInversion5408D0 binds the original shared scan DWORD, native
// map index, spell registry, logic RNG and ordinary MonsterCast action stack.
func (s *Server) MonsterCastInversion5408D0(unit *Object) bool {
	return monsterCastInversion5408D0(unit, monsterCastInversionHooks5408D0{
		frame:        s.Frame,
		balance:      s.Balance.Float,
		eachMissile:  s.Map.EachMissileInCircle,
		loadThreat:   func() uint32 { return memmap.Uint32(0x5D4594, 2489156) },
		storeThreat:  func(value uint32) { *memmap.PtrUint32(0x5D4594, 2489156) = value },
		spellAllowed: func(id spell.ID) bool { return s.Spells.HasFlags(id, things.SpellMobsCanCast) },
		random:       s.Rand.Logic.IntClamp,
		cast:         (*Object).MonsterCast,
	})
}
