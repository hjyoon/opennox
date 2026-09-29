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

func nativeTradeTestValue[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	*ptr = value
	t.Cleanup(free)
	return ptr
}

func addTestNativeTradeItem(t *testing.T, s *Server, session *TradeSession, item *Object, cost uint32) *TradeItem {
	t.Helper()
	state := s.tradeNative.sessions[session]
	if state == nil {
		t.Fatal("test session is not native")
	}
	node, freeNode := alloc.New(TradeItem{})
	node.Item0 = item
	node.Cost4 = cost
	insertShopItem50EE00(&session.Field20, node)
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

func TestQuestShopSessionCacheLifecycle50E8F0(t *testing.T) {
	player1, freePlayer1 := alloc.New(Object{})
	defer freePlayer1()
	merchant1, freeMerchant1 := alloc.New(Object{})
	defer freeMerchant1()
	player2, freePlayer2 := alloc.New(Object{})
	defer freePlayer2()
	merchant2, freeMerchant2 := alloc.New(Object{})
	defer freeMerchant2()

	s := &Server{}
	session, reused := s.OpenShopSessionNative50E8F0(player1, merchant1, true, 7)
	if reused || session == nil {
		t.Fatalf("fresh Quest shop = %p, reused %t; want non-nil/false", session, reused)
	}
	item, freeItem := alloc.New(Object{})
	defer freeItem()
	node := addTestNativeTradeItem(t, s, session, item, 41)
	state := s.tradeNative.sessions[session]
	freeCalls := 0
	originalFree := state.freeSession
	state.freeSession = func() {
		freeCalls++
		originalFree()
	}

	if !s.CacheQuestShopSessionNative50F4C0(7, session) {
		t.Fatal("Quest shop session was not cached")
	}
	if !s.IsTradeSessionNative(session) || freeCalls != 0 {
		t.Fatalf("cached Quest shop ownership/free calls = %t/%d, want true/0", s.IsTradeSessionNative(session), freeCalls)
	}

	reopened, reused := s.OpenShopSessionNative50E8F0(player2, merchant2, true, 7)
	if !reused || reopened != session {
		t.Fatalf("reopened Quest shop = %p, reused %t; want %p/true", reopened, reused, session)
	}
	if reopened.Field8 != player2 || reopened.Field12 != merchant2 || reopened.Field16 != 1 {
		t.Fatalf("reopened participants = (%p,%p,%d), want (%p,%p,1)", reopened.Field8, reopened.Field12, reopened.Field16, player2, merchant2)
	}
	if reopened.Field20 != node || node.Item0 != item || node.Cost4 != 41 {
		t.Fatal("reopened Quest shop did not retain its generated inventory")
	}
	if cached := s.tradeNative.questShopSession[7]; cached != nil {
		t.Fatalf("taken Quest cache slot = %p, want nil", cached)
	}

	if !s.CacheQuestShopSessionNative50F4C0(7, reopened) {
		t.Fatal("reopened Quest shop session was not cached")
	}
	if !s.ClearQuestShopSessionNative510E20(7) {
		t.Fatal("cached Quest shop session was not cleared")
	}
	if freeCalls != 1 || s.IsTradeSessionNative(session) || s.tradeNative.questShopSession[7] != nil {
		t.Fatalf("cleared Quest shop = frees:%d owned:%t cached:%p, want 1/false/nil", freeCalls, s.IsTradeSessionNative(session), s.tradeNative.questShopSession[7])
	}
	if s.ClearQuestShopSessionNative510E20(7) {
		t.Fatal("empty Quest shop cache slot was cleared twice")
	}
}

func TestQuestShopSessionCacheRejectsInvalidAndClearsOnRelease50F4C0(t *testing.T) {
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(nil, nil)
	if s.CacheQuestShopSessionNative50F4C0(-1, session) ||
		s.CacheQuestShopSessionNative50F4C0(questShopSessionSlots50E8F0, session) ||
		s.CacheQuestShopSessionNative50F4C0(0, &TradeSession{}) {
		t.Fatal("Quest shop cache accepted an invalid index or unowned session")
	}
	if !s.CacheQuestShopSessionNative50F4C0(3, session) {
		t.Fatal("valid Quest shop session was not cached")
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("cached Quest shop session was not released")
	}
	if s.tradeNative.questShopSession[3] != nil || s.ClearQuestShopSessionNative510E20(3) {
		t.Fatal("release left a dangling Quest shop cache entry")
	}
}

func TestNativeTradeSessionAllocation50E870LinksGoldAndReleases(t *testing.T) {
	s := &Server{}
	var allocated []*Object
	var released []*Object
	newGold := func() (*Object, func()) {
		obj, free := alloc.New(Object{})
		allocated = append(allocated, obj)
		return obj, func() {
			released = append(released, obj)
			free()
		}
	}

	first := s.newTradeSessionNative50E870(newGold)
	second := s.newTradeSessionNative50E870(newGold)
	third := s.newTradeSessionNative50E870(newGold)
	if len(allocated) != 6 {
		t.Fatalf("Gold allocations = %d, want 6", len(allocated))
	}
	if first.Field48 != allocated[0] || first.Field52 != allocated[1] ||
		second.Field48 != allocated[2] || second.Field52 != allocated[3] ||
		third.Field48 != allocated[4] || third.Field52 != allocated[5] {
		t.Fatal("session Gold objects do not preserve the original allocation order")
	}
	if s.tradeNative.head != third || third.Field56 != second || third.Field60 != nil ||
		second.Field56 != first || second.Field60 != third ||
		first.Field56 != nil || first.Field60 != second {
		t.Fatalf("native session list = head %p; third (%p,%p), second (%p,%p), first (%p,%p)",
			s.tradeNative.head, third.Field56, third.Field60, second.Field56, second.Field60, first.Field56, first.Field60)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{
			unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(third),
			unsafe.Pointer(first.Field48), unsafe.Pointer(first.Field52),
		} {
			if uintptr(ptr) <= uintptr(^uint32(0)) {
				t.Fatalf("native allocation address %#x did not exercise the high half", uintptr(ptr))
			}
		}
	}

	if !s.ReleaseTradeSessionNative510000(second) {
		t.Fatal("middle session was not released")
	}
	if s.tradeNative.head != third || third.Field56 != first || first.Field60 != third {
		t.Fatalf("middle unlink left head/links = %p/(%p,%p)", s.tradeNative.head, third.Field56, first.Field60)
	}
	if !reflect.DeepEqual(released, allocated[2:4]) {
		t.Fatalf("middle release order = %p, want %p", released, allocated[2:4])
	}
	if !s.ReleaseTradeSessionNative510000(third) || s.tradeNative.head != first || first.Field60 != nil {
		t.Fatalf("head unlink left head/previous = %p/%p", s.tradeNative.head, first.Field60)
	}
	if !s.ReleaseTradeSessionNative510000(first) || s.tradeNative.head != nil || len(s.tradeNative.sessions) != 0 {
		t.Fatalf("final unlink left head/sessions = %p/%d", s.tradeNative.head, len(s.tradeNative.sessions))
	}
	if want := []*Object{allocated[2], allocated[3], allocated[4], allocated[5], allocated[0], allocated[1]}; !reflect.DeepEqual(released, want) {
		t.Fatalf("Gold release order = %p, want %p", released, want)
	}
}

func TestNativeShopLifecycle50E2A0ResetAndFree(t *testing.T) {
	player, freePlayer := alloc.New(Object{})
	defer freePlayer()
	merchant, freeMerchant := alloc.New(Object{})
	defer freeMerchant()

	s := &Server{}
	if !s.TradeInit50E2A0() || s.tradeNative.sessions == nil {
		t.Fatal("native trade registry was not initialized")
	}
	session := s.NewShopSessionNative50E8F0(player, merchant)
	state := s.tradeNative.sessions[session]
	if state == nil {
		t.Fatal("native trade session was not registered")
	}

	var events []string
	originalFreeSession := state.freeSession
	state.freeSession = func() {
		events = append(events, "session")
		originalFreeSession()
	}
	node, freeNode := alloc.New(TradeItem{})
	state.items[node] = nativeTradeItemAllocation{
		freeObject: func() { events = append(events, "object") },
		freeNode: func() {
			events = append(events, "node")
			freeNode()
		},
	}

	s.TradeReset50E360()
	if want := []string{"object", "node", "session"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("reset cleanup events = %v, want %v", events, want)
	}
	if s.tradeNative.sessions == nil || len(s.tradeNative.sessions) != 0 {
		t.Fatalf("reset registry = %#v, want initialized and empty", s.tradeNative.sessions)
	}
	if s.IsTradeSessionNative(session) {
		t.Fatal("reset session remained registered")
	}

	next := s.NewShopSessionNative50E8F0(player, merchant)
	if !s.IsTradeSessionNative(next) {
		t.Fatal("reset registry did not accept a new session")
	}
	s.TradeFree50E300()
	if s.tradeNative.sessions != nil || s.IsTradeSessionNative(next) {
		t.Fatalf("free registry = %#v, want nil", s.tradeNative.sessions)
	}
	// Shutdown and a later new-session initialization are both idempotent.
	s.TradeFree50E300()
	if !s.TradeInit50E2A0() || s.tradeNative.sessions == nil || len(s.tradeNative.sessions) != 0 {
		t.Fatal("trade registry did not reinitialize after shutdown")
	}
}

func TestInsertShopItem50EE00OriginalOrder(t *testing.T) {
	items := []*TradeItem{
		{Cost4: 20},
		{Cost4: 10},
		{Cost4: 20},
		{Cost4: 15},
	}
	var head *TradeItem
	for _, item := range items {
		insertShopItem50EE00(&head, item)
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

func TestShopItemSortKey50EEC0Categories(t *testing.T) {
	tests := []struct {
		name     string
		class    object.Class
		subclass object.SubClass
		category uint32
	}{
		{name: "weapon", class: object.ClassWeapon, category: 0xff},
		{name: "wand", class: object.ClassWand, category: 0xff},
		{name: "armor", class: object.ClassArmor, category: 0xfe},
		{name: "field guide", class: object.ClassInfoBook, subclass: object.SubClass(object.BookFieldGuide), category: 0xfd},
		{name: "spell book", class: object.ClassInfoBook, subclass: object.SubClass(object.BookSpell), category: 0xfc},
		{name: "ability book", class: object.ClassInfoBook, subclass: object.SubClass(object.BookAbility), category: 0xfb},
		{name: "potion", class: object.ClassFood, subclass: object.SubClass(object.FoodPotion), category: 0xfa},
		{name: "food", class: object.ClassFood, category: 0xf9},
		{name: "default", class: object.ClassMonster, category: 0xf8},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item := &Object{ObjClass: tc.class, ObjSubClass: tc.subclass}
			if got, want := shopItemSortKey50EEC0(item, 7), tc.category<<24|7; got != want {
				t.Fatalf("sort key = %#x, want %#x", got, want)
			}
		})
	}
	if got, want := shopItemSortKey50EEC0(&Object{ObjClass: object.ClassMonster}, 0x01000000), uint32(0xf9000000); got != want {
		t.Fatalf("wide-cost OR key = %#x, want %#x", got, want)
	}
}

func TestShopItemBuyCost50E3D0(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.BuyMultiplier = 1.5
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	player := &Object{ObjClass: object.ClassPlayer}
	session := &TradeSession{Field8: player, Field12: merchant, Field16: 1}
	health := &HealthData{Cur: 1, Max: 2}
	item := &Object{Worth: 101, HealthData: health}
	s := &Server{}
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 76 {
		t.Fatalf("simple health-adjusted cost = %d, %t, want 76, true", got, ok)
	}
	health.Cur = health.Max
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 152 {
		t.Fatalf("round-to-even full cost = %d, %t, want 152, true", got, ok)
	}
	idata.BuyMultiplier = 0
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 1 {
		t.Fatalf("minimum cost = %d, %t, want 1, true", got, ok)
	}
	modifier := &ModifierEff{ind4: 7, Price20: 20}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	item.ObjClass = object.ClassWeapon
	item.InitData = unsafe.Pointer(attrs)
	item.Worth = 101
	idata.BuyMultiplier = 1
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 121 {
		t.Fatalf("modified weapon cost = %d, %t, want 121, true", got, ok)
	}
	item.ObjClass = object.ClassFood
	item.InitData = nil
	item.Worth = 0x01000000
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 0x01000000 {
		t.Fatalf("wide sort-key cost = %#x, %t, want 0x01000000, true", got, ok)
	}
	idata.BuyMultiplier = float32(math.NaN())
	if got, ok := s.shopItemBuyCost50E3D0(session, item); !ok || got != 1 {
		t.Fatalf("NaN cost = %d, %t, want original minimum 1", got, ok)
	}
}

func TestShopkeeperModifierIDEncoding50E970(t *testing.T) {
	for _, id := range []int{0, 1, 127, 254} {
		slot, ok := EncodeShopkeeperModifierID(id)
		if !ok {
			t.Fatalf("modifier ID %d was rejected", id)
		}
		got, present, valid := DecodeShopkeeperModifierID(slot)
		if !valid || !present || got != id {
			t.Fatalf("modifier ID %d round trip = %d/%t/%t through %#x", id, got, present, valid, slot)
		}
	}
	if _, ok := EncodeShopkeeperModifierID(-1); ok {
		t.Fatal("negative modifier ID was accepted")
	}
	if _, ok := EncodeShopkeeperModifierID(255); ok {
		t.Fatal("no-modifier sentinel was accepted")
	}
	if id, present, valid := DecodeShopkeeperModifierID(0); id != 0 || present || !valid {
		t.Fatalf("empty modifier token = %d/%t/%t", id, present, valid)
	}
	if _, _, valid := DecodeShopkeeperModifierID(0x100); valid {
		t.Fatal("out-of-range modifier token was accepted")
	}
}

func TestShopDefinitionMatchesItem5103F0(t *testing.T) {
	s := &Server{}
	s.Types.byInd = make([]*ObjectType, 6)
	s.Types.byInd[5] = &ObjectType{ind: 5, id: "Wasp"}

	slot0, ok := EncodeShopkeeperModifierID(0)
	if !ok {
		t.Fatal("modifier ID zero was not encodable")
	}
	slot7, ok := EncodeShopkeeperModifierID(7)
	if !ok {
		t.Fatal("modifier ID seven was not encodable")
	}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{
		&ModifierEff{ind4: 0},
		nil,
		&ModifierEff{ind4: 7},
		nil,
	}}
	weapon := &Object{TypeInd: 11, ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(attrs)}
	weaponDef := ShopkeeperItemDefinition{TypeInd: 11, ModifierSlots: [4]uint32{slot0, 0, slot7, 0}}
	if !s.shopDefinitionMatchesItem5103F0(weapon, &weaponDef) {
		t.Fatal("matching modifier-bearing weapon definition was rejected")
	}
	weaponDef.ModifierSlots[2] = slot0
	if s.shopDefinitionMatchesItem5103F0(weapon, &weaponDef) {
		t.Fatal("mismatched modifier-bearing weapon definition was accepted")
	}

	spellData := &SpellRewardUseData{Spell: 9}
	spellBook := &Object{
		TypeInd:     12,
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookSpell),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(spellData)},
	}
	if !s.shopDefinitionMatchesItem5103F0(spellBook, &ShopkeeperItemDefinition{TypeInd: 12, Param: 9}) ||
		s.shopDefinitionMatchesItem5103F0(spellBook, &ShopkeeperItemDefinition{TypeInd: 12, Param: 8}) {
		t.Fatal("spell reward parameter matching is not exact")
	}

	abilityData := &AbilityRewardUseData{Ability: 4}
	abilityBook := &Object{
		TypeInd:     13,
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookAbility),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(abilityData)},
	}
	if !s.shopDefinitionMatchesItem5103F0(abilityBook, &ShopkeeperItemDefinition{TypeInd: 13, Param: 4}) ||
		s.shopDefinitionMatchesItem5103F0(abilityBook, &ShopkeeperItemDefinition{TypeInd: 13, Param: 3}) {
		t.Fatal("ability reward parameter matching is not exact")
	}

	guideData := &FieldGuideUseData{}
	guideData.SetCreature("Wasp")
	fieldGuide := &Object{
		TypeInd:     14,
		ObjClass:    object.ClassInfoBook,
		ObjSubClass: object.SubClass(object.BookFieldGuide),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(guideData)},
	}
	if !s.shopDefinitionMatchesItem5103F0(fieldGuide, &ShopkeeperItemDefinition{TypeInd: 14, Param: 5}) ||
		s.shopDefinitionMatchesItem5103F0(fieldGuide, &ShopkeeperItemDefinition{TypeInd: 14, Param: 4}) {
		t.Fatal("field guide creature matching is not exact")
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
	mods, freeMods := alloc.Make([]ModifierEff{}, 4)
	defer freeMods()
	for i, id := range [4]uint32{1, 17, 128, 255} {
		mods[i].ind4 = id
	}
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
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{GoldVal: 100, ProtPlayerGold: 0xfedcba98})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)})
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	update.Trade70 = session
	item := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40})
	other := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1235, Worth: 40})
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
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{GoldVal: 3})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)})
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	item := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40})
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
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{GoldVal: 100})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)})
	for i := 0; i < 3; i++ {
		playerUnit.InvFirstItem = nativeTradeTestValue(t, Object{TypeInd: 7, InvNextItem: playerUnit.InvFirstItem})
	}
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	item := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40})
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

