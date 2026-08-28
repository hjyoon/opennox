package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var (
	cheatAllowAll = false
)

func nox_xxx_inventoryServPlace_4F36F0(obj *server.Object, it *server.Object, a3 int, a4 int) bool {
	s := noxServer
	return nox_xxx_inventoryServPlaceRaw_4F36F0(
		s,
		obj,
		it,
		int32(a3),
		int32(a4),
	) != 0
}

func nox_xxx_inventoryServPlaceRaw_4F36F0(
	s *Server,
	obj, it *server.Object,
	a3, a4 int32,
) int32 {
	return s.S().InventoryServPlace4F36F0(
		obj,
		it,
		a3,
		a4,
		server.InventoryServPlaceRuntime4F36F0{
			DefaultPickup: pickupDefaultRuntime4F31E0(s),
			RefreshCollide: func(item *server.Object) {
				legacy.Sub_5117F0(item)
			},
			ScriptPickup: func(callback *server.ScriptCallback, owner, item *server.Object) {
				s.noxScript.ScriptCallback(callback, owner, item, server.NoxEventInventoryPlace)
			},
		},
	)
}

func nox_xxx_pickupDefault_4F31E0(obj, item *server.Object, a3, a4 int) bool {
	s := noxServer
	return s.S().PickupDefault4F31E0(
		obj,
		item,
		int32(a3),
		int32(a4),
		pickupDefaultRuntime4F31E0(s),
	) != 0
}

func pickupDefaultRuntime4F31E0(s *Server) server.PickupDefaultRuntime4F31E0 {
	return server.PickupDefaultRuntime4F31E0{
		DeleteWorldObject: s.ObjectDeleteLast,
		InventoryPut: func(owner, item *server.Object, report int32) {
			legacy.Nox_xxx_inventoryPutImpl_4F3070(owner, item, int(report))
		},
	}
}

func nox_xxx_pickupFood_4F3350(obj, item *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupFood4F3350(
		obj,
		item,
		a3,
		a4,
		server.PickupFoodRuntime4F3350{
			DefaultPickup: pickupDefaultRuntime4F31E0(s),
		},
	)
}

func nox_xxx_pickupUse_4F34D0(obj, item *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupUse4F34D0(
		obj,
		item,
		a3,
		a4,
		server.PickupUseRuntime4F34D0{
			DefaultPickup: pickupDefaultRuntime4F31E0(s),
		},
	)
}

func nox_xxx_pickupTrap_4F3510(obj, item *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupTrap4F3510(
		obj,
		item,
		a3,
		a4,
		server.PickupTrapRuntime4F3510{
			DefaultPickup: pickupDefaultRuntime4F31E0(s),
		},
	)
}

func nox_xxx_pickupTreasure_4F3580(obj, item *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupTreasure4F3580(
		obj,
		item,
		a3,
		a4,
		server.PickupTreasureRuntime4F3580{
			DefaultPickup:      pickupDefaultRuntime4F31E0(s),
			TreasureMax:        legacy.Nox_xxx_scavengerTreasureMax_4D1600,
			IncrementElimDeath: s.playerIncrementElimDeath4D8D40,
		},
	)
}

// playerIncrementElimDeath4D8D40 is the native-pointer form of the nested
// service TreasurePickup uses when an unteamed player completes Scavenger.
// The scalar match-state helpers remain legacy-owned, while Player and Team
// iteration and every object/update/player access stay native-width.
func (s *Server) playerIncrementElimDeath4D8D40(obj *server.Object) {
	if uint8(obj.ObjClass)&uint8(object.ClassPlayer) == 0 {
		return
	}
	update := (*server.PlayerUpdateData)(obj.UpdateData)
	player := update.Player
	player.Field2140++
	if !noxflags.HasGame(noxflags.GameModeElimination) {
		return
	}

	if legacy.Sub_40AA00() != 0 && legacy.Sub_40AA20() == 0 {
		for current := s.Players.First(); current != nil; current = s.Players.Next(current) {
			if current.Field3680&1 != 0 {
				s.NeedPlayerStatus4174F0(current, 256)
			}
		}
		legacy.Sub_40AA30(1)
	}
	if noxflags.HasGame(noxflags.GameSuddenDeath) || s.GetFlag3592() || legacy.Sub_40AA00() == 0 {
		return
	}

	if !noxflags.HasGamePlay(noxflags.GameplayFlag4) {
		if legacy.Sub_40A770() >= legacy.Sub_40AA40() {
			return
		}
		s.ServStartCountdown(int(s.Balance.Float("SuddenDeathCountdown")), "Settings.c:SuddenDeathImminent")
		return
	}
	if s.Teams.Count() >= legacy.Sub_40AA40() {
		return
	}
	for team := s.Teams.First(); team != nil; team = s.Teams.Next(team) {
		if legacy.Nox_xxx_countNonEliminatedPlayersInTeam_40A830(team) == 1 {
			s.ServStartCountdown(int(s.Balance.Float("SuddenDeathCountdown")), "Settings.c:SuddenDeathImminent")
			return
		}
	}
}

