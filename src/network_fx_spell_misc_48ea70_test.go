package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeMiscSpellFXPackets48EA70(t *testing.T) {
	t.Run("Delta-Z", func(t *testing.T) {
		data := []byte{0x9f, 0x34, 0x92, 0xff, 0x80, 0x7f, 0xcc}
		state, ok := decodeDeltaZFXState48EA70(data)
		if !ok || state.Code != 0x9234 || state.Height != 0xff || state.Velocity != -128 || state.Target != 0x7f {
			t.Fatalf("decoded state = %+v, ok=%t", state, ok)
		}
		for n := 0; n < 6; n++ {
			if _, ok := decodeDeltaZFXState48EA70(data[:n]); ok {
				t.Fatalf("%d-byte Delta-Z packet was accepted", n)
			}
		}
	})

	t.Run("arrow trap", func(t *testing.T) {
		data := []byte{0xa1, 0x00, 0x80, 0xff, 0x7f, 2, 0xcc}
		state, ok := decodeArrowTrapFXState48EA70(data)
		if !ok || state.Position != image.Pt(-32768, 32767) || state.Variant != 2 {
			t.Fatalf("decoded state = %+v, ok=%t", state, ok)
		}
		for n := 0; n < 6; n++ {
			if _, ok := decodeArrowTrapFXState48EA70(data[:n]); ok {
				t.Fatalf("%d-byte arrow-trap packet was accepted", n)
			}
		}
	})

	t.Run("Vampirism", func(t *testing.T) {
		data := []byte{0xa2, 0xff, 0xff, 0x00, 0x80, 0x34, 0x12, 0x78, 0x56, 0xcd, 0xab, 0xcc}
		state, ok := decodeVampirismFXState48EA70(data)
		if !ok || state.From != image.Pt(65535, 32768) || state.To != image.Pt(0x1234, 0x5678) || state.Amount != 0xabcd {
			t.Fatalf("decoded state = %+v, ok=%t", state, ok)
		}
		for n := 0; n < 11; n++ {
			if _, ok := decodeVampirismFXState48EA70(data[:n]); ok {
				t.Fatalf("%d-byte Vampirism packet was accepted", n)
			}
		}
	})
}

func TestVampirismFXOrbCount48EA70(t *testing.T) {
	for _, tc := range []struct {
		amount uint16
		want   int
	}{
		{0, 1}, {3, 1}, {4, 2}, {27, 7}, {28, 8}, {31, 8}, {32, 8}, {math.MaxUint16, 8},
	} {
		if got := vampirismFXOrbCount48EA70(tc.amount); got != tc.want {
			t.Errorf("amount %d produced %d orbs, want %d", tc.amount, got, tc.want)
		}
	}
}

