package opennox

import (
	"fmt"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeSmokeBlastState48EA70PacketWidth(t *testing.T) {
	data := []byte{0x8a, 0x00, 0x80, 0xff, 0x7f, 0xaa, 0xbb}
	state, ok := decodeSmokeBlastState48EA70(data)
	if !ok || state.Pos != image.Pt(-32768, 32767) {
		t.Fatalf("decoded state = %+v, ok=%t", state, ok)
	}
	for n := 0; n < 5; n++ {
		if _, ok := decodeSmokeBlastState48EA70(data[:n]); ok {
			t.Fatalf("%d-byte smoke-blast packet was accepted", n)
		}
	}
}

func TestResolveSmokeBlastTypesNative48EA70Cache(t *testing.T) {
	var smoke, puff int
	var calls []string
	lookup := func(name string) int {
		calls = append(calls, name)
		if name == "Smoke" {
			return 41
		}
		return 42
	}
	if gotSmoke, gotPuff := resolveSmokeBlastTypesNative48EA70(&smoke, &puff, lookup); gotSmoke != 41 || gotPuff != 42 {
		t.Fatalf("resolved types = (%d, %d), want (41, 42)", gotSmoke, gotPuff)
	}
	if gotSmoke, gotPuff := resolveSmokeBlastTypesNative48EA70(&smoke, &puff, lookup); gotSmoke != 41 || gotPuff != 42 {
		t.Fatalf("cached types = (%d, %d), want (41, 42)", gotSmoke, gotPuff)
	}
	if want := []string{"Smoke", "Puff"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("type lookups = %v, want %v", calls, want)
	}
}

func TestResolveSmokeBlastTypesNative48EA70RetriesZeroSmoke(t *testing.T) {
	var smoke, puff, calls int
	lookup := func(name string) int {
		calls++
		if name == "Puff" {
			return 42
		}
		return 0
	}
	resolveSmokeBlastTypesNative48EA70(&smoke, &puff, lookup)
	resolveSmokeBlastTypesNative48EA70(&smoke, &puff, lookup)
	if calls != 4 {
		t.Fatalf("type lookups = %d, want 4 when Smoke remains unresolved", calls)
	}
}

func TestHandleSmokeBlastNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := smokeBlastHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		types: func() (int, int) {
			calls = append(calls, "types")
			return 0, 0
		},
	}
	if got := handleSmokeBlastNative48EA70([]byte{0x8a, 1, 0, 2, 0, 0x99}, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if want := []string{"connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleSmokeBlastNative48EA70HighAddress(t *testing.T) {
	const (
		smokeType = 41
		puffType  = 42
	)
	base := image.Pt(-123, 321)
	randomValues := []int{1, -2, 5, 3, -4, 6, 5, -6, 7, 7, -8, 8, 9, -10, 9, 11, -12, 10}
	wantPos := []image.Point{
		base,
		image.Pt(-125, 322),
		image.Pt(-127, 324),
		image.Pt(-129, 326),
		image.Pt(-131, 328),
		image.Pt(-133, 330),
		image.Pt(-135, 332),
	}
	drawables := make([]*client.Drawable, len(wantPos))
	indices := make(map[*client.Drawable]int, len(wantPos))
	for i := range drawables {
		drawables[i] = new(client.Drawable)
		indices[drawables[i]] = i
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(drawables[i])) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low address: %p", drawables[i])
		}
	}

	var calls []string
	spawnIndex := 0
	randomIndex := 0
	var activated []*client.Drawable
	hooks := smokeBlastHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		types: func() (int, int) {
			calls = append(calls, "types")
			return smokeType, puffType
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			calls = append(calls, fmt.Sprintf("spawn:%d:%d:%d", typ, pos.X, pos.Y))
			wantType := puffType
			if spawnIndex == 0 {
				wantType = smokeType
			}
			if typ != wantType || pos != wantPos[spawnIndex] {
				t.Fatalf("spawn %d = type %d at %v, want type %d at %v", spawnIndex, typ, pos, wantType, wantPos[spawnIndex])
			}
			dr := drawables[spawnIndex]
			spawnIndex++
			return dr
		},
		random: func(min, max int) int {
			calls = append(calls, fmt.Sprintf("random:%d:%d", min, max))
			if randomIndex >= len(randomValues) {
				t.Fatalf("unexpected random call %d", randomIndex)
			}
			if randomIndex%3 == 2 {
				if min != 5 || max != 25 {
					t.Fatalf("Z random range = (%d, %d), want (5, 25)", min, max)
				}
			} else if min != -15 || max != 15 {
				t.Fatalf("position random range = (%d, %d), want (-15, 15)", min, max)
			}
			v := randomValues[randomIndex]
			randomIndex++
			return v
		},
		activate: func(dr *client.Drawable) {
			calls = append(calls, fmt.Sprintf("activate:%d", indices[dr]))
			activated = append(activated, dr)
		},
	}

	data := []byte{0x8a, 0x85, 0xff, 0x41, 0x01, 0xaa, 0xbb}
	before := append([]byte(nil), data...)
	if got := handleSmokeBlastNative48EA70(data, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if spawnIndex != len(drawables) || randomIndex != len(randomValues) {
		t.Fatalf("spawn/random calls = %d/%d, want %d/%d", spawnIndex, randomIndex, len(drawables), len(randomValues))
	}
	if !reflect.DeepEqual(activated, drawables) {
		t.Fatalf("activated drawables = %p, want %p", activated, drawables)
	}
	for i, dr := range drawables {
		wantZ := uint16(20)
		if i != 0 {
			wantZ = uint16(randomValues[(i-1)*3+2])
		}
		if dr.ZVal != wantZ {
			t.Fatalf("drawable %d Z = %d, want %d", i, dr.ZVal, wantZ)
		}
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}

	wantCalls := []string{"connected", "types", "spawn:41:-123:321", "activate:0"}
	for i := 1; i < len(drawables); i++ {
		wantCalls = append(wantCalls,
			"random:-15:15",
			"random:-15:15",
			fmt.Sprintf("spawn:42:%d:%d", wantPos[i].X, wantPos[i].Y),
			"random:5:25",
			fmt.Sprintf("activate:%d", i),
		)
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("callback order = %v, want %v", calls, wantCalls)
	}
}

