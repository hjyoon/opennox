package legacy

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_offer_remove_original_v1.json.gz
var tradeOfferRemoveOriginal []byte

type tradeOfferRemoveCase struct {
	Input struct {
		Name    string
		Players [2]bool
		Lists   [2][]struct {
			Name string
			Code uint32
		}
		Code      uint32
		Flags     [2]uint32
		Before    [2]uint32
		Repriced  [2]uint32
		Insertion bool
	}
	Observed struct {
		Result int
		Totals [2]uint32
		Flags  [2]uint32
		Lists  [2][]string
		Events []struct {
			Service   string
			Recipient string
			Item      string
			Flags     [2]uint32
			Totals    [2]uint32
		}
	}
}

func TestTradeOfferRemoveNativeCABI50FE20(t *testing.T) {
	rows := tradeOfferOriginalRows[tradeOfferRemoveCase](t, tradeOfferRemoveOriginal,
		"7b040e08664bc8fc52e552ccc238acef68950c60aeb528cca897384aa65fbd52", 360)
	for route := range 2 {
		for _, row := range rows {
			if row.Input.Players[0] || row.Input.Players[1] {
				t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
					checkTradeOfferRemoveCase(t, route, row)
				})
			}
		}
	}
}

func checkTradeOfferRemoveCase(t *testing.T, route int, row tradeOfferRemoveCase) {
	t.Helper()
	in, want := row.Input, row.Observed
	world, session, parties := newTradeOfferFixture(t, in.Players)
	nodes := make(map[*server.TradeItem]string)
	items := make(map[*server.Object]string)
	var original [2][]string
	for side, values := range in.Lists {
		for _, value := range slices.Backward(values) {
			node := addTradeOfferFixture(t, world, session, parties[side], value.Code)
			nodes[node], items[node.Item0] = value.Name, value.Name
		}
		for _, value := range values {
			original[side] = append(original[side], value.Name)
		}
	}
	if in.Players[0] != in.Players[1] {
		session.Field16 = 1
		for index, player := range in.Players {
			if !player && !in.Insertion {
				// An unavailable merchant price represents a rejected child insertion.
				parties[index].InitData = nil
			}
		}
	}
	session.Field24, session.Field28 = in.Flags[0], in.Flags[1]
	session.Field40, session.Field44 = in.Before[0], in.Before[1]
	side := func(name string) int { return slices.Index([]string{"first", "second"}, name) }
	assertLists := func(wantLists [2][]string) {
		t.Helper()
		for slot, head := range [...]*server.TradeItem{session.Field32, session.Field36} {
			var got []string
			var previous *server.TradeItem
			for node := head; node != nil; node = node.Field8 {
				label, known := nodes[node]
				if node.Field12 != previous || !known || len(got) >= len(nodes) {
					t.Fatalf("bad offer links on side %d at %p", slot, node)
				}
				got = append(got, label)
				previous = node
			}
			if !slices.Equal(got, wantLists[slot]) {
				t.Fatalf("list %d = %v, want %v", slot, got, wantLists[slot])
			}
		}
	}
	var removedNode *server.TradeItem
	var removedItem *server.Object
	for _, event := range want.Events {
		if event.Service == "free_node" {
			for node, label := range nodes {
				if label == event.Item {
					removedNode, removedItem = node, node.Item0
					break
				}
			}
		}
	}
	var gotEvents, wantEvents []string
	game4ConsumerServices(t, game4TradeServices{
		world: world,
		reprice: func(got *server.TradeSession) {
			if got != session || [2]uint32{session.Field24, session.Field28} != in.Flags {
				t.Fatal("repricing called after approval reset or for a different session")
			}
			assertLists(want.Lists)
			gotEvents = append(gotEvents, "reprice")
			session.Field40, session.Field44 = in.Repriced[0], in.Repriced[1]
		},
		p2p: func() server.TradeP2PRuntime50F3A0 {
			return server.TradeP2PRuntime50F3A0{
				PutInventory: func(owner, item *server.Object) {
					assertLists(original)
					gotEvents = append(gotEvents, fmt.Sprintf("put:%d:%s", slices.Index(parties[:], owner)+1, items[item]))
				},
				Send: func(player *server.Player, packet []byte) {
					assertLists(want.Lists)
					if removedNode == nil || removedNode.Item0 != removedItem {
						t.Fatal("removed node was released before its notifications")
					}
					gotEvents = append(gotEvents, fmt.Sprintf("send:%d:%x", player.Index(), packet))
				},
			}
		},
	})
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		assertLists(original)
		code := tradeOfferStockCode(t, packet, related, remove, sequence)
		if session.Field20 == nil || session.Field20.Item0 != removedItem || code != uint16(in.Code) {
			t.Fatal("stock notification did not use the removed item")
		}
		gotEvents = append(gotEvents, fmt.Sprintf("stock:%d", index))
		return 1
	}
	for _, event := range want.Events {
		index := side(event.Recipient)
		var packet []byte
		switch event.Service {
		case "reprice":
			wantEvents = append(wantEvents, "reprice")
		case "put_inventory":
			wantEvents = append(wantEvents, fmt.Sprintf("put:%d:%s", index+1, event.Item))
		case "send_stock":
			wantEvents = append(wantEvents, fmt.Sprintf("stock:%d", index+1))
		case "balance":
			packet = make([]byte, 14)
			packet[0], packet[1] = 0xc9, 6
			if !in.Players[1-index] {
				binary.LittleEndian.PutUint32(packet[6:10], event.Totals[1-index]-event.Totals[index])
			}
		case "remove":
			packet = []byte{0xc9, 5, 0, 0}
			binary.LittleEndian.PutUint16(packet[2:], uint16(in.Code))
		case "status":
			packet = []byte{0xc9, 3, 0}
			if event.Flags[index] != 0 {
				packet[2] |= 1
			}
			if event.Flags[1-index] != 0 {
				packet[2] |= 2
			}
		}
		if packet != nil {
			wantEvents = append(wantEvents, fmt.Sprintf("send:%d:%x", index+1, packet))
		}
	}
	if got := cfixture.CallTradeOfferRemove50FE20(route, session, in.Code); got != want.Result {
		t.Fatalf("remove result = %d, want %d", got, want.Result)
	}
	if !slices.Equal(gotEvents, wantEvents) {
		t.Fatalf("events = %v, want %v", gotEvents, wantEvents)
	}
	assertLists(want.Lists)
	if [2]uint32{session.Field24, session.Field28} != want.Flags ||
		[2]uint32{session.Field40, session.Field44} != want.Totals {
		t.Fatal("final approval flags or balances differ from stock")
	}
}

func TestTradeOfferRemoveNativeCallbackCanRelease50FE20(t *testing.T) {
	world, session, parties := newTradeOfferFixture(t, [2]bool{true, true})
	addTradeOfferFixture(t, world, session, parties[0], 17)
	sends := 0
	game4ConsumerServices(t, game4TradeServices{world: world, p2p: func() server.TradeP2PRuntime50F3A0 {
		return server.TradeP2PRuntime50F3A0{Send: func(*server.Player, []byte) {
			sends++
			if !world.ReleaseTradeSessionNative510000(session) {
				t.Fatal("release during removal notification")
			}
		}}
	}})
	if got := cfixture.CallTradeOfferRemove50FE20(1, session, 17); got != 0 || sends != 1 || world.IsTradeSessionNative(session) {
		t.Fatalf("interrupted removal = %d, sends = %d, live session = %t", got, sends, world.IsTradeSessionNative(session))
	}
}
