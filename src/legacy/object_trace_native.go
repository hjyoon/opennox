package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func objectTraceNetCode4E4C00(owner *server.Object) uint32 {
	if owner == nil {
		return 0
	}
	return owner.NetCode
}

func objectTraceType4E4C10(owner *server.Object) uint16 {
	if owner == nil {
		return 0
	}
	return owner.TypeInd
}

func objectTraceInit4E4C30(owner *server.Object) unsafe.Pointer {
	if owner == nil {
		return nil
	}
	return owner.InitData
}

func objectTraceID4E4C80(owner *server.Object) unsafe.Pointer {
	if owner == nil {
		return nil
	}
	return owner.IDPtr
}

func objectTracePoint537760() *ntype.Point32 {
	if Get_dword_5d4594_2488620() == 0 {
		return nil
	}
	return memmap.PtrT[ntype.Point32](0x5D4594, 2488612)
}

func objectTraceFlags53B300(owner *server.Object) uint8 {
	server.LoopAndDamageUpdate53B300(owner)
	// The public C callback declares char; its upper EAX bits are incidental.
	return uint8(owner.ObjFlags)
}
