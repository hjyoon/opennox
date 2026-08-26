package server

import (
	"encoding/binary"
	"math"
	"strings"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	ShopStartPacketSize50F0F0       = 86
	ShopItemPacketSize50F2B0        = 18
	ShopItemRemovePacketSize50E820  = 4
	ShopMissingGoldPacketSize5104F0 = 4
	ShopGoldReportPacketSize4D8870  = 5
	ShopSellQuotePacketSize5109C0   = 8
	ShopRepairQuotePacketSize5108D0 = 8
	ShopItemHealthPacketSize4D87A0  = 7

	shopModifierClassMask50E3D0 = object.ClassWand | object.ClassWeapon | object.ClassArmor | object.ClassFlag
	shopSpecialClassMask50E3D0  = shopModifierClassMask50E3D0 | object.ClassInfoBook
	shopSimpleMaxCost50EEC0     = uint32(0x00ffffff)
)

type shopPriceMode50E3D0 uint8

const (
	shopPriceSell50E3D0 shopPriceMode50E3D0 = iota
	shopPriceBuy50E3D0
	shopPriceRepair50E3D0
)

type ShopBuyResult5100C0 uint8

const (
	ShopBuyNoItem5100C0 ShopBuyResult5100C0 = iota
	ShopBuyUnsupported5100C0
	ShopBuyMissingGold5100C0
	ShopBuyMaxSameItem5100C0
	ShopBuyComplete5100C0
)

// ShopBuyRuntime5100C0 binds services that still cross the legacy runtime.
// The native transaction owns all Object, Player, TradeSession, and TradeItem
// pointers; callbacks receive native pointers without an IA-32 integer cast.
type ShopBuyRuntime5100C0 struct {
	ExpandedFoodLimit bool
	QuestPersistent   func(*Object) bool
	PutInventory      func(player, item *Object)
	CallPickup        func(player, item *Object)
	PlayPickupSound   func(*Object)
	ProtectGold       func(token uint32, delta int32)
	SendItemRemoved   func(*Player, *Object)
	ReportGold        func(*Player, *Object)
	ReportMissingGold func(*Player, uint16)
	ReportMaxSameItem func(*Object)
}

type ShopSellResult5109C0 uint8

const (
	ShopSellNoItem5109C0 ShopSellResult5109C0 = iota
	ShopSellUnsupported5109C0
	ShopSellQuestItem5109C0
	ShopSellGlyph5109C0
	ShopSellQuoted5109C0
	ShopSellComplete5109C0
)

// ShopSellRuntime5109C0 binds the native single-item quote and completion
// paths restored from 005109C0 and 00510BE0. All item and player arguments
// remain native pointers; only the protocol net code is intentionally narrow.
type ShopSellRuntime5109C0 struct {
	ItemIsQuest           func(*Object) bool
	ItemIsGlyph           func(*Object) bool
	DetachInventory       func(player, item *Object)
	DelayedDelete         func(*Object)
	ProtectGold           func(token uint32, delta int32)
	SendQuote             func(*Player, *Object, uint32)
	ReportGold            func(*Player, *Object)
	ReportCannotSellQuest func(*Object)
	ReportCannotSellItem  func(*Object)
	PlayRejectSound       func(*Object)
	PlaySellSound         func(*Object)
}

type ShopRepairResult5108D0 uint8

const (
	ShopRepairNoItem5108D0 ShopRepairResult5108D0 = iota
	ShopRepairUnsupported5108D0
	ShopRepairNotDamaged5108D0
	ShopRepairQuoted5108D0
	ShopRepairComplete5108D0
)

// ShopRepairRuntime5108D0 binds the regular/Coop health-durability subset of
// the quote and completion paths at 005108D0 and 00510AE0. Charge-bearing
// wands and ammo weapons remain explicitly outside this first native subset.
type ShopRepairRuntime5108D0 struct {
	RepairCoefficient float32
	SetHealth         func(*Object, uint16)
	ProtectGold       func(token uint32, delta int32)
	SendQuote         func(*Player, *Object, uint32)
	ReportHealth      func(*Player, *Object)
	ReportGold        func(*Player, *Object)
	PlayRejectSound   func(*Object)
	PlayRepairSound   func(*Object)
}

