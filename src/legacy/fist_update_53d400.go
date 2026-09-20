package legacy

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func fistUpdateNative53D400(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.FistUpdate53D400(source, server.FistUpdateRuntime53D400{
		AudioEvent: func(id uint32, obj *server.Object) {
			world.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		MakeScorch: Nox_xxx_sMakeScorch_537AF0,
		SendPointFX: func(code uint8, position types.Pointf) {
			world.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), position)
		},
		Earthquake:    world.Nox_xxx_earthquakeSend_4D9110,
		DelayedDelete: outer.DelayedDelete,
	})
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var fistUpdateCall53D400 = fistUpdateNative53D400
