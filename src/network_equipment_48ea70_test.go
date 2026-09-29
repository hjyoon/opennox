package opennox

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/server"
)

func equipmentTestPacket48EA70(op netmsg.Op, code uint16, itemType uint32, modifierIDs ...byte) []byte {
	size, ok := equipmentPacketSize48EA70(op)
	if !ok {
		panic("unsupported equipment test opcode")
	}
	packet := make([]byte, size)
	packet[0] = byte(op)
	binary.LittleEndian.PutUint16(packet[1:3], code)
	binary.LittleEndian.PutUint32(packet[3:7], itemType)
	copy(packet[7:], modifierIDs)
	return packet
}

func requireEquipmentHighAddress48EA70(t *testing.T, ptr unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ptr)
	}
}

func TestHandleEquipmentNative48EA70RejectsShortPackets(t *testing.T) {
	ops := []netmsg.Op{
		netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP,
		netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_MODIFIABLE_ARMOR_EQUIP,
		netmsg.MSG_REPORT_ARMOR_DEQUIP,
		netmsg.MSG_REPORT_WEAPON_DEQUIP,
	}
	for _, op := range ops {
		size, _ := equipmentPacketSize48EA70(op)
		for n := 0; n < size; n++ {
			if got := handleEquipmentNative48EA70(op, make([]byte, n), equipmentPacketHooks48EA70{}); got != -1 {
				t.Fatalf("op %d accepted %d-byte packet: %d", op, n, got)
			}
		}
	}
	if got := handleEquipmentNative48EA70(netmsg.MSG_PING, make([]byte, 32), equipmentPacketHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported opcode result = %d, want -1", got)
	}
}

func TestHandleEquipmentNative48EA70DisconnectedConsumesPackets(t *testing.T) {
	ops := []netmsg.Op{
		netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP,
		netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP,
		netmsg.MSG_REPORT_MODIFIABLE_ARMOR_EQUIP,
		netmsg.MSG_REPORT_ARMOR_DEQUIP,
		netmsg.MSG_REPORT_WEAPON_DEQUIP,
	}
	hooks := equipmentPacketHooks48EA70{
		connected: func() bool { return false },
		playerByID: func(int) *server.Player {
			t.Fatal("player lookup while disconnected")
			return nil
		},
		npcByID: func(int) *server.NPC {
			t.Fatal("NPC lookup while disconnected")
			return nil
		},
		modifierByID: func(byte) *server.ModifierEff {
			t.Fatal("modifier lookup while disconnected")
			return nil
		},
	}
	for _, op := range ops {
		packet := equipmentTestPacket48EA70(op, 1, 2, 3, 4, 5, 6)
		size, _ := equipmentPacketSize48EA70(op)
		if got := handleEquipmentNative48EA70(op, packet, hooks); got != size {
			t.Fatalf("op %d consumed %d bytes, want %d", op, got, size)
		}
	}
}

func TestHandleEquipmentNative48EA70RoutesModifiablePlayerWeapon(t *testing.T) {
	player := new(server.Player)
	requireEquipmentHighAddress48EA70(t, unsafe.Pointer(player))
	player.ArmorEquip = 0x400
	modifiers := [4]*server.ModifierEff{new(server.ModifierEff), nil, new(server.ModifierEff), new(server.ModifierEff)}
	packet := equipmentTestPacket48EA70(netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP, 0x9234, 0x20000, 7, 0xff, 9, 10)
	wantPacket := append([]byte(nil), packet...)
	var modifierIDs []byte
	hooks := equipmentPacketHooks48EA70{
		connected: func() bool { return true },
		playerByID: func(id int) *server.Player {
			if id != 0x1234 {
				t.Fatalf("player ID = %#x, want 0x1234", id)
			}
			return player
		},
		npcByID: func(int) *server.NPC {
			t.Fatal("NPC lookup for player equipment")
			return nil
		},
		modifierByID: func(id byte) *server.ModifierEff {
			modifierIDs = append(modifierIDs, id)
			switch id {
			case 7:
				return modifiers[0]
			case 9:
				return modifiers[2]
			case 10:
				return modifiers[3]
			default:
				return nil
			}
		},
	}
	if got := handleEquipmentNative48EA70(netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP, packet, hooks); got != 11 {
		t.Fatalf("consumed bytes = %d, want 11", got)
	}
	wantModifiers := [4]unsafe.Pointer{modifiers[0].C(), nil, modifiers[2].C(), modifiers[3].C()}
	if player.WeaponEquip != 0x20000 || player.Weapon[0].Field0 != 0x20000 || player.Weapon[0].Field4 != wantModifiers {
		t.Fatalf("player weapon state = mask:%#x slot:%#v", player.WeaponEquip, player.Weapon[0])
	}
	if player.ArmorEquip != 0x400 {
		t.Fatalf("adjacent armor mask = %#x, want 0x400", player.ArmorEquip)
	}
	if want := []byte{7, 0xff, 9, 10}; !reflect.DeepEqual(modifierIDs, want) {
		t.Fatalf("modifier IDs = %v, want %v", modifierIDs, want)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %x, want %x", packet, wantPacket)
	}
}

