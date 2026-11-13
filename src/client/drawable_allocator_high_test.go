package client

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestDrawableAllocatorExplicitHighOwnership(t *testing.T) {
	var ordinary, high clientDrawables
	ordinary.Init(2)
	defer ordinary.Free()
	high.InitHigh(2)
	defer high.Free()
	a, b := ordinary.Alloc.NewObject(), high.Alloc.NewObject()
	if a == nil || b == nil || a == b {
		t.Fatal("distinct drawable owners did not receive distinct native records")
	}
	if runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(b)) <= uintptr(^uint32(0)) {
		t.Fatal("explicit drawable fixture did not receive high native address bits")
	}
	a.ObjFlags = 99
	b.ObjFlags = 77
	high.Alloc.FreeObjectFirst(b)
	reused := high.Alloc.NewObject()
	if reused != b || reused.ObjFlags != 0 || a.ObjFlags != 99 {
		t.Fatal("drawable reuse lost zero initialization or changed another owner's record")
	}
}
