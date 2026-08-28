package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

// exitMapNameSizeWithNULNative4F4B90 deliberately has no 80-byte bound.
// GAME.EXE 004F4B90 runs strlen on the entry-time collide-data pointer before
// reading the transfer version, so malformed data faults or scans identically.
func exitMapNameSizeWithNULNative4F4B90(data *server.ExitCollideData) uint32 {
	ptr := unsafe.Pointer(data)
	for size := uint32(1); ; size++ {
		if *(*byte)(unsafe.Add(ptr, uintptr(size-1))) == 0 {
			return size
		}
	}
}

// exitReadWriteBytesNative4F4B90 transfers the exact wire length from the
// supplied collide-data offset. The original function does not clamp the map
// name transfer to the 80-byte member.
func exitReadWriteBytesNative4F4B90(
	cf *cryptfile.CryptFile,
	data *server.ExitCollideData,
	offset uintptr,
	size uint32,
) {
	ptr := unsafe.Add(unsafe.Pointer(data), offset)
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(ptr), int(size)))
}

func Nox_xxx_XFerExitNative4F4B90(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return exitXfer4F4B90(
		object,
		exitXferDeps4F4B90[*server.Object, *server.ExitCollideData]{
			loadCollideData: func(object *server.Object) *server.ExitCollideData {
				// This direct native-width load is intentional. Allocation-liveness
				// validation would add behavior that GAME.EXE does not have here.
				return (*server.ExitCollideData)(object.CollideData)
			},
			mapNameSizeWithNUL: exitMapNameSizeWithNULNative4F4B90,
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
			rwMapNameSize: func(value uint32) uint32 {
				return objectReadOldRWU32Native4F4170(cf, value)
			},
			rwMapName: func(data *server.ExitCollideData, size uint32) {
				exitReadWriteBytesNative4F4B90(cf, data, 0, size)
			},
			rwLegacyMapNameByte: func(data *server.ExitCollideData, offset uint32) {
				exitReadWriteBytesNative4F4B90(cf, data, uintptr(offset), 1)
			},
			loadMapNameByte: func(data *server.ExitCollideData, offset uint32) byte {
				return *(*byte)(unsafe.Add(unsafe.Pointer(data), uintptr(offset)))
			},
			rwDestinationX: func(data *server.ExitCollideData) {
				exitReadWriteBytesNative4F4B90(
					cf, data, unsafe.Offsetof(server.ExitCollideData{}.DestinationX), 4,
				)
			},
			rwDestinationY: func(data *server.ExitCollideData) {
				exitReadWriteBytesNative4F4B90(
					cf, data, unsafe.Offsetof(server.ExitCollideData{}.DestinationY), 4,
				)
			},
			transferInventory: func(version uint16, object *server.Object, count int32) int32 {
				return xferInventoryCall4F3E30(object, version, count)
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}
