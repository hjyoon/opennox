package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestInventoryPut4F3070NativeLayout(t *testing.T) {
	checks32 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 780},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 488},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 490},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 492},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 496},
		{"Object.Field125", unsafe.Offsetof(Object{}.Field125), 500},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 504},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), 508},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 748},
		{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 556},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 276},
		{"Player.size", unsafe.Sizeof(Player{}), 4828},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2064},
		{"Player.Field3656", unsafe.Offsetof(Player{}.Field3656), 3656},
		{"Player.Prot4632", unsafe.Offsetof(Player{}.Prot4632), 4632},
	}
	checks64 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 12},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 20},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), 516},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 518},
		{"Object.InvHolder", unsafe.Offsetof(Object{}.InvHolder), 520},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), 528},
		{"Object.Field125", unsafe.Offsetof(Object{}.Field125), 536},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), 544},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), 552},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), 872},
		{"PlayerUpdateData.size", unsafe.Sizeof(PlayerUpdateData{}), 640},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), 320},
		{"Player.size", unsafe.Sizeof(Player{}), 6160},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), 2068},
		{"Player.Field3656", unsafe.Offsetof(Player{}.Field3656), 4952},
		{"Player.Prot4632", unsafe.Offsetof(Player{}.Prot4632), 5936},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestInventoryPutNative4F3070PreservesPointersAndLiveState(t *testing.T) {
	initialPlayer := &Player{PlayerInd: 7, Prot4632: 77}
	replacementPlayer := &Player{PlayerInd: 9, Prot4632: 1234}
	update := &PlayerUpdateData{Player: initialPlayer}
	staleHead := &Object{}
	owner := &Object{InvFirstItem: staleHead}
	item := &Object{Field125: new(Object)}
	weight2 := &Object{Weight: 5}
	weight1 := &Object{Weight: 250, InvNextItem: weight2}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{
			unsafe.Pointer(owner),
			unsafe.Pointer(item),
			unsafe.Pointer(staleHead),
			unsafe.Pointer(weight1),
			unsafe.Pointer(weight2),
			unsafe.Pointer(update),
			unsafe.Pointer(initialPlayer),
			unsafe.Pointer(replacementPlayer),
		}
		for index, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	var events []string
	inventoryPutNative4F3070(owner, item, -1, inventoryPutNativeDeps4F3070{
		setOwner: func(gotOwner, gotItem *Object) {
			events = append(events, "owner")
			if gotOwner != owner || gotItem != item {
				t.Fatalf("owner args = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
			}
			item.ObjOwner = owner
			owner.ObjClass = object.ClassPlayer
			owner.UpdateData = unsafe.Pointer(update)
		},
		report: func(index uint8, gotItem *Object) {
			events = append(events, "report")
			if index != initialPlayer.PlayerInd || gotItem != item {
				t.Fatalf("report args = %d/%p, want %d/%p", index, gotItem, initialPlayer.PlayerInd, item)
			}
			initialPlayer.Prot4632 = 99
			update.Player = replacementPlayer
		},
		protect: func(value uint32, gotItem *Object) {
			events = append(events, "protect")
			if value != 99 || gotItem != item {
				t.Fatalf("protect args = %d/%p, want 99/%p", value, gotItem, item)
			}
			owner.InvFirstItem = weight1
			owner.CarryCapacity = 200
			item.ObjClass = object.Class(0x40)
		},
		audioEvent: func(id int32, gotOwner *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != inventoryPutSound4F3070 || gotOwner != owner || kind != 0 || code != 0 {
				t.Fatalf("audio args = %d/%p/%d/%d", id, gotOwner, kind, code)
			}
		},
	})

	if want := []string{"owner", "report", "protect", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if item.Field125 != nil || item.InvNextItem != staleHead || staleHead.Field125 != item {
		t.Fatalf("inserted links = prev %p next %p old-prev %p", item.Field125, item.InvNextItem, staleHead.Field125)
	}
	if item.InvHolder != owner || item.ObjOwner != owner {
		t.Fatalf("holder/owner = %p/%p, want %p", item.InvHolder, item.ObjOwner, owner)
	}
	if initialPlayer.Field3656 != 1 || replacementPlayer.Field3656 != 0 {
		t.Fatalf("cached/replacement overweight = %d/%d, want 1/0", initialPlayer.Field3656, replacementPlayer.Field3656)
	}
	runtime.KeepAlive(update)
}

func TestInventoryPutNative4F3070EarlyGatesDoNotCallServices(t *testing.T) {
	deps := inventoryPutNativeDeps4F3070{
		setOwner:   func(*Object, *Object) { t.Fatal("owner callback reached") },
		report:     func(uint8, *Object) { t.Fatal("report callback reached") },
		protect:    func(uint32, *Object) { t.Fatal("protect callback reached") },
		audioEvent: func(int32, *Object, int32, uint32) { t.Fatal("audio callback reached") },
	}
	inventoryPutNative4F3070(nil, new(Object), 1, deps)
	inventoryPutNative4F3070(new(Object), nil, 1, deps)
	inventoryPutNative4F3070(&Object{ObjFlags: object.FlagDestroyed}, new(Object), 1, deps)
	inventoryPutNative4F3070(new(Object), &Object{ObjFlags: object.FlagDestroyed}, 1, deps)
}

func TestInventoryPutNative4F3070PreservesUnguardedPlayerFaults(t *testing.T) {
	for _, test := range []struct {
		name   string
		update *PlayerUpdateData
		report int32
	}{
		{name: "nil update", report: 0},
		{name: "nil player report", update: &PlayerUpdateData{}, report: 1},
		{name: "nil player no report", update: &PlayerUpdateData{}, report: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(test.update)}
			item := &Object{}
			called := false
			defer func() {
				if recover() == nil {
					t.Fatal("unguarded Player path did not fault")
				}
				if called {
					t.Fatal("service callback ran past the original fault")
				}
				if owner.InvFirstItem != item || item.InvHolder != owner {
					t.Fatalf("fault prefix lost insertion: head/holder = %p/%p", owner.InvFirstItem, item.InvHolder)
				}
			}()
			inventoryPutNative4F3070(owner, item, test.report, inventoryPutNativeDeps4F3070{
				setOwner: func(gotOwner, gotItem *Object) {
					gotItem.ObjOwner = gotOwner
				},
				report: func(uint8, *Object) {
					called = true
				},
				protect: func(uint32, *Object) {
					called = true
				},
				audioEvent: func(int32, *Object, int32, uint32) {
					called = true
				},
			})
		})
	}
}