func TestHandleSmokeBlastNative48EA70NilPuffsSkipZRandom(t *testing.T) {
	smoke := new(client.Drawable)
	spawnCalls := 0
	randomCalls := 0
	activateCalls := 0
	hooks := smokeBlastHooks48EA70{
		connected: func() bool { return true },
		types:     func() (int, int) { return 41, 42 },
		spawn: func(int, image.Point) *client.Drawable {
			spawnCalls++
			if spawnCalls == 1 {
				return smoke
			}
			return nil
		},
		random: func(min, max int) int {
			randomCalls++
			if min != -15 || max != 15 {
				t.Fatalf("random range = (%d, %d), want (-15, 15)", min, max)
			}
			return 0
		},
		activate: func(dr *client.Drawable) {
			activateCalls++
			if dr != smoke || dr.ZVal != 20 {
				t.Fatalf("activated drawable = %p with Z %d, want smoke %p with Z 20", dr, dr.ZVal, smoke)
			}
		},
	}
	if got := handleSmokeBlastNative48EA70([]byte{0x8a, 1, 0, 2, 0}, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if spawnCalls != 7 || randomCalls != 12 || activateCalls != 1 {
		t.Fatalf("spawn/random/activate calls = %d/%d/%d, want 7/12/1", spawnCalls, randomCalls, activateCalls)
	}
}

func TestHandleSmokeBlastNative48EA70RejectsShortPacket(t *testing.T) {
	data := []byte{0x8a, 1, 0, 2}
	for n := 0; n <= len(data); n++ {
		if got := handleSmokeBlastNative48EA70(data[:n], smokeBlastHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed bytes = %d, want -1", n, got)
		}
	}
}
