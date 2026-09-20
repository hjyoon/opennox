package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestPushUpdate53B030NativeLayoutAndArguments(t *testing.T) {
	wantPosOffset := uintptr(56)
	wantUpdateDataOffset := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPosOffset = 60
		wantUpdateDataOffset = 872
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosOffset},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateDataOffset},
		{"PushUpdateData size", unsafe.Sizeof(PushUpdateData53B030{}), 12},
		{"PushUpdateData.Radius", unsafe.Offsetof(PushUpdateData53B030{}.Radius), 0},
		{"PushUpdateData.RadiusCopy", unsafe.Offsetof(PushUpdateData53B030{}.RadiusCopy), 4},
		{"PushUpdateData.Force", unsafe.Offsetof(PushUpdateData53B030{}.Force), 8},
	} {
		if field.got != field.want {
			t.Errorf("%s = %d, want %d", field.name, field.got, field.want)
		}
	}

	update := &PushUpdateData53B030{Radius: 81.25, RadiusCopy: 912.5, Force: -17.75}
	source := &Object{
		PosVec:     types.Ptf(123.5, -456.25),
		UpdateData: unsafe.Pointer(update),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 || uintptr(source.UpdateData) <= math.MaxUint32 {
			t.Fatalf("native source/update pointers = %p/%p, want both above ABI32 range", source, source.UpdateData)
		}
	}

	var calls int
	if !PushUpdate53B030(source, PushUpdateRuntime53B030{
		PushUnits: func(position types.Pointf, outer, inner, force float32) {
			calls++
			if position != source.PosVec || outer != update.Radius || inner != 0 || force != update.Force {
				t.Fatalf("push = (%+v, %g, %g, %g), want (%+v, %g, 0, %g)",
					position, outer, inner, force, source.PosVec, update.Radius, update.Force)
			}
		},
	}) {
		t.Fatal("native PushUpdate was not handled")
	}
	if calls != 1 {
		t.Fatalf("push calls = %d, want 1", calls)
	}
	if update.RadiusCopy != 912.5 {
		t.Fatalf("unused radius copy = %g, want 912.5", update.RadiusCopy)
	}
}

func TestPushUpdate53B030RejectsMissingState(t *testing.T) {
	if PushUpdate53B030(nil, PushUpdateRuntime53B030{}) {
		t.Fatal("nil source was accepted")
	}
	if PushUpdate53B030(&Object{}, PushUpdateRuntime53B030{PushUnits: func(types.Pointf, float32, float32, float32) {}}) {
		t.Fatal("nil update data was accepted")
	}
	data := new(PushUpdateData53B030)
	if PushUpdate53B030(&Object{UpdateData: unsafe.Pointer(data)}, PushUpdateRuntime53B030{}) {
		t.Fatal("missing push callback was accepted")
	}
}
