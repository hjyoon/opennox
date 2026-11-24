package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func objectStatusTestServer(t *testing.T, capacity int) *server.Server {
	t.Helper()
	handles.InitHigh()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	pool := alloc.NewClassHigh("object-status-packet-test", importantPacketSizeC(), capacity)
	setImportantAllocClassC(pool.UPtr())
	gameFrameHook = func() uint32 { return 19 }

	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer := GetServer
	GetServer = func() Server { return &netClientSendTestServer{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	return srv
}

func checkObjectStatusPacket(t *testing.T, address unsafe.Pointer, recipient uint8, data []byte, sequenced uint8, remove uint32) {
	t.Helper()
	if address == nil {
		t.Fatalf("no packet for opcode %#x", data[0])
	}
	packet := (*importantPacketLegacy)(address)
	if got := packet.Payload[:int(packet.PayloadSize)]; !bytes.Equal(got, data) {
		t.Fatalf("payload %x, want %x", got, data)
	}
	if uint8(packet.Recipient) != recipient || packet.SequenceEnabled != sequenced ||
		packet.RemoveIfDisconnected != remove || packet.CreatedFrame != 19 || importantPacketRelatedObjectC(address) != nil {
		t.Fatalf("opcode %#x recipient %d, sequence %d, disconnect policy %d",
			data[0], packet.Recipient, packet.SequenceEnabled, packet.RemoveIfDisconnected)
	}
}

func TestObjectStatusPacketsKeepNativeQueueOptions(t *testing.T) {
	objectStatusTestServer(t, 10)
	setImportantRecipientMaskC(1 << 6)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	*unit = server.Object{NetCode: 0x2345, ObjClass: object.ClassMonster}
	calls := objectStatusPacketNativeCalls(6, unit)
	for i, tc := range []struct {
		payload   []byte
		sequenced uint8
		remove    uint32
	}{
		{[]byte{0xd8, 0x7f, 0x81, 0x35, 0xcd, 0xab}, 1, 1},
		{[]byte{0xd9, 0xfe, 0x12, 0x70}, 1, 1},
		{[]byte{0xdf, 0xef, 0xcd, 0xab, 0x89, 0x7f}, 0, 1},
		{[]byte{0xe0, 0x45, 0x23, 0x80}, 1, 0},
		{[]byte{0xe1, 0x45, 0x23}, 1, 0},
		{[]byte{0x71}, 1, 0},
		{[]byte{0x34, 0x45, 0x23}, 1, 1},
		{[]byte{0x35, 0x45, 0x23}, 1, 1},
		{[]byte{0x36}, 1, 1},
		{[]byte{0xe4, 1, 0}, 1, 1},
	} {
		if result := calls[i](); result != 1 {
			t.Fatalf("opcode %#x returned %d", tc.payload[0], result)
		}
		first, _ := importantNativeListC()
		checkObjectStatusPacket(t, first, 6, tc.payload, tc.sequenced, tc.remove)
	}
}

func TestDestroyObjectNativeKeepsVisibilityAndFriendRecipients(t *testing.T) {
	srv := objectStatusTestServer(t, 3)
	p1, inactive, p2 := srv.Players.NewRaw(101), srv.Players.NewRaw(102), srv.Players.NewRaw(103)
	if p1 == nil || inactive == nil || p2 == nil {
		t.Fatal("unable to allocate native players")
	}
	inactive.Active = 0
	setImportantRecipientMaskC((1 << p1.PlayerInd) | (1 << p2.PlayerInd))
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	*unit = server.Object{
		NetCode: 0x2345, ObjClass: object.ClassMonster, Field5: 0xfeed00c0,
		Field37: (1 << p1.PlayerInd) | (1 << inactive.PlayerInd),
	}

	objectStatusDestroyNativeCall(unit)
	first, _ := importantNativeListC()
	checkObjectStatusPacket(t, first, p2.PlayerInd, []byte{0x35, 0x45, 0x23}, 1, 1)
	second := importantPacketNextC(first)
	checkObjectStatusPacket(t, second, p1.PlayerInd, []byte{0x35, 0x45, 0x23}, 1, 1)
	third := importantPacketNextC(second)
	checkObjectStatusPacket(t, third, p1.PlayerInd, []byte{0x33, 0x45, 0x23}, 0, 1)
	if importantPacketNextC(third) != nil {
		t.Fatal("object destruction queued an extra packet")
	}
}
