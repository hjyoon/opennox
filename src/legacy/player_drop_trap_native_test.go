package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDropATrapNative10002030UsesNativeLinksAndFirstMatchingItem(t *testing.T) {
	wrong := &server.Object{ObjClass: object.Class(0x1010000)}
	trap := &server.Object{ObjClass: object.Class(0x110000)}
	later := &server.Object{ObjClass: object.Class(0x110001)}
	wrong.InvNextItem = trap
	trap.InvNextItem = later
	player := &server.Player{}
	wantPos := types.Ptf(124.5, 87.25)
	player.SetPos3632(wantPos)
	update := &server.PlayerUpdateData{Player: player}
	owner := &server.Object{UpdateData: unsafe.Pointer(update), InvFirstItem: wrong}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= 0xffffffff {
		t.Fatal("expected native owner pointer above 4 GiB")
	}
	calls := 0
	if !playerDropATrapNative10002030(owner, func(gotOwner, gotItem *server.Object, pos *types.Pointf) {
		calls++
		if gotOwner != owner || gotItem != trap || *pos != wantPos {
			t.Fatalf("drop = (%p, %p, %v), want (%p, %p, %v)", gotOwner, gotItem, *pos, owner, trap, wantPos)
		}
	}) || calls != 1 {
		t.Fatalf("drop calls = %d, want 1", calls)
	}
}

func TestPlayerDropATrapNative10002030Gates(t *testing.T) {
	trap := &server.Object{ObjClass: object.Class(0x110000)}
	player := &server.Player{}
	update := &server.PlayerUpdateData{Player: player}
	owner := &server.Object{UpdateData: unsafe.Pointer(update), InvFirstItem: trap}
	failDrop := func(*server.Object, *server.Object, *types.Pointf) { t.Fatal("unexpected drop") }
	for _, status := range []uint32{1, 2, 3} {
		player.Field3680 = status
		if playerDropATrapNative10002030(owner, failDrop) {
			t.Fatalf("status %d permitted drop", status)
		}
	}
	player.Field3680 = 0
	update.State = server.PlayerState1
	if playerDropATrapNative10002030(owner, failDrop) {
		t.Fatal("state 1 permitted drop")
	}
	update.State = server.PlayerState0
	owner.InvFirstItem = nil
	if playerDropATrapNative10002030(owner, failDrop) {
		t.Fatal("empty inventory permitted drop")
	}
	if playerDropATrapNative10002030(nil, failDrop) ||
		playerDropATrapNative10002030(&server.Object{}, failDrop) {
		t.Fatal("missing owner/update permitted drop")
	}
}
