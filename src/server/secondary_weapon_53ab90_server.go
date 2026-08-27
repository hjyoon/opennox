package server

import "github.com/opennox/libs/player"

// SecondaryWeaponReport53AB90 binds GAME.EXE 0053AB90 to native Object,
// PlayerUpdateData, Player, and secondary-weapon pointers. Field27 remains the
// exact fixed-width PE32 storage; the server sidecar owns the lossless pointer
// on wider targets.
func (s *Server) SecondaryWeaponReport53AB90(
	owner, item *Object,
	classCanUseItem func(*Object, player.Class) bool,
	checkStrength func(*Object, *Object) bool,
	clearClient func(byte),
) {
	secondaryWeaponReport53AB90(owner, item, secondaryWeaponHooks53AB90[*Object, *PlayerUpdateData, *Player]{
		loadUpdate: func(owner *Object) *PlayerUpdateData {
			return owner.UpdateDataPlayer()
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(playerValue *Player) player.Class {
			return playerValue.Info().PlayerClass()
		},
		classCanUseItem: classCanUseItem,
		checkStrength:   checkStrength,
		loadPlayerIndex: func(playerValue *Player) byte {
			return playerValue.PlayerInd
		},
		clearClient: clearClient,
		store: func(update *PlayerUpdateData, item *Object) {
			storeSecondaryWeaponABI32(update, item)
			if item == nil {
				delete(s.secondaryWeapons53AB90, owner)
				return
			}
			if s.secondaryWeapons53AB90 == nil {
				s.secondaryWeapons53AB90 = make(map[*Object]*Object)
			}
			s.secondaryWeapons53AB90[owner] = item
		},
	})
}

func (s *Server) SecondaryWeapon53AB90(owner *Object) *Object {
	return s.secondaryWeapons53AB90[owner]
}
