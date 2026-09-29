package server

import (
	"encoding/binary"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
)

type questPlayerStateFixture4D79C0 struct {
	server  *Server
	units   []Object
	updates []PlayerUpdateData
}

func newQuestPlayerStateFixture4D79C0(count int) *questPlayerStateFixture4D79C0 {
	f := &questPlayerStateFixture4D79C0{
		server:  new(Server),
		units:   make([]Object, count),
		updates: make([]PlayerUpdateData, count),
	}
	f.server.Players.list = make([]Player, count)
	for index := 0; index < count; index++ {
		player := &f.server.Players.list[index]
		*player = Player{Active: 1, PlayerInd: uint8(index), Field4792: 1}
		f.updates[index].Player = player
		f.units[index] = Object{
			ObjClass:   object.ClassPlayer,
			TypeInd:    uint16(0x1200 + index),
			NetCode:    uint32(0x3400 + index),
			UpdateData: unsafe.Pointer(&f.updates[index]),
		}
		player.PlayerUnit = &f.units[index]
	}
	return f
}

func assertQuestPlayerStatePointerAbove32Bits4D79C0(t *testing.T, name string, pointer unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(pointer) <= math.MaxUint32 {
		t.Fatalf("%s pointer = %#x, want address above 32-bit range", name, uintptr(pointer))
	}
}

func TestQuestPlayerStateLayout4D79C0(t *testing.T) {
	wantState, wantMarkers := uintptr(324), uintptr(452)
	wantFlagsA, wantFlagsB := uintptr(484), uintptr(516)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		wantState, wantMarkers = 420, 548
		wantFlagsA, wantFlagsB = 580, 612
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"QuestPlayerState", unsafe.Offsetof(PlayerUpdateData{}.QuestPlayerState), wantState},
		{"RespawnMarkers", unsafe.Offsetof(PlayerUpdateData{}.RespawnMarkers), wantMarkers},
		{"QuestPlayerFlagsA", unsafe.Offsetof(PlayerUpdateData{}.QuestPlayerFlagsA), wantFlagsA},
		{"QuestPlayerFlagsB", unsafe.Offsetof(PlayerUpdateData{}.QuestPlayerFlagsB), wantFlagsB},
		{"QuestPlayerState size", unsafe.Sizeof(PlayerUpdateData{}.QuestPlayerState), 128},
		{"RespawnMarkers size", unsafe.Sizeof(PlayerUpdateData{}.RespawnMarkers), 32},
		{"QuestPlayerFlagsA size", unsafe.Sizeof(PlayerUpdateData{}.QuestPlayerFlagsA), 32},
		{"QuestPlayerFlagsB size", unsafe.Sizeof(PlayerUpdateData{}.QuestPlayerFlagsB), 32},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %d, want %d", test.name, test.got, test.want)
		}
	}
}

func TestQuestPlayerStateRemove4D79C0PreservesPointersReloadsIndexAndOrder(t *testing.T) {
	f := newQuestPlayerStateFixture4D79C0(2)
	unit := &f.units[0]
	cached := &f.updates[0]
	replacement := &PlayerUpdateData{Player: &f.server.Players.list[0]}
	reloadedPlayer := &Player{PlayerInd: 7}

	assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "unit", unsafe.Pointer(unit))
	assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "cached update", unsafe.Pointer(cached))
	assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "replacement update", unsafe.Pointer(replacement))
	assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "reloaded player", unsafe.Pointer(reloadedPlayer))

	for _, update := range []*PlayerUpdateData{replacement, &f.updates[1]} {
		update.QuestPlayerState[3], update.QuestPlayerState[7], update.QuestPlayerState[8] = 0x33, 0x77, 0x88
		update.RespawnMarkers[3], update.RespawnMarkers[7], update.RespawnMarkers[8] = 0x33, 0x77, 0x88
		update.QuestPlayerFlagsA[3], update.QuestPlayerFlagsA[7], update.QuestPlayerFlagsA[8] = 0x33, 0x77, 0x88
		update.QuestPlayerFlagsB[3], update.QuestPlayerFlagsB[7], update.QuestPlayerFlagsB[8] = 0x33, 0x77, 0x88
	}

	var events []string
	got := f.server.QuestPlayerStateRemove4D79C0(unit, QuestPlayerStateRuntime4D79C0{
		Notify: func(recipient int, gotUnit *Object) int {
			if recipient != 255 || gotUnit != unit {
				t.Fatalf("notify = recipient %d unit %p, want 255/%p", recipient, gotUnit, unit)
			}
			events = append(events, "notify")
			unit.UpdateData = unsafe.Pointer(replacement)
			return 0x55
		},
		Reset: func(gotUnit *Object) {
			if gotUnit != unit {
				t.Fatalf("reset unit = %p, want %p", gotUnit, unit)
			}
			events = append(events, "reset")
			cached.Player = reloadedPlayer
		},
	})
	if got != 0 || !reflect.DeepEqual(events, []string{"notify", "reset"}) {
		t.Fatalf("result/events = %d/%q, want 0/[notify reset]", got, events)
	}
	for index, update := range []*PlayerUpdateData{replacement, &f.updates[1]} {
		if update.QuestPlayerState[7] != 0 || update.RespawnMarkers[7] != 0 ||
			update.QuestPlayerFlagsA[7] != 0 || update.QuestPlayerFlagsB[7] != 0 {
			t.Fatalf("update %d slot 7 was not cleared", index)
		}
		if update.QuestPlayerState[3] != 0x33 || update.RespawnMarkers[3] != 0x33 ||
			update.QuestPlayerFlagsA[3] != 0x33 || update.QuestPlayerFlagsB[3] != 0x33 ||
			update.QuestPlayerState[8] != 0x88 || update.RespawnMarkers[8] != 0x88 ||
			update.QuestPlayerFlagsA[8] != 0x88 || update.QuestPlayerFlagsB[8] != 0x88 {
			t.Fatalf("update %d adjacent slots changed", index)
		}
	}
	runtime.KeepAlive(f)
}

