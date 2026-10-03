package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Nox_client_playerQuestKeysReceived reads the bytes written by the actual
// F0/22 and F0/23 client decoder. Use the native C field, not PE32 offsets or
// the server's inventory/report markers. It never supplies a received result.
func Nox_client_playerQuestKeysReceived(player *server.Player) [2]byte {
	if player == nil {
		return [2]byte{}
	}
	info := (*nox_playerInfo)(unsafe.Pointer(player))
	value := uint32(info.tail_padding[2])
	return [2]byte{byte(value), byte(value >> 8)}
}
