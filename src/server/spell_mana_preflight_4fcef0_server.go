package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	spellManaPreflightTableBase4FCEF0   = uintptr(0x587000)
	spellManaPreflightTableOffset4FCEF0 = uintptr(217668)
)

type spellManaPreflightNativeDeps4FCEF0 struct {
	loadGodMode   func() int32
	loadOldMana   func(*Object) uint16
	summonCost    func(int32, *Object) int32
	spellManaCost func(int32, int32) int32
}

func spellManaPreflightNative4FCEF0(
	unit *Object,
	sequence *int32,
	count int32,
	deps spellManaPreflightNativeDeps4FCEF0,
) int32 {
	return spellManaPreflight4FCEF0(spellManaPreflightHooks4FCEF0[*Object, *int32]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadSequenceArg: func() *int32 {
			return sequence
		},
		loadCountArg: func() int32 {
			return count
		},
		loadGodMode: deps.loadGodMode,
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadOldMana: deps.loadOldMana,
		loadSpell: func(sequence *int32, index int32) int32 {
			return *(*int32)(unsafe.Add(unsafe.Pointer(sequence), uintptr(index)*unsafe.Sizeof(*sequence)))
		},
		summonCost:    deps.summonCost,
		spellManaCost: deps.spellManaCost,
	})
}

func spellManaPreflightServerDeps4FCEF0(s *Server) spellManaPreflightNativeDeps4FCEF0 {
	return spellManaPreflightNativeDeps4FCEF0{
		loadGodMode: func() int32 {
			if noxflags.HasEngine(noxflags.EngineGodMode) {
				return 1
			}
			return 0
		},
		loadOldMana: UnitGetOldMana4EEC80,
		summonCost: func(spellID int32, unit *Object) int32 {
			if uint8(unit.ObjClass)&uint8(object.ClassPlayer) == 0 {
				return 0
			}
			return int32(memmap.Uint32(
				spellManaPreflightTableBase4FCEF0,
				spellManaPreflightTableOffset4FCEF0+4*uintptr(spellID),
			))
		},
		spellManaCost: func(spellID, costType int32) int32 {
			return int32(s.Spells.ManaCost(spell.ID(spellID), int(costType)))
		},
	}
}

// SpellManaPreflight4FCEF0 binds GAME.EXE 004FCEF0 to native Object and
// sequence pointers. sequence points at the first signed 32-bit spell ID;
// positive count is intentionally not capped here because the original
// function walks exactly count entries.
//
//go:noinline
func (s *Server) SpellManaPreflight4FCEF0(unit *Object, sequence *int32, count int32) int32 {
	return spellManaPreflightNative4FCEF0(unit, sequence, count, spellManaPreflightServerDeps4FCEF0(s))
}
