package server

// PlayerFollowLastTargetRuntime4F9E10 supplies the retained camera-toggle
// service. All Object, update-data, and Player traversal stays native-width.
type PlayerFollowLastTargetRuntime4F9E10 struct {
	CameraFollow func(*Object, *Object)
}

func playerFollowLastTargetNative4F9E10(
	unit *Object,
	runtime PlayerFollowLastTargetRuntime4F9E10,
) int32 {
	return playerFollowLastTarget4F9E10(unit, playerFollowLastTargetHooks4F9E10[
		*Object, *PlayerUpdateData, *Player,
	]{
		loadLastTarget: func(unit *Object) *Object {
			return unit.Obj130
		},
		findOwnerChainPlayer: func(target *Object) *Object {
			return target.FindOwnerChainPlayer()
		},
		loadFlagsByte: func(target *Object) uint8 {
			return uint8(target.ObjFlags)
		},
		loadClassByte: func(target *Object) uint8 {
			return uint8(target.ObjClass)
		},
		loadUpdateData: func(target *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(target.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		cameraFollow: runtime.CameraFollow,
	})
}

// PlayerFollowLastTarget4F9E10 selects the attributed live unit for the
// player's observer camera using native-width links throughout.
//
//go:noinline
func (s *Server) PlayerFollowLastTarget4F9E10(
	unit *Object,
	runtime PlayerFollowLastTargetRuntime4F9E10,
) int32 {
	_ = s
	return playerFollowLastTargetNative4F9E10(unit, runtime)
}
