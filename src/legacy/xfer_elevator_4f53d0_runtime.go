package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type elevatorXferNativeDeps4F53D0 struct {
	transferInventory func(uint16, *server.Object, int32) int32
}

func elevatorXferRuntimeDeps4F53D0() elevatorXferNativeDeps4F53D0 {
	return elevatorXferNativeDeps4F53D0{
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func elevatorXferReadWriteNative4F53D0(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func elevatorXferNative4F53D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps elevatorXferNativeDeps4F53D0,
) int32 {
	return elevatorXfer4F53D0(
		object,
		elevatorXferDeps4F53D0[*server.Object, *server.ElevatorUpdateData]{
			loadUpdateData: func(object *server.Object) *server.ElevatorUpdateData {
				// Preserve the entry pointer without allocation or class validation.
				return object.UpdateDataElevator()
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
			rwShaftExtent: func(data *server.ElevatorUpdateData) {
				elevatorXferReadWriteNative4F53D0(cf, unsafe.Pointer(&data.Field_2), 4)
			},
			rwField4: func(data *server.ElevatorUpdateData) {
				elevatorXferReadWriteNative4F53D0(cf, unsafe.Pointer(&data.Field_4), 4)
			},
			rwField3: func(data *server.ElevatorUpdateData) {
				elevatorXferReadWriteNative4F53D0(cf, unsafe.Pointer(&data.Field_3), 1)
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

func Nox_xxx_XFerElevatorNative4F53D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return elevatorXferNative4F53D0(cf, object, elevatorXferRuntimeDeps4F53D0())
}
