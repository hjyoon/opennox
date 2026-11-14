package legacy

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func buffIterationCountNative424CB0(unit *server.Object) byte {
	var count byte
	for i := int32(0); i < buffIterationCountScalar424CB0(); i++ {
		buff := int32(*memmap.PtrUint32(0x587000, 66000+4*uintptr(i)))
		if unitBuffTestLegacy4FF350(unit, buff) != 0 {
			count++
		}
	}
	return count
}

func buffIterationFirstNative424D00() int32 {
	if buffIterationCountScalar424CB0() <= 0 {
		return -1
	}
	return int32(*memmap.PtrUint32(0x587000, 66000))
}

func buffIterationNextNative424D20(buff int32) int32 {
	count := buffIterationCountScalar424CB0()
	for i := int32(0); i < count; i++ {
		if int32(*memmap.PtrUint32(0x587000, 66000+4*uintptr(i))) == buff && i < count-1 {
			return int32(*memmap.PtrUint32(0x587000, 66004+4*uintptr(i)))
		}
	}
	return -1
}
