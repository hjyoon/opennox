package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerRespawnItem4EF750NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantInit := uintptr(688)
	wantUpdateData := uintptr(748)
	wantModifierSize := uintptr(20)
	wantModifierField16 := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantInit = 752
		wantUpdateData = 872
		wantModifierSize = 40
		wantModifierField16 = 32
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Init", unsafe.Offsetof(Object{}.Init), wantInit},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantModifierSize},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"ModifierInitData.Field16", unsafe.Offsetof(ModifierInitData{}.Field16), wantModifierField16},
		{"update prefix size", unsafe.Sizeof(playerRespawnItemUpdatePrefix4EF750{}), 8},
		{"update prefix Mark4", unsafe.Offsetof(playerRespawnItemUpdatePrefix4EF750{}.Mark4), 4},
		{"placement int32 size", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerRespawnItem4EF750NativeFieldsAndCallbacks(t *testing.T) {
	player := &Object{}
	attrs := &ModifierInitData{Field16: 0x12345678}
	update := &playerRespawnItemUpdatePrefix4EF750{Field0: 0xaabbccdd, Mark4: 0x10}
	initToken := uint32(0xfeedface)
	init := unsafe.Pointer(&initToken)
	item := &Object{
		Init:       init,
		ObjFlags:   object.Flags(0xa5a80040),
		ObjClass:   object.ClassWeapon,
		UpdateData: unsafe.Pointer(update),
	}
	events := make([]string, 0, 4)

	got := playerRespawnItemNative4EF750(player, "Longsword", attrs, -4, 5, playerRespawnItemNativeDeps4EF750{
		newObject: func(typeID string) *Object {
			if typeID != "Longsword" {
				t.Fatalf("type ID = %q", typeID)
			}
			events = append(events, "new")
			return item
		},
		callInit: func(gotInit unsafe.Pointer, gotItem *Object, value uint32) {
			if gotInit != init || gotItem != item || value != 0 {
				t.Fatalf("init call = %p/%p/%d", gotInit, gotItem, value)
			}
			events = append(events, "init")
		},
		applyAttrs: func(gotItem *Object, gotAttrs *ModifierInitData) {
			if gotItem != item || gotAttrs != attrs {
				t.Fatalf("modifier call = %p/%p", gotItem, gotAttrs)
			}
			events = append(events, "attrs")
		},
		placeInventory: func(gotPlayer, gotItem *Object, a4, a5 int32) bool {
			if gotPlayer != player || gotItem != item || a4 != -4 || a5 != 5 {
				t.Fatalf("placement = %p/%p/%d/%d", gotPlayer, gotItem, a4, a5)
			}
			events = append(events, "place")
			return false
		},
	})
	if got != item {
		t.Fatalf("result = %p, want %p", got, item)
	}
	wantEvents := []string{"new", "init", "attrs", "place"}
	if len(events) != len(wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	for i := range wantEvents {
		if events[i] != wantEvents[i] {
			t.Fatalf("events = %v, want %v", events, wantEvents)
		}
	}
	if uint32(item.ObjFlags) != 0xa5a00040 || update.Field0 != 0xaabbccdd || update.Mark4 != 0x11 {
		t.Fatalf("native state = flags %#x update %#x/%#x", uint32(item.ObjFlags), update.Field0, update.Mark4)
	}
}

func TestPlayerRespawnItem4EF750NativeNilAndUnmarkedPaths(t *testing.T) {
	called := false
	deps := playerRespawnItemNativeDeps4EF750{
		newObject: func(string) *Object {
			return nil
		},
		callInit: func(unsafe.Pointer, *Object, uint32) {
			called = true
		},
		applyAttrs: func(*Object, *ModifierInitData) {
			called = true
		},
		placeInventory: func(*Object, *Object, int32, int32) bool {
			called = true
			return true
		},
	}
	if got := playerRespawnItemNative4EF750(nil, "Missing", nil, 1, 2, deps); got != nil || called {
		t.Fatalf("nil creation = %p, later callback = %v", got, called)
	}

	item := &Object{ObjFlags: object.FlagRespawn, ObjClass: object.ClassFood}
	deps.newObject = func(string) *Object { return item }
	deps.placeInventory = func(*Object, *Object, int32, int32) bool { return false }
	if got := playerRespawnItemNative4EF750(nil, "Food", nil, 1, 2, deps); got != item {
		t.Fatalf("unmarked result = %p, want %p", got, item)
	}
	if item.ObjFlags.Has(object.FlagRespawn) {
		t.Fatalf("respawn flag was not cleared: %#x", uint32(item.ObjFlags))
	}
}

func TestPlayerRespawnItem4EF750NativeMarkedNilUpdateFaultsAfterFlagStore(t *testing.T) {
	item := &Object{ObjFlags: object.FlagRespawn, ObjClass: object.ClassArmor}
	defer func() {
		if recover() == nil {
			t.Fatal("marked nil UpdateData did not fault")
		}
		if item.ObjFlags.Has(object.FlagRespawn) {
			t.Fatalf("flag store did not precede UpdateData fault: %#x", uint32(item.ObjFlags))
		}
	}()
	playerRespawnItemNative4EF750(nil, "Armor", nil, 1, 2, playerRespawnItemNativeDeps4EF750{
		newObject:      func(string) *Object { return item },
		callInit:       func(unsafe.Pointer, *Object, uint32) {},
		applyAttrs:     func(*Object, *ModifierInitData) {},
		placeInventory: func(*Object, *Object, int32, int32) bool { return true },
	})
}
