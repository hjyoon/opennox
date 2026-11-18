package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_offer_transfer_original_v1.json.gz
var tradeOfferTransferOriginal []byte

//go:embed testdata/trade_offer_accept_original_v1.json.gz
var tradeOfferAcceptOriginal []byte

func tradeOfferOriginalRows[T any](t *testing.T, compressed []byte, digest string, cases int) []T {
	t.Helper()
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != digest {
		t.Fatalf("stock fixture digest = %s, want %s", got, digest)
	}
	var fixture struct{ Rows []T }
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != cases {
		t.Fatalf("stock fixture has %d cases, want %d", len(fixture.Rows), cases)
	}
	return fixture.Rows
}

func newTradeOfferFixture(t *testing.T, players [2]bool) (*server.Server, *server.TradeSession, [2]*server.Object) {
	t.Helper()
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.SetGame(oldFlags) })
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	if !world.Objs.InitHigh(12) {
		t.Fatal("object pool")
	}
	t.Cleanup(world.Objs.FreeObjects)
	world.FreeObjectTypes()
	t.Cleanup(world.FreeObjectTypes)
	t.Cleanup(world.TradeFree50E300)
	if err := world.Types.ReadObjectType(&things.Thing{Name: "NativeTradeOfferItem"}); err != nil {
		t.Fatal(err)
	}
	session := world.NewTradeSessionNative50E870()
	var participants [2]*server.Object
	for side, isPlayer := range players {
		unit, player := game4ConsumerPlayer(t, 0)
		player.PlayerInd = byte(side + 1)
		if isPlayer {
			unit.UpdateDataPlayer().Trade70 = session
		} else {
			unit.ObjClass = object.ClassMonster
			unit.InitData = unsafe.Pointer(game4ConsumerValue(t, server.ShopkeeperInitData{
				BuyMultiplier: 1, SellMultiplier: 1,
			}))
		}
		participants[side] = unit
	}
	session.Field0 = 1
	session.Field8, session.Field12 = participants[0], participants[1]
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= 0xffffffff {
		t.Fatalf("session did not exercise native pointer width: %p", session)
	}
	return world, session, participants
}

func addTradeOfferFixture(t *testing.T, world *server.Server, session *server.TradeSession, party *server.Object, code uint32) *server.TradeItem {
	t.Helper()
	item := world.NewObjectByTypeID("NativeTradeOfferItem")
	if item == nil {
		t.Fatal("allocate offer item")
	}
	item.NetCode = code
	item.Worth = 10 + code
	if !world.AddP2PTradeOfferNative50F820(session, party, item, server.TradeP2PRuntime50F3A0{}) {
		t.Fatal("create offer node")
	}
	node := session.Field36
	if party == session.Field8 {
		node = session.Field32
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(node)) <= 0xffffffff || uintptr(unsafe.Pointer(item)) <= 0xffffffff) {
		t.Fatalf("offer did not exercise native pointer width: %p -> %p", node, item)
	}
	return node
}

func tradeOfferStockCode(t *testing.T, packet []byte, related *server.Object, remove, sequence int) uint16 {
	t.Helper()
	if len(packet) != 18 || packet[0] != 0xc9 || packet[1] != 0x08 || related != nil || remove != 1 || sequence != 1 {
		t.Fatalf("shop stock packet = %x, related %p, remove %d, sequence %d", packet, related, remove, sequence)
	}
	return binary.LittleEndian.Uint16(packet[4:6])
}

