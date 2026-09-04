package server

import (
	"github.com/opennox/libs/spell"
)

type spellPrecheckNativeDeps4FD0E0 struct {
	spellFlags            func(int32) uint32
	spellEnabled          func(int32) int32
	checkPlayerSpellClass func(uint8, int32) int32
	summonAllowed         func(int32, *Object) int32
}

func spellPrecheckNative4FD0E0(
	unit *Object,
	spellID int32,
	deps spellPrecheckNativeDeps4FD0E0,
) int32 {
	return spellPrecheck4FD0E0(spellPrecheckHooks4FD0E0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadSpellArg: func() int32 {
			return spellID
		},
		spellFlags: deps.spellFlags,
		loadUnitArg: func() *Object {
			return unit
		},
		findParentPlayer: (*Object).FindOwnerChainPlayer,
		spellEnabled:     deps.spellEnabled,
		loadUnitClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			// Access the byte directly: Player.Info and PlayerClass are nil-safe,
			// while GAME.EXE faults on a nil Player at this exact load.
			return player.info[66]
		},
		checkPlayerSpellClass: deps.checkPlayerSpellClass,
		summonAllowed:         deps.summonAllowed,
	})
}

func spellPrecheckServerDeps4FD0E0(s *Server) spellPrecheckNativeDeps4FD0E0 {
	return spellPrecheckNativeDeps4FD0E0{
		spellFlags: func(spellID int32) uint32 {
			return uint32(s.Spells.Flags(spell.ID(spellID)))
		},
		spellEnabled: func(spellID int32) int32 {
			if s.Spells.DefByInd(spell.ID(spellID)).IsEnabled() {
				return 1
			}
			return 0
		},
		checkPlayerSpellClass: func(class uint8, spellID int32) int32 {
			// 0057AEA0 performs its own live flags lookup before selecting the
			// class mask, so this must not reuse the priming lookup above.
			flags := s.Spells.Flags(spell.ID(spellID))
			return rewardSpellClassCheck4F09F0(flags, uint32(class))
		},
		summonAllowed: func(spellID int32, owner *Object) int32 {
			if Sub_57AEE0(spell.ID(spellID), owner) {
				return 1
			}
			return 0
		},
	}
}

// SpellPrecheck4FD0E0 binds GAME.EXE 004FD0E0 to native-width Object,
// PlayerUpdateData, and Player pointers. Spell IDs and result codes retain
// their original signed 32-bit ABI. There are deliberately no nil guards.
//
//go:noinline
func (s *Server) SpellPrecheck4FD0E0(unit *Object, spellID spell.ID) int32 {
	return spellPrecheckNative4FD0E0(unit, int32(spellID), spellPrecheckServerDeps4FD0E0(s))
}
