package server

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestNetworkTradeStartContractOrder51BAD0(t *testing.T) {
	events := make([]string, 0, 9)
	unit := "player-unit"
	update := 17
	hooks := networkTradeStartHooks51BAD0[string, int, string]{
		gameBlocked: func() bool {
			events = append(events, "game")
			return false
		},
		loadPlayer: func(got int) string {
			events = append(events, "player")
			if got != update {
				t.Fatalf("update = %d, want %d", got, update)
			}
			return "player"
		},
		loadPlayerStatus: func(got string) uint32 {
			events = append(events, "status")
			if got != "player" {
				t.Fatalf("player = %q", got)
			}
			return 0
		},
		loadWireCode: func() uint16 {
			events = append(events, "wire")
			return 0x8123
		},
		dynamicUnitCode: func(code uint16) uint32 {
			events = append(events, "dynamic")
			if code != 0x8123 {
				t.Fatalf("wire code = %#x", code)
			}
			return 0x4567
		},
		objectFromNetCode: func(code uint32) string {
			events = append(events, "lookup")
			if code != 0x4567 {
				t.Fatalf("dynamic code = %#x", code)
			}
			return "merchant"
		},
		loadMonsterSubclassLow: func(got string) uint8 {
			events = append(events, "subclass")
			if got != "merchant" {
				t.Fatalf("merchant = %q", got)
			}
			return 0x8
		},
		startShop: func(gotUnit, gotMerchant string) {
			events = append(events, "start")
			if gotUnit != unit || gotMerchant != "merchant" {
				t.Fatalf("start = (%q, %q)", gotUnit, gotMerchant)
			}
		},
	}
	if got := networkTradeStart51BAD0(unit, update, hooks); got != NetworkTradeStartPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, NetworkTradeStartPacketSize51BAD0)
	}
	want := []string{"game", "player", "status", "wire", "dynamic", "lookup", "subclass", "start"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetworkTradeStartGatesStillConsume51BAD0(t *testing.T) {
	tests := []struct {
		name        string
		gameBlocked bool
		status      uint32
		merchant    string
		subclass    uint8
		wantEvents  []string
	}{
		{name: "game", gameBlocked: true, wantEvents: []string{"game"}},
		{name: "status", status: 1, wantEvents: []string{"game", "player", "status"}},
		{name: "missing merchant", wantEvents: []string{"game", "player", "status", "wire", "dynamic", "lookup"}},
		{name: "not shopkeeper", merchant: "monster", subclass: 0x10, wantEvents: []string{"game", "player", "status", "wire", "dynamic", "lookup", "subclass"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events := make([]string, 0, 8)
			starts := 0
			got := networkTradeStart51BAD0("unit", 1, networkTradeStartHooks51BAD0[string, int, string]{
				gameBlocked:            func() bool { events = append(events, "game"); return test.gameBlocked },
				loadPlayer:             func(int) string { events = append(events, "player"); return "player" },
				loadPlayerStatus:       func(string) uint32 { events = append(events, "status"); return test.status },
				loadWireCode:           func() uint16 { events = append(events, "wire"); return 9 },
				dynamicUnitCode:        func(uint16) uint32 { events = append(events, "dynamic"); return 9 },
				objectFromNetCode:      func(uint32) string { events = append(events, "lookup"); return test.merchant },
				loadMonsterSubclassLow: func(string) uint8 { events = append(events, "subclass"); return test.subclass },
				startShop:              func(string, string) { starts++ },
			})
			if got != NetworkTradeStartPacketSize51BAD0 || starts != 0 || !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("result = (%d, starts %d, events %v), want (%d, 0, %v)", got, starts, events, NetworkTradeStartPacketSize51BAD0, test.wantEvents)
			}
		})
	}
}

func TestNetworkTradeStartNativePointers51BAD0(t *testing.T) {
	const netCode = uint32(0x4567)
	unit := &Object{}
	merchant := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterShopkeeper),
		NetCode:     netCode,
	}
	s := &Server{}
	s.Objs.List = merchant
	update := &PlayerUpdateData{Player: &Player{}}
	packet := &[NetworkTradeStartPacketSize51BAD0]byte{0: 0xc9, 1: 0x15}
	binary.LittleEndian.PutUint16(packet[2:4], uint16(netCode))
	starts := 0
	got := s.NetworkTradeStart51BAD0(unit, update, packet, NetworkTradeStartRuntime51BAD0{
		GameBlocked: func() bool { return false },
		StartShop: func(gotUnit, gotMerchant *Object) {
			starts++
			if gotUnit != unit || gotMerchant != merchant {
				t.Fatalf("native pointer round trip = (%p, %p), want (%p, %p)", gotUnit, gotMerchant, unit, merchant)
			}
		},
	})
	if got != NetworkTradeStartPacketSize51BAD0 || starts != 1 {
		t.Fatalf("result = (%d, starts %d), want (%d, 1)", got, starts, NetworkTradeStartPacketSize51BAD0)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(merchant)) <= uintptr(^uint32(0))) {
		t.Fatalf("native addresses (%#x, %#x) did not exercise the high half", uintptr(unsafe.Pointer(unit)), uintptr(unsafe.Pointer(merchant)))
	}
}

func TestNetworkTradeExitContract51BAD0(t *testing.T) {
	events := make([]string, 0, 2)
	session := &TradeSession{}
	update := &PlayerUpdateData{Trade70: session}
	got := networkTradeExit51BAD0(update, networkTradeExitHooks51BAD0[*PlayerUpdateData, TradeSession]{
		loadSession: func(got *PlayerUpdateData) *TradeSession {
			events = append(events, "load-session")
			if got != update {
				t.Fatalf("update = %p, want %p", got, update)
			}
			return got.Trade70
		},
		exitSession: func(got *TradeSession) {
			events = append(events, "exit-session")
			if got != session {
				t.Fatalf("session = %p, want %p", got, session)
			}
		},
	})
	if got != NetworkTradeExitPacketSize51BAD0 {
		t.Fatalf("consumed = %d, want %d", got, NetworkTradeExitPacketSize51BAD0)
	}
	if want := []string{"load-session", "exit-session"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= uintptr(^uint32(0)) {
		t.Fatalf("session address %#x did not exercise the high native half", uintptr(unsafe.Pointer(session)))
	}
}

func TestNetworkTradeExitWithoutSessionStillConsumes51BAD0(t *testing.T) {
	exitCalls := 0
	got := NetworkTradeExit51BAD0(&PlayerUpdateData{}, func(*TradeSession) {
		exitCalls++
	})
	if got != NetworkTradeExitPacketSize51BAD0 || exitCalls != 0 {
		t.Fatalf("result = (%d, exits %d), want (%d, exits 0)", got, exitCalls, NetworkTradeExitPacketSize51BAD0)
	}
}
