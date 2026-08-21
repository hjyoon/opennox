package server

import (
	"math"
	"unsafe"
)

// PlayerUnitInitRuntime4EFE80 supplies services that remain owned by the
// legacy runtime. Object, UpdateData, and Player pointers remain native-width;
// gold, flags, rewards, controls, and the converted life count retain their
// original fixed-width scalar contracts.
type PlayerUnitInitRuntime4EFE80 struct {
	ProtectGold           func(uint32, int32)
	SyncLevel             func(*Object)
	AwardBeastScrolls     func(*Player)
	AwardSpells           func(*Player)
	ReadValues            func(*Object, int32)
	AwardWarriorAbilities func(*Player)
	GameFlag              func(uint32) int32
	BalanceFloat          func(string) float32
	MakeDefaultItems      func(*Object, int32, int32) uint8
}

func playerUnitInitGetGoldNative4EFE80(unit *Object) uint32 {
	// GAME.EXE 004FA6D0 tests only the low ObjClass byte before following
	// UpdateData and Player. Preserve that gate instead of broadening it to a
	// native-width class comparison.
	if unit == nil || uint8(unit.ObjClass)&0x04 == 0 {
		return 0
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	return update.Player.GoldVal
}

func playerUnitInitSubGoldNative4EFE80(
	unit *Object,
	amount uint32,
	protect func(token uint32, delta int32),
) {
	update := (*PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	gold := player.GoldVal
	if gold >= amount {
		player.GoldVal = gold - amount
	} else {
		player.GoldVal = 0
	}

	// 004FA5D0 reloads Player from its own cached UpdateData for the
	// protection token and negates amount in uint32 arithmetic before passing
	// C int.
	player = update.Player
	protect(player.ProtPlayerGold, int32(uint32(0)-amount))
}

// playerUnitInitFloatToInt4EFE80 models nox_float2int at GAME.EXE 00419A70:
// x87 FISTP under the default round-to-nearest-even mode. Invalid and
// out-of-range conversions return the signed integer-indefinite value.
func playerUnitInitFloatToInt4EFE80(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func playerUnitInitNative4EFE80(unit *Object, runtime PlayerUnitInitRuntime4EFE80) uint8 {
	return playerUnitInit4EFE80(playerUnitInitHooks4EFE80[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		getGold: playerUnitInitGetGoldNative4EFE80,
		subGold: func(unit *Object, amount uint32) {
			playerUnitInitSubGoldNative4EFE80(unit, amount, runtime.ProtectGold)
		},
		syncLevel:             runtime.SyncLevel,
		loadPlayer:            func(update *PlayerUpdateData) *Player { return update.Player },
		awardBeastScrolls:     runtime.AwardBeastScrolls,
		awardSpells:           runtime.AwardSpells,
		readValues:            runtime.ReadValues,
		awardWarriorAbilities: runtime.AwardWarriorAbilities,
		gameFlag:              runtime.GameFlag,
		balanceFloat:          runtime.BalanceFloat,
		floatToInt:            playerUnitInitFloatToInt4EFE80,
		storeExtraLives: func(update *PlayerUpdateData, value int32) {
			update.ExtraLives = uint32(value)
		},
		makeDefaultItems: runtime.MakeDefaultItems,
	})
}

// PlayerUnitInit4EFE80 binds GAME.EXE 004EFE80 to native-width Object,
// PlayerUpdateData, and Player storage while retaining the original callback
// and binary32 conversion order.
func (s *Server) PlayerUnitInit4EFE80(unit *Object, runtime PlayerUnitInitRuntime4EFE80) uint8 {
	return playerUnitInitNative4EFE80(unit, runtime)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(PlayerUpdateData{}.ExtraLives)]
)