type TradeSession struct {
	Field0  uint32        // 0, 0
	Field4  uint32        // 1, 4
	Field8  *Object       // 2, 8
	Field12 *Object       // 3, 12
	Field16 uint32        // 4, 16
	Field20 *TradeItem    // 5, 20
	Field24 uint32        // 6, 24
	Field28 uint32        // 7, 28
	Field32 *TradeItem    // 8, 32
	Field36 *TradeItem    // 9, 36
	Field40 uint32        // 10, 40
	Field44 uint32        // 11, 44
	Field48 *Object       // 12, 48
	Field52 *Object       // 13, 52
	Field56 *TradeSession // 14, 56
	Field60 *TradeSession // 15, 60
}

type TradeItem struct {
	Item0   *Object    // 0, 0
	Cost4   uint32     // 1, 4
	Field8  *TradeItem // 2, 8
	Field12 *TradeItem // 3, 12
}

type nativeTradeItemAllocation struct {
	freeNode   func()
	freeObject func()
}

type nativeTradeSessionAllocation struct {
	freeSession func()
	items       map[*TradeItem]nativeTradeItemAllocation
}

type serverTradeNativeState struct {
	sessions map[*TradeSession]*nativeTradeSessionAllocation
}

func (t *serverTradeNativeState) init() {
	if t.sessions == nil {
		t.sessions = make(map[*TradeSession]*nativeTradeSessionAllocation)
	}
}

func (t *serverTradeNativeState) close() {
	for session, state := range t.sessions {
		delete(t.sessions, session)
		for item, allocation := range state.items {
			delete(state.items, item)
			allocation.freeObject()
			allocation.freeNode()
		}
		state.freeSession()
	}
}

// NewShopSessionNative50E8F0 allocates a pointer-width-safe shop session on
// the C heap. The original PE32 pool is fixed at 64 bytes, while the same
// typed structure grows with native pointers on 64-bit targets.
func (s *Server) NewShopSessionNative50E8F0(player, merchant *Object) *TradeSession {
	s.tradeNative.init()
	session, free := alloc.New(TradeSession{})
	session.Field8 = player
	session.Field12 = merchant
	session.Field16 = 1
	s.tradeNative.sessions[session] = &nativeTradeSessionAllocation{
		freeSession: free,
		items:       make(map[*TradeItem]nativeTradeItemAllocation),
	}
	return session
}

// IsTradeSessionNative reports whether session is owned by the native-width
// allocator rather than the original PE32 pool.
func (s *Server) IsTradeSessionNative(session *TradeSession) bool {
	if session == nil || s.tradeNative.sessions == nil {
		return false
	}
	_, ok := s.tradeNative.sessions[session]
	return ok
}

// ReleaseTradeSessionNative510000 releases only sessions allocated by the
// native-width trade subsystem. It is deliberately safe for legacy sessions.
func (s *Server) ReleaseTradeSessionNative510000(session *TradeSession) bool {
	if session == nil || s.tradeNative.sessions == nil {
		return false
	}
	state, ok := s.tradeNative.sessions[session]
	if !ok {
		return false
	}
	delete(s.tradeNative.sessions, session)
	for item, allocation := range state.items {
		delete(state.items, item)
		allocation.freeObject()
		allocation.freeNode()
	}
	state.freeSession()
	return true
}

func insertSimpleShopItem50EE00(head **TradeItem, item *TradeItem) {
	if *head == nil || item.Cost4 <= (*head).Cost4 {
		item.Field8 = *head
		if item.Field8 != nil {
			item.Field8.Field12 = item
		}
		*head = item
		return
	}
	prev := *head
	for prev.Field8 != nil && item.Cost4 > prev.Field8.Cost4 {
		prev = prev.Field8
	}
	item.Field8 = prev.Field8
	if item.Field8 != nil {
		item.Field8.Field12 = item
	}
	prev.Field8 = item
	item.Field12 = prev
}

