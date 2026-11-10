package opennox

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestGame4TradePacketNativeAllThirteenWidthsAndTail(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	player := &server.Player{Field3680: 3}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	player.PlayerUnit = unit
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= 0xffffffff {
		t.Fatal("native trade formal owner did not retain its high address")
	}
	for _, tc := range []struct {
		subtype byte
		width   int
	}{
		{0x0e, 2}, {0x0f, 4}, {0x10, 4}, {0x11, 2}, {0x12, 2}, {0x15, 4},
		{0x16, 4}, {0x17, 5}, {0x18, 4}, {0x19, 5}, {0x1a, 4}, {0x1c, 4}, {0x1e, 4},
	} {
		t.Run(fmt.Sprintf("%02x", tc.subtype), func(t *testing.T) {
			packet := make([]byte, tc.width+2)
			packet[0] = 0xc9
			packet[1] = tc.subtype
			packet[tc.width] = 0xa5
			packet[tc.width+1] = 0x5a
			before := append([]byte(nil), packet...)
			if got := s.Game4TradePacketNative(unit, packet); got != tc.width || !bytes.Equal(packet, before) {
				t.Fatalf("consumed=%d packet=%x, want%d unchanged", got, packet, tc.width)
			}
			for n := 1; n < tc.width; n++ {
				if got := s.Game4TradePacketNative(unit, packet[:n]); got != -1 {
					t.Fatalf("short size%d returned%d", n, got)
				}
			}
		})
	}
	if got := s.Game4TradePacketNative(unit, []byte{0xc9, 0xff}); got != -1 {
		t.Fatalf("unknown subtype=%d", got)
	}
	if got := s.Game4TradePacketNative(unit, []byte{0xc8, 0x0e}); got != -1 {
		t.Fatalf("wrong opcode=%d", got)
	}
}
