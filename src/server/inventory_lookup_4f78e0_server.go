package server

// InventoryContains4F78E0 binds GAME.EXE 004F78E0 to the native Object
// layout. Both object identities and every inventory link remain pointer-width.
func InventoryContains4F78E0(holder, item *Object) int32 {
	return inventoryContains4F78E0(holder, item, inventoryContainsHooks4F78E0[*Object]{
		loadItemHolder: func(item *Object) *Object {
			return item.InvHolder
		},
		loadHolderFirst: func(holder *Object) *Object {
			return holder.InvFirstItem
		},
		loadItemNext: func(item *Object) *Object {
			return item.InvNextItem
		},
	})
}

// EquippedItemByCode4F7920 binds GAME.EXE 004F7920 to the native Object
// layout. Its result is an Object pointer rather than the PE32 integer used by
// the decompiler, and the full four-byte NetCode comparison is preserved.
func EquippedItemByCode4F7920(holder *Object, code uint32) *Object {
	return equippedItemByCode4F7920(holder, code, equippedItemByCodeHooks4F7920[*Object]{
		loadHolderFirst: func(holder *Object) *Object {
			return holder.InvFirstItem
		},
		loadItemNetCode: func(item *Object) uint32 {
			return item.NetCode
		},
		loadItemNext: func(item *Object) *Object {
			return item.InvNextItem
		},
	})
}
