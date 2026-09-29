package server

import "unsafe"

// ClientEquipPlayerNative417AA0 updates the client-side player equipment
// mirror without assuming GAME.EXE's 24-byte equipment record. EquipmentData
// contains native pointers and is 48 bytes on 64-bit hosts.
func ClientEquipPlayerNative417AA0(player *Player, weapon bool, itemType uint32, modifiers [4]unsafe.Pointer) *Player {
	if player == nil {
		return nil
	}
	if weapon {
		// GAME.EXE updates the mask before looking for a free record.
		player.WeaponEquip |= itemType
		for i := range player.Weapon {
			if player.Weapon[i].Field0 != 0 {
				continue
			}
			player.Weapon[i].Field0 = itemType
			player.Weapon[i].Field4 = modifiers
			break
		}
		return player
	}
	player.ArmorEquip |= itemType
	for i := range player.Armor {
		if player.Armor[i].Field0 != 0 {
			continue
		}
		player.Armor[i].Field0 = itemType
		player.Armor[i].Field4 = modifiers
		break
	}
	return player
}

// ClientDequipPlayerNative417B80 removes the first matching equipment record.
// Modifier pointers and the trailing word intentionally remain unchanged, as
// in GAME.EXE; a later equip overwrites the pointers when reusing the slot.
func ClientDequipPlayerNative417B80(player *Player, weapon bool, itemType uint32) *Player {
	if player == nil {
		return nil
	}
	if weapon {
		player.WeaponEquip &^= itemType
		for i := range player.Weapon {
			if player.Weapon[i].Field0 == itemType {
				player.Weapon[i].Field0 = 0
				break
			}
		}
		return player
	}
	player.ArmorEquip &^= itemType
	for i := range player.Armor {
		if player.Armor[i].Field0 == itemType {
			player.Armor[i].Field0 = 0
			break
		}
	}
	return player
}

// ClientEquipNPCNative49A3D0 updates the native-width NPC equipment mirror.
// Unlike the player helper, the original NPC path changes the mask only after
// it finds a free slot.
func ClientEquipNPCNative49A3D0(npc *NPC, weapon bool, itemType uint32, modifiers [4]unsafe.Pointer) *NPC {
	if npc == nil {
		return nil
	}
	if weapon {
		for i := range npc.Weapon {
			if npc.Weapon[i].Field0 != 0 {
				continue
			}
			npc.Weapon[i].Field0 = itemType
			npc.Weapon[i].Field4 = modifiers
			npc.WeaponEquip |= itemType
			return npc
		}
		return npc
	}
	for i := range npc.Armor {
		if npc.Armor[i].Field0 != 0 {
			continue
		}
		npc.Armor[i].Field0 = itemType
		npc.Armor[i].Field4 = modifiers
		npc.ArmorEquip |= itemType
		return npc
	}
	return npc
}
