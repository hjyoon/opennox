package opennox

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/server"
)

const (
	equipmentMundanePacketSize48EA70    = 7
	equipmentModifiablePacketSize48EA70 = 11
)

type equipmentPacketHooks48EA70 struct {
	connected    func() bool
	playerByID   func(int) *server.Player
	npcByID      func(int) *server.NPC
	modifierByID func(byte) *server.ModifierEff
}

func equipmentPacketSize48EA70(op netmsg.Op) (int, bool) {
	switch op {
	case netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP,
		netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_ARMOR_DEQUIP,
		netmsg.MSG_REPORT_WEAPON_DEQUIP:
		return equipmentMundanePacketSize48EA70, true
	case netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_MODIFIABLE_ARMOR_EQUIP:
		return equipmentModifiablePacketSize48EA70, true
	default:
		return 0, false
	}
}

func equipmentWeaponOpcode48EA70(op netmsg.Op) bool {
	return op == netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP ||
		op == netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP ||
		op == netmsg.MSG_REPORT_WEAPON_DEQUIP
}

// handleEquipmentNative48EA70 preserves cdecode.c's equipment packet
// semantics while keeping native pointer fields out of the PE32 C decoder.
func handleEquipmentNative48EA70(op netmsg.Op, data []byte, hooks equipmentPacketHooks48EA70) int {
	size, ok := equipmentPacketSize48EA70(op)
	if !ok || len(data) < size {
		return -1
	}
	if hooks.connected == nil || !hooks.connected() {
		return size
	}

	rawCode := binary.LittleEndian.Uint16(data[1:3])
	itemType := binary.LittleEndian.Uint32(data[3:7])
	weapon := equipmentWeaponOpcode48EA70(op)
	if op == netmsg.MSG_REPORT_ARMOR_DEQUIP || op == netmsg.MSG_REPORT_WEAPON_DEQUIP {
		// The server emits dequip reports only for players and does not set the
		// holder-selection high bit on this packet family.
		if hooks.playerByID != nil {
			server.ClientDequipPlayerNative417B80(hooks.playerByID(int(rawCode)), weapon, itemType)
		}
		return size
	}

	modifierIDs := [4]byte{0xff, 0xff, 0xff, 0xff}
	if size == equipmentModifiablePacketSize48EA70 {
		copy(modifierIDs[:], data[7:11])
	}
	var modifiers [4]unsafe.Pointer
	if hooks.modifierByID != nil {
		for i, id := range modifierIDs {
			if modifier := hooks.modifierByID(id); modifier != nil {
				modifiers[i] = modifier.C()
			}
		}
	}

	id := int(rawCode & 0x7fff)
	if rawCode&0x8000 != 0 {
		if hooks.playerByID != nil {
			server.ClientEquipPlayerNative417AA0(hooks.playerByID(id), weapon, itemType, modifiers)
		}
	} else if hooks.npcByID != nil {
		server.ClientEquipNPCNative49A3D0(hooks.npcByID(id), weapon, itemType, modifiers)
	}
	return size
}

func (c *Client) handleEquipmentPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleEquipmentNative48EA70(op, data, equipmentPacketHooks48EA70{
		connected:  nox_client_isConnected,
		playerByID: c.srv.Players.ByID,
		npcByID:    c.srv.NPCs.ByID,
		modifierByID: func(id byte) *server.ModifierEff {
			return c.srv.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
	})
}