func TestQuestPlayerStateExpire4D7A80UsesExactQuestStateStrictDeadlineAndUnsignedWrap(t *testing.T) {
	f := newQuestPlayerStateFixture4D79C0(2)
	f.server.SetFrame(4)
	f.server.SetTickRate(2)
	f.server.Players.list[1].Field4792 = 2

	var timestamps [questPlayerStateSlots4D79C0]uint32
	timestamps[0] = 123
	timestamps[1] = math.MaxUint32 - 55 // elapsed == 60: strict deadline does not expire.
	timestamps[2] = math.MaxUint32 - 56 // elapsed == 61: expires across wrap.
	for updateIndex := range f.updates {
		update := &f.updates[updateIndex]
		for _, index := range []int{1, 2} {
			update.QuestPlayerState[index] = uint32(0x100 + index)
			update.RespawnMarkers[index] = byte(0x20 + index)
			update.QuestPlayerFlagsA[index] = byte(0x40 + index)
			update.QuestPlayerFlagsB[index] = byte(0x60 + index)
		}
	}

	var loads [questPlayerStateSlots4D79C0]int
	got := f.server.QuestPlayerStateExpire4D7A80(QuestPlayerStateRuntime4D79C0{
		LoadTimestamp: func(index int) uint32 {
			loads[index]++
			return timestamps[index]
		},
		StoreTimestamp: func(index int, value uint32) {
			timestamps[index] = value
		},
	})
	if got != questPlayerStateSlots4D79C0 {
		t.Fatalf("result = %d, want %d", got, questPlayerStateSlots4D79C0)
	}
	if loads[0] != 0 || timestamps[0] != 0 {
		t.Fatalf("active exact-one slot loaded/stored = %d/%#x, want 0/0", loads[0], timestamps[0])
	}
	if timestamps[1] != math.MaxUint32-55 {
		t.Fatalf("equal-deadline timestamp = %#x, want unchanged", timestamps[1])
	}
	if timestamps[2] != 0 {
		t.Fatalf("wrapped expired timestamp = %#x, want 0", timestamps[2])
	}
	for index, update := range f.updates {
		if update.QuestPlayerState[1] != 0x101 || update.RespawnMarkers[1] != 0x21 ||
			update.QuestPlayerFlagsA[1] != 0x41 || update.QuestPlayerFlagsB[1] != 0x61 {
			t.Fatalf("update %d equal-deadline slot = %#x/%#x/%#x/%#x", index,
				update.QuestPlayerState[1], update.RespawnMarkers[1],
				update.QuestPlayerFlagsA[1], update.QuestPlayerFlagsB[1])
		}
		if update.QuestPlayerState[2] != 0 || update.RespawnMarkers[2] != 0 ||
			update.QuestPlayerFlagsA[2] != 0 || update.QuestPlayerFlagsB[2] != 0 {
			t.Fatalf("update %d wrapped expired slot was not cleared", index)
		}
	}
	runtime.KeepAlive(f)
}

