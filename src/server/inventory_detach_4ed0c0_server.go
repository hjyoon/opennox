package server

// InventoryDetachRuntime4ED0C0 contains the legacy services called by
// GAME.EXE 004ED0C0. Object, update-data, and Player values remain native
// pointers; the callbacks that are still implemented in legacy C are kept
// explicit so their remaining ABI32 internals are not mistaken for a fully
// widened product boundary.
type InventoryDetachRuntime4ED0C0 struct {
	GameFlag        func(uint32) uint32
	NetReportDequip func(uint8, *Object)
	DequipArmor     func(*Object, *Object, int32, int32)
	DequipWeapon    func(*Object, *Object, int32, int32)
	NetReportDrop   func(uint8, *Object)
	ProtectItem     func(uint32, *Object)
	NPCSetItemEquip func(*Object, *Object, int32)
}

type inventoryDetachNativeDeps4ED0C0 struct {
	gameFlag        func(uint32) uint32
	netReportDequip func(uint8, *Object)
	dequipArmor     func(*Object, *Object, int32, int32)
	dequipWeapon    func(*Object, *Object, int32, int32)
	netReportDrop   func(uint8, *Object)
	protectItem     func(uint32, *Object)
	npcSetItemEquip func(*Object, *Object, int32)
	clearOwner      func(*Object)
}

func detachInventoryNative4ED0C0(
	owner, item *Object,
	deps inventoryDetachNativeDeps4ED0C0,
) {
	detachInventory4ED0C0(inventoryDetachHooks4ED0C0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadItemArg: func() *Object {
			return item
		},
		loadObjectClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadObjectSubclass: func(obj *Object) uint8 {
			return uint8(obj.ObjSubClass)
		},
		loadObjectUpdate: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		gameFlag: deps.gameFlag,
		loadUpdatePlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerField4: func(player *Player) uint32 {
			return player.WeaponEquip
		},
		storePlayerField4: func(player *Player, value uint32) {
			player.WeaponEquip = value
		},
		netReportDequip: deps.netReportDequip,
		dequipArmor:     deps.dequipArmor,
		dequipWeapon:    deps.dequipWeapon,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		netReportDrop: deps.netReportDrop,
		loadPlayerProtect: func(player *Player) uint32 {
			return player.Prot4632
		},
		protectItem:     deps.protectItem,
		npcSetItemEquip: deps.npcSetItemEquip,
		loadInventoryPrev: func(item *Object) *Object {
			return item.Field125
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		storeInventoryNext: func(item, next *Object) {
			item.InvNextItem = next
		},
		storeInventoryPrev: func(item, previous *Object) {
			item.Field125 = previous
		},
		loadInventoryFirst: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		storeInventoryFirst: func(owner, first *Object) {
			owner.InvFirstItem = first
		},
		storeInventoryHolder: func(item, holder *Object) {
			item.InvHolder = holder
		},
		clearOwner: deps.clearOwner,
		loadItemWeight: func(item *Object) uint8 {
			return item.Weight
		},
		loadCarryCapacity: func(owner *Object) uint16 {
			return owner.CarryCapacity
		},
		storePlayerOverweight: func(player *Player, value uint32) {
			player.Field3656 = value
		},
	})
}

// DetachInventory4ED0C0 removes item from owner's inventory while preserving
// the original dequip, network, protection, owner-clear, and encumbrance order.
func (s *Server) DetachInventory4ED0C0(
	owner, item *Object,
	runtime InventoryDetachRuntime4ED0C0,
) {
	detachInventoryNative4ED0C0(owner, item, inventoryDetachNativeDeps4ED0C0{
		gameFlag:        runtime.GameFlag,
		netReportDequip: runtime.NetReportDequip,
		dequipArmor:     runtime.DequipArmor,
		dequipWeapon:    runtime.DequipWeapon,
		netReportDrop:   runtime.NetReportDrop,
		protectItem:     runtime.ProtectItem,
		npcSetItemEquip: runtime.NPCSetItemEquip,
		clearOwner:      s.ObjClearOwner,
	})
}
