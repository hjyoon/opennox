package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func pixieTeleportBits4FD050(object *Object) [6]uint32 {
	return [6]uint32{
		math.Float32bits(object.NewPos.X),
		math.Float32bits(object.NewPos.Y),
		math.Float32bits(object.PosVec.X),
		math.Float32bits(object.PosVec.Y),
		math.Float32bits(object.PrevPos.X),
		math.Float32bits(object.PrevPos.Y),
	}
}

func TestPixieTeleportNative4FD050PreservesNativePointersAndRawBits(t *testing.T) {
	const (
		xBits = uint32(0x7fa12345)
		yBits = uint32(0x80000000)
	)
	owner := &Object{PosVec: types.Pointf{
		X: math.Float32frombits(xBits),
		Y: math.Float32frombits(yBits),
	}}
	pixie := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"owner": uintptr(unsafe.Pointer(owner)),
			"pixie": uintptr(unsafe.Pointer(pixie)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	moveCalls := 0
	pixieTeleportNative4FD050(pixie, owner, pixieTeleportNativeDeps4FD050{
		moveUpdate: func(got *Object) {
			moveCalls++
			if got != pixie {
				t.Fatalf("move object = %p, want %p", got, pixie)
			}
			want := [6]uint32{xBits, yBits, xBits, yBits, xBits, yBits}
			if gotBits := pixieTeleportBits4FD050(got); gotBits != want {
				t.Fatalf("coordinates at move = %#v, want %#v", gotBits, want)
			}
		},
	})
	if moveCalls != 1 {
		t.Fatalf("move calls = %d, want 1", moveCalls)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(pixie)
}

func TestPixieTeleportNative4FD050NilOwnerFaultsBeforeStores(t *testing.T) {
	pixie := &Object{
		NewPos:  types.Pointf{X: 1, Y: 2},
		PosVec:  types.Pointf{X: 3, Y: 4},
		PrevPos: types.Pointf{X: 5, Y: 6},
	}
	want := pixieTeleportBits4FD050(pixie)
	moveCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault")
		}
		if got := pixieTeleportBits4FD050(pixie); got != want {
			t.Fatalf("coordinates after owner fault = %#v, want unchanged %#v", got, want)
		}
		if moveCalls != 0 {
			t.Fatalf("nil owner invoked move update %d times", moveCalls)
		}
	}()
	pixieTeleportNative4FD050(pixie, nil, pixieTeleportNativeDeps4FD050{
		moveUpdate: func(*Object) { moveCalls++ },
	})
}

func TestPixieTeleportNative4FD050NilPixieFaultsBeforeMove(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{X: 7, Y: 8}}
	moveCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil Pixie did not fault")
		}
		if moveCalls != 0 {
			t.Fatalf("nil Pixie invoked move update %d times", moveCalls)
		}
	}()
	pixieTeleportNative4FD050(nil, owner, pixieTeleportNativeDeps4FD050{
		moveUpdate: func(*Object) { moveCalls++ },
	})
}

func TestPixieTeleport4FD050ServerMethodUsesNativeAdapter(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{
		X: math.Float32frombits(0x80000000),
		Y: math.Float32frombits(0x7fcabcde),
	}}
	pixie := &Object{}
	called := false
	new(Server).PixieTeleport4FD050(pixie, owner, func(got *Object) {
		called = true
		if got != pixie {
			t.Fatalf("move object = %p, want %p", got, pixie)
		}
	})
	if !called {
		t.Fatal("server method did not invoke move update")
	}
	want := [6]uint32{
		math.Float32bits(owner.PosVec.X), math.Float32bits(owner.PosVec.Y),
		math.Float32bits(owner.PosVec.X), math.Float32bits(owner.PosVec.Y),
		math.Float32bits(owner.PosVec.X), math.Float32bits(owner.PosVec.Y),
	}
	if got := pixieTeleportBits4FD050(pixie); got != want {
		t.Fatalf("coordinates = %#v, want %#v", got, want)
	}
}

func TestPixieTeleport4FD050NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantPos := uintptr(56)
	wantNew := uintptr(64)
	wantPrev := uintptr(72)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantPos = 60
		wantNew = 68
		wantPrev = 76
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), wantNew},
		{"Object.PrevPos", unsafe.Offsetof(Object{}.PrevPos), wantPrev},
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
