package legacy

import (
	"encoding/binary"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func game4ConsumerValue[T comparable](t *testing.T, value T) *T {
	t.Helper()
	p, free := alloc.New(value)
	*p = value
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatalf("fixture address %p does not exercise native pointer width", p)
	}
	return p
}

func game4ConsumerPlayer(t *testing.T, gold uint32) (*server.Object, *server.Player) {
	t.Helper()
	p := game4ConsumerValue(t, server.Player{GoldVal: gold})
	u := game4ConsumerValue(t, server.PlayerUpdateData{Player: p})
	return game4ConsumerValue(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(u)}), p
}

func game4ConsumerServices(t *testing.T, rt game4TradeServices) {
	t.Helper()
	old := game4TradeServicesForTest
	game4TradeServicesForTest = &rt
	t.Cleanup(func() { game4TradeServicesForTest = old })
}

func TestGame4GoldConsumersActualCABIAndNativeGold(t *testing.T) {
	if mask := game4ConsumerFixtureABI(); mask != 255 {
		t.Fatalf("native C signature mask %#x, want255", mask)
	}
	owner, player := game4ConsumerPlayer(t, 0xfffffffc)
	data := game4ConsumerValue(t, server.GoldInitData{Amount: 7})
	gold := game4ConsumerValue(t, server.Object{InitData: unsafe.Pointer(data)})
	game4ConsumerFixtureReturn(owner, gold)
	if player.GoldVal != 3 {
		t.Fatalf("DWORD wrapped gold=%d, want3", player.GoldVal)
	}
	other := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
	if got := game4ConsumerFixtureReturn(other, gold); got != uintptr(unsafe.Pointer(other)) {
		t.Fatalf("nonplayer native return %#x, want %p", got, other)
	}
}

func TestGame4GoldConsumersActualCRepriceNativeFieldsAndCanaries(t *testing.T) {
	first, player := game4ConsumerPlayer(t, 100)
	second := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
	a := game4ConsumerValue(t, server.GoldInitData{Amount: 7})
	b := game4ConsumerValue(t, server.GoldInitData{Amount: 11})
	goldA := game4ConsumerValue(t, server.Object{InitData: unsafe.Pointer(a)})
	goldB := game4ConsumerValue(t, server.Object{InitData: unsafe.Pointer(b)})
	nodeA := game4ConsumerValue(t, server.TradeItem{Cost4: 30})
	nodeB := game4ConsumerValue(t, server.TradeItem{Cost4: 10})
	wrapped := game4ConsumerValue(t, struct {
		Before  [32]byte
		Session server.TradeSession
		After   [32]byte
	}{})
	for i := range wrapped.Before {
		wrapped.Before[i] = 0xa5
		wrapped.After[i] = 0x5a
	}
	wrapped.Session = server.TradeSession{Field8: first, Field12: second, Field48: goldA, Field52: goldB, Field32: nodeA, Field36: nodeB}
	if got := game4ConsumerFixtureReprice(&wrapped.Session); got != uintptr(unsafe.Pointer(&b.Amount)) {
		t.Fatalf("native amount return %#x, want %p", got, &b.Amount)
	}
	if player.GoldVal != 107 || a.Amount != 0 || b.Amount != 20 || wrapped.Session.Field40 != 30 || wrapped.Session.Field44 != 30 {
		t.Fatalf("reprice state: player=%d offers=%d/%d totals=%d/%d", player.GoldVal, a.Amount, b.Amount, wrapped.Session.Field40, wrapped.Session.Field44)
	}
	for i := range wrapped.Before {
		if wrapped.Before[i] != 0xa5 || wrapped.After[i] != 0x5a {
			t.Fatal("native session canary changed")
		}
	}
	player.GoldVal = 5
	a.Amount = 0
	b.Amount = 0
	nodeA.Cost4 = 10
	nodeB.Cost4 = 100
	game4ConsumerFixtureReprice(&wrapped.Session)
	if player.GoldVal != 0 || a.Amount != 5 || wrapped.Session.Field40 != 15 || wrapped.Session.Field44 != 100 {
		t.Fatalf("limited balance: player=%d offer=%d totals=%d/%d", player.GoldVal, a.Amount, wrapped.Session.Field40, wrapped.Session.Field44)
	}
}

