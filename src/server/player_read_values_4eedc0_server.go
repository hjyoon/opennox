package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// PlayerReadValuesRuntime4EEDC0 supplies the services that remain outside the
// native-width player model. Protection tokens stay uint32 because they are
// opaque GAME.EXE values, while every Object, PlayerUpdateData, Player, and
// PlayerInfo pointer remains native-width.
type PlayerReadValuesRuntime4EEDC0 struct {
	SetHP         func(*Object, uint16)
	SoloMode      func() int32
	Ability       AbilityGivePlayerAllRuntime4EED40
	ProtectInt    func(uint32, uint32)
	ProtectUint16 func(uint32, uint16)
	WideLen       func(*PlayerInfo) uint32
	ProtectName   func(*PlayerInfo, uint32, uint32) int32
}

type playerReadValuesNativeDeps4EEDC0 struct {
	gameFlagsCheck func(uint32) int32
	setHP          func(*Object, uint16)
	soloMode       func() int32
	abilityGiveAll func(*Object, int8, int32)
	protectInt     func(uint32, uint32)
	protectUint16  func(uint32, uint16)
	wideLen        func(*PlayerInfo) uint32
	protectName    func(*PlayerInfo, uint32, uint32) int32
}

func playerReadValuesNative4EEDC0(
	unit *Object,
	rewardArg int32,
	players *serverPlayers,
	deps playerReadValuesNativeDeps4EEDC0,
) int32 {
	return playerReadValues4EEDC0(playerReadValuesHooks4EEDC0[
		*Object,
		*PlayerUpdateData,
		*Player,
		*ClassStats,
		*HealthData,
		*PlayerInfo,
		*Object,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(obj *Object) *PlayerUpdateData {
			// Do not use UpdateDataPlayer: 004EEDC0 has no class or nil gate.
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadBaseStats: func() *ClassStats {
			return players.BaseStats()
		},
		loadPlayerClass: func(player *Player) uint8 {
			return uint8(playerReadValuesInfoPtr4EEDC0(player).PlayerClass())
		},
		loadClassStats: func(class uint8) *ClassStats {
			switch player.Class(class) {
			case player.Warrior, player.Wizard, player.Conjurer:
				return players.ClassStats(player.Class(class))
			default:
				panic("invalid player class")
			}
		},
		loadStat:       playerReadValuesLoadStatNative4EEDC0,
		gameFlagsCheck: deps.gameFlagsCheck,
		floatToInt:     playerReadValuesRound4EEDC0,
		floatToInt16Abs: func(value float32) int16 {
			return int16(playerReadValuesRound4EEDC0(float32(math.Abs(float64(value)))))
		},
		loadHealthData: func(obj *Object) *HealthData {
			return obj.HealthData
		},
		storeHealthMax: func(health *HealthData, value uint16) {
			health.Max = value
		},
		loadHealthMax: func(health *HealthData) uint16 {
			return health.Max
		},
		setHP: deps.setHP,
		storeManaMax: func(update *PlayerUpdateData, value uint16) {
			update.ManaMax = value
		},
		loadManaMax: func(update *PlayerUpdateData) uint16 {
			return update.ManaMax
		},
		storeManaCurrent: func(update *PlayerUpdateData, value uint16) {
			update.ManaCur = value
		},
		storeStrength: func(player *Player, value uint32) {
			playerReadValuesInfoPtr4EEDC0(player).SetField2239(value)
		},
		loadStrength: func(player *Player) uint32 {
			return playerReadValuesInfoPtr4EEDC0(player).Field2239()
		},
		storeSpeedStat: func(player *Player, value uint32) {
			playerReadValuesInfoPtr4EEDC0(player).SetField2235(value)
		},
		loadSpeedStat: func(player *Player) uint32 {
			return playerReadValuesInfoPtr4EEDC0(player).Field2235()
		},
		storeSpeedBase: func(obj *Object, value float32) {
			obj.SpeedBase = value
		},
		loadLevel: func(player *Player) uint8 {
			return player.Level
		},
		soloMode: deps.soloMode,
		loadRewardArg: func() int32 {
			return rewardArg
		},
		abilityGiveAll: deps.abilityGiveAll,
		storeMass: func(obj *Object, value float32) {
			obj.Mass = value
		},
		floatToInt64Trunc: playerReadValuesTruncInt64_4EEDC0,
		storeCapacityWord: func(player *Player, value uint16) {
			player.field3652 = player.field3652&0xffff0000 | uint32(value)
		},
		loadCapacityWord: func(player *Player) uint16 {
			return uint16(player.field3652)
		},
		storeCarry: func(obj *Object, value uint16) {
			obj.CarryCapacity = value
		},
		loadStrengthToken: func(player *Player) uint32 {
			return player.ProtPlayerField2239
		},
		loadSpeedToken: func(player *Player) uint32 {
			return player.ProtPlayerField2235
		},
		loadManaMaxToken: func(player *Player) uint32 {
			return player.ProtUnitManaMax
		},
		loadHealthToken: func(player *Player) uint32 {
			return player.ProtUnitHPMax
		},
		protectInt:    deps.protectInt,
		protectUint16: deps.protectUint16,
		loadFirstItem: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadItemWeight: func(item *Object) uint8 {
			return item.Weight
		},
		loadNextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadCarry: func(obj *Object) uint16 {
			return obj.CarryCapacity
		},
		storeOverweight: func(player *Player, value uint32) {
			player.Field3656 = value
		},
		loadNameToken: func(player *Player) uint32 {
			return player.ProtPlayerOrigName
		},
		loadName: func(player *Player) *PlayerInfo {
			return playerReadValuesInfoPtr4EEDC0(player)
		},
		wideLen:     deps.wideLen,
		protectName: deps.protectName,
		storeInitialized: func(player *Player, value uint8) {
			*playerReadValuesInitializedPtr4EEDC0(player) = value
		},
	})
}

func playerReadValuesLoadStatNative4EEDC0(stats *ClassStats, stat playerReadValuesStat4EEDC0) float32 {
	switch stat {
	case playerReadValuesHealth4EEDC0:
		return stats.Health
	case playerReadValuesMana4EEDC0:
		return stats.Mana
	case playerReadValuesSpeed4EEDC0:
		return stats.Speed
	case playerReadValuesStrength4EEDC0:
		return stats.Strength
	default:
		panic("invalid player stat")
	}
}

func playerReadValuesInitializedPtr4EEDC0(player *Player) *uint8 {
	offset := unsafe.Offsetof(Player{}.info) - 1
	return (*uint8)(unsafe.Add(unsafe.Pointer(player), offset))
}

func playerReadValuesInfoPtr4EEDC0(player *Player) *PlayerInfo {
	return (*PlayerInfo)(unsafe.Pointer(&player.info[0]))
}

// playerReadValuesRound4EEDC0 models nox_float2int's default x87 FISTP
// rounding mode. Invalid and out-of-range inputs return integer-indefinite.
func playerReadValuesRound4EEDC0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// playerReadValuesTruncInt64_4EEDC0 models GAME.EXE 00566DCC, which changes
// the x87 control word to truncate toward zero for one signed-qword store.
func playerReadValuesTruncInt64_4EEDC0(value float64) int64 {
	if math.IsNaN(value) || value >= 9223372036854775808 || value < -9223372036854775808 {
		return math.MinInt64
	}
	return int64(math.Trunc(value))
}

func playerReadValuesServerDeps4EEDC0(
	server *Server,
	runtime PlayerReadValuesRuntime4EEDC0,
) playerReadValuesNativeDeps4EEDC0 {
	return playerReadValuesNativeDeps4EEDC0{
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		setHP:    runtime.SetHP,
		soloMode: runtime.SoloMode,
		abilityGiveAll: func(unit *Object, count int8, rewardArg int32) {
			server.AbilityGivePlayerAll4EED40(unit, count, rewardArg, runtime.Ability)
		},
		protectInt:    runtime.ProtectInt,
		protectUint16: runtime.ProtectUint16,
		wideLen:       runtime.WideLen,
		protectName:   runtime.ProtectName,
	}
}

// PlayerReadValues4EEDC0 binds GAME.EXE 004EEDC0 to native-width server
// structures while retaining the original signed low-byte and ABI32 token
// behavior.
func (s *Server) PlayerReadValues4EEDC0(
	unit *Object,
	rewardArg int32,
	runtime PlayerReadValuesRuntime4EEDC0,
) int32 {
	return playerReadValuesNative4EEDC0(
		unit,
		rewardArg,
		&s.Players,
		playerReadValuesServerDeps4EEDC0(s, runtime),
	)
}
