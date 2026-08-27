package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func defaultPickupUseNativeDeps4F34D0() pickupUseNativeDeps4F34D0 {
	return pickupUseNativeDeps4F34D0{
		useByNetCode:  func(*Object, *Object) int32 { return 0 },
		defaultPickup: func(*Object, *Object, int32, int32) int32 { return 0 },
	}
}

func TestPickupUse4F34D0NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantFlags := uintptr(16)
	wantUse := uintptr(732)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantFlags = 20
		wantUse = 840
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Use", unsafe.Offsetof(Object{}.Use), wantUse},
		{"Object.ObjFlags size", unsafe.Sizeof(Object{}.ObjFlags), 4},
		{"UseFuncPtr size", unsafe.Sizeof(UseFuncPtr{}), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPickupUseNative4F34D0BindsUseAndLiveDestroyedFlag(t *testing.T) {
	var token byte
	usePointer := unsafe.Pointer(&token)
	owner := &Object{}
	item := &Object{Use: UseFuncPtr{Ptr: usePointer}}
	calls := 0
	objUse.Register(usePointer, func(gotOwner, gotItem *Object) bool {
		calls++
		if gotOwner != owner || gotItem != item {
			t.Fatalf("Use args = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
		item.ObjFlags = object.FlagDestroyed | object.Flags(0x80000000)
		return false
	})
	s := &Server{}
	deps := defaultPickupUseNativeDeps4F34D0()
	deps.useByNetCode = s.pickupUseByNetCode4F34D0
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("destroyed Use path called DefaultPickup")
		return 0
	}
	if got := pickupUseNative4F34D0(owner, item, -1, -2, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if calls != 1 {
		t.Fatalf("Use calls = %d, want 1", calls)
	}
}

func TestPickupUseByNetCode4F34D0NilItemAndNilUseSkipOwnerState(t *testing.T) {
	s := &Server{}
	owner := &Object{ObjClass: object.ClassPlayer}
	for _, item := range []*Object{nil, {}} {
		if got := s.pickupUseByNetCode4F34D0(owner, item); got != 1 {
			t.Fatalf("item %p result = %d, want 1", item, got)
		}
	}
}

func TestPickupUseByNetCode4F34D0SpecialPlayerSkipsUse(t *testing.T) {
	var token byte
	usePointer := unsafe.Pointer(&token)
	calls := 0
	objUse.Register(usePointer, func(*Object, *Object) bool {
		calls++
		return false
	})
	pl := &Player{PlayerInd: 3}
	update := &PlayerUpdateData{Player: pl}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	item := &Object{Use: UseFuncPtr{Ptr: usePointer}}
	s := &Server{}
	s.Players.playersXxx = uint32(1) << pl.PlayerInd
	if got := s.pickupUseByNetCode4F34D0(owner, item); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if calls != 0 {
		t.Fatalf("Use calls = %d, want 0", calls)
	}
}

func TestPickupUseNative4F34D0ForwardsFourArgsAndExactResult(t *testing.T) {
	owner := &Object{}
	item := &Object{ObjFlags: object.FlagActive}
	events := make([]string, 0, 2)
	deps := defaultPickupUseNativeDeps4F34D0()
	deps.useByNetCode = func(gotOwner, gotItem *Object) int32 {
		events = append(events, "use")
		if gotOwner != owner || gotItem != item {
			t.Fatalf("Use args = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
		return math.MinInt32
	}
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("DefaultPickup args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		return math.MinInt32
	}
	if got := pickupUseNative4F34D0(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if want := []string{"use", "default"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupUse4F34D0ServerBindingDefaultPickup(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	owner := &Object{CarryCapacity: 50}
	item := &Object{TypeInd: 17, ObjFlags: object.FlagActive, Weight: 3}
	events := make([]string, 0, 2)
	runtime := PickupUseRuntime4F34D0{
		DefaultPickup: PickupDefaultRuntime4F31E0{
			DeleteWorldObject: func(gotItem *Object) {
				events = append(events, "delete")
				if gotItem != item {
					t.Fatalf("deleted = %p, want %p", gotItem, item)
				}
				item.ObjFlags &^= object.FlagActive
			},
			InventoryPut: func(gotOwner, gotItem *Object, report int32) {
				events = append(events, "put")
				if gotOwner != owner || gotItem != item || report != math.MinInt32 {
					t.Fatalf("put args = %p/%p/%d", gotOwner, gotItem, report)
				}
			},
		},
	}
	if got := s.PickupUse4F34D0(owner, item, math.MinInt32, math.MaxInt32, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"delete", "put"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupUseNative4F34D0NilItemFaultsAfterUseHelper(t *testing.T) {
	called := false
	deps := defaultPickupUseNativeDeps4F34D0()
	deps.useByNetCode = func(*Object, *Object) int32 {
		called = true
		return 1
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item flags did not fault")
		}
		if !called {
			t.Fatal("nil item faulted before use helper")
		}
	}()
	pickupUseNative4F34D0(&Object{}, nil, 0, 0, deps)
}
