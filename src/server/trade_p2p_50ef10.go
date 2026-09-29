package server

import (
	"encoding/binary"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	TradeP2PStartPacketSize50F1A0   = 52
	TradeP2PClosePacketSize50F450   = 2
	TradeP2PStatusPacketSize50F720  = 3
	TradeP2PBalancePacketSize50FA00 = 14
	TradeP2PItemPacketSize50FAE0    = 15
	TradeP2PRemovePacketSize50FF90  = 4
)

type TradeP2PStartResult50EF10 uint8

const (
	TradeP2PStartInvalid50EF10 TradeP2PStartResult50EF10 = iota
	TradeP2PStartStarterBusy50EF10
	TradeP2PStartOtherBusy50EF10
	TradeP2PStartComplete50EF10
	TradeP2PStartSamePartner50EF10
)

// TradeP2PRuntime50F3A0 contains the object-bearing services around the
// native player-to-player transaction. The session and list nodes themselves
// remain owned by Server's native-width trade registry.
type TradeP2PRuntime50F3A0 struct {
	Send         func(*Player, []byte)
	SendClose    func(*Player, []byte)
	PutInventory func(player, item *Object)
	AddGold      func(player *Object, amount uint32)
}

// BuildTradeP2PStartPacket50F1A0 constructs the original C9/0C packet. The
// peer name occupies 24 UTF-16LE code units followed by the fixed terminator.
func BuildTradeP2PStartPacket50F1A0(name string) [TradeP2PStartPacketSize50F1A0]byte {
	var packet [TradeP2PStartPacketSize50F1A0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x0c
	name16 := utf16.Encode([]rune(name))
	if len(name16) > 24 {
		name16 = name16[:24]
	}
	for i, ch := range name16 {
		binary.LittleEndian.PutUint16(packet[2+2*i:4+2*i], ch)
	}
	return packet
}

func BuildTradeP2PClosePacket50F450() [TradeP2PClosePacketSize50F450]byte {
	return [TradeP2PClosePacketSize50F450]byte{byte(netmsg.MSG_TRADE), 0x01}
}

func BuildTradeP2PStatusPacket50F720(session *TradeSession, recipient *Object) [TradeP2PStatusPacketSize50F720]byte {
	packet := [TradeP2PStatusPacketSize50F720]byte{byte(netmsg.MSG_TRADE), 0x03, 0}
	if session == nil {
		return packet
	}
	if session.Field24 != 0 {
		packet[2] = 1
		if session.Field8 != recipient {
			packet[2] = 2
		}
	}
	if session.Field28 != 0 {
		if session.Field12 == recipient {
			packet[2] |= 1
		} else {
			packet[2] |= 2
		}
	}
	return packet
}

func tradeP2PGoldAmount50FA00(obj *Object) uint32 {
	if obj == nil || obj.InitData == nil {
		return 0
	}
	return obj.InitDataGold().Amount
}

func BuildTradeP2PBalancePacket50FA00(session *TradeSession, recipient *Object) [TradeP2PBalancePacketSize50FA00]byte {
	packet := [TradeP2PBalancePacketSize50FA00]byte{byte(netmsg.MSG_TRADE), 0x06}
	if session == nil {
		return packet
	}
	if session.Field8 == recipient {
		binary.LittleEndian.PutUint32(packet[2:6], tradeP2PGoldAmount50FA00(session.Field48))
		// Both participants are players, so GAME.EXE deliberately reports a
		// zero price delta here rather than balancing either private Gold item.
		binary.LittleEndian.PutUint32(packet[10:14], tradeP2PGoldAmount50FA00(session.Field52))
	} else if session.Field12 == recipient {
		binary.LittleEndian.PutUint32(packet[2:6], tradeP2PGoldAmount50FA00(session.Field52))
		binary.LittleEndian.PutUint32(packet[10:14], tradeP2PGoldAmount50FA00(session.Field48))
	}
	return packet
}

func BuildTradeP2PItemPacket50FAE0(recipient, offeringPlayer, item *Object, cost uint32) [TradeP2PItemPacketSize50FAE0]byte {
	packet := [TradeP2PItemPacketSize50FAE0]byte{byte(netmsg.MSG_TRADE), 0x04}
	if recipient == offeringPlayer {
		packet[2] = 1
	}
	for i := 11; i < len(packet); i++ {
		packet[i] = 0xff
	}
	if item == nil {
		return packet
	}
	binary.LittleEndian.PutUint16(packet[3:5], item.TypeInd)
	binary.LittleEndian.PutUint16(packet[5:7], uint16(item.NetCode))
	binary.LittleEndian.PutUint32(packet[7:11], cost)
	if item.Class().HasAny(shopModifierClassMask50E3D0) && item.InitData != nil {
		for i, modifier := range item.InitDataModifier().Modifiers {
			if modifier != nil {
				packet[11+i] = byte(modifier.Index())
			}
		}
	}
	return packet
}

func BuildTradeP2PRemovePacket50FF90(item *Object) [TradeP2PRemovePacketSize50FF90]byte {
	packet := [TradeP2PRemovePacketSize50FF90]byte{byte(netmsg.MSG_TRADE), 0x05}
	if item != nil {
		binary.LittleEndian.PutUint16(packet[2:4], uint16(item.NetCode))
	}
	return packet
}

func tradeP2PPlayer50EF10(unit *Object) (*PlayerUpdateData, *Player, bool) {
	if unit == nil || !unit.Class().Has(object.ClassPlayer) || unit.UpdateData == nil {
		return nil, nil, false
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	if update.Player == nil {
		return nil, nil, false
	}
	return update, update.Player, true
}

// StartP2PTradeNative50EF10 restores the player/player branch of 0050EF10.
// It records native Object pointers in both PlayerUpdateData records and sends
// each participant the other player's approved name.
func (s *Server) StartP2PTradeNative50EF10(
	starter, other *Object,
	frame uint32,
	send func(*Player, []byte),
) (*TradeSession, TradeP2PStartResult50EF10) {
	starterUpdate, starterPlayer, ok := tradeP2PPlayer50EF10(starter)
	if !ok {
		return nil, TradeP2PStartInvalid50EF10
	}
	otherUpdate, otherPlayer, ok := tradeP2PPlayer50EF10(other)
	if !ok || starter == other {
		return nil, TradeP2PStartInvalid50EF10
	}
	if active := starterUpdate.Trade70; active != nil {
		// The original start routine quietly ignores another request for the
		// participant already attached to this session. Only a different
		// target produces StarterAlreadyTrading feedback.
		if s.IsTradeSessionNative(active) && active.Field12 == other {
			return nil, TradeP2PStartSamePartner50EF10
		}
		return nil, TradeP2PStartStarterBusy50EF10
	}
	if otherUpdate.Trade70 != nil {
		return nil, TradeP2PStartOtherBusy50EF10
	}

	session := s.NewTradeSessionNative50E870()
	if session == nil {
		return nil, TradeP2PStartInvalid50EF10
	}
	session.Field0 = 1
	session.Field4 = frame
	session.Field8 = starter
	session.Field12 = other
	session.Field16 = 0
	starterUpdate.Trade70 = session
	otherUpdate.Trade70 = session
	if send != nil {
		starterPacket := BuildTradeP2PStartPacket50F1A0(otherPlayer.Name())
		send(starterPlayer, starterPacket[:])
		otherPacket := BuildTradeP2PStartPacket50F1A0(starterPlayer.Name())
		send(otherPlayer, otherPacket[:])
	}
	return session, TradeP2PStartComplete50EF10
}

func (s *Server) nativeP2PSession50F3A0(session *TradeSession) (*nativeTradeSessionAllocation, bool) {
	if session == nil || s.tradeNative.sessions == nil {
		return nil, false
	}
	state, ok := s.tradeNative.sessions[session]
	if !ok || session.Field16 != 0 {
		return nil, false
	}
	return state, true
}

func tradeP2PSend50F3A0(runtime TradeP2PRuntime50F3A0, unit *Object, packet []byte) {
	if runtime.Send == nil {
		return
	}
	_, player, ok := tradeP2PPlayer50EF10(unit)
	if ok {
		runtime.Send(player, packet)
	}
}

func tradeP2PSendClose50F450(runtime TradeP2PRuntime50F3A0, unit *Object, packet []byte) {
	if runtime.SendClose == nil {
		return
	}
	_, player, ok := tradeP2PPlayer50EF10(unit)
	if ok {
		runtime.SendClose(player, packet)
	}
}

func tradeP2PClearPlayer50F490(session *TradeSession, unit *Object) {
	if update, _, ok := tradeP2PPlayer50EF10(unit); ok && update.Trade70 == session {
		update.Trade70 = nil
	}
}

func tradeP2PReturnList50F3A0(owner *Object, head *TradeItem, runtime TradeP2PRuntime50F3A0) {
	if runtime.PutInventory == nil {
		return
	}
	for node := head; node != nil; node = node.Field8 {
		if node.Item0 != nil {
			runtime.PutInventory(owner, node.Item0)
		}
	}
}

func tradeP2PAddGold50F3A0(owner, gold *Object, runtime TradeP2PRuntime50F3A0) {
	amount := tradeP2PGoldAmount50FA00(gold)
	if amount != 0 && runtime.AddGold != nil {
		runtime.AddGold(owner, amount)
	}
}

// CancelP2PTradeNative50F3A0 returns both offer lists, restores any private
// Gold amounts, clears both player links, emits C9/01, and releases the native
// session. Ownership is checked before any session field is dereferenced.
func (s *Server) CancelP2PTradeNative50F3A0(session *TradeSession, runtime TradeP2PRuntime50F3A0) bool {
	if _, ok := s.nativeP2PSession50F3A0(session); !ok {
		return false
	}
	tradeP2PReturnList50F3A0(session.Field8, session.Field32, runtime)
	tradeP2PAddGold50F3A0(session.Field8, session.Field48, runtime)
	tradeP2PReturnList50F3A0(session.Field12, session.Field36, runtime)
	tradeP2PAddGold50F3A0(session.Field12, session.Field52, runtime)
	closePacket := BuildTradeP2PClosePacket50F450()
	tradeP2PSendClose50F450(runtime, session.Field8, closePacket[:])
	tradeP2PSendClose50F450(runtime, session.Field12, closePacket[:])
	tradeP2PClearPlayer50F490(session, session.Field8)
	tradeP2PClearPlayer50F490(session, session.Field12)
	session.Field0 = 0
	return s.ReleaseTradeSessionNative510000(session)
}

func tradeP2PCanAddOffer50FD60(head *TradeItem, item *Object) bool {
	if item == nil {
		return false
	}
	var types [4]uint16
	unique := 0
	for node := head; node != nil; node = node.Field8 {
		if node.Item0 == nil {
			continue
		}
		typeInd := node.Item0.TypeInd
		found := false
		for i := 0; i < unique; i++ {
			if types[i] == typeInd {
				found = true
				break
			}
		}
		if !found && unique < len(types) {
			types[unique] = typeInd
			unique++
		}
	}
	if unique == 0 {
		return true
	}
	for i := 0; i < unique; i++ {
		if types[i] == item.TypeInd && !item.Class().HasAny(shopModifierClassMask50E3D0) {
			return true
		}
	}
	return unique < len(types)
}

func (s *Server) addP2POfferNode50F820(session *TradeSession, item *Object, cost uint32, first bool) *TradeItem {
	state, ok := s.nativeP2PSession50F3A0(session)
	if !ok || item == nil {
		return nil
	}
	node, freeNode := alloc.New(TradeItem{})
	node.Item0 = item
	node.Cost4 = cost
	head := &session.Field36
	if first {
		head = &session.Field32
	}
	node.Field8 = *head
	if node.Field8 != nil {
		node.Field8.Field12 = node
	}
	*head = node
	state.items[node] = nativeTradeItemAllocation{freeNode: freeNode}
	return node
}

func tradeP2PNotifyOffer50F820(session *TradeSession, offeringPlayer, item *Object, cost uint32, runtime TradeP2PRuntime50F3A0) {
	for _, recipient := range [...]*Object{session.Field8, session.Field12} {
		balance := BuildTradeP2PBalancePacket50FA00(session, recipient)
		tradeP2PSend50F3A0(runtime, recipient, balance[:])
		itemPacket := BuildTradeP2PItemPacket50FAE0(recipient, offeringPlayer, item, cost)
		tradeP2PSend50F3A0(runtime, recipient, itemPacket[:])
		status := BuildTradeP2PStatusPacket50F720(session, recipient)
		tradeP2PSend50F3A0(runtime, recipient, status[:])
	}
}

// AddP2PTradeOfferNative50F820 links an inventory Object into the correct
// native offer list. The caller detaches the item only when this returns true,
// matching the decoder's original operation order.
func (s *Server) AddP2PTradeOfferNative50F820(
	session *TradeSession,
	offeringPlayer, item *Object,
	runtime TradeP2PRuntime50F3A0,
) bool {
	if _, ok := s.nativeP2PSession50F3A0(session); !ok || item == nil {
		return false
	}
	first := session.Field8 == offeringPlayer
	if !first && session.Field12 != offeringPlayer {
		return false
	}
	head := session.Field36
	if first {
		head = session.Field32
	}
	if !tradeP2PCanAddOffer50FD60(head, item) {
		return false
	}
	cost := uint32(s.ShopItemCostNoMerchant50E3D0(item))
	if s.addP2POfferNode50F820(session, item, cost, first) == nil {
		return false
	}
	session.Field24 = 0
	session.Field28 = 0
	tradeP2PNotifyOffer50F820(session, offeringPlayer, item, cost, runtime)
	return true
}

func tradeP2PFindOffer50FFE0(head *TradeItem, netCode uint32) *TradeItem {
	for node := head; node != nil; node = node.Field8 {
		if node.Item0 != nil && node.Item0.NetCode == netCode {
			return node
		}
	}
	return nil
}

func (s *Server) unlinkP2POffer50FE20(session *TradeSession, node *TradeItem, first bool) bool {
	state, ok := s.nativeP2PSession50F3A0(session)
	if !ok {
		return false
	}
	allocation, ok := state.items[node]
	if !ok {
		return false
	}
	head := &session.Field36
	if first {
		head = &session.Field32
	}
	if node.Field8 != nil {
		node.Field8.Field12 = node.Field12
	}
	if node.Field12 != nil {
		node.Field12.Field8 = node.Field8
	} else if *head == node {
		*head = node.Field8
	} else {
		return false
	}
	delete(state.items, node)
	if allocation.freeNode != nil {
		allocation.freeNode()
	}
	return true
}

// RemoveP2PTradeOfferNative50FE20 returns the selected Object to its original
// participant and broadcasts the exact balance/remove/status packet sequence.
func (s *Server) RemoveP2PTradeOfferNative50FE20(
	session *TradeSession,
	netCode uint32,
	runtime TradeP2PRuntime50F3A0,
) bool {
	if _, ok := s.nativeP2PSession50F3A0(session); !ok {
		return false
	}
	first := true
	node := tradeP2PFindOffer50FFE0(session.Field32, netCode)
	owner := session.Field8
	if node == nil {
		first = false
		node = tradeP2PFindOffer50FFE0(session.Field36, netCode)
		owner = session.Field12
	}
	if node == nil {
		return false
	}
	item := node.Item0
	if runtime.PutInventory != nil && item != nil {
		runtime.PutInventory(owner, item)
	}
	if !s.unlinkP2POffer50FE20(session, node, first) {
		return false
	}
	session.Field24 = 0
	session.Field28 = 0
	for _, recipient := range [...]*Object{session.Field8, session.Field12} {
		balance := BuildTradeP2PBalancePacket50FA00(session, recipient)
		tradeP2PSend50F3A0(runtime, recipient, balance[:])
		removed := BuildTradeP2PRemovePacket50FF90(item)
		tradeP2PSend50F3A0(runtime, recipient, removed[:])
		status := BuildTradeP2PStatusPacket50F720(session, recipient)
		tradeP2PSend50F3A0(runtime, recipient, status[:])
	}
	return true
}

func (s *Server) transferP2POfferList50F790(session *TradeSession, recipient *Object, head **TradeItem, runtime TradeP2PRuntime50F3A0) {
	state := s.tradeNative.sessions[session]
	for node := *head; node != nil; {
		next := node.Field8
		if runtime.PutInventory != nil && node.Item0 != nil {
			runtime.PutInventory(recipient, node.Item0)
		}
		if allocation, ok := state.items[node]; ok {
			delete(state.items, node)
			if allocation.freeNode != nil {
				allocation.freeNode()
			}
		}
		node = next
	}
	*head = nil
}

// AcceptP2PTradeNative50F5A0 marks the accepting side, reports both flags,
// and atomically swaps the two offer lists once both players accept.
func (s *Server) AcceptP2PTradeNative50F5A0(
	session *TradeSession,
	acceptingPlayer *Object,
	runtime TradeP2PRuntime50F3A0,
) bool {
	if _, ok := s.nativeP2PSession50F3A0(session); !ok {
		return false
	}
	if session.Field8 == acceptingPlayer {
		session.Field24 = 1
	} else if session.Field12 == acceptingPlayer {
		session.Field28 = 1
	} else {
		return false
	}
	for _, recipient := range [...]*Object{session.Field8, session.Field12} {
		status := BuildTradeP2PStatusPacket50F720(session, recipient)
		tradeP2PSend50F3A0(runtime, recipient, status[:])
	}
	if session.Field24 != 1 || session.Field28 != 1 {
		return true
	}
	s.transferP2POfferList50F790(session, session.Field12, &session.Field32, runtime)
	s.transferP2POfferList50F790(session, session.Field8, &session.Field36, runtime)
	tradeP2PAddGold50F3A0(session.Field12, session.Field48, runtime)
	tradeP2PAddGold50F3A0(session.Field8, session.Field52, runtime)
	if session.Field48 != nil && session.Field48.InitData != nil {
		session.Field48.InitDataGold().Amount = 0
	}
	if session.Field52 != nil && session.Field52.InitData != nil {
		session.Field52.InitDataGold().Amount = 0
	}
	session.Field40 = 0
	session.Field44 = 0
	return s.CancelP2PTradeNative50F3A0(session, runtime)
}
