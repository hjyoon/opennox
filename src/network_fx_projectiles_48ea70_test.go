package opennox

import (
	"encoding/binary"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/sound"
)

func TestDecodeProjectileFXPackets48EA70(t *testing.T) {
	t.Run("sentry ray", func(t *testing.T) {
		data := []byte{0x95, 0xff, 0xff, 0x00, 0x80, 0x34, 0x12, 0x78, 0x56, 0xcc}
		state, ok := decodeSentryRayFXState48EA70(data)
		if !ok || state.From != image.Pt(65535, 32768) || state.To != image.Pt(0x1234, 0x5678) {
			t.Fatalf("decoded state = %+v, ok=%t", state, ok)
		}
		for n := 0; n < 9; n++ {
			if _, ok := decodeSentryRayFXState48EA70(data[:n]); ok {
				t.Fatalf("%d-byte sentry-ray packet was accepted", n)
			}
		}
	})

	t.Run("ricochet", func(t *testing.T) {
		data := []byte{0x96, 0x00, 0x80, 0xff, 0x7f, 0xcc}
		pos, ok := decodeRicochetFXPosition48EA70(data)
		if !ok || pos != image.Pt(-32768, 32767) {
			t.Fatalf("decoded position = %v, ok=%t", pos, ok)
		}
		for n := 0; n < 5; n++ {
			if _, ok := decodeRicochetFXPosition48EA70(data[:n]); ok {
				t.Fatalf("%d-byte ricochet packet was accepted", n)
			}
		}
	})

	t.Run("green bolt", func(t *testing.T) {
		data := []byte{0x98, 0xff, 0xff, 0x00, 0x80, 0x34, 0x12, 0x78, 0x56, 0xcd, 0xab, 0xcc}
		state, ok := decodeGreenBoltFXState48EA70(data)
		if !ok || state.From != image.Pt(65535, 32768) || state.To != image.Pt(0x1234, 0x5678) || state.Duration != 0xabcd {
			t.Fatalf("decoded state = %+v, ok=%t", state, ok)
		}
		for n := 0; n < 11; n++ {
			if _, ok := decodeGreenBoltFXState48EA70(data[:n]); ok {
				t.Fatalf("%d-byte green-bolt packet was accepted", n)
			}
		}
	})
}

