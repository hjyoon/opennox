package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type invLightXferNativeDeps4F5AA0 struct {
	gameFlags         func(uint32) int32
	firstDrawable     func() *client.Drawable
	staticDrawable    func(uint32) *client.Drawable
	dynamicDrawable   func(uint32) *client.Drawable
	transferInventory func(uint16, *server.Object, int32) int32
}

func invLightXferRuntimeDeps4F5AA0() invLightXferNativeDeps4F5AA0 {
	return invLightXferNativeDeps4F5AA0{
		gameFlags: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		firstDrawable: func() *client.Drawable {
			return GetClient().Cli().Objs.FirstList1()
		},
		staticDrawable: func(code uint32) *client.Drawable {
			return GetClient().Cli().Objs.ByNetCodeStatic(int(code))
		},
		dynamicDrawable: func(code uint32) *client.Drawable {
			return GetClient().Cli().Objs.ByNetCodeDynamic(int(code))
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func invLightXferNative4F5AA0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps invLightXferNativeDeps4F5AA0,
) int32 {
	return invLightXfer4F5AA0(
		object,
		invLightXferDeps4F5AA0[*server.Object, *client.Drawable]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			gameFlags:     deps.gameFlags,
			firstDrawable: deps.firstDrawable,
			nextDrawable: func(drawable *client.Drawable) *client.Drawable {
				return drawable.NextPtr
			},
			loadDrawableCode: func(drawable *client.Drawable) uint32 {
				return drawable.NetCode32
			},
			loadExtent: func(object *server.Object) uint32 {
				return object.Extent
			},
			loadClass: func(object *server.Object) uint32 {
				return uint32(object.ObjClass)
			},
			loadNetCode: func(object *server.Object) uint32 {
				return object.NetCode
			},
			staticDrawable:  deps.staticDrawable,
			dynamicDrawable: deps.dynamicDrawable,
			copyDrawableLight: func(drawable *client.Drawable, light *[invLightXferPayloadSize4F5AA0]byte) {
				// GAME.EXE directly dereferences failed static/dynamic lookups.
				// LightXferData is nil-tolerant, so force the same invariant here.
				_ = drawable.LightFlags
				*light = drawable.LightXferData()
			},
			rwLight: func(light *[invLightXferPayloadSize4F5AA0]byte, offset, size int) {
				_, _ = cf.ReadWrite(light[offset : offset+size])
			},
			rwLegacyField43: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			legacyTooBright: invLightLegacyTooBright4F5AA0,
			clampLegacyLight: func(light *[invLightXferPayloadSize4F5AA0]byte) {
				invLightClampLegacy4F5AA0(light, func(value float32) uint32 {
					return uint32(client.LightRadius(value))
				})
			},
			copyObjectLight: func(object *server.Object, light *[invLightXferPayloadSize4F5AA0]byte) {
				copy(
					unsafe.Slice((*byte)(unsafe.Add(object.Field189, 2432)), len(light)),
					light[:],
				)
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerInvLightNative4F5AA0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return invLightXferNative4F5AA0(cf, object, invLightXferRuntimeDeps4F5AA0())
}
