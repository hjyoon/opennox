package server

import (
	"reflect"
	"testing"
	"unsafe"

	noxplayer "github.com/opennox/libs/player"
)

type clientPlayerRespawnEquipCall417680 struct {
	opcode    uint8
	player    *Player
	itemType  uint32
	modifiers [4]uint8
}

func TestClientPlayerRespawn417680NilDoesNothing(t *testing.T) {
	called := false
	clientPlayerRespawn417680(nil, 0xff, clientPlayerRespawnDeps417680{
		gameFlag: func(uint32) bool {
			called = true
			return false
		},
		modifierID: func(string) int {
			called = true
			return 0
		},
		modifierDesc: func(int) *ModifierEff {
			called = true
			return nil
		},
		equip: func(uint8, *Player, uint32, [4]uint8) {
			called = true
		},
	})
	if called {
		t.Fatal("nil player invoked a dependency")
	}
}

func TestClientPlayerRespawn417680ClearsNativeEquipment(t *testing.T) {
	player := &Player{WeaponEquip: 0x1234, ArmorEquip: 0x5678}
	requireNativeEquipmentAddress(t, player.C())
	modifier := unsafe.Pointer(new(byte))
	for i := range player.Weapon {
		player.Weapon[i] = EquipmentData{
			Field0:  uint32(i + 1),
			Field4:  [4]unsafe.Pointer{modifier, modifier, modifier, modifier},
			Field20: uint32(1000 + i),
		}
	}
	for i := range player.Armor {
		player.Armor[i] = EquipmentData{
			Field0:  uint32(i + 1),
			Field4:  [4]unsafe.Pointer{modifier, modifier, modifier, modifier},
			Field20: uint32(2000 + i),
		}
	}
	var gameFlags []uint32
	clientPlayerRespawn417680(player, 0xff, clientPlayerRespawnDeps417680{
		gameFlag: func(flag uint32) bool {
			gameFlags = append(gameFlags, flag)
			return false
		},
		modifierID: func(name string) int {
			if name != "UserColor1" {
				t.Fatalf("modifier name = %q", name)
			}
			return 7
		},
		modifierDesc: func(id int) *ModifierEff {
			if id != 7 {
				t.Fatalf("modifier id = %d", id)
			}
			return nil
		},
		equip: func(uint8, *Player, uint32, [4]uint8) {
			t.Fatal("equipment rebuilt without UserColor1")
		},
	})
	if player.WeaponEquip != 0 || player.ArmorEquip != 0 {
		t.Fatalf("equipment masks = weapon:%#x armor:%#x", player.WeaponEquip, player.ArmorEquip)
	}
	if !reflect.DeepEqual(gameFlags, []uint32{1, 1}) {
		t.Fatalf("game flags = %v, want [1 1]", gameFlags)
	}
	for i, got := range player.Weapon {
		if got.Field0 != 0 || got.Field4 != [4]unsafe.Pointer{} || got.Field20 != uint32(1000+i) {
			t.Fatalf("weapon slot %d = %#v", i, got)
		}
	}
	for i, got := range player.Armor {
		if got.Field0 != 0 || got.Field4 != [4]unsafe.Pointer{} || got.Field20 != uint32(2000+i) {
			t.Fatalf("armor slot %d = %#v", i, got)
		}
	}
}

func TestClientPlayerRespawn417680PreservesMasksForGameFlag1(t *testing.T) {
	player := &Player{WeaponEquip: 0x1234, ArmorEquip: 0x5678}
	clientPlayerRespawn417680(player, 0, clientPlayerRespawnDeps417680{
		gameFlag: func(flag uint32) bool {
			return flag == 1
		},
		modifierID: func(string) int { return 7 },
		modifierDesc: func(int) *ModifierEff {
			return nil
		},
		equip: func(uint8, *Player, uint32, [4]uint8) {
			t.Fatal("unexpected equip")
		},
	})
	if player.WeaponEquip != 0x1234 || player.ArmorEquip != 0x5678 {
		t.Fatalf("equipment masks = weapon:%#x armor:%#x", player.WeaponEquip, player.ArmorEquip)
	}
}

