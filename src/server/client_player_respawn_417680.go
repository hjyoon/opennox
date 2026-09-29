package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const clientPlayerRespawnNoModifier417680 = uint8(0xff)

type clientPlayerRespawnDeps417680 struct {
	gameFlag     func(uint32) bool
	modifierID   func(string) int
	modifierDesc func(int) *ModifierEff
	equip        func(uint8, *Player, uint32, [4]uint8)
}

func clientPlayerRespawnModifierIndex417680(
	deps clientPlayerRespawnDeps417680,
	id int,
) uint8 {
	modifier := deps.modifierDesc(id)
	if modifier == nil {
		return clientPlayerRespawnNoModifier417680
	}
	return uint8(modifier.ind4)
}

// clientPlayerRespawn417680 rebuilds the client-side equipment mirror without
// walking GAME.EXE's packed 24-byte records. EquipmentData contains native
// pointers and grows to 48 bytes on 64-bit hosts.
func clientPlayerRespawn417680(
	player *Player,
	equipmentMask uint8,
	deps clientPlayerRespawnDeps417680,
) {
	if player == nil {
		return
	}
	if !deps.gameFlag(1) {
		player.WeaponEquip = 0
	}
	for i := range player.Weapon {
		player.Weapon[i].Field0 = 0
		player.Weapon[i].Field4 = [4]unsafe.Pointer{}
	}
	if !deps.gameFlag(1) {
		player.ArmorEquip = 0
	}
	for i := range player.Armor {
		player.Armor[i].Field0 = 0
		player.Armor[i].Field4 = [4]unsafe.Pointer{}
	}

	userColor := deps.modifierDesc(deps.modifierID("UserColor1"))
	if userColor == nil {
		return
	}
	colorBase := userColor.ind4
	modifiers := [4]uint8{
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
	}
	if player.PlayerClass() != 0 || deps.gameFlag(2048) {
		colors := &player.Info().Colors
		modifiers[1] = clientPlayerRespawnModifierIndex417680(deps, int(colorBase+uint32(colors.Shirt1)))
		modifiers[2] = clientPlayerRespawnModifierIndex417680(deps, int(colorBase+uint32(colors.Shirt2)))
		if equipmentMask&1 != 0 {
			deps.equip(82, player, 1024, modifiers)
		}
	}
	modifiers = [4]uint8{
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
	}
	modifiers[1] = clientPlayerRespawnModifierIndex417680(
		deps,
		int(colorBase+uint32(player.Info().Colors.Pants)),
	)
	if equipmentMask&2 != 0 {
		deps.equip(82, player, 4, modifiers)
	}
	modifiers = [4]uint8{
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
	}
	colors := &player.Info().Colors
	modifiers[0] = clientPlayerRespawnModifierIndex417680(deps, int(colorBase+uint32(colors.Shoes2)))
	modifiers[1] = clientPlayerRespawnModifierIndex417680(deps, int(colorBase+uint32(colors.Shoes1)))
	if equipmentMask&4 != 0 {
		deps.equip(82, player, 1, modifiers)
	}

	modifiers = [4]uint8{
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
		clientPlayerRespawnNoModifier417680,
	}
	if player.PlayerClass() == 1 {
		if deps.gameFlag(2048) {
			if equipmentMask&8 != 0 {
				modifiers[0] = clientPlayerRespawnModifierIndex417680(
					deps,
					deps.modifierID("ArmorQuality1"),
				)
				deps.equip(80, player, 0x8000, modifiers)
			}
		} else if deps.gameFlag(4096) {
			modifiers[2] = clientPlayerRespawnModifierIndex417680(
				deps,
				deps.modifierID("Replenishment1"),
			)
			deps.equip(80, player, 0x10000, modifiers)
		} else if equipmentMask&0x10 != 0 {
			deps.equip(79, player, 0x4000, modifiers)
		}
	}
	if player.PlayerClass() == 0 {
		if deps.gameFlag(2048) {
			if equipmentMask&0x20 != 0 {
				modifiers[0] = clientPlayerRespawnModifierIndex417680(
					deps,
					deps.modifierID("Material1"),
				)
				deps.equip(80, player, 256, modifiers)
			}
		} else if deps.gameFlag(4096) {
			deps.equip(80, player, 256, modifiers)
		} else {
			if equipmentMask&0x40 != 0 {
				deps.equip(80, player, 512, modifiers)
			}
			if equipmentMask&0x80 != 0 {
				deps.equip(79, player, 0x1000000, modifiers)
			}
		}
	}
	if player.PlayerClass() == 2 {
		if deps.gameFlag(2048) {
			if equipmentMask&8 != 0 {
				modifiers[0] = clientPlayerRespawnModifierIndex417680(
					deps,
					deps.modifierID("ArmorQuality1"),
				)
				deps.equip(80, player, 0x8000, modifiers)
			}
		} else if deps.gameFlag(4096) {
			deps.equip(80, player, 4, modifiers)
		}
	}
}

func clientPlayerRespawnServerDeps417680(s *Server) clientPlayerRespawnDeps417680 {
	return clientPlayerRespawnDeps417680{
		gameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		modifierID:   s.Modif.Nox_xxx_modifGetIdByName413290,
		modifierDesc: s.Modif.Nox_xxx_modifGetDescById413330,
		equip: func(opcode uint8, source *Player, itemType uint32, modifierIDs [4]uint8) {
			// 00417AA0 resolves the player again from its network code before
			// updating the mirror. Preserve that failure boundary and order.
			player := s.Players.ByID(source.NetCode())
			if player == nil {
				return
			}
			var modifiers [4]unsafe.Pointer
			for i, id := range modifierIDs {
				if modifier := s.Modif.Nox_xxx_modifGetDescById413330(int(id)); modifier != nil {
					modifiers[i] = modifier.C()
				}
			}
			ClientEquipPlayerNative417AA0(player, opcode == 80 || opcode == 81, itemType, modifiers)
		},
	}
}

// ClientPlayerRespawn417680 binds GAME.EXE 00417680 to native-width Player,
// EquipmentData, and ModifierEff pointers.
func (s *Server) ClientPlayerRespawn417680(player *Player, equipmentMask uint8) {
	clientPlayerRespawn417680(player, equipmentMask, clientPlayerRespawnServerDeps417680(s))
}
