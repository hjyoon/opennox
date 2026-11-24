package legacy

/*
#include "GAME1.h"
#include "GAME2_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func clientEquipNPCNative49A3D0(npc *server.NPC, opcode byte, itemType uint32, modifiers [4]unsafe.Pointer) *server.NPC {
	return server.ClientEquipNPCNative49A3D0(npc, opcode == 80 || opcode == 81, itemType, modifiers)
}

//export nox_xxx_clientEquip_49A3D0
func nox_xxx_clientEquip_49A3D0(opcode C.uint8_t, npcID C.int, itemType C.uint32_t, modifierIDs C.nox_equipment_modifier_ids) *C.char {
	npc := GetServer().S().NPCs.ByID(int(npcID))
	if npc == nil {
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
	return (*C.char)(clientEquipNPCNative49A3D0(npc, byte(opcode), uint32(itemType), modifiers).C())
}
