package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_start_dispatch_original_v2.json.gz
var tradeStartDispatchStock50EF10 []byte

type originalTradeStartEvent50EF10 struct {
	Kind    string
	Party   int
	Parties [2]int
	Key     string
}

type originalTradeStartCase50EF10 struct {
	Input struct {
		Name       string
		Players    [2]bool
		OldLinks   [2]*string `json:"old_links"`
		Flags      uint32
		Frame      uint32
		Same       bool `json:"same_partner"`
		Swap       bool `json:"swap_storage"`
		Allocate   bool
		TargetName string `json:"other_name"`
	}
	Observed struct {
		Events   []originalTradeStartEvent50EF10
		Links    [2]*string
		Returned bool
		Frame    uint32
		Active   uint32
		Shop     bool
		Parties  [2]int
	}
}

func TestTradeStartDispatchNativeCABI50EF10(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(tradeStartDispatchStock50EF10))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "431fa7730160f4512f7fe7b9ebed527d290e87e464aa0e3c201b2061e489fd17" {
		t.Fatalf("stock dispatch digest %s", got)
	}
	var stock struct {
		Rows []originalTradeStartCase50EF10
	}
	if err := json.Unmarshal(raw, &stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 54 {
		t.Fatalf("stock dispatch cases = %d", len(stock.Rows))
	}
	for route := range 2 {
		for _, row := range stock.Rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				checkTradeStartDispatch50EF10(t, route, row)
			})
		}
	}
}

