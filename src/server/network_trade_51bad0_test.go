package server

import (
	"reflect"
	"testing"
	"unsafe"
)

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