// addSimpleShopItemNative50EE00 owns item until it is detached by a completed
// purchase or the session is released. This is the exact equal-cost insertion
// order from 0050EE00 for the currently restored simple-item category.
func (s *Server) addSimpleShopItemNative50EE00(session *TradeSession, item *Object, cost uint32) *TradeItem {
	state := s.tradeNative.sessions[session]
	if state == nil || item == nil {
		return nil
	}
	node, freeNode := alloc.New(TradeItem{})
	node.Item0 = item
	node.Cost4 = cost
	insertSimpleShopItem50EE00(&session.Field20, node)
	state.items[node] = nativeTradeItemAllocation{
		freeNode: freeNode,
		freeObject: func() {
			s.Objs.FreeObject(item)
		},
	}
	return node
}

func (s *Server) findNativeTradeItem5100C0(session *TradeSession, netCode uint16) *TradeItem {
	if session == nil || !s.IsTradeSessionNative(session) {
		return nil
	}
	for node := session.Field20; node != nil; node = node.Field8 {
		if node.Item0 != nil && node.Item0.NetCode == uint32(netCode) {
			return node
		}
	}
	return nil
}

func decrementSimpleShopDefinition510320(session *TradeSession, item *Object) {
	if session == nil || item == nil {
		return
	}
	merchant := session.Field8
	if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
		merchant = session.Field12
	}
	if merchant == nil || merchant.InitData == nil {
		return
	}
	idata := merchant.InitDataShopkeeper()
	count := int(idata.Count)
	if count > len(idata.Items) {
		count = len(idata.Items)
	}
	for i := 0; i < count; i++ {
		def := &idata.Items[i]
		if def.TypeInd != uint32(item.TypeInd) || hasUnsupportedShopDefinition50E970(def) {
			continue
		}
		def.Count--
		if def.Count != 0 {
			return
		}
		copy(idata.Items[i:count-1], idata.Items[i+1:count])
		idata.Items[count-1] = ShopkeeperItemDefinition{}
		idata.Count--
		return
	}
}

// detachNativeTradeItem50E7A0 unlinks and releases a native TradeItem node
// after its Object has moved into the buyer's inventory. The Object itself is
// deliberately removed from session ownership and remains alive.
func (s *Server) detachNativeTradeItem50E7A0(session *TradeSession, node *TradeItem, notify func(*Object)) bool {
	if session == nil || node == nil || s.tradeNative.sessions == nil {
		return false
	}
	state := s.tradeNative.sessions[session]
	if state == nil {
		return false
	}
	allocation, ok := state.items[node]
	if !ok {
		return false
	}
	if next := node.Field8; next != nil {
		next.Field12 = node.Field12
	}
	if prev := node.Field12; prev != nil {
		prev.Field8 = node.Field8
	} else if session.Field20 == node {
		session.Field20 = node.Field8
	} else {
		return false
	}
	item := node.Item0
	if notify != nil {
		notify(item)
	}
	delete(state.items, node)
	allocation.freeNode()
	return true
}

