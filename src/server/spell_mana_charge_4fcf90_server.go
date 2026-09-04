package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const (
	spellManaChargeTableBase4FCF90   = uintptr(0x587000)
	spellManaChargeTableOffset4FCF90 = uintptr(217668)
)

type spellManaChargeNativeDeps4FCF90 struct {
	loadGodMode   func() bool
	summonCost    func(int32, *Object) int32
	spellManaCost func(int32, int32) int32
	subtractMana  func(*Object, int32)
	loadTickRate  func() uint32
}

func spellManaChargeNative4FCF90(
	unit *Object,
	spellID int32,
	costType int32,
	deps spellManaChargeNativeDeps4FCF90,
) int32 {
	return spellManaCharge4FCF90(spellManaChargeHooks4FCF90[*Object, *PlayerUpdateData]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadSpellArg: func() int32 {
			return spellID
		},
		loadGodMode: deps.loadGodMode,
		loadCostTypeArg: func() int32 {
			return costType
		},
		summonCost:    deps.summonCost,
		spellManaCost: deps.spellManaCost,
		loadCurrentMana: func(update *PlayerUpdateData) uint16 {
			return update.ManaCur
		},
		subtractMana: deps.subtractMana,
		storeRechargeCost: func(update *PlayerUpdateData, value uint16) {
			update.Field20_0 = value
		},
		loadTickRate: deps.loadTickRate,
		storeRechargeFrame: func(update *PlayerUpdateData, value uint16) {
			update.Field20_1 = value
		},
	})
}

func spellManaChargeServerDeps4FCF90(
	s *Server,
	subtractMana func(*Object, int32),
) spellManaChargeNativeDeps4FCF90 {
	return spellManaChargeNativeDeps4FCF90{
		loadGodMode: func() bool {
			return noxflags.HasEngine(noxflags.EngineGodMode)
		},
		summonCost: func(spellID int32, unit *Object) int32 {
			// sub_500CA0 rechecks the live Player bit even though 004FCF90
			// already gated the cached unit before calling it.
			if unit == nil || uint8(unit.ObjClass)&uint8(object.ClassPlayer) == 0 {
				return 0
			}
			return int32(memmap.Uint32(
				spellManaChargeTableBase4FCF90,
				spellManaChargeTableOffset4FCF90+4*uintptr(spellID),
			))
		},
		spellManaCost: func(spellID, costType int32) int32 {
			return int32(s.Spells.ManaCost(spell.ID(spellID), int(costType)))
		},
		subtractMana: subtractMana,
		loadTickRate: s.TickRate,
	}
}

// SpellManaCharge4FCF90 binds GAME.EXE 004FCF90 to native Object and player
// update layouts. subtractMana owns the still-separate 004EEBF0 service; its
// return value is intentionally discarded, as in the original caller.
//
//go:noinline
func (s *Server) SpellManaCharge4FCF90(
	unit *Object,
	spellID int32,
	costType int32,
	subtractMana func(*Object, int32),
) int32 {
	return spellManaChargeNative4FCF90(
		unit,
		spellID,
		costType,
		spellManaChargeServerDeps4FCF90(s, subtractMana),
	)
}
