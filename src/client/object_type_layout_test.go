package client

import (
	"testing"
	"unsafe"
)

func TestObjectTypeNativeLayout(t *testing.T) {
	typ := ObjectType{}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want := struct {
		lightColor, drawFunc, drawData, clientUpdate, next, size uintptr
	}{48, 88, 92, 100, 108, 128}
	if ptrSize == 8 {
		want = struct {
			lightColor, drawFunc, drawData, clientUpdate, next, size uintptr
		}{64, 120, 128, 144, 160, 184}
	}
	got := struct {
		lightColor, drawFunc, drawData, clientUpdate, next, size uintptr
	}{
		unsafe.Offsetof(typ.LightColor),
		unsafe.Offsetof(typ.DrawFunc),
		unsafe.Offsetof(typ.DrawData),
		unsafe.Offsetof(typ.ClientUpdate),
		unsafe.Offsetof(typ.ObjNext),
		unsafe.Sizeof(typ),
	}
	if got != want {
		t.Fatalf("ObjectType native layout mismatch: got %+v, want %+v", got, want)
	}
}