func TestShopInventoryItemCost50E3D0SellAndRepair(t *testing.T) {
	s := &Server{}
	idata, freeShop := alloc.New(ShopkeeperInitData{})
	defer freeShop()
	idata.BuyMultiplier = 2
	idata.SellMultiplier = 0.5
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	session := &TradeSession{Field8: &Object{ObjClass: object.ClassPlayer}, Field12: merchant, Field16: 1}
	attrs, freeAttrs := alloc.New(ModifierInitData{})
	defer freeAttrs()
	modifier := nativeTradeTestValue(t, ModifierEff{Price20: 20})
	attrs.Modifiers[0] = modifier
	health := nativeTradeTestValue(t, HealthData{Cur: 25, Max: 100})
	item := &Object{
		ObjClass:   object.ClassWeapon,
		Worth:      100,
		HealthData: health,
		InitData:   unsafe.Pointer(attrs),
	}
	if got, ok := s.shopInventoryItemCost50E3D0(session, item, shopPriceSell50E3D0, 0); !ok || got != 15 {
		t.Fatalf("sell cost = %d, %t, want 15, true", got, ok)
	}
	if got, ok := s.shopInventoryItemCost50E3D0(session, item, shopPriceRepair50E3D0, 0.5); !ok || got != 90 {
		t.Fatalf("repair cost = %d, %t, want 90, true", got, ok)
	}
	item.ObjSubClass = 0x82
	if got, ok := s.shopInventoryItemCost50E3D0(session, item, shopPriceSell50E3D0, 0); !ok || got != 15 {
		t.Fatalf("ammo weapon without optional charge data = %d, %t, want 15, true", got, ok)
	}
}

