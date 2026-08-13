package legacy

import "github.com/opennox/opennox/v1/server"

var inventoryContainsEquivalentNativeHooks4E7EC0 = inventoryContainsEquivalentHooks4E7EC0[*server.Object]{
	loadFirst: func(owner *server.Object) *server.Object {
		return owner.InvFirstItem
	},
	equivalent: func(candidate, item *server.Object) bool {
		return itemEquivalent4E7DE0(candidate, item, itemEquivalentNativeHooks4E7DE0)
	},
	loadNext: func(candidate *server.Object) *server.Object {
		return candidate.InvNextItem
	},
}

// Sub_4E7EC0 exposes the native-width Go path. The two historical C callers
// use the typed C implementation while their wider pickup functions retain
// separately tracked ABI32 debt.
func Sub_4E7EC0(owner, item *server.Object) int32 {
	if inventoryContainsEquivalent4E7EC0(owner, item, inventoryContainsEquivalentNativeHooks4E7EC0) {
		return 1
	}
	return 0
}
