package legacy

/*
#include "defs.h"
#include "mapgen_legacy_ptr.h"
float* nox_mapgenMakeHallStructNative_523E30(int, int, int);
int nox_xxx_mapGenSetRoomPos_521880(uint32_t*, float2*);
void* nox_mapgenRoomAtNative_521200(uint8_t*);
int nox_xxx_mapGenAddNewRoom_521730(uint32_t*);
int nox_mapgenRoomLinkBothNative_521A70(uint8_t*, uint8_t*, int);
void sub_521A10(void*);
*/
import "C"

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"unsafe"
)

func game5ConnectorCount() *uint32         { return memmap.PtrUint32(0x5D4594, 2491608) }
func game5ConnectorToken(i uint32) *uint32 { return memmap.PtrUint32(0x5D4594, 2491612+4*uintptr(i)) }
func game5ConnectorRoom(i uint32) *C.uint8_t {
	return (*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(*game5ConnectorToken(i))))
}
func game5ConnectorNew(kind, width, length int32) unsafe.Pointer {
	room := unsafe.Pointer(C.nox_mapgenMakeHallStructNative_523E30(C.int(kind), C.int(width), C.int(length)))
	*game5ConnectorToken(*game5ConnectorCount()) = uint32(C.nox_mapgenLegacyPtrRegister(room))
	return room
}
func game5ConnectorPosition(room unsafe.Pointer, x, y int32) {
	p := types.Ptf(float32(float64(x)*32.526913), float32(float64(y)*32.526913))
	C.nox_xxx_mapGenSetRoomPos_521880((*C.uint32_t)(room), (*C.float2)(unsafe.Pointer(&p)))
}
func game5ConnectorLink(first, second uint32, direction int32) {
	C.nox_mapgenRoomLinkBothNative_521A70(game5ConnectorRoom(first), game5ConnectorRoom(second), C.int(direction))
}
func game5ConnectorCommit(first, second uintptr, d1, d2 int32) C.int {
	count := int32(*game5ConnectorCount())
	for i := int32(0); i < count; i++ {
		room := game5ConnectorRoom(uint32(i))
		if room != nil && C.nox_mapgenRoomAtNative_521200(room) != nil {
			for j := int32(0); j < count; j++ {
				if allocated := game5ConnectorRoom(uint32(j)); allocated != nil {
					C.sub_521A10(unsafe.Pointer(allocated))
				}
				*game5ConnectorToken(uint32(j)) = 0
			}
			*game5ConnectorCount() = 0
			return 0
		}
	}
	for i := int32(0); i < count; i++ {
		if room := game5ConnectorRoom(uint32(i)); room != nil {
			C.nox_xxx_mapGenAddNewRoom_521730((*C.uint32_t)(unsafe.Pointer(room)))
		}
	}
	if count <= 0 {
		return 0
	}
	C.nox_mapgenRoomLinkBothNative_521A70((*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(first))), game5ConnectorRoom(0), C.int(d1))
	C.nox_mapgenRoomLinkBothNative_521A70((*C.uint8_t)(C.nox_mapgenLegacyPtrResolve(C.uintptr_t(second))), game5ConnectorRoom(uint32(count-1)), C.int(d2))
	return 1
}

//export sub_54BA60
func sub_54BA60(first, second C.uintptr_t, d1, d2 C.int) C.int {
	return game5ConnectorCommit(uintptr(first), uintptr(second), int32(d1), int32(d2))
}

