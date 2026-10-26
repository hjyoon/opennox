package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestServerCloseReleasesNativeTables(t *testing.T) {
	for cycle := 0; cycle < 3; cycle++ {
		s := New(nil, nil, strman.New())
		playerTable := unsafe.Pointer(&s.Players.list[0])
		teamTable := unsafe.Pointer(&s.Teams.Arr[0])
		t.Logf("cycle %d: player bytes=%d team bytes=%d", cycle,
			uintptr(len(s.Players.list))*unsafe.Sizeof(s.Players.list[0]),
			uintptr(len(s.Teams.Arr))*unsafe.Sizeof(s.Teams.Arr[0]))
		s.Close()
		if getServer(s.handle) != nil {
			t.Error("closed server is still registered")
		}
		for name, ptr := range map[string]unsafe.Pointer{"players": playerTable, "teams": teamTable} {
			// FreePtr checks its ownership map before touching C memory. A
			// released allocation must reject this second release; if Close
			// leaked it, this probe also frees the failed fixture's allocation.
			var released bool
			func() {
				defer func() { released = recover() != nil }()
				alloc.FreePtr(ptr)
			}()
			if !released {
				t.Errorf("cycle %d: Close retained the native %s allocation", cycle, name)
			}
		}
		if s.Players.list != nil || s.Teams.Arr != nil {
			t.Error("closed native tables remain addressable")
		}
		s.Close() // Repeated shutdown must not release the same C memory twice.
	}
}
