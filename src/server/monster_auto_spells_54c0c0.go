package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type monsterAutoSpellsKind54C0C0 uint8

const (
	monsterAutoSpellsDefault54C0C0 monsterAutoSpellsKind54C0C0 = iota
	monsterAutoSpellsUrchinShaman54C0C0
	monsterAutoSpellsWizard54C0C0
	monsterAutoSpellsBeholder54C0C0
	monsterAutoSpellsLich54C0C0
	monsterAutoSpellsLichLord54C0C0
	monsterAutoSpellsDemon54C0C0
	monsterAutoSpellsWizardGreen54C0C0
	monsterAutoSpellsWillOWisp54C0C0
)

func setMonsterAutoSpellsInitialized54C0C0(update *MonsterUpdateData) {
	update.Field509 = update.Field509&^0xff | 1
}

// monsterAutoSpells54C0C0 preserves the type-specific writes from
// GAME.EXE 0054C0C0..0054C474 using the native MonsterUpdateData layout.
// The original writes only the low byte of Field509 and returns a short whose
// incidental value differs between branches, so both details stay explicit.
func monsterAutoSpells54C0C0(
	typeInd uint16,
	kind monsterAutoSpellsKind54C0C0,
	update *MonsterUpdateData,
	fps uint16,
	quest bool,
) uint16 {
	ret := typeInd
	switch kind {
	case monsterAutoSpellsUrchinShaman54C0C0:
		update.Field510 = 2
		update.Field410 = 0x08000000
		update.Field362_0 = 0
		update.Field362_2 = fps >> 1
		update.Field384 = 0x20000000
		update.Field366_0 = 3 * fps
		update.Field368_0 = 0
		update.Field366_2 = 5 * fps
		update.Field430 = 0x40000000
		update.Field424 = 0x40000000
		update.Field438 = 0x80000000
		update.Field368_2 = 3 * fps
		update.Field370_0 = fps
		ret = fps
		setMonsterAutoSpellsInitialized54C0C0(update)
		update.Field370_2 = 3 * fps
		return ret
	case monsterAutoSpellsWizard54C0C0:
		update.Field510 = 3
		update.Field385 = 0x08000000
		update.Field410 = 0x08000000
		update.Field411 = 0x10000000
		update.Field444 = 0x20000000
		update.Field399 = 0x40000000
		update.Field422 = 0x40000000
		update.Field415 = 0x40000000
		update.Field396 = 0x40000000
		update.Field376 = 0x80000000
		ret = 0
	case monsterAutoSpellsBeholder54C0C0:
		update.Field510 = 3
		update.Field396 = 0x40000000
		if !quest {
			update.Field376 = 0x80000000
			ret = 0
		} else {
			ret = 1
		}
	case monsterAutoSpellsLich54C0C0:
		update.Field510 = 3
		update.Field385 = 0x08000000
		update.Field410 = 0x08000000
		update.Field443 = 0x10000000
		update.Field444 = 0x20000000
		update.Field405 = 0x20000000
		update.Field411 = 0x80000000
		update.Field399 = 0x40000000
		update.Field415 = 0x40000000
		update.Field396 = 0x40000000
		ret = 0
	case monsterAutoSpellsLichLord54C0C0:
		update.Field510 = 3
		update.Field385 = 0x08000000
		update.Field410 = 0x08000000
		update.Field443 = 0x10000000
		update.Field368_0 = 3 * fps
		update.Field403 = 0x40000000
		update.Field411 = 0x80000000
		ret = fps
		setMonsterAutoSpellsInitialized54C0C0(update)
		update.Field368_2 = 5 * fps
		return ret
	case monsterAutoSpellsDemon54C0C0:
		update.Field510 = 3
		update.Field385 = 0x08000000
		update.Field446 = 0x20000000
		update.Field399 = 0x40000000
		update.Field382 = 0x40000000
		update.Field411 = 0x80000000
		ret = 0
	case monsterAutoSpellsWizardGreen54C0C0:
		update.Field510 = 3
		update.Field410 = 0x08000000
		update.Field384 = 0x20000000
		update.Field430 = 0x40000000
		update.Field432 = 0x40000000
		update.Field422 = 0x40000000
		update.Field376 = 0x80000000
		ret = 0
	case monsterAutoSpellsWillOWisp54C0C0:
		update.Field510 = 3
		update.Field415 = 0x40000000
	}
	setMonsterAutoSpellsInitialized54C0C0(update)
	return ret
}

func (s *Server) monsterAutoSpellsKind54C0C0(typeInd uint16) monsterAutoSpellsKind54C0C0 {
	match := func(id string) bool {
		ind := s.Types.IndByID(id)
		return ind != 0 && uint16(ind) == typeInd
	}
	switch {
	case match("UrchinShaman"):
		return monsterAutoSpellsUrchinShaman54C0C0
	case match("Wizard"), match("WizardWhite"):
		return monsterAutoSpellsWizard54C0C0
	case match("Beholder"):
		return monsterAutoSpellsBeholder54C0C0
	case match("Lich"):
		return monsterAutoSpellsLich54C0C0
	case match("LichLord"):
		return monsterAutoSpellsLichLord54C0C0
	case match("Demon"):
		return monsterAutoSpellsDemon54C0C0
	case match("WizardGreen"):
		return monsterAutoSpellsWizardGreen54C0C0
	case match("WillOWisp"):
		return monsterAutoSpellsWillOWisp54C0C0
	default:
		return monsterAutoSpellsDefault54C0C0
	}
}

// MonsterAutoSpells54C0C0 binds the restored writes to a native monster
// object. This avoids reading Object.UpdateData at the PE32 +748 slot on
// 64-bit targets.
func (s *Server) MonsterAutoSpells54C0C0(unit *Object) uint16 {
	if unit == nil {
		return 0
	}
	return monsterAutoSpells54C0C0(
		unit.TypeInd,
		s.monsterAutoSpellsKind54C0C0(unit.TypeInd),
		unit.UpdateDataMonster(),
		uint16(s.TickRate()),
		noxflags.HasGame(noxflags.GameModeQuest),
	)
}
