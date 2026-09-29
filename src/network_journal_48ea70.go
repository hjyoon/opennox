package opennox

import (
	"bytes"
	"encoding/binary"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const journalPacketSize48EA70 = 68

type journalPacketState48EA70 struct {
	Action    byte
	Message   string
	EntryType uint16
}

type journalPacketHooks48EA70 struct {
	connected   func() bool
	localPlayer func() *server.Player
	add         func(*server.Player, string, uint16)
	remove      func(*server.Player, string)
	update      func(*server.Player, string, uint16)
	rebuild     func()
}

func decodeJournalPacketState48EA70(data []byte) (journalPacketState48EA70, bool) {
	if len(data) < journalPacketSize48EA70 {
		return journalPacketState48EA70{}, false
	}
	message := data[2:66]
	if n := bytes.IndexByte(message, 0); n >= 0 {
		message = message[:n]
	}
	return journalPacketState48EA70{
		Action:    data[1],
		Message:   string(message),
		EntryType: binary.LittleEndian.Uint16(data[66:68]),
	}, true
}

func handleJournalPacketNative48EA70(data []byte, hooks journalPacketHooks48EA70) int {
	state, ok := decodeJournalPacketState48EA70(data)
	if !ok {
		return -1
	}
	if state.Action < 1 || state.Action > 3 {
		return -1
	}
	if !hooks.connected() {
		return journalPacketSize48EA70
	}
	player := hooks.localPlayer()
	switch state.Action {
	case 1:
		if player != nil {
			hooks.add(player, state.Message, state.EntryType)
		}
		hooks.rebuild()
	case 2:
		if player != nil {
			hooks.remove(player, state.Message)
		}
		hooks.rebuild()
	case 3:
		if player != nil {
			hooks.update(player, state.Message, state.EntryType)
		}
	}
	return journalPacketSize48EA70
}

func (c *Client) handleJournalPacketNative48EA70(data []byte) int {
	return handleJournalPacketNative48EA70(data, journalPacketHooks48EA70{
		connected:   nox_client_isConnected,
		localPlayer: getCurPlayer,
		add: func(player *server.Player, message string, entryType uint16) {
			server.JournalEntryAdd427490(player, message, entryType)
		},
		remove: func(player *server.Player, message string) {
			server.JournalEntryRemove427590(player, message)
		},
		update: func(player *server.Player, message string, entryType uint16) {
			server.JournalEntryUpdate4276B0(player, message, entryType)
		},
		rebuild: legacy.Nox_xxx_cliBuildJournalString_469BC0,
	})
}
