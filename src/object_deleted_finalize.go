package opennox

import "github.com/opennox/opennox/v1/server"

type finalizeDeletingObjects4E5EC0Hooks struct {
	deletedList    func() *server.Object
	setDeletedList func(*server.Object)
	finish         func(*server.Object)
}

// finalizeDeletingObjects_4E5EC0 saves each successor before completely
// deleting the current object, then clears the global list after traversal.
func finalizeDeletingObjects_4E5EC0(hooks finalizeDeletingObjects4E5EC0Hooks) {
	var next *server.Object
	for it := hooks.deletedList(); it != nil; it = next {
		next = it.DeletedNext
		hooks.finish(it)
	}
	hooks.setDeletedList(nil)
}