func TestTradeOfferTransferNativeCABI50F790(t *testing.T) {
	type transferCase struct {
		Input struct {
			Name        string
			Count       int
			Player      bool
			FirstPlayer bool `json:"first_player"`
			Insert      []bool
		}
		Observed struct {
			Events []struct {
				Service   string
				Recipient string
				Item      int
				Node      string
			}
		}
	}
	rows := tradeOfferOriginalRows[transferCase](t, tradeOfferTransferOriginal,
		"fc32bddba4b7e3f6d3b9c5c89bbe943a145cd3c511376f739dc66c4fc20c8ad5", 48)
	for route := range 2 {
		for _, row := range rows {
			in := row.Input
			t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
				world, session, parties := newTradeOfferFixture(t, [2]bool{in.FirstPlayer, !in.FirstPlayer})
				recipient := parties[0]
				if in.Player != in.FirstPlayer {
					recipient = parties[1]
				}
				nodes := make([]*server.TradeItem, in.Count)
				before := make([]server.TradeItem, in.Count)
				for i := in.Count - 1; i >= 0; i-- {
					nodes[i] = addTradeOfferFixture(t, world, session, parties[0], uint32(i+1))
					// Represent a denied child insertion with an empty item.
					if !in.Player && !in.Insert[i] {
						nodes[i].Item0 = nil
					}
				}
				for i, node := range nodes {
					before[i] = *node
				}
				session.Field16 = 1
				beforeSession := *session
				var events, want []string
				game4ConsumerServices(t, game4TradeServices{world: world, p2p: func() server.TradeP2PRuntime50F3A0 {
					return server.TradeP2PRuntime50F3A0{PutInventory: func(owner, item *server.Object) {
						if owner != recipient {
							t.Fatalf("inventory recipient = %p, want %p", owner, recipient)
						}
						events = append(events, fmt.Sprintf("put:%d", item.NetCode-1))
					}}
				}})
				world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
					code := tradeOfferStockCode(t, packet, related, remove, sequence)
					events = append(events, fmt.Sprintf("stock:%d:new%d", index, code-1))
					return 1
				}
				for _, event := range row.Observed.Events {
					switch event.Service {
					case "put_inventory":
						want = append(want, fmt.Sprintf("put:%d", event.Item))
					case "send_stock":
						side := slices.Index([]string{"first", "second"}, event.Recipient)
						want = append(want, fmt.Sprintf("stock:%d:%s", side+1, event.Node))
					}
				}
				callTradeOfferTransfer50F790(route, session, recipient, session.Field32)
				beforeSession.Field20 = session.Field20
				if !slices.Equal(events, want) || *session != beforeSession {
					t.Fatalf("transfer events = %v, want %v, session changed = %t", events, want, *session != beforeSession)
				}
				stock := session.Field20
				for i, node := range nodes {
					if *node != before[i] {
						t.Fatalf("offer node %d changed", i)
					}
					if !in.Player && in.Insert[i] {
						if stock == nil || stock.Item0 != node.Item0 || stock.Cost4 != node.Item0.Worth {
							t.Fatalf("stock node for offer %d = %p", i, stock)
						}
						stock = stock.Field8
					}
				}
				if stock != nil {
					t.Fatal("unexpected stock item")
				}
			})
		}
	}
}

type tradeOfferAcceptCase struct {
	Input struct {
		Name    string
		Players [2]bool
		Actor   string
		Before  [2]uint32
		Mode    uint32
		Heads   [2]*string
		Gold    [2]uint32
		Totals  [2]uint32
	}
	Observed struct {
		Flags  [2]uint32
		Heads  [2]*string
		Gold   [2]uint32
		Totals [2]uint32
		Events []struct {
			Service   string
			Recipient string
			Head      *string
			Flags     [2]uint32
			Amount    uint32
		}
	}
}

func TestTradeOfferAcceptNativeCABI50F5A0(t *testing.T) {
	rows := tradeOfferOriginalRows[tradeOfferAcceptCase](t, tradeOfferAcceptOriginal,
		"1ae3103948b523f808bfec949fd05387c2e7c3ec76e0b9e563585b4130257269", 120)
	for route := range 2 {
		for _, row := range rows {
			// The game starts a trade for two players and a shop for one.
			in := row.Input
			if (in.Mode == 0 && in.Players == [2]bool{true, true}) ||
				(in.Mode == 1 && in.Players[0] != in.Players[1]) {
				t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
					checkTradeOfferAcceptCase(t, route, row)
				})
			}
		}
	}
}

