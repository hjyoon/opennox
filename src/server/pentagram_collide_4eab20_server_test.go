package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestPentagramCollide4EAB20NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateData := uintptr(748)
	wantPentagramSize := uintptr(24)
	wantDestination := uintptr(12)
	wantAnimationStep := uintptr(20)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateData = 872
		wantPentagramSize = 32
		wantDestination = 16
		wantAnimationStep = 28
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PentagramUpdateDataPrefix size", unsafe.Sizeof(PentagramUpdateDataPrefix{}), 8},
		{"PentagramUpdateDataPrefix.Triggered", unsafe.Offsetof(PentagramUpdateDataPrefix{}.Triggered), 4},
		{"PentagramUpdateData size", unsafe.Sizeof(PentagramUpdateData{}), wantPentagramSize},
		{"PentagramUpdateData.Triggered", unsafe.Offsetof(PentagramUpdateData{}.Triggered), 4},
		{"PentagramUpdateData.AnimationFrame", unsafe.Offsetof(PentagramUpdateData{}.AnimationFrame), 8},
		{"PentagramUpdateData.Destination", unsafe.Offsetof(PentagramUpdateData{}.Destination), wantDestination},
		{"PentagramUpdateData.AnimationStep", unsafe.Offsetof(PentagramUpdateData{}.AnimationStep), wantAnimationStep},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPentagramCollideNative4EAB20UsesNativeUpdatePointer(t *testing.T) {
	data := &PentagramUpdateDataPrefix{
		Reserved0: [4]byte{1, 2, 3, 4},
		Triggered: 0xaabbccdd,
	}
	source := &Object{UpdateData: unsafe.Pointer(data), Field188: 0x11223344}
	target := &Object{Field188: 0x55667788}
	collision := &types.Pointf{X: 3, Y: 4}
	got := pentagramCollideNative4EAB20(source, target, collision)
	if got != source {
		t.Fatalf("return = %p, want %p", got, source)
	}
	if data.Reserved0 != [4]byte{1, 2, 3, 4} || data.Triggered != 1 {
		t.Fatalf("data = %+v", data)
	}
	if source.Field188 != 0x11223344 || target.Field188 != 0x55667788 || collision.X != 3 || collision.Y != 4 {
		t.Fatalf("state = source %#x target %#x collision %+v", source.Field188, target.Field188, collision)
	}
}

func TestPentagramCollide4EAB20ServerBinding(t *testing.T) {
	s := &Server{}
	data := &PentagramUpdateDataPrefix{Triggered: 9}
	source := &Object{UpdateData: unsafe.Pointer(data)}
	target := &Object{Field188: 0x89abcdef}
	collision := &types.Pointf{X: 5, Y: 6}
	s.PentagramCollide4EAB20(source, target, collision)
	if data.Triggered != 1 || target.Field188 != 0x89abcdef || collision.X != 5 || collision.Y != 6 {
		t.Fatalf("data/target/collision = %#x/%#x/%+v", data.Triggered, target.Field188, collision)
	}
}

func TestPentagramCollideNative4EAB20NilUpdateDataFaults(t *testing.T) {
	source := &Object{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update data did not fault")
		}
	}()
	pentagramCollideNative4EAB20(source, nil, nil)
}
