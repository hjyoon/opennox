package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultPickupAnkhTradableNativeDeps4F3DD0() pickupAnkhTradableNativeDeps4F3DD0 {
	return pickupAnkhTradableNativeDeps4F3DD0{
		delayedDelete: func(*Object) {},
		audio:         func(uint32, *Object, int32, uint32) {},
	}
}

func TestPickupAnkhTradable4F3DD0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantExtraLives := uintptr(320)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantUpdateSize = 640
		wantExtraLives = 400
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ExtraLives", unsafe.Offsetof(PlayerUpdateData{}.ExtraLives), wantExtraLives},
		{"Object.ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"PlayerUpdateData.ExtraLives width", unsafe.Sizeof(PlayerUpdateData{}.ExtraLives), 4},
		{"callback result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPickupAnkhTradableNative4F3DD0BindsFieldsAndNativePointers(t *testing.T) {
	entryUpdate := &PlayerUpdateData{ExtraLives: 9}
	replacementUpdate := &PlayerUpdateData{ExtraLives: 77}
	owner := &Object{
		ObjClass:   object.ClassPlayer | object.Class(0xa5000000),
		UpdateData: unsafe.Pointer(entryUpdate),
	}
	item := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(item)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(entryUpdate)) <= math.MaxUint32) {
		t.Fatalf("native pointer test did not allocate above 32 bits: owner=%p item=%p update=%p", owner, item, entryUpdate)
	}

	var events []string
	deps := defaultPickupAnkhTradableNativeDeps4F3DD0()
	deps.delayedDelete = func(gotItem *Object) {
		events = append(events, "delete")
		if gotItem != item || entryUpdate.ExtraLives != 10 {
			t.Fatalf("delete item/lives = %p/%d, want %p/10", gotItem, entryUpdate.ExtraLives, item)
		}
		owner.UpdateData = unsafe.Pointer(replacementUpdate)
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 1004 || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%#x", id, gotOwner, kind, code)
		}
		if entryUpdate.ExtraLives != 10 || replacementUpdate.ExtraLives != 77 {
			t.Fatalf("cached/replacement ExtraLives = %d/%d", entryUpdate.ExtraLives, replacementUpdate.ExtraLives)
		}
	}

	if got := pickupAnkhTradableNative4F3DD0(owner, item, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"delete", "audio"}) {
		t.Fatalf("events = %v, want delete/audio", events)
	}
}

func TestPickupAnkhTradableNative4F3DD0WrapsExtraLives(t *testing.T) {
	update := &PlayerUpdateData{ExtraLives: math.MaxUint32}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	if got := pickupAnkhTradableNative4F3DD0(owner, nil, 0x12345678, -1, defaultPickupAnkhTradableNativeDeps4F3DD0()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if update.ExtraLives != 0 {
		t.Fatalf("ExtraLives = %#x, want 0", update.ExtraLives)
	}
}

func TestPickupAnkhTradableNative4F3DD0NonPlayerSkipsEffects(t *testing.T) {
	owner := &Object{ObjClass: object.ClassMonster | object.Class(0xff000000)}
	deps := pickupAnkhTradableNativeDeps4F3DD0{
		delayedDelete: func(*Object) { t.Fatal("non-player deleted item") },
		audio:         func(uint32, *Object, int32, uint32) { t.Fatal("non-player emitted audio") },
	}
	if got := pickupAnkhTradableNative4F3DD0(owner, nil, math.MinInt32, math.MaxInt32, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestPickupAnkhTradableNative4F3DD0HasNoNilGuards(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil owner did not preserve the class-load fault")
			}
		}()
		pickupAnkhTradableNative4F3DD0(nil, nil, 0, 0, defaultPickupAnkhTradableNativeDeps4F3DD0())
	})

	t.Run("update", func(t *testing.T) {
		owner := &Object{ObjClass: object.ClassPlayer}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not preserve the ExtraLives fault")
			}
		}()
		pickupAnkhTradableNative4F3DD0(owner, nil, 0, 0, defaultPickupAnkhTradableNativeDeps4F3DD0())
	})
}
