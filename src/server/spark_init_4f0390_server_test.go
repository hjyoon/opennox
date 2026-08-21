package server

import (
	"testing"
	"unsafe"
)

func TestSparkInit4F0390NativeLayout(t *testing.T) {
	wantUpdateData := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdateData = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"SparkUpdateData size", unsafe.Sizeof(SparkUpdateData{}), 16},
		{"SparkUpdateData.LifetimeInitial", unsafe.Offsetof(SparkUpdateData{}.LifetimeInitial), 0},
		{"SparkUpdateData.LifetimeRemaining", unsafe.Offsetof(SparkUpdateData{}.LifetimeRemaining), 4},
		{"SparkUpdateData.Field8", unsafe.Offsetof(SparkUpdateData{}.Field8), 8},
		{"SparkUpdateData.Kind", unsafe.Offsetof(SparkUpdateData{}.Kind), 12},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSparkInit4F0390NativeStoresAndReturnsCachedUpdate(t *testing.T) {
	update := &SparkUpdateData{
		LifetimeInitial:   0x11111111,
		LifetimeRemaining: 0x22222222,
		Field8:            0xa5a5a5a5,
		Kind:              0x5a5a5a5a,
	}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	got := SparkInit4F0390(unit)
	if got != update {
		t.Fatalf("return = %p, want %p", got, update)
	}
	if update.LifetimeInitial != sparkInitLifetime4F0390 || update.LifetimeRemaining != sparkInitLifetime4F0390 {
		t.Fatalf("lifetimes = %d/%d", update.LifetimeInitial, update.LifetimeRemaining)
	}
	if update.Field8 != 0xa5a5a5a5 || update.Kind != 0x5a5a5a5a {
		t.Fatalf("unrelated fields = %#x/%#x", update.Field8, update.Kind)
	}
	if unit.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("Object.UpdateData changed to %p", unit.UpdateData)
	}
}

func TestSparkInit4F0390NativeNilUnitFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original UpdateData-load fault")
		}
	}()
	SparkInit4F0390(nil)
}

func TestSparkInit4F0390NativeNilUpdateFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not preserve the original first-store fault")
		}
	}()
	SparkInit4F0390(&Object{})
}
