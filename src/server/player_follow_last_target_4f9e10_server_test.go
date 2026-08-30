package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerFollowLastTargetNative4F9E10Layouts(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantLast := uintptr(520)
	wantUpdate := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantPlayerStatus := uintptr(3680)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
		wantLast = 576
		wantUpdate = 872
		wantUpdatePlayer = 336
		wantPlayerStatus = 4976
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Obj130", unsafe.Offsetof(Object{}.Obj130), wantLast},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerStatus},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerFollowLastTargetNative4F9E10PreservesPointers(t *testing.T) {
	pl := &Player{Field3680: 0x80000000}
	update := &PlayerUpdateData{Player: pl}
	target := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	leaf := &Object{ObjOwner: target}
	unit := &Object{Obj130: leaf}
	var gotUnit, gotTarget *Object

	got := playerFollowLastTargetNative4F9E10(unit, PlayerFollowLastTargetRuntime4F9E10{
		CameraFollow: func(gotU, gotT *Object) {
			gotUnit, gotTarget = gotU, gotT
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if gotUnit != unit || gotTarget != target {
		t.Fatalf("camera pointers = %p/%p, want %p/%p", gotUnit, gotTarget, unit, target)
	}
}

func TestPlayerFollowLastTargetNative4F9E10RejectsObserverAndMonster(t *testing.T) {
	called := false
	follow := PlayerFollowLastTargetRuntime4F9E10{
		CameraFollow: func(*Object, *Object) { called = true },
	}
	pl := &Player{Field3680: playerFollowObserverBit4F9E10}
	update := &PlayerUpdateData{Player: pl}
	playerTarget := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if got := playerFollowLastTargetNative4F9E10(&Object{Obj130: playerTarget}, follow); got != 0 {
		t.Fatalf("observer result = %d, want 0", got)
	}
	monster := &Object{ObjClass: object.ClassMonster}
	if got := playerFollowLastTargetNative4F9E10(&Object{Obj130: monster}, follow); got != 0 {
		t.Fatalf("monster result = %d, want 0", got)
	}
	if called {
		t.Fatal("rejected target reached camera callback")
	}
}
