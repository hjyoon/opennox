package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func arachnaphobiaUpdateNative53DA60(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.ArachnaphobiaUpdate53DA60(source, server.ArachnaphobiaUpdateRuntime53DA60{
		NewObjectByTypeID: world.NewObjectByTypeID,
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		RandomInt:     world.Rand.Logic.IntClamp,
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var arachnaphobiaUpdateCall53DA60 = arachnaphobiaUpdateNative53DA60
