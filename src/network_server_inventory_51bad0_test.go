package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestOnPacketInventoryNative51BAD0ConsumesExactPackets(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })

	player := &server.Player{}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	player.PlayerUnit = unit
	for _, tc := range []struct {
		op   netmsg.Op
		size int
	}{
		{netmsg.MSG_TRY_DROP, server.NetworkTryDropPacketSize51BAD0},
		{netmsg.MSG_TRY_GET, server.NetworkTryGetPacketSize51BAD0},
		{netmsg.MSG_TRY_USE, server.NetworkTryUsePacketSize51BAD0},
		{netmsg.MSG_TRY_EQUIP, server.NetworkTryEquipPacketSize51BAD0},
		{netmsg.MSG_TRY_DEQUIP, server.NetworkTryDequipPacketSize51BAD0},
		{netmsg.MSG_TRY_COLLIDE, server.NetworkTryCollidePacketSize51BAD0},
		{netmsg.MSG_INFO_BOOK_DATA, server.NetworkInfoBookPacketSize51BAD0},
		{netmsg.MSG_INVENTORY_FAIL, server.NetworkInventoryFailPacketSize51BAD0},
	} {
		t.Run(tc.op.String(), func(t *testing.T) {
			packet := make([]byte, tc.size)
			packet[0] = byte(tc.op)
			n, handled, valid := s.onPacketInventoryNative51BAD0(tc.op, packet, unit, update)
			if n != tc.size || !handled || !valid {
				t.Fatalf("dispatch = (%d,%t,%t), want (%d,true,true)", n, handled, valid, tc.size)
			}
			if n, valid := s.onPacketOp(0, tc.op, packet, player, unit); n != tc.size || !valid {
				t.Fatalf("server dispatch = (%d,%t), want (%d,true)", n, valid, tc.size)
			}
		})
	}

	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(player)) <= math.MaxUint32) {
		t.Fatalf("native packet identities did not exercise high halves: unit=%p update=%p player=%p", unit, update, player)
	}
}

func TestOnPacketInventoryNative51BAD0RejectsShortPacketsWithoutFallback(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	update := &server.PlayerUpdateData{Player: &server.Player{}}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}

	for _, tc := range []struct {
		op   netmsg.Op
		size int
	}{
		{netmsg.MSG_TRY_DROP, server.NetworkTryDropPacketSize51BAD0},
		{netmsg.MSG_TRY_GET, server.NetworkTryGetPacketSize51BAD0},
		{netmsg.MSG_TRY_USE, server.NetworkTryUsePacketSize51BAD0},
		{netmsg.MSG_TRY_EQUIP, server.NetworkTryEquipPacketSize51BAD0},
		{netmsg.MSG_TRY_DEQUIP, server.NetworkTryDequipPacketSize51BAD0},
		{netmsg.MSG_TRY_COLLIDE, server.NetworkTryCollidePacketSize51BAD0},
		{netmsg.MSG_INFO_BOOK_DATA, server.NetworkInfoBookPacketSize51BAD0},
		{netmsg.MSG_INVENTORY_FAIL, server.NetworkInventoryFailPacketSize51BAD0},
	} {
		t.Run(tc.op.String(), func(t *testing.T) {
			for size := 0; size < tc.size; size++ {
				packet := make([]byte, size)
				if size != 0 {
					packet[0] = byte(tc.op)
				}
				n, handled, valid := s.onPacketInventoryNative51BAD0(tc.op, packet, unit, update)
				if n != 0 || !handled || valid {
					t.Fatalf("size %d dispatch = (%d,%t,%t), want (0,true,false)", size, n, handled, valid)
				}
			}
		})
	}

	if n, handled, valid := s.onPacketInventoryNative51BAD0(netmsg.MSG_KEEP_ALIVE, []byte{byte(netmsg.MSG_KEEP_ALIVE)}, unit, update); n != 0 || handled || valid {
		t.Fatalf("unrelated dispatch = (%d,%t,%t), want (0,false,false)", n, handled, valid)
	}
}
