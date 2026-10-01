package server

import "unsafe"

// QuestRecordDeath4D6130 keeps Object, cached PlayerUpdateData, and both live
// Player loads native-width while preserving the original mixed pointer result.
func QuestRecordDeath4D6130(unit *Object) unsafe.Pointer {
	result := questRecordDeath4D6130(unit, questRecordDeathHooks4D6130[*Object, *PlayerUpdateData, *Player]{
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadDeaths: func(player *Player) uint32 {
			return player.field4660
		},
		storeDeaths: func(player *Player, value uint32) {
			player.field4660 = value
		},
		loadMask: func(player *Player) uint32 {
			return player.field4692
		},
		storeMask: func(player *Player, value uint32) {
			player.field4692 = value
		},
	})
	if result.isPlayer {
		return unsafe.Pointer(result.player)
	}
	return unsafe.Pointer(result.unit)
}
