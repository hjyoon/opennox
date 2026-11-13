package noxrender

import (
	"math"
	"testing"
	"unsafe"
)

func TestRenderDataExplicitHighOwnership(t *testing.T) {
	ordinary, freeOrdinary := NewRenderData()
	high, freeHigh := NewRenderDataHigh()
	t.Cleanup(freeOrdinary)
	t.Cleanup(freeHigh)
	if high == ordinary || *high != *ordinary {
		t.Fatal("explicit render data lost the reset state or independent ownership")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(high.C()) <= math.MaxUint32 {
		t.Fatal("explicit render data does not exercise high native address bits")
	}
	before := *ordinary
	high.SetClip(!ordinary.Clip())
	if *ordinary != before {
		t.Fatal("render data owners shared mutable storage")
	}
}
