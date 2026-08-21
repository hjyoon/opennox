package server

import (
	"testing"
	"unsafe"
)

func TestDirectionInit4F0490NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantDirection1 = 128
		wantDirection2 = 130
		wantInitData = 760
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection1},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"DirectionInitData size", unsafe.Sizeof(DirectionInitData{}), 8},
		{"DirectionInitData.X", unsafe.Offsetof(DirectionInitData{}.X), 0},
		{"DirectionInitData.Y", unsafe.Offsetof(DirectionInitData{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDirectionInit4F0490NativeUsesExactFields(t *testing.T) {
	initData := &DirectionInitData{X: 1, Y: -1}
	unit := &Object{
		Field29:    0x11111111,
		Direction1: 0xaaaa,
		Direction2: 0xbbbb,
		Field32:    0x22222222,
		InitData:   unsafe.Pointer(initData),
	}
	if got := DirectionInit4F0490(unit); got != 224 {
		t.Fatalf("result = %d, want 224", got)
	}
	if unit.Direction1 != 224 || unit.Direction2 != 224 {
		t.Fatalf("directions = %d/%d, want 224", unit.Direction1, unit.Direction2)
	}
	if unit.Field29 != 0x11111111 || unit.Field32 != 0x22222222 {
		t.Fatalf("adjacent fields changed to %#x/%#x", unit.Field29, unit.Field32)
	}
}

func TestDirectionInit4F0490NativeNilUnitFaultsOnInitDataLoad(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not preserve the original InitData-load fault")
		}
	}()
	DirectionInit4F0490(nil)
}

func TestDirectionInit4F0490NativeNilInitDataFaultsBeforeStores(t *testing.T) {
	unit := &Object{Direction1: 0xaaaa, Direction2: 0xbbbb}
	defer func() {
		if recover() == nil {
			t.Fatal("nil DirectionInitData did not fault")
		}
		if unit.Direction1 != 0xaaaa || unit.Direction2 != 0xbbbb {
			t.Fatalf("directions changed to %#x/%#x", unit.Direction1, unit.Direction2)
		}
	}()
	DirectionInit4F0490(unit)
}

func TestDirectionInit4F0490NativeRejectsUnsealedAdjacentTableDataBeforeStores(t *testing.T) {
	initData := &DirectionInitData{X: 5, Y: 0}
	unit := &Object{
		Direction1: 0xaaaa,
		Direction2: 0xbbbb,
		InitData:   unsafe.Pointer(initData),
	}
	defer func() {
		if recover() == nil {
			t.Fatal("out-of-range centered table index did not fault")
		}
		if unit.Direction1 != 0xaaaa || unit.Direction2 != 0xbbbb {
			t.Fatalf("directions changed to %#x/%#x", unit.Direction1, unit.Direction2)
		}
	}()
	DirectionInit4F0490(unit)
}
