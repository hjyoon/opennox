package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type minimapBroadcastLegacyServer417470 struct {
	Server
	srv *server.Server
}

func (s *minimapBroadcastLegacyServer417470) S() *server.Server {
	return s.srv
}

func TestMinimapBroadcast417470NativePointersPlayerSlotsAndFlags(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &minimapBroadcastLegacyServer417470{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	obj, freeObject := alloc.New(server.Object{})
	t.Cleanup(freeObject)
	defer srv.Players.Nox_xxx_netUnmarkMinimapSpec_417470(obj, math.MaxUint32)
	first := srv.Players.NewRaw(1001)
	inactive := srv.Players.NewRaw(1002)
	last := srv.Players.NewRaw(1003)
	if first == nil || inactive == nil || last == nil {
		t.Fatal("cannot allocate native player fixtures")
	}
	inactive.Active = 0
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"object": unsafe.Pointer(obj), "first player": first.C(), "last player": last.C(),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above 4 GiB", name, ptr)
			}
		}
	}

	if got := minimapMarkBroadcastCall4174B0(obj, 0x80000001); got != nil {
		t.Fatalf("mark result = %p, want terminal nil iterator", got)
	}
	minimapMarkBroadcastCall4174B0(obj, 2)
	for _, pl := range []*server.Player{first, last} {
		if !pl.MinimapTracks(obj) || pl.MinimapTrackCount() != 1 {
			t.Fatalf("player %d did not receive exactly one tracked native object", pl.Index())
		}
		if pl.Field4580.Field4 != obj || pl.Field4580.Field0 != 0x80000003 {
			t.Fatalf("player %d minimap object/flags = %p/%#x", pl.Index(), pl.Field4580.Field4, pl.Field4580.Field0)
		}
	}
	if inactive.Field4580 != nil {
		t.Fatal("inactive player received a minimap mark")
	}

	if got := minimapUnmarkBroadcastCall417470(obj, 1); got != nil {
		t.Fatalf("unmark result = %p, want terminal nil iterator", got)
	}
	for _, pl := range []*server.Player{first, last} {
		if !pl.MinimapTracks(obj) || pl.Field4580.Field0 != 0x80000002 {
			t.Fatalf("player %d lost the remaining minimap flags", pl.Index())
		}
	}
	minimapUnmarkBroadcastCall417470(obj, 0x80000002)
	for _, pl := range []*server.Player{first, last} {
		if pl.Field4580 != nil || pl.MinimapTracks(obj) {
			t.Fatalf("player %d still tracks the fully unmarked object", pl.Index())
		}
	}

	minimapMarkBroadcastCall4174B0(nil, 1)
	minimapUnmarkBroadcastCall417470(nil, 1)
	first.Active = 0
	last.Active = 0
	minimapMarkBroadcastCall4174B0(obj, 1)
	if first.Field4580 != nil || last.Field4580 != nil {
		t.Fatal("empty player list acquired a minimap mark")
	}
}
