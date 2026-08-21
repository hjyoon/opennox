package server

import "unsafe"

// PlayerUnitInitRuntime4EFE80 supplies services that remain owned by the
// legacy runtime. Object, UpdateData, and Player pointers remain native-width;
// gold, flags, rewards, controls, and the converted life count retain their
// original fixed-width scalar contracts.
type PlayerUnitInitRuntime4EFE80 struct {
	GetGold               func(*Object) uint32
	SubGold               func(*Object, uint32)
	SyncLevel             func(*Object)
	AwardBeastScrolls     func(*Player)
	AwardSpells           func(*Player)
	ReadValues            func(*Object, int32)
	AwardWarriorAbilities func(*Player)
	GameFlag              func(uint32) int32
	BalanceFloat          func(string) float32
	FloatToInt            func(float32) int32
	MakeDefaultItems      func(*Object, int32, int32) uint8
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
		getGold:               runtime.GetGold,
		subGold:               runtime.SubGold,
		syncLevel:             runtime.SyncLevel,
		loadPlayer:            func(update *PlayerUpdateData) *Player { return update.Player },
		awardBeastScrolls:     runtime.AwardBeastScrolls,
		awardSpells:           runtime.AwardSpells,
		readValues:            runtime.ReadValues,
		awardWarriorAbilities: runtime.AwardWarriorAbilities,
		gameFlag:              runtime.GameFlag,
		balanceFloat:          runtime.BalanceFloat,
		floatToInt:            runtime.FloatToInt,
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
