package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultObjectDropDispatchNativeDeps4ED790() objectDropDispatchNativeDeps4ED790 {
	return objectDropDispatchNativeDeps4ED790{
		gameFlag:    func(uint32) int32 { return 0 },
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 0 },
		refreshUnit: func(*Object) {},
	}
}

func TestObjectDropDispatch4ED790NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantDrop := uintptr(712)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantDrop = 800
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Drop", unsafe.Offsetof(Object{}.Drop), wantDrop},
		{"DropFuncPtr size", unsafe.Sizeof(DropFuncPtr{}), unsafe.Sizeof(uintptr(0))},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestObjectDropDispatch4ED790NativeBindsLiveHandlerAndFields(t *testing.T) {
	var entryToken, liveToken byte
	entryPtr := unsafe.Pointer(&entryToken)
	livePtr := unsafe.Pointer(&liveToken)
	owner := &Object{}
	item := &Object{
		ObjClass: object.ClassFood,
		ObjFlags: object.Flags(0x80000005),
		Drop:     DropFuncPtr{Ptr: entryPtr},
	}
	point := &types.Pointf{X: 12.5, Y: -3.25}
	wantResult := int32(-0x1234567)
	events := make([]string, 0, 4)
	objDrop.Register(entryPtr, func(*Object, *Object, *types.Pointf) int32 {
		t.Fatal("entry handler was loaded before refresh")
		return 0
	})
	objDrop.Register(livePtr, func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "drop")
		if gotOwner != owner || gotItem != item || gotPoint != point {
			t.Fatalf("drop args = %p/%p/%p, want %p/%p/%p", gotOwner, gotItem, gotPoint, owner, item, point)
		}
		return wantResult
	})

	deps := defaultObjectDropDispatchNativeDeps4ED790()
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "game")
		if flag == objectDropOnlineFlag4ED790 {
			return -1
		}
		if flag == objectDropQuestFlag4ED790 {
			return 0
		}
		t.Fatalf("game flag = %#x", flag)
		return 0
	}
	deps.refreshUnit = func(gotItem *Object) {
		events = append(events, "refresh")
		if gotItem != item || uint32(item.ObjFlags) != 0x80000045 {
			t.Fatalf("refresh item/flags = %p/%#x, want %p/0x80000045", gotItem, uint32(item.ObjFlags), item)
		}
		item.Drop.Ptr = livePtr
	}
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 {
		t.Fatal("default drop called with a live handler")
		return 0
	}

	if got := objectDropDispatchNative4ED790(owner, item, point, deps); got != wantResult {
		t.Fatalf("result = %#x, want %#x", uint32(got), uint32(wantResult))
	}
	if want := []string{"game", "game", "refresh", "drop"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestObjectDropDispatch4ED790NativeDefaultFallback(t *testing.T) {
	item := &Object{}
	deps := defaultObjectDropDispatchNativeDeps4ED790()
	deps.defaultDrop = func(owner, gotItem *Object, point *types.Pointf) int32 {
		if owner != nil || gotItem != item || point != nil {
			t.Fatalf("default args = %p/%p/%p, want nil/%p/nil", owner, gotItem, point, item)
		}
		return int32(0x76543210)
	}
	if got := objectDropDispatchNative4ED790(nil, item, nil, deps); got != int32(0x76543210) {
		t.Fatalf("result = %#x, want 0x76543210", uint32(got))
	}
}

func TestObjectDropDispatch4ED790NativeHandlerReceivesNilPoint(t *testing.T) {
	var token byte
	ptr := unsafe.Pointer(&token)
	item := &Object{Drop: DropFuncPtr{Ptr: ptr}}
	objDrop.Register(ptr, func(owner, gotItem *Object, point *types.Pointf) int32 {
		if owner != nil || gotItem != item || point != nil {
			t.Fatalf("drop args = %p/%p/%p, want nil/%p/nil", owner, gotItem, point, item)
		}
		return -1
	})
	deps := defaultObjectDropDispatchNativeDeps4ED790()
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 {
		t.Fatal("default drop called with a live handler")
		return 0
	}
	if got := objectDropDispatchNative4ED790(nil, item, nil, deps); got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
}

func TestObjectDropDispatch4ED790NativeNilItemAvoidsRuntime(t *testing.T) {
	deps := objectDropDispatchNativeDeps4ED790{
		gameFlag: func(uint32) int32 { t.Fatal("game flag called"); return 0 },
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 {
			t.Fatal("default drop called")
			return 0
		},
		refreshUnit: func(*Object) { t.Fatal("refresh called") },
	}
	if got := objectDropDispatchNative4ED790(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}
