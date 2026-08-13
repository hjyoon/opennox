package server

import (
	"testing"
	"unsafe"
)

func objectPixieTargetClearCachedDeps4E81D0(typeID uint32) objectPixieTargetClearNativeDeps4E81D0 {
	return objectPixieTargetClearNativeDeps4E81D0{
		loadPixieTypeID:  func() uint32 { return typeID },
		lookupObjectType: func(string) uint32 { panic("unexpected Pixie lookup") },
		storePixieTypeID: func(uint32) { panic("unexpected Pixie cache store") },
	}
}

func TestObjectPixieTargetClearNative4E81D0UsesNamedUpdateFields(t *testing.T) {
	owner := &Object{TypeInd: 3}
	target := &Object{TypeInd: 4}
	update := PixieUpdateData{
		Owner:                 owner,
		Target:                target,
		Field8:                0x11223344,
		SpellID:               -17,
		Field16:               0x55667788,
		Deadline:              0x99aabbcc,
		LastOwnerVisibleFrame: 0xddeeff00,
	}
	obj := &Object{TypeInd: 37, UpdateData: unsafe.Pointer(&update)}
	got := objectPixieTargetClearNative4E81D0(obj, objectPixieTargetClearCachedDeps4E81D0(37))
	if !got.returnsUpdate || got.updateData != &update || update.Target != nil {
		t.Fatalf("result/target = (%#v, %p), want matched nil target", got, update.Target)
	}
	if update.Owner != owner || update.Field8 != 0x11223344 || update.SpellID != -17 ||
		update.Field16 != 0x55667788 || update.Deadline != 0x99aabbcc || update.LastOwnerVisibleFrame != 0xddeeff00 {
		t.Fatalf("adjacent update fields changed: %#v", update)
	}
}

func TestObjectPixieTargetClearNative4E81D0NilAndMismatchSkipUpdate(t *testing.T) {
	if got := objectPixieTargetClearNative4E81D0(nil, objectPixieTargetClearCachedDeps4E81D0(37)); got.returnsUpdate || got.typeID != 37 {
		t.Fatalf("nil result = %#v, want scalar 37", got)
	}
	obj := &Object{TypeInd: 36}
	if got := objectPixieTargetClearNative4E81D0(obj, objectPixieTargetClearCachedDeps4E81D0(37)); got.returnsUpdate || got.typeID != 37 {
		t.Fatalf("mismatch result = %#v, want scalar 37", got)
	}
}

func TestObjectPixieTargetClearNative4E81D0NilUpdateFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("matching object with nil update data did not fault")
		}
	}()
	objectPixieTargetClearNative4E81D0(&Object{TypeInd: 37}, objectPixieTargetClearCachedDeps4E81D0(37))
}

func TestServerClearPixieTarget4E81D0BindsActualTypeCache(t *testing.T) {
	s := &Server{}
	s.Types.byID = map[string]*ObjectType{"pixie": {ind: 37}}
	update := PixieUpdateData{Target: &Object{TypeInd: 99}}
	obj := &Object{TypeInd: 37, UpdateData: unsafe.Pointer(&update)}
	s.ClearPixieTarget4E81D0(obj)
	if s.Types.fast.pixie != 37 || update.Target != nil {
		t.Fatalf("cache/target = (%d, %p), want (37, nil)", s.Types.fast.pixie, update.Target)
	}
}

func TestPixieUpdateDataNativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantSize := uintptr(28)
	wantTarget := uintptr(4)
	wantDeadline := uintptr(20)
	wantLast := uintptr(24)
	if ptrSize == 8 {
		wantSize = 40
		wantTarget = 8
		wantDeadline = 28
		wantLast = 32
	}
	if got := unsafe.Sizeof(PixieUpdateData{}); got != wantSize {
		t.Fatalf("PixieUpdateData size = %d, want %d", got, wantSize)
	}
	if got := unsafe.Offsetof(PixieUpdateData{}.Target); got != wantTarget {
		t.Fatalf("Target offset = %d, want %d", got, wantTarget)
	}
	if got := unsafe.Offsetof(PixieUpdateData{}.Deadline); got != wantDeadline {
		t.Fatalf("Deadline offset = %d, want %d", got, wantDeadline)
	}
	if got := unsafe.Offsetof(PixieUpdateData{}.LastOwnerVisibleFrame); got != wantLast {
		t.Fatalf("LastOwnerVisibleFrame offset = %d, want %d", got, wantLast)
	}
}
