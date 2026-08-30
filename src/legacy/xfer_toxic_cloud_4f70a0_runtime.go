package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type toxicCloudXferNativeDeps4F70A0 struct {
	transferInventory func(uint16, *server.Object, int32) int32
}

func toxicCloudXferRuntimeDeps4F70A0() toxicCloudXferNativeDeps4F70A0 {
	return toxicCloudXferNativeDeps4F70A0{
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func toxicCloudXferNative4F70A0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps toxicCloudXferNativeDeps4F70A0,
) int32 {
	return toxicCloudXfer4F70A0(
		object,
		toxicCloudXferDeps4F70A0[*server.Object, *server.ToxicCloudUpdateData]{
			loadUpdateData: func(object *server.Object) *server.ToxicCloudUpdateData {
				// Cache the raw native-width field at entry. Do not add class or
				// liveness checks that the original PE32 routine did not perform.
				return (*server.ToxicCloudUpdateData)(object.UpdateData)
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
			rwDuration: func(data *server.ToxicCloudUpdateData) {
				data.Duration = objectReadOldRWI32Native4F4170(cf, data.Duration)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerToxicCloudNative4F70A0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return toxicCloudXferNative4F70A0(cf, object, toxicCloudXferRuntimeDeps4F70A0())
}
