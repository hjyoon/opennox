package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func skullUpdateNative54F9A0(source *server.Object) {
	outer := GetServer()
	world := outer.S()
	world.SkullUpdate54F9A0(source, server.SkullUpdateRuntime54F9A0{
		CreateAt: func(object, owner *server.Object, position types.Pointf) {
			outer.CreateObjectAt(object, owner, position)
		},
		InFront: func(source, target *server.Object) bool {
			return Nox_server_testTwoPointsAndDirection_4E6E50(
				source.PosVec,
				int16(source.Direction1),
				target.PosVec,
			)&1 != 0
		},
	})
}

// Keep the indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var skullUpdateCall54F9A0 = skullUpdateNative54F9A0
