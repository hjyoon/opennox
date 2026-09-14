package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// TeleportToMB4E7190 exposes the already restored movement gate to native Go
// spell callbacks without a round trip through a PE32 C function.
func TeleportToMB4E7190(obj *server.Object, pos types.Pointf) {
	teleportToMBObject4E7190(obj, func(obj *server.Object) {
		Nox_xxx_unitMove_4E7010(obj, pos)
	})
}

// teleportToMBObject4E7190 binds the independently tested gate contract to
// the native-width object and game-flag state used by the production path.
func teleportToMBObject4E7190(obj *server.Object, move func(*server.Object)) {
	teleportToMB4E7190(obj, teleportToMB4E7190Hooks[*server.Object]{
		anchored: func(obj *server.Object) bool { return obj.HasEnchant(server.ENCHANT_ANCHORED) },
		flags:    func(obj *server.Object) object.Flags { return obj.ObjFlags },
		quest:    func() bool { return noxflags.HasGame(noxflags.GameModeQuest) },
		class:    func(obj *server.Object) object.Class { return obj.ObjClass },
		subclass: func(obj *server.Object) object.SubClass { return obj.ObjSubClass },
		coop:     func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		move:     move,
	})
}
