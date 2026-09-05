package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestPositionDeltaNative4FEA70PreservesNativePointers(t *testing.T) {
	object := &Object{PosVec: types.Pointf{X: 10, Y: 20}}
	point := &types.Pointf{X: 14, Y: 25}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"object": uintptr(unsafe.Pointer(object)),
			"point":  uintptr(unsafe.Pointer(point)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := positionDeltaNative4FEA70(object, point); got != 1 {
		t.Fatalf("result = %d, want canonical one from Y threshold", got)
	}
	if object.PosVec != (types.Pointf{X: 10, Y: 20}) || *point != (types.Pointf{X: 14, Y: 25}) {
		t.Fatalf("inputs mutated to object=%+v point=%+v", object.PosVec, *point)
	}
	runtime.KeepAlive(object)
	runtime.KeepAlive(point)
}

func TestPositionDelta4FEA70ServerMethodUsesNativeAdapter(t *testing.T) {
	object := &Object{PosVec: types.Pointf{X: -5, Y: 1}}
	point := &types.Pointf{X: 0, Y: 1}
	if got := (*Server)(nil).PositionDelta4FEA70(object, point); got != 1 {
		t.Fatalf("server result = %d, want canonical one", got)
	}
}

func TestPositionDeltaNative4FEA70NilFaults(t *testing.T) {
	tests := []struct {
		name   string
		object *Object
		point  *types.Pointf
	}{
		{name: "nil point", object: &Object{}},
		{name: "nil object", point: &types.Pointf{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid native pointer did not fault")
				}
			}()
			positionDeltaNative4FEA70(tc.object, tc.point)
		})
	}
}

func TestPositionDeltaNativeLayout4FEA70(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantPosition := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantPosition = 60
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
