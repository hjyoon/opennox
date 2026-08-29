package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type moverXferNativeDeps4F5730 struct {
	waypointIndex     func(*server.Object, *server.MoverUpdateData, int) uint32
	transferInventory func(uint16, *server.Object, int32) int32
}

func moverXferRuntimeDeps4F5730() moverXferNativeDeps4F5730 {
	return moverXferNativeDeps4F5730{
		waypointIndex: func(object *server.Object, data *server.MoverUpdateData, slot int) uint32 {
			waypoint := object.MoverWaypointFor(data, slot)
			if waypoint == nil {
				return 0
			}
			return waypoint.Index
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func moverXferReadWriteNative4F5730(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func moverXferNative4F5730(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps moverXferNativeDeps4F5730,
) int32 {
	return moverXfer4F5730(
		object,
		moverXferDeps4F5730[*server.Object, *server.MoverUpdateData]{
			loadUpdateData: func(object *server.Object) *server.MoverUpdateData {
				// Preserve the entry pointer without allocation or class validation.
				return object.UpdateDataMover()
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
			rwField1: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_1), 4)
			},
			rwField2: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_2), 4)
			},
			rwField8: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_8), 4)
			},
			rwField0: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_0), 1)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			rwField4: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_4), 4)
			},
			rwField6: func(data *server.MoverUpdateData) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&data.Field_6), 4)
			},
			waypointIndex: func(object *server.Object, data *server.MoverUpdateData, slot int) uint32 {
				return deps.waypointIndex(object, data, slot)
			},
			rwWaypointIndex: func(value uint32) {
				// The original write-side temporary is a dword. Never write a
				// serialized index into either transient pointer slot.
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&value), 4)
			},
			rwSpeedBase: func(object *server.Object) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&object.SpeedBase), 4)
			},
			rwSpeedCur: func(object *server.Object) {
				moverXferReadWriteNative4F5730(cf, unsafe.Pointer(&object.SpeedCur), 4)
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerMoverNative4F5730(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return moverXferNative4F5730(cf, object, moverXferRuntimeDeps4F5730())
}
