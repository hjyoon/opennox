package opennox

import "github.com/opennox/opennox/v1/server"

func unitRemoveChild4EC470(parent *server.Object) {
	unitRemoveChildContract4EC470(parent, unitRemoveChildHooks4EC470[*server.Object]{
		loadFirstOwned: func(parent *server.Object) *server.Object {
			return parent.Field129
		},
		loadNextOwned: func(child *server.Object) *server.Object {
			return child.Field128
		},
		storeOwner: func(child, owner *server.Object) {
			child.ObjOwner = owner
		},
		storeNextOwned: func(child, next *server.Object) {
			child.Field128 = next
		},
		storeFirstOwned: func(parent, first *server.Object) {
			parent.Field129 = first
		},
	})
}
