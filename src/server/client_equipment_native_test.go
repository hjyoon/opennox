package server

import (
	"testing"
	"unsafe"
)

func requireNativeEquipmentAddress(t *testing.T, ptr unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ptr)
	}
}

func TestClientEquipPlayerNative417AA0PreservesNativePointers(t *testing.T) {
	player := new(Player)
	requireNativeEquipmentAddress(t, unsafe.Pointer(player))
	player.ArmorEquip = 0x4000
	player.Weapon[0].Field0 = 0x20
	player.Weapon[0].Field20 = 77
	modifiers := [4]unsafe.Pointer{
		unsafe.Pointer(new(byte)),
		nil,
		unsafe.Pointer(new(byte)),
		unsafe.Pointer(new(byte)),
	}

	if got := ClientEquipPlayerNative417AA0(player, true, 0x20000, modifiers); got != player {
		t.Fatalf("result = %p, want %p", got, player)
	}
	if player.WeaponEquip != 0x20000 {
		t.Fatalf("weapon mask = %#x, want 0x20000", player.WeaponEquip)
	}
	if player.Weapon[0].Field0 != 0x20 || player.Weapon[0].Field20 != 77 {
		t.Fatalf("occupied slot changed: %#v", player.Weapon[0])
	}
	if player.Weapon[1].Field0 != 0x20000 || player.Weapon[1].Field4 != modifiers {
		t.Fatalf("new weapon slot = %#v", player.Weapon[1])
	}
	if player.ArmorEquip != 0x4000 {
		t.Fatalf("adjacent armor mask = %#x, want 0x4000", player.ArmorEquip)
	}
}

func TestClientEquipPlayerNative417AA0FullInventoryStillChangesMask(t *testing.T) {
	player := &Player{WeaponEquip: 0x20, ArmorEquip: 0x40}
	for i := range player.Weapon {
		player.Weapon[i].Field0 = uint32(i + 1)
	}
	for i := range player.Armor {
		player.Armor[i].Field0 = uint32(i + 1)
	}

	ClientEquipPlayerNative417AA0(player, true, 0x10000, [4]unsafe.Pointer{})
	ClientEquipPlayerNative417AA0(player, false, 0x20000, [4]unsafe.Pointer{})
	if player.WeaponEquip != 0x10020 || player.ArmorEquip != 0x20040 {
		t.Fatalf("full masks = weapon:%#x armor:%#x", player.WeaponEquip, player.ArmorEquip)
	}
	for i := range player.Weapon {
		if player.Weapon[i].Field0 != uint32(i+1) {
			t.Fatalf("weapon slot %d changed to %#x", i, player.Weapon[i].Field0)
		}
	}
	for i := range player.Armor {
		if player.Armor[i].Field0 != uint32(i+1) {
			t.Fatalf("armor slot %d changed to %#x", i, player.Armor[i].Field0)
		}
	}
}

func TestClientDequipPlayerNative417B80ClearsFirstMatchOnly(t *testing.T) {
	player := &Player{WeaponEquip: 0x120, ArmorEquip: 0x240}
	weaponModifiers := [4]unsafe.Pointer{unsafe.Pointer(new(byte)), nil, nil, nil}
	armorModifiers := [4]unsafe.Pointer{nil, unsafe.Pointer(new(byte)), nil, nil}
	player.Weapon[0] = EquipmentData{Field0: 0x100, Field4: weaponModifiers, Field20: 7}
	player.Weapon[1] = EquipmentData{Field0: 0x100, Field20: 8}
	player.Armor[0] = EquipmentData{Field0: 0x200, Field4: armorModifiers, Field20: 9}
	player.Armor[1] = EquipmentData{Field0: 0x200, Field20: 10}

	ClientDequipPlayerNative417B80(player, true, 0x100)
	ClientDequipPlayerNative417B80(player, false, 0x200)
	if player.WeaponEquip != 0x20 || player.ArmorEquip != 0x40 {
		t.Fatalf("masks = weapon:%#x armor:%#x", player.WeaponEquip, player.ArmorEquip)
	}
	if player.Weapon[0].Field0 != 0 || player.Weapon[0].Field4 != weaponModifiers || player.Weapon[0].Field20 != 7 {
		t.Fatalf("first weapon slot = %#v", player.Weapon[0])
	}
	if player.Weapon[1].Field0 != 0x100 {
		t.Fatalf("second weapon slot changed: %#v", player.Weapon[1])
	}
	if player.Armor[0].Field0 != 0 || player.Armor[0].Field4 != armorModifiers || player.Armor[0].Field20 != 9 {
		t.Fatalf("first armor slot = %#v", player.Armor[0])
	}
	if player.Armor[1].Field0 != 0x200 {
		t.Fatalf("second armor slot changed: %#v", player.Armor[1])
	}
}

func TestClientEquipmentNativeNilHolders(t *testing.T) {
	if ClientEquipPlayerNative417AA0(nil, true, 1, [4]unsafe.Pointer{}) != nil {
		t.Fatal("nil player equip returned non-nil")
	}
	if ClientDequipPlayerNative417B80(nil, true, 1) != nil {
		t.Fatal("nil player dequip returned non-nil")
	}
	if ClientEquipNPCNative49A3D0(nil, true, 1, [4]unsafe.Pointer{}) != nil {
		t.Fatal("nil NPC equip returned non-nil")
	}
}
