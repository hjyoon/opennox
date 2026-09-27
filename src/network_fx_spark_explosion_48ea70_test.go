package opennox

import (
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDecodeSparkExplosionFXState48EA70PacketWidth(t *testing.T) {
	data := []byte{0x93, 0x00, 0x80, 0xff, 0x7f, 0xab, 0xcc}
	state, ok := decodeSparkExplosionFXState48EA70(data)
	if !ok || state.Pos != image.Pt(-32768, 32767) || state.Strength != 0xab {
		t.Fatalf("decoded state = %+v, ok=%t", state, ok)
	}
	for n := 0; n < 6; n++ {
		if _, ok := decodeSparkExplosionFXState48EA70(data[:n]); ok {
			t.Fatalf("%d-byte spark-explosion packet was accepted", n)
		}
	}
}

func TestResolveSparkExplosionFXTypesNative48EA70Cache(t *testing.T) {
	var types [3]int
	var calls []string
	lookup := func(name string) int {
		calls = append(calls, name)
		return len(calls) + 40
	}
	for range 2 {
		spark, medium, fire := resolveSparkExplosionFXTypesNative48EA70(&types, lookup)
		if spark != 41 || medium != 42 || fire != 43 {
			t.Fatalf("resolved types = (%d, %d, %d), want (41, 42, 43)", spark, medium, fire)
		}
	}
	if want := []string{"Spark", "MediumFireBoom", "FireBoom"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("type lookups = %v, want %v", calls, want)
	}
}

func TestResolveSparkExplosionFXTypesNative48EA70RetriesZeroSpark(t *testing.T) {
	var types [3]int
	var calls int
	lookup := func(name string) int {
		calls++
		if name == "Spark" {
			return 0
		}
		return calls
	}
	resolveSparkExplosionFXTypesNative48EA70(&types, lookup)
	resolveSparkExplosionFXTypesNative48EA70(&types, lookup)
	if calls != 6 {
		t.Fatalf("type lookups = %d, want 6 when Spark remains unresolved", calls)
	}
}

func TestSparkExplosionFXParamsAndBoomBoundary48EA70(t *testing.T) {
	for _, tc := range []struct {
		strength              byte
		count, speed, minLife int
		boomType              int
	}{
		{strength: 0, count: 10, speed: 200, minLife: 5, boomType: 12},
		{strength: 0xaa, count: 130, speed: 1800, minLife: 11, boomType: 12},
		{strength: 0xab, count: 130, speed: 1809, minLife: 11, boomType: 13},
		{strength: 0xff, count: 190, speed: 2600, minLife: 15, boomType: 13},
	} {
		count, speed, minLife := sparkExplosionFXParams48EA70(tc.strength)
		if count != tc.count || speed != tc.speed || minLife != tc.minLife {
			t.Fatalf("strength %d params = (%d, %d, %d), want (%d, %d, %d)",
				tc.strength, count, speed, minLife, tc.count, tc.speed, tc.minLife)
		}
		if got := sparkExplosionBoomType48EA70(tc.strength, 12, 13); got != tc.boomType {
			t.Fatalf("strength %d boom type = %d, want %d", tc.strength, got, tc.boomType)
		}
	}
}

func TestHandleSparkExplosionFXNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := sparkExplosionFXHooks48EA70{
		types: func() (int, int, int) {
			calls = append(calls, "types")
			return 41, 42, 43
		},
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		spawn: func(int, image.Point) *client.Drawable {
			calls = append(calls, "spawn")
			return nil
		},
	}
	if got := handleSparkExplosionFXNative48EA70([]byte{0x93, 1, 0, 2, 0, 255}, hooks); got != 6 {
		t.Fatalf("consumed bytes = %d, want 6", got)
	}
	if want := []string{"types", "connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleSparkExplosionFXNative48EA70HighAddress(t *testing.T) {
	const (
		sparkType      = 41
		mediumBoomType = 42
		fireBoomType   = 43
	)
	pos := image.Pt(-32768, 32767)
	drawables := make([]*client.Drawable, 11)
	for i := range drawables {
		drawables[i] = new(client.Drawable)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(drawables[i])) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low address: %p", drawables[i])
		}
	}

	spawnIndex := 0
	randomIndex := 0
	frameCalls := 0
	var activated []*client.Drawable
	hooks := sparkExplosionFXHooks48EA70{
		connected: func() bool { return true },
		types:     func() (int, int, int) { return sparkType, mediumBoomType, fireBoomType },
		random: func(min, max int) int {
			want := [][2]int{{0, 255}, {1, 200}, {5, 96}, {5, 15}, {0, 8}}[randomIndex%5]
			if min != want[0] || max != want[1] {
				t.Fatalf("random call %d range = (%d, %d), want (%d, %d)", randomIndex, min, max, want[0], want[1])
			}
			randomIndex++
			return min
		},
		frame: func() uint32 {
			frameCalls++
			return 100
		},
		spawn: func(typ int, at image.Point) *client.Drawable {
			wantType := sparkType
			if spawnIndex == 10 {
				wantType = mediumBoomType
			}
			if typ != wantType || at != pos {
				t.Fatalf("spawn %d = type %d at %v, want type %d at %v", spawnIndex, typ, at, wantType, pos)
			}
			dr := drawables[spawnIndex]
			dr.PosVec = at
			spawnIndex++
			return dr
		},
		activate: func(dr *client.Drawable) {
			activated = append(activated, dr)
		},
	}

	data := []byte{0x93, 0x00, 0x80, 0xff, 0x7f, 0x00, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleSparkExplosionFXNative48EA70(data, hooks); got != 6 {
		t.Fatalf("consumed bytes = %d, want 6", got)
	}
	if spawnIndex != 11 || randomIndex != 50 || frameCalls != 20 {
		t.Fatalf("spawn/random/frame calls = %d/%d/%d, want 11/50/20", spawnIndex, randomIndex, frameCalls)
	}
	if !reflect.DeepEqual(activated, drawables) {
		t.Fatalf("activated drawables differ from native-pointer spawn results")
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
	for i, dr := range drawables[:10] {
		effect := dr.UnionEffect()
		if effect.Field_108 != uint32(pos.X)<<12 || effect.Field_109 != uint32(pos.Y)<<12 ||
			effect.Field_110 != 1 || effect.Field_111 != 100 || effect.Field_112 != 105 ||
			dr.Field_74_4 != 0 || dr.ZVal != 5 || dr.ZVal2 != 0 || dr.VelZ != 0 {
			t.Fatalf("spark %d = %+v, effect = %+v", i, dr, effect)
		}
	}
}

func TestHandleSparkExplosionFXNative48EA70NilSpawnsAndInvalidPacket(t *testing.T) {
	spawnCalls := 0
	randomCalls := 0
	frameCalls := 0
	activateCalls := 0
	hooks := sparkExplosionFXHooks48EA70{
		connected: func() bool { return true },
		types:     func() (int, int, int) { return 41, 42, 43 },
		random: func(int, int) int {
			randomCalls++
			return 0
		},
		frame: func() uint32 {
			frameCalls++
			return 0
		},
		spawn: func(int, image.Point) *client.Drawable {
			spawnCalls++
			return nil
		},
		activate: func(*client.Drawable) { activateCalls++ },
	}
	if got := handleSparkExplosionFXNative48EA70([]byte{0x93, 1, 0, 2, 0, 0}, hooks); got != 6 {
		t.Fatalf("nil-spawn consumed bytes = %d, want 6", got)
	}
	if spawnCalls != 11 || randomCalls != 0 || frameCalls != 0 || activateCalls != 0 {
		t.Fatalf("spawn/random/frame/activate calls = %d/%d/%d/%d, want 11/0/0/0",
			spawnCalls, randomCalls, frameCalls, activateCalls)
	}
	for n := 0; n < 6; n++ {
		if got := handleSparkExplosionFXNative48EA70([]byte{0x93, 1, 0, 2, 0, 0}[:n], sparkExplosionFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed bytes = %d, want -1", n, got)
		}
	}
}
