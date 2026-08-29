package server

import (
	"sync/atomic"
	"testing"
	"unsafe"
)

func newMoverStateServer(t *testing.T) *Server {
	t.Helper()
	handle := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: handle}
	s.Objs.init(handle)
	servers.Store(handle, s)
	t.Cleanup(func() { servers.Delete(handle) })
	return s
}

func TestMoverUpdateDataLayoutAndNativePointerIdentity(t *testing.T) {
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "size", got: unsafe.Sizeof(MoverUpdateData{}), want: 36},
		{name: "Field_0", got: unsafe.Offsetof(MoverUpdateData{}.Field_0), want: 0},
		{name: "Field_1", got: unsafe.Offsetof(MoverUpdateData{}.Field_1), want: 4},
		{name: "Field_2", got: unsafe.Offsetof(MoverUpdateData{}.Field_2), want: 8},
		{name: "Field_3", got: unsafe.Offsetof(MoverUpdateData{}.Field_3), want: 12},
		{name: "Field_4", got: unsafe.Offsetof(MoverUpdateData{}.Field_4), want: 16},
		{name: "Field_5", got: unsafe.Offsetof(MoverUpdateData{}.Field_5), want: 20},
		{name: "Field_6", got: unsafe.Offsetof(MoverUpdateData{}.Field_6), want: 24},
		{name: "Field_7", got: unsafe.Offsetof(MoverUpdateData{}.Field_7), want: 28},
		{name: "Field_8", got: unsafe.Offsetof(MoverUpdateData{}.Field_8), want: 32},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Mover update-data %s = %d, want %d", check.name, check.got, check.want)
		}
	}

	s := newMoverStateServer(t)
	entryData := &MoverUpdateData{
		Field_0: 0x7d,
		Field_1: 1.25,
		Field_2: -7,
		Field_3: 0xffffffff,
		Field_4: 0x11223344,
		Field_5: 0xeeeeeeee,
		Field_6: 0x55667788,
		Field_7: 0xdddddddd,
		Field_8: 0x99aabbcc,
	}
	liveData := &MoverUpdateData{Field_4: 2, Field_6: 3, Field_8: 4}
	obj := &Object{UpdateData: unsafe.Pointer(entryData), serverHandle: s.handle}
	waypoint3 := &Waypoint{Index: 0x01020304}
	waypoint5 := &Waypoint{Index: 0x05060708}
	target := &Object{serverHandle: s.handle}

	obj.SetMoverWaypointFor(entryData, 3, waypoint3)
	obj.SetMoverWaypointFor(entryData, 5, waypoint5)
	obj.SetMoverTargetFor(entryData, target)
	if entryData.Field_3 != 0 || entryData.Field_5 != 0 || entryData.Field_7 != 0 {
		t.Fatalf("entry PE32 slots = %#x/%#x/%#x, want zero", entryData.Field_3, entryData.Field_5, entryData.Field_7)
	}
	if entryData.Field_0 != 0x7d || entryData.Field_1 != 1.25 || entryData.Field_2 != -7 ||
		entryData.Field_4 != 0x11223344 || entryData.Field_6 != 0x55667788 || entryData.Field_8 != 0x99aabbcc {
		t.Fatalf("fixed fields changed while binding native pointers: %+v", *entryData)
	}
	if got := obj.MoverWaypointFor(entryData, 3); got != waypoint3 {
		t.Fatalf("slot-3 waypoint = %p, want %p", got, waypoint3)
	}
	if got := obj.MoverWaypointFor(entryData, 5); got != waypoint5 {
		t.Fatalf("slot-5 waypoint = %p, want %p", got, waypoint5)
	}
	if got := obj.MoverTargetFor(entryData); got != target {
		t.Fatalf("target = %p, want %p", got, target)
	}

	obj.UpdateData = unsafe.Pointer(liveData)
	if obj.MoverWaypointFor(liveData, 3) != nil || obj.MoverTarget() != nil {
		t.Fatal("replacement update data inherited stale native pointers")
	}
	obj.SetMoverWaypointFor(liveData, 3, waypoint5)
	if liveData.Field_3 != 0 || liveData.Field_5 != 0 || liveData.Field_7 != 0 ||
		liveData.Field_4 != 2 || liveData.Field_6 != 3 || liveData.Field_8 != 4 {
		t.Fatalf("live PE32/fixed fields = %#x/%#x/%#x/%#x/%#x/%#x",
			liveData.Field_3, liveData.Field_5, liveData.Field_7,
			liveData.Field_4, liveData.Field_6, liveData.Field_8)
	}
	if obj.MoverWaypointFor(entryData, 3) != nil || obj.MoverTargetFor(entryData) != nil {
		t.Fatal("replacing update-data identity retained stale native pointers")
	}

	standaloneData := &MoverUpdateData{Field_3: 1, Field_5: 2, Field_7: 3, Field_8: 4}
	standalone := &Object{UpdateData: unsafe.Pointer(standaloneData)}
	standalone.SetMoverWaypointFor(standaloneData, 5, waypoint5)
	standalone.SetMoverTarget(target)
	if standaloneData.Field_3 != 1 || standaloneData.Field_5 != 0 || standaloneData.Field_7 != 0 ||
		standaloneData.Field_8 != 4 || standalone.MoverTarget() != nil {
		t.Fatalf("standalone PE32/native state = %#x/%#x/%#x/%#x/%p",
			standaloneData.Field_3, standaloneData.Field_5, standaloneData.Field_7,
			standaloneData.Field_8, standalone.MoverTarget())
	}
}

func TestMoverWaypointSlotRejectsUnknownField(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("unknown waypoint slot did not panic")
		}
	}()
	(&Object{}).SetMoverWaypointFor(&MoverUpdateData{}, 4, nil)
}
