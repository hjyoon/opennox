package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectGoldNativeKeepsPointerWidthAndUpdatesProtection(t *testing.T) {
	player := &server.Player{GoldVal: 250, ProtPlayerGold: 0x89abcdef}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("expected native unit pointer above 4 GiB")
	}

	type call struct {
		kind  string
		token uint32
		value int32
	}
	var calls []call
	protect := func(token uint32, delta int32) {
		calls = append(calls, call{kind: "protect", token: token, value: delta})
	}
	reset := func(token uint32, value int32) {
		calls = append(calls, call{kind: "reset", token: token, value: value})
	}

	if got := objectGetGoldNative4FA6D0(unit); got != 250 {
		t.Fatalf("initial gold = %d, want 250", got)
	}
	objectSetGoldNative4FA620(unit, -200, protect, reset)
	if player.GoldVal != 50 {
		t.Fatalf("gold after -200 = %d, want 50", player.GoldVal)
	}
	objectSetGoldNative4FA620(unit, 25, protect, reset)
	if player.GoldVal != 75 {
		t.Fatalf("gold after +25 = %d, want 75", player.GoldVal)
	}
	want := []call{
		{kind: "protect", token: 0x89abcdef, value: -200},
		{kind: "protect", token: 0x89abcdef, value: 25},
	}
	if len(calls) != len(want) {
		t.Fatalf("protection calls = %#v, want %#v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("protection call %d = %#v, want %#v", i, calls[i], want[i])
		}
	}
}

func TestObjectGoldNativeSaturatesAndResetsProtection(t *testing.T) {
	player := &server.Player{GoldVal: 50, ProtPlayerGold: 0x12345678}
	update := &server.PlayerUpdateData{Player: player}
	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	protectCalls := 0
	var resetToken uint32
	var resetValue int32 = -1

	objectSetGoldNative4FA620(
		unit,
		-200,
		func(uint32, int32) { protectCalls++ },
		func(token uint32, value int32) {
			resetToken = token
			resetValue = value
		},
	)

	if player.GoldVal != 0 {
		t.Fatalf("gold after saturating subtraction = %d, want 0", player.GoldVal)
	}
	if protectCalls != 0 {
		t.Fatalf("incremental protection calls = %d, want 0", protectCalls)
	}
	if resetToken != 0x12345678 || resetValue != 0 {
		t.Fatalf("protection reset = (%#x, %d), want (%#x, 0)", resetToken, resetValue, uint32(0x12345678))
	}
}

func TestObjectGoldNativeIgnoresNonPlayersAndBrokenLinks(t *testing.T) {
	called := false
	hook := func(uint32, int32) { called = true }

	for _, unit := range []*server.Object{
		nil,
		{},
		{ObjClass: object.ClassPlayer},
		{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&server.PlayerUpdateData{})},
	} {
		objectSetGoldNative4FA620(unit, 10, hook, hook)
		if got := objectGetGoldNative4FA6D0(unit); got != 0 {
			t.Fatalf("gold for invalid player %p = %d, want 0", unit, got)
		}
	}
	if called {
		t.Fatal("protection hook called for invalid player")
	}
}
