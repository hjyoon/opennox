package server

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func addTestNativeTradeItem(t *testing.T, s *Server, session *TradeSession, item *Object, cost uint32) *TradeItem {
	t.Helper()
	state := s.tradeNative.sessions[session]
	if state == nil {
		t.Fatal("test session is not native")
	}
	node, freeNode := alloc.New(TradeItem{})
	node.Item0 = item
	node.Cost4 = cost
	insertSimpleShopItem50EE00(&session.Field20, node)
	state.items[node] = nativeTradeItemAllocation{
		freeNode:   freeNode,
		freeObject: func() {},
	}
	return node
}

func TestShopkeeperInitDataLayout50E970(t *testing.T) {
	if got := unsafe.Sizeof(ShopkeeperItemDefinition{}); got != 28 {
		t.Fatalf("shop item definition size = %d, want 28", got)
	}
	if got := unsafe.Offsetof(ShopkeeperItemDefinition{}.Count); got != 4 {
		t.Fatalf("shop item count offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(ShopkeeperItemDefinition{}.Param); got != 8 {
		t.Fatalf("shop item param offset = %d, want 8", got)
	}
	if got := unsafe.Offsetof(ShopkeeperItemDefinition{}.ModifierSlots); got != 12 {
		t.Fatalf("shop item modifier slots offset = %d, want 12", got)
	}
	if got := unsafe.Sizeof(ShopkeeperInitData{}); got != 1724 {
		t.Fatalf("shopkeeper init data size = %d, want 1724", got)
	}
	if got := unsafe.Offsetof(ShopkeeperInitData{}.Items); got != 4 {
		t.Fatalf("shopkeeper items offset = %d, want 4", got)
	}
	if got := unsafe.Offsetof(ShopkeeperInitData{}.ShopText); got != 1684 {
		t.Fatalf("shopkeeper text offset = %d, want 1684", got)
	}
	if got := unsafe.Offsetof(ShopkeeperInitData{}.BuyMultiplier); got != 1716 {
		t.Fatalf("shopkeeper buy multiplier offset = %d, want 1716", got)
	}
	if got := unsafe.Offsetof(ShopkeeperInitData{}.SellMultiplier); got != 1720 {
		t.Fatalf("shopkeeper sell multiplier offset = %d, want 1720", got)
	}
}

func TestShopObjectNameKey4E39F0(t *testing.T) {
	tests := []struct {
		objectID string
		typeID   string
		want     string
	}{
		{objectID: "MapGroup:Inn_Keeper", typeID: "Shopkeeper", want: "NPC:InnKeeper"},
		{objectID: "Inn_Keeper", typeID: "Shopkeeper", want: "NPC:InnKeeper"},
		{typeID: "Shop_keeper", want: "NPC:Shopkeeper"},
	}
	for _, tc := range tests {
		if got := shopObjectNameKey4E39F0(tc.objectID, tc.typeID); got != tc.want {
			t.Errorf("shopObjectNameKey4E39F0(%q, %q) = %q, want %q", tc.objectID, tc.typeID, got, tc.want)
		}
	}
}

func TestShopObjectNameKey4E39F0KeepsNativePointer(t *testing.T) {
	id, freeID := alloc.CString("MapGroup:Inn_Keeper")
	defer freeID()
	obj, freeObj := alloc.New(Object{})
	defer freeObj()
	obj.IDPtr = unsafe.Pointer(id)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= uintptr(^uint32(0)) {
		t.Fatalf("test object address = %#x, want native high address", uintptr(unsafe.Pointer(obj)))
	}
	if got := shopObjectNameKey4E39F0(obj.ID(), "ignored"); got != "NPC:InnKeeper" {
		t.Fatalf("native object name key = %q, want NPC:InnKeeper", got)
	}
}

func TestNativeShopSessionAllocation50E8F0(t *testing.T) {
	player, freePlayer := alloc.New(Object{})
	defer freePlayer()
	merchant, freeMerchant := alloc.New(Object{})
	defer freeMerchant()
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(player, merchant)
	if session.Field8 != player || session.Field12 != merchant || session.Field16 != 1 {
		t.Fatalf("session = %+v, want native player/merchant shop session", session)
	}
	if !s.IsTradeSessionNative(session) {
		t.Fatal("native session was not tracked")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= uintptr(^uint32(0)) {
		t.Fatalf("session address %#x did not exercise the high native half", uintptr(unsafe.Pointer(session)))
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
	if s.IsTradeSessionNative(session) || s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("released session remained owned or was released twice")
	}
	if s.ReleaseTradeSessionNative510000(&TradeSession{}) {
		t.Fatal("legacy/untracked session was released by native allocator")
	}
}

func TestInsertSimpleShopItem50EE00OriginalOrder(t *testing.T) {
	items := []*TradeItem{
		{Cost4: 20},
		{Cost4: 10},
		{Cost4: 20},
		{Cost4: 15},
	}
	var head *TradeItem
	for _, item := range items {
		insertSimpleShopItem50EE00(&head, item)
	}
	want := []*TradeItem{items[1], items[3], items[2], items[0]}
	var prev *TradeItem
	for i, item := range want {
		if head != item {
			t.Fatalf("item[%d] = %p cost=%d, want %p cost=%d", i, head, head.Cost4, item, item.Cost4)
		}
		if head.Field12 != prev {
			t.Fatalf("item[%d] previous = %p, want %p", i, head.Field12, prev)
		}
		prev = head
		head = head.Field8
	}
	if head != nil {
		t.Fatalf("list has unexpected tail %p", head)
	}
}

func TestSimpleShopItemCost50E3D0(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.BuyMultiplier = 1.5
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	player := &Object{ObjClass: object.ClassPlayer}
	session := &TradeSession{Field8: player, Field12: merchant, Field16: 1}
	health := &HealthData{Cur: 1, Max: 2}
	item := &Object{Worth: 101, HealthData: health}
	if got, ok := simpleShopItemCost50E3D0(session, item); !ok || got != 76 {
		t.Fatalf("simple health-adjusted cost = %d, %t, want 76, true", got, ok)
	}
	health.Cur = health.Max
	if got, ok := simpleShopItemCost50E3D0(session, item); !ok || got != 152 {
		t.Fatalf("round-to-even full cost = %d, %t, want 152, true", got, ok)
	}
	idata.BuyMultiplier = 0
	if got, ok := simpleShopItemCost50E3D0(session, item); !ok || got != 1 {
		t.Fatalf("minimum cost = %d, %t, want 1, true", got, ok)
	}
	item.ObjClass = object.ClassWeapon
	if _, ok := simpleShopItemCost50E3D0(session, item); ok {
		t.Fatal("modifier-capable weapon entered simple cost subset")
	}
	item.ObjClass = object.ClassInfoBook
	if _, ok := simpleShopItemCost50E3D0(session, item); ok {
		t.Fatal("categorized info book entered simple cost subset")
	}
	item.ObjClass = object.ClassFood
	item.Worth = shopSimpleMaxCost50EEC0 + 1
	idata.BuyMultiplier = 1
	if _, ok := simpleShopItemCost50E3D0(session, item); ok {
		t.Fatal("cost outside the default-category sort-key range entered simple subset")
	}
	idata.BuyMultiplier = float32(math.NaN())
	if _, ok := simpleShopItemCost50E3D0(session, item); ok {
		t.Fatal("NaN cost entered simple subset")
	}
}

func TestUnsupportedShopDefinition50E970(t *testing.T) {
	if hasUnsupportedShopDefinition50E970(&ShopkeeperItemDefinition{}) {
		t.Fatal("plain definition was rejected")
	}
	if !hasUnsupportedShopDefinition50E970(&ShopkeeperItemDefinition{Param: 1}) {
		t.Fatal("reward parameter was accepted")
	}
	def := &ShopkeeperItemDefinition{}
	def.ModifierSlots[3] = 1
	if !hasUnsupportedShopDefinition50E970(def) {
		t.Fatal("ABI32 modifier slot was accepted")
	}
}

func TestBuildShopStartPacket50F0F0(t *testing.T) {
	packet := BuildShopStartPacket50F0F0(0x1234, "상점 Merchant", []byte("ShopGreeting\x00ignored"))
	if len(packet) != ShopStartPacketSize50F0F0 || packet[0] != byte(netmsg.MSG_TRADE) || packet[1] != 0x0d {
		t.Fatalf("header = % x", packet[:4])
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != 0x1234 {
		t.Fatalf("type = %#x, want 0x1234", got)
	}
	wantName := utf16.Encode([]rune("상점 Merchant"))
	for i, want := range wantName {
		if got := binary.LittleEndian.Uint16(packet[4+2*i : 6+2*i]); got != want {
			t.Fatalf("name[%d] = %#x, want %#x", i, got, want)
		}
	}
	if packet[52] != 0 || packet[53] != 0 {
		t.Fatalf("name terminator = % x", packet[52:54])
	}
	if got := string(packet[54 : 54+len("ShopGreeting")]); got != "ShopGreeting" {
		t.Fatalf("shop text = %q", got)
	}
	for i, ch := range packet[54+len("ShopGreeting"):] {
		if ch != 0 {
			t.Fatalf("shop text tail[%d] = %#x, want zero", i, ch)
		}
	}
}

func TestBuildShopStartPacketBounds50F0F0(t *testing.T) {
	packet := BuildShopStartPacket50F0F0(7, "123456789012345678901234EXTRA", []byte("1234567890123456789012345678901EXTRA"))
	name := make([]uint16, 24)
	for i := range name {
		name[i] = binary.LittleEndian.Uint16(packet[4+2*i : 6+2*i])
	}
	if got := string(utf16.Decode(name)); got != "123456789012345678901234" {
		t.Fatalf("bounded name = %q", got)
	}
	if got := string(packet[54:85]); got != "1234567890123456789012345678901" || packet[85] != 0 {
		t.Fatalf("bounded shop text = %q tail=%#x", got, packet[85])
	}
}

func TestBuildShopItemPacket50F2B0(t *testing.T) {
	health := &HealthData{Cur: 0x1122, Field2: 0x3344}
	item := &Object{TypeInd: 0x5566, NetCode: 0x778899aa, HealthData: health}
	packet := BuildShopItemPacket50F2B0(item, 0xaabbccdd)
	if len(packet) != ShopItemPacketSize50F2B0 || packet[0] != byte(netmsg.MSG_TRADE) || packet[1] != 0x08 {
		t.Fatalf("header = % x", packet[:2])
	}
	if got := binary.LittleEndian.Uint16(packet[2:4]); got != item.TypeInd {
		t.Fatalf("type = %#x, want %#x", got, item.TypeInd)
	}
	if got := binary.LittleEndian.Uint16(packet[4:6]); got != uint16(item.NetCode) {
		t.Fatalf("net code = %#x, want %#x", got, uint16(item.NetCode))
	}
	if got := binary.LittleEndian.Uint32(packet[6:10]); got != 0xaabbccdd {
		t.Fatalf("cost = %#x, want 0xaabbccdd", got)
	}
	if got := binary.LittleEndian.Uint32(packet[10:14]); got != 0x33441122 {
		t.Fatalf("health dword = %#x, want 0x33441122", got)
	}
	if got := packet[14:18]; got[0] != 0xff || got[1] != 0xff || got[2] != 0xff || got[3] != 0xff {
		t.Fatalf("plain modifiers = % x, want ff ff ff ff", got)
	}
}

func TestBuildShopItemPacketModifierIDs50F2B0(t *testing.T) {
	attrs, freeAttrs := alloc.New(ModifierInitData{})
	defer freeAttrs()
	mods := [4]ModifierEff{{ind4: 1}, {ind4: 17}, {ind4: 128}, {ind4: 255}}
	for i := range mods {
		attrs.Modifiers[i] = &mods[i]
	}
	item := &Object{
		ObjClass: object.ClassWeapon,
		InitData: unsafe.Pointer(attrs),
	}
	packet := BuildShopItemPacket50F2B0(item, 1)
	want := [4]byte{1, 17, 128, 255}
	if got := [4]byte(packet[14:18]); got != want {
		t.Fatalf("modifier IDs = %v, want %v", got, want)
	}
}

func TestBuyShopItemNative5100C0TransfersOwnershipAndGold(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.Count = 1
	idata.Items[0] = ShopkeeperItemDefinition{TypeInd: 7, Count: 2}
	idata.BuyMultiplier = 1
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	player := &Player{GoldVal: 100, ProtPlayerGold: 0xfedcba98}
	update := &PlayerUpdateData{Player: player}
	playerUnit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	update.Trade70 = session
	item := &Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40}
	other := &Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1235, Worth: 40}
	targetNode := addTestNativeTradeItem(t, s, session, item, 40)
	otherNode := addTestNativeTradeItem(t, s, session, other, 40)

	events := make([]string, 0, 5)
	got := s.BuyShopItemNative5100C0(playerUnit, session, 0x1234, ShopBuyRuntime5100C0{
		PutInventory: func(gotPlayer, gotItem *Object) {
			events = append(events, "put")
			if gotPlayer != playerUnit || gotItem != item {
				t.Fatalf("inventory = %p/%p, want %p/%p", gotPlayer, gotItem, playerUnit, item)
			}
			gotPlayer.InvFirstItem = gotItem
			gotItem.InvHolder = gotPlayer
		},
		CallPickup: func(*Object, *Object) {
			t.Fatal("food purchase called its Pickup callback")
		},
		PlayPickupSound: func(gotPlayer *Object) {
			events = append(events, "sound")
			if gotPlayer != playerUnit {
				t.Fatalf("sound player = %p, want %p", gotPlayer, playerUnit)
			}
		},
		SendItemRemoved: func(gotPlayer *Player, gotItem *Object) {
			events = append(events, "removed")
			if gotPlayer != player || gotItem != item || player.GoldVal != 100 || idata.Items[0].Count != 1 {
				t.Fatalf("remove state = player %p item %p gold %d definition count %d", gotPlayer, gotItem, player.GoldVal, idata.Items[0].Count)
			}
		},
		ProtectGold: func(token uint32, delta int32) {
			events = append(events, "protect")
			if token != 0xfedcba98 || delta != -40 || player.GoldVal != 60 {
				t.Fatalf("protection = token %#x delta %d gold %d", token, delta, player.GoldVal)
			}
		},
		ReportGold: func(gotPlayer *Player, gotUnit *Object) {
			events = append(events, "gold")
			if gotPlayer != player || gotUnit != playerUnit || player.GoldVal != 60 {
				t.Fatalf("gold report = %p/%p gold %d", gotPlayer, gotUnit, player.GoldVal)
			}
		},
	})
	if got != ShopBuyComplete5100C0 {
		t.Fatalf("result = %d, want complete", got)
	}
	if want := []string{"put", "sound", "removed", "protect", "gold"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if session.Field20 != otherNode || otherNode.Field12 != nil || playerUnit.InvFirstItem != item || item.InvHolder != playerUnit {
		t.Fatalf("post-purchase ownership/list = head %p prev %p inventory %p holder %p", session.Field20, otherNode.Field12, playerUnit.InvFirstItem, item.InvHolder)
	}
	if _, ok := s.tradeNative.sessions[session].items[targetNode]; ok {
		t.Fatal("purchased node remained owned by the shop session")
	}
	if _, ok := s.tradeNative.sessions[session].items[otherNode]; !ok {
		t.Fatal("remaining node lost shop-session ownership")
	}
	if idata.Count != 1 || idata.Items[0].Count != 1 || player.GoldVal != 60 {
		t.Fatalf("post-purchase definition/gold = %d/%d/%d, want 1/1/60", idata.Count, idata.Items[0].Count, player.GoldVal)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(session)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(targetNode)) <= uintptr(^uint32(0))) {
		t.Fatalf("native trade pointers did not exercise the high half")
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("remaining native session was not released")
	}
}

func TestBuyShopItemNative5100C0RejectsMissingGold(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.Count = 1
	idata.Items[0] = ShopkeeperItemDefinition{TypeInd: 7, Count: 1}
	idata.BuyMultiplier = 1
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	player := &Player{GoldVal: 3}
	update := &PlayerUpdateData{Player: player}
	playerUnit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	item := &Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40}
	node := addTestNativeTradeItem(t, s, session, item, 40)
	missingCalls := 0
	got := s.BuyShopItemNative5100C0(playerUnit, session, 0x1234, ShopBuyRuntime5100C0{
		ReportMissingGold: func(gotPlayer *Player, amount uint16) {
			missingCalls++
			if gotPlayer != player || amount != 37 {
				t.Fatalf("missing gold = %p/%d, want %p/37", gotPlayer, amount, player)
			}
		},
		PutInventory: func(*Object, *Object) { t.Fatal("rejected purchase moved inventory") },
	})
	if got != ShopBuyMissingGold5100C0 || missingCalls != 1 || player.GoldVal != 3 || session.Field20 != node || idata.Items[0].Count != 1 {
		t.Fatalf("rejected purchase = result %d calls %d gold %d head %p count %d", got, missingCalls, player.GoldVal, session.Field20, idata.Items[0].Count)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestBuyShopItemNative5100C0EnforcesFoodLimit(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.Count = 1
	idata.Items[0] = ShopkeeperItemDefinition{TypeInd: 7, Count: 1}
	idata.BuyMultiplier = 1
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	player := &Player{GoldVal: 100}
	update := &PlayerUpdateData{Player: player}
	playerUnit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	for i := 0; i < 3; i++ {
		playerUnit.InvFirstItem = &Object{TypeInd: 7, InvNextItem: playerUnit.InvFirstItem}
	}
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	item := &Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40}
	node := addTestNativeTradeItem(t, s, session, item, 40)
	maxCalls := 0
	got := s.BuyShopItemNative5100C0(playerUnit, session, 0x1234, ShopBuyRuntime5100C0{
		ReportMaxSameItem: func(gotPlayer *Object) {
			maxCalls++
			if gotPlayer != playerUnit {
				t.Fatalf("max-item player = %p, want %p", gotPlayer, playerUnit)
			}
		},
		PutInventory: func(*Object, *Object) { t.Fatal("limited purchase moved inventory") },
	})
	if got != ShopBuyMaxSameItem5100C0 || maxCalls != 1 || player.GoldVal != 100 || session.Field20 != node {
		t.Fatalf("limited purchase = result %d calls %d gold %d head %p", got, maxCalls, player.GoldVal, session.Field20)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestShopPurchasePacketsMatchOriginalBytes(t *testing.T) {
	item := &Object{NetCode: 0x12345678}
	if got, want := BuildShopItemRemovePacket50E820(item), [4]byte{0xc9, 0x09, 0x78, 0x56}; got != want {
		t.Fatalf("remove packet = % x, want % x", got, want)
	}
	if got, want := BuildShopMissingGoldPacket5104F0(0xabcd), [4]byte{0xc9, 0x1b, 0xcd, 0xab}; got != want {
		t.Fatalf("missing-gold packet = % x, want % x", got, want)
	}
	if got, want := BuildShopGoldReportPacket4D8870(0x12345678), [5]byte{0x4a, 0x78, 0x56, 0x34, 0x12}; got != want {
		t.Fatalf("gold packet = % x, want % x", got, want)
	}
}