func TestHandleDeltaZFXNative48EA70HighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	var calls []string
	hooks := deltaZFXHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		byCode: func(code uint16) *client.Drawable {
			calls = append(calls, "lookup")
			if code != 0x9234 {
				t.Fatalf("lookup code = %#x, want 0x9234", code)
			}
			return dr
		},
		frame: func() uint32 {
			calls = append(calls, "frame")
			return 0x10203040
		},
	}
	data := []byte{0x9f, 0x34, 0x92, 0xfe, 0x80, 0x7d, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleDeltaZFXNative48EA70(data, hooks); got != 6 {
		t.Fatalf("consumed bytes = %d, want 6", got)
	}
	effect := dr.UnionEffect()
	if effect.Field_108 != 0x10203040 || math.Float32frombits(effect.Field_109) != 254 ||
		math.Float32frombits(effect.Field_110) != -128 || math.Float32frombits(effect.Field_111) != 125 {
		t.Fatalf("Delta-Z state = frame:%#x height:%g velocity:%g target:%g",
			effect.Field_108, math.Float32frombits(effect.Field_109),
			math.Float32frombits(effect.Field_110), math.Float32frombits(effect.Field_111))
	}
	if want := []string{"connected", "lookup", "frame"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleDeltaZFXNative48EA70Gates(t *testing.T) {
	data := []byte{0x9f, 1, 0, 2, 3, 4}
	for n := 0; n < 6; n++ {
		if got := handleDeltaZFXNative48EA70(data[:n], deltaZFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed %d bytes, want -1", n, got)
		}
	}
	called := false
	if got := handleDeltaZFXNative48EA70(data, deltaZFXHooks48EA70{
		connected: func() bool { return false },
		byCode: func(uint16) *client.Drawable {
			called = true
			return nil
		},
	}); got != 6 || called {
		t.Fatalf("disconnected result/callback = %d/%t, want 6/false", got, called)
	}
	if got := handleDeltaZFXNative48EA70(data, deltaZFXHooks48EA70{
		connected: func() bool { return true },
		byCode:    func(uint16) *client.Drawable { return nil },
		frame: func() uint32 {
			called = true
			return 0
		},
	}); got != 6 || called {
		t.Fatalf("missing drawable result/frame = %d/%t, want 6/false", got, called)
	}
}

func TestHandleArrowTrapFXNative48EA70Variants(t *testing.T) {
	for _, tc := range []struct {
		name    string
		variant byte
		wantTyp int
		wantPos image.Point
	}{
		{"first", 1, 41, image.Pt(-85, 200)},
		{"second", 2, 42, image.Pt(-103, 200)},
		{"fallback", 0xff, 42, image.Pt(-103, 200)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dr := new(client.Drawable)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
				t.Skipf("allocator returned a low address: %p", dr)
			}
			var calls []string
			var packet [6]byte
			packet[0] = 0xa1
			binary.LittleEndian.PutUint16(packet[1:3], 0xff9c)
			binary.LittleEndian.PutUint16(packet[3:5], 200)
			packet[5] = tc.variant
			if got := handleArrowTrapFXNative48EA70(packet[:], arrowTrapFXHooks48EA70{
				connected: func() bool {
					calls = append(calls, "connected")
					return true
				},
				typeIDs: func() [2]int {
					calls = append(calls, "types")
					return [2]int{41, 42}
				},
				spawn: func(typ int, pos image.Point) *client.Drawable {
					calls = append(calls, "spawn")
					if typ != tc.wantTyp || pos != tc.wantPos {
						t.Fatalf("spawn = type %d at %v, want type %d at %v", typ, pos, tc.wantTyp, tc.wantPos)
					}
					return dr
				},
				activate: func(got *client.Drawable) {
					calls = append(calls, "activate")
					if got != dr {
						t.Fatalf("activated drawable = %p, want %p", got, dr)
					}
				},
			}); got != 6 {
				t.Fatalf("consumed bytes = %d, want 6", got)
			}
			if want := []string{"connected", "types", "spawn", "activate"}; !reflect.DeepEqual(calls, want) {
				t.Fatalf("callback order = %v, want %v", calls, want)
			}
		})
	}
}

func TestHandleArrowTrapFXNative48EA70Gates(t *testing.T) {
	data := []byte{0xa1, 1, 0, 2, 0, 1}
	for n := 0; n < 6; n++ {
		if got := handleArrowTrapFXNative48EA70(data[:n], arrowTrapFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed %d bytes, want -1", n, got)
		}
	}
	called := false
	if got := handleArrowTrapFXNative48EA70(data, arrowTrapFXHooks48EA70{
		connected: func() bool { return false },
		typeIDs: func() [2]int {
			called = true
			return [2]int{}
		},
	}); got != 6 || called {
		t.Fatalf("disconnected result/callback = %d/%t, want 6/false", got, called)
	}
	if got := handleArrowTrapFXNative48EA70(data, arrowTrapFXHooks48EA70{
		connected: func() bool { return true },
		typeIDs:   func() [2]int { return [2]int{1, 2} },
		spawn:     func(int, image.Point) *client.Drawable { return nil },
		activate: func(*client.Drawable) {
			called = true
		},
	}); got != 6 || called {
		t.Fatalf("failed spawn result/activation = %d/%t, want 6/false", got, called)
	}
}

