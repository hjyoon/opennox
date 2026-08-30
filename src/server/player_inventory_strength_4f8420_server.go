package server

// PlayerInventoryStrengthRuntime4F8420 supplies the restored force-drop
// service. Inventory traversal, flags, strength checks, and all Object links
// stay inside native-width server code.
type PlayerInventoryStrengthRuntime4F8420 struct {
	ForceDrop func(*Object, *Object) int32
}

type playerInventoryStrengthNativeDeps4F8420 struct {
	checkStrength func(*Object, *Object) int32
	forceDrop     func(*Object, *Object) int32
}

func playerInventoryStrengthNative4F8420(
	player *Object,
	deps playerInventoryStrengthNativeDeps4F8420,
) {
	playerInventoryStrength4F8420(player, playerInventoryStrengthHooks4F8420[*Object]{
		loadInventoryHead: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		loadItemFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		checkStrength: deps.checkStrength,
		forceDrop:     deps.forceDrop,
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
	})
}

// PlayerInventoryStrength4F8420 scans the player's live inventory, dropping
// equipped items whose restored native strength check returns zero.
//
//go:noinline
func (s *Server) PlayerInventoryStrength4F8420(
	player *Object,
	runtime PlayerInventoryStrengthRuntime4F8420,
) {
	playerInventoryStrengthNative4F8420(player, playerInventoryStrengthNativeDeps4F8420{
		checkStrength: s.PlayerCheckStrength4F3180,
		forceDrop:     runtime.ForceDrop,
	})
}