func TestShopSellQuoteAndCompletion5109C0(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.SellMultiplier = 0.5
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{GoldVal: 60, ProtPlayerGold: 0xfedcba98})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	item := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 7, NetCode: 0x1234, Worth: 40})
	other := nativeTradeTestValue(t, Object{ObjClass: object.ClassFood, TypeInd: 8, NetCode: 0x1235, Worth: 10})
	playerUnit := nativeTradeTestValue(t, Object{
		ObjClass:     object.ClassPlayer,
		UpdateData:   unsafe.Pointer(update),
		InvFirstItem: item,
	})
	item.InvHolder = playerUnit
	item.InvNextItem = other
	other.InvHolder = playerUnit
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	update.Trade70 = session

	events := make([]string, 0, 8)
	runtime := ShopSellRuntime5109C0{
		ItemIsQuest: func(got *Object) bool {
			events = append(events, "quest")
			return false
		},
		ItemIsGlyph: func(got *Object) bool {
			events = append(events, "glyph")
			return false
		},
		SendQuote: func(gotPlayer *Player, gotItem *Object, cost uint32) {
			events = append(events, "quote")
			if gotPlayer != player || gotItem != item || cost != 20 {
				t.Fatalf("quote = %p/%p/%d, want %p/%p/20", gotPlayer, gotItem, cost, player, item)
			}
		},
		DetachInventory: func(gotPlayer, gotItem *Object) {
			events = append(events, "detach")
			if gotPlayer != playerUnit || gotItem != item || player.GoldVal != 60 {
				t.Fatalf("detach state = %p/%p gold %d", gotPlayer, gotItem, player.GoldVal)
			}
			gotPlayer.InvFirstItem = gotItem.InvNextItem
			gotItem.InvHolder = nil
			gotItem.InvNextItem = nil
		},
		DelayedDelete: func(gotItem *Object) {
			events = append(events, "delete")
			if gotItem != item || playerUnit.InvFirstItem != other || player.GoldVal != 60 {
				t.Fatalf("delete state = item %p head %p gold %d", gotItem, playerUnit.InvFirstItem, player.GoldVal)
			}
		},
		ProtectGold: func(token uint32, delta int32) {
			events = append(events, "protect")
			if token != 0xfedcba98 || delta != 20 || player.GoldVal != 80 {
				t.Fatalf("protection = %#x/%d gold %d", token, delta, player.GoldVal)
			}
		},
		ReportGold: func(gotPlayer *Player, gotUnit *Object) {
			events = append(events, "gold")
			if gotPlayer != player || gotUnit != playerUnit || player.GoldVal != 80 {
				t.Fatalf("gold report = %p/%p/%d", gotPlayer, gotUnit, player.GoldVal)
			}
		},
		PlaySellSound: func(gotPlayer *Object) {
			events = append(events, "sound")
			if gotPlayer != playerUnit {
				t.Fatalf("sound player = %p, want %p", gotPlayer, playerUnit)
			}
		},
	}
	if got := s.QuoteShopSellNative5109C0(playerUnit, session, 0x1234, runtime); got != ShopSellQuoted5109C0 {
		t.Fatalf("quote result = %d, want quoted", got)
	}
	if want := []string{"quest", "glyph", "quote"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("quote events = %v, want %v", events, want)
	}
	events = events[:0]
	if got := s.SellShopItemNative510BE0(playerUnit, session, 0x1234, runtime); got != ShopSellComplete5109C0 {
		t.Fatalf("sell result = %d, want complete", got)
	}
	if want := []string{"quest", "glyph", "detach", "delete", "protect", "gold", "sound"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("sell events = %v, want %v", events, want)
	}
	if playerUnit.InvFirstItem != other || player.GoldVal != 80 {
		t.Fatalf("sell result state = head %p gold %d, want %p/80", playerUnit.InvFirstItem, player.GoldVal, other)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestShopSellUsesFullInventoryNetCode5109C0(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.SellMultiplier = 1
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	lowHalfOnly := nativeTradeTestValue(t, Object{NetCode: 0x10001234, Worth: 99})
	exact := nativeTradeTestValue(t, Object{NetCode: 0x1234, Worth: 7})
	lowHalfOnly.InvNextItem = exact
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: lowHalfOnly})
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	quoted := (*Object)(nil)
	got := s.QuoteShopSellNative5109C0(playerUnit, session, 0x1234, ShopSellRuntime5109C0{
		SendQuote: func(_ *Player, item *Object, cost uint32) {
			quoted = item
			if cost != 7 {
				t.Fatalf("quote cost = %d, want 7", cost)
			}
		},
	})
	if got != ShopSellQuoted5109C0 || quoted != exact {
		t.Fatalf("full-code quote = result %d item %p, want quoted/%p", got, quoted, exact)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestSellShopItemsByTypeNative510D10UsesNativePointers(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.SellMultiplier = 1
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{GoldVal: 10, ProtPlayerGold: 0x89abcdef})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	first := nativeTradeTestValue(t, Object{TypeInd: 7, Worth: 20})
	other := nativeTradeTestValue(t, Object{TypeInd: 9, Worth: 100})
	second := nativeTradeTestValue(t, Object{TypeInd: 7, Worth: 30})
	first.InvNextItem = other
	other.InvNextItem = second
	playerUnit := nativeTradeTestValue(t, Object{
		ObjClass:     object.ClassPlayer,
		UpdateData:   unsafe.Pointer(update),
		InvFirstItem: first,
	})
	for _, item := range []*Object{first, other, second} {
		item.InvHolder = playerUnit
	}
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	update.Trade70 = session
	events := make([]string, 0, 9)
	runtime := ShopSellRuntime5109C0{
		ItemIsQuest: func(*Object) bool {
			t.Fatal("bulk sale called the single-item quest filter")
			return false
		},
		ItemIsGlyph: func(*Object) bool {
			t.Fatal("bulk sale called the single-item glyph filter")
			return false
		},
		DetachInventory: func(gotPlayer, gotItem *Object) {
			events = append(events, "detach")
			if gotPlayer != playerUnit || gotItem.TypeInd != 7 {
				t.Fatalf("detach = %p/%p", gotPlayer, gotItem)
			}
			var prev *Object
			for it := gotPlayer.InvFirstItem; it != nil; it = it.InvNextItem {
				if it == gotItem {
					if prev == nil {
						gotPlayer.InvFirstItem = it.InvNextItem
					} else {
						prev.InvNextItem = it.InvNextItem
					}
					it.InvHolder = nil
					it.InvNextItem = nil
					return
				}
				prev = it
			}
			t.Fatal("detached item was not in inventory")
		},
		DelayedDelete: func(item *Object) {
			events = append(events, "delete")
			if item.InvHolder != nil || item.InvNextItem != nil {
				t.Fatalf("deleted item remained linked: holder %p next %p", item.InvHolder, item.InvNextItem)
			}
		},
		ProtectGold: func(token uint32, delta int32) {
			events = append(events, "protect")
			if token != 0x89abcdef || (delta != 20 && delta != 30) {
				t.Fatalf("protection = %#x/%d", token, delta)
			}
		},
		ReportGold: func(gotPlayer *Player, gotUnit *Object) {
			events = append(events, "gold")
			if gotPlayer != player || gotUnit != playerUnit {
				t.Fatalf("gold report = %p/%p", gotPlayer, gotUnit)
			}
		},
		PlaySellSound: func(gotPlayer *Object) {
			events = append(events, "sound")
			if gotPlayer != playerUnit {
				t.Fatalf("sound player = %p, want %p", gotPlayer, playerUnit)
			}
		},
	}
	if got := s.SellShopItemsByTypeNative510D10(playerUnit, session, 7, 2, runtime); got != 2 {
		t.Fatalf("sold = %d, want 2", got)
	}
	if want := []string{"detach", "delete", "protect", "gold", "detach", "delete", "protect", "gold", "sound"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if player.GoldVal != 60 || playerUnit.InvFirstItem != other || other.InvNextItem != nil {
		t.Fatalf("state = gold %d head %p next %p, want 60/%p/nil", player.GoldVal, playerUnit.InvFirstItem, other.InvNextItem, other)
	}
	for name, ptr := range map[string]unsafe.Pointer{
		"player":  unsafe.Pointer(playerUnit),
		"session": unsafe.Pointer(session),
		"first":   unsafe.Pointer(first),
		"second":  unsafe.Pointer(second),
	} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= uintptr(^uint32(0)) {
			t.Fatalf("%s address %#x did not exercise the high native half", name, uintptr(ptr))
		}
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestSellShopItemsByTypeNative510D10StopsWithoutCompletionSound(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.SellMultiplier = 1
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	item := nativeTradeTestValue(t, Object{TypeInd: 7, Worth: 10})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: item})
	item.InvHolder = playerUnit
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	sounds := 0
	sold := s.SellShopItemsByTypeNative510D10(playerUnit, session, 7, 2, ShopSellRuntime5109C0{
		DetachInventory: func(player, got *Object) {
			player.InvFirstItem = nil
			got.InvHolder = nil
		},
		DelayedDelete: func(*Object) {},
		PlaySellSound: func(*Object) {
			sounds++
		},
	})
	if sold != 1 || sounds != 0 || player.GoldVal != 10 {
		t.Fatalf("partial sale = sold %d sounds %d gold %d, want 1/0/10", sold, sounds, player.GoldVal)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestShopRepairQuoteAndCompletion5108D0(t *testing.T) {
	idata, freeShop := alloc.New(ShopkeeperInitData{})
	defer freeShop()
	idata.BuyMultiplier = 2
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	attrs, freeAttrs := alloc.New(ModifierInitData{})
	defer freeAttrs()
	health := nativeTradeTestValue(t, HealthData{Cur: 25, Max: 100})
	item := nativeTradeTestValue(t, Object{
		ObjClass:   object.ClassWeapon,
		NetCode:    0x4321,
		Worth:      100,
		HealthData: health,
		InitData:   unsafe.Pointer(attrs),
	})
	player := nativeTradeTestValue(t, Player{GoldVal: 100, ProtPlayerGold: 0x89abcdef})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: item})
	item.InvHolder = playerUnit
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	update.Trade70 = session
	events := make([]string, 0, 7)
	runtime := ShopRepairRuntime5108D0{
		RepairCoefficient: 0.5,
		SendQuote: func(gotPlayer *Player, gotItem *Object, cost uint32) {
			events = append(events, "quote")
			if gotPlayer != player || gotItem != item || cost != 75 {
				t.Fatalf("repair quote = %p/%p/%d, want %p/%p/75", gotPlayer, gotItem, cost, player, item)
			}
		},
		ProtectGold: func(token uint32, delta int32) {
			events = append(events, "protect")
			if token != 0x89abcdef || delta != -75 || player.GoldVal != 25 {
				t.Fatalf("repair protection = %#x/%d gold %d", token, delta, player.GoldVal)
			}
		},
		SetHealth: func(gotItem *Object, amount uint16) {
			events = append(events, "health")
			if gotItem != item || amount != 100 || player.GoldVal != 25 {
				t.Fatalf("set health = %p/%d gold %d", gotItem, amount, player.GoldVal)
			}
			gotItem.HealthData.Cur = amount
		},
		ReportHealth: func(gotPlayer *Player, gotItem *Object) {
			events = append(events, "report-health")
			if gotPlayer != player || gotItem != item || health.Cur != health.Max {
				t.Fatalf("health report = %p/%p health %d/%d", gotPlayer, gotItem, health.Cur, health.Max)
			}
		},
		ReportGold: func(gotPlayer *Player, gotUnit *Object) {
			events = append(events, "gold")
			if gotPlayer != player || gotUnit != playerUnit || player.GoldVal != 25 {
				t.Fatalf("repair gold report = %p/%p/%d", gotPlayer, gotUnit, player.GoldVal)
			}
		},
		PlayRepairSound: func(gotPlayer *Object) {
			events = append(events, "sound")
			if gotPlayer != playerUnit {
				t.Fatalf("repair sound player = %p, want %p", gotPlayer, playerUnit)
			}
		},
	}
	if got := s.QuoteShopRepairNative5108D0(playerUnit, session, 0x4321, runtime); got != ShopRepairQuoted5108D0 {
		t.Fatalf("repair quote result = %d, want quoted", got)
	}
	if want := []string{"quote"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("repair quote events = %v, want %v", events, want)
	}
	events = events[:0]
	if got := s.RepairShopItemNative510AE0(playerUnit, session, 0x4321, runtime); got != ShopRepairComplete5108D0 {
		t.Fatalf("repair result = %d, want complete", got)
	}
	if want := []string{"protect", "health", "report-health", "gold", "sound"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("repair events = %v, want %v", events, want)
	}
	if player.GoldVal != 25 || health.Cur != health.Max {
		t.Fatalf("repair state = gold %d health %d/%d, want 25/100/100", player.GoldVal, health.Cur, health.Max)
	}
	if !s.ReleaseTradeSessionNative510000(session) {
		t.Fatal("native session was not released")
	}
}

