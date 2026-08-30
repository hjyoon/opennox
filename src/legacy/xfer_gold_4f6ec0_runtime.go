package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type goldXferNativeDeps4F6EC0 struct {
	transferInventory func(uint16, *server.Object, int32) int32
}

func goldXferRuntimeDeps4F6EC0() goldXferNativeDeps4F6EC0 {
	return goldXferNativeDeps4F6EC0{
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func goldXferNative4F6EC0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps goldXferNativeDeps4F6EC0,
) int32 {
	return goldXfer4F6EC0(
		object,
		goldXferDeps4F6EC0[*server.Object, *server.GoldInitData]{
			loadInitData: func(object *server.Object) *server.GoldInitData {
				return object.InitDataGold()
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
			rwGoldAmount: func(data *server.GoldInitData) {
				data.Amount = objectReadOldRWU32Native4F4170(cf, data.Amount)
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

func Nox_xxx_XFerGoldNative4F6EC0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return goldXferNative4F6EC0(cf, object, goldXferRuntimeDeps4F6EC0())
}
