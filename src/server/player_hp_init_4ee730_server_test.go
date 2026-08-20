package server

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerHPInit4EE730NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantHealth := uintptr(556)
	wantUpdate := uintptr(748)
	wantPlayerUpdateSize := uintptr(556)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantHealth = 616
		wantUpdate = 872
		wantPlayerUpdateSize = 640
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantPlayerUpdateSize},
		{"PlayerUpdateData.HealthSamples", unsafe.Offsetof(PlayerUpdateData{}.HealthSamples), 12},
		{"PlayerUpdateData.HealthSamples size", unsafe.Sizeof(PlayerUpdateData{}.HealthSamples), 64},
		{"PlayerUpdateData.HealthSampleCur", unsafe.Offsetof(PlayerUpdateData{}.HealthSampleCur), 76},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerHPInit4EE730NativeBindsPlayerUpdateAndHealth(t *testing.T) {
	health := &HealthData{Cur: 0x9abc}
	update := &PlayerUpdateData{HealthSampleCur: 0xdef0}
	for i := range update.HealthSamples {
		update.HealthSamples[i] = uint16(i)
	}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
	}

	PlayerHPInit4EE730(unit)
	for i, got := range update.HealthSamples {
		if got != health.Cur {
			t.Fatalf("HealthSamples[%d] = %#x, want %#x", i, got, health.Cur)
		}
	}
	if update.HealthSampleCur != health.Cur {
		t.Fatalf("HealthSampleCur = %#x, want %#x", update.HealthSampleCur, health.Cur)
	}
}

func TestPlayerHPInit4EE730NativeEntryGates(t *testing.T) {
	PlayerHPInit4EE730(nil)
	PlayerHPInit4EE730(&Object{ObjClass: object.ClassMonster})

	update := &PlayerUpdateData{HealthSampleCur: 0x1234}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	PlayerHPInit4EE730(unit)
	if update.HealthSampleCur != 0x1234 {
		t.Fatalf("nil HealthData changed Player update = %#x", update.HealthSampleCur)
	}
}

func TestPlayerHPInit4EE730NativeNilUpdateFault(t *testing.T) {
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: &HealthData{Cur: 77},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player UpdateData did not preserve the original store fault")
		}
	}()
	PlayerHPInit4EE730(unit)
}
