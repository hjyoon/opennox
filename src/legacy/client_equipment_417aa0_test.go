package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestClientEquipPlayerNative417AA0UsesOpcodeFamily(t *testing.T) {
	weapon := &server.Player{}
	armor := &server.Player{}
	weaponModifiers := [4]unsafe.Pointer{unsafe.Pointer(new(byte)), nil, nil, nil}
	armorModifiers := [4]unsafe.Pointer{nil, unsafe.Pointer(new(byte)), nil, nil}

	if got := clientEquipPlayerNative417AA0(weapon, 81, 0x100, weaponModifiers); got != weapon {
		t.Fatalf("weapon result = %p, want %p", got, weapon)
	}
	if got := clientEquipPlayerNative417AA0(armor, 82, 0x200, armorModifiers); got != armor {
		t.Fatalf("armor result = %p, want %p", got, armor)
	}
	if weapon.WeaponEquip != 0x100 || weapon.Weapon[0].Field0 != 0x100 || weapon.Weapon[0].Field4 != weaponModifiers {
		t.Fatalf("weapon state = mask:%#x slot:%#v", weapon.WeaponEquip, weapon.Weapon[0])
	}
	if armor.ArmorEquip != 0x200 || armor.Armor[0].Field0 != 0x200 || armor.Armor[0].Field4 != armorModifiers {
		t.Fatalf("armor state = mask:%#x slot:%#v", armor.ArmorEquip, armor.Armor[0])
	}
}

func TestClientDequipPlayerNative417B80UsesOpcodeFamily(t *testing.T) {
	player := &server.Player{WeaponEquip: 0x100, ArmorEquip: 0x200}
	player.Weapon[0].Field0 = 0x100
	player.Armor[0].Field0 = 0x200
	clientDequipPlayerNative417B80(player, 84, 0x100)
	clientDequipPlayerNative417B80(player, 83, 0x200)
	if player.WeaponEquip != 0 || player.Weapon[0].Field0 != 0 {
		t.Fatalf("weapon state = mask:%#x slot:%#v", player.WeaponEquip, player.Weapon[0])
	}
	if player.ArmorEquip != 0 || player.Armor[0].Field0 != 0 {
		t.Fatalf("armor state = mask:%#x slot:%#v", player.ArmorEquip, player.Armor[0])
	}
}

func TestClientEquipmentPlayerNativeNilPlayer(t *testing.T) {
	if clientEquipPlayerNative417AA0(nil, 80, 1, [4]unsafe.Pointer{}) != nil {
		t.Fatal("nil equip returned non-nil")
	}
	if clientDequipPlayerNative417B80(nil, 84, 1) != nil {
		t.Fatal("nil dequip returned non-nil")
	}
}
