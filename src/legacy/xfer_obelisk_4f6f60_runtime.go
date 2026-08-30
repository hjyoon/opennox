package legacy

import (
	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type obeliskXferNativeDeps4F6F60 struct {
	syncManaLevel     func(*server.Object, float32)
	gameFlags         func(uint32) int32
	staticDrawable    func(uint32) *client.Drawable
	firstMinimap      func() *client.Drawable
	transferInventory func(uint16, *server.Object, int32) int32
}

func obeliskXferRuntimeDeps4F6F60() obeliskXferNativeDeps4F6F60 {
	return obeliskXferNativeDeps4F6F60{
		// GAME.EXE 004E4770 (nullsub_35) intentionally has no side effects.
		syncManaLevel: func(*server.Object, float32) {},
		gameFlags: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		staticDrawable: func(code uint32) *client.Drawable {
			return GetClient().Cli().Objs.ByNetCodeStatic(int(code))
		},
		firstMinimap: func() *client.Drawable {
			return GetClient().Cli().Objs.FirstMinimapList()
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func obeliskXferNative4F6F60(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps obeliskXferNativeDeps4F6F60,
) int32 {
	return obeliskXfer4F6F60(
		object,
		obeliskXferDeps4F6F60[*server.Object, *server.ObeliskUpdateData, *client.Drawable]{
			loadUpdateData: func(object *server.Object) *server.ObeliskUpdateData {
				// Cache the raw native-width field at entry. Do not add class or
				// liveness checks that the original PE32 routine did not perform.
				return (*server.ObeliskUpdateData)(object.UpdateData)
			},
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			rwMana: func(data *server.ObeliskUpdateData) {
				data.Mana = int32(objectReadOldRWU32Native4F4170(cf, uint32(data.Mana)))
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			loadMana: func(data *server.ObeliskUpdateData) int32 {
				return data.Mana
			},
			syncManaLevel: deps.syncManaLevel,
			gameFlags:     deps.gameFlags,
			loadExtent: func(object *server.Object) uint32 {
				return object.Extent
			},
			staticDrawable: deps.staticDrawable,
			firstMinimap:   deps.firstMinimap,
			nextMinimap: func(drawable *client.Drawable) *client.Drawable {
				return drawable.Nox_xxx_cliNextMinimapObj_459EC0(drawable)
			},
			rwMinimapPresent: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerObeliskNative4F6F60(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return obeliskXferNative4F6F60(cf, object, obeliskXferRuntimeDeps4F6F60())
}
