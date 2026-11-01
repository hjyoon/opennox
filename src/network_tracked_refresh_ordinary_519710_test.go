package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestTrackedRefreshOrdinaryBindings519710(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	previous := legacy.GetServer
	legacy.GetServer = func() legacy.Server { return s }
	t.Cleanup(func() { legacy.GetServer = previous })
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	player := base.Players.ByIndRaw(7)
	player.PlayerInd = 7
	update.Player = player
	var nodes []*server.MinimapItem
	for i := 0; i < 61; i++ {
		node, free := alloc.New(server.MinimapItem{})
		t.Cleanup(free)
		nodes = append(nodes, node)
	}
	t.Cleanup(func() { player.Field4580 = nil })
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(nodes[0]), unsafe.Pointer(nodes[60])} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatal("production binding allocations must exceed4GiB")
		}
	}
	// These strict-boundary outcomes also occur in the frozen original fixture.
	for _, c := range []struct {
		name        string
		count       int
		last, frame uint32
		want        int
	}{
		{"empty", 0, 80, 100, 0},
		{"one-equal", 1, 80, 140, 0},
		{"one-next", 1, 80, 141, 1},
		{"four-equal", 4, 80, 95, 0},
		{"four-next", 4, 80, 96, 1},
		{"sixty-equal", 60, 80, 81, 0},
		{"sixty-next", 60, 80, 82, 1},
		{"sixtyone", 61, 80, 80, 1},
		{"wrapped", 4, 0xfffffff0, 0, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			player.Field4580 = nil
			if c.count != 0 {
				for i := 0; i < c.count; i++ {
					nodes[i].Field8 = nodes[(i+1)%c.count]
				}
				player.Field4580 = nodes[0]
			}
			base.SetFrame(c.frame)
			update.Field67 = c.last
			if got := player.MinimapTrackCount(); got != c.count {
				t.Fatalf("actual nodes=%d", got)
			}
			for _, mode := range []struct {
				name string
				call func() int
			}{
				{"C", func() int { return legacy.Sub_519710(unsafe.Pointer(update)) }},
				{"Go", func() int { return int(legacy.TrackedObjectRefresh519710(update)) }},
				{"ordinary", func() int {
					if s.netTrackedObjectRefreshDueNative519710(update) {
						return 1
					}
					return 0
				}},
			} {
				t.Run(mode.name, func(t *testing.T) {
					if got := mode.call(); got != c.want {
						t.Fatalf("got=%d want=%d", got, c.want)
					}
				})
			}
			if update.Field67 != c.last || update.Player != player || base.Frame() != c.frame || player.MinimapTrackCount() != c.count {
				t.Fatal("production read-only entry changed state")
			}
		})
	}
}

func TestTrackedRefreshOrdinaryNilPolicy519710(t *testing.T) {
	s := &Server{}
	if s.netTrackedObjectRefreshDueNative519710(nil) || s.netTrackedObjectRefreshDueNative519710(&server.PlayerUpdateData{}) {
		t.Fatal("ordinary nil policy changed")
	}
}
