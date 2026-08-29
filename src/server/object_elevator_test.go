package server

import (
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func newElevatorStateServer(t *testing.T) *Server {
	t.Helper()
	handle := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: handle}
	s.Objs.init(handle)
	servers.Store(handle, s)
	t.Cleanup(func() { servers.Delete(handle) })
	return s
}

func TestElevatorUpdateDataLayoutAndNativeLinkIdentity(t *testing.T) {
	if got := unsafe.Sizeof(ElevatorUpdateData{}); got != 20 {
		t.Fatalf("Elevator update-data size = %d, want 20", got)
	}
	if got := unsafe.Offsetof(ElevatorUpdateData{}.Field_1); got != 4 {
		t.Fatalf("Elevator Field_1 offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(ElevatorUpdateData{}.Field_2); got != 8 {
		t.Fatalf("Elevator Field_2 offset = %d, want 8", got)
	}
	if got := unsafe.Offsetof(ElevatorUpdateData{}.Field_3); got != 12 {
		t.Fatalf("Elevator Field_3 offset = %d, want 12", got)
	}
	if got := unsafe.Offsetof(ElevatorUpdateData{}.Field_4); got != 16 {
		t.Fatalf("Elevator Field_4 offset = %d, want 16", got)
	}
	if got := unsafe.Sizeof(ElevatorShaftUpdateData{}); got != 16 {
		t.Fatalf("Elevator shaft update-data size = %d, want 16", got)
	}
	if got := unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_1); got != 4 {
		t.Fatalf("Elevator shaft Field_1 offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_2); got != 8 {
		t.Fatalf("Elevator shaft Field_2 offset = %d, want 8", got)
	}
	if got := unsafe.Offsetof(ElevatorShaftUpdateData{}.Field_3); got != 12 {
		t.Fatalf("Elevator shaft Field_3 offset = %d, want 12", got)
	}

	s := newElevatorStateServer(t)
	entryData := &ElevatorUpdateData{
		Field_0: 0x01020304,
		Field_1: 0xffffffff,
		Field_2: 0x11223344,
		Field_3: 3,
		Field_4: 0x55667788,
	}
	liveData := &ElevatorUpdateData{
		Field_0: 0x11121314,
		Field_1: 0xeeeeeeee,
		Field_2: 0x99aabbcc,
		Field_3: 1,
		Field_4: 0xddeeff00,
	}
	obj := &Object{UpdateData: unsafe.Pointer(entryData), serverHandle: s.handle}
	target := &Object{serverHandle: s.handle}

	obj.SetElevatorLinkFor(unsafe.Pointer(entryData), target)
	if entryData.Field_1 != 0 || entryData.Field_0 != 0x01020304 ||
		entryData.Field_2 != 0x11223344 || entryData.Field_3 != 3 ||
		entryData.Field_4 != 0x55667788 {
		t.Fatalf("entry data changed outside PE32 slot: %+v", *entryData)
	}
	if got := obj.ElevatorLinkFor(unsafe.Pointer(entryData)); got != target {
		t.Fatalf("entry link = %p, want %p", got, target)
	}

	obj.UpdateData = unsafe.Pointer(liveData)
	if got := obj.ElevatorLink(); got != nil {
		t.Fatalf("replacement-data link = %p, want nil", got)
	}
	if got := obj.ElevatorLinkFor(unsafe.Pointer(entryData)); got != target {
		t.Fatalf("cached-data link = %p, want %p", got, target)
	}
	obj.SetElevatorLinkFor(unsafe.Pointer(liveData), nil)
	if liveData.Field_1 != 0 || liveData.Field_2 != 0x99aabbcc {
		t.Fatalf("live PE32/extent = %#x/%#x, want 0/0x99aabbcc",
			liveData.Field_1, liveData.Field_2)
	}
	if got := obj.ElevatorLinkFor(unsafe.Pointer(entryData)); got != nil {
		t.Fatalf("stale-data link = %p, want nil", got)
	}

	standaloneData := &ElevatorShaftUpdateData{Field_1: 0xabcdef01, Field_2: 7}
	standalone := &Object{UpdateData: unsafe.Pointer(standaloneData)}
	standalone.SetElevatorLink(target)
	if standaloneData.Field_1 != 0 || standaloneData.Field_2 != 7 || standalone.ElevatorLink() != nil {
		t.Fatalf("standalone PE32/extent/link = %#x/%d/%p, want 0/7/nil",
			standaloneData.Field_1, standaloneData.Field_2, standalone.ElevatorLink())
	}
}

func TestAttachPendingElevatorKeepsPE32SlotsBesideReciprocalNativeLinks(t *testing.T) {
	s := newElevatorStateServer(t)
	const shaftExtent = uint32(0xf1234567)
	elevatorData := &ElevatorUpdateData{
		Field_0: 0x01020304,
		Field_1: 0xffffffff,
		Field_2: shaftExtent,
		Field_3: 1,
		Field_4: 0x55667788,
	}
	shaftData := &ElevatorShaftUpdateData{
		Field_0: 0x11121314,
		Field_1: 0xeeeeeeee,
		Field_2: 0x99aabbcc,
		Field_3: 3,
	}
	shaft := &Object{
		ObjClass:     object.ClassElevatorShaft,
		Extent:       shaftExtent,
		UpdateData:   unsafe.Pointer(shaftData),
		serverHandle: s.handle,
	}
	distractor := &Object{
		Extent:       shaftExtent,
		ObjNext:      shaft,
		serverHandle: s.handle,
	}
	elevator := &Object{
		ObjClass:     object.ClassElevator,
		UpdateData:   unsafe.Pointer(elevatorData),
		ObjNext:      distractor,
		serverHandle: s.handle,
	}
	s.Objs.Pending = elevator

	s.AttachPending()

	if got := elevator.ElevatorLink(); got != shaft {
		t.Fatalf("elevator link = %p, want shaft %p", got, shaft)
	}
	if got := shaft.ElevatorLink(); got != elevator {
		t.Fatalf("shaft link = %p, want elevator %p", got, elevator)
	}
	if elevatorData.Field_1 != 0 || elevatorData.Field_2 != shaftExtent {
		t.Fatalf("elevator PE32/extent = %#x/%#x, want 0/%#x",
			elevatorData.Field_1, elevatorData.Field_2, shaftExtent)
	}
	if shaftData.Field_1 != 0 || shaftData.Field_2 != 0x99aabbcc {
		t.Fatalf("shaft PE32/field2 = %#x/%#x, want 0/0x99aabbcc",
			shaftData.Field_1, shaftData.Field_2)
	}
}

func TestAttachPendingElevatorDoesNotResolveShaftAsSource(t *testing.T) {
	s := newElevatorStateServer(t)
	const elevatorExtent = uint32(0x12345678)
	shaftData := &ElevatorShaftUpdateData{Field_2: elevatorExtent}
	elevatorData := &ElevatorUpdateData{Field_2: 0x87654321}
	elevator := &Object{
		ObjClass:     object.ClassElevator,
		Extent:       elevatorExtent,
		UpdateData:   unsafe.Pointer(elevatorData),
		serverHandle: s.handle,
	}
	shaft := &Object{
		ObjClass:     object.ClassElevatorShaft,
		UpdateData:   unsafe.Pointer(shaftData),
		ObjNext:      elevator,
		serverHandle: s.handle,
	}
	s.Objs.Pending = shaft

	s.AttachPending()

	if got := shaft.ElevatorLink(); got != nil {
		t.Fatalf("shaft-source link = %p, want nil", got)
	}
	if got := elevator.ElevatorLink(); got != nil {
		t.Fatalf("unmatched elevator link = %p, want nil", got)
	}
}

func TestAIPathElevatorLinksStayNativeWidth(t *testing.T) {
	tests := []struct {
		name    string
		class   object.Class
		flag    AIMapIndexFlags
		newData func() (unsafe.Pointer, func() uint32)
	}{
		{
			name:  "elevator",
			class: object.ClassElevator,
			flag:  AIIndexElevator,
			newData: func() (unsafe.Pointer, func() uint32) {
				data := &ElevatorUpdateData{Field_1: 0xffffffff, Field_2: 0x76543210}
				return unsafe.Pointer(data), func() uint32 { return data.Field_1 }
			},
		},
		{
			name:  "shaft",
			class: object.ClassElevatorShaft,
			flag:  AIIndexElevatorShaft,
			newData: func() (unsafe.Pointer, func() uint32) {
				data := &ElevatorShaftUpdateData{Field_1: 0xffffffff, Field_2: 0x76543210}
				return unsafe.Pointer(data), func() uint32 { return data.Field_1 }
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newElevatorStateServer(t)
			s.Map.Init()
			t.Cleanup(s.Map.Free)

			const x, y = 4, 5
			pos := types.Ptf(float32(x*23), float32(y*23))
			data, pe32Link := tc.newData()
			unit := &Object{
				ObjClass:     tc.class,
				ObjFlags:     object.FlagActive | object.FlagEnabled,
				PosVec:       pos,
				NewPos:       pos,
				UpdateData:   data,
				serverHandle: s.handle,
			}
			unit.Shape.Kind = ShapeKindCircle
			unit.Shape.Circle.R = 1
			target := &Object{PosVec: types.Ptf(230, 460), serverHandle: s.handle}
			unit.SetElevatorLink(target)
			if got := pe32Link(); got != 0 {
				t.Fatalf("PE32 link slot = %#x, want 0", got)
			}
			s.Map.AddObjectToIndex(unit)

			paths := serverAIPaths{s: s}
			paths.MapIndex(x, y).Flags8 = tc.flag
			var out [2]uint16
			if got := paths.Sub_50AC20(&AIVisitNode{X0: x, Y2: y}, &out); got != 1 {
				t.Fatalf("path link result = %d, want 1", got)
			}
			if want := [2]uint16{10, 20}; out != want {
				t.Fatalf("path destination = %v, want %v", out, want)
			}
		})
	}
}
