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
	ShopStartPacketSize50F0F0 = 86
	ShopItemPacketSize50F2B0  = 18

	shopModifierClassMask50E3D0 = object.ClassWand | object.ClassWeapon | object.ClassArmor | object.ClassFlag
	shopSpecialClassMask50E3D0  = shopModifierClassMask50E3D0 | object.ClassInfoBook
	shopSimpleMaxCost50EEC0     = uint32(0x00ffffff)
)

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
