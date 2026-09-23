package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func highAddressNPC49A3D0(t *testing.T) *server.NPC {
	t.Helper()
	npc := new(server.NPC)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(npc)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", npc)
	}
	return npc
}

func TestClientEquipNPCNative49A3D0WeaponPreservesNativePointers(t *testing.T) {
	npc := highAddressNPC49A3D0(t)
	npc.ArmorEquip = 0x4000
	npc.Field1312 = 0x55667788
	modifiers := [4]unsafe.Pointer{
		unsafe.Pointer(new(byte)),
		unsafe.Pointer(new(byte)),
		unsafe.Pointer(new(byte)),
		unsafe.Pointer(new(byte)),
	}
	if got := clientEquipNPCNative49A3D0(npc, 81, 0x20000, modifiers); got != npc {
		t.Fatalf("result = %p, want %p", got, npc)
	}
	if npc.WeaponEquip != 0x20000 {
		t.Fatalf("weapon mask = %#x, want 0x20000", npc.WeaponEquip)
	}
	if npc.Weapon[0].Field0 != 0x20000 || npc.Weapon[0].Field4 != modifiers {
		t.Fatalf("weapon slot = %#v", npc.Weapon[0])
	}
	if npc.ArmorEquip != 0x4000 || npc.Field1312 != 0x55667788 {
		t.Fatalf("adjacent state changed: armor=%#x field=%#x", npc.ArmorEquip, npc.Field1312)
	}
}

func TestClientEquipNPCNative49A3D0ArmorUsesFirstEmptySlot(t *testing.T) {
	npc := highAddressNPC49A3D0(t)
	npc.ArmorEquip = 0x10
	npc.Armor[0].Field0 = 0x10
	npc.Armor[0].Field20 = 77
	modifiers := [4]unsafe.Pointer{nil, unsafe.Pointer(new(byte)), nil, unsafe.Pointer(new(byte))}

	clientEquipNPCNative49A3D0(npc, 82, 0x800000, modifiers)

	if npc.ArmorEquip != 0x800010 {
		t.Fatalf("armor mask = %#x, want 0x800010", npc.ArmorEquip)
	}
	if npc.Armor[0].Field0 != 0x10 || npc.Armor[0].Field20 != 77 {
		t.Fatalf("occupied armor slot changed: %#v", npc.Armor[0])
	}
	if npc.Armor[1].Field0 != 0x800000 || npc.Armor[1].Field4 != modifiers {
		t.Fatalf("new armor slot = %#v", npc.Armor[1])
	}
}

func TestClientEquipNPCNative49A3D0MundaneOpcodes(t *testing.T) {
	weaponNPC := &server.NPC{}
	armorNPC := &server.NPC{}
	clientEquipNPCNative49A3D0(weaponNPC, 80, 0x400, [4]unsafe.Pointer{})
	clientEquipNPCNative49A3D0(armorNPC, 79, 0x800, [4]unsafe.Pointer{})
	if weaponNPC.WeaponEquip != 0x400 || weaponNPC.Weapon[0].Field0 != 0x400 || weaponNPC.ArmorEquip != 0 {
		t.Fatalf("mundane weapon state = weapon %#x/%#x, armor %#x",
			weaponNPC.WeaponEquip, weaponNPC.Weapon[0].Field0, weaponNPC.ArmorEquip)
	}
	if armorNPC.ArmorEquip != 0x800 || armorNPC.Armor[0].Field0 != 0x800 || armorNPC.WeaponEquip != 0 {
		t.Fatalf("mundane armor state = armor %#x/%#x, weapon %#x",
			armorNPC.ArmorEquip, armorNPC.Armor[0].Field0, armorNPC.WeaponEquip)
	}
}

func TestClientEquipNPCNative49A3D0FullInventoryDoesNotChangeMask(t *testing.T) {
	npc := &server.NPC{WeaponEquip: 0x20, ArmorEquip: 0x40}
	for i := range npc.Weapon {
		npc.Weapon[i].Field0 = uint32(i + 1)
	}
	for i := range npc.Armor {
		npc.Armor[i].Field0 = uint32(i + 1)
	}
	clientEquipNPCNative49A3D0(npc, 80, 0x10000, [4]unsafe.Pointer{})
	clientEquipNPCNative49A3D0(npc, 79, 0x20000, [4]unsafe.Pointer{})
	if npc.WeaponEquip != 0x20 || npc.ArmorEquip != 0x40 {
		t.Fatalf("full masks changed: weapon=%#x armor=%#x", npc.WeaponEquip, npc.ArmorEquip)
	}
}

func TestClientEquipNPCNative49A3D0NilNPC(t *testing.T) {
	if got := clientEquipNPCNative49A3D0(nil, 80, 1, [4]unsafe.Pointer{}); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
}
