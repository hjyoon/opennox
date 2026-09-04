package opennox

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestManaBombCancelFloatToInt48EA70(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0, 0},
		{1.5, 2},
		{2.5, 2},
		{-1.5, -2},
		{-2.5, -2},
		{2147483520, 2147483520},
		{-2147483648, math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.value), func(t *testing.T) {
			if got := manaBombCancelFloatToInt48EA70(tc.value); got != tc.want {
				t.Fatalf("float conversion = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestResolveManaBombCancelTypeNative48EA70(t *testing.T) {
	var cache int
	var calls []string
	lookup := func(name string) int {
		calls = append(calls, name)
		return 83
	}
	for range 2 {
		if got := resolveManaBombCancelTypeNative48EA70(&cache, lookup); got != 83 {
			t.Fatalf("resolved type = %d, want 83", got)
		}
	}
	if want := []string{"CyanSpark"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("type lookups = %v, want %v", calls, want)
	}
}

func TestResolveManaBombCancelTypeNative48EA70RetriesZero(t *testing.T) {
	var cache, calls int
	lookup := func(string) int {
		calls++
		return 0
	}
	resolveManaBombCancelTypeNative48EA70(&cache, lookup)
	resolveManaBombCancelTypeNative48EA70(&cache, lookup)
	if calls != 2 {
		t.Fatalf("type lookups = %d, want 2 when CyanSpark remains unresolved", calls)
	}
}

func TestHandleManaBombCancelNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := manaBombCancelHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		radius: func() float32 {
			calls = append(calls, "radius")
			return 64
		},
	}
	if got := handleManaBombCancelNative48EA70([]byte{0xa3, 1, 0, 2, 0}, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if want := []string{"connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandleManaBombCancelNative48EA70HighAddress(t *testing.T) {
	const (
		radius = int32(64)
		typeID = 83
	)
	center := image.Pt(-32700, 32600)
	drawables := make([]*client.Drawable, manaBombCancelSparkCount48EA70)
	for i := range drawables {
		drawables[i] = new(client.Drawable)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(drawables[i])) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low address: %p", drawables[i])
		}
	}

	var calls []string
	spawned := 0
	activated := 0
	radiusRandom := 0
	angleRandom := 0
	expiryRandom := 0
	velocityRandom := 0
	frameCalls := 0
	hooks := manaBombCancelHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		radius: func() float32 {
			calls = append(calls, "radius")
			return float32(radius)
		},
		typeID: func() int {
			calls = append(calls, "type")
			return typeID
		},
		random: func(min, max int) int {
			switch {
			case min == 0 && max == int(radius):
				v := radiusRandom % (int(radius) + 1)
				radiusRandom++
				return v
			case min == 0 && max == 255:
				v := angleRandom % 256
				angleRandom++
				return v
			case min == 30 && max == 40:
				expiryRandom++
				return 35
			case min == 4 && max == 10:
				velocityRandom++
				return 7
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		frame: func() uint32 {
			v := uint32(1000 + frameCalls)
			frameCalls++
			return v
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != typeID {
				t.Fatalf("spawn type = %d, want %d", typ, typeID)
			}
			i := spawned
			spawned++
			distance := int32(uint32(radius)>>2) + int32(i%(int(radius)+1))
			if distance > radius {
				distance = radius
			}
			dir := sincosTable16[uint8(i)]
			want := image.Pt(
				int(int32(center.X)+distance*int32(dir.X)/16),
				int(int32(center.Y)+distance*int32(dir.Y)/16),
			)
			if pos != want {
				t.Fatalf("spark %d position = %v, want %v", i, pos, want)
			}
			dr := drawables[i]
			dr.PosVec = pos
			dr.ZVal = 35
			dr.VelZ = 2
			dr.Field_74_4 = 0xff
			return dr
		},
		activate: func(dr *client.Drawable) {
			i := activated
			if dr != drawables[i] {
				t.Fatalf("activated spark %d = %p, want native pointer %p", i, dr, drawables[i])
			}
			effect := dr.UnionEffect()
			if effect.Field_108 != uint32(dr.PosVec.X)<<12 || effect.Field_109 != uint32(dr.PosVec.Y)<<12 {
				t.Fatalf("spark %d fixed position = (%#x, %#x), want (%#x, %#x)", i,
					effect.Field_108, effect.Field_109, uint32(dr.PosVec.X)<<12, uint32(dr.PosVec.Y)<<12)
			}
			wantExpire := uint32(1000 + 2*i + 35)
			wantStart := uint32(1001 + 2*i)
			if effect.Field_110 != 0 || effect.Field_111 != wantStart || effect.Field_112 != wantExpire {
				t.Fatalf("spark %d timing = (%d, %d, %d), want (0, %d, %d)", i,
					effect.Field_110, effect.Field_111, effect.Field_112, wantStart, wantExpire)
			}
			if dr.Field_74_4 != 0 || dr.ZVal != 0 || dr.VelZ != 7 {
				t.Fatalf("spark %d state = angle:%d Z:%d VelZ:%d, want 0/0/7", i, dr.Field_74_4, dr.ZVal, dr.VelZ)
			}
			activated++
		},
	}
	data := []byte{0xa3, 0x44, 0x80, 0x58, 0x7f, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleManaBombCancelNative48EA70(data, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if want := []string{"connected", "radius", "type"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("initial callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
	if spawned != manaBombCancelSparkCount48EA70 || activated != manaBombCancelSparkCount48EA70 {
		t.Fatalf("spark calls = spawn:%d activate:%d, want %d each", spawned, activated, manaBombCancelSparkCount48EA70)
	}
	if radiusRandom != manaBombCancelSparkCount48EA70 || angleRandom != manaBombCancelSparkCount48EA70 ||
		expiryRandom != manaBombCancelSparkCount48EA70 || velocityRandom != manaBombCancelSparkCount48EA70 ||
		frameCalls != 2*manaBombCancelSparkCount48EA70 {
		t.Fatalf("effect calls = radius:%d angle:%d expiry:%d velocity:%d frame:%d", radiusRandom, angleRandom, expiryRandom, velocityRandom, frameCalls)
	}
}

func TestHandleManaBombCancelNative48EA70NilAndShort(t *testing.T) {
	spawnCalls := 0
	randomCalls := 0
	hooks := manaBombCancelHooks48EA70{
		connected: func() bool { return true },
		radius:    func() float32 { return 64 },
		typeID:    func() int { return 83 },
		random: func(min, max int) int {
			randomCalls++
			return min
		},
		spawn: func(int, image.Point) *client.Drawable {
			spawnCalls++
			return nil
		},
	}
	if got := handleManaBombCancelNative48EA70([]byte{0xa3, 1, 0, 2, 0}, hooks); got != 5 {
		t.Fatalf("nil-spawn consumed bytes = %d, want 5", got)
	}
	if spawnCalls != manaBombCancelSparkCount48EA70 || randomCalls != 2*manaBombCancelSparkCount48EA70 {
		t.Fatalf("nil-spawn calls = spawn:%d random:%d, want %d/%d", spawnCalls, randomCalls,
			manaBombCancelSparkCount48EA70, 2*manaBombCancelSparkCount48EA70)
	}
	for n := 0; n < 5; n++ {
		if got := handleManaBombCancelNative48EA70([]byte{0xa3, 1, 0, 2, 0}[:n], manaBombCancelHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed bytes = %d, want -1", n, got)
		}
	}
}