func game5ConnectVertical(first, second uintptr, a, b [2]int32, width int32) C.int {
	dx := b[0] - a[0]
	distance := dx
	if distance < 0 {
		distance = -distance
	}
	*game5ConnectorCount() = 0
	if distance != 0 {
		if distance < 3 {
			return 0
		}
		middle := a[1] - (a[1]-b[1])/2 - 1
		room := game5ConnectorNew(2, width, a[1]-middle)
		game5ConnectorPosition(room, a[0], middle)
		*game5ConnectorCount()++
		if dx <= 0 {
			room = game5ConnectorNew(5, width, a[0]-b[0])
			game5ConnectorLink(0, 1, 3)
			game5ConnectorPosition(room, b[0], middle)
		} else {
			room = game5ConnectorNew(4, width, b[0]-a[0])
			game5ConnectorLink(0, 1, 2)
			game5ConnectorPosition(room, width+a[0], middle)
		}
		*game5ConnectorCount()++
		room = game5ConnectorNew(2, width, middle-b[1]-1)
		game5ConnectorPosition(room, b[0], b[1]+1)
		game5ConnectorLink(1, 2, 0)
	} else {
		room := game5ConnectorNew(2, width, a[1]-b[1]-1)
		game5ConnectorPosition(room, b[0], b[1]+1)
	}
	*game5ConnectorCount()++
	return game5ConnectorCommit(first, second, 0, 1)
}

//export sub_54B810
func sub_54B810(first, second C.uintptr_t, a *C.int, b *C.int2, width C.int) C.int {
	return game5ConnectVertical(uintptr(first), uintptr(second), *(*[2]int32)(unsafe.Pointer(a)), *(*[2]int32)(unsafe.Pointer(b)), int32(width))
}

func game5ConnectHorizontal(first, second uintptr, a, b [2]int32, width int32) C.int {
	dy := b[1] - a[1]
	distance := dy
	if distance < 0 {
		distance = -distance
	}
	*game5ConnectorCount() = 0
	if distance != 0 {
		if distance < 3 {
			return 0
		}
		middle := a[0] + (b[0]-a[0])/2
		room := game5ConnectorNew(4, width, middle-a[0]+width-1)
		game5ConnectorPosition(room, a[0]+1, a[1])
		*game5ConnectorCount()++
		if dy <= 0 {
			room = game5ConnectorNew(2, width, a[1]-b[1])
			game5ConnectorLink(0, 1, 0)
			game5ConnectorPosition(room, middle, b[1])
		} else {
			room = game5ConnectorNew(3, width, b[1]-a[1])
			game5ConnectorLink(0, 1, 1)
			game5ConnectorPosition(room, middle, a[1]+width)
		}
		*game5ConnectorCount()++
		room = game5ConnectorNew(4, width, b[0]-middle-width)
		game5ConnectorPosition(room, width+middle, b[1])
		game5ConnectorLink(1, 2, 2)
	} else {
		room := game5ConnectorNew(4, width, b[0]-a[0]-1)
		game5ConnectorPosition(room, a[0]+1, a[1])
	}
	*game5ConnectorCount()++
	return game5ConnectorCommit(first, second, 2, 3)
}

//export sub_54BB20
func sub_54BB20(first, second C.uintptr_t, a *C.int, b *C.uint32_t, width C.int) C.int {
	return game5ConnectHorizontal(uintptr(first), uintptr(second), *(*[2]int32)(unsafe.Pointer(a)), *(*[2]int32)(unsafe.Pointer(b)), int32(width))
}

func game5ConnectElbow(first, second uintptr, a, b [2]int32, width int32, south bool) C.int {
	if south {
		if b[1] <= a[1] {
			return 0
		}
	} else if b[1] > a[1]-width {
		return 0
	}
	*game5ConnectorCount() = 0
	var room unsafe.Pointer
	if south {
		room = game5ConnectorNew(3, width, b[1]-a[1]+width-1)
		game5ConnectorPosition(room, a[0], a[1]+1)
	} else {
		room = game5ConnectorNew(2, width, a[1]-b[1])
		game5ConnectorPosition(room, a[0], b[1])
	}
	*game5ConnectorCount()++
	direction := int32(2)
	if b[0]-a[0] <= 0 {
		room = game5ConnectorNew(5, width, a[0]-b[0]-1)
		game5ConnectorLink(0, 1, 3)
		game5ConnectorPosition(room, b[0]+1, b[1])
	} else {
		room = game5ConnectorNew(4, width, b[0]-a[0]-width)
		game5ConnectorLink(0, 1, 2)
		game5ConnectorPosition(room, width+a[0], b[1])
		direction = 3
	}
	*game5ConnectorCount()++
	if south {
		return game5ConnectorCommit(first, second, 1, direction)
	}
	return game5ConnectorCommit(first, second, 0, direction)
}

