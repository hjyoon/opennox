package legacy

import "github.com/opennox/opennox/v1/server"

func breakUpdateNative53DB30(source *server.Object) {
	GetServer().S().BreakUpdate53DB30(source)
}

func breakAndRemoveUpdateNative53DC30(source *server.Object) {
	outer := GetServer()
	outer.S().BreakAndRemoveUpdate53DC30(source, server.BreakAndRemoveUpdateRuntime53DC30{
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep these indirections dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback addresses stored in thing.bin objects.
var (
	breakUpdateCall53DB30          = breakUpdateNative53DB30
	breakAndRemoveUpdateCall53DC30 = breakAndRemoveUpdateNative53DC30
)
