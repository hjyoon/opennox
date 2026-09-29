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

func TestOnPacketTradeDoesNotFallbackToLegacyDecoder51BAD0(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)

	player := &server.Player{}
	update := &server.PlayerUpdateData{
		Player:  player,
		Trade70: &server.TradeSession{},
	}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	player.PlayerUnit = unit
	if s.Server.IsTradeSessionNative(update.Trade70) {
		t.Fatal("unregistered trade session unexpectedly has native ownership")
	}

	for _, tc := range []struct {
		name    string
		subtype byte
		size    int
	}{
		{"cancel", 0x0e, 2},
		{"add offer", 0x0f, 4},
		{"remove offer", 0x10, 4},
		{"accept", 0x11, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packet := make([]byte, tc.size)
			packet[0] = byte(netmsg.MSG_TRADE)
			packet[1] = tc.subtype
			if n, valid := s.onPacketOp(0, netmsg.MSG_TRADE, packet, player, unit); n != tc.size || !valid {
				t.Fatalf("dispatch = (%d,%t), want (%d,true)", n, valid, tc.size)
			}
		})
	}

	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update.Trade70)) <= math.MaxUint32) {
		t.Fatalf("trade identities did not exercise high halves: unit=%p update=%p session=%p", unit, update, update.Trade70)
	}
}

func TestOnPacketTradeRejectsMalformedWithoutLegacyDecoder51BAD0(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)

	player := &server.Player{}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	player.PlayerUnit = unit

	for _, packet := range [][]byte{
		{byte(netmsg.MSG_TRADE), 0x0f},
		{byte(netmsg.MSG_TRADE), 0x0f, 0},
		{byte(netmsg.MSG_TRADE), 0x10},
		{byte(netmsg.MSG_TRADE), 0x10, 0},
		{byte(netmsg.MSG_TRADE), 0xff},
	} {
		if n, valid := s.onPacketOp(0, netmsg.MSG_TRADE, packet, player, unit); n != 0 || valid {
			t.Fatalf("packet %x dispatch = (%d,%t), want (0,false)", packet, n, valid)
		}
	}

	unknown := []byte{0xfe}
	if n, valid := s.onPacketOp(0, netmsg.Op(unknown[0]), unknown, player, unit); n != 0 || valid {
		t.Fatalf("unknown opcode dispatch = (%d,%t), want (0,false)", n, valid)
	}
}
