package opennox

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeObjectZState48EA70PacketWidth(t *testing.T) {
	data := []byte{0x5e, 0x34, 0x92, 0x68, 0xaa, 0xbb}
	state, ok := decodeObjectZState48EA70(data)
	if !ok || state.Code != 0x9234 || state.Magnitude != 0x68 {
		t.Fatalf("decoded state = %+v, ok=%t", state, ok)
	}
	for n := 0; n < 4; n++ {
		if _, ok := decodeObjectZState48EA70(data[:n]); ok {
			t.Fatalf("%d-byte object-Z packet was accepted", n)
		}
	}
}

func TestHandleObjectZNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := objectZHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		byNetCode: func(uint16) *client.Drawable {
			calls = append(calls, "lookup")
			return nil
		},
	}
	if got := handleObjectZNative48EA70([]byte{0x5e, 1, 0, 7, 0x99}, false, hooks); got != 4 {
		t.Fatalf("consumed bytes = %d, want 4", got)
	}
	if want := []string{"connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleObjectZNative48EA70PlusHighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	var calls []string
	hooks := objectZHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		byNetCode: func(code uint16) *client.Drawable {
			calls = append(calls, "lookup")
			if code != 0x9234 {
				t.Fatalf("lookup code = %#x, want 0x9234", code)
			}
			return dr
		},
	}
	data := []byte{0x5e, 0x34, 0x92, 0x68, 0xaa, 0xbb}
	before := append([]byte(nil), data...)
	if got := handleObjectZNative48EA70(data, false, hooks); got != 4 {
		t.Fatalf("consumed bytes = %d, want 4", got)
	}
	if dr.ZVal != 0x68 {
		t.Fatalf("Z value = %#x, want 0x68", dr.ZVal)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
	if want := []string{"connected", "lookup"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleObjectZNative48EA70MinusAndMissing(t *testing.T) {
	dr := new(client.Drawable)
	hooks := objectZHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			if code == 0x1234 {
				return dr
			}
			return nil
		},
	}
	if got := handleObjectZNative48EA70([]byte{0x5f, 0x34, 0x12, 0x68}, true, hooks); got != 4 {
		t.Fatalf("minus consumed bytes = %d, want 4", got)
	}
	if want := uint16(0xff98); dr.ZVal != want {
		t.Fatalf("negative Z value = %#x, want %#x", dr.ZVal, want)
	}
	dr.ZVal = 0x7777
	if got := handleObjectZNative48EA70([]byte{0x5e, 0x35, 0x12, 0x22}, false, hooks); got != 4 {
		t.Fatalf("missing consumed bytes = %d, want 4", got)
	}
	if dr.ZVal != 0x7777 {
		t.Fatalf("missing lookup changed Z value to %#x", dr.ZVal)
	}
}

func TestHandleObjectZNative48EA70RejectsShortPacket(t *testing.T) {
	if got := handleObjectZNative48EA70([]byte{0x5e, 1, 0}, false, objectZHooks48EA70{}); got != -1 {
		t.Fatalf("short packet consumed bytes = %d, want -1", got)
	}
}
