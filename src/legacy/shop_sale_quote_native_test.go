package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_sale_quote_original_v1.json.gz
var shopSaleQuoteOriginal []byte

type shopSaleQuoteCase struct {
	Input struct {
		Name       string
		Placement  string
		Quest      bool
		Glyph      bool
		Merchant   bool
		NetCode    uint32 `json:"net_code"`
		PlayerCode uint32 `json:"player_code"`
		Receiver   byte
		Queue      uint32
	}
	Observed struct {
		Events []struct {
			Service  string
			Item     string
			Key      string
			Receiver int
			Packet   string
			Remove   int
		}
		ReturnWord uint32 `json:"return_word"`
	}
}

func TestShopSaleQuoteNativeCABI5109C0(t *testing.T) {
	rows := tradeOfferOriginalRows[shopSaleQuoteCase](t, shopSaleQuoteOriginal,
		"6a00913ed435ee344a9bb1666a4e27acbe0c1b959dbf91b2308df37ae9a1e216", 96)
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkShopSaleQuoteCase(t, route, row)
			})
		}
	}
}

func checkShopSaleQuoteCase(t *testing.T, route int, row shopSaleQuoteCase) {
	t.Helper()
	in, expected := row.Input, row.Observed
	world, session, parties := newTradeOfferFixture(t, [2]bool{true, false})
	session.Field16 = 1
	if !in.Merchant {
		session = nil
	}
	player := parties[0]
	player.NetCode = in.PlayerCode
	player.UpdateDataPlayer().Player.PlayerInd = in.Receiver
	const price = 27
	first := game4ConsumerValue(t, server.Object{NetCode: in.NetCode ^ 0x6789, Worth: price})
	last := game4ConsumerValue(t, server.Object{NetCode: in.NetCode ^ 0x6789, Worth: price})
	first.InvNextItem = last
	if in.Placement != "empty" {
		player.InvFirstItem = first
	}
	labels := map[*server.Object]string{first: "first", last: "last"}
	switch in.Placement {
	case "first":
		first.NetCode = in.NetCode
	case "last":
		last.NetCode = in.NetCode
	}
	var events, want []string
	recordRejection := func(actor *server.Object, event string) {
		if actor != player || actor.NetCode != in.PlayerCode {
			t.Fatal("wrong sale rejection owner")
		}
		events = append(events, event)
	}
	game4ConsumerServices(t, game4TradeServices{
		world: world,
		sell: func() server.ShopSellRuntime5109C0 {
			return server.ShopSellRuntime5109C0{
				ItemIsQuest: func(item *server.Object) bool {
					events = append(events, "quest:"+labels[item])
					return in.Quest
				},
				ItemIsGlyph: func(item *server.Object) bool {
					if in.Quest || labels[item] != in.Placement {
						t.Fatal("wrong glyph check")
					}
					return in.Glyph
				},
				ReportCannotSellQuest: func(actor *server.Object) { recordRejection(actor, "line:CantSellQuestItem") },
				ReportCannotSellItem:  func(actor *server.Object) { recordRejection(actor, "line:CantSellItem") },
				PlayRejectSound: func(actor *server.Object) uintptr {
					recordRejection(actor, "sound")
					return 0
				},
			}
		},
	})
	world.NetSendPacketXxx = func(index int, data []byte, related *server.Object, remove, sequence int) int {
		if len(expected.Events) == 0 || expected.Events[len(expected.Events)-1].Service != "quote" {
			t.Fatalf("unexpected sale quote %x", data)
		}
		event := expected.Events[len(expected.Events)-1]
		packet, err := hex.DecodeString(event.Packet)
		if err != nil {
			t.Fatal(err)
		}
		binary.LittleEndian.PutUint32(packet[4:], price)
		if index != event.Receiver || !bytes.Equal(data, packet) || related != nil || remove != event.Remove || sequence != 0 {
			t.Fatalf("sale quote = %d %x %p %d %d, want %d %x", index, data, related, remove, sequence, event.Receiver, packet)
		}
		events = append(events, "quote")
		return int(int32(in.Queue))
	}
	wantWord := expected.ReturnWord
	for _, event := range expected.Events {
		switch event.Service {
		case "quest_flag":
			want = append(want, "quest:"+event.Item)
		case "line":
			want = append(want, "line:"+event.Key)
		case "sound":
			want = append(want, "sound")
			wantWord = 0
		case "quote":
			want = append(want, "quote")
		}
	}
	beforePlayer, beforeFirst, beforeLast := *player, *first, *last
	if got := callShopSaleQuote5109C0(route, player, session, in.NetCode); got != wantWord || !slices.Equal(events, want) {
		t.Fatalf("sale quote = %#x, %v, want %#x, %v", got, events, wantWord, want)
	}
	if *player != beforePlayer || *first != beforeFirst || *last != beforeLast {
		t.Fatal("sale quote changed a player or inventory object")
	}
}
