package server

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func journalTestPlayer427500(t *testing.T, index byte) (*Object, *Player) {
	t.Helper()
	unit, freeUnit := alloc.New(Object{})
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	player, freePlayer := alloc.New(Player{})
	t.Cleanup(freePlayer)
	t.Cleanup(freeUpdate)
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	player.PlayerUnit = unit
	player.PlayerInd = index
	return unit, player
}

func TestJournalEntryAdd427500NativeLinksAndTruncates(t *testing.T) {
	unit, player := journalTestPlayer427500(t, HostPlayerIndex)
	first := new(Server).JournalEntryAdd427500(unit, "first", 2)
	if first == nil || player.Journal != first {
		t.Fatal("first journal entry was not linked")
	}
	if got := string(bytes.TrimRight(first.EntryBuf[:], "\x00")); got != "first" || first.Field3 != 2 {
		t.Fatalf("first entry = %q/%d, want first/2", got, first.Field3)
	}

	message := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-overflow"
	second := new(Server).JournalEntryAdd427500(unit, message, 8)
	if player.Journal != second || second.Next != first || first.Prev != second {
		t.Fatal("native-width journal links do not match the original head insertion")
	}
	if got := string(second.EntryBuf[:63]); got != message[:63] || second.EntryBuf[63] != 0 {
		t.Fatalf("truncated entry = %q/%d", got, second.EntryBuf[63])
	}
}

func TestJournalEntryAdd427500RemotePacket(t *testing.T) {
	unit, _ := journalTestPlayer427500(t, 7)
	var recipient, remove, sequence int
	var packet []byte
	s := &Server{NetSendPacketXxx: func(gotRecipient int, payload []byte, _ *Object, gotRemove, gotSequence int) int {
		recipient, remove, sequence = gotRecipient, gotRemove, gotSequence
		packet = append([]byte(nil), payload...)
		return 1
	}}
	entry := s.JournalEntryAdd427500(unit, "War01AFirstQuest", 4)
	if entry == nil {
		t.Fatal("entry allocation failed")
	}
	if recipient != 7 || remove != 1 || sequence != 0 {
		t.Fatalf("send args = %d/%d/%d, want 7/1/0", recipient, remove, sequence)
	}
	if len(packet) != 68 || packet[0] != 0xd5 || packet[1] != 1 ||
		string(bytes.TrimRight(packet[2:66], "\x00")) != "War01AFirstQuest" ||
		packet[66] != 4 || packet[67] != 0 {
		t.Fatalf("journal packet = %v", packet)
	}
}

func TestJournalEntryAdd427500HostDoesNotSend(t *testing.T) {
	unit, _ := journalTestPlayer427500(t, HostPlayerIndex)
	s := &Server{NetSendPacketXxx: func(int, []byte, *Object, int, int) int {
		t.Fatal("host journal entry sent a network packet")
		return 0
	}}
	if s.JournalEntryAdd427500(unit, "local", 2) == nil {
		t.Fatal("host journal entry was not added")
	}
}
