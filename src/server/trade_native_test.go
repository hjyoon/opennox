package server

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

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
