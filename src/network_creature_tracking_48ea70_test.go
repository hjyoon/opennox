package opennox

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func requireCreatureTrackingHighAddress48EA70(t *testing.T, dr *client.Drawable) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
}

func TestDecodeCreatureTrackingPackets48EA70(t *testing.T) {
	acquire := []byte{0x6c, 0x34, 0x92, 0x78, 0x80}
	state, ok := decodeCreatureAcquireState48EA70(acquire)
	if !ok || state.RawCode != 0x9234 || state.Code != 0x1234 || state.TypeID != 0x78 || !state.Silent {
		t.Fatalf("acquire state = %+v, ok=%t", state, ok)
	}

	monitor := []byte{0xdb, 0x34, 0x92, 0xcd, 0xab}
	ref, ok := decodeCreatureReference48EA70(monitor, creatureMonitorPacketSize48EA70)
	if !ok || ref.RawCode != 0x9234 || ref.Code != 0x1234 || ref.TypeID != 0xabcd {
		t.Fatalf("monitor state = %+v, ok=%t", ref, ok)
	}

	interesting := []byte{0xd2, 0x34, 0x12, 0x78, 0x56, 2, 3}
	interestingState, ok := decodeInterestingIDState48EA70(interesting)
	if !ok || interestingState.RawCode != 0x1234 || interestingState.Code != 0x1234 ||
		interestingState.TypeID != 0x5678 || interestingState.Action != 2 || interestingState.Flags != 3 {
		t.Fatalf("interesting-ID state = %+v, ok=%t", interestingState, ok)
	}
}

func TestHandleCreatureTrackingNative48EA70PacketGates(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		handler func([]byte, creatureTrackingHooks48EA70) int
	}{
		{"acquire", creatureAcquirePacketSize48EA70, handleCreatureAcquireNative48EA70},
		{"lose", creatureLosePacketSize48EA70, handleCreatureLoseNative48EA70},
		{"monitor", creatureMonitorPacketSize48EA70, handleCreatureMonitorNative48EA70},
		{"unmonitor", creatureUnmonitorPacketSize48EA70, handleCreatureUnmonitorNative48EA70},
		{"interesting-ID", interestingIDPacketSize48EA70, handleInterestingIDNative48EA70},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.size)
			for n := 0; n < tc.size; n++ {
				if got := tc.handler(data[:n], creatureTrackingHooks48EA70{}); got != -1 {
					t.Fatalf("%d-byte packet consumed %d bytes, want -1", n, got)
				}
			}
			called := false
			hooks := creatureTrackingHooks48EA70{
				connected: func() bool { return false },
				byStatic: func(uint16) *client.Drawable {
					called = true
					return nil
				},
			}
			if got := tc.handler(data, hooks); got != tc.size || called {
				t.Fatalf("disconnected result/callback = %d/%t, want %d/false", got, called, tc.size)
			}
		})
	}
}

func TestHandleCreatureAcquireNative48EA70PreservesHighAddress(t *testing.T) {
	dr := new(client.Drawable)
	requireCreatureTrackingHighAddress48EA70(t, dr)
	packet := []byte{0x6c, 0x34, 0x92, 0x78, 0x80}
	before := append([]byte(nil), packet...)
	var calls []string
	hooks := creatureTrackingHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		summonAcquire: func(code, typeID uint16, silent bool) {
			calls = append(calls, "summon")
			if code != 0x9234 || typeID != 0x78 || !silent {
				t.Fatalf("summon = code:%#x type:%#x silent:%t", code, typeID, silent)
			}
		},
		byStatic: func(code uint16) *client.Drawable {
			calls = append(calls, "static")
			if code != 0x1234 {
				t.Fatalf("static code = %#x, want 0x1234", code)
			}
			return dr
		},
		byDynamic: func(uint16) *client.Drawable {
			t.Fatal("dynamic namespace was queried")
			return nil
		},
		create: func(uint16, uint16) *client.Drawable {
			t.Fatal("existing drawable was recreated")
			return nil
		},
		minimapAdd: func(got *client.Drawable, flags byte) {
			calls = append(calls, "minimap")
			if got != dr || flags != 1 {
				t.Fatalf("minimap add = %p/%d, want %p/1", got, flags, dr)
			}
		},
		monitorAdd: func(code uint16) {
			calls = append(calls, "monitor")
			if code != 0x9234 {
				t.Fatalf("monitor code = %#x, want raw 0x9234", code)
			}
		},
	}
	if got := handleCreatureAcquireNative48EA70(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "summon", "static", "minimap", "monitor"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(packet, before) {
		t.Fatalf("packet mutated: got %x, want %x", packet, before)
	}
}

func TestHandleCreatureLoseNative48EA70UsesDynamicNamespace(t *testing.T) {
	dr := new(client.Drawable)
	requireCreatureTrackingHighAddress48EA70(t, dr)
	packet := []byte{0x6d, 0x34, 0x92}
	before := append([]byte(nil), packet...)
	var calls []string
	hooks := creatureTrackingHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		summonLose: func(code uint16, silent bool) {
			calls = append(calls, "summon")
			if code != 0x1234 || !silent {
				t.Fatalf("summon lose = code:%#x silent:%t", code, silent)
			}
		},
		monitorRemove: func(code uint16) {
			calls = append(calls, "monitor")
			if code != 0x1234 {
				t.Fatalf("monitor removal code = %#x, want 0x1234", code)
			}
		},
		byDynamic: func(code uint16) *client.Drawable {
			calls = append(calls, "dynamic")
			if code != 0x1234 {
				t.Fatalf("dynamic code = %#x, want 0x1234", code)
			}
			return dr
		},
		byStatic: func(uint16) *client.Drawable {
			t.Fatal("lose packet queried the static namespace")
			return nil
		},
		minimapRemove: func(got *client.Drawable, flags byte) {
			calls = append(calls, "minimap")
			if got != dr || flags != 1 {
				t.Fatalf("minimap removal = %p/%d, want %p/1", got, flags, dr)
			}
		},
	}
	if got := handleCreatureLoseNative48EA70(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "summon", "monitor", "dynamic", "minimap"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(packet, before) {
		t.Fatalf("packet mutated: got %x, want %x", packet, before)
	}
}

