package legacy

import (
	"testing"
	"unsafe"
)

func TestMapgenClosestPrefabRoomsKeepsSixAndStableTies(t *testing.T) {
	old := mapgenRooms
	mapgenRooms = mapgenRoomList{}
	locations := [...][2]int32{{0, -9}, {0, -5}, {0, -3}, {3, -4}, {0, -10}, {0, -2}, {0, -8}, {0, -1}, {8, -5}, {0, 10}, {0, -1}, {0, 0}}
	var nodes [len(locations)]*mapgenRankedRoom
	t.Cleanup(func() {
		mapgenRooms = old
		for _, room := range nodes {
			if room != nil {
				mapgenReleaseDecorRecord(unsafe.Pointer(room))
			}
		}
	})
	for i := len(locations) - 1; i >= 0; i-- {
		raw, _ := mapgenAllocateDecorRecord(mapgenRoomRecordSize)
		if raw == nil {
			t.Fatal("could not allocate closest room fixture")
		}
		room := (*mapgenRankedRoom)(raw)
		nodes[i] = room
		room.kind = 1
		if i == 10 {
			room.kind = 2
		}
		*(*[2]int32)(unsafe.Pointer(&room.grid)) = locations[i]
		if !mapgenRooms.add((*mapgenRoomListNode)(raw), mapgenRegisterListedRoom) {
			t.Fatal("could not register closest room fixture")
		}
	}

	room := mapgenClosestPrefabRooms(mapgenPrefabExit5262F0{}, mapgenPrefabNorth5262F0)
	want := [...]int{7, 5, 2, 1, 3, 6}
	for idx, original := range want {
		if room != nodes[original] {
			t.Fatalf("candidate %d = %p, want original room %d at %p", idx, room, original, nodes[original])
		}
		room = mapgenRankedRoomByToken(room.prefabNext)
	}
	if room != nil {
		t.Fatalf("candidate chain continued past its six slots at %p", room)
	}
}