// BuyShopItemNative5100C0 restores the regular/Coop, non-modified single-item
// purchase path from GAME.EXE 005100C0. Quest-persistent gems and AnkhTradable
// are rejected until their clone, life-limit, and replenishment rules are
// restored. Every accepted Object and list pointer remains native-width.
func (s *Server) BuyShopItemNative5100C0(
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopBuyRuntime5100C0,
) ShopBuyResult5100C0 {
	node := s.findNativeTradeItem5100C0(session, netCode)
	if node == nil || playerUnit == nil || !playerUnit.Class().Has(object.ClassPlayer) {
		return ShopBuyNoItem5100C0
	}
	item := node.Item0
	if runtime.QuestPersistent != nil && runtime.QuestPersistent(item) {
		return ShopBuyUnsupported5100C0
	}
	cost, ok := simpleShopItemCost50E3D0(session, item)
	if !ok {
		return ShopBuyUnsupported5100C0
	}
	update := (*PlayerUpdateData)(playerUnit.UpdateData)
	if update == nil || update.Player == nil {
		return ShopBuyUnsupported5100C0
	}
	player := update.Player
	if cost > player.GoldVal {
		if runtime.ReportMissingGold != nil {
			runtime.ReportMissingGold(player, uint16(cost-player.GoldVal))
		}
		return ShopBuyMissingGold5100C0
	}
	if item.Class().Has(object.ClassFood) {
		limit := int32(3)
		if runtime.ExpandedFoodLimit {
			limit = 9
		}
		if playerUnit.CountInventoryWithType(int32(item.TypeInd)) >= limit {
			if runtime.ReportMaxSameItem != nil {
				runtime.ReportMaxSameItem(playerUnit)
			}
			return ShopBuyMaxSameItem5100C0
		}
	}

	if item.Class().HasAny(object.ClassFood|object.ClassInfoBook) || item.Pickup.Ptr == nil {
		runtime.PutInventory(playerUnit, item)
		if runtime.PlayPickupSound != nil {
			runtime.PlayPickupSound(playerUnit)
		}
	} else {
		runtime.CallPickup(playerUnit, item)
	}
	decrementSimpleShopDefinition510320(session, item)
	if !s.detachNativeTradeItem50E7A0(session, node, func(item *Object) {
		if runtime.SendItemRemoved != nil {
			runtime.SendItemRemoved(player, item)
		}
	}) {
		return ShopBuyUnsupported5100C0
	}
	gold := player.GoldVal
	if gold >= cost {
		player.GoldVal = gold - cost
	} else {
		player.GoldVal = 0
	}
	if runtime.ProtectGold != nil {
		runtime.ProtectGold(player.ProtPlayerGold, int32(uint32(0)-cost))
	}
	if runtime.ReportGold != nil {
		runtime.ReportGold(player, playerUnit)
	}
	return ShopBuyComplete5100C0
}

func (s *Server) findInventoryItemByNetCode5108D0(playerUnit *Object, netCode uint16) *Object {
	if playerUnit == nil {
		return nil
	}
	for item := playerUnit.InvFirstItem; item != nil; item = item.InvNextItem {
		// GAME.EXE compares the full object NetCode dword against the
		// zero-extended uint16 request. Do not accept a matching low half.
		if item.NetCode == uint32(netCode) {
			return item
		}
	}
	return nil
}

func shopPlayer5108D0(playerUnit *Object) (*Player, bool) {
	if playerUnit == nil || !playerUnit.Class().Has(object.ClassPlayer) || playerUnit.UpdateData == nil {
		return nil, false
	}
	update := (*PlayerUpdateData)(playerUnit.UpdateData)
	if update.Player == nil {
		return nil, false
	}
	return update.Player, true
}

// shopInventoryItemCost50E3D0 restores the regular/Coop pricing path used by
// single-item sell and repair. It includes native modifier pointers and item
// health, but deliberately rejects guide/reward, ammo, and rechargeable-wand
// branches until their distinct native payloads and balance inputs are sealed.
func shopInventoryItemCost50E3D0(
	session *TradeSession,
	item *Object,
	mode shopPriceMode50E3D0,
	repairCoefficient float32,
) (uint32, bool) {
	if session == nil || item == nil || session.Field16 == 0 {
		return 0, false
	}
	merchant := session.Field8
	if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
		merchant = session.Field12
	}
	if merchant == nil || merchant.InitData == nil || item.Class().Has(object.ClassInfoBook) {
		return 0, false
	}

	price := float64(item.Worth)
	if item.Class().HasAny(shopModifierClassMask50E3D0) {
		if item.InitData == nil {
			return 0, false
		}
		for _, modifier := range item.InitDataModifier().Modifiers {
			if modifier != nil {
				price += float64(modifier.Price20)
			}
		}
	}
	// 0050E3D0 scales ammo weapons and rechargeable wands through layouts
	// separate from ordinary weapon durability. Keep those paths explicit.
	if item.Class().Has(object.ClassWeapon) && uint32(item.SubClass())&0x82 != 0 {
		return 0, false
	}
	if item.Class().Has(object.ClassWand) && uint32(item.SubClass())&0x047f0000 != 0 {
		return 0, false
	}

	fullPrice := price
	idata := merchant.InitDataShopkeeper()
	if mode != shopPriceSell50E3D0 {
		multiplier := float64(idata.BuyMultiplier)
		price *= multiplier
		fullPrice *= multiplier
	}
	if health := item.HealthData; health != nil && health.Max != 0 {
		price = float64(health.Cur) / float64(health.Max) * price
	}
	if mode == shopPriceSell50E3D0 {
		multiplier := float64(idata.SellMultiplier)
		price *= multiplier
		fullPrice *= multiplier
	}
	if price < 1 {
		price = 1
	}
	if fullPrice < 1 {
		fullPrice = 1
	}
	if mode == shopPriceRepair50E3D0 {
		price = float64(repairCoefficient) * (fullPrice - price)
		if price < 1 {
			price = 1
		}
	}
	if math.IsNaN(price) || math.IsInf(price, 0) || price > math.MaxInt32 {
		return 0, false
	}
	return uint32(int32(math.RoundToEven(price))), true
}

