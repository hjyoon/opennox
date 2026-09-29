package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
)

func newTradeP2PTestPlayer50EF10(t *testing.T, name string, index byte) (*Object, *PlayerUpdateData, *Player) {
	t.Helper()
	player := nativeTradeTestValue(t, Player{PlayerInd: index})
	player.SetName(name)
	update := nativeTradeTestValue(t, PlayerUpdateData{Player: player})
	unit := nativeTradeTestValue(t, Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	})
	player.PlayerUnit = unit
	return unit, update, player
}

type tradeP2PTestPacket50EF10 struct {
	player *Player
	data   []byte
}

func tradeP2PTestRuntime50EF10(
	packets *[]tradeP2PTestPacket50EF10,
	puts *map[*Object][]*Object,
) TradeP2PRuntime50F3A0 {
	send := func(player *Player, packet []byte) {
		*packets = append(*packets, tradeP2PTestPacket50EF10{
			player: player,
			data:   append([]byte(nil), packet...),
		})
	}
	return TradeP2PRuntime50F3A0{
		Send:      send,
		SendClose: send,
		PutInventory: func(player, item *Object) {
			(*puts)[player] = append((*puts)[player], item)
		},
		AddGold: func(player *Object, amount uint32) {
			update := (*PlayerUpdateData)(player.UpdateData)
			update.Player.GoldVal += amount
		},
	}
}

func TestBuildTradeP2PStartPacket50F1A0(t *testing.T) {
	packet := BuildTradeP2PStartPacket50F1A0("Dúnmír-곰-12345678901234567890")
	if packet[0] != byte(netmsg.MSG_TRADE) || packet[1] != 0x0c {
		t.Fatalf("header = % x, want c9 0c", packet[:2])
	}
	want := utf16.Encode([]rune("Dúnmír-곰-12345678901234567890"))[:24]
	for i, ch := range want {
		if got := binary.LittleEndian.Uint16(packet[2+2*i : 4+2*i]); got != ch {
			t.Fatalf("name unit %d = %#x, want %#x", i, got, ch)
		}
	}
	if got := binary.LittleEndian.Uint16(packet[50:52]); got != 0 {
		t.Fatalf("terminator = %#x, want 0", got)
	}
}

func TestP2PTradeNativeLifecycle50EF10(t *testing.T) {
	first, firstUpdate, firstPlayer := newTradeP2PTestPlayer50EF10(t, "Conjurer", 3)
	second, secondUpdate, secondPlayer := newTradeP2PTestPlayer50EF10(t, "Dun Mir", 7)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(first)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(second)) <= uintptr(^uint32(0))) {
		t.Fatal("test participants did not exercise native high addresses")
	}

	var packets []tradeP2PTestPacket50EF10
	puts := make(map[*Object][]*Object)
	runtime := tradeP2PTestRuntime50EF10(&packets, &puts)
	s := &Server{}
	session, result := s.StartP2PTradeNative50EF10(first, second, 0x12345678, runtime.Send)
	if result != TradeP2PStartComplete50EF10 || session == nil {
		t.Fatalf("start = (%p, %d), want complete", session, result)
	}
	if session.Field8 != first || session.Field12 != second || session.Field4 != 0x12345678 || session.Field16 != 0 || session.Field0 != 1 {
		t.Fatalf("session fields = %+v", session)
	}
	if firstUpdate.Trade70 != session || secondUpdate.Trade70 != session || !s.IsTradeSessionNative(session) {
		t.Fatal("native session was not attached to both players")
	}
	if len(packets) != 2 || packets[0].player != firstPlayer || packets[1].player != secondPlayer {
		t.Fatalf("start packets = %+v", packets)
	}
	wantStarterPacket := BuildTradeP2PStartPacket50F1A0("Dun Mir")
	if got := packets[0].data; !reflect.DeepEqual(got, wantStarterPacket[:]) {
		t.Fatalf("starter packet = % x", got)
	}
	wantOtherPacket := BuildTradeP2PStartPacket50F1A0("Conjurer")
	if got := packets[1].data; !reflect.DeepEqual(got, wantOtherPacket[:]) {
		t.Fatalf("other packet = % x", got)
	}

	firstItem := nativeTradeTestValue(t, Object{TypeInd: 101, ObjClass: object.ClassFood, NetCode: 0x2345, Worth: 25})
	secondItem := nativeTradeTestValue(t, Object{TypeInd: 202, ObjClass: object.ClassFood, NetCode: 0x4567, Worth: 40})
	packets = nil
	if !s.AddP2PTradeOfferNative50F820(session, first, firstItem, runtime) ||
		!s.AddP2PTradeOfferNative50F820(session, second, secondItem, runtime) {
		t.Fatal("player offers were rejected")
	}
	if session.Field32 == nil || session.Field32.Item0 != firstItem || session.Field32.Cost4 != 25 ||
		session.Field36 == nil || session.Field36.Item0 != secondItem || session.Field36.Cost4 != 40 {
		t.Fatalf("offer lists = %+v / %+v", session.Field32, session.Field36)
	}
	if len(packets) != 12 {
		t.Fatalf("offer packet count = %d, want 12", len(packets))
	}
	for i, opcode := range []byte{0x06, 0x04, 0x03, 0x06, 0x04, 0x03} {
		if packets[i].data[1] != opcode {
			t.Fatalf("first offer packet %d opcode = %#x, want %#x", i, packets[i].data[1], opcode)
		}
	}
	if packets[1].data[2] != 1 || packets[4].data[2] != 0 {
		t.Fatalf("offer ownership flags = %d/%d, want 1/0", packets[1].data[2], packets[4].data[2])
	}

	packets = nil
	if !s.AcceptP2PTradeNative50F5A0(session, first, runtime) {
		t.Fatal("first acceptance failed")
	}
	if session.Field24 != 1 || session.Field28 != 0 || len(packets) != 2 || packets[0].data[2] != 1 || packets[1].data[2] != 2 {
		t.Fatalf("first acceptance state = %d/%d packets %+v", session.Field24, session.Field28, packets)
	}
	if !s.AcceptP2PTradeNative50F5A0(session, second, runtime) {
		t.Fatal("second acceptance failed")
	}
	if got := puts[second]; !reflect.DeepEqual(got, []*Object{firstItem}) {
		t.Fatalf("second inventory additions = %p, want first item", got)
	}
	if got := puts[first]; !reflect.DeepEqual(got, []*Object{secondItem}) {
		t.Fatalf("first inventory additions = %p, want second item", got)
	}
	if firstUpdate.Trade70 != nil || secondUpdate.Trade70 != nil || s.IsTradeSessionNative(session) {
		t.Fatal("completed P2P trade remained active")
	}
	if len(packets) != 6 || packets[4].data[1] != 0x01 || packets[5].data[1] != 0x01 {
		t.Fatalf("completion packets = %+v", packets)
	}
}