func TestClientPlayerRespawn417680RebuildsColoredClothing(t *testing.T) {
	player := new(Player)
	player.Info().SetPlayerClass(1)
	player.Info().Colors = PlayerColors{
		Pants:  3,
		Shirt1: 1,
		Shirt2: 2,
		Shoes1: 4,
		Shoes2: 5,
	}
	descriptors := map[int]*ModifierEff{
		10: {ind4: 40},
		41: {ind4: 141},
		42: {ind4: 142},
		43: {ind4: 143},
		44: {ind4: 144},
		45: {ind4: 145},
	}
	var calls []clientPlayerRespawnEquipCall417680
	clientPlayerRespawn417680(player, 7, clientPlayerRespawnDeps417680{
		gameFlag: func(uint32) bool { return false },
		modifierID: func(name string) int {
			if name != "UserColor1" {
				t.Fatalf("modifier name = %q", name)
			}
			return 10
		},
		modifierDesc: func(id int) *ModifierEff {
			return descriptors[id]
		},
		equip: func(opcode uint8, gotPlayer *Player, itemType uint32, modifiers [4]uint8) {
			calls = append(calls, clientPlayerRespawnEquipCall417680{opcode, gotPlayer, itemType, modifiers})
		},
	})
	want := []clientPlayerRespawnEquipCall417680{
		{82, player, 1024, [4]uint8{0xff, 141, 142, 0xff}},
		{82, player, 4, [4]uint8{0xff, 143, 0xff, 0xff}},
		{82, player, 1, [4]uint8{145, 144, 0xff, 0xff}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("equip calls = %#v, want %#v", calls, want)
	}
}

func TestClientPlayerRespawn417680ClassAndModeEquipment(t *testing.T) {
	tests := []struct {
		name      string
		class     byte
		flags     uint32
		mask      uint8
		wantCalls []clientPlayerRespawnEquipCall417680
	}{
		{
			name:      "wizard normal",
			class:     1,
			mask:      0x10,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 79, itemType: 0x4000, modifiers: [4]uint8{0xff, 0xff, 0xff, 0xff}}},
		},
		{
			name:      "wizard mode 2048",
			class:     1,
			flags:     2048,
			mask:      8,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 0x8000, modifiers: [4]uint8{30, 0xff, 0xff, 0xff}}},
		},
		{
			name:      "wizard mode 4096",
			class:     1,
			flags:     4096,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 0x10000, modifiers: [4]uint8{0xff, 0xff, 32, 0xff}}},
		},
		{
			name:  "mode 2048 takes precedence",
			class: 1,
			flags: 2048 | 4096,
		},
		{
			name:      "warrior mode 2048",
			class:     0,
			flags:     2048,
			mask:      0x20,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 256, modifiers: [4]uint8{31, 0xff, 0xff, 0xff}}},
		},
		{
			name:      "warrior mode 4096",
			class:     0,
			flags:     4096,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 256, modifiers: [4]uint8{0xff, 0xff, 0xff, 0xff}}},
		},
		{
			name:  "warrior normal",
			class: 0,
			mask:  0xc0,
			wantCalls: []clientPlayerRespawnEquipCall417680{
				{opcode: 80, itemType: 512, modifiers: [4]uint8{0xff, 0xff, 0xff, 0xff}},
				{opcode: 79, itemType: 0x1000000, modifiers: [4]uint8{0xff, 0xff, 0xff, 0xff}},
			},
		},
		{
			name:      "conjurer mode 2048",
			class:     2,
			flags:     2048,
			mask:      8,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 0x8000, modifiers: [4]uint8{30, 0xff, 0xff, 0xff}}},
		},
		{
			name:      "conjurer mode 4096",
			class:     2,
			flags:     4096,
			wantCalls: []clientPlayerRespawnEquipCall417680{{opcode: 80, itemType: 4, modifiers: [4]uint8{0xff, 0xff, 0xff, 0xff}}},
		},
		{name: "conjurer normal", class: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := new(Player)
			player.Info().SetPlayerClass(noxplayer.Class(tc.class))
			descriptors := map[int]*ModifierEff{
				10: {ind4: 20},
				20: {ind4: 20},
				30: {ind4: 30},
				31: {ind4: 31},
				32: {ind4: 32},
			}
			ids := map[string]int{
				"UserColor1":     10,
				"ArmorQuality1":  30,
				"Material1":      31,
				"Replenishment1": 32,
			}
			var calls []clientPlayerRespawnEquipCall417680
			clientPlayerRespawn417680(player, tc.mask, clientPlayerRespawnDeps417680{
				gameFlag: func(flag uint32) bool {
					return tc.flags&flag != 0
				},
				modifierID: func(name string) int {
					return ids[name]
				},
				modifierDesc: func(id int) *ModifierEff {
					return descriptors[id]
				},
				equip: func(opcode uint8, gotPlayer *Player, itemType uint32, modifiers [4]uint8) {
					calls = append(calls, clientPlayerRespawnEquipCall417680{opcode, gotPlayer, itemType, modifiers})
				},
			})
			for i := range tc.wantCalls {
				tc.wantCalls[i].player = player
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("equip calls = %#v, want %#v", calls, tc.wantCalls)
			}
		})
	}
}

func TestClientPlayerRespawnServerDeps417680ResolvesNativePointers(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, 1)
	target := &s.Players.list[0]
	target.Active = 1
	target.NetCodeVal = 42
	requireNativeEquipmentAddress(t, target.C())
	source := &Player{NetCodeVal: 42}
	modifier := &ModifierEff{ind4: 7}
	s.Modif.types[0] = modifier

	deps := clientPlayerRespawnServerDeps417680(s)
	deps.equip(80, source, 0x100, [4]uint8{7, 0xff, 0xff, 0xff})
	if target.WeaponEquip != 0x100 || target.Weapon[0].Field0 != 0x100 {
		t.Fatalf("weapon state = mask:%#x slot:%#v", target.WeaponEquip, target.Weapon[0])
	}
	if target.Weapon[0].Field4[0] != modifier.C() {
		t.Fatalf("modifier pointer = %p, want %p", target.Weapon[0].Field4[0], modifier.C())
	}
	if source.WeaponEquip != 0 {
		t.Fatalf("source player was updated: %#x", source.WeaponEquip)
	}

	deps.equip(79, source, 0x200, [4]uint8{0xff, 0xff, 0xff, 0xff})
	if target.ArmorEquip != 0x200 || target.Armor[0].Field0 != 0x200 {
		t.Fatalf("armor state = mask:%#x slot:%#v", target.ArmorEquip, target.Armor[0])
	}
}