func shopSellCandidate5109C0(
	s *Server,
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopSellRuntime5109C0,
) (*Player, *Object, uint32, ShopSellResult5109C0) {
	if session == nil || !s.IsTradeSessionNative(session) {
		return nil, nil, 0, ShopSellUnsupported5109C0
	}
	player, ok := shopPlayer5108D0(playerUnit)
	if !ok {
		return nil, nil, 0, ShopSellUnsupported5109C0
	}
	item := s.findInventoryItemByNetCode5108D0(playerUnit, netCode)
	if item == nil {
		return player, nil, 0, ShopSellNoItem5109C0
	}
	if runtime.ItemIsQuest != nil && runtime.ItemIsQuest(item) {
		if runtime.ReportCannotSellQuest != nil {
			runtime.ReportCannotSellQuest(playerUnit)
		}
		if runtime.PlayRejectSound != nil {
			runtime.PlayRejectSound(playerUnit)
		}
		return player, item, 0, ShopSellQuestItem5109C0
	}
	if runtime.ItemIsGlyph != nil && runtime.ItemIsGlyph(item) {
		if runtime.ReportCannotSellItem != nil {
			runtime.ReportCannotSellItem(playerUnit)
		}
		if runtime.PlayRejectSound != nil {
			runtime.PlayRejectSound(playerUnit)
		}
		return player, item, 0, ShopSellGlyph5109C0
	}
	cost, ok := shopInventoryItemCost50E3D0(session, item, shopPriceSell50E3D0, 0)
	if !ok {
		return player, item, 0, ShopSellUnsupported5109C0
	}
	return player, item, cost, ShopSellQuoted5109C0
}

// QuoteShopSellNative5109C0 restores the C9/1C request and C9/1D response
// half of the original single-item sale.
func (s *Server) QuoteShopSellNative5109C0(
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopSellRuntime5109C0,
) ShopSellResult5109C0 {
	player, item, cost, result := shopSellCandidate5109C0(s, playerUnit, session, netCode, runtime)
	if result != ShopSellQuoted5109C0 {
		return result
	}
	if runtime.SendQuote == nil {
		return ShopSellUnsupported5109C0
	}
	runtime.SendQuote(player, item, cost)
	return ShopSellQuoted5109C0
}

// SellShopItemNative510BE0 restores the C9/18 single-item completion path.
// GAME.EXE detaches and schedules deletion before adding and reporting gold;
// the same observable callback order is preserved here.
func (s *Server) SellShopItemNative510BE0(
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopSellRuntime5109C0,
) ShopSellResult5109C0 {
	player, item, cost, result := shopSellCandidate5109C0(s, playerUnit, session, netCode, runtime)
	if result != ShopSellQuoted5109C0 {
		return result
	}
	if runtime.DetachInventory == nil || runtime.DelayedDelete == nil {
		return ShopSellUnsupported5109C0
	}
	runtime.DetachInventory(playerUnit, item)
	runtime.DelayedDelete(item)
	player.GoldVal += cost
	if runtime.ProtectGold != nil {
		runtime.ProtectGold(player.ProtPlayerGold, int32(cost))
	}
	if runtime.ReportGold != nil {
		runtime.ReportGold(player, playerUnit)
	}
	if runtime.PlaySellSound != nil {
		runtime.PlaySellSound(playerUnit)
	}
	return ShopSellComplete5109C0
}

