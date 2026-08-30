package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestCustomWaypointNativeLayout4F7950(t *testing.T) {
	wantSize := uintptr(556)
	wantWaypoints := uintptr(168)
	wantWrite := uintptr(180)
	wantRead := uintptr(181)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 656
		wantWaypoints = 200
		wantWrite = 224
		wantRead = 225
	}
	for _, tc := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"PlayerUpdateData size", unsafe.Sizeof(server.PlayerUpdateData{}), wantSize},
		{"CustomWaypoints offset", unsafe.Offsetof(server.PlayerUpdateData{}.CustomWaypoints), wantWaypoints},
		{"CustomWaypoints size", unsafe.Sizeof(server.PlayerUpdateData{}.CustomWaypoints), 3 * unsafe.Sizeof(uintptr(0))},
		{"CustomWaypointWrite offset", unsafe.Offsetof(server.PlayerUpdateData{}.CustomWaypointWrite), wantWrite},
		{"CustomWaypointRead offset", unsafe.Offsetof(server.PlayerUpdateData{}.CustomWaypointRead), wantRead},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestCustomWaypointPresence4F9A80PreservesNativePointer(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	defer freeUnit()
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeUpdate()
	waypoint, freeWaypoint := alloc.New(server.Object{})
	defer freeWaypoint()

	unit.UpdateData = unsafe.Pointer(update)
	update.CustomWaypointWrite = 1
	update.CustomWaypointRead = 2
	update.CustomWaypoints[2] = waypoint

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit":     unsafe.Pointer(unit),
			"update":   unsafe.Pointer(update),
			"waypoint": unsafe.Pointer(waypoint),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above 4GiB", name, pointer)
			}
		}
	}

	if got := Sub_4F9A80(unit); got != 1 {
		t.Fatalf("C custom-waypoint presence = %d, want 1", got)
	}
	if got := update.CustomWaypoints[2]; got != waypoint {
		t.Fatalf("custom waypoint after C read = %p, want %p", got, waypoint)
	}

	update.CustomWaypoints[2] = nil
	if got := Sub_4F9A80(unit); got != 0 {
		t.Fatalf("C empty custom-waypoint presence = %d, want 0", got)
	}
	Sub_4F7950(unit)
	if update.CustomWaypointWrite != 0 || update.CustomWaypointRead != 0 {
		t.Fatalf("C cleanup indices = %d/%d, want 0/0", update.CustomWaypointWrite, update.CustomWaypointRead)
	}

	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(waypoint)
}
