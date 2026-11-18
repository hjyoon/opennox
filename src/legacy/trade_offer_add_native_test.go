package legacy

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_offer_eligibility_original_v2.json.gz
var tradeOfferEligibilityOriginal []byte

//go:embed testdata/trade_offer_add_original_v1.json.gz
var tradeOfferAddOriginal []byte

func TestTradeOfferEligibilityNativeCABI50FD60(t *testing.T) {
	type eligibilityCase struct {
		Input struct {
			Name       string
			Existing   []uint16
			NewType    uint16       `json:"new_type"`
			ClassValue object.Class `json:"class_value"`
		}
		Observed int
	}
	rows := tradeOfferOriginalRows[eligibilityCase](t, tradeOfferEligibilityOriginal,
		"75f810d9ab39d5d8cee5d012dfcc12b11ee270c631d9510b2c8201b533fa305d", 462)
	for route := range 2 {
		for _, row := range rows {
			in := row.Input
			t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
				var head *server.TradeItem
				for _, typ := range slices.Backward(in.Existing) {
					item := game4ConsumerValue(t, server.Object{TypeInd: typ})
					head = game4ConsumerValue(t, server.TradeItem{Item0: item, Field8: head})
				}
				item := game4ConsumerValue(t, server.Object{TypeInd: in.NewType, ObjClass: in.ClassValue})
				if got := callTradeOfferEligible50FD60(route, head, item); got != row.Observed {
					t.Fatalf("item eligibility = %d, want %d", got, row.Observed)
				}
			})
		}
	}
}

type tradeOfferAddCase struct {
	Input struct {
		Name     string
		Players  [2]bool
		Actor    string
		TypeID   uint16 `json:"type_id"`
		Lists    [2][]uint16
		Flags    [2]uint32
		Before   [2]uint32
		Repriced [2]uint32
	}
	Observed struct {
		Result int
		Totals [2]uint32
		Flags  [2]uint32
		Events []tradeOfferAddEvent
	}
}

type tradeOfferAddEvent struct {
	Service   string
	Recipient string
	Offering  string
	Flags     [2]uint32
	Totals    [2]uint32
	Granted   bool
}

func TestTradeOfferAddNativeCABI50F820(t *testing.T) {
	rows := tradeOfferOriginalRows[tradeOfferAddCase](t, tradeOfferAddOriginal,
		"1769a5bdc62831bf210b473c790a5fdab6e773459a88f4a17ccbd396539f740c", 240)
	for route := range 2 {
		for _, row := range rows {
			if !row.Input.Players[0] && !row.Input.Players[1] ||
				slices.ContainsFunc(row.Observed.Events, func(event tradeOfferAddEvent) bool {
					return event.Service == "allocate" && !event.Granted
				}) {
				continue
			}
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkTradeOfferAddCase(t, route, row)
			})
		}
	}
}

func checkTradeOfferAddCase(t *testing.T, route int, row tradeOfferAddCase) {
	t.Helper()
	in, want := row.Input, row.Observed
	world, session, parties := newTradeOfferFixture(t, in.Players)
	for side, values := range in.Lists {
		for i, typ := range slices.Backward(values) {
			node := addTradeOfferFixture(t, world, session, parties[side], uint32(40+i))
			node.Item0.TypeInd = typ
		}
	}
	before := [2]*server.TradeItem{session.Field32, session.Field36}
	if in.Players[0] != in.Players[1] {
		session.Field16 = 1
	}
	session.Field24, session.Field28 = in.Flags[0], in.Flags[1]
	session.Field40, session.Field44 = in.Before[0], in.Before[1]
	side := func(name string) int { return slices.Index([]string{"first", "second"}, name) }
	ownerSide := side(in.Actor)
	actor := game4ConsumerValue(t, server.Object{})
	if ownerSide >= 0 {
		actor = parties[ownerSide]
	}
	const code = uint32(0x5a10)
	const price = uint32(37)
	item := game4ConsumerValue(t, server.Object{TypeInd: in.TypeID, Worth: price, NetCode: code})
	head := func(side int) *server.TradeItem {
		if side == 0 {
			return session.Field32
		}
		return session.Field36
	}
	var gotEvents, wantEvents []string
	game4ConsumerServices(t, game4TradeServices{
		world: world,
		reprice: func(got *server.TradeSession) {
			if got != session || ownerSide < 0 {
				t.Fatal("repriced an unrelated trade")
			}
			node := head(ownerSide)
			if node == nil || node == before[ownerSide] || node.Item0 != item || node.Cost4 != price ||
				node.Field8 != before[ownerSide] || node.Field12 != nil || head(1-ownerSide) != before[1-ownerSide] {
				t.Fatal("repriced before linking the new offer")
			}
			if node.Field8 != nil && node.Field8.Field12 != node {
				t.Fatal("old head is missing its backward link")
			}
			if flags := [2]uint32{session.Field24, session.Field28}; flags != in.Flags {
				t.Fatalf("repriced after resetting approval flags: %v", flags)
			}
			gotEvents = append(gotEvents, "reprice")
			session.Field40, session.Field44 = in.Repriced[0], in.Repriced[1]
		},
		p2p: func() server.TradeP2PRuntime50F3A0 {
			return server.TradeP2PRuntime50F3A0{Send: func(player *server.Player, packet []byte) {
				gotEvents = append(gotEvents, fmt.Sprintf("send:%d:%x", player.Index(), packet))
			}}
		},
	})
	for _, event := range want.Events {
		index := side(event.Recipient)
		var packet []byte
		switch event.Service {
		case "reprice":
			wantEvents = append(wantEvents, "reprice")
		case "balance":
			packet = make([]byte, 14)
			packet[0], packet[1] = 0xc9, 6
			if !in.Players[1-index] {
				binary.LittleEndian.PutUint32(packet[6:10], event.Totals[1-index]-event.Totals[index])
			}
		case "item":
			packet = []byte{0xc9, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}
			if index == side(event.Offering) {
				packet[2] = 1
			}
			binary.LittleEndian.PutUint16(packet[3:5], in.TypeID)
			binary.LittleEndian.PutUint16(packet[5:7], uint16(code))
			binary.LittleEndian.PutUint32(packet[7:11], price)
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
	if got := callTradeOfferAdd50F820(route, session, actor, item); got != want.Result {
		t.Fatalf("add result = %d, want %d", got, want.Result)
	}
	if !slices.Equal(gotEvents, wantEvents) {
		t.Fatalf("events = %v, want %v", gotEvents, wantEvents)
	}
	if [2]uint32{session.Field24, session.Field28} != want.Flags ||
		[2]uint32{session.Field40, session.Field44} != want.Totals {
		t.Fatal("final approval flags or balances differ from stock")
	}
	if want.Result == 0 && [2]*server.TradeItem{session.Field32, session.Field36} != before {
		t.Fatal("rejected add changed an offer list")
	}
}
