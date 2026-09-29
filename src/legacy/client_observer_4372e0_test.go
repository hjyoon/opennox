package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestClientIsObserver4372E0UsesNativePlayerLayout(t *testing.T) {
	old := Get_dword_8531A0_2576()
	t.Cleanup(func() {
		Set_dword_8531A0_2576(old)
	})

	Set_dword_8531A0_2576(nil)
	if got := Nox_xxx_clientIsObserver_4372E0(); got != 0 {
		t.Fatalf("nil player observer state = %d, want 0", got)
	}

	player := new(server.Player)
	var pin runtime.Pinner
	pin.Pin(player)
	t.Cleanup(pin.Unpin)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("player pointer = %p, want native address above 4 GiB", player)
	}
	Set_dword_8531A0_2576(player)

	tests := []struct {
		name   string
		active byte
		flags  uint32
		want   int
	}{
		{name: "inactive with observer bit", flags: 1},
		{name: "active without observer bits", active: 1},
		{name: "active with low observer bit", active: 1, flags: 1, want: 1},
		{name: "active with high observer bit", active: 1, flags: 2, want: 1},
		{name: "active with both observer bits", active: 1, flags: 3, want: 1},
		{name: "active with unrelated bit", active: 1, flags: 4},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			player.Active = test.active
			player.Field3680 = test.flags
			if got := Nox_xxx_clientIsObserver_4372E0(); got != test.want {
				t.Fatalf("observer state = %d, want %d (active=%d flags=%#x)", got, test.want, test.active, test.flags)
			}
		})
	}
	runtime.KeepAlive(player)
}
