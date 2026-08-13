package server

import (
	"testing"
	"unsafe"
)

func TestPickupCollideNative4E8DF0GuardReturnsFullUnitPointer(t *testing.T) {
	collision := uint32(0xfeedface)
	deps := pickupCollideNativeDeps4E8DF0{
		frame: func() uint32 {
			t.Fatal("nil/Monster guard read frame")
			return 0
		},
		fps: func() uint32 {
			t.Fatal("nil/Monster guard read FPS")
			return 0
		},
		placeInventory: func(*Object, *Object, int, int) bool {
			t.Fatal("nil/Monster guard placed inventory")
			return false
		},
	}
	if got := pickupCollideNative4E8DF0(nil, nil, unsafe.Pointer(&collision), deps); got != 0 {
		t.Fatalf("nil-unit result = %#x, want 0", got)
	}

	unit := &Object{ObjClass: 0x02}
	got := pickupCollideNative4E8DF0(nil, unit, unsafe.Pointer(&collision), deps)
	if want := uintptr(unit.CObj()); got != want {
		t.Fatalf("guard result = %#x, want full pointer %#x", got, want)
	}
}

func TestPickupCollideNative4E8DF0InventoryResultAndArguments(t *testing.T) {
	s := &Server{}
	s.SetFrame(100)
	s.SetTickRate(20)
	item := &Object{Field32: 90}
	unit := &Object{ObjClass: 0x80}
	called := 0
	runtime := PickupCollideRuntime4E8DF0{
		PlaceInventory: func(gotUnit, gotItem *Object, flag1, flag2 int) bool {
			called++
			if gotUnit != unit || gotItem != item || flag1 != 1 || flag2 != 1 {
				t.Fatalf("place args = (%p,%p,%d,%d)", gotUnit, gotItem, flag1, flag2)
			}
			return called == 2
		},
	}
	if got := s.PickupCollide4E8DF0(item, unit, nil, runtime); got != 0 {
		t.Fatalf("false inventory result = %#x, want 0", got)
	}
	if got := s.PickupCollide4E8DF0(item, unit, nil, runtime); got != 1 {
		t.Fatalf("true inventory result = %#x, want 1", got)
	}
	if called != 2 {
		t.Fatalf("place calls = %d, want 2", called)
	}
}

func TestPickupCollideNative4E8DF0UsesCachedClassAndMovementFlags(t *testing.T) {
	update := &PlayerUpdateData{MovementFlags: 0}
	unit := &Object{ObjClass: 0x04, UpdateData: unsafe.Pointer(update)}
	item := &Object{}
	placed := false
	got := pickupCollideNative4E8DF0(item, unit, nil, pickupCollideNativeDeps4E8DF0{
		frame: func() uint32 {
			unit.ObjClass = 0
			return 10
		},
		fps: func() uint32 { return 20 },
		placeInventory: func(*Object, *Object, int, int) bool {
			placed = true
			return true
		},
	})
	if want := uintptr(unit.CObj()); got != want || placed {
		t.Fatalf("cached Player gate result = %#x, want %#x; placed=%v", got, want, placed)
	}

	unit.ObjClass = 0x04
	update.MovementFlags = 0x101
	got = pickupCollideNative4E8DF0(item, unit, nil, pickupCollideNativeDeps4E8DF0{
		frame:          func() uint32 { return 10 },
		fps:            func() uint32 { return 20 },
		placeInventory: func(*Object, *Object, int, int) bool { return true },
	})
	if got != 1 {
		t.Fatalf("movement-enabled result = %#x, want 1", got)
	}
}

func TestPickupCollideNative4E8DF0NilItemFaultOrder(t *testing.T) {
	unit := &Object{ObjClass: 0x80}
	frameRead := false
	fpsRead := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not fault")
		}
		if !frameRead || fpsRead {
			t.Fatalf("frameRead=%v fpsRead=%v, want true/false", frameRead, fpsRead)
		}
	}()
	pickupCollideNative4E8DF0(nil, unit, nil, pickupCollideNativeDeps4E8DF0{
		frame: func() uint32 {
			frameRead = true
			return 1
		},
		fps: func() uint32 {
			fpsRead = true
			return 1
		},
		placeInventory: func(*Object, *Object, int, int) bool { return false },
	})
}

func TestPickupCollideNative4E8DF0Layouts(t *testing.T) {
	wantClass, wantFrame, wantUpdate, wantMovement := uintptr(12), uintptr(132), uintptr(872), uintptr(284)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		wantClass, wantFrame, wantUpdate, wantMovement = 8, 128, 748, 240
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Field32", unsafe.Offsetof(Object{}.Field32), wantFrame},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData.MovementFlags", unsafe.Offsetof(PlayerUpdateData{}.MovementFlags), wantMovement},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}
