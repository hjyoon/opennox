package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_repair_quote_original_v1.json.gz
var shopRepairQuoteOriginal []byte

type shopRepairQuoteCase struct {
	Input struct {
		Name      string
		Placement string
		Health    []uint16
		Wand      [4]uint32
		NetCode   uint32 `json:"net_code"`
		Receiver  byte
		SendWord  uint32 `json:"send_word"`
		Shop      bool
	}
	Observed struct {
		Events []struct {
			Service  string
			Receiver int
			Packet   string
			Remove   int
			ID       int
		}
		ReturnWord uint32 `json:"return_word"`
	}
}

func TestShopRepairQuoteNativeCABI5108D0(t *testing.T) {
	rows := tradeOfferOriginalRows[shopRepairQuoteCase](t, shopRepairQuoteOriginal,
		"4f4eee99364a5d53db3a532126f3467857a0eeeafa0f3fef60b91562268f1075", 144)
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkShopRepairQuoteCase(t, route, row)
			})
		}
	}
}

func checkShopRepairQuoteCase(t *testing.T, route int, row shopRepairQuoteCase) {
	t.Helper()
	in, want := row.Input, row.Observed
	world, session, parties := newTradeOfferFixture(t, [2]bool{true, false})
	session.Field16 = 1
	if !in.Shop {
		session = nil
	}
	playerUnit := parties[0]
	playerUnit.UpdateDataPlayer().Player.PlayerInd = in.Receiver
	first := game4ConsumerValue(t, server.Object{NetCode: in.NetCode ^ 0x6789})
	last := game4ConsumerValue(t, server.Object{NetCode: in.NetCode ^ 0x6789})
	first.InvNextItem = last
	if in.Placement != "empty" {
		playerUnit.InvFirstItem = first
	}
	var matched *server.Object
	if in.Placement == "first" {
		matched = first
	} else if in.Placement == "last" {
		matched = last
	}
	if matched != nil {
		matched.NetCode = in.NetCode
		matched.ObjClass = object.Class(in.Wand[0])
		matched.ObjSubClass = object.SubClass(in.Wand[1])
		matched.UseData.Ptr = unsafe.Pointer(game4ConsumerValue(t, server.WandUseData{
			Charge: uint8(in.Wand[2]), MaxCharge: uint8(in.Wand[3]),
		}))
		if in.Health != nil {
			matched.HealthData = game4ConsumerValue(t, server.HealthData{Cur: in.Health[0], Max: in.Health[1]})
		}
	}
	var expectedQuote, expectedReject bool
	for _, e := range want.Events {
		expectedQuote = expectedQuote || e.Service == "quote"
		expectedReject = expectedReject || e.Service == "sound"
	}
	rejections, quotes := 0, 0
	game4ConsumerServices(t, game4TradeServices{
		world: world,
		repair: func() server.ShopRepairRuntime5108D0 {
			return server.ShopRepairRuntime5108D0{
				RepairCoefficient: 1,
				PlayRejectSound: func(owner *server.Object) {
					if owner != playerUnit {
						t.Fatalf("reject sound owner %p, want %p", owner, playerUnit)
					}
					rejections++
				},
			}
		},
	})
	world.NetSendPacketXxx = func(index int, data []byte, related *server.Object, remove, sequence int) int {
		if !expectedQuote {
			t.Fatalf("unexpected quote %x", data)
		}
		quotes++
		event := want.Events[len(want.Events)-1]
		packet, err := hex.DecodeString(event.Packet)
		if err != nil {
			t.Fatal(err)
		}
		// The original price child had a controlled return. A zero-worth
		// item through the actual pricing service yields the minimum cost.
		binary.LittleEndian.PutUint32(packet[4:], 1)
		if index != event.Receiver || !bytes.Equal(data, packet) || related != nil || remove != event.Remove || sequence != 0 {
			t.Fatalf("quote = %d %x %p %d %d, want %d %x", index, data, related, remove, sequence, event.Receiver, packet)
		}
		return int(int32(in.SendWord))
	}
	beforePlayer, beforeFirst, beforeLast := *playerUnit, *first, *last
	got := callShopRepairQuote5108D0(route, playerUnit, session, in.NetCode)
	wantResult := uint32(0)
	if expectedQuote {
		wantResult = want.ReturnWord
	}
	if got != wantResult || (quotes == 1) != expectedQuote || (rejections == 1) != expectedReject || quotes > 1 || rejections > 1 {
		t.Fatalf("quote result = %#x, quote events %d, reject events %d, want %#x/%t/%t",
			got, quotes, rejections, wantResult, expectedQuote, expectedReject)
	}
	if *playerUnit != beforePlayer || *first != beforeFirst || *last != beforeLast {
		t.Fatal("quote changed a player or inventory object")
	}
}
