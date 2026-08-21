package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestBoulderInit4F0420NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantPosVec := uintptr(56)
	wantPos39 := uintptr(156)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantPosVec = 60
		wantPos39 = 160
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosVec},
		{"Object.PosVec.Y", unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.Y), wantPosVec + 4},
		{"Object.Pos39", unsafe.Offsetof(Object{}.Pos39), wantPos39},
		{"Object.Pos39.Y", unsafe.Offsetof(Object{}.Pos39) + unsafe.Offsetof(types.Pointf{}.Y), wantPos39 + 4},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBoulderInit4F0420NativeCopiesExactBitsAndReturnsUnit(t *testing.T) {
	const (
		xBits = uint32(0x7fa12345)
		yBits = uint32(0x80000000)
	)
	unit := &Object{
		PosVec:  types.Pointf{X: math.Float32frombits(xBits), Y: math.Float32frombits(yBits)},
		Field38: 0xa5a5a5a5,
		Pos39:   types.Pointf{X: math.Float32frombits(0x11111111), Y: math.Float32frombits(0x22222222)},
		Field41: 0x5a5a5a5a,
	}

	got := BoulderInit4F0420(unit)
	if got != unit {
		t.Fatalf("return = %p, want entry unit %p", got, unit)
	}
	if gotX, gotY := math.Float32bits(unit.Pos39.X), math.Float32bits(unit.Pos39.Y); gotX != xBits || gotY != yBits {
		t.Fatalf("Pos39 bits = %#x/%#x, want %#x/%#x", gotX, gotY, xBits, yBits)
	}
	if gotX, gotY := math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y); gotX != xBits || gotY != yBits {
		t.Fatalf("PosVec changed to %#x/%#x", gotX, gotY)
	}
	if unit.Field38 != 0xa5a5a5a5 || unit.Field41 != 0x5a5a5a5a {
		t.Fatalf("adjacent fields changed to %#x/%#x", unit.Field38, unit.Field41)
	}
}

func TestBoulderInit4F0420NativeNilUnitFaultsOnFirstLoad(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original PosVec.X-load fault")
		}
	}()
	BoulderInit4F0420(nil)
}
