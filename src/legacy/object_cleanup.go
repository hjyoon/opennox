package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type objectCleanup4E5BF0Hooks struct {
	moonglowTypeInd           func() int
	firstObject               func() *server.Object
	nextObject                func(*server.Object) *server.Object
	firstMissile              func() *server.Object
	nextMissile               func(*server.Object) *server.Object
	isOfflineMigratingMonster func(*server.Object) bool
	isCoopPlayerPixie         func(*server.Object) bool
	delayedDelete             func(*server.Object)
}

func cleanupObjectsForMapLoad_4E5BF0(mode int, hooks objectCleanup4E5BF0Hooks) {
	moonglowTypeInd := hooks.moonglowTypeInd()
	for obj := hooks.firstObject(); obj != nil; {
		next := hooks.nextObject(obj)
		keep := false
		if mode != 0 {
			if obj.Class().Has(object.ClassPlayer) {
				keep = true
			} else if holder := obj.InvHolder; holder != nil && holder.Class().Has(object.ClassPlayer) {
				keep = true
			} else if int(obj.TypeInd) == moonglowTypeInd && obj.ObjOwner != nil &&
				obj.ObjOwner.Class().Has(object.ClassPlayer) {
				keep = true
			} else if hooks.isOfflineMigratingMonster(obj) {
				keep = true
			}
		}
		if !keep {
			hooks.delayedDelete(obj)
		}
		obj = next
	}

	for obj := hooks.firstMissile(); obj != nil; {
		next := hooks.nextMissile(obj)
		if mode != 1 || !hooks.isCoopPlayerPixie(obj) {
			hooks.delayedDelete(obj)
		}
		obj = next
	}
}
