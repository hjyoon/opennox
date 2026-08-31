package server

import (
	"encoding/binary"
	"runtime"
	"testing"
	"unsafe"
)

func TestNetworkTryCollideNativePointersAndLayout51BAD0(t *testing.T) {
	wantCollide := uintptr(696)
	wantUpdatePlayer := uintptr(276)
	wantUpdateTrade := uintptr(280)
	wantUpdateDialog := uintptr(284)
	wantPlayerStatus := uintptr(3680)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantCollide = 768
		wantUpdatePlayer = 336
		wantUpdateTrade = 344
		wantUpdateDialog = 352
		wantPlayerStatus = 4976
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.Collide", unsafe.Offsetof(Object{}.Collide), wantCollide},
		{"Object.Collide width", unsafe.Sizeof(Object{}.Collide), unsafe.Sizeof(uintptr(0))},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"PlayerUpdateData.Trade70", unsafe.Offsetof(PlayerUpdateData{}.Trade70), wantUpdateTrade},
		{"PlayerUpdateData.DialogWith", unsafe.Offsetof(PlayerUpdateData{}.DialogWith), wantUpdateDialog},
		{"PlayerUpdateData.Player width", unsafe.Sizeof(PlayerUpdateData{}.Player), unsafe.Sizeof(uintptr(0))},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerStatus},
		{"Player.Field3680 width", unsafe.Sizeof(Player{}.Field3680), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}

	const (
		extent  = uint32(0x123)
		netCode = uint32(0x4567)
	)
	callbackToken := uint64(0x1122334455667788)
	callback := unsafe.Pointer(&callbackToken)
	target := &Object{Extent: extent, NetCode: netCode, Collide: callback}
	unit := &Object{}
	player := &Player{}
	update := &PlayerUpdateData{Player: player}
	s := &Server{}
	s.Objs.List = target
	packet := &[NetworkTryCollidePacketSize51BAD0]byte{0: 0x7b}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(extent)|0x8000)

	debugCalls := 0
	collideCalls := 0
	got := s.NetworkTryCollide51BAD0(unit, update, packet, NetworkTryCollideRuntime51BAD0{
		NetDebug: func() bool { return true },
		TestHighBit: func(code uint16) {
			debugCalls++
			if code != uint16(extent)|0x8000 {
				t.Fatalf("debug code = %#x", code)
			}
		},
		CallCollide: func(gotCallback unsafe.Pointer, gotTarget, gotUnit *Object) {
			collideCalls++
			if gotCallback != callback || gotTarget != target || gotUnit != unit {
				t.Fatalf("native callback = (%p, %p, %p), want (%p, %p, %p)", gotCallback, gotTarget, gotUnit, callback, target, unit)
			}
		},
	})
	if got != 3 || debugCalls != 1 || collideCalls != 1 {
		t.Fatalf("result = (%d, debug %d, collide %d), want (3,1,1)", got, debugCalls, collideCalls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(target)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(update)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(player)) <= uintptr(^uint32(0)) ||
			uintptr(callback) <= uintptr(^uint32(0))) {
		t.Fatalf("test pointers did not exercise high native halves: target=%p unit=%p update=%p player=%p callback=%p", target, unit, update, player, callback)
	}
	runtime.KeepAlive(callbackToken)
}

func TestNetworkTryCollideNativeGates51BAD0(t *testing.T) {
	callbackToken := byte(1)
	target := &Object{NetCode: 9, Collide: unsafe.Pointer(&callbackToken)}
	packet := &[NetworkTryCollidePacketSize51BAD0]byte{0x7b, 9, 0}
	for _, tc := range []struct {
		name   string
		status uint32
		trade  bool
		dialog bool
	}{
		{name: "player status", status: 1},
		{name: "trade", trade: true},
		{name: "dialog", dialog: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			update := &PlayerUpdateData{Player: &Player{Field3680: tc.status}}
			if tc.trade {
				update.Trade70 = &TradeSession{}
			}
			if tc.dialog {
				update.DialogWith = &Object{}
			}
			s := &Server{}
			s.Objs.List = target
			calls := 0
			if got := s.NetworkTryCollide51BAD0(&Object{}, update, packet, NetworkTryCollideRuntime51BAD0{
				NetDebug:    func() bool { return false },
				TestHighBit: func(uint16) { t.Fatal("unexpected debug callback") },
				CallCollide: func(unsafe.Pointer, *Object, *Object) { calls++ },
			}); got != 3 || calls != 0 {
				t.Fatalf("result = (%d, calls %d), want (3,0)", got, calls)
			}
		})
	}
	runtime.KeepAlive(callbackToken)
}
