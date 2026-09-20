package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func telekinesisUpdateNative53D330(source *server.Object) {
	outer := GetServer()
	outer.S().TelekinesisUpdate53D330(source, server.TelekinesisUpdateRuntime53D330{
		Move: func(obj *server.Object, point types.Pointf) {
			Nox_xxx_unitMove_4E7010(obj, point)
		},
		DelayedDelete: outer.DelayedDelete,
		BuffOff:       Nox_xxx_spellBuffOff_4FF5B0,
	})
}

// Keep the indirection dynamic so tests can prove that thing.bin dispatches
// directly to Go while retaining the original C callback address as its key.
var telekinesisUpdateCall53D330 = telekinesisUpdateNative53D330
