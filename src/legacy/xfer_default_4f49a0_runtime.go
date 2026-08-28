package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_XFerDefaultNative4F49A0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return defaultXfer4F49A0(object, defaultXferDeps4F49A0[*server.Object]{
		loadField34: func(object *server.Object) uint32 {
			return object.Field34
		},
		rwVersion: func(value uint16) uint16 {
			return objectReadOldRWU16Native4F4170(cf, value)
		},
		mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
			return objectMapReadWriteNative4F4530(cf, object, mapVersion)
		},
		readOnly: func() int32 {
			if cf.ReadOnly() {
				return 1
			}
			return 0
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
		storeField34: func(object *server.Object, value uint32) {
			object.Field34 = value
		},
	})
}
