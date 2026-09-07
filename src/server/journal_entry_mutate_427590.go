package server

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func journalMessage427590(entry *PlayerJournal) string {
	if entry == nil {
		return ""
	}
	n := bytes.IndexByte(entry.EntryBuf[:], 0)
	if n < 0 {
		n = len(entry.EntryBuf)
	}
	return string(entry.EntryBuf[:n])
}

func journalCString427590(message string) string {
	if n := strings.IndexByte(message, 0); n >= 0 {
		return message[:n]
	}
	return message
}

func journalEntryFind427590(player *Player, message string) *PlayerJournal {
	if player == nil {
		return nil
	}
	message = journalCString427590(message)
	for entry := player.Journal; entry != nil; entry = entry.Next {
		if journalMessage427590(entry) == message {
			return entry
		}
	}
	return nil
}

func journalPlayer427630(unit *Object) *Player {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassPlayer) {
		return nil
	}
	return unit.UpdateDataPlayer().Player
}

// journalEntryRemove427590 restores GAME.EXE 00427590 with native-width list
// links. It removes only the first case-sensitive C-string match.
func journalEntryRemove427590(player *Player, message string) ([64]byte, bool) {
	entry := journalEntryFind427590(player, message)
	if entry == nil {
		return [64]byte{}, false
	}
	entryBuf := entry.EntryBuf
	if entry.Prev != nil {
		entry.Prev.Next = entry.Next
	}
	if entry.Next != nil {
		entry.Next.Prev = entry.Prev
	}
	if player.Journal == entry {
		player.Journal = entry.Next
	}
	entry.Next = nil
	entry.Prev = nil
	alloc.Free(entry)
	return entryBuf, true
}

// JournalEntryRemove427630 restores the server-side GAME.EXE 00427630
// dispatch without narrowing Object, PlayerUpdateData, Player, or journal
// pointers to the original PE32 integer representation.
func (s *Server) JournalEntryRemove427630(unit *Object, message string) bool {
	player := journalPlayer427630(unit)
	if player == nil {
		return false
	}
	entryBuf, ok := journalEntryRemove427590(player, message)
	if !ok {
		return false
	}

	// The original host branch only recalculates an open journal panel's text
	// height. The linked list above is the authoritative local state.
	if player.PlayerInd != HostPlayerIndex && s != nil && s.NetSendPacketXxx != nil {
		var packet [68]byte
		packet[0] = 0xd5 // MSG_JOURNAL
		packet[1] = 2    // remove
		copy(packet[2:66], entryBuf[:])
		s.NetSendPacketXxx0(int(player.PlayerInd), packet[:], nil, 1)
	}
	return true
}

// journalEntryUpdate4276B0 restores GAME.EXE 004276B0 with native-width list
// links. It edits only the first case-sensitive C-string match.
func journalEntryUpdate4276B0(player *Player, message string, entryType uint16) *PlayerJournal {
	entry := journalEntryFind427590(player, message)
	if entry == nil {
		return nil
	}
	entry.Field3 = entryType
	return entry
}

// JournalEntryUpdate427720 restores the server-side GAME.EXE 00427720
// dispatch without narrowing Object, PlayerUpdateData, Player, or journal
// pointers to the original PE32 integer representation.
func (s *Server) JournalEntryUpdate427720(unit *Object, message string, entryType uint16) *PlayerJournal {
	player := journalPlayer427630(unit)
	if player == nil {
		return nil
	}
	entry := journalEntryUpdate4276B0(player, message, entryType)
	if entry == nil {
		return nil
	}
	if player.PlayerInd != HostPlayerIndex && s != nil && s.NetSendPacketXxx != nil {
		var packet [68]byte
		packet[0] = 0xd5 // MSG_JOURNAL
		packet[1] = 3    // update
		copy(packet[2:66], entry.EntryBuf[:])
		binary.LittleEndian.PutUint16(packet[66:68], entry.Field3)
		s.NetSendPacketXxx0(int(player.PlayerInd), packet[:], nil, 1)
	}
	return entry
}
