package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// The original 36-byte leaf reads the live six-DWORD table in order and
// returns the first match. Keep the table address native-width throughout.
func sub409A70TableNative(value int16, table unsafe.Pointer) int32 {
	masked := uint32(uint16(value)) & 0x17f0
	for index := 0; index < 6; index++ {
		if *(*uint32)(unsafe.Add(table, 4*index)) == masked {
			runtime.KeepAlive(table)
			return int32(index)
		}
	}
	runtime.KeepAlive(table)
	return 0
}

func sub409A70Native(value int16) int32 {
	return sub409A70TableNative(value, memmap.PtrOff(0x587000, 4704))
}
