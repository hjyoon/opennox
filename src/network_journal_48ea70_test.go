package opennox

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/server"
)

func journalPacket48EA70(action byte, message string, entryType uint16) []byte {
	data := make([]byte, journalPacketSize48EA70)
	data[0] = byte(netmsg.MSG_JOURNAL_MSG)
	data[1] = action
	copy(data[2:66], message)
	binary.LittleEndian.PutUint16(data[66:68], entryType)
	return data
}

func freeJournalEntries48EA70(t *testing.T, player *server.Player) {
	t.Helper()
	for player.Journal != nil {
		message := string(bytes.TrimRight(player.Journal.EntryBuf[:], "\x00"))
		if _, ok := server.JournalEntryRemove427590(player, message); !ok {
			t.Fatalf("failed to release journal entry %q", message)
		}
	}
}

func TestDecodeJournalPacketState48EA70(t *testing.T) {
	data := journalPacket48EA70(3, "War01A:Quest\x00ignored", 0x1234)
	state, ok := decodeJournalPacketState48EA70(data)
	if !ok || state.Action != 3 || state.Message != "War01A:Quest" || state.EntryType != 0x1234 {
		t.Fatalf("decoded state = %+v/%v", state, ok)
	}

	full := strings.Repeat("x", 64)
	state, ok = decodeJournalPacketState48EA70(journalPacket48EA70(1, full, 8))
	if !ok || state.Message != full || state.EntryType != 8 {
		t.Fatalf("full-width state = %+v/%v", state, ok)
	}

	for size := 0; size < journalPacketSize48EA70; size++ {
		if _, ok := decodeJournalPacketState48EA70(make([]byte, size)); ok {
			t.Fatalf("accepted short journal packet of size %d", size)
		}
	}
}

func TestHandleJournalPacketNative48EA70Gates(t *testing.T) {
	var connectedCalls, localPlayerCalls, mutationCalls, rebuildCalls int
	hooks := journalPacketHooks48EA70{
		connected: func() bool {
			connectedCalls++
			return false
		},
		localPlayer: func() *server.Player {
			localPlayerCalls++
			return new(server.Player)
		},
		add:     func(*server.Player, string, uint16) { mutationCalls++ },
		remove:  func(*server.Player, string) { mutationCalls++ },
		update:  func(*server.Player, string, uint16) { mutationCalls++ },
		rebuild: func() { rebuildCalls++ },
	}

	if got := handleJournalPacketNative48EA70(make([]byte, 67), hooks); got != -1 {
		t.Fatalf("short packet returned %d, want -1", got)
	}
	if got := handleJournalPacketNative48EA70(journalPacket48EA70(0, "bad", 0), hooks); got != -1 {
		t.Fatalf("unknown action returned %d, want -1", got)
	}
	if connectedCalls != 0 {
		t.Fatalf("invalid packets queried connection state %d times", connectedCalls)
	}

	for action := byte(1); action <= 3; action++ {
		if got := handleJournalPacketNative48EA70(journalPacket48EA70(action, "offline", 2), hooks); got != journalPacketSize48EA70 {
			t.Fatalf("offline action %d returned %d", action, got)
		}
	}
	if connectedCalls != 3 || localPlayerCalls != 0 || mutationCalls != 0 || rebuildCalls != 0 {
		t.Fatalf("offline calls = connected:%d player:%d mutation:%d rebuild:%d",
			connectedCalls, localPlayerCalls, mutationCalls, rebuildCalls)
	}
}

func TestHandleJournalPacketNative48EA70MutatesNativeList(t *testing.T) {
	player := new(server.Player)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(player)) <= uintptr(^uint32(0)) {
		t.Fatalf("player pointer %#x did not exercise native-width storage", uintptr(unsafe.Pointer(player)))
	}
	t.Cleanup(func() { freeJournalEntries48EA70(t, player) })
	var rebuilds int
	hooks := journalPacketHooks48EA70{
		connected:   func() bool { return true },
		localPlayer: func() *server.Player { return player },
		add: func(player *server.Player, message string, entryType uint16) {
			server.JournalEntryAdd427490(player, message, entryType)
		},
		remove: func(player *server.Player, message string) {
			server.JournalEntryRemove427590(player, message)
		},
		update: func(player *server.Player, message string, entryType uint16) {
			server.JournalEntryUpdate4276B0(player, message, entryType)
		},
		rebuild: func() { rebuilds++ },
	}

	if got := handleJournalPacketNative48EA70(journalPacket48EA70(1, "first", 2), hooks); got != 68 {
		t.Fatalf("first add returned %d", got)
	}
	first := player.Journal
	if first == nil {
		t.Fatal("first add did not publish an entry")
	}
	if message := string(bytes.TrimRight(first.EntryBuf[:], "\x00")); message != "first" || first.Field3 != 2 || rebuilds != 1 {
		t.Fatalf("first add = %p/%q/%d, rebuilds %d", first, message, first.Field3, rebuilds)
	}

	handleJournalPacketNative48EA70(journalPacket48EA70(1, "second", 8), hooks)
	second := player.Journal
	if second == nil || second.Next != first || first.Prev != second || rebuilds != 2 {
		t.Fatal("second add did not preserve native-width doubly linked list")
	}
	handleJournalPacketNative48EA70(journalPacket48EA70(3, "first", 0x1234), hooks)
	if first.Field3 != 0x1234 || rebuilds != 2 {
		t.Fatalf("update type/rebuilds = %#x/%d", first.Field3, rebuilds)
	}
	handleJournalPacketNative48EA70(journalPacket48EA70(2, "second", 0), hooks)
	if player.Journal != first || first.Prev != nil || first.Next != nil || rebuilds != 3 {
		t.Fatal("remove did not normalize the new native-width list head")
	}
}

func TestHandleJournalPacketNative48EA70NilPlayerRebuildParity(t *testing.T) {
	var mutations, rebuilds int
	hooks := journalPacketHooks48EA70{
		connected:   func() bool { return true },
		localPlayer: func() *server.Player { return nil },
		add:         func(*server.Player, string, uint16) { mutations++ },
		remove:      func(*server.Player, string) { mutations++ },
		update:      func(*server.Player, string, uint16) { mutations++ },
		rebuild:     func() { rebuilds++ },
	}
	for action := byte(1); action <= 3; action++ {
		if got := handleJournalPacketNative48EA70(journalPacket48EA70(action, "entry", 1), hooks); got != 68 {
			t.Fatalf("action %d returned %d", action, got)
		}
	}
	if mutations != 0 || rebuilds != 2 {
		t.Fatalf("nil-player calls = mutations:%d rebuilds:%d, want 0/2", mutations, rebuilds)
	}
}
