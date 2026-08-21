package server

import (
	"testing"
	"unsafe"
)

func TestSkullInit4F0450NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantInitData := uintptr(692)
	wantUpdateData := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantDirection1 = 128
		wantDirection2 = 130
		wantInitData = 760
		wantUpdateData = 872
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
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"DirectionInitData size", unsafe.Sizeof(DirectionInitData{}), 8},
		{"DirectionInitData.X", unsafe.Offsetof(DirectionInitData{}.X), 0},
		{"DirectionInitData.Y", unsafe.Offsetof(DirectionInitData{}.Y), 4},
		{"SkullUpdateData size", unsafe.Sizeof(SkullUpdateData{}), 52},
		{"SkullUpdateData.ScanDelay", unsafe.Offsetof(SkullUpdateData{}.ScanDelay), 0},
		{"SkullUpdateData.FireDelay", unsafe.Offsetof(SkullUpdateData{}.FireDelay), 4},
		{"SkullUpdateData.TargetReady", unsafe.Offsetof(SkullUpdateData{}.TargetReady), 8},
		{"SkullUpdateData.ProjectileType", unsafe.Offsetof(SkullUpdateData{}.ProjectileType), 12},
		{"SkullUpdateData.ProjectileName", unsafe.Offsetof(SkullUpdateData{}.ProjectileName), 16},
		{"SkullUpdateData.Enabled", unsafe.Offsetof(SkullUpdateData{}.Enabled), 48},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDirectionToAngleNative509E00MatchesSealedTable(t *testing.T) {
	want := [3][3]uint32{
		{160, 192, 224},
		{128, 0, 0},
		{96, 64, 32},
	}
	for y := int32(-1); y <= 1; y++ {
		for x := int32(-1); x <= 1; x++ {
			data := &DirectionInitData{X: x, Y: y}
			if got := directionToAngleNative509E00(data); got != want[y+1][x+1] {
				t.Errorf("angle(%d,%d) = %d, want %d", x, y, got, want[y+1][x+1])
			}
		}
	}
}

func TestDirectionToAngleNative509E00RejectsUnsealedAdjacentData(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("out-of-range centered table index did not fault")
		}
	}()
	directionToAngleNative509E00(&DirectionInitData{X: 5, Y: 0})
}

func TestSkullInitNative4F0450UsesCachedPointersAndExactFields(t *testing.T) {
	initData := &DirectionInitData{X: 1, Y: -1}
	replacementInit := &DirectionInitData{X: -1, Y: 1}
	update := &SkullUpdateData{
		ScanDelay: 0x11111111, FireDelay: 0x22222222,
		TargetReady: 0x33, ProjectileType: 0x44444444, Enabled: 0x55,
		Field9: [3]byte{0x61, 0x62, 0x63}, Field49: [3]byte{0x71, 0x72, 0x73},
	}
	copy(update.ProjectileName[:], "MercArcherArrow")
	replacementUpdate := &SkullUpdateData{ProjectileType: 0x66666666}
	unit := &Object{
		Direction1: 0xaaaa, Direction2: 0xbbbb,
		InitData: unsafe.Pointer(initData), UpdateData: unsafe.Pointer(update),
	}
	lookupCalls := 0
	got := skullInitNative4F0450(unit, skullInitNativeDeps4F0450{
		lookupType: func(name string) int32 {
			lookupCalls++
			if name != "MercArcherArrow" {
				t.Fatalf("projectile name = %q", name)
			}
			unit.InitData = unsafe.Pointer(replacementInit)
			unit.UpdateData = unsafe.Pointer(replacementUpdate)
			return int32(-2147483647)
		},
	})
	if got != int32(-2147483647) || lookupCalls != 1 {
		t.Fatalf("result/calls = %d/%d", got, lookupCalls)
	}
	if unit.Direction1 != 224 || unit.Direction2 != 224 {
		t.Fatalf("directions = %d/%d, want 224", unit.Direction1, unit.Direction2)
	}
	if update.ProjectileType != 0x80000001 {
		t.Fatalf("cached update projectile type = %#x", update.ProjectileType)
	}
	if replacementUpdate.ProjectileType != 0x66666666 {
		t.Fatalf("replacement update changed to %#x", replacementUpdate.ProjectileType)
	}
	if update.ScanDelay != 0x11111111 || update.FireDelay != 0x22222222 ||
		update.TargetReady != 0x33 || update.Enabled != 0x55 ||
		update.Field9 != [3]byte{0x61, 0x62, 0x63} ||
		update.Field49 != [3]byte{0x71, 0x72, 0x73} {
		t.Fatalf("adjacent update fields changed: %+v", update)
	}
}

func TestServerSkullInit4F0450ResolvesObjectType(t *testing.T) {
	var s Server
	s.Types.byID = map[string]*ObjectType{
		"mercarcherarrow": {ind: 321, id: "MercArcherArrow"},
	}
	initData := &DirectionInitData{X: -1, Y: 0}
	update := &SkullUpdateData{}
	copy(update.ProjectileName[:], "MERCARCHERARROW")
	unit := &Object{InitData: unsafe.Pointer(initData), UpdateData: unsafe.Pointer(update)}
	if got := s.SkullInit4F0450(unit); got != 321 {
		t.Fatalf("result = %d, want 321", got)
	}
	if update.ProjectileType != 321 || unit.Direction1 != 128 || unit.Direction2 != 128 {
		t.Fatalf("type/directions = %d/%d/%d", update.ProjectileType, unit.Direction1, unit.Direction2)
	}
}

func TestSkullInitNative4F0450NilUnitFaultsBeforeLookup(t *testing.T) {
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object did not fault")
		}
		if called {
			t.Fatal("lookup ran before nil Object fault")
		}
	}()
	skullInitNative4F0450(nil, skullInitNativeDeps4F0450{
		lookupType: func(string) int32 { called = true; return 0 },
	})
}

func TestSkullInitNative4F0450NilInitDataFaultsBeforeStores(t *testing.T) {
	update := &SkullUpdateData{}
	unit := &Object{Direction1: 0xaaaa, Direction2: 0xbbbb, UpdateData: unsafe.Pointer(update)}
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil DirectionInitData did not fault")
		}
		if called || unit.Direction1 != 0xaaaa || unit.Direction2 != 0xbbbb {
			t.Fatalf("post-fault state = called:%t directions:%#x/%#x", called, unit.Direction1, unit.Direction2)
		}
	}()
	skullInitNative4F0450(unit, skullInitNativeDeps4F0450{
		lookupType: func(string) int32 { called = true; return 0 },
	})
}

func TestSkullInitNative4F0450NilUpdateDataFaultsAfterDirections(t *testing.T) {
	initData := &DirectionInitData{X: 0, Y: 1}
	unit := &Object{Direction1: 0xaaaa, Direction2: 0xbbbb, InitData: unsafe.Pointer(initData)}
	called := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil SkullUpdateData did not fault")
		}
		if called {
			t.Fatal("type lookup ran with nil SkullUpdateData")
		}
		if unit.Direction1 != 64 || unit.Direction2 != 64 {
			t.Fatalf("directions = %d/%d, want 64", unit.Direction1, unit.Direction2)
		}
	}()
	skullInitNative4F0450(unit, skullInitNativeDeps4F0450{
		lookupType: func(string) int32 { called = true; return 0 },
	})
}
