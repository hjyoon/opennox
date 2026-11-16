package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

func objectFieldMonsterAction5408A0(owner *server.Object) int32 {
	data := (*server.MonsterUpdateData)(owner.UpdateData)
	var action uint32
	if data.AIStackInd == -1 {
		// The original empty-head read addresses Field132 before the PE32
		// stack. Keep that named field separate from native stack indexing.
		action = data.Field132
	} else {
		action = data.AIStack[data.AIStackInd].Action
	}
	if action >= 18 && action <= 20 {
		return 1
	}
	return 0
}

func objectFieldMask1404E4500(owner *server.Object, value, mask uint32, set bool) unsafe.Pointer {
	owner.Sub_4E4500(value, mask, set)
	return unsafe.Add(unsafe.Pointer(&owner.Field140[0]), unsafe.Sizeof(owner.Field140))
}

func objectFieldSameArmor53E7B0(owner, item *server.Object) *server.Object {
	for it := owner.InvFirstItem; it != nil; it = it.InvNextItem {
		if uint32(it.ObjFlags)&0x100 != 0 && uint32(it.ObjClass)&0x2000000 != 0 && it.ObjSubClass == item.ObjSubClass {
			return it
		}
	}
	return nil
}

func objectFieldDrawableCode578B00(drawable *client.Drawable) uint32 {
	if drawable == nil || drawable.NetCode32 >= 0x8000 {
		return 0
	}
	code := drawable.NetCode32
	if uint32(drawable.ObjClass)&0x20400000 != 0 {
		code |= 0x8000
	}
	return code
}

// Keep a Go-owned receiver pinned until its complete C call has returned.
// Pinner ignores foreign C owners; the callback keeps its typed pointer result.
func pinObjectForC(owner *server.Object) func() {
	var pin runtime.Pinner
	pin.Pin(owner)
	return pin.Unpin
}
