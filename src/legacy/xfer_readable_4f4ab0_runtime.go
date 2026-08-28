package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

// readableTextSizeWithNULNative4F4AB0 deliberately has no 256-byte bound.
// GAME.EXE 004F4AB0 calls strlen on the entry-time use-data pointer before it
// reads the transfer version, so malformed data faults or scans identically.
func readableTextSizeWithNULNative4F4AB0(data *server.ReadableUseData) uint32 {
	ptr := unsafe.Pointer(data)
	for size := uint32(1); ; size++ {
		if *(*byte)(unsafe.Add(ptr, uintptr(size-1))) == 0 {
			return size
		}
	}
}

// readableReadWriteTextNative4F4AB0 transfers the exact wire length. The
// original function does not clamp it to the 256-byte text member.
func readableReadWriteTextNative4F4AB0(
	cf *cryptfile.CryptFile,
	data *server.ReadableUseData,
	size uint32,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size)))
}

func Nox_xxx_XFerReadableNative4F4AB0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return readableXfer4F4AB0(
		object,
		readableXferDeps4F4AB0[*server.Object, *server.ReadableUseData]{
			loadUseData: func(object *server.Object) *server.ReadableUseData {
				// This direct native-width load is intentional. AsReadable performs
				// allocation-liveness validation that GAME.EXE does not perform here.
				return (*server.ReadableUseData)(object.UseData.Ptr)
			},
			textSizeWithNUL: readableTextSizeWithNULNative4F4AB0,
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			rwTextSize: func(value uint32) uint32 {
				return objectReadOldRWU32Native4F4170(cf, value)
			},
			rwText: func(data *server.ReadableUseData, size uint32) {
				readableReadWriteTextNative4F4AB0(cf, data, size)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			clearTransientRead: func(data *server.ReadableUseData) {
				data.TransientReadState = 0
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
