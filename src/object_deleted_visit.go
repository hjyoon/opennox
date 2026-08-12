package opennox

import "github.com/opennox/opennox/v1/server"

type deletedObjectVisit4E5F00 func(*server.Object, int32) int32

// visitDeletedObjects_4E5F00 calls visit for objects not marked in the
// current frame. The successor is intentionally read after the callback.
func visitDeletedObjects_4E5F00(first *server.Object, frame func() uint32, visit deletedObjectVisit4E5F00, arg int32) {
	for it := first; it != nil; {
		deletedAt := it.DeletedAt
		currentFrame := frame()
		if deletedAt != currentFrame {
			visit(it, arg)
		}
		it = it.DeletedNext
	}
}
