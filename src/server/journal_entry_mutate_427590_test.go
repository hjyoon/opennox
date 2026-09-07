package server

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func journalMutationTestPlayer427590(t *testing.T, index byte) (*Object, *Player) {
	t.Helper()
	unit, player := journalTestPlayer427500(t, index)
	t.Cleanup(func() {
		for entry := player.Journal; entry != nil; {
			next := entry.Next
			alloc.Free(entry)
			entry = next
		}
		player.Journal = nil
	})
	return unit, player
}

func TestJournalEntryRemove427630NativeWidthAndLinks(t *testing.T) {
	unit, player := journalMutationTestPlayer427590(t, HostPlayerIndex)
	s := new(Server)
	tail := s.JournalEntryAdd427500(unit, "tail", 1)
	middle := s.JournalEntryAdd427500(unit, "middle", 2)
	head := s.JournalEntryAdd427500(unit, "head", 4)
	if tail == nil || middle == nil || head == nil {
		t.Fatal("journal setup failed")
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(middle)) <= uintptr(^uint32(0)) {
		t.Fatalf("journal entry pointer %#x did not exercise the native-width path", uintptr(unsafe.Pointer(middle)))
	}

	if !s.JournalEntryRemove427630(unit, "middle") {
		t.Fatal("middle entry was not removed")
	}
	if player.Journal != head || head.Prev != nil || head.Next != tail || tail.Prev != head || tail.Next != nil {
		t.Fatal("middle removal did not preserve the native doubly linked list")
	}
	if !s.JournalEntryRemove427630(unit, "head") {
		t.Fatal("head entry was not removed")
	}
	if player.Journal != tail || tail.Prev != nil || tail.Next != nil {
		t.Fatal("head removal did not publish and normalize the new head")
	}
	if !s.JournalEntryRemove427630(unit, "tail") || player.Journal != nil {
		t.Fatal("tail removal did not empty the journal")
	}
}

func TestJournalEntryRemove427630FirstExactCStringMatch(t *testing.T) {
	unit, player := journalMutationTestPlayer427590(t, HostPlayerIndex)
	s := new(Server)
	caseEntry := s.JournalEntryAdd427500(unit, "Case", 1)
	older := s.JournalEntryAdd427500(unit, "duplicate", 2)
	newer := s.JournalEntryAdd427500(unit, "duplicate", 4)
	if caseEntry == nil || older == nil || newer == nil {
		t.Fatal("journal setup failed")
	}

	if !s.JournalEntryRemove427630(unit, "duplicate") {
		t.Fatal("duplicate entry was not removed")
	}
	if player.Journal != older || older.Prev != nil || older.Next != caseEntry || caseEntry.Prev != older {
		t.Fatal("removal did not stop at the first matching entry")
	}
	if s.JournalEntryRemove427630(unit, "case") {
		t.Fatal("journal matching must remain case-sensitive")
	}
	if player.Journal != older {
		t.Fatal("a failed match changed the journal head")
	}

	nulEntry := s.JournalEntryAdd427500(unit, "prefix\x00ignored", 8)
	if nulEntry == nil || !s.JournalEntryRemove427630(unit, "prefix\x00different") {
		t.Fatal("journal matching did not preserve C-string termination semantics")
	}
}

func TestJournalEntryUpdate427720FirstExactMatch(t *testing.T) {
	unit, player := journalMutationTestPlayer427590(t, HostPlayerIndex)
	s := new(Server)
	older := s.JournalEntryAdd427500(unit, "quest", 1)
	newer := s.JournalEntryAdd427500(unit, "quest", 2)
	if older == nil || newer == nil {
		t.Fatal("journal setup failed")
	}

	got := s.JournalEntryUpdate427720(unit, "quest", 0x1234)
	if got != newer || newer.Field3 != 0x1234 || older.Field3 != 1 {
		t.Fatalf("first-match update = %p/%#x/%#x, want newer/0x1234/0x1", got, newer.Field3, older.Field3)
	}
	if got := s.JournalEntryUpdate427720(unit, "Quest", 7); got != nil {
		t.Fatalf("case-mismatched update returned %p", got)
	}
	if player.Journal != newer || newer.Next != older || older.Prev != newer {
		t.Fatal("journal update changed list links")
	}
}

