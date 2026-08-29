package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type transporterXferNativeDeps4F5300 struct {
	hasTarget         func(*server.Object, *server.TransporterUpdateData) bool
	transferInventory func(uint16, *server.Object, int32) int32
}

func transporterXferRuntimeDeps4F5300() transporterXferNativeDeps4F5300 {
	return transporterXferNativeDeps4F5300{
		hasTarget: func(object *server.Object, data *server.TransporterUpdateData) bool {
			return object.TransporterTargetFor(data) != nil
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func transporterXferReadWriteNative4F5300(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func transporterXferNative4F5300(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps transporterXferNativeDeps4F5300,
) int32 {
	return transporterXfer4F5300(
		object,
		transporterXferDeps4F5300[*server.Object, *server.TransporterUpdateData]{
			loadUpdateData: func(object *server.Object) *server.TransporterUpdateData {
				// Preserve the entry pointer without allocation or class validation.
				return object.UpdateDataTransporter()
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
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			rwTargetExtent: func(data *server.TransporterUpdateData) {
				transporterXferReadWriteNative4F5300(cf, unsafe.Pointer(&data.TargetExtent), 4)
			},
			hasTarget: func(data *server.TransporterUpdateData) bool {
				return deps.hasTarget(object, data)
			},
			loadTargetExtent: func(data *server.TransporterUpdateData) uint32 {
				return data.TargetExtent
			},
			rwLocalTargetExtent: func(value uint32) {
				transporterXferReadWriteNative4F5300(cf, unsafe.Pointer(&value), 4)
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerTransporterNative4F5300(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return transporterXferNative4F5300(cf, object, transporterXferRuntimeDeps4F5300())
}
