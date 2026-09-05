package server

import "github.com/opennox/libs/spell"

type spellDurationCancelNativeDeps4FE9D0 struct {
	loadCaster        func(*DurSpell) *Object
	loadClassLowByte  func(*Object) byte
	loadSpell         func(*DurSpell) uint32
	loadUpdate        func(*Object) *PlayerUpdateData
	loadPlayer        func(*PlayerUpdateData) *Player
	loadPlayerIndex   func(*Player) byte
	reportSpellStat   func(byte, uint32, byte)
	loadSub108        func(*DurSpell) *DurSpell
	loadTarget        func(*DurSpell) *Object
	stopRay           func(*DurSpell, *Object)
	loadNext          func(*DurSpell) *DurSpell
	loadFlagsLowByte  func(*DurSpell) byte
	storeFlagsLowByte func(*DurSpell, byte)
}

func spellDurationCancelNative4FE9D0(
	record *DurSpell,
	deps spellDurationCancelNativeDeps4FE9D0,
) byte {
	return SpellDurationCancel4FE9D0(record, SpellDurationCancelHooks4FE9D0[
		*DurSpell,
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		LoadCaster:        deps.loadCaster,
		LoadClassLowByte:  deps.loadClassLowByte,
		LoadSpell:         deps.loadSpell,
		LoadUpdate:        deps.loadUpdate,
		LoadPlayer:        deps.loadPlayer,
		LoadPlayerIndex:   deps.loadPlayerIndex,
		ReportSpellStat:   deps.reportSpellStat,
		LoadSub108:        deps.loadSub108,
		LoadTarget:        deps.loadTarget,
		StopRay:           deps.stopRay,
		LoadNext:          deps.loadNext,
		LoadFlagsLowByte:  deps.loadFlagsLowByte,
		StoreFlagsLowByte: deps.storeFlagsLowByte,
	})
}

func spellDurationCancelServerDeps4FE9D0(sp *SpellsDuration) spellDurationCancelNativeDeps4FE9D0 {
	return spellDurationCancelNativeDeps4FE9D0{
		loadCaster: func(record *DurSpell) *Object {
			return record.Caster16
		},
		loadClassLowByte: func(object *Object) byte {
			return byte(object.ObjClass)
		},
		loadSpell: func(record *DurSpell) uint32 {
			return record.Spell
		},
		loadUpdate: func(object *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(object.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) byte {
			return player.PlayerInd
		},
		reportSpellStat: func(index byte, spellID uint32, status byte) {
			_ = sp.s.NetReportSpellStat(int(index), spell.ID(spellID), status)
		},
		loadSub108: func(record *DurSpell) *DurSpell {
			return record.Sub108
		},
		loadTarget: func(record *DurSpell) *Object {
			return record.Target48
		},
		stopRay: func(record *DurSpell, target *Object) {
			sp.s.NetStopRaySpell(record, target)
		},
		loadNext: func(record *DurSpell) *DurSpell {
			return record.Next
		},
		loadFlagsLowByte: func(record *DurSpell) byte {
			return byte(record.Flags88)
		},
		storeFlagsLowByte: func(record *DurSpell, flags byte) {
			record.Flags88 = record.Flags88&^0xff | uint32(flags)
		},
	}
}

// SpellDurationCancel4FE9D0 binds GAME.EXE 004FE9D0 to native-width duration,
// object, update-data, and player pointers. Field access stays raw so no
// nil/dead-object guard is added around the executable's original fault
// boundaries. Only the low byte of Flags88 is updated and returned.
//
//go:noinline
func (sp *SpellsDuration) SpellDurationCancel4FE9D0(record *DurSpell) byte {
	return spellDurationCancelNative4FE9D0(record, spellDurationCancelServerDeps4FE9D0(sp))
}
