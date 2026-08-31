package server

import (
	"encoding/binary"
	"unsafe"
)

// NetworkTryCollidePacketSize51BAD0 is the exact MSG_TRY_COLLIDE packet width.
const NetworkTryCollidePacketSize51BAD0 = networkTryCollidePacketSize51BAD0

// NetworkTryCollideRuntime51BAD0 supplies the engine-debug observation and
// the final native-width callback dispatch owned by the legacy boundary.
type NetworkTryCollideRuntime51BAD0 struct {
	NetDebug    func() bool
	TestHighBit func(uint16)
	CallCollide func(unsafe.Pointer, *Object, *Object)
}

// NetworkTryCollide51BAD0 binds MSG_TRY_COLLIDE to native Object,
// PlayerUpdateData, Player, and collide-callback pointers on every pointer
// width. No runtime pointer crosses a fixed-width integer representation.
func (s *Server) NetworkTryCollide51BAD0(
	unit *Object,
	update *PlayerUpdateData,
	packet *[NetworkTryCollidePacketSize51BAD0]byte,
	runtime NetworkTryCollideRuntime51BAD0,
) int32 {
	return networkTryCollide51BAD0(unit, update, networkTryCollideHooks51BAD0[
		*Object,
		*PlayerUpdateData,
		*Player,
		unsafe.Pointer,
	]{
		loadWireCode: func() uint16 {
			return binary.LittleEndian.Uint16(packet[1:3])
		},
		dynamicUnitCode: s.packetDynamicUnitCode578B40,
		netDebug:        runtime.NetDebug,
		testHighBit:     runtime.TestHighBit,
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		loadTradeActive: func(update *PlayerUpdateData) bool {
			return update.Trade70 != nil
		},
		loadDialogActive: func(update *PlayerUpdateData) bool {
			return update.DialogWith != nil
		},
		objectFromNetCode: s.ObjectFromNetCode4ECCB0,
		loadCollide: func(target *Object) unsafe.Pointer {
			return target.Collide
		},
		callCollide: runtime.CallCollide,
	})
}
