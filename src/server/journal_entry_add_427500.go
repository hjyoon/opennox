package server

import (
	"encoding/binary"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// JournalEntryAdd427490 restores GAME.EXE 00427490 with native-width Player
// and PlayerJournal pointers. Entries are allocated from the shared C heap so
// legacy drawing code may traverse them, while alloc keeps ownership paired
// with the native remove paths.
func JournalEntryAdd427490(player *Player, message string, entryType uint16) *PlayerJournal {
	if player == nil {
		return nil
	}
	entry, _ := alloc.New(PlayerJournal{})
	message = journalCString427590(message)
	copy(entry.EntryBuf[:63], message)
	entry.EntryBuf[63] = 0
	entry.Field3 = entryType
	entry.Next = player.Journal
	if player.Journal != nil {
		player.Journal.Prev = entry
	}
	player.Journal = entry
	return entry
}

// JournalEntryAdd427500 restores the server-side part of GAME.EXE 00427500
// and its 00427490 allocator using native-width PlayerUpdateData, Player, and
// PlayerJournal pointers. The local client rebuild in the original only
// recalculates an open journal panel's text height; the linked list remains the
// authoritative journal state and is updated here on every platform.
func (s *Server) JournalEntryAdd427500(unit *Object, message string, entryType uint16) *PlayerJournal {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassPlayer) {
		return nil
	}
	player := unit.UpdateDataPlayer().Player
	if player == nil {
		return nil
	}

	entry := JournalEntryAdd427490(player, message, entryType)

	if player.PlayerInd != HostPlayerIndex && s != nil && s.NetSendPacketXxx != nil {
		var packet [68]byte
		packet[0] = 0xd5 // MSG_JOURNAL
		packet[1] = 1    // add
		copy(packet[2:66], entry.EntryBuf[:])
		binary.LittleEndian.PutUint16(packet[66:68], entry.Field3)
		s.NetSendPacketXxx0(int(player.PlayerInd), packet[:], nil, 1)
	}
	return entry
}
