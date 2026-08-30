package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func TestUnitGiveXPNative4EF270BindsFieldsAndServices(t *testing.T) {
	player := &Player{ProtUnitExperience: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{Experience: 100, UpdateData: unsafe.Pointer(update)}
	var events []string
	got := unitGiveXPNative4EF270(unit, 200, unitGiveXPNativeDeps4EF270{
		protectExperience: func(token uint32, award float32) {
			events = append(events, "protect")
			if token != 0x89abcdef || math.Float32bits(award) != 0x40000000 {
				t.Fatalf("protection args = %08x/%08x", token, math.Float32bits(award))
			}
			if unit.Experience != 102 {
				t.Fatalf("experience at protection = %v, want 102", unit.Experience)
			}
		},
		reportExperience: func(got *Object) {
			events = append(events, "report")
			if got != unit || unit.Experience != 102 {
				t.Fatalf("report unit/state = %p/%v", got, unit.Experience)
			}
		},
		syncLevel: func(got *Object) {
			events = append(events, "sync")
			if got != unit {
				t.Fatalf("sync unit = %p, want %p", got, unit)
			}
			unit.Experience = 777
		},
	})
	if math.Float64bits(got) != math.Float64bits(2) {
		t.Fatalf("award = %v, want 2", got)
	}
	if len(events) != 3 || events[0] != "protect" || events[1] != "report" || events[2] != "sync" {
		t.Fatalf("service order = %q, want protect/report/sync", events)
	}
	if unit.Experience != 777 {
		t.Fatalf("sync mutation was lost: %v", unit.Experience)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestUnitGiveXPNative4EF270EarlyPathSkipsPointerServices(t *testing.T) {
	marker := byte(1)
	unit := &Object{Experience: 200, UpdateData: unsafe.Pointer(&marker)}
	got := unitGiveXPNative4EF270(unit, 200, unitGiveXPNativeDeps4EF270{
		protectExperience: func(uint32, float32) { t.Fatal("protection called") },
		reportExperience:  func(*Object) { t.Fatal("report called") },
		syncLevel:         func(*Object) { t.Fatal("sync called") },
	})
	if math.Float64bits(got) != 0 || unit.Experience != 200 {
		t.Fatalf("result/experience = %016x/%v", math.Float64bits(got), unit.Experience)
	}
	runtime.KeepAlive(unit)
}

func TestUnitGiveXPNative4EF270HasNoClassGate(t *testing.T) {
	player := &Player{ProtUnitExperience: 7}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 0, Experience: 1, UpdateData: unsafe.Pointer(update)}
	called := 0
	if got := unitGiveXPNative4EF270(unit, 2, unitGiveXPNativeDeps4EF270{
		protectExperience: func(uint32, float32) { called++ },
		reportExperience:  func(*Object) { called++ },
		syncLevel:         func(*Object) { called++ },
	}); got <= 0 || called != 3 {
		t.Fatalf("award/calls = %v/%d", got, called)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestUnitGiveXP4EF270NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantExperience := uintptr(28)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantProtection := uintptr(4604)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantExperience = 32
		wantUpdate = 872
		wantUpdateSize = 656
		wantPlayer = 336
		wantPlayerSize = 6160
		wantProtection = 5908
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.Experience", unsafe.Offsetof(Object{}.Experience), wantExperience},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.ProtUnitExperience", unsafe.Offsetof(Player{}.ProtUnitExperience), wantProtection},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
