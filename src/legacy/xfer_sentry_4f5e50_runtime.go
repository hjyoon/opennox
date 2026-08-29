package legacy

import (
	"fmt"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type sentryXferNativeDeps4F5E50 struct {
	gameFlags         func(uint32) int32
	transferInventory func(uint16, *server.Object, int32) int32
}

func sentryXferRuntimeDeps4F5E50() sentryXferNativeDeps4F5E50 {
	return sentryXferNativeDeps4F5E50{
		gameFlags: func(mask uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(mask)) {
				return 1
			}
			return 0
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func sentryXferReadWriteUpdateNative4F5E50(
	cf *cryptfile.CryptFile,
	data *server.SentryUpdateData,
	offset int,
) {
	switch offset {
	case 0:
		data.Field0 = objectReadOldRWU32Native4F4170(cf, data.Field0)
	case 4:
		data.Field4 = objectReadOldRWU32Native4F4170(cf, data.Field4)
	case 8:
		data.Field8 = objectReadOldRWU32Native4F4170(cf, data.Field8)
	default:
		panic(fmt.Sprintf("SentryXfer update offset %d", offset))
	}
}

func sentryXferLoadUpdateNative4F5E50(data *server.SentryUpdateData, offset int) uint32 {
	switch offset {
	case 0:
		return data.Field0
	case 4:
		return data.Field4
	case 8:
		return data.Field8
	default:
		panic(fmt.Sprintf("SentryXfer update offset %d", offset))
	}
}

func sentryXferStoreUpdateNative4F5E50(data *server.SentryUpdateData, offset int, value uint32) {
	switch offset {
	case 0:
		data.Field0 = value
	case 4:
		data.Field4 = value
	case 8:
		data.Field8 = value
	default:
		panic(fmt.Sprintf("SentryXfer update offset %d", offset))
	}
}

func sentryXferNative4F5E50(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps sentryXferNativeDeps4F5E50,
) int32 {
	return sentryXfer4F5E50(
		object,
		sentryXferDeps4F5E50[*server.Object, *server.SentryUpdateData]{
			loadUpdateData: func(object *server.Object) *server.SentryUpdateData {
				return object.UpdateDataSentry()
			},
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			rwUpdateData: func(data *server.SentryUpdateData, offset int) {
				sentryXferReadWriteUpdateNative4F5E50(cf, data, offset)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			gameFlags: deps.gameFlags,
			loadUpdateU32: func(data *server.SentryUpdateData, offset int) uint32 {
				return sentryXferLoadUpdateNative4F5E50(data, offset)
			},
			storeUpdateU32: func(data *server.SentryUpdateData, offset int, value uint32) {
				sentryXferStoreUpdateNative4F5E50(data, offset, value)
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerSentryNative4F5E50(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return sentryXferNative4F5E50(cf, object, sentryXferRuntimeDeps4F5E50())
}
