package legacy

import (
	"image"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// Keep server lookup inside the actual world services: a disabled source or
// filtered candidate does not require a live world in the original entry.
var blowUpdateRuntime53C160 = func() server.BlowUpdateRuntime53C160 {
	return server.BlowUpdateRuntime53C160{
		IndexedDirection: blowIndexedDirectionNative53C160,
		EachInRect: func(rect types.Rectf, fn func(*server.Object) bool) {
			GetServer().S().Map.EachObjAndMissileInRect(rect, fn)
		},
		CanInteract: func(source, candidate *server.Object) bool {
			return GetServer().S().CanInteract(source, candidate, 0)
		},
		DirectionFloat: func(index int32) float32 {
			return *(*float32)(memmap.Ptr(uintptr(int64(0x5b6658) + int64(index)*4)))
		},
	}
}

func blowIndexedDirectionNative53C160(direction int16) image.Point {
	read := func(axis int64) int32 {
		return *(*int32)(memmap.Ptr(uintptr(int64(0x5b5e58) + int64(direction)*8 + axis*4)))
	}
	classify := func(value, threshold int32) int {
		if value > threshold {
			return 1
		}
		if value >= -threshold {
			return 0
		}
		return -1
	}
	// Each component reloads the original threshold. int32 negation retains the
	// machine's wrapping behavior for the signed minimum value.
	thresholdX := *(*int32)(memmap.Ptr(uintptr(0x5bf2cc)))
	x := classify(read(0), thresholdX)
	valueY := read(1)
	thresholdY := *(*int32)(memmap.Ptr(uintptr(0x5bf2cc)))
	y := classify(valueY, thresholdY)
	return image.Pt(x, y)
}
