package opennox

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

const playerDiedPersistentArmorMask48EA70 = uint32(0x0c0d)

// clearClientPlayerEquipmentOnDeath48EA70 preserves MSG_PLAYER_DIED's
// non-Quest equipment cleanup without walking GAME.EXE's 24-byte PE32 records.
// EquipmentData grows to 48 bytes on a 64-bit host because it contains four
// pointers, while its semantic fields remain unchanged.
func clearClientPlayerEquipmentOnDeath48EA70(player *server.Player) {
	player.WeaponEquip = 0
	for i := range player.Weapon {
		player.Weapon[i].Field4 = [4]unsafe.Pointer{}
		player.Weapon[i].Field20 = 0
	}
	for i := range player.Armor {
		mask := player.Armor[i].Field0
		if mask&playerDiedPersistentArmorMask48EA70 != 0 {
			continue
		}
		player.ArmorEquip &^= mask
		player.Armor[i].Field0 = 0
		player.Armor[i].Field4 = [4]unsafe.Pointer{}
	}
}

func playerDiedNetCode48EA70(data []byte) (uint16, bool) {
	if len(data) < 3 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(data[1:3]), true
}