// QuoteShopRepairNative5108D0 restores the C9/1E request and C9/1F response
// for ordinary health-durability items.
func (s *Server) QuoteShopRepairNative5108D0(
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopRepairRuntime5108D0,
) ShopRepairResult5108D0 {
	if session == nil || !s.IsTradeSessionNative(session) {
		return ShopRepairUnsupported5108D0
	}
	player, ok := shopPlayer5108D0(playerUnit)
	if !ok {
		return ShopRepairUnsupported5108D0
	}
	item := s.findInventoryItemByNetCode5108D0(playerUnit, netCode)
	if item == nil {
		return ShopRepairNoItem5108D0
	}
	health := item.HealthData
	if health == nil || health.Max == 0 || health.Cur == health.Max {
		if runtime.PlayRejectSound != nil {
			runtime.PlayRejectSound(playerUnit)
		}
		return ShopRepairNotDamaged5108D0
	}
	cost, ok := shopInventoryItemCost50E3D0(session, item, shopPriceRepair50E3D0, runtime.RepairCoefficient)
	if !ok || runtime.SendQuote == nil {
		return ShopRepairUnsupported5108D0
	}
	runtime.SendQuote(player, item, cost)
	return ShopRepairQuoted5108D0
}

// RepairShopItemNative510AE0 restores the C9/1A single-item completion path
// for ordinary health durability. Like GAME.EXE, it trusts the prior client
// quote for affordability, clamps gold at zero, then repairs and reports.
func (s *Server) RepairShopItemNative510AE0(
	playerUnit *Object,
	session *TradeSession,
	netCode uint16,
	runtime ShopRepairRuntime5108D0,
) ShopRepairResult5108D0 {
	if session == nil || !s.IsTradeSessionNative(session) {
		return ShopRepairUnsupported5108D0
	}
	player, ok := shopPlayer5108D0(playerUnit)
	if !ok {
		return ShopRepairUnsupported5108D0
	}
	item := s.findInventoryItemByNetCode5108D0(playerUnit, netCode)
	if item == nil {
		return ShopRepairNoItem5108D0
	}
	health := item.HealthData
	if health == nil || health.Max == 0 || runtime.SetHealth == nil {
		return ShopRepairUnsupported5108D0
	}
	cost, ok := shopInventoryItemCost50E3D0(session, item, shopPriceRepair50E3D0, runtime.RepairCoefficient)
	if !ok {
		return ShopRepairUnsupported5108D0
	}
	if player.GoldVal >= cost {
		player.GoldVal -= cost
	} else {
		player.GoldVal = 0
	}
	if runtime.ProtectGold != nil {
		runtime.ProtectGold(player.ProtPlayerGold, -int32(cost))
	}
	runtime.SetHealth(item, health.Max)
	if runtime.ReportHealth != nil {
		runtime.ReportHealth(player, item)
	}
	if runtime.ReportGold != nil {
		runtime.ReportGold(player, playerUnit)
	}
	if runtime.PlayRepairSound != nil {
		runtime.PlayRepairSound(playerUnit)
	}
	return ShopRepairComplete5108D0
}

