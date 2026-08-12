package opennox

import "github.com/opennox/opennox/v1/server"

type deletedObjectsUpdate4E5E20Hooks struct {
	deletedList         func() *server.Object
	setDeletedList      func(*server.Object)
	frame               func() uint32
	removeFromUpdatable func(*server.Object)
	finish              func(*server.Object)
}

func deletedObjectsUpdate_4E5E20(hooks deletedObjectsUpdate4E5E20Hooks) {
	var list *server.Object
	for it := hooks.deletedList(); it != nil; {
		deletedAt := it.DeletedAt
		frame := hooks.frame()
		next := it.DeletedNext
		if deletedAt == frame {
			it.DeletedNext = list
			list = it
			hooks.removeFromUpdatable(it)
		} else {
			hooks.finish(it)
		}
		it = next
	}
	hooks.setDeletedList(list)
}
