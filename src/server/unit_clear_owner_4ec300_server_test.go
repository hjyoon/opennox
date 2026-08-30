package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestUnitClearOwnerNative4EC300Layout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubClass := uintptr(12)
	wantOwner := uintptr(508)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	wantUpdateData := uintptr(748)
	wantPlayerDataSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerInd := uintptr(2064)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubClass = 16
		wantOwner = 552
		wantNextOwned = 560
		wantFirstOwned = 568
		wantUpdateData = 872
		wantPlayerDataSize = 656
		wantPlayer = 336
		wantPlayerSize = 6160
		wantPlayerInd = 2068
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantPlayerDataSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerInd},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitClearOwnerNative4EC300UsesLiveOwnersAndPlayers(t *testing.T) {
	firstPlayer := &Player{PlayerInd: 0xfe}
	secondPlayer := &Player{PlayerInd: 0x81}
	data := &PlayerUpdateData{Player: firstPlayer}
	entryOwner := &Object{ObjClass: object.ClassPlayer}
	notifyOwner := &Object{UpdateData: unsafe.Pointer(data)}
	next := &Object{}
	listOwner := &Object{}
	obj := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(0xaabbccff),
		ObjOwner:    entryOwner,
		Field128:    next,
	}
	listOwner.Field129 = obj
	events := make([]string, 0, 5)
	unitClearOwnerNative4EC300(obj, unitClearOwnerNativeDeps4EC300{
		isMonitored: func(gotOwner, gotObj *Object) bool {
			events = append(events, "monitored")
			if gotOwner != entryOwner || gotObj != obj {
				t.Fatalf("monitored args = %p/%p", gotOwner, gotObj)
			}
			gotObj.ObjOwner = notifyOwner
			return true
		},
		netFxShield: func(ind uint8, got *Object) {
			events = append(events, "shield")
			if ind != firstPlayer.PlayerInd || got != obj {
				t.Fatalf("shield args = %#x/%p", ind, got)
			}
			data.Player = secondPlayer
		},
		unmarkMinimap: func(ind uint8, got *Object, flags uint32) {
			events = append(events, "unmark")
			if ind != secondPlayer.PlayerInd || got != obj || flags != 1 {
				t.Fatalf("unmark args = %#x/%p/%d", ind, got, flags)
			}
			got.ObjOwner = listOwner
		},
		resetMonster: func(got *Object) {
			events = append(events, "reset")
			got.ObjClass = object.ClassPlayer
		},
		markUnitUpdate: func(got *Object) {
			events = append(events, "mark")
			if got != obj {
				t.Fatalf("mark object = %p", got)
			}
		},
	})
	wantEvents := []string{"monitored", "shield", "unmark", "reset", "mark"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if uint32(obj.ObjSubClass) != 0xaabbcc7f {
		t.Fatalf("subclass = %#x", obj.ObjSubClass)
	}
	if obj.ObjOwner != nil || obj.Field128 != next || listOwner.Field129 != next {
		t.Fatalf("ownership = owner %p next %p head %p", obj.ObjOwner, obj.Field128, listOwner.Field129)
	}
}

func TestUnitClearOwner4EC300ServerBindingRepairsOwnedList(t *testing.T) {
	s := &Server{}
	tail := &Object{}
	obj := &Object{Field128: tail}
	first := &Object{Field128: obj}
	owner := &Object{Field129: first}
	obj.ObjOwner = owner

	s.ObjClearOwner(obj)
	if first.Field128 != tail || obj.ObjOwner != nil || obj.Field128 != tail {
		t.Fatalf("middle removal = predecessor %p owner %p object next %p", first.Field128, obj.ObjOwner, obj.Field128)
	}

	obj.ObjOwner = owner
	obj.Field128 = tail
	owner.Field129 = obj
	s.ObjClearOwner(obj)
	if owner.Field129 != tail || obj.ObjOwner != nil || obj.Field128 != tail {
		t.Fatalf("head removal = head %p owner %p object next %p", owner.Field129, obj.ObjOwner, obj.Field128)
	}

	s.ObjClearOwner(nil)
	s.ObjClearOwner(&Object{})
}
