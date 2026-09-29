package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestNetworkNewAliasContractOrderAndNativeIdentity51BAD0(t *testing.T) {
	const player = uint64(0x7f11223344556677)
	events := make([]string, 0, 7)
	hooks := networkNewAliasHooks51BAD0[uint64]{
		loadAlias: func() byte {
			events = append(events, "alias")
			return 0xfe
		},
		loadCode: func() uint16 {
			events = append(events, "code")
			return 0x1234
		},
		loadTypeID: func() uint16 {
			events = append(events, "type")
			return 0x5678
		},
		loadDeadline: func() uint32 {
			events = append(events, "deadline")
			return 0x9abcdef0
		},
		storeCode: func(gotPlayer uint64, alias byte, value uint16) {
			events = append(events, "store-code")
			if gotPlayer != player || alias != 0xfe || value != 0x1234 {
				t.Fatalf("code store = (%#x,%#x,%#x)", gotPlayer, alias, value)
			}
		},
		storeTypeID: func(gotPlayer uint64, alias byte, value uint16) {
			events = append(events, "store-type")
			if gotPlayer != player || alias != 0xfe || value != 0x5678 {
				t.Fatalf("type store = (%#x,%#x,%#x)", gotPlayer, alias, value)
			}
		},
		storeDeadline: func(gotPlayer uint64, alias byte, value uint32) {
			events = append(events, "store-deadline")
			if gotPlayer != player || alias != 0xfe || value != 0x9abcdef0 {
				t.Fatalf("deadline store = (%#x,%#x,%#x)", gotPlayer, alias, value)
			}
		},
	}
	if got := networkNewAlias51BAD0(player, hooks); got != networkNewAliasPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, networkNewAliasPacketSize51BAD0)
	}
	want := []string{"alias", "code", "store-code", "type", "store-type", "deadline", "store-deadline"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkNewAliasReservedEscapeDoesNotStore51BAD0(t *testing.T) {
	events := make([]string, 0, 2)
	hooks := networkNewAliasHooks51BAD0[uint64]{
		loadAlias: func() byte {
			events = append(events, "alias")
			return 0xff
		},
		loadCode:      func() uint16 { t.Fatal("unexpected code read"); return 0 },
		loadTypeID:    func() uint16 { t.Fatal("unexpected type read"); return 0 },
		loadDeadline:  func() uint32 { t.Fatal("unexpected deadline read"); return 0 },
		storeCode:     func(uint64, byte, uint16) { t.Fatal("unexpected code store") },
		storeTypeID:   func(uint64, byte, uint16) { t.Fatal("unexpected type store") },
		storeDeadline: func(uint64, byte, uint32) { t.Fatal("unexpected deadline store") },
	}
	if got := networkNewAlias51BAD0(uint64(7), hooks); got != networkNewAliasPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, networkNewAliasPacketSize51BAD0)
	}
	if !reflect.DeepEqual(events, []string{"alias"}) {
		t.Fatalf("events = %v, want alias only", events)
	}
}

func TestNetworkNewAliasNativePlayerTable51BAD0(t *testing.T) {
	player := &Player{}
	player.NetData16[0xfd] = PlayerNetData{Field0: 1, Field2: 2, Frame4: 3}
	player.NetData16[0xfe] = PlayerNetData{Field0: 4, Field2: 5, Frame4: 6}
	packet := &[NetworkNewAliasPacketSize51BAD0]byte{0: 0xa5, 1: 0xfe}
	binary.LittleEndian.PutUint16(packet[2:4], 0x1234)
	binary.LittleEndian.PutUint16(packet[4:6], 0x5678)
	binary.LittleEndian.PutUint32(packet[6:10], 0x9abcdef0)

	if got := (&Server{}).NetworkNewAlias51BAD0(player, packet); got != NetworkNewAliasPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, NetworkNewAliasPacketSize51BAD0)
	}
	if got, want := player.NetData16[0xfe], (PlayerNetData{Field0: 0x1234, Field2: 0x5678, Frame4: 0x9abcdef0}); got != want {
		t.Fatalf("alias = %#v, want %#v", got, want)
	}
	if got, want := player.NetData16[0xfd], (PlayerNetData{Field0: 1, Field2: 2, Frame4: 3}); got != want {
		t.Fatalf("adjacent alias = %#v, want %#v", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= uintptr(^uint32(0)) {
		t.Fatalf("test pointer did not exercise high native half: player=%p", player)
	}
}

func TestNetworkNewAliasEscapeDoesNotOverwritePlayerUnit51BAD0(t *testing.T) {
	unit := &Object{}
	player := &Player{PlayerUnit: unit, NetCodeVal: 0x12345678}
	packet := &[NetworkNewAliasPacketSize51BAD0]byte{0: 0xa5, 1: 0xff}
	for i := 2; i < len(packet); i++ {
		packet[i] = 0xa5
	}

	if got := (&Server{}).NetworkNewAlias51BAD0(player, packet); got != NetworkNewAliasPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, NetworkNewAliasPacketSize51BAD0)
	}
	if player.PlayerUnit != unit || player.NetCodeVal != 0x12345678 {
		t.Fatalf("player tail changed: unit=%p/%p netcode=%#x", player.PlayerUnit, unit, player.NetCodeVal)
	}
}
