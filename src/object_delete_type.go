package opennox

import "github.com/opennox/opennox/v1/server"

// deleteAllObjectsOfType_4E5DB0 walks the active object list and each object's
// inventory in the same order as the original function. Both successors must
// be saved before delayedDelete because deletion may unlink the current node.
func deleteAllObjectsOfType_4E5DB0(first *server.Object, typeInd int, delayedDelete func(*server.Object)) {
	var next *server.Object
	for it := first; it != nil; it = next {
		next = it.Next()
		var nextItem *server.Object
		for item := it.FirstItem(); item != nil; item = nextItem {
			nextItem = item.NextItem()
			if int(item.TypeInd) == typeInd {
				delayedDelete(item)
			}
		}
		if int(it.TypeInd) == typeInd {
			delayedDelete(it)
		}
	}
}
