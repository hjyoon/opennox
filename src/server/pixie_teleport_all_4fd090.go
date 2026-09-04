package server

const pixieDeadFlag4FD090 = uint32(0x8000)

type pixieTeleportAllHooks4FD090[Object comparable, UpdateData any] struct {
	loadOwnerArg    func() Object
	loadFirstOwned  func(Object) Object
	loadPixieTypeID func() uint32
	loadTypeInd     func(Object) uint16
	loadFlags       func(Object) uint32
	loadUpdateData  func(Object) UpdateData
	loadTarget      func(UpdateData) Object
	teleport        func(Object, Object)
	loadNextOwned   func(Object) Object
}

// pixieTeleportAll4FD090 preserves GAME.EXE 004FD090's exact owned-list
// traversal and branch-dependent load order. The full uint32 Pixie type ID is
// reloaded before each object's zero-extended uint16 TypeInd. Flags and update
// data are read only after their preceding gates pass, and NextOwned is loaded
// from the current object only after an optional teleport callback returns.
// Thus, a callback mutation of that link affects the next iteration.
//
// There are deliberately no nil guards. A nil owner reaches loadFirstOwned,
// and a matching live Pixie with nil update data reaches loadTarget, matching
// the original PE32 fault boundaries.
func pixieTeleportAll4FD090[Object comparable, UpdateData any](
	hooks pixieTeleportAllHooks4FD090[Object, UpdateData],
) {
	var zero Object
	owner := hooks.loadOwnerArg()
	pixie := hooks.loadFirstOwned(owner)
	for pixie != zero {
		typeID := hooks.loadPixieTypeID()
		typeInd := uint32(hooks.loadTypeInd(pixie))
		if typeInd == typeID {
			flags := hooks.loadFlags(pixie)
			if flags&pixieDeadFlag4FD090 == 0 {
				updateData := hooks.loadUpdateData(pixie)
				target := hooks.loadTarget(updateData)
				if target == zero {
					hooks.teleport(pixie, owner)
				}
			}
		}
		pixie = hooks.loadNextOwned(pixie)
	}
}