func TestQuestPlayerStateMarkAndClear4D7A60(t *testing.T) {
	s := new(Server)
	s.SetFrame(0x89abcdef)
	var timestamps [questPlayerStateSlots4D79C0]uint32
	runtimeHooks := QuestPlayerStateRuntime4D79C0{
		StoreTimestamp: func(index int, value uint32) {
			timestamps[index] = value
		},
	}
	if got := s.QuestPlayerStateMark4D7A60(17, runtimeHooks); got != 17 || timestamps[17] != 0x89abcdef {
		t.Fatalf("mark = result %d timestamp %#x", got, timestamps[17])
	}
	for index := range timestamps {
		timestamps[index] = uint32(index + 1)
	}
	if got := s.QuestPlayerStateClear4D7B40(runtimeHooks); got != 0 {
		t.Fatalf("clear result = %d, want 0", got)
	}
	if timestamps != [questPlayerStateSlots4D79C0]uint32{} {
		t.Fatalf("timestamps after clear = %v", timestamps)
	}
}

func TestQuestNotifyPlayer4D9D20PreservesNativePointerAndPacket(t *testing.T) {
	s := new(Server)
	unit := &Object{NetCode: 0x12345678}
	assertQuestPlayerStatePointerAbove32Bits4D79C0(t, "notify unit", unsafe.Pointer(unit))

	var recipient, remove, sequence int
	var packet []byte
	var related *Object
	s.NetSendPacketXxx = func(gotRecipient int, gotPacket []byte, gotRelated *Object, gotRemove, gotSequence int) int {
		recipient, remove, sequence = gotRecipient, gotRemove, gotSequence
		packet = append([]byte(nil), gotPacket...)
		related = gotRelated
		return -77
	}
	if got := s.QuestNotifyPlayer4D9D20(19, unit); got != -77 {
		t.Fatalf("send result = %d, want -77", got)
	}
	if recipient != 19 || !reflect.DeepEqual(packet, []byte{0xf0, 0x01, 0x78, 0x56}) ||
		related != nil || remove != 1 || sequence != 1 {
		t.Fatalf("send = recipient %d packet %v related %p remove %d sequence %d", recipient, packet, related, remove, sequence)
	}
	runtime.KeepAlive(unit)
}

func TestPlayerInterestingReset4D7E50BroadcastsAndResetsState(t *testing.T) {
	f := newQuestPlayerStateFixture4D79C0(2)
	target, update := &f.units[0], &f.updates[0]
	target.TypeInd = 0x4567
	target.NetCode = 0x3456
	update.Field62, update.Field63, update.Field64, update.IsCamping = 1, 2, 3, 1
	f.server.SetFrame(0x89abcdef)

	type sendRecord struct {
		recipient, remove, sequence int
		packet                      []byte
		related                     *Object
	}
	var sends []sendRecord
	f.server.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
		sends = append(sends, sendRecord{recipient, remove, sequence, append([]byte(nil), packet...), related})
		return 0
	}
	f.server.Sub_4D7E50(target)

	if update.Field62 != 0 || update.Field63 != 0 || update.Field64 != 0x89abcdef || update.IsCamping != 0 {
		t.Fatalf("reset state = %d/%d/%#x/%d", update.Field62, update.Field63, update.Field64, update.IsCamping)
	}
	if len(sends) != 2 {
		t.Fatalf("send count = %d, want 2", len(sends))
	}
	wantPacket := make([]byte, 7)
	wantPacket[0] = byte(netmsg.MSG_INTERESTING_ID)
	binary.LittleEndian.PutUint16(wantPacket[1:3], 0x3456)
	binary.LittleEndian.PutUint16(wantPacket[3:5], 0x4567)
	wantPacket[5], wantPacket[6] = 2, 2
	for index, send := range sends {
		if send.recipient != index || !reflect.DeepEqual(send.packet, wantPacket) ||
			send.related != nil || send.remove != 1 || send.sequence != 0 {
			t.Fatalf("send %d = recipient %d packet %v related %p remove %d sequence %d", index, send.recipient, send.packet, send.related, send.remove, send.sequence)
		}
	}
	runtime.KeepAlive(f)
}

func TestPlayersInterestingReset4D7EA0TraversesNativePlayerUnits(t *testing.T) {
	f := newQuestPlayerStateFixture4D79C0(2)
	f.server.SetFrame(77)
	for index := range f.updates {
		f.updates[index].Field62 = uint32(index + 1)
		f.updates[index].Field63 = uint32(index + 2)
		f.updates[index].Field64 = uint32(index + 3)
	}
	f.server.Sub_4D7EA0()
	for index, update := range f.updates {
		if update.Field62 != 0 || update.Field63 != 0 || update.Field64 != 77 || update.IsCamping != 0 {
			t.Fatalf("update %d = %d/%d/%d/%d", index, update.Field62, update.Field63, update.Field64, update.IsCamping)
		}
	}
	runtime.KeepAlive(f)
}
