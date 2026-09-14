package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

var moonglowUpdateCall53D270 = func(visual *server.Object) {
	outer := GetServer()
	world := outer.S()
	server.MoonglowUpdate53D270(visual, server.MoonglowUpdateRuntime53D270{
		Frame:         world.Frame,
		FPS:           world.TickRate,
		ValidPosition: world.Map.ValidIndexPos,
		Move: func(obj *server.Object, point types.Pointf) {
			Nox_xxx_unitMove_4E7010(obj, point)
		},
		DelayedDelete: outer.DelayedDelete,
		BuffOff:       Nox_xxx_spellBuffOff_4FF5B0,
	})
}