func nox_objectPickupAudEvent_4F3D50(obj1 *server.Object, obj2 *server.Object, a3, a4 int) bool {
	s := noxServer
	if obj1 == nil || obj2 == nil {
		return false
	}
	if !nox_xxx_pickupDefault_4F31E0(obj1, obj2, a3, a4) {
		return false
	}
	if snd := s.PickupSound(obj2.TypeInd); snd != 0 {
		s.Audio.EventObj(snd, obj1, 0, 0)
	}
	return true
}

func sub_57B370(cl object.Class, sub object.SubClass, typ int) byte {
	s := noxServer
	if cl.HasAny(object.ClassWeapon | object.ClassWand) {
		m := s.Modif.Nox_xxx_getProjectileClassById413250(typ)
		if m == nil {
			return 0
		}
		return m.Classes62
	}
	if cl.Has(object.ClassArmor) {
		m := s.Modif.Nox_xxx_equipClothFindDefByTT413270(typ)
		if m == nil {
			return 0
		}
		return m.Classes62
	}
	if cl.Has(object.ClassFood) {
		return byte(^(uint32(sub) >> 5) | 0xFE)
	}
	return 0xFF
}

func nox_xxx_playerClassCanUseItem_57B3D0(item *server.Object, cl player.Class) bool {
	if cheatAllowAll {
		return true
	}
	return ((byte(1) << cl) & sub_57B370(item.Class(), item.SubClass(), int(item.TypeInd))) != 0
}

func nox_server_tryPickup_51BAD0(unit, item *server.Object) {
	// GAME.EXE 0051BF53 calls inventoryServPlace directly. Do not apply the
	// later GameEx library-notice class filter to the original pickup packet.
	nox_xxx_inventoryServPlace_4F36F0(unit, item, 1, 1)
}

func nox_xxx_pickupPotion_4F37D0(obj *server.Object, potion *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupPotion4F37D0(
		obj,
		potion,
		a3,
		a4,
		server.PickupPotionRuntime4F37D0{
			DefaultPickup: pickupDefaultRuntime4F31E0(s),
			PlayerClassCanUse: func(item *server.Object, class uint8) int32 {
				if nox_xxx_playerClassCanUseItem_57B3D0(item, player.Class(class)) {
					return 1
				}
				return 0
			},
			AdjustHealth: func(owner *server.Object, amount int32) {
				legacy.Nox_xxx_unitAdjustHP_4EE460(owner, int(amount))
			},
			AddMana: func(owner *server.Object, amount int32) {
				legacy.Nox_xxx_playerManaAdd_4EEB80(owner, int(amount))
			},
			DelayedDelete: s.DelayedDelete,
		},
	)
}

func nox_xxx_pickupGold_4F3A60(obj, item *server.Object, a3, a4 int32) int32 {
	s := noxServer
	return s.S().PickupGold4F3A60(
		obj,
		item,
		a3,
		a4,
		server.PickupGoldRuntime4F3A60{
			DefaultPickup:   pickupDefaultRuntime4F31E0(s),
			ProtectGold:     legacy.Nox_xxx_protectGoldDelta_56F920,
			DelayedDelete:   s.DelayedDelete,
			SendLineMessage: legacy.Nox_xxx_pickupGoldSendLine_4F3A60,
		},
	)
}
