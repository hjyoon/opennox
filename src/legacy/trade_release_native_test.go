package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestTradeReleaseNativeCABI510000(t *testing.T) {
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			world := server.NewHigh(nil, nil, strman.New())
			defer world.Close()
			game4ConsumerServices(t, game4TradeServices{world: world})
			a, _ := game4ConsumerPlayer(t, 0)
			b, _ := game4ConsumerPlayer(t, 0)
			older := world.NewTradeSessionNative50E870()
			session, result := world.StartP2PTradeNative50EF10(a, b, 35, nil)
			if result != server.TradeP2PStartComplete50EF10 || session == nil || older.Field60 != session {
				t.Fatalf("start trade: %v", result)
			}
			first := game4ConsumerValue(t, server.Object{NetCode: 0x1acb, Worth: 41})
			second := game4ConsumerValue(t, server.Object{NetCode: 0xb78d, Worth: 93})
			for _, row := range []struct{ owner, item *server.Object }{{a, first}, {a, second}, {b, first}} {
				if !world.AddP2PTradeOfferNative50F820(session, row.owner, row.item, server.TradeP2PRuntime50F3A0{}) {
					t.Fatal("offer fixture")
				}
			}
			left, right := session.Field32, session.Field36
			if unsafe.Sizeof(uintptr(0)) == 8 &&
				(uintptr(unsafe.Pointer(session)) <= 0xffffffff || uintptr(unsafe.Pointer(left)) <= 0xffffffff) {
				t.Fatal("session and node must exercise high addresses")
			}
			if got := tradeOfferReleaseFixtureCall(route, left); got != 0 || session.Field32 != nil || session.Field36 != right {
				t.Fatalf("offer release returned %d or changed an unrelated list", got)
			}
			if first.Worth != 41 || second.Worth != 93 || !world.IsTradeSessionNative(session) {
				t.Fatal("offer release changed an object or the session owner")
			}
			if got := tradeOfferReleaseFixtureCall(route, nil); got != 0 {
				t.Fatalf("empty offers returned %d", got)
			}
			a.UpdateDataPlayer().Trade70 = nil
			b.UpdateDataPlayer().Trade70 = nil
			tradeReleaseFixtureCall(route, session)
			if world.IsTradeSessionNative(session) || !world.IsTradeSessionNative(older) ||
				older.Field60 != nil || first.NetCode != 0x1acb || second.NetCode != 0xb78d {
				t.Fatal("session release changed an adjacent owner or freed offered objects")
			}
			tradeReleaseFixtureCall(route, session)
		})
	}
}

func TestTradeReleasePE32Pools510000(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 4 {
		t.Skip("fixed-width trade pools run on 32-bit hosts")
	}
	handles.Init()
	t.Cleanup(handles.Release)
	sessions := alloc.NewClass("TestTradeReleaseSessions", unsafe.Sizeof(server.TradeSession{}), 2)
	defer sessions.Free()
	items := alloc.NewClass("TestTradeReleaseItems", unsafe.Sizeof(server.TradeItem{}), 3)
	defer items.Free()
	head := (*server.TradeSession)(sessions.NewObject())
	older := (*server.TradeSession)(sessions.NewObject())
	head.Field56, older.Field60 = older, head
	offer := (*server.TradeItem)(items.NewObject())
	other := (*server.TradeItem)(items.NewObject())
	tail := (*server.TradeItem)(items.NewObject())
	head.Field32, head.Field36 = offer, other
	offer.Field8, tail.Field12 = tail, offer
	restore := tradeReleaseFixtureUsePE32Pools(sessions, items, head)
	defer restore()
	game4ConsumerServices(t, game4TradeServices{world: &server.Server{}})
	if got := tradeOfferReleaseFixtureCall(1, offer); got != 0 ||
		items.HasActiveObject(unsafe.Pointer(offer)) || items.HasActiveObject(unsafe.Pointer(tail)) ||
		!items.HasActiveObject(unsafe.Pointer(other)) {
		t.Fatal("offer-path allocator ownership")
	}
	head.Field32 = nil
	tradeReleaseFixtureCall(0, head)
	if sessions.HasActiveObject(unsafe.Pointer(head)) || items.HasActiveObject(unsafe.Pointer(other)) ||
		tradeReleaseFixturePE32Head() != uintptr(unsafe.Pointer(older)) ||
		!sessions.HasActiveObject(unsafe.Pointer(older)) || older.Field60 != nil {
		t.Fatal("legacy pool release or linked head")
	}
	tradeReleaseFixtureCall(1, head)
}