func TestHandleCreatureMonitorNative48EA70CreatesAtNativeWidth(t *testing.T) {
	dr := new(client.Drawable)
	requireCreatureTrackingHighAddress48EA70(t, dr)
	packet := []byte{0xdb, 0x34, 0x92, 0x78, 0x56}
	var calls []string
	hooks := creatureTrackingHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		byStatic: func(code uint16) *client.Drawable {
			calls = append(calls, "static")
			if code != 0x1234 {
				t.Fatalf("static code = %#x, want 0x1234", code)
			}
			return nil
		},
		byDynamic: func(uint16) *client.Drawable {
			t.Fatal("dynamic namespace was queried")
			return nil
		},
		create: func(typeID, code uint16) *client.Drawable {
			calls = append(calls, "create")
			if typeID != 0x5678 || code != 0x1234 {
				t.Fatalf("create = type:%#x code:%#x", typeID, code)
			}
			return dr
		},
		minimapAdd: func(got *client.Drawable, flags byte) {
			calls = append(calls, "minimap")
			if got != dr || flags != 1 {
				t.Fatalf("minimap add = %p/%d, want %p/1", got, flags, dr)
			}
		},
		monitorAdd: func(code uint16) {
			calls = append(calls, "monitor")
			if code != 0x1234 {
				t.Fatalf("monitor code = %#x, want stripped 0x1234", code)
			}
		},
	}
	if got := handleCreatureMonitorNative48EA70(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "static", "create", "minimap", "monitor"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleCreatureUnmonitorNative48EA70PreservesHighAddress(t *testing.T) {
	dr := new(client.Drawable)
	requireCreatureTrackingHighAddress48EA70(t, dr)
	packet := []byte{0xdc, 0x34, 0x92}
	var calls []string
	hooks := creatureTrackingHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		monitorRemove: func(code uint16) {
			calls = append(calls, "monitor")
			if code != 0x9234 {
				t.Fatalf("monitor removal code = %#x, want raw 0x9234", code)
			}
		},
		byStatic: func(code uint16) *client.Drawable {
			calls = append(calls, "static")
			if code != 0x1234 {
				t.Fatalf("static code = %#x, want 0x1234", code)
			}
			return dr
		},
		byDynamic: func(uint16) *client.Drawable {
			t.Fatal("dynamic namespace was queried")
			return nil
		},
		minimapRemove: func(got *client.Drawable, flags byte) {
			calls = append(calls, "minimap")
			if got != dr || flags != 1 {
				t.Fatalf("minimap removal = %p/%d, want %p/1", got, flags, dr)
			}
		},
	}
	if got := handleCreatureUnmonitorNative48EA70(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "monitor", "static", "minimap"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleInterestingIDNative48EA70AddAndRemove(t *testing.T) {
	dr := new(client.Drawable)
	requireCreatureTrackingHighAddress48EA70(t, dr)
	var packet [interestingIDPacketSize48EA70]byte
	packet[0] = 0xd2
	binary.LittleEndian.PutUint16(packet[1:3], 0x1234)
	binary.LittleEndian.PutUint16(packet[3:5], 0x5678)
	packet[5] = 1
	packet[6] = 3
	var calls []string
	hooks := creatureTrackingHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		byDynamic: func(code uint16) *client.Drawable {
			calls = append(calls, "dynamic")
			if code != 0x1234 {
				t.Fatalf("dynamic code = %#x, want 0x1234", code)
			}
			return nil
		},
		create: func(typeID, code uint16) *client.Drawable {
			calls = append(calls, "create")
			if typeID != 0x5678 || code != 0x1234 {
				t.Fatalf("create = type:%#x code:%#x", typeID, code)
			}
			return dr
		},
		minimapAdd: func(got *client.Drawable, flags byte) {
			calls = append(calls, "add")
			if got != dr || flags != 3 {
				t.Fatalf("minimap add = %p/%d, want %p/3", got, flags, dr)
			}
		},
	}
	if got := handleInterestingIDNative48EA70(packet[:], hooks); got != len(packet) {
		t.Fatalf("add consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "dynamic", "create", "add"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("add callback order = %v, want %v", calls, want)
	}

	packet[5] = 2
	packet[6] = 2
	calls = nil
	hooks.byDynamic = func(uint16) *client.Drawable {
		calls = append(calls, "dynamic")
		return dr
	}
	hooks.minimapRemove = func(got *client.Drawable, flags byte) {
		calls = append(calls, "remove")
		if got != dr || flags != 2 {
			t.Fatalf("minimap removal = %p/%d, want %p/2", got, flags, dr)
		}
	}
	if got := handleInterestingIDNative48EA70(packet[:], hooks); got != len(packet) {
		t.Fatalf("remove consumed bytes = %d, want %d", got, len(packet))
	}
	if want := []string{"connected", "dynamic", "remove"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("remove callback order = %v, want %v", calls, want)
	}
}
