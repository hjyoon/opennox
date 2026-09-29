package opennox

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func networkPacketNetDebug51BAD0() bool {
	return noxflags.HasEngine(noxflags.EngineNetDebug)
}

func networkPacketTestHighBit51BAD0(code uint16) {
	_ = code & 0x8000
}

// onPacketInventoryNative51BAD0 routes the inventory-related branches already
// restored in Go before the legacy C decoder can narrow live pointers into its
// PE32 temporaries. handled distinguishes an unknown opcode from a malformed
// packet; malformed packets must not fall back to the unbounded C decoder.
func (s *Server) onPacketInventoryNative51BAD0(
	op netmsg.Op,
	data []byte,
	unit *server.Object,
	update *server.PlayerUpdateData,
) (n int, handled, valid bool) {
	switch op {
	case netmsg.MSG_TRY_DROP:
		if len(data) < server.NetworkTryDropPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryDropPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryDrop51BAD0(unit, update, packet, server.NetworkTryDropRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			Drop: func(owner, item *server.Object, point *types.Pointf) {
				_ = legacy.ObjectDropBoundedCall4ED810(owner, item, point)
			},
		}))
	case netmsg.MSG_TRY_GET:
		if len(data) < server.NetworkTryGetPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryGetPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryGet51BAD0(unit, update, packet, server.NetworkTryGetRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			GameBlocked: nox_xxx_gameGet_4DB1B0,
			Pickup:      legacy.Nox_server_tryPickup_51BAD0,
			CarryingTooMuch: func(unit *server.Object) {
				s.Server.NetPriMsgToPlayer(unit, "pickup.c:CarryingTooMuch", 0)
			},
		}))
	case netmsg.MSG_TRY_USE:
		if len(data) < server.NetworkTryUsePacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryUsePacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryUse51BAD0(unit, update, packet, server.NetworkTryUseRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			GameBlocked: nox_xxx_gameGet_4DB1B0,
		}))
	case netmsg.MSG_TRY_EQUIP:
		if len(data) < server.NetworkTryEquipPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryEquipPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryEquip51BAD0(unit, update, packet, server.NetworkTryEquipRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			GameBlocked: nox_xxx_gameGet_4DB1B0,
			TryEquip: func(owner, item *server.Object) {
				legacy.Nox_xxx_playerTryEquip_4F2F70(owner, item)
			},
		}))
	case netmsg.MSG_TRY_DEQUIP:
		if len(data) < server.NetworkTryDequipPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryDequipPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryDequip51BAD0(unit, update, packet, server.NetworkTryDequipRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			TryDequip: func(owner, item *server.Object) {
				legacy.Nox_xxx_playerTryDequip_4F2FB0(owner, item)
			},
		}))
	case netmsg.MSG_TRY_COLLIDE:
		if len(data) < server.NetworkTryCollidePacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkTryCollidePacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkTryCollide51BAD0(unit, update, packet, server.NetworkTryCollideRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			CallCollide: func(callback unsafe.Pointer, target, unit *server.Object) {
				server.CallObjectCollide(callback, target, unit, nil)
			},
		}))
	case netmsg.MSG_INFO_BOOK_DATA:
		if len(data) < server.NetworkInfoBookPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkInfoBookPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkInfoBook51BAD0(unit, update, packet, server.NetworkInfoBookRuntime51BAD0{
			NetDebug:    networkPacketNetDebug51BAD0,
			TestHighBit: networkPacketTestHighBit51BAD0,
			Send: func(recipient uint8, response [server.NetworkInfoBookPacketSize51BAD0]byte) {
				s.NetSendPacketXxx0(int(recipient), response[:], nil, 1)
			},
		}))
	case netmsg.MSG_INVENTORY_FAIL:
		if len(data) < server.NetworkInventoryFailPacketSize51BAD0 {
			return 0, true, false
		}
		packet := (*[server.NetworkInventoryFailPacketSize51BAD0]byte)(unsafe.Pointer(&data[0]))
		n = int(s.Server.NetworkInventoryFail51BAD0(unit, packet, server.NetworkInventoryFailRuntime51BAD0{
			Drop: func(owner, item *server.Object, point *types.Pointf) {
				_ = legacy.ObjectDropDispatchCall4ED790(owner, item, point)
			},
			CarryingTooMuch: func(unit *server.Object) {
				s.Server.NetPriMsgToPlayer(unit, "pickup.c:CarryingTooMuch", 0)
			},
		}))
	default:
		return 0, false, false
	}
	if n <= 0 || n > len(data) {
		return 0, true, false
	}
	return n, true, true
}