func TestHandleVampirismFXNative48EA70HighAddresses(t *testing.T) {
	drawables := []*client.Drawable{new(client.Drawable), new(client.Drawable), new(client.Drawable)}
	for _, dr := range drawables {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low address: %p", dr)
		}
	}
	var randomCalls [][2]int
	spawned := 0
	var activated []*client.Drawable
	random := func(min, max int) int {
		randomCalls = append(randomCalls, [2]int{min, max})
		switch [2]int{min, max} {
		case [2]int{6, 12}:
			return 9
		case [2]int{-20, 20}:
			if len(randomCalls)%4 == 2 {
				return -2
			}
			return 3
		case [2]int{3, 10}:
			return 5
		default:
			t.Fatalf("unexpected random range [%d,%d]", min, max)
			return 0
		}
	}
	data := []byte{0xa2, 100, 0, 200, 0, 44, 1, 144, 1, 8, 0, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleVampirismFXNative48EA70(data, vampirismFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return 77 },
		random:    random,
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 77 || pos != image.Pt(303, 398) {
				t.Fatalf("spawn %d = type %d at %v, want type 77 at (303,398)", spawned, typ, pos)
			}
			dr := drawables[spawned]
			spawned++
			return dr
		},
		activate: func(dr *client.Drawable) { activated = append(activated, dr) },
	}); got != 11 {
		t.Fatalf("consumed bytes = %d, want 11", got)
	}
	if spawned != 3 || !reflect.DeepEqual(activated, drawables) {
		t.Fatalf("spawned/activated = %d/%p, want 3/%p", spawned, activated, drawables)
	}
	wantRandom := make([][2]int, 0, 12)
	for range 3 {
		wantRandom = append(wantRandom, [2]int{6, 12}, [2]int{-20, 20}, [2]int{-20, 20}, [2]int{3, 10})
	}
	if !reflect.DeepEqual(randomCalls, wantRandom) {
		t.Fatalf("random calls = %v, want %v", randomCalls, wantRandom)
	}
	for i, dr := range drawables {
		payload := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 15)
		if binary.LittleEndian.Uint16(payload[0:2]) != 100 || binary.LittleEndian.Uint16(payload[2:4]) != 200 ||
			payload[11] != 9 || payload[12] != 5 || payload[13] != 0 || payload[14] != 0 {
			t.Fatalf("orb %d payload = %x", i, payload)
		}
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleVampirismFXNative48EA70GatesAndFailedSpawn(t *testing.T) {
	data := []byte{0xa2, 1, 0, 2, 0, 3, 0, 4, 0, 0, 0}
	for n := 0; n < 11; n++ {
		if got := handleVampirismFXNative48EA70(data[:n], vampirismFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed %d bytes, want -1", n, got)
		}
	}
	called := false
	if got := handleVampirismFXNative48EA70(data, vampirismFXHooks48EA70{
		connected: func() bool { return false },
		typeID: func() int {
			called = true
			return 0
		},
	}); got != 11 || called {
		t.Fatalf("disconnected result/callback = %d/%t, want 11/false", got, called)
	}
	var randomCalls [][2]int
	if got := handleVampirismFXNative48EA70(data, vampirismFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return 1 },
		random: func(min, max int) int {
			randomCalls = append(randomCalls, [2]int{min, max})
			return min
		},
		spawn: func(int, image.Point) *client.Drawable { return nil },
		activate: func(*client.Drawable) {
			called = true
		},
	}); got != 11 || called {
		t.Fatalf("failed spawn result/activation = %d/%t, want 11/false", got, called)
	}
	if want := [][2]int{{6, 12}, {-20, 20}, {-20, 20}}; !reflect.DeepEqual(randomCalls, want) {
		t.Fatalf("failed spawn random calls = %v, want %v", randomCalls, want)
	}
}
