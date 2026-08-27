package server

import "github.com/opennox/opennox/v1/common/sound"

// InventoryPutRuntime4F3070 supplies the two services in GAME.EXE 004F3070
// that remain implemented by the legacy network/protection layer. Object and
// Player references stay native-width across both callbacks.
type InventoryPutRuntime4F3070 struct {
	ReportPickup func(uint8, *Object)
	ProtectItem  func(uint32, *Object)
}

type inventoryPutNativeDeps4F3070 struct {
	setOwner   func(*Object, *Object)
	report     func(uint8, *Object)
	protect    func(uint32, *Object)
	audioEvent func(int32, *Object, int32, uint32)
}

func inventoryPutNative4F3070(
	owner, item *Object,
	report int32,
	deps inventoryPutNativeDeps4F3070,
) {
	inventoryPut4F3070(owner, item, report, inventoryPutHooks4F3070[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadFlagsLow: func(object *Object) uint8 {
			return uint8(object.ObjFlags)
		},
		loadClassLow: func(object *Object) uint8 {
			return uint8(object.ObjClass)
		},
		loadUpdate: func(object *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(object.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		storeInventoryPrev: func(object, previous *Object) {
			object.Field125 = previous
		},
		loadInventoryFirst: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		storeInventoryNext: func(object, next *Object) {
			object.InvNextItem = next
		},
		storeInventoryFirst: func(owner, first *Object) {
			owner.InvFirstItem = first
		},
		storeInventoryHolder: func(item, holder *Object) {
			item.InvHolder = holder
		},
		setOwner: deps.setOwner,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportPickup: deps.report,
		loadPlayerProtect: func(player *Player) uint32 {
			return player.Prot4632
		},
		protectItem: deps.protect,
		loadItemWeight: func(item *Object) uint8 {
			return item.Weight
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadCarryCapacity: func(owner *Object) uint16 {
			return owner.CarryCapacity
		},
		storePlayerOverweight: func(player *Player, value uint32) {
			player.Field3656 = value
		},
		audioEvent: deps.audioEvent,
	})
}

// InventoryPutImpl4F3070 prepends item to owner's inventory using native
// Object links and exact 32-bit report semantics. It deliberately retains the
// original unguarded PlayerUpdateData and Player loads.
//
//go:noinline
func (s *Server) InventoryPutImpl4F3070(
	owner, item *Object,
	report int32,
	runtime InventoryPutRuntime4F3070,
) {
	inventoryPutNative4F3070(owner, item, report, inventoryPutNativeDeps4F3070{
		setOwner: s.ObjSetOwner,
		report:   runtime.ReportPickup,
		protect:  runtime.ProtectItem,
		audioEvent: func(id int32, object *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), object, int(kind), code)
		},
	})
}