func TestHandleSentryRayFXNative48EA70HighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	var queued [2]image.Point
	var randomCalls [][2]int
	var sounds []struct {
		id     sound.ID
		volume int
	}
	frameCalls := 0
	typeCalls := 0
	spawnCalls := 0
	activateCalls := 0
	hooks := sentryRayFXHooks48EA70{
		connected: func() bool { return true },
		queueRay: func(from, to image.Point) {
			queued = [2]image.Point{from, to}
		},
		random: func(min, max int) int {
			randomCalls = append(randomCalls, [2]int{min, max})
			return min
		},
		playerPos: func() (image.Point, bool) { return image.Pt(130, 140), true },
		playSound: func(id sound.ID, volume int) {
			sounds = append(sounds, struct {
				id     sound.ID
				volume int
			}{id, volume})
		},
		paused: func() bool { return false },
		typeID: func() int {
			typeCalls++
			return 77
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawnCalls++
			if typ != 77 || pos != image.Pt(128, 237) {
				t.Fatalf("spawn = type %d at %v, want type 77 at (128,237)", typ, pos)
			}
			dr.PosVec = pos
			return dr
		},
		frame: func() uint32 {
			frameCalls++
			return 100
		},
		activate: func(got *client.Drawable) {
			activateCalls++
			if got != dr {
				t.Fatalf("activated drawable = %p, want %p", got, dr)
			}
		},
	}
	data := []byte{0x95, 100, 0, 200, 0, 130, 0, 240, 0, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleSentryRayFXNative48EA70(data, hooks); got != 9 {
		t.Fatalf("consumed bytes = %d, want 9", got)
	}
	if queued != [2]image.Point{image.Pt(100, 200), image.Pt(130, 240)} {
		t.Fatalf("queued ray = %v", queued)
	}
	if len(sounds) != 1 || sounds[0].id != sound.SoundSentryRayHitWall || sounds[0].volume != 83 {
		t.Fatalf("sounds = %+v, want SentryRayHitWall at volume 83", sounds)
	}
	wantRandom := [][2]int{{0, 100}, {0, 255}, {1, 1500}, {5, 20}, {-4, 4}}
	if !reflect.DeepEqual(randomCalls, wantRandom) {
		t.Fatalf("random calls = %v, want %v", randomCalls, wantRandom)
	}
	if typeCalls != 1 || spawnCalls != 1 || frameCalls != 2 || activateCalls != 1 {
		t.Fatalf("type/spawn/frame/activate calls = %d/%d/%d/%d, want 1/1/2/1",
			typeCalls, spawnCalls, frameCalls, activateCalls)
	}
	effect := dr.UnionEffect()
	if effect.Field_108 != uint32(128)<<12 || effect.Field_109 != uint32(237)<<12 ||
		effect.Field_110 != 1 || effect.Field_111 != 100 || effect.Field_112 != 105 ||
		dr.Field_74_4 != 0 || dr.ZVal != 22 || dr.VelZ != -4 {
		t.Fatalf("drawable = %+v, effect = %+v", dr, effect)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleSentryRayFXNative48EA70Gates(t *testing.T) {
	data := []byte{0x95, 1, 0, 2, 0, 3, 0, 4, 0}
	for n := 0; n < 9; n++ {
		if got := handleSentryRayFXNative48EA70(data[:n], sentryRayFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed %d bytes, want -1", n, got)
		}
	}
	called := false
	if got := handleSentryRayFXNative48EA70(data, sentryRayFXHooks48EA70{
		connected: func() bool { return false },
		queueRay:  func(image.Point, image.Point) { called = true },
	}); got != 9 || called {
		t.Fatalf("disconnected result/callback = %d/%t, want 9/false", got, called)
	}

	var calls []string
	if got := handleSentryRayFXNative48EA70(data, sentryRayFXHooks48EA70{
		connected: func() bool { return true },
		queueRay:  func(image.Point, image.Point) { calls = append(calls, "queue") },
		random: func(min, max int) int {
			calls = append(calls, "sound roll")
			return 25
		},
		playerPos: func() (image.Point, bool) {
			calls = append(calls, "player")
			return image.Point{}, false
		},
		paused: func() bool {
			calls = append(calls, "paused")
			return true
		},
		typeID: func() int {
			calls = append(calls, "type")
			return 0
		},
	}); got != 9 {
		t.Fatalf("paused result = %d, want 9", got)
	}
	if want := []string{"queue", "sound roll", "paused"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("paused callback order = %v, want %v", calls, want)
	}
}

func TestHandleRicochetFXNative48EA70HighAddress(t *testing.T) {
	pos := image.Pt(-32768, 32767)
	drawables := make([]*client.Drawable, 5)
	for i := range drawables {
		drawables[i] = &client.Drawable{PosVec: pos}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(drawables[i])) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low address: %p", drawables[i])
		}
	}
	spawnIndex := 0
	frameCalls := 0
	var randomCalls [][2]int
	var activated []*client.Drawable
	data := []byte{0x96, 0x00, 0x80, 0xff, 0x7f, 0xcc}
	before := append([]byte(nil), data...)
	if got := handleRicochetFXNative48EA70(data, ricochetFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return 88 },
		spawn: func(typ int, at image.Point) *client.Drawable {
			if typ != 88 || at != pos {
				t.Fatalf("spawn %d = %d/%v, want 88/%v", spawnIndex, typ, at, pos)
			}
			dr := drawables[spawnIndex]
			spawnIndex++
			return dr
		},
		random: func(min, max int) int {
			randomCalls = append(randomCalls, [2]int{min, max})
			return min
		},
		frame: func() uint32 {
			frameCalls++
			return 200
		},
		activate: func(dr *client.Drawable) { activated = append(activated, dr) },
	}); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if spawnIndex != 5 || frameCalls != 10 || !reflect.DeepEqual(activated, drawables) {
		t.Fatalf("spawn/frame/activated = %d/%d/%d, want 5/10/5", spawnIndex, frameCalls, len(activated))
	}
	wantRandom := [][2]int{{0, 255}, {1333, 4000}, {5, 20}, {-5, 5}}
	if len(randomCalls) != 20 || !reflect.DeepEqual(randomCalls[:4], wantRandom) {
		t.Fatalf("random calls = %v, want 5 repetitions of %v", randomCalls, wantRandom)
	}
	for i, dr := range drawables {
		effect := dr.UnionEffect()
		if effect.Field_108 != uint32(pos.X)<<12 || effect.Field_109 != uint32(pos.Y)<<12 ||
			effect.Field_110 != 1333 || effect.Field_111 != 200 || effect.Field_112 != 205 ||
			dr.Field_74_4 != 0 || dr.ZVal != 20 || dr.VelZ != -5 {
			t.Fatalf("drawable %d = %+v, effect = %+v", i, dr, effect)
		}
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleRicochetFXNative48EA70Gates(t *testing.T) {
	data := []byte{0x96, 1, 0, 2, 0}
	if got := handleRicochetFXNative48EA70(data[:4], ricochetFXHooks48EA70{}); got != -1 {
		t.Fatalf("short packet result = %d, want -1", got)
	}
	typeCalls := 0
	if got := handleRicochetFXNative48EA70(data, ricochetFXHooks48EA70{
		connected: func() bool { return false },
		typeID: func() int {
			typeCalls++
			return 0
		},
	}); got != 5 || typeCalls != 0 {
		t.Fatalf("disconnected result/type calls = %d/%d, want 5/0", got, typeCalls)
	}

	spawnCalls := 0
	randomCalls := 0
	if got := handleRicochetFXNative48EA70(data, ricochetFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func() int { return 1 },
		spawn: func(int, image.Point) *client.Drawable {
			spawnCalls++
			return nil
		},
		random: func(int, int) int {
			randomCalls++
			return 0
		},
	}); got != 5 || spawnCalls != 5 || randomCalls != 0 {
		t.Fatalf("nil-spawn result/spawn/random = %d/%d/%d, want 5/5/0", got, spawnCalls, randomCalls)
	}
}