func TestShopRepairQuoteRejectsPristineItem5108D0(t *testing.T) {
	idata, freeInit := alloc.New(ShopkeeperInitData{})
	defer freeInit()
	idata.BuyMultiplier = 1
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)})
	player := nativeTradeTestValue(t, Player{})
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	health := nativeTradeTestValue(t, HealthData{Cur: 10, Max: 10})
	item := nativeTradeTestValue(t, Object{NetCode: 9, HealthData: health})
	playerUnit := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: item})
	s := &Server{}
	session := s.NewShopSessionNative50E8F0(playerUnit, merchant)
	rejects := 0
	got := s.QuoteShopRepairNative5108D0(playerUnit, session, 9, ShopRepairRuntime5108D0{
		PlayRejectSound: func(gotPlayer *Object) {
			rejects++
			if gotPlayer != playerUnit {
				t.Fatalf("reject player = %p, want %p", gotPlayer, playerUnit)
			}
		},
		SendQuote: func(*Player, *Object, uint32) { t.Fatal("pristine item was quoted") },
	})
	if got != ShopRepairNotDamaged5108D0 || rejects != 1 {
		t.Fatalf("pristine quote = result %d rejects %d", got, rejects)
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
	if got, want := BuildShopSellQuotePacket5109C0(item, 0xaabbccdd), [8]byte{0xc9, 0x1d, 0x78, 0x56, 0xdd, 0xcc, 0xbb, 0xaa}; got != want {
		t.Fatalf("sell quote packet = % x, want % x", got, want)
	}
	if got, want := BuildShopRepairQuotePacket5108D0(item, 0x01020304), [8]byte{0xc9, 0x1f, 0x78, 0x56, 0x04, 0x03, 0x02, 0x01}; got != want {
		t.Fatalf("repair quote packet = % x, want % x", got, want)
	}
	healthItem := &Object{NetCode: 0x12345678, HealthData: &HealthData{Cur: 0x1122, Max: 0x3344}}
	if got, want := BuildShopItemHealthPacket4D87A0(healthItem), [7]byte{0x44, 0x78, 0x56, 0x22, 0x11, 0x44, 0x33}; got != want {
		t.Fatalf("item health packet = % x, want % x", got, want)
	}
}
