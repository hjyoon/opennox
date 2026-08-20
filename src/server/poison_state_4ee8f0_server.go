package server

import (
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
)

// PoisonStateRuntime4EE8F0 keeps the three status/network services whose
// bodies remain separate restoration scopes. Object and player traversal is
// performed in package server with native-width pointers.
type PoisonStateRuntime4EE8F0 struct {
	NeedPlayerStatus  func(*Player, uint32)
	UnsetPlayerStatus func(*Player, uint32)
	ReportPoison      func(*Object, *Object, int32)
}

type poisonStateNativeDeps4EE8F0 struct {
	needPlayerStatus  func(*Player, uint32)
	unsetPlayerStatus func(*Player, uint32)
	priorityMessage   func(*Object, string, uint8)
	gameFlag          func(uint32) int32
	playerByIndex     func(int32) *Player
	reportPoison      func(*Object, *Object, int32)
	frame             func() uint32
}

func poisonClearNativeHooks4EE8F0(
	unit *Object,
	amount int32,
	deps poisonStateNativeDeps4EE8F0,
) poisonClearHooks4EE8F0[*Object, *HealthData, *PlayerUpdateData, *Player, *Player] {
	return poisonClearHooks4EE8F0[*Object, *HealthData, *PlayerUpdateData, *Player, *Player]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadAmountArg: func() int32 {
			return amount
		},
		loadPoison: func(unit *Object) uint8 {
			return unit.Poison540
		},
		storePoison: func(unit *Object, value uint8) {
			unit.Poison540 = value
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		clearHealthFrame: func(health *HealthData) {
			health.Field16 = 0
		},
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		unsetPlayerStatus: deps.unsetPlayerStatus,
		priorityMessage:   deps.priorityMessage,
		gameFlag:          deps.gameFlag,
		loadSubClass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		playerInfoByIndex: deps.playerByIndex,
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadOwner: func(unit *Object) *Object {
			return unit.ObjOwner
		},
		reportPoison: deps.reportPoison,
	}
}

func updatePoisonNative4EE8F0(unit *Object, amount int32, deps poisonStateNativeDeps4EE8F0) {
	updatePoison4EE8F0(poisonClearNativeHooks4EE8F0(unit, amount, deps))
}

func removePoisonNative4EE9D0(unit *Object, deps poisonStateNativeDeps4EE8F0) {
	removePoison4EE9D0(poisonClearNativeHooks4EE8F0(unit, 0, deps))
}

func setPoisonNative4EEA90(unit *Object, value int32, deps poisonStateNativeDeps4EE8F0) {
	setPoison4EEA90(poisonSetHooks4EEA90[*Object, *HealthData, *PlayerUpdateData, *Player, *Player]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadCurrent: func(unit *Object) uint8 {
			return unit.Poison540
		},
		loadValueArg: func() int32 {
			return value
		},
		loadHealth: func(unit *Object) *HealthData {
			return unit.HealthData
		},
		frame: deps.frame,
		storeHealthFrame: func(health *HealthData, frame uint32) {
			health.Field16 = frame
		},
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		storePoison: func(unit *Object, value uint8) {
			unit.Poison540 = value
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		needPlayerStatus:  deps.needPlayerStatus,
		unsetPlayerStatus: deps.unsetPlayerStatus,
		gameFlag:          deps.gameFlag,
		loadSubClass: func(unit *Object) uint32 {
			return uint32(unit.ObjSubClass)
		},
		playerInfoByIndex: deps.playerByIndex,
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadOwner: func(unit *Object) *Object {
			return unit.ObjOwner
		},
		reportPoison: deps.reportPoison,
		storePoisonTimer: func(unit *Object, timer uint16) {
			unit.Field542 = timer
		},
	})
}

func (s *Server) poisonStateNativeDeps4EE8F0(runtime PoisonStateRuntime4EE8F0) poisonStateNativeDeps4EE8F0 {
	return poisonStateNativeDeps4EE8F0{
		needPlayerStatus:  runtime.NeedPlayerStatus,
		unsetPlayerStatus: runtime.UnsetPlayerStatus,
		priorityMessage: func(unit *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), value)
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		playerByIndex: func(index int32) *Player {
			return s.Players.ByInd(ntype.PlayerInd(index))
		},
		reportPoison: runtime.ReportPoison,
		frame:        s.Frame,
	}
}

func (s *Server) UpdatePoison4EE8F0(unit *Object, amount int32, runtime PoisonStateRuntime4EE8F0) {
	updatePoisonNative4EE8F0(unit, amount, s.poisonStateNativeDeps4EE8F0(runtime))
}

func (s *Server) RemovePoison4EE9D0(unit *Object, runtime PoisonStateRuntime4EE8F0) {
	removePoisonNative4EE9D0(unit, s.poisonStateNativeDeps4EE8F0(runtime))
}

func (s *Server) SetPoison4EEA90(unit *Object, value int32, runtime PoisonStateRuntime4EE8F0) {
	setPoisonNative4EEA90(unit, value, s.poisonStateNativeDeps4EE8F0(runtime))
}
