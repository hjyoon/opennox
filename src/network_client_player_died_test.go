package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDiedNetCode48EA70(t *testing.T) {
	if _, ok := playerDiedNetCode48EA70([]byte{0xe8, 1}); ok {
		t.Fatal("short packet was accepted")
	}
	if got, ok := playerDiedNetCode48EA70([]byte{0xe8, 0x34, 0x12}); !ok || got != 0x1234 {
		t.Fatalf("net code = %#x/%t, want 0x1234/true", got, ok)
	}
}

func TestClearClientPlayerEquipmentOnDeath48EA70(t *testing.T) {
	ptr := unsafe.Pointer(new(byte))
	player := &server.Player{WeaponEquip: 0xffffffff, ArmorEquip: 0xffffffff}
	for i := range player.Weapon {
		player.Weapon[i].Field0 = uint32(i + 1)
		player.Weapon[i].Field4 = [4]unsafe.Pointer{ptr, ptr, ptr, ptr}
		player.Weapon[i].Field20 = uint32(100 + i)
	}
	player.Armor[0] = server.EquipmentData{
		Field0:  0x1000,
		Field4:  [4]unsafe.Pointer{ptr, ptr, ptr, ptr},
		Field20: 77,
	}
	player.Armor[1] = server.EquipmentData{
		Field0:  0x0004,
		Field4:  [4]unsafe.Pointer{ptr, ptr, ptr, ptr},
		Field20: 88,
	}

	clearClientPlayerEquipmentOnDeath48EA70(player)
	if player.WeaponEquip != 0 {
		t.Fatalf("weapon equip = %#x", player.WeaponEquip)
	}
	for i, item := range player.Weapon {
		if item.Field0 != uint32(i+1) || item.Field4 != [4]unsafe.Pointer{} || item.Field20 != 0 {
			t.Fatalf("weapon[%d] = %#v", i, item)
		}
	}
	if player.ArmorEquip != 0xffffefff {
		t.Fatalf("armor equip = %#x, want 0xffffefff", player.ArmorEquip)
	}
	if player.Armor[0].Field0 != 0 || player.Armor[0].Field4 != [4]unsafe.Pointer{} || player.Armor[0].Field20 != 77 {
		t.Fatalf("ordinary armor = %#v", player.Armor[0])
	}
	if player.Armor[1].Field0 != 0x0004 || player.Armor[1].Field4 != [4]unsafe.Pointer{ptr, ptr, ptr, ptr} ||
		player.Armor[1].Field20 != 88 {
		t.Fatalf("persistent armor = %#v", player.Armor[1])
	}
}
