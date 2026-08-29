package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type elevatorShaftXferNativeDeps4F54A0 struct {
	transferInventory func(uint16, *server.Object, int32) int32
}

func elevatorShaftXferRuntimeDeps4F54A0() elevatorShaftXferNativeDeps4F54A0 {
	return elevatorShaftXferNativeDeps4F54A0{
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func elevatorShaftXferReadWriteNative4F54A0(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func elevatorShaftXferNative4F54A0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps elevatorShaftXferNativeDeps4F54A0,
) int32 {
	return elevatorShaftXfer4F54A0(
		object,
		elevatorShaftXferDeps4F54A0[*server.Object, *server.ElevatorShaftUpdateData]{
			loadUpdateData: func(object *server.Object) *server.ElevatorShaftUpdateData {
				// Preserve the entry pointer without allocation or class validation.
				return object.UpdateDataElevatorShaft()
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
			rwElevatorExtent: func(data *server.ElevatorShaftUpdateData) {
				elevatorShaftXferReadWriteNative4F54A0(cf, unsafe.Pointer(&data.Field_2), 4)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerElevatorShaftNative4F54A0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return elevatorShaftXferNative4F54A0(cf, object, elevatorShaftXferRuntimeDeps4F54A0())
}
