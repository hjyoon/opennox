package handles

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestHighOpaqueFixtureBank(t *testing.T) {
	InitHigh()
	defer Release()
	a, b := NewPtr(), NewPtr()
	if a == nil || b == nil || a == b || !IsValid(uintptr(a)) || !IsValid(uintptr(b)) {
		t.Fatal("opaque fixture bank lost distinct valid native handles")
	}
	if runtime.GOOS == "windows" && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(a) <= uintptr(^uint32(0)) {
		t.Fatalf("opaque fixture handle is below 4 GiB: %p", a)
	}
	AssertValidPtr(a)
	AssertValidPtr(b)
	Release()
	if IsValid(uintptr(a)) || IsValid(uintptr(b)) {
		t.Fatal("released high opaque handles remained valid")
	}
}
