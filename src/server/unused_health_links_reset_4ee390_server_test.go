package server

import (
	"strings"
	"testing"
	"unsafe"
)

func TestUnusedHealthLinksReset4EE390NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantHealthOffset := uintptr(556)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantHealthOffset = 616
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealthOffset},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData current", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData maximum", unsafe.Offsetof(HealthData{}.Max), 4},
		{"HealthData next ABI32 link", unsafe.Offsetof(HealthData{}.field8), 8},
		{"HealthData previous ABI32 link", unsafe.Offsetof(HealthData{}.field12), 12},
		{"HealthData field16", unsafe.Offsetof(HealthData{}.Field16), 16},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnusedHealthLinksResetNative4EE390ClearsLinksAndReturnsHealth(t *testing.T) {
	health := &HealthData{
		Cur:     11,
		Field2:  12,
		Max:     13,
		field6:  14,
		field8:  0x11223344,
		field12: 0x55667788,
		Field16: 0x99aabbcc,
	}
	obj := &Object{HealthData: health}

	if got := unusedHealthLinksResetNative4EE390(obj); got != health {
		t.Fatalf("result = %p, want HealthData %p", got, health)
	}
	if health.field8 != 0 || health.field12 != 0 {
		t.Fatalf("links = %#x/%#x, want zero", health.field8, health.field12)
	}
	if health.Cur != 11 || health.Field2 != 12 || health.Max != 13 || health.field6 != 14 || health.Field16 != 0x99aabbcc {
		t.Fatalf("non-link health fields changed: %+v", *health)
	}
	if obj.HealthData != health {
		t.Fatalf("object HealthData = %p, want %p", obj.HealthData, health)
	}
}

func TestUnusedHealthLinksReset4EE390ServerBindingNullObject(t *testing.T) {
	if got := new(Server).UnusedHealthLinksReset4EE390(nil); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
}

func TestUnusedHealthLinksResetNative4EE390PreservesNullHealthFault(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("expected absolute-null-write panic")
		}
		message, ok := got.(string)
		if !ok || !strings.Contains(message, "0x0000000C") {
			t.Fatalf("panic = %#v, want original absolute null address", got)
		}
	}()
	unusedHealthLinksResetNative4EE390(&Object{})
}