func simpleShopItemCost50E3D0(session *TradeSession, item *Object) (uint32, bool) {
	if session == nil || item == nil || session.Field16 == 0 {
		return 0, false
	}
	// These classes take modifier, guide/reward, ammo, charge, durability, or
	// category branches in 0050E3D0/0050EEC0. Keep them outside this first
	// native subset, whose items all use the original default category.
	if item.Class().HasAny(shopSpecialClassMask50E3D0) {
		return 0, false
	}
	merchant := session.Field8
	if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
		merchant = session.Field12
	}
	if merchant == nil || merchant.InitData == nil {
		return 0, false
	}
	price := float64(item.Worth) * float64(merchant.InitDataShopkeeper().BuyMultiplier)
	if health := item.HealthData; health != nil && health.Max != 0 {
		price = float64(health.Cur) / float64(health.Max) * price
	}
	if price < 1 {
		price = 1
	}
	if math.IsNaN(price) || price > math.MaxInt32 || price < math.MinInt32 {
		return 0, false
	}
	cost := uint32(int32(math.RoundToEven(price)))
	if cost > shopSimpleMaxCost50EEC0 {
		return 0, false
	}
	return cost, true
}

func hasUnsupportedShopDefinition50E970(def *ShopkeeperItemDefinition) bool {
	if def.Param != 0 {
		return true
	}
	for _, slot := range def.ModifierSlots {
		if slot != 0 {
			return true
		}
	}
	return false
}

// LoadSimpleShopItemsNative50E970 restores the regular-game, unmodified item
// subset of 0050E970. It reports complete=false when the source contains a
// reward parameter, ABI32 modifier slot, unsupported cost branch, or an
// invalid definition count, so callers cannot mistake a partial list for a
// fully restored merchant inventory.
func (s *Server) LoadSimpleShopItemsNative50E970(session *TradeSession) (loaded int, complete bool) {
	if session == nil || !s.IsTradeSessionNative(session) {
		return 0, false
	}
	merchant := session.Field8
	if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
		merchant = session.Field12
	}
	if merchant == nil || merchant.InitData == nil {
		return 0, false
	}
	idata := merchant.InitDataShopkeeper()
	count := int(idata.Count)
	complete = true
	if count > len(idata.Items) {
		count = len(idata.Items)
		complete = false
	}
	for i := 0; i < count; i++ {
		def := &idata.Items[i]
		if hasUnsupportedShopDefinition50E970(def) {
			complete = false
			continue
		}
		for j := 0; j < int(def.Count); j++ {
			item := s.NewObjectByTypeInd(int(def.TypeInd))
			if item == nil {
				complete = false
				continue
			}
			cost, ok := simpleShopItemCost50E3D0(session, item)
			if !ok {
				s.Objs.FreeObject(item)
				complete = false
				continue
			}
			if s.addSimpleShopItemNative50EE00(session, item, cost) == nil {
				s.Objs.FreeObject(item)
				complete = false
				continue
			}
			loaded++
		}
	}
	return loaded, complete
}

// BuildShopItemPacket50F2B0 constructs the exact 18-byte C9/08 merchant-item
// packet: type, object net code, cost, the first health dword, and four
// modifier IDs (or FF for an unmodified item).
func BuildShopItemPacket50F2B0(item *Object, cost uint32) [ShopItemPacketSize50F2B0]byte {
	var packet [ShopItemPacketSize50F2B0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x08
	if item == nil {
		for i := 14; i < len(packet); i++ {
			packet[i] = 0xff
		}
		return packet
	}
	binary.LittleEndian.PutUint16(packet[2:4], item.TypeInd)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(item.NetCode))
	binary.LittleEndian.PutUint32(packet[6:10], cost)
	if health := item.HealthData; health != nil {
		binary.LittleEndian.PutUint16(packet[10:12], health.Cur)
		binary.LittleEndian.PutUint16(packet[12:14], health.Field2)
	}
	for i := 14; i < len(packet); i++ {
		packet[i] = 0xff
	}
	if item.Class().HasAny(shopModifierClassMask50E3D0) && item.InitData != nil {
		for i, modifier := range item.InitDataModifier().Modifiers {
			if modifier != nil {
				packet[14+i] = byte(modifier.Index())
			}
		}
	}
	return packet
}

func BuildShopItemRemovePacket50E820(item *Object) [ShopItemRemovePacketSize50E820]byte {
	var packet [ShopItemRemovePacketSize50E820]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x09
	if item != nil {
		binary.LittleEndian.PutUint16(packet[2:4], uint16(item.NetCode))
	}
	return packet
}

