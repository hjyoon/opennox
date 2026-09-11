package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// AudioEventPacketRuntime501FD0 supplies the client viewport value still
// owned by the root game package.
type AudioEventPacketRuntime501FD0 struct {
	WindowWidth func() int32
}

// SendAudioEvent501FD0 binds the verified 00501FD0 packet path to native
// Object, PlayerUpdateData, Player, and AudioEvent layouts. It returns the
// original message-list enqueue result.
func (s *Server) SendAudioEvent501FD0(
	unit *Object,
	event *AudioEvent,
	percentage int32,
	runtime AudioEventPacketRuntime501FD0,
) bool {
	return sendAudioEventPacket501FD0(unit, event, percentage, audioEventPacketHooks501FD0[
		*Object,
		unsafe.Pointer,
		*Player,
		*AudioEvent,
	]{
		loadUpdate: func(unit *Object) unsafe.Pointer {
			return unit.UpdateData
		},
		loadEventObject: func(event *AudioEvent) *Object {
			return event.Obj
		},
		loadEventPositionX: func(event *AudioEvent) float32 {
			return event.Pos.X
		},
		loadEventSound: func(event *AudioEvent) int32 {
			return int32(event.Sound)
		},
		loadPlayer: func(update unsafe.Pointer) *Player {
			return (*PlayerUpdateData)(update).Player
		},
		loadPlayerPositionX: func(player *Player) float32 {
			return player.Pos3632Vec.X
		},
		floatToInt:      audioEventZoneFloatToInt501C00,
		loadWindowWidth: runtime.WindowWidth,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		enqueue: func(index, kind uint8, packet [4]byte) bool {
			return s.NetList.AddToMsgListCli(ntype.PlayerInd(index), netlist.Kind(kind), packet[:])
		},
	})
}
