package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestDispatchClientPlayerRespawn417680PreservesNativePointer(t *testing.T) {
	player := new(server.Player)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(player.C()) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", player)
	}
	const mask = uint8(0xa5)
	called := false
	dispatchClientPlayerRespawn417680(player, mask, func(got *server.Player, gotMask uint8) {
		called = true
		if got != player || gotMask != mask {
			t.Fatalf("dispatch = (%p, %#x), want (%p, %#x)", got, gotMask, player, mask)
		}
	})
	if !called {
		t.Fatal("respawn callback was not called")
	}
}

func TestDispatchClientPlayerRespawn417680ForwardsNil(t *testing.T) {
	called := false
	dispatchClientPlayerRespawn417680(nil, 7, func(got *server.Player, mask uint8) {
		called = true
		if got != nil || mask != 7 {
			t.Fatalf("dispatch = (%p, %d), want (nil, 7)", got, mask)
		}
	})
	if !called {
		t.Fatal("respawn callback was not called")
	}
}
