package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestUnitHPSetOnMax4EE6F0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantHealth := uintptr(556)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantHealth = 616
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Field2", unsafe.Offsetof(HealthData{}.Field2), 2},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
		{"HP width", unsafe.Sizeof(uint16(0)), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUnitHPSetOnMax4EE6F0NativeBindsLiveHealthAndClass(t *testing.T) {
	entry := &HealthData{Cur: 3, Field2: 4, Max: 500}
	live := &HealthData{Cur: 77, Field2: 8, Max: 900}
	unit := &Object{HealthData: entry}
	var events []string
	unitHPSetOnMaxNative4EE6F0(unit, unitHPSetOnMaxNativeDeps4EE6F0{
		setHP: func(got *Object, value uint16) {
			events = append(events, "set")
			if got != unit || value != 500 {
				t.Fatalf("SetHP(%p, %d), want (%p, 500)", got, value, unit)
			}
			got.HealthData.Cur = value
			got.HealthData = live
			got.ObjClass = object.ClassMonster
		},
		informOwner: func(got *Object) {
			events = append(events, "inform")
			if got != unit {
				t.Fatalf("inform owner object = %p, want %p", got, unit)
			}
		},
	})
	if !reflect.DeepEqual(events, []string{"set", "inform"}) {
		t.Fatalf("events = %q", events)
	}
	if entry.Cur != 500 || entry.Field2 != 4 {
		t.Fatalf("entry health = %+v", *entry)
	}
	if live.Cur != 77 || live.Field2 != 77 {
		t.Fatalf("live health = %+v", *live)
	}
}

func TestUnitHPSetOnMax4EE6F0ServerMethodUsesRuntimeSetter(t *testing.T) {
	health := &HealthData{Cur: 6, Field2: 7, Max: 321}
	unit := &Object{HealthData: health}
	server := &Server{}
	server.UnitHPSetOnMax4EE6F0(unit, UnitHPSetOnMaxRuntime4EE6F0{
		SetHP: func(got *Object, value uint16) {
			if got != unit || value != 321 {
				t.Fatalf("SetHP(%p, %d), want (%p, 321)", got, value, unit)
			}
			got.HealthData.Cur = value
		},
	})
	if health.Cur != 321 || health.Field2 != 321 || health.Max != 321 {
		t.Fatalf("health = %+v", *health)
	}
}

func TestUnitHPSetOnMax4EE6F0NativeEntryGates(t *testing.T) {
	called := false
	deps := unitHPSetOnMaxNativeDeps4EE6F0{
		setHP:       func(*Object, uint16) { called = true },
		informOwner: func(*Object) { called = true },
	}
	unitHPSetOnMaxNative4EE6F0(nil, deps)
	unitHPSetOnMaxNative4EE6F0(&Object{}, deps)
	if called {
		t.Fatal("entry gates invoked a callback")
	}
}
