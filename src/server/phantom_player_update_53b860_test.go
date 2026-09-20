package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestPhantomPlayerUpdateDataNativeLayout53B860(t *testing.T) {
	wantSize := unsafe.Sizeof(uintptr(0)) + 8
	if got := unsafe.Offsetof(PhantomPlayerUpdateData{}.Target); got != 0 {
		t.Fatalf("Target offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(PhantomPlayerUpdateData{}.Position); got != unsafe.Sizeof(uintptr(0)) {
		t.Fatalf("Position offset = %d, want %d", got, unsafe.Sizeof(uintptr(0)))
	}
	if got := unsafe.Sizeof(PhantomPlayerUpdateData{}); got != wantSize {
		t.Fatalf("size = %d, want %d", got, wantSize)
	}

	target := new(Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
		t.Fatalf("target pointer = %p, want address above the PE32 range", target)
	}
	runtime.KeepAlive(target)
}

func TestPhantomPlayerUpdateNative53B860TracksTarget(t *testing.T) {
	target := &Object{
		PosVec:     types.Ptf(120, 250),
		Direction1: 200,
	}
	update := &PhantomPlayerUpdateData{
		Target:   target,
		Position: types.Ptf(100, 200),
	}
	source := &Object{
		NewPos:     types.Ptf(-1, -2),
		Direction2: 17,
		UpdateData: unsafe.Pointer(update),
	}
	deleted := false

	phantomPlayerUpdateNative53B860(source, func(*Object) { deleted = true })

	if deleted {
		t.Fatal("phantom was deleted inside the 400-unit radius")
	}
	if source.NewPos != update.Position {
		t.Fatalf("NewPos = %v, want %v", source.NewPos, update.Position)
	}
	if source.Direction2 != 72 {
		t.Fatalf("Direction2 = %d, want 72", source.Direction2)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(update)
}

func TestPhantomPlayerUpdateNative53B860KeepsExactBoundary(t *testing.T) {
	target := &Object{PosVec: types.Ptf(500, 200), Direction1: 255}
	update := &PhantomPlayerUpdateData{Target: target, Position: types.Ptf(100, 200)}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	deleted := false

	phantomPlayerUpdateNative53B860(source, func(*Object) { deleted = true })

	if deleted || source.NewPos != update.Position || source.Direction2 != 127 {
		t.Fatalf("deleted/NewPos/Direction2 = %t/%v/%d, want false/%v/127", deleted, source.NewPos, source.Direction2, update.Position)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(update.Target)) <= math.MaxUint32 {
		t.Fatalf("target pointer = %p, want address above the PE32 range", update.Target)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(update)
}

func TestPhantomPlayerUpdateNative53B860DeletesOutsideRadius(t *testing.T) {
	target := &Object{PosVec: types.Ptf(501, 200), Direction1: 90}
	update := &PhantomPlayerUpdateData{Target: target, Position: types.Ptf(100, 200)}
	source := &Object{
		NewPos:     types.Ptf(7, 8),
		Direction2: 33,
		UpdateData: unsafe.Pointer(update),
	}
	var deleted *Object

	phantomPlayerUpdateNative53B860(source, func(obj *Object) { deleted = obj })

	if deleted != source {
		t.Fatalf("deleted = %p, want %p", deleted, source)
	}
	if source.NewPos != (types.Ptf(7, 8)) || source.Direction2 != 33 {
		t.Fatalf("state changed to NewPos=%v Direction2=%d", source.NewPos, source.Direction2)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(update)
}

func TestPhantomPlayerUpdateNative53B860DeletesNaNDistance(t *testing.T) {
	target := &Object{PosVec: types.Ptf(float32(math.NaN()), 200)}
	update := &PhantomPlayerUpdateData{Target: target, Position: types.Ptf(100, 200)}
	source := &Object{
		NewPos:     types.Ptf(7, 8),
		Direction2: 33,
		UpdateData: unsafe.Pointer(update),
	}
	var deleted *Object

	phantomPlayerUpdateNative53B860(source, func(obj *Object) { deleted = obj })

	if deleted != source {
		t.Fatalf("deleted = %p, want %p", deleted, source)
	}
	if source.NewPos != (types.Ptf(7, 8)) || source.Direction2 != 33 {
		t.Fatalf("state changed to NewPos=%v Direction2=%d", source.NewPos, source.Direction2)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(update)
}

func TestPhantomPlayerUpdateNative53B860DeletesInvalidTrackingData(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update *PhantomPlayerUpdateData
	}{
		{name: "missing update data"},
		{name: "missing target", update: new(PhantomPlayerUpdateData)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := new(Object)
			if tc.update != nil {
				source.UpdateData = unsafe.Pointer(tc.update)
			}
			var deleted *Object
			phantomPlayerUpdateNative53B860(source, func(obj *Object) { deleted = obj })
			if deleted != source {
				t.Fatalf("deleted = %p, want %p", deleted, source)
			}
			runtime.KeepAlive(tc.update)
		})
	}
}

func TestPhantomPlayerUpdate53B860NilSource(t *testing.T) {
	called := false
	new(Server).PhantomPlayerUpdate53B860(nil, PhantomPlayerUpdateRuntime53B860{
		DelayedDelete: func(*Object) { called = true },
	})
	if called {
		t.Fatal("nil source reached delayed delete")
	}
}
