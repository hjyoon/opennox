package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
)

func requireExplicitHighServerPointer(t *testing.T, ptr unsafe.Pointer) {
	t.Helper()
	if ptr == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32) {
		t.Fatalf("explicit native owner pointer %p does not exercise host address width", ptr)
	}
}

func TestServerExplicitHighOwnerIsolation(t *testing.T) {
	ordinary, high := New(nil, nil, strman.New()), NewHigh(nil, nil, strman.New())
	t.Cleanup(ordinary.Close)
	t.Cleanup(high.Close)
	t.Cleanup(ordinary.Types.Free)
	t.Cleanup(high.Types.Free)
	if ordinary.highAddresses || ordinary.Types.highAddresses || ordinary.Teams.highAddresses || ordinary.tradeNative.highAddresses {
		t.Fatal("ordinary server inherited another owner's high allocation policy")
	}
	if !high.highAddresses || !high.Types.highAddresses || !high.Teams.highAddresses || !high.tradeNative.highAddresses {
		t.Fatal("explicit server policy did not reach its native owners")
	}
	if ordinary.Spells.Dur.s != ordinary || high.Spells.Dur.s != high {
		t.Fatal("duration owner did not retain its server policy source")
	}
	requireExplicitHighServerPointer(t, unsafe.Pointer(&high.Players.list[31]))
	requireExplicitHighServerPointer(t, high.Teams.Arr[1].C())
	ordinary.Players.list[31].GoldVal, high.Players.list[31].GoldVal = 17, 29
	for _, srv := range []*Server{ordinary, high} {
		srv.Types.Free()
		health := 43
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "ExplicitHighOwnerHealth", Health: &health}); err != nil {
			t.Fatal(err)
		}
		if srv.Spells.Dur.SpellCreateDurations4FE850() != 1 {
			t.Fatal("duration allocation class was not initialized")
		}
		t.Cleanup(srv.Spells.Dur.Free)
	}
	requireExplicitHighServerPointer(t, high.Types.ByID("ExplicitHighOwnerHealth").Health().C())
	a, b := ordinary.Spells.Dur.NewRaw(), high.Spells.Dur.NewRaw()
	if a == nil || b == nil || a == b {
		t.Fatal("duration owners did not allocate distinct records")
	}
	requireExplicitHighServerPointer(t, b.C())
	a.Spell, b.Spell = 19, 37
	high.Spells.Dur.alloc.FreeObjectFirst(b)
	reused := high.Spells.Dur.NewRaw()
	if reused != b || reused.Spell != 0 || a.Spell != 19 {
		t.Fatal("high duration reuse failed to clear storage or affected another owner")
	}
	first := ordinary.newTradeSessionNative50E870(nil)
	second := high.newTradeSessionNative50E870(nil)
	if first == nil || second == nil || first == second {
		t.Fatal("trade owners did not allocate distinct sessions")
	}
	requireExplicitHighServerPointer(t, unsafe.Pointer(second))
	first.Field40, second.Field40 = 53, 71
	if !high.ReleaseTradeSessionNative510000(second) || high.IsTradeSessionNative(second) {
		t.Fatal("high trade owner did not release its record")
	}
	if first.Field40 != 53 || ordinary.Players.list[31].GoldVal != 17 || high.Players.list[31].GoldVal != 29 {
		t.Fatal("native ownership crossed server boundaries")
	}
}

func TestAIPathExplicitHighStorageIsolation(t *testing.T) {
	var ordinary, high serverAIPaths
	ordinary.Init(new(Server))
	high.InitHigh(new(Server))
	t.Cleanup(ordinary.Free)
	t.Cleanup(high.Free)
	a, b := ordinary.NewVisitNode(), high.NewVisitNode()
	if a == nil || b == nil || a == b {
		t.Fatal("path owners did not allocate distinct visit nodes")
	}
	requireExplicitHighServerPointer(t, unsafe.Pointer(b))
	requireExplicitHighServerPointer(t, unsafe.Pointer(unsafe.SliceData(high.points)))
	a.X0, b.X0 = 11, 23
	high.allocVisit.FreeObjectFirst(b)
	reused := high.NewVisitNode()
	if reused != b || reused.X0 != 0 || a.X0 != 11 {
		t.Fatal("path record reuse failed or modified another owner")
	}
	high.points[1023].X, ordinary.points[1023].X = 31, 47
	high.Free()
	if high.Valid() || high.points != nil || !ordinary.Valid() || ordinary.points[1023].X != 47 {
		t.Fatal("path storage release crossed owner boundaries")
	}
}
