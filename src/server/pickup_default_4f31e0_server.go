package server

import (
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
)

// PickupDefaultRuntime4F31E0 supplies the two object-list operations that are
// still owned by the root/legacy runtime. Object pointers remain native-width
// across both callbacks.
type PickupDefaultRuntime4F31E0 struct {
	DeleteWorldObject func(*Object)
	InventoryPut      func(*Object, *Object, int32)
}

type pickupDefaultNativeDeps4F31E0 struct {
	gameFlagsCheck    func(uint32) int32
	findTeam          func(uint8) *Team
	informTeam        func(uint8, uint8, uint32)
	primaryMessage    func(*Object, string, uint8)
	deleteWorldObject func(*Object)
	inventoryPut      func(*Object, *Object, int32)
}

func pickupDefaultNative4F31E0(
	owner, item *Object,
	report, ignored int32,
	deps pickupDefaultNativeDeps4F31E0,
) int32 {
	return pickupDefault4F31E0(
		owner,
		item,
		report,
		ignored,
		pickupDefaultHooks4F31E0[*Object, *Team, *PlayerUpdateData, *Player]{
			gameFlagsCheck: deps.gameFlagsCheck,
			itemHasTeam: func(item *Object) bool {
				return item.TeamVal.Has()
			},
			teamsSame: func(owner, item *Object) bool {
				return owner.TeamVal.SameAs(&item.TeamVal)
			},
			loadTeamID: func(item *Object) uint8 {
				return uint8(item.TeamVal.ID)
			},
			findTeam: deps.findTeam,
			loadClassLow: func(owner *Object) uint8 {
				return uint8(owner.ObjClass)
			},
			loadUpdate: func(owner *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(owner.UpdateData)
			},
			loadTeamColor: func(team *Team) uint8 {
				return uint8(team.ColorInd)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadPlayerInd: func(player *Player) uint8 {
				return player.PlayerInd
			},
			informTeam: deps.informTeam,
			loadInventoryHolder: func(item *Object) *Object {
				return item.InvHolder
			},
			loadCarryCapacity: func(owner *Object) uint16 {
				return owner.CarryCapacity
			},
			loadInventoryFirst: func(owner *Object) *Object {
				return owner.InvFirstItem
			},
			loadWeight: func(item *Object) uint8 {
				return item.Weight
			},
			loadInventoryNext: func(item *Object) *Object {
				return item.InvNextItem
			},
			primaryMessage: deps.primaryMessage,
			loadItemClass: func(item *Object) uint32 {
				return uint32(item.ObjClass)
			},
			loadItemType: func(item *Object) uint16 {
				return item.TypeInd
			},
			countInventory: func(owner *Object, typeInd int32) int32 {
				return owner.CountInventoryWithType(typeInd)
			},
			deleteWorldObject: deps.deleteWorldObject,
			inventoryPut:      deps.inventoryPut,
		},
	)
}

func pickupDefaultServerDeps4F31E0(
	s *Server,
	runtime PickupDefaultRuntime4F31E0,
) pickupDefaultNativeDeps4F31E0 {
	return pickupDefaultNativeDeps4F31E0{
		gameFlagsCheck: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		findTeam: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		informTeam: func(index, code uint8, color uint32) {
			s.NetInformTextMsg(ntype.PlayerInd(index), code, int(color))
		},
		primaryMessage: func(owner *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), value)
		},
		deleteWorldObject: runtime.DeleteWorldObject,
		inventoryPut:      runtime.InventoryPut,
	}
}

// PickupDefault4F31E0 runs GAME.EXE's four-argument default-pickup callback
// against native-width Object, Team, PlayerUpdateData, and Player layouts.
// The fourth argument is preserved at the ABI boundary and intentionally
// ignored by the reconstructed body.
func (s *Server) PickupDefault4F31E0(
	owner, item *Object,
	report, ignored int32,
	runtime PickupDefaultRuntime4F31E0,
) int32 {
	return pickupDefaultNative4F31E0(
		owner,
		item,
		report,
		ignored,
		pickupDefaultServerDeps4F31E0(s, runtime),
	)
}
