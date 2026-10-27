package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

//export nox_xxx_updatePhantomPlayer_53B860_go
func nox_xxx_updatePhantomPlayer_53B860_go(source *nox_object_t) {
	phantomPlayerUpdateCall53B860(asObjectS(source))
}

func phantomPlayerUpdateNative53B860(source *server.Object) {
	outer := GetServer()
	outer.S().PhantomPlayerUpdate53B860(source, server.PhantomPlayerUpdateRuntime53B860{
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep the indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var phantomPlayerUpdateCall53B860 = phantomPlayerUpdateNative53B860
