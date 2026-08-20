package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

type playerRespawnStateResetNativeDeps4EF660 struct {
	gameFlag    func(uint32) int32
	countGlyphs func(*Object) int32
}

func playerRespawnStateResetNative4EF660(
	unit *Object,
	deps playerRespawnStateResetNativeDeps4EF660,
) *Player {
	return playerRespawnStateReset4EF660(unit, playerRespawnStateResetHooks4EF660[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			// Do not use UpdateDataPlayer: 004EF660 has no class or nil gate.
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		storePendingObject: func(update *PlayerUpdateData, index int, value *Object) {
			update.Field29[index] = value
		},
		storeSoulGate: func(update *PlayerUpdateData, value *Object) {
			update.SoulGate = value
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		storeQuestAnkh: func(player *Player, index int, value *Object) {
			player.QuestAnkhs[index] = value
		},
		gameFlag:    deps.gameFlag,
		countGlyphs: deps.countGlyphs,
		storeCurTrapsLowByte: func(update *PlayerUpdateData, value uint8) {
			update.CurTraps = update.CurTraps&^uint32(0xff) | uint32(value)
		},
		storeField66: func(update *PlayerUpdateData, value uint32) {
			update.Field66 = value
		},
		storeAttribution: func(unit, value *Object) {
			unit.Obj130 = value
		},
		resetPlayerMarkers: func(player *Player) *Player {
			player.Sub422140()
			return player
		},
	})
}

// playerRespawnGlyphCount4EF6F0 is the pointer-native inventory portion of
// GAME.EXE 004EF6F0 used by the reset. It intentionally counts matching
// Destroyed items too and loads each live successor only after comparing the
// current item's zero-extended TypeInd. The surrounding cache/name lookup is
// supplied by Server.Types.GlyphID and is audited with 004EF6F0 itself.
func playerRespawnGlyphCount4EF6F0(unit *Object, glyphType uint32) int32 {
	var count int32
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if uint32(item.TypeInd) == glyphType {
			count++
		}
	}
	return count
}

func playerRespawnStateResetServerDeps4EF660(s *Server) playerRespawnStateResetNativeDeps4EF660 {
	return playerRespawnStateResetNativeDeps4EF660{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		countGlyphs: func(unit *Object) int32 {
			return playerRespawnGlyphCount4EF6F0(unit, uint32(s.Types.GlyphID()))
		},
	}
}

// PlayerRespawnStateReset4EF660 binds GAME.EXE 004EF660 to native-width
// Object, PlayerUpdateData, Player, pending-object, SoulGate, attribution, and
// Quest-Ankh pointers. Fixed-width game flags, counters, and low-byte stores
// retain their original widths.
func (s *Server) PlayerRespawnStateReset4EF660(unit *Object) *Player {
	return playerRespawnStateResetNative4EF660(
		unit,
		playerRespawnStateResetServerDeps4EF660(s),
	)
}
