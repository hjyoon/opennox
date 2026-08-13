package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDoorCheckKey4E8910NativeUsesLowTypeWordAndNamedDoorFields(t *testing.T) {
	s := &Server{}
	s.Types.byInd = make([]*ObjectType, 4)
	s.Types.byInd[2] = &ObjectType{ind: 2, id: "GoldKey"}

	key := &Object{TypeInd: 2, ObjClass: object.ClassKey}
	unit := &Object{InvFirstItem: key}
	data := &DoorUpdateData{LockCode: 2, TileX: 0x1020304, TileY: -17, QuestSync: 0xa5}
	door := &Object{UpdateData: unsafe.Pointer(data)}

	if got := s.DoorCheckKey(unit, door); got != key {
		t.Fatalf("result = %p, want key %p", got, key)
	}
	if data.LockCode != 2 || data.TileX != 0x1020304 || data.TileY != -17 || data.QuestSync != 0xa5 {
		t.Fatalf("door update data changed: %+v", data)
	}
	if key.TypeInd != 2 || key.ObjClass != object.ClassKey || unit.InvFirstItem != key {
		t.Fatal("inventory key or unit link changed")
	}
}

func TestDoorCheckKey4E8910NativeEarlyGatesAndNilUnitFault(t *testing.T) {
	s := &Server{}
	mechanism := &DoorUpdateData{LockCode: 5}
	if got := s.DoorCheckKey(nil, &Object{UpdateData: unsafe.Pointer(mechanism)}); got != nil {
		t.Fatalf("mechanism lock returned %p", got)
	}
	owned := &DoorUpdateData{LockCode: 1}
	if got := s.DoorCheckKey(nil, &Object{UpdateData: unsafe.Pointer(owned), ObjOwner: &Object{}}); got != nil {
		t.Fatalf("owned door returned %p", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault after the door gates")
		}
	}()
	s.DoorCheckKey(nil, &Object{UpdateData: unsafe.Pointer(&DoorUpdateData{LockCode: 1})})
}

func TestPlayersHaveSilverKey4E8A10NativeSelectsLastActiveUnit(t *testing.T) {
	s := &Server{}
	s.Types.byID = map[string]*ObjectType{"silverkey": {ind: 7, id: "SilverKey"}}
	s.Players.list = make([]Player, 3)

	firstKey := &Object{TypeInd: 7}
	firstUnit := &Object{ObjClass: object.ClassPlayer, InvFirstItem: firstKey}
	secondWrong := &Object{TypeInd: 8}
	secondKey := &Object{TypeInd: 7}
	secondWrong.InvNextItem = secondKey
	secondUnit := &Object{ObjClass: object.ClassPlayer, InvFirstItem: secondWrong}

	for i := range s.Players.list {
		s.Players.list[i].PlayerInd = byte(i)
	}
	s.Players.list[0].Active = 1
	s.Players.list[0].PlayerUnit = firstUnit
	s.Players.list[1].Active = 1
	s.Players.list[1].PlayerUnit = nil
	s.Players.list[2].Active = 1
	s.Players.list[2].PlayerUnit = secondUnit
	firstUnit.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &s.Players.list[0]})
	secondUnit.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &s.Players.list[2]})

	if got := s.PlayersHaveSilverKey(); got != secondKey {
		t.Fatalf("result = %p, want last active unit key %p", got, secondKey)
	}
	if s.Types.fast.silverKey != 7 {
		t.Fatalf("SilverKey cache = %d, want 7", s.Types.fast.silverKey)
	}
	if firstUnit.InvFirstItem != firstKey || secondUnit.InvFirstItem != secondWrong || secondWrong.InvNextItem != secondKey {
		t.Fatal("player inventory links changed")
	}
}

func TestDoorCheckKey4E8910NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantTypeInd, wantClass := uintptr(8), uintptr(12)
	wantNext, wantFirst, wantOwner, wantUpdate := uintptr(528), uintptr(544), uintptr(552), uintptr(872)
	if ptrSize == 4 {
		wantTypeInd, wantClass = 4, 8
		wantNext, wantFirst, wantOwner, wantUpdate = 496, 504, 508, 748
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"DoorUpdateData.LockCode", unsafe.Offsetof(DoorUpdateData{}.LockCode), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
	if unsafe.Sizeof(DoorUpdateData{}) != 52 {
		t.Fatalf("DoorUpdateData size = %d, want 52", unsafe.Sizeof(DoorUpdateData{}))
	}
}