func checkTradeOfferAcceptCase(t *testing.T, route int, row tradeOfferAcceptCase) {
	t.Helper()
	in, expected := row.Input, row.Observed
	world, session, parties := newTradeOfferFixture(t, in.Players)
	var heads [2]*server.TradeItem
	var gold [2]*server.GoldInitData
	for side := range parties {
		if in.Heads[side] != nil {
			heads[side] = addTradeOfferFixture(t, world, session, parties[side], uint32(side+1))
		}
		gold[side] = game4ConsumerValue(t, server.GoldInitData{Amount: in.Gold[side]})
	}
	session.Field48 = game4ConsumerValue(t, server.Object{InitData: unsafe.Pointer(gold[0])})
	session.Field52 = game4ConsumerValue(t, server.Object{InitData: unsafe.Pointer(gold[1])})
	session.Field16 = in.Mode
	session.Field24, session.Field28 = in.Before[0], in.Before[1]
	session.Field40, session.Field44 = in.Totals[0], in.Totals[1]
	side := func(name string) int { return slices.Index([]string{"first", "second"}, name) }
	actor := game4ConsumerValue(t, server.Object{})
	if index := side(in.Actor); index >= 0 {
		actor = parties[index]
	}
	var events, want []string
	assertHeads := func() {
		t.Helper()
		if session.Field32 != heads[0] || session.Field36 != heads[1] {
			t.Fatal("offer nodes released during the exchange")
		}
	}
	assertSettled := func() {
		t.Helper()
		if session.Field32 != nil || session.Field36 != nil || session.Field40 != 0 || session.Field44 != 0 ||
			gold[0].Amount != 0 || gold[1].Amount != 0 {
			t.Fatal("sent completion before clearing both offers and balances")
		}
	}
	send := func(player *server.Player, packet []byte) {
		index := player.Index()
		if packet[1] == 0x03 {
			assertHeads()
		} else {
			assertSettled()
		}
		events = append(events, fmt.Sprintf("send:%d:%x", index, packet))
	}
	game4ConsumerServices(t, game4TradeServices{world: world, p2p: func() server.TradeP2PRuntime50F3A0 {
		return server.TradeP2PRuntime50F3A0{
			Send: send, SendClose: send,
			PutInventory: func(owner, item *server.Object) {
				assertHeads()
				events = append(events, fmt.Sprintf("put:%d:%d", slices.Index(parties[:], owner)+1, item.NetCode))
			},
			AddGold: func(owner *server.Object, amount uint32) {
				assertHeads()
				events = append(events, fmt.Sprintf("gold:%d:%d", slices.Index(parties[:], owner)+1, amount))
			},
		}
	}})
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		assertHeads()
		code := tradeOfferStockCode(t, packet, related, remove, sequence)
		events = append(events, fmt.Sprintf("stock:%d:%d", index, code))
		return 1
	}
	for _, event := range expected.Events {
		index := side(event.Recipient)
		switch event.Service {
		case "status":
			var bits byte
			if event.Flags[index] != 0 {
				bits |= 1
			}
			if event.Flags[1-index] != 0 {
				bits |= 2
			}
			want = append(want, fmt.Sprintf("send:%d:c903%02x", index+1, bits))
		case "transfer":
			if event.Head != nil {
				if in.Players[index] {
					want = append(want, fmt.Sprintf("put:%d:%d", index+1, side(*event.Head)+1))
				} else {
					want = append(want, fmt.Sprintf("stock:%d:%d", 2-index, side(*event.Head)+1))
				}
			}
		case "transfer_gold":
			if in.Players[index] {
				want = append(want, fmt.Sprintf("gold:%d:%d", index+1, event.Amount))
			}
		case "refresh_shop":
			want = append(want, fmt.Sprintf("send:%d:c907", index+1))
		case "close_trade":
			want = append(want, "send:1:c901", "send:2:c901")
		}
	}
	callTradeOfferAccept50F5A0(route, session, actor)
	if !slices.Equal(events, want) {
		t.Fatalf("accept events = %v, want %v", events, want)
	}
	closed := in.Mode == 0 && expected.Flags == [2]uint32{1, 1}
	if world.IsTradeSessionNative(session) == closed {
		t.Fatalf("session ownership did not match close=%t", closed)
	}
	for index, player := range in.Players {
		if player && (parties[index].UpdateDataPlayer().Trade70 == session) == closed {
			t.Fatalf("participant %d link did not match close=%t", index, closed)
		}
		if gold[index].Amount != expected.Gold[index] {
			t.Fatalf("gold %d = %d, want %d", index, gold[index].Amount, expected.Gold[index])
		}
	}
	if !closed {
		gotHeads := [2]bool{session.Field32 != nil, session.Field36 != nil}
		wantHeads := [2]bool{expected.Heads[0] != nil, expected.Heads[1] != nil}
		if [2]uint32{session.Field24, session.Field28} != expected.Flags ||
			[2]uint32{session.Field40, session.Field44} != expected.Totals || gotHeads != wantHeads {
			t.Fatal("live session state differs from stock")
		}
	}
}

func TestTradeOfferAcceptNativeCallbackCanRelease50F5A0(t *testing.T) {
	world, session, parties := newTradeOfferFixture(t, [2]bool{true, true})
	addTradeOfferFixture(t, world, session, parties[0], 1)
	session.Field24 = 1
	sends := 0
	game4ConsumerServices(t, game4TradeServices{world: world, p2p: func() server.TradeP2PRuntime50F3A0 {
		return server.TradeP2PRuntime50F3A0{
			Send: func(*server.Player, []byte) {
				sends++
				if !world.ReleaseTradeSessionNative510000(session) {
					t.Fatal("release during status send")
				}
			},
			PutInventory: func(*server.Object, *server.Object) { t.Fatal("transferred from a released session") },
		}
	}})
	callTradeOfferAccept50F5A0(1, session, parties[1])
	if sends != 1 || world.IsTradeSessionNative(session) {
		t.Fatalf("status sends = %d, live session = %t", sends, world.IsTradeSessionNative(session))
	}
}