func TestHandleEquipmentNative48EA70RoutesMundaneNPCArmor(t *testing.T) {
	npc := new(server.NPC)
	requireEquipmentHighAddress48EA70(t, unsafe.Pointer(npc))
	npc.WeaponEquip = 0x20
	packet := equipmentTestPacket48EA70(netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP, 0x2345, 0x800000)
	var modifierIDs []byte
	hooks := equipmentPacketHooks48EA70{
		connected: func() bool { return true },
		playerByID: func(int) *server.Player {
			t.Fatal("player lookup for NPC equipment")
			return nil
		},
		npcByID: func(id int) *server.NPC {
			if id != 0x2345 {
				t.Fatalf("NPC ID = %#x, want 0x2345", id)
			}
			return npc
		},
		modifierByID: func(id byte) *server.ModifierEff {
			modifierIDs = append(modifierIDs, id)
			return nil
		},
	}
	if got := handleEquipmentNative48EA70(netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP, packet, hooks); got != 7 {
		t.Fatalf("consumed bytes = %d, want 7", got)
	}
	if npc.ArmorEquip != 0x800000 || npc.Armor[0].Field0 != 0x800000 || npc.Armor[0].Field4 != [4]unsafe.Pointer{} {
		t.Fatalf("NPC armor state = mask:%#x slot:%#v", npc.ArmorEquip, npc.Armor[0])
	}
	if npc.WeaponEquip != 0x20 {
		t.Fatalf("adjacent weapon mask = %#x, want 0x20", npc.WeaponEquip)
	}
	if want := []byte{0xff, 0xff, 0xff, 0xff}; !reflect.DeepEqual(modifierIDs, want) {
		t.Fatalf("mundane modifier IDs = %v, want %v", modifierIDs, want)
	}
}

func TestHandleEquipmentNative48EA70DequipsPlayer(t *testing.T) {
	player := &server.Player{WeaponEquip: 0x120}
	modifier := unsafe.Pointer(new(byte))
	player.Weapon[0] = server.EquipmentData{Field0: 0x100, Field4: [4]unsafe.Pointer{modifier}, Field20: 77}
	packet := equipmentTestPacket48EA70(netmsg.MSG_REPORT_WEAPON_DEQUIP, 0x1234, 0x100)
	hooks := equipmentPacketHooks48EA70{
		connected: func() bool { return true },
		playerByID: func(id int) *server.Player {
			if id != 0x1234 {
				t.Fatalf("player ID = %#x, want 0x1234", id)
			}
			return player
		},
		npcByID: func(int) *server.NPC {
			t.Fatal("NPC lookup during player dequip")
			return nil
		},
		modifierByID: func(byte) *server.ModifierEff {
			t.Fatal("modifier lookup during dequip")
			return nil
		},
	}
	if got := handleEquipmentNative48EA70(netmsg.MSG_REPORT_WEAPON_DEQUIP, packet, hooks); got != 7 {
		t.Fatalf("consumed bytes = %d, want 7", got)
	}
	if player.WeaponEquip != 0x20 || player.Weapon[0].Field0 != 0 || player.Weapon[0].Field4[0] != modifier || player.Weapon[0].Field20 != 77 {
		t.Fatalf("dequipped state = mask:%#x slot:%#v", player.WeaponEquip, player.Weapon[0])
	}
}
