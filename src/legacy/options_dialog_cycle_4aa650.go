package legacy

import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_4AA650
func sub_4AA650() { optionsDialogCycleNative4AA650() }

func optionsDialogCycleNative4AA650() {
	if sub_44D930() != 0 {
		return
	}
	counter := memmap.PtrUint32(0x5D4594, 1309744)
	index := *counter
	if index >= 3 {
		return
	}
	key := (*C.char)(*memmap.PtrPtr(0x587000, 172892+4*uintptr(index)))
	*counter = index + 1
	var file *C.char
	const source = `C:\\NoxPost\\src\\client\\shell\\Options.c`
	nox_strman_loadString_40F1D0(key, &file, internCStr(source), 131)
	*counter = uint32(int32(*counter) % 3)
	if file != nil {
		nox_xxx_playDialogFile_44D900((*byte)(unsafe.Pointer(file)), 100)
	}
}
