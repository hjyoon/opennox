package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// ModifierDefendActualCallbacks4E03D0 returns the actual registered C
// identities only when all four public entries have native signatures and
// ordinary production field/parent/call services are active.
func ModifierDefendActualCallbacks4E03D0() ([2]unsafe.Pointer, bool) {
	if modifierPredicateObserve4E03D0 != nil || inversionInventoryObserveMemory4FA4F0 != nil || inversionInventoryObserveCall4FA4F0 != nil || server.ObserveOwnerChainPlayerRead4EC580 != nil || inversionInventoryABI4FA4F0() != 2 {
		return [2]unsafe.Pointer{}, false
	}
	for i := 0; i < 3; i++ {
		if modifierPredicatesABI4E03D0(i) != 2 {
			return [2]unsafe.Pointer{}, false
		}
	}
	return [2]unsafe.Pointer{modifierPredicatesCallback4E03D0(0), modifierPredicatesCallback4E03D0(1)}, true
}

// ModifierDefendActualDefaultC4E03D0 is transport for product diagnostics.
// Entries 0/1 call the real Inversion/Grip function pointers; 2/3 call the
// public Grip scan/inventory checker. No game algorithm or world result is
// supplied here. The caller owns the objects and output for the whole call.
func ModifierDefendActualDefaultC4E03D0(entry int, root, item, target *server.Object, effect *server.ModifierEff, output *int32) (int32, bool) {
	functions, ok := ModifierDefendActualCallbacks4E03D0()
	if !ok {
		return 0, false
	}
	switch entry {
	case 0, 1:
		args := [4]uintptr{uintptr(unsafe.Pointer(item)), uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(target)), uintptr(unsafe.Pointer(root))}
		return modifierPredicateInvoke4E03D0(functions[entry], unsafe.Pointer(effect), args, output), true
	case 2:
		return modifierPredicateNativeScanCall4E03D0(uintptr(unsafe.Pointer(root)), uintptr(unsafe.Pointer(target)), unsafe.Pointer(item), uintptr(unsafe.Pointer(root))), true
	case 3:
		return inversionInventoryPublic4FA4F0(unsafe.Pointer(root), unsafe.Pointer(target)), true
	default:
		return 0, false
	}
}
