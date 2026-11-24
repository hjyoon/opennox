package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
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

func TestInventoryJournalPacketsKeepNativeWireAndReturnValues(t *testing.T) {
	objectStatusTestServer(t, 7)
	setImportantRecipientMaskC(1 << 6)

	plain, releasePlain := alloc.NewHigh(server.Object{})
	t.Cleanup(releasePlain)
	*plain = server.Object{NetCode: 0x2345, TypeInd: 0x7056}

	modified, releaseModified := alloc.NewHigh(server.Object{})
	t.Cleanup(releaseModified)
	attrs, releaseAttrs := alloc.NewHigh(server.ModifierInitData{})
	t.Cleanup(releaseAttrs)
	modifier, releaseModifier := alloc.NewHigh(server.ModifierEff{})
	t.Cleanup(releaseModifier)
	(*server.Modifier)(unsafe.Pointer(modifier)).TypeInd = 0x1234
	attrs.Modifiers[2] = modifier
	*modified = server.Object{NetCode: 0x2345, TypeInd: 0x7056, ObjClass: object.Class(0x01000000), InitData: unsafe.Pointer(attrs)}

	journal := [68]byte{0xd5, 1, 66: 0x76, 67: 0x98}
	for i := 2; i < 65; i++ {
		journal[i] = 'A'
	}
	update := journal
	update[1] = 3
	removed := [68]byte{0xd5, 2, 'K', 'e', 'y'}

	calls := inventoryJournalNativeCalls(6, plain, modified)
	for i, tc := range []struct {
		payload  []byte
		sequence uint8
		remove   uint32
		result   int64
	}{
		{[]byte{75, 0x45, 0x23, 0x56, 0x70}, 1, 0, 0},
		{[]byte{76, 0x45, 0x23, 0x56, 0x70, 0xff, 0xff, 0x34, 0xff}, 1, 0, 0},
		{[]byte{77, 0x45, 0x23, 0x56, 0x70}, 1, 0, 1},
		{journal[:], 0, 1, 1},
		{removed[:], 0, 1, 1},
		{update[:], 0, 1, 1},
		{[]byte{0xed, 0x81}, 1, 1, 1},
	} {
		if result := calls[i](); result != tc.result {
			t.Fatalf("opcode %#x returned %d, want %d", tc.payload[0], result, tc.result)
		}
		first, _ := importantNativeListC()
		checkObjectStatusPacket(t, first, 6, tc.payload, tc.sequence, tc.remove)
	}
}

func TestCreatureStatusPacketsKeepHealthOrderingAndMetadata(t *testing.T) {
	objectStatusTestServer(t, 6)
	setImportantRecipientMaskC(1 << 6)
	const acquireFlag = noxflags.GameFlag28
	oldFlags := noxflags.GetGame()
	noxflags.SetGame(acquireFlag)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	creature, freeCreature := alloc.NewHigh(server.Object{})
	t.Cleanup(freeCreature)
	*creature = server.Object{NetCode: 0x2345, TypeInd: 0x7056, ObjClass: object.ClassMonster}
	calls := creatureStatusPacketNativeCalls(6, creature)
	if got := calls[0](); got != 0 {
		t.Fatalf("unit without health data returned %d", got)
	}
	if first, _ := importantNativeListC(); first != nil {
		t.Fatal("unit without health data queued a packet")
	}
	health, freeHealth := alloc.NewHigh(server.HealthData{})
	t.Cleanup(freeHealth)
	*health = server.HealthData{Cur: 0x5123, Field2: 0xcafe, Max: 0x7a99}
	creature.HealthData = health
	healthPacket := []byte{0xdd, 0x45, 0x23, 0x23, 0x51, 0x99, 0x7a}
	for i, tc := range []struct {
		enqueueOrder [][]byte
		sequence     uint8
	}{
		{[][]byte{{108, 0x45, 0x23, 0x56, 0xf0}, healthPacket}, 1},
		{[][]byte{{0xdb, 0x45, 0x23, 0x56, 0x70}, healthPacket}, 1},
		{[][]byte{{0xdc, 0x45, 0x23}}, 1},
		{[][]byte{{104, 0x45, 0x23, 0, 0, 0x2f, 0xc4, 0xf7}}, 0},
	} {
		if got := calls[i+1](); got != 1 {
			t.Fatalf("opcode %#x returned %d", tc.enqueueOrder[0][0], got)
		}
		first, _ := importantNativeListC()
		for j := len(tc.enqueueOrder) - 1; j >= 0; j-- {
			checkObjectStatusPacket(t, first, 6, tc.enqueueOrder[j], tc.sequence, 1)
			first = importantPacketNextC(first)
		}
	}
}