func TestHandleGreenBoltFXNative48EA70HighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	sentinel := new(client.Drawable)
	dr.NextPtr = sentinel
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 20)
	for i := range payload {
		payload[i] = 0xa5
	}
	data := []byte{0x98, 100, 0, 200, 0, 130, 0, 241, 0, 0x34, 0x12, 0xcc}
	before := append([]byte(nil), data...)
	typeCalls := 0
	spawnCalls := 0
	if got := handleGreenBoltFXNative48EA70(data, greenBoltFXHooks48EA70{
		connected: func() bool { return true },
		typeID: func() int {
			typeCalls++
			return 99
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawnCalls++
			if typ != 99 || pos != image.Pt(115, 220) {
				t.Fatalf("spawn = %d/%v, want 99/(115,220)", typ, pos)
			}
			return dr
		},
	}); got != 11 {
		t.Fatalf("consumed bytes = %d, want 11", got)
	}
	want := []byte{0, 0x34, 0x12, 0, 0, 100, 0, 200, 0, 130, 0, 241, 0}
	if !reflect.DeepEqual(payload[:13], want) {
		t.Fatalf("payload = %x, want %x", payload[:13], want)
	}
	for i, got := range payload[13:] {
		if got != 0xa5 {
			t.Fatalf("payload tail byte %d = %#x, want %#x", i+13, got, 0xa5)
		}
	}
	if dr.NextPtr != sentinel {
		t.Fatalf("native pointer field was corrupted: got %p, want %p", dr.NextPtr, sentinel)
	}
	if typeCalls != 1 || spawnCalls != 1 || !reflect.DeepEqual(data, before) {
		t.Fatalf("type/spawn/packet = %d/%d/%x, want 1/1/%x", typeCalls, spawnCalls, data, before)
	}
}

func TestHandleGreenBoltFXNative48EA70GatesAndMidpoint(t *testing.T) {
	data := make([]byte, 11)
	data[0] = 0x98
	binary.LittleEndian.PutUint16(data[1:3], 10)
	binary.LittleEndian.PutUint16(data[3:5], 11)
	binary.LittleEndian.PutUint16(data[5:7], 5)
	binary.LittleEndian.PutUint16(data[7:9], 4)
	state, ok := decodeGreenBoltFXState48EA70(data)
	if !ok || greenBoltFXMidpoint48EA70(state) != image.Pt(8, 8) {
		t.Fatalf("descending midpoint = %v, ok=%t, want (8,8)/true", greenBoltFXMidpoint48EA70(state), ok)
	}
	if got := handleGreenBoltFXNative48EA70(data[:10], greenBoltFXHooks48EA70{}); got != -1 {
		t.Fatalf("short packet result = %d, want -1", got)
	}
	var calls []string
	if got := handleGreenBoltFXNative48EA70(data, greenBoltFXHooks48EA70{
		typeID: func() int {
			calls = append(calls, "type")
			return 7
		},
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		spawn: func(int, image.Point) *client.Drawable {
			calls = append(calls, "spawn")
			return nil
		},
	}); got != 11 {
		t.Fatalf("disconnected result = %d, want 11", got)
	}
	if want := []string{"type", "connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}