func TestP2PTradeRemoveAndCancel50FE20(t *testing.T) {
	first, firstUpdate, _ := newTradeP2PTestPlayer50EF10(t, "First", 1)
	second, secondUpdate, _ := newTradeP2PTestPlayer50EF10(t, "Second", 2)
	var packets []tradeP2PTestPacket50EF10
	puts := make(map[*Object][]*Object)
	runtime := tradeP2PTestRuntime50EF10(&packets, &puts)
	s := &Server{}
	session, _ := s.StartP2PTradeNative50EF10(first, second, 9, runtime.Send)
	firstItem := nativeTradeTestValue(t, Object{TypeInd: 1, ObjClass: object.ClassFood, NetCode: 11, Worth: 3})
	secondItem := nativeTradeTestValue(t, Object{TypeInd: 2, ObjClass: object.ClassFood, NetCode: 22, Worth: 4})
	if !s.AddP2PTradeOfferNative50F820(session, first, firstItem, runtime) ||
		!s.AddP2PTradeOfferNative50F820(session, second, secondItem, runtime) {
		t.Fatal("offer setup failed")
	}
	packets = nil
	if !s.RemoveP2PTradeOfferNative50FE20(session, firstItem.NetCode, runtime) {
		t.Fatal("offer removal failed")
	}
	if session.Field32 != nil || !reflect.DeepEqual(puts[first], []*Object{firstItem}) {
		t.Fatalf("removed offer state = %p / %p", session.Field32, puts[first])
	}
	if len(packets) != 6 || packets[1].data[1] != 0x05 || binary.LittleEndian.Uint16(packets[1].data[2:4]) != 11 {
		t.Fatalf("remove packets = %+v", packets)
	}
	packets = nil
	if !s.CancelP2PTradeNative50F3A0(session, runtime) {
		t.Fatal("trade cancellation failed")
	}
	if !reflect.DeepEqual(puts[second], []*Object{secondItem}) || firstUpdate.Trade70 != nil || secondUpdate.Trade70 != nil {
		t.Fatalf("cancel state = first %p second %p links %p/%p", puts[first], puts[second], firstUpdate.Trade70, secondUpdate.Trade70)
	}
	if len(packets) != 2 || packets[0].data[1] != 0x01 || packets[1].data[1] != 0x01 {
		t.Fatalf("cancel packets = %+v", packets)
	}
}

func TestP2PTradeStartBusy50EF10(t *testing.T) {
	first, firstUpdate, _ := newTradeP2PTestPlayer50EF10(t, "First", 1)
	second, secondUpdate, _ := newTradeP2PTestPlayer50EF10(t, "Second", 2)
	s := &Server{}
	firstUpdate.Trade70 = nativeTradeTestValue(t, TradeSession{})
	if session, result := s.StartP2PTradeNative50EF10(first, second, 1, nil); session != nil || result != TradeP2PStartStarterBusy50EF10 {
		t.Fatalf("starter-busy result = %p/%d", session, result)
	}
	firstUpdate.Trade70 = nil
	secondUpdate.Trade70 = nativeTradeTestValue(t, TradeSession{})
	if session, result := s.StartP2PTradeNative50EF10(first, second, 1, nil); session != nil || result != TradeP2PStartOtherBusy50EF10 {
		t.Fatalf("other-busy result = %p/%d", session, result)
	}
}

func TestP2PTradeStartSamePartnerIsQuiet50EF10(t *testing.T) {
	first, firstUpdate, _ := newTradeP2PTestPlayer50EF10(t, "First", 1)
	second, secondUpdate, _ := newTradeP2PTestPlayer50EF10(t, "Second", 2)
	var packets []tradeP2PTestPacket50EF10
	puts := make(map[*Object][]*Object)
	runtime := tradeP2PTestRuntime50EF10(&packets, &puts)
	s := &Server{}
	session, result := s.StartP2PTradeNative50EF10(first, second, 1, runtime.Send)
	if session == nil || result != TradeP2PStartComplete50EF10 {
		t.Fatalf("initial start = %p/%d, want complete", session, result)
	}

	again, result := s.StartP2PTradeNative50EF10(first, second, 2, runtime.Send)
	if again != nil || result != TradeP2PStartSamePartner50EF10 {
		t.Fatalf("same-partner start = %p/%d, want quiet refusal", again, result)
	}
	if firstUpdate.Trade70 != session || secondUpdate.Trade70 != session || len(packets) != 2 {
		t.Fatalf("existing session changed: links %p/%p packets %d", firstUpdate.Trade70, secondUpdate.Trade70, len(packets))
	}
}
