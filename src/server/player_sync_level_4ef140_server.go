package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// PlayerSyncLevelRuntime4EF140 supplies the two legacy services whose fixed
// width contracts remain outside the native Object and Player layouts.
type PlayerSyncLevelRuntime4EF140 struct {
	ProtectLevel func(uint32, uint8)
	ReadValues   func(*Object, int32) int32
}

type playerSyncLevelNativeDeps4EF140 struct {
	gameFlagsCheck func(uint32) int32
	loadXPTable    func(int32) float64
	protectLevel   func(uint32, uint8)
	readValues     func(*Object, int32) int32
}

func playerSyncLevelNative4EF140(
	unit *Object,
	deps playerSyncLevelNativeDeps4EF140,
) int32 {
	return playerSyncLevel4EF140(playerSyncLevelHooks4EF140[
		*Object,
		*PlayerUpdateData,
		*Player,
		int32,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			// 004EF140 has no nil or class gate before this native pointer load.
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		gameFlagsCheck: deps.gameFlagsCheck,
		loadXPTable:    deps.loadXPTable,
		loadExperience: func(unit *Object) float32 {
			return unit.Experience
		},
		loadLevelProtection: func(player *Player) uint32 {
			return player.ProtPlayerLevel
		},
		storeLevel: func(player *Player, level uint8) {
			player.Level = level
		},
		protectLevel: deps.protectLevel,
		readValues:   deps.readValues,
	})
}

func playerSyncLevelServerDeps4EF140(
	server *Server,
	runtime PlayerSyncLevelRuntime4EF140,
) playerSyncLevelNativeDeps4EF140 {
	return playerSyncLevelNativeDeps4EF140{
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		loadXPTable: func(index int32) float64 {
			return server.Balance.FloatInd("XPTable", int(index))
		},
		protectLevel: runtime.ProtectLevel,
		readValues:   runtime.ReadValues,
	}
}

// PlayerSyncLevel4EF140 binds GAME.EXE 004EF140 to native pointer-width
// Object, PlayerUpdateData, and Player layouts. Protection tokens remain
// fixed-width opaque values and the final read-values result remains int32.
func (s *Server) PlayerSyncLevel4EF140(
	unit *Object,
	runtime PlayerSyncLevelRuntime4EF140,
) int32 {
	return playerSyncLevelNative4EF140(
		unit,
		playerSyncLevelServerDeps4EF140(s, runtime),
	)
}