func BuildShopMissingGoldPacket5104F0(amount uint16) [ShopMissingGoldPacketSize5104F0]byte {
	var packet [ShopMissingGoldPacketSize5104F0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x1b
	binary.LittleEndian.PutUint16(packet[2:4], amount)
	return packet
}

func BuildShopGoldReportPacket4D8870(gold uint32) [ShopGoldReportPacketSize4D8870]byte {
	var packet [ShopGoldReportPacketSize4D8870]byte
	packet[0] = byte(netmsg.MSG_REPORT_GOLD)
	binary.LittleEndian.PutUint32(packet[1:5], gold)
	return packet
}

func BuildShopSellQuotePacket5109C0(item *Object, cost uint32) [ShopSellQuotePacketSize5109C0]byte {
	var packet [ShopSellQuotePacketSize5109C0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x1d
	if item != nil {
		binary.LittleEndian.PutUint16(packet[2:4], uint16(item.NetCode))
	}
	binary.LittleEndian.PutUint32(packet[4:8], cost)
	return packet
}

func BuildShopRepairQuotePacket5108D0(item *Object, cost uint32) [ShopRepairQuotePacketSize5108D0]byte {
	var packet [ShopRepairQuotePacketSize5108D0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x1f
	if item != nil {
		binary.LittleEndian.PutUint16(packet[2:4], uint16(item.NetCode))
	}
	binary.LittleEndian.PutUint32(packet[4:8], cost)
	return packet
}

func BuildShopItemHealthPacket4D87A0(item *Object) [ShopItemHealthPacketSize4D87A0]byte {
	var packet [ShopItemHealthPacketSize4D87A0]byte
	packet[0] = byte(netmsg.MSG_REPORT_ITEM_HEALTH)
	if item == nil {
		return packet
	}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(item.NetCode))
	if health := item.HealthData; health != nil {
		binary.LittleEndian.PutUint16(packet[3:5], health.Cur)
		binary.LittleEndian.PutUint16(packet[5:7], health.Max)
	}
	return packet
}

func shopObjectNameKey4E39F0(objectID, typeID string) string {
	id := objectID
	if id == "" {
		id = typeID
	}
	if i := strings.LastIndexByte(id, ':'); i >= 0 {
		id = id[i+1:]
	}
	return "NPC:" + strings.ReplaceAll(id, "_", "")
}

// ShopObjectName4E39F0 reproduces the original object database display-name
// lookup without passing a native-width Object pointer through the PE32 helper.
func (s *Server) ShopObjectName4E39F0(obj *Object) string {
	if obj == nil {
		return ""
	}
	typeID := ""
	if typ := s.Types.ByInd(int(obj.TypeInd)); typ != nil {
		typeID = typ.ID()
	}
	key := shopObjectNameKey4E39F0(obj.ID(), typeID)
	return s.Strings().GetStringInFile(strman.ID(key), "C:\\NoxPost\\src\\Server\\DBase\\objdb.c")
}

// BuildShopStartPacket50F0F0 constructs the original 86-byte C9/0D packet:
// merchant type, 24 UTF-16LE name code units plus terminator, and a bounded
// 32-byte shop-text C string.
func BuildShopStartPacket50F0F0(typeInd uint16, name string, shopText []byte) [ShopStartPacketSize50F0F0]byte {
	var packet [ShopStartPacketSize50F0F0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x0d
	binary.LittleEndian.PutUint16(packet[2:4], typeInd)
	name16 := utf16.Encode([]rune(name))
	if len(name16) > 24 {
		name16 = name16[:24]
	}
	for i, ch := range name16 {
		binary.LittleEndian.PutUint16(packet[4+2*i:6+2*i], ch)
	}
	// packet[52:54] remains the explicit UTF-16 terminator.
	for i, ch := range shopText {
		if i >= 31 || ch == 0 {
			break
		}
		packet[54+i] = ch
	}
	return packet
}
