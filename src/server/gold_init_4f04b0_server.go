package server

import "unsafe"

type goldInitNativeDeps4F04B0 struct {
	firstPlayer func() *Player
	nextPlayer  func(*Player) *Player
	randomInt   func(int32, int32, string, int32) int32
}

func goldInitNative4F04B0(unit *Object, deps goldInitNativeDeps4F04B0) int32 {
	return goldInit4F04B0(goldInitHooks4F04B0[*Object, *GoldInitData, *Player]{
		loadUnitArg: func() (*Object, int32) {
			// GAME.EXE leaves the object argument in EAX on the nonzero-Amount
			// path. Preserve only that numeric low dword; it is never converted
			// back into a pointer.
			return unit, int32(uint32(uintptr(unsafe.Pointer(unit))))
		},
		loadInitData: func(unit *Object) *GoldInitData {
			return (*GoldInitData)(unit.InitData)
		},
		loadAmount: func(data *GoldInitData) uint32 {
			return data.Amount
		},
		firstPlayer: deps.firstPlayer,
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadExperience: func(unit *Object) float32 {
			return unit.Experience
		},
		nextPlayer:    deps.nextPlayer,
		truncQwordLow: goldInitTruncQwordLow4F04B0,
		randomInt:     deps.randomInt,
		storeAmount: func(data *GoldInitData, amount uint32) {
			data.Amount = amount
		},
	})
}

// GoldInit4F04B0 binds GAME.EXE 004F04B0 to native-width Object and Player
// pointers, fixed-width GoldInitData, the server's active-player traversal,
// and its logic RNG. Debug path and line values remain in the semantic call
// contract even though the RNG service does not consume them. There are
// deliberately no nil guards.
//
//go:noinline
func (s *Server) GoldInit4F04B0(unit *Object) int32 {
	return goldInitNative4F04B0(unit, goldInitNativeDeps4F04B0{
		firstPlayer: s.Players.First,
		nextPlayer:  s.Players.Next,
		randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
	})
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(GoldInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(GoldInitData{}.Amount)]
)