//export sub_54BD90
func sub_54BD90(first, second C.uintptr_t, a, b *C.int, width C.int) C.int {
	return game5ConnectElbow(uintptr(first), uintptr(second), *(*[2]int32)(unsafe.Pointer(a)), *(*[2]int32)(unsafe.Pointer(b)), int32(width), false)
}

//export sub_54BF20
func sub_54BF20(first, second C.uintptr_t, a, b *C.int, width C.int) C.int {
	return game5ConnectElbow(uintptr(first), uintptr(second), *(*[2]int32)(unsafe.Pointer(a)), *(*[2]int32)(unsafe.Pointer(b)), int32(width), true)
}

//export sub_54B2D0
func sub_54B2D0(prefabRaw *C.uint8_t, direction C.int, candidateRaw *C.uint8_t) C.int {
	// Invalid side values have no native room-link operation.
	if direction < 0 || direction > 3 {
		return 0
	}
	prefab := (*[38]int32)(unsafe.Pointer(prefabRaw))
	candidate := (*[5]int32)(unsafe.Pointer(candidateRaw))
	side := int(direction)
	width := prefab[4*side+22]
	first, second := uintptr(uint32(prefab[37])), uintptr(unsafe.Pointer(candidateRaw))
	start := func(offset int) [2]int32 { return [2]int32{prefab[offset], prefab[offset+1]} }
	if side == 0 || side == 1 {
		limit := candidate[3] - width
		p := [2]int32{candidate[1] + candidate[3]/2, candidate[2]}
		if side == 0 {
			p[1] = candidate[4] + candidate[2] - 1
		}
		for i := int32(0); i <= limit; i++ {
			var ok C.int
			if side == 0 {
				ok = game5ConnectVertical(first, second, start(20), p, prefab[22])
			} else {
				ok = game5ConnectVertical(second, first, p, start(24), prefab[26])
			}
			if ok != 0 {
				return 1
			}
			p[0]++
			if p[0] > candidate[1]+limit {
				p[0] = candidate[1]
			}
		}
		limit = candidate[4] - width
		p[1] = candidate[2] + candidate[4]/2
		for i := int32(0); i <= limit; i++ {
			p[0] = candidate[1]
			offset := 20
			if side == 1 {
				offset = 24
			}
			if game5ConnectElbow(first, second, start(offset), p, prefab[offset+2], side == 1) != 0 {
				return 1
			}
			p[0] += candidate[3] - 1
			if game5ConnectElbow(first, second, start(offset), p, prefab[offset+2], side == 1) != 0 {
				return 1
			}
			p[1]++
			if p[1] > candidate[2]+limit {
				p[1] = candidate[2]
			}
		}
		return 0
	}
	limit := candidate[3] - width // Preserve the original width-based trial count on both axes.
	p := [2]int32{candidate[1], candidate[2] + candidate[4]/2}
	if side == 3 {
		p[0] += candidate[3] - 1
	}
	for i := int32(0); i <= limit; i++ {
		var ok C.int
		if side == 2 {
			ok = game5ConnectHorizontal(first, second, start(28), p, prefab[30])
		} else {
			ok = game5ConnectHorizontal(second, first, p, start(32), prefab[34])
		}
		if ok != 0 {
			return 1
		}
		p[1]++
		if p[1] > candidate[2]+limit {
			p[1] = candidate[2]
		}
	}
	limit = candidate[3] - width
	p[0] = candidate[1] + candidate[3]/2
	for i := int32(0); i <= limit; i++ {
		p[1] = candidate[2]
		offset := 28
		if side == 3 {
			offset = 32
		}
		if game5ConnectElbow(second, first, p, start(offset), prefab[offset+2], false) != 0 {
			return 1
		}
		p[1] += candidate[4] - 1
		if game5ConnectElbow(second, first, p, start(offset), prefab[offset+2], true) != 0 {
			return 1
		}
		p[0]++
		if p[0] > candidate[1]+limit {
			p[0] = candidate[1]
		}
	}
	return 0
}
