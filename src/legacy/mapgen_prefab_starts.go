package legacy

/*
#include "GAME4_2.h"

extern uint32_t dword_5d4594_2487620, dword_5d4594_2487624;
extern uint32_t dword_5d4594_2487628, dword_5d4594_2487632, dword_5d4594_2487652;
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func mapgenStorePrefabStartPositions(positions [][2]float32) {
	cells := [...]*uint32{
		memmap.PtrUint32(0x5D4594, 2487612), memmap.PtrUint32(0x5D4594, 2487616),
		(*uint32)(unsafe.Pointer(&C.dword_5d4594_2487620)), (*uint32)(unsafe.Pointer(&C.dword_5d4594_2487624)),
		(*uint32)(unsafe.Pointer(&C.dword_5d4594_2487628)), (*uint32)(unsafe.Pointer(&C.dword_5d4594_2487632)),
		memmap.PtrUint32(0x5D4594, 2487636), memmap.PtrUint32(0x5D4594, 2487640),
		memmap.PtrUint32(0x5D4594, 2487644), memmap.PtrUint32(0x5D4594, 2487648),
	}
	for index, point := range positions {
		*cells[2*index] = math.Float32bits(point[0])
		*cells[2*index+1] = math.Float32bits(point[1])
	}
}

func mapgenPrefabStarts(count int, size float32, roll func() int, coordinate func() float32) [5][2]float32 {
	var positions [5][2]float32
	switch count {
	case 1:
		positions[0] = [2]float32{coordinate(), coordinate()}
	case 2:
		if roll() >= 50 {
			positions[0] = [2]float32{-size, coordinate()}
			positions[1] = [2]float32{size, coordinate()}
		} else {
			positions[0] = [2]float32{coordinate(), -size}
			positions[1] = [2]float32{coordinate(), size}
		}
	case 3:
		switch roll := roll(); {
		case roll < 25:
			positions = [5][2]float32{{-size, -size}, {size, -size}, {0, size}}
		case roll < 50:
			positions = [5][2]float32{{-size, size}, {size, size}, {0, -size}}
		case roll < 75:
			positions = [5][2]float32{{size, -size}, {size, size}, {-size, 0}}
		default:
			positions = [5][2]float32{{-size, -size}, {-size, size}, {size, 0}}
		}
	default:
		positions = [5][2]float32{{-size, -size}, {size, -size}, {-size, size}, {size, size}}
	}
	return positions
}

func mapgenShufflePrefabStarts(order []uint32, random func(int, int) int) {
	for i := range order {
		order[i] = uint32(i)
	}
	switch count := len(order); {
	case count < 2:
	case count == 2:
		if random(0, 100) >= 50 {
			order[0], order[1] = order[1], order[0]
		}
	default:
		for range order {
			index := random(0, count-2)
			order[index], order[index+1] = order[index+1], order[index]
		}
	}
}

//export sub_525DF0
func sub_525DF0(raw *C.uint8_t) {
	theme := (*[mapgenThemeWordCount]C.uint32_t)(unsafe.Pointer(raw))
	count := min(int(int32(theme[21])), 5)
	size := float32(mapgenRoomFloat(theme[16])) * 0.5
	*memmap.PtrUint32(0x5D4594, 2487608) = 0
	C.dword_5d4594_2487652 = C.uint32_t(count)
	if count <= 0 {
		return
	}
	order := (*[5]uint32)(memmap.PtrOff(0x5D4594, 2487588))
	random := func(low, high int) int {
		return int(nox_xxx_mapGenRandFunc_526AC0(C.int(low), C.int(high)))
	}
	mapgenShufflePrefabStarts(order[:count], random)
	positions := mapgenPrefabStarts(count, size, func() int { return random(0, 100) }, func() float32 {
		return float32(C.sub_526BC0(C.float(-size), C.float(size)))
	})
	written := count
	if count >= 4 {
		written = 5
	}
	mapgenStorePrefabStartPositions(positions[:written])
}
