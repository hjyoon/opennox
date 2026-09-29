package legacy

/*
#include <stdint.h>
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func clientEquipPlayerNative417AA0(player *server.Player, opcode byte, itemType uint32, modifiers [4]unsafe.Pointer) *server.Player {
	return server.ClientEquipPlayerNative417AA0(player, opcode == 80 || opcode == 81, itemType, modifiers)
}

func clientDequipPlayerNative417B80(player *server.Player, opcode byte, itemType uint32) *server.Player {
	return server.ClientDequipPlayerNative417B80(player, opcode == 84, itemType)
}

//export nox_xxx_clientEquipPlayer_native_417AA0
func nox_xxx_clientEquipPlayer_native_417AA0(opcode C.uint8_t, playerID C.int, itemType C.uint32_t, modifierIDs *C.uint8_t) unsafe.Pointer {
	player := GetServer().S().Players.ByID(int(playerID))
	if player == nil {
		return nil
	}
	var modifiers [4]unsafe.Pointer
	if modifierIDs != nil {
		ids := unsafe.Slice((*byte)(unsafe.Pointer(modifierIDs)), len(modifiers))
		for i, id := range ids {
			if modifier := GetServer().S().Modif.Nox_xxx_modifGetDescById413330(int(id)); modifier != nil {
				modifiers[i] = modifier.C()
			}
		}
	}
	return clientEquipPlayerNative417AA0(player, byte(opcode), uint32(itemType), modifiers).C()
}

//export nox_xxx_clientDequipPlayer_native_417B80
func nox_xxx_clientDequipPlayer_native_417B80(opcode C.uint8_t, playerID C.int, itemType C.uint32_t) unsafe.Pointer {
	player := GetServer().S().Players.ByID(int(playerID))
	if player == nil {
		return nil
	}
	return clientDequipPlayerNative417B80(player, byte(opcode), uint32(itemType)).C()
}