func checkTradeStartDispatch50EF10(t *testing.T, route int, row originalTradeStartCase50EF10) {
	in, want := row.Input, row.Observed
	world := server.NewHigh(nil, nil, strman.New())
	defer world.Close()
	world.SetFrame(in.Frame)
	var units [2]*server.Object
	var updates [2]*server.PlayerUpdateData
	titles := [2]string{"Conjurer", in.TargetName}
	indices := [2]uint8{3, 29}
	for i := range units {
		unit, player := game4ConsumerPlayer(t, 0)
		player.PlayerInd = indices[i]
		player.SetName(titles[i])
		updates[i] = unit.UpdateDataPlayer()
		if !in.Players[i] {
			unit.ObjClass = object.ClassMonster
			merchant := game4ConsumerValue(t, server.ShopkeeperInitData{})
			copy(merchant.ShopText[:], "shop dialogue")
			unit.InitData = unsafe.Pointer(merchant)
			unit.TypeInd = uint16(401 + i)
		}
		units[i] = unit
	}
	oldStarter := world.NewTradeSessionNative50E870()
	oldTarget := world.NewTradeSessionNative50E870()
	oldStarter.Field4, oldTarget.Field4 = 0x9b5290c1, 0x8c4489d4
	if in.Same {
		oldStarter.Field12 = units[1]
	}
	previous := map[string]*server.TradeSession{"starter": oldStarter, "target": oldTarget}
	findLink := func(name *string) *server.TradeSession {
		if name == nil {
			return nil
		}
		return previous[*name]
	}
	for i, link := range in.OldLinks {
		updates[i].Trade70 = findLink(link)
	}
	unitIndex := func(unit *server.Object) int {
		pos := slices.Index(units[:], unit)
		if pos < 0 {
			t.Fatalf("unexpected participant %p", unit)
		}
		return pos
	}

	var events, expected []originalTradeStartEvent50EF10
	reason := server.TradeP2PStartInvalid50EF10
	for _, event := range want.Events {
		switch event.Kind {
		case "allocate_shop", "trade_packet", "shop_packet", "stock_packet", "freeze":
			expected = append(expected, event)
		case "load":
			switch event.Key {
			case "StarterAlreadyTrading":
				reason = server.TradeP2PStartStarterBusy50EF10
			case "OtherAlreadyTrading":
				reason = server.TradeP2PStartOtherBusy50EF10
			}
		}
	}
	switch {
	case want.Returned:
		reason = server.TradeP2PStartComplete50EF10
	case in.Same:
		reason = server.TradeP2PStartSamePartner50EF10
	case !in.Players[0] && in.OldLinks[1] != nil:
		reason = server.TradeP2PStartOtherBusy50EF10
	}

	item := game4ConsumerValue(t, server.Object{TypeInd: 71, NetCode: 755})
	openShop := func(player, merchant *server.Object) *server.TradeSession {
		events = append(events, originalTradeStartEvent50EF10{Kind: "allocate_shop", Parties: [2]int{unitIndex(player), unitIndex(merchant)}})
		if !in.Allocate {
			return nil
		}
		session := world.NewShopSessionNative50E8F0(player, merchant)
		if in.Swap {
			session.Field8, session.Field12 = session.Field12, session.Field8
		}
		session.InsertShopItem50EE00(game4ConsumerValue(t, server.TradeItem{Item0: item, Cost4: 194}))
		return session
	}
	freeze := func(unit *server.Object) {
		if in.Flags&0x800 != 0 {
			events = append(events, originalTradeStartEvent50EF10{Kind: "freeze", Party: unitIndex(unit)})
		}
	}
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		party := 0
		if index == int(indices[1]) {
			party = 1
		}
		session := updates[party].Trade70
		if index != int(indices[party]) || len(packet) < 2 || packet[0] != 0xc9 || related != nil || remove != 1 || sequence != 1 ||
			session == nil || !world.IsTradeSessionNative(session) || session.Field0 != 1 || session.Field4 != in.Frame {
			t.Fatalf("invalid queue call for participant %d", party)
		}
		for i, isPlayer := range in.Players {
			if isPlayer && updates[i].Trade70 != session {
				t.Fatalf("participant %d was not linked before the first packet", i)
			}
		}
		event := originalTradeStartEvent50EF10{Party: party}
		switch packet[1] {
		case 0x0c:
			event.Kind = "trade_packet"
			p := server.BuildTradeP2PStartPacket50F1A0(titles[1-party])
			if !bytes.Equal(packet, p[:]) {
				t.Fatal("trade packet did not use the peer name")
			}
		case 0x0d:
			event.Kind = "shop_packet"
			p := server.BuildShopStartPacket50F0F0(units[1-party].TypeInd, "MISSING:'NPC:'", []byte("shop dialogue"))
			if !bytes.Equal(packet, p[:]) {
				t.Fatal("shop packet did not use merchant data")
			}
		case 0x08:
			event.Kind = "stock_packet"
			p := server.BuildShopItemPacket50F2B0(item, 194)
			if !bytes.Equal(packet, p[:]) {
				t.Fatal("stock packet did not use the session item")
			}
		default:
			t.Fatalf("unexpected trade opcode %#x", packet[1])
		}
		events = append(events, event)
		return 1
	}
	game4ConsumerServices(t, game4TradeServices{world: world, start: func(starter, target *server.Object) *server.TradeSession {
		session, result := world.StartTradeNative50EF10(starter, target, world.Frame(), openShop, freeze)
		if result != reason {
			t.Fatalf("dispatch result = %d, want %d", result, reason)
		}
		return session
	}})

	session := tradeStartFixtureCall(route, units[0], units[1])
	if !slices.Equal(events, expected) || (session != nil) != want.Returned {
		t.Fatalf("session %p and events %+v, want returned=%t and events %+v", session, events, want.Returned, expected)
	}
	if session != nil {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= 0xffffffff {
			t.Fatal("returned session must exercise high pointer bits")
		}
		if !world.IsTradeSessionNative(session) || session.Field0 != want.Active || session.Field4 != want.Frame ||
			session.Field8 != units[want.Parties[0]] || session.Field12 != units[want.Parties[1]] ||
			(session.Field16 != 0) != want.Shop {
			t.Fatal("incorrect session owner, participants or start state")
		}
		previous["session"] = session
	}
	for i, link := range want.Links {
		if updates[i].Trade70 != findLink(link) {
			t.Fatalf("participant %d link %p, want %p", i, updates[i].Trade70, findLink(link))
		}
	}
	if oldStarter.Field4 != 0x9b5290c1 || oldTarget.Field4 != 0x8c4489d4 {
		t.Fatal("pre-existing trade state changed")
	}
}

func TestTradeStartDispatchNativeSendCanClose50EF10(t *testing.T) {
	world := server.NewHigh(nil, nil, strman.New())
	defer world.Close()
	first, _ := game4ConsumerPlayer(t, 0)
	second, _ := game4ConsumerPlayer(t, 0)
	game4ConsumerServices(t, game4TradeServices{world: world, start: func(starter, target *server.Object) *server.TradeSession {
		session, _ := world.StartTradeNative50EF10(starter, target, 9, nil, nil)
		return session
	}})
	calls := 0
	world.NetSendPacketXxx = func(_ int, packet []byte, _ *server.Object, _, _ int) int {
		calls++
		if calls != 1 || len(packet) != server.TradeP2PStartPacketSize50F1A0 {
			t.Fatalf("send after session release: %x", packet)
		}
		firstUpdate, secondUpdate := first.UpdateDataPlayer(), second.UpdateDataPlayer()
		session := firstUpdate.Trade70
		if session == nil || secondUpdate.Trade70 != session {
			t.Fatal("session was not bound to both participants")
		}
		firstUpdate.Trade70, secondUpdate.Trade70 = nil, nil
		if !world.ReleaseTradeSessionNative510000(session) {
			t.Fatal("session release failed")
		}
		return 0
	}
	if session := tradeStartFixtureCall(1, first, second); session != nil || calls != 1 {
		t.Fatalf("start returned %p and %d sends after release", session, calls)
	}
}
