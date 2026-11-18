package legacy

import (
	_ "embed"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_cancel_cache_original_v1.json.gz
var tradeCancelCacheOriginal []byte

type tradeCancelCacheCase struct {
	Input struct {
		Name      string
		Mode      string
		Shop      uint32
		Index     int
		Present   bool
		Placement int
	}
	Observed struct {
		Child string
		Fault *uint32
	}
}

type tradeCancelCacheServer struct {
	Server
	world     *server.Server
	cancelled *server.TradeSession
	cleared   []int
	packets   [][2]byte
}

func (s *tradeCancelCacheServer) ShopCancelSessionNative510DC0(session *server.TradeSession) bool {
	s.cancelled = session
	if !s.world.IsTradeSessionNative(session) {
		return false
	}
	if session.Field16 != 0 {
		return s.world.CloseShopNative50F4C0(session, nil)
	}
	return s.world.CancelP2PTradeNative50F3A0(session, server.TradeP2PRuntime50F3A0{
		SendClose: func(_ *server.Player, packet []byte) {
			s.packets = append(s.packets, [2]byte(packet))
		},
	})
}

func (s *tradeCancelCacheServer) ClearQuestShopSessionNative510E20(index int) bool {
	s.cleared = append(s.cleared, index)
	return s.world.ClearQuestShopSessionNative510E20(index)
}

func TestTradeCancelCacheNativeCABI510DC0(t *testing.T) {
	rows := tradeOfferOriginalRows[tradeCancelCacheCase](t, tradeCancelCacheOriginal,
		"522cbd928c39e19c5c570296061b35bd393548e17aadda05ea7d89a3ad72df2e", 141)
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	game4ConsumerServices(t, game4TradeServices{world: world})
	host := &tradeCancelCacheServer{world: world}
	world.NetSendPacketXxx = func(_ int, packet []byte, _ *server.Object, _, _ int) int {
		host.packets = append(host.packets, [2]byte(packet))
		return 1
	}
	getServer, flags := GetServer, noxflags.GetGame()
	GetServer = func() Server { return host }
	noxflags.UnsetGame(noxflags.GameModeQuest)
	t.Cleanup(func() {
		GetServer = getServer
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				host.cancelled, host.cleared, host.packets = nil, nil, nil
				switch row.Input.Mode {
				case "cancel":
					checkTradeCancelNative(t, route, row, host)
				case "clear":
					checkQuestCacheClearNative(t, route, row, host)
				default:
					t.Fatal("unknown original entry", row.Input.Mode)
				}
			})
		}
		host.cleared = nil
		callQuestCacheClear510E20(route, -1)
		callQuestCacheClear510E20(route, 32)
		if len(host.cleared) != 0 {
			t.Fatal("invalid Quest cache indices were forwarded")
		}
	}
}

func checkTradeCancelNative(t *testing.T, route int, row tradeCancelCacheCase, host *tradeCancelCacheServer) {
	t.Helper()
	if row.Observed.Fault != nil {
		// The native entry keeps the established nil guard.
		callTradeCancel510DC0(route, nil)
		if *row.Observed.Fault != 16 || host.cancelled != nil || len(host.packets) != 0 {
			t.Fatal("nil cancellation reached a child")
		}
		return
	}
	world := host.world
	first, _ := game4ConsumerPlayer(t, 101)
	session := world.NewTradeSessionNative50E870()
	defer world.ReleaseTradeSessionNative510000(session)
	session.Field0, session.Field8, session.Field16 = 1, first, row.Input.Shop
	first.UpdateDataPlayer().Trade70 = session
	second, _ := game4ConsumerPlayer(t, 179)
	if row.Observed.Child == "p2p" {
		session.Field12 = second
		second.UpdateDataPlayer().Trade70 = session
	} else {
		session.Field12 = game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
	}
	if row.Input.Placement != 0 {
		session.Field8, session.Field12 = session.Field12, session.Field8
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= 0xffffffff {
		t.Fatal("test did not allocate a high session")
	}
	callTradeCancel510DC0(route, session)
	var want [][2]byte
	if row.Observed.Child == "shop" {
		want = [][2]byte{{0xc9, 0x02}}
	} else {
		want = [][2]byte{{0xc9, 0x01}, {0xc9, 0x01}}
	}
	if host.cancelled != session || !reflect.DeepEqual(host.packets, want) || world.IsTradeSessionNative(session) ||
		first.UpdateDataPlayer().Trade70 != nil || second.UpdateDataPlayer().Trade70 != nil {
		t.Fatalf("cancellation callback or released session state differs: packets %v, want %v", host.packets, want)
	}
}

func checkQuestCacheClearNative(t *testing.T, route int, row tradeCancelCacheCase, host *tradeCancelCacheServer) {
	t.Helper()
	world := host.world
	index := row.Input.Index
	otherIndex := (index + 1) % 32
	other := world.NewTradeSessionNative50E870()
	defer world.ReleaseTradeSessionNative510000(other)
	if !world.CacheQuestShopSessionNative50F4C0(otherIndex, other) {
		t.Fatal("could not cache the adjacent fixture session")
	}
	var selected *server.TradeSession
	legacyOnly := cgoABIPointerSize == 4 && row.Input.Placement != 0
	if row.Input.Present {
		selected = world.NewTradeSessionNative50E870()
		defer world.ReleaseTradeSessionNative510000(selected)
		if !legacyOnly && !world.CacheQuestShopSessionNative50F4C0(index, selected) {
			t.Fatal("could not cache the selected fixture session")
		}
	}
	if cgoABIPointerSize == 4 {
		slot, otherSlot := questShopSessionSlotPE32(index), questShopSessionSlotPE32(otherIndex)
		before, otherBefore := *slot, *otherSlot
		defer func() { *slot, *otherSlot = before, otherBefore }()
		*slot, *otherSlot = unsafe.Pointer(selected), unsafe.Pointer(other)
	}
	callQuestCacheClear510E20(route, index)
	if !reflect.DeepEqual(host.cleared, []int{index}) || world.IsTradeSessionNative(selected) ||
		world.ClearQuestShopSessionNative510E20(index) || !world.IsTradeSessionNative(other) {
		t.Fatalf("cache clear changed the wrong owner or left the selected slot live: %v", host.cleared)
	}
	if row.Input.Present != (row.Observed.Child == "release") || len(host.packets) != 0 {
		t.Fatal("fixture child or unexpected packet")
	}
	if cgoABIPointerSize == 4 &&
		(*questShopSessionSlotPE32(index) != nil || *questShopSessionSlotPE32(otherIndex) != unsafe.Pointer(other)) {
		t.Fatal("PE32 cache clear changed the wrong slot")
	}
	if !world.ClearQuestShopSessionNative510E20(otherIndex) {
		t.Fatal("cache clear displaced the adjacent session")
	}
}
