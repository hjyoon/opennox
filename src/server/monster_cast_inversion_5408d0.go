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

type MonsterCastInversionRuntime5408D0 struct {
	Frame        func() uint32
	Balance      func(string) float64
	EachMissile  func(types.Pointf, float32, func(*Object) bool)
	LoadThreat   func() uint32
	StoreThreat  func(uint32)
	SpellAllowed func(spell.ID) bool
	Random       func(int, int) int
	Cast         func(*Object, spell.ID, *Object)
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

// MonsterCastInversionNative5408D0 restores GAME.EXE 005408D0. The update-data
// pointer is cached before the enable gate and retained across casting. Spell
// flags are scanned in ascending ID order over exactly 1..136. The cooldown
// bounds are read after the cast, its RNG is called before the live frame is
// loaded, and the final addition wraps as a DWORD. ANTI_MAGIC belongs to the
// caller's gate at 0054742D, not to this selector.
func MonsterCastInversionNative5408D0(unit *Object, h MonsterCastInversionRuntime5408D0) bool {
	update := unit.UpdateDataMonster()
	if !unit.ObjFlags.Has(object.FlagEnabled) || !update.StatusFlags.Has(object.MonStatusCanCastSpells) {
		return false
	}
	if h.Frame() < update.Field363 {
		return false
	}
	if action := update.AIStackHead().Type(); action >= ai.ACTION_CAST_SPELL_ON_OBJECT && action <= ai.ACTION_CAST_DURATION_SPELL {
		return false
	}
	h.StoreThreat(0)
	radius := pointOnLineSpill57C8A0(h.Balance("InversionRange") * 0.5)
	h.EachMissile(unit.PosVec, radius, func(missile *Object) bool {
		monsterUnitIsMagicMissile540B60(missile, unit, h.StoreThreat)
		return true
	})
	if h.LoadThreat() == 0 {
		return false
	}
	candidates := monsterFightSpellCandidates540B90(update, monsterInversionSpellMask5408D0, h.SpellAllowed)
	if len(candidates) == 0 {
		return false
	}
	id := candidates[h.Random(0, len(candidates)-1)]
	h.Cast(unit, id, unit)
	cooldown := h.Random(int(update.Field362_0), int(update.Field362_2))
	update.Field363 = h.Frame() + uint32(cooldown)
	return true
}

// MonsterCastInversion5408D0 binds the original shared scan DWORD, native
// map index, spell registry, logic RNG and ordinary MonsterCast action stack.
func (s *Server) MonsterCastInversion5408D0(unit *Object) bool {
	return MonsterCastInversionNative5408D0(unit, MonsterCastInversionRuntime5408D0{
		Frame:        s.Frame,
		Balance:      s.Balance.Float,
		EachMissile:  s.Map.EachMissileInCircle,
		LoadThreat:   func() uint32 { return memmap.Uint32(0x5D4594, 2489156) },
		StoreThreat:  storeMonsterInversionThreat5408D0,
		SpellAllowed: func(id spell.ID) bool { return s.Spells.HasFlags(id, things.SpellMobsCanCast) },
		Random:       s.Rand.Logic.IntClamp,
		Cast:         (*Object).MonsterCast,
	})
}

func storeMonsterInversionThreat5408D0(value uint32) {
	*memmap.PtrUint32(0x5D4594, 2489156) = value
}

func MonsterInversionMissile540B60(missile, target *Object) *Object {
	return monsterUnitIsMagicMissile540B60(missile, target, storeMonsterInversionThreat5408D0)
}