func TestJournalEntryMutationRemotePackets(t *testing.T) {
	unit, _ := journalMutationTestPlayer427590(t, 7)
	setup := new(Server)
	entry := setup.JournalEntryAdd427500(unit, "War01AFirstQuest", 1)
	if entry == nil {
		t.Fatal("journal setup failed")
	}

	type sentPacket struct {
		recipient int
		remove    int
		sequence  int
		data      []byte
	}
	var packets []sentPacket
	s := &Server{NetSendPacketXxx: func(recipient int, packet []byte, _ *Object, remove, sequence int) int {
		packets = append(packets, sentPacket{
			recipient: recipient,
			remove:    remove,
			sequence:  sequence,
			data:      append([]byte(nil), packet...),
		})
		return 1
	}}
	if got := s.JournalEntryUpdate427720(unit, "War01AFirstQuest", 0x1234); got != entry {
		t.Fatalf("updated entry = %p, want %p", got, entry)
	}
	if !s.JournalEntryRemove427630(unit, "War01AFirstQuest") {
		t.Fatal("updated entry was not removed")
	}
	if len(packets) != 2 {
		t.Fatalf("sent %d packets, want 2", len(packets))
	}
	for i, packet := range packets {
		if packet.recipient != 7 || packet.remove != 1 || packet.sequence != 0 || len(packet.data) != 68 {
			t.Fatalf("packet %d metadata = %d/%d/%d/%d", i, packet.recipient, packet.remove, packet.sequence, len(packet.data))
		}
		if packet.data[0] != 0xd5 || string(bytes.TrimRight(packet.data[2:66], "\x00")) != "War01AFirstQuest" {
			t.Fatalf("packet %d payload = %v", i, packet.data)
		}
	}
	if packets[0].data[1] != 3 || packets[0].data[66] != 0x34 || packets[0].data[67] != 0x12 {
		t.Fatalf("update packet = %v", packets[0].data)
	}
	if packets[1].data[1] != 2 {
		t.Fatalf("remove packet = %v", packets[1].data)
	}
}

func TestJournalEntryMutationHostDoesNotSend(t *testing.T) {
	unit, _ := journalMutationTestPlayer427590(t, HostPlayerIndex)
	setup := new(Server)
	if setup.JournalEntryAdd427500(unit, "local", 2) == nil {
		t.Fatal("journal setup failed")
	}
	s := &Server{NetSendPacketXxx: func(int, []byte, *Object, int, int) int {
		t.Fatal("host journal mutation sent a network packet")
		return 0
	}}
	if s.JournalEntryUpdate427720(unit, "local", 4) == nil {
		t.Fatal("host journal entry was not updated")
	}
	if !s.JournalEntryRemove427630(unit, "local") {
		t.Fatal("host journal entry was not removed")
	}
}

func TestJournalEntryMutationRejectsInvalidUnits(t *testing.T) {
	s := new(Server)
	if s.JournalEntryRemove427630(nil, "entry") || s.JournalEntryUpdate427720(nil, "entry", 1) != nil {
		t.Fatal("nil unit was accepted")
	}
	nonPlayer, freeNonPlayer := alloc.New(Object{})
	t.Cleanup(freeNonPlayer)
	nonPlayer.ObjClass = object.ClassMonster
	if s.JournalEntryRemove427630(nonPlayer, "entry") || s.JournalEntryUpdate427720(nonPlayer, "entry", 1) != nil {
		t.Fatal("non-player unit was accepted")
	}
	playerUnit, freePlayerUnit := alloc.New(Object{})
	t.Cleanup(freePlayerUnit)
	playerUnit.ObjClass = object.ClassPlayer
	if s.JournalEntryRemove427630(playerUnit, "entry") || s.JournalEntryUpdate427720(playerUnit, "entry", 1) != nil {
		t.Fatal("player without update data was accepted")
	}
}