func TestGame4GoldConsumersPinnedSessionAcrossGC(t *testing.T) {
	a, b := &server.Object{ObjClass: object.ClassPlayer}, &server.Object{ObjClass: object.ClassPlayer}
	session := &server.TradeSession{Field8: a, Field12: b, Field40: 0xfedcba98, Field44: 0x87654321}
	var pins runtime.Pinner
	pins.Pin(a)
	pins.Pin(b)
	pins.Pin(session)
	defer pins.Unpin()
	runtime.GC()
	if got := game4ConsumerFixtureReprice(session); got != 4 || session.Field40 != 0xfedcba98 || session.Field44 != 0x87654321 {
		t.Fatalf("player/player early return=%#x totals=%#x/%#x", got, session.Field40, session.Field44)
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	runtime.KeepAlive(session)
}

func TestGame4GoldConsumersActualCShopSaleRepairAndBulk(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	world := &server.Server{}
	owner, player := game4ConsumerPlayer(t, 100)
	idata := game4ConsumerValue(t, server.ShopkeeperInitData{BuyMultiplier: 1, SellMultiplier: 1})
	merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	session := world.NewShopSessionNative50E8F0(owner, merchant)
	t.Cleanup(func() { world.ReleaseTradeSessionNative510000(session) })
	owner.UpdateDataPlayer().Trade70 = session
	var events []string
	report := func(p *server.Player, u *server.Object) {
		if p != player || u != owner {
			t.Fatal("native report owner changed")
		}
		events = append(events, "gold")
	}
	sell := server.ShopSellRuntime5109C0{
		DetachInventory: func(u, it *server.Object) {
			if u != owner {
				t.Fatal("native detach owner")
			}
			u.InvFirstItem = it.InvNextItem
			events = append(events, "detach")
		},
		DelayedDelete: func(*server.Object) { events = append(events, "delete") },
		ProtectGold:   func(uint32, int32) { events = append(events, "protect") }, ReportGold: report,
		PlaySellSound: func(*server.Object) { events = append(events, "sound") },
	}
	repair := server.ShopRepairRuntime5108D0{RepairCoefficient: 1, SetHealth: func(o *server.Object, h uint16) { o.HealthData.Cur = h; events = append(events, "health") }, ProtectGold: func(uint32, int32) { events = append(events, "protect") }, ReportHealth: func(*server.Player, *server.Object) { events = append(events, "report-health") }, ReportGold: report, PlayRepairSound: func(*server.Object) { events = append(events, "repair-sound") }}
	buys := 0
	game4ConsumerServices(t, game4TradeServices{world: world, sell: func() server.ShopSellRuntime5109C0 { return sell }, repair: func() server.ShopRepairRuntime5108D0 { return repair }, buy: func() server.ShopBuyRuntime5100C0 { buys++; return server.ShopBuyRuntime5100C0{} }})
	item := game4ConsumerValue(t, server.Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40})
	owner.InvFirstItem = item
	if got := game4ConsumerFixtureSell(owner, session, 0x1234); got != 0x1234 {
		t.Fatalf("wire residue=%#x", got)
	}
	if player.GoldVal != 140 || owner.InvFirstItem != nil || !reflect.DeepEqual(events, []string{"detach", "delete", "protect", "gold", "sound"}) {
		t.Fatalf("sale gold=%d events=%v", player.GoldVal, events)
	}
	events = nil
	health := game4ConsumerValue(t, server.HealthData{Cur: 50, Max: 100})
	item.HealthData = health
	item.Worth = 100
	owner.InvFirstItem = item
	game4ConsumerFixtureRepair(owner, session, 0x1234)
	if player.GoldVal != 90 || health.Cur != 100 || !reflect.DeepEqual(events, []string{"protect", "health", "report-health", "gold", "repair-sound"}) {
		t.Fatalf("repair gold=%d health=%d events=%v", player.GoldVal, health.Cur, events)
	}
	events = nil
	item.HealthData = nil
	item.Worth = 10
	other := game4ConsumerValue(t, server.Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1235, Worth: 10})
	item.InvNextItem = other
	game4ConsumerFixtureSellMany(owner, session, 7, 2)
	if player.GoldVal != 110 || owner.InvFirstItem != nil || !reflect.DeepEqual(events, []string{"detach", "delete", "protect", "gold", "detach", "delete", "protect", "gold", "sound"}) {
		t.Fatalf("bulk sale gold=%d events=%v", player.GoldVal, events)
	}
	// Missing inventory is the existing native failure path, not a fake
	// purchase. It also verifies both public buy ABIs dispatch their runtimes.
	game4ConsumerFixtureBuy(owner, session, 0x9999)
	if got := game4ConsumerFixtureBuyMany(owner, session, 99, 1); got != 0 {
		t.Fatalf("missing buy count=%d", got)
	}
	if buys != 2 || player.GoldVal != 110 {
		t.Fatalf("buy dispatch=%d gold=%d", buys, player.GoldVal)
	}
}

func TestGame4GoldConsumersActualCCancelAndPacketHandoff(t *testing.T) {
	world := &server.Server{}
	a, pa := game4ConsumerPlayer(t, 20)
	b, pb := game4ConsumerPlayer(t, 30)
	session := world.NewTradeSessionNative50E870()
	session.Field8 = a
	session.Field12 = b
	a.UpdateDataPlayer().Trade70 = session
	b.UpdateDataPlayer().Trade70 = session
	var packets [][]byte
	game4ConsumerServices(t, game4TradeServices{world: world, p2p: func() server.TradeP2PRuntime50F3A0 {
		return server.TradeP2PRuntime50F3A0{SendClose: func(p *server.Player, d []byte) {
			if p != pa && p != pb {
				t.Fatal("native peer changed")
			}
			packets = append(packets, append([]byte(nil), d...))
		}}
	}, packet: func(got *server.Object, data []byte) int {
		if got != a || !reflect.DeepEqual(data, []byte{0xc9, 0x19, 0x34, 0x12, 2, 0xa5}) {
			t.Fatalf("native packet=%p/%x", got, data)
		}
		return 5
	}})
	game4ConsumerFixtureCancel(session)
	if world.IsTradeSessionNative(session) || a.UpdateDataPlayer().Trade70 != nil || b.UpdateDataPlayer().Trade70 != nil || !reflect.DeepEqual(packets, [][]byte{{0xc9, 1}, {0xc9, 1}}) {
		t.Fatalf("cancel packets=%x", packets)
	}
	packet := []byte{0xc9, 0x19, 0x34, 0x12, 2, 0xa5}
	if got := game4ConsumerFixturePacket(a, packet); got != 5 || packet[5] != 0xa5 {
		t.Fatalf("native handoff consumed=%d packet=%x", got, packet)
	}
	if got := game4ConsumerFixturePacket(a, []byte{0xc9}); got != -1 {
		t.Fatalf("short handoff=%d", got)
	}
	// Packet layout checks are fixed-width scalars, independent of host pointers.
	if binary.LittleEndian.Uint16(packet[2:4]) != 0x1234 || sound.SoundTreasurePickup != 307 {
		t.Fatal("wire scalar contract changed")
	}
}
