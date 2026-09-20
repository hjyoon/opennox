package legacy

import "github.com/opennox/opennox/v1/server"

func selfDestructUpdateNative53CC90(source *server.Object) {
	outer := GetServer()
	outer.S().SelfDestructUpdate53CC90(source, server.SelfDestructUpdateRuntime53CC90{
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var selfDestructUpdateCall53CC90 = selfDestructUpdateNative53CC90
