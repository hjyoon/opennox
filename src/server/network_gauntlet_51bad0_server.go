package server

// NetworkGauntletPacketSize51BAD0 is the exact MSG_GAUNTLET packet width.
const NetworkGauntletPacketSize51BAD0 = int(networkGauntletPacketSize51BAD0)

type NetworkGauntletRuntime51BAD0 struct {
	Respawn func(*Object)
	Exit    func(*Object)
}

// NetworkGauntlet51BAD0 binds the original packet branch to native-width
// Object, PlayerUpdateData, and Player pointers.
func (s *Server) NetworkGauntlet51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkGauntletPacketSize51BAD0]byte,
	runtime NetworkGauntletRuntime51BAD0,
) int32 {
	return networkGauntlet51BAD0(unit, update, networkGauntletHooks51BAD0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadSubtype: func() uint8 {
			return packet[1]
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		clearField137: func(update *PlayerUpdateData) {
			update.Field137 = 0
		},
		respawn: runtime.Respawn,
		exit:    runtime.Exit,
	})
}
