package server

import (
	"reflect"
	"testing"
)

type barrelCollideTestObject4EAAA0 struct {
	last   uint32
	marker uint32
}

func defaultBarrelCollideHooks4EAAA0() barrelCollideHooks4EAAA0[*barrelCollideTestObject4EAAA0] {
	return barrelCollideHooks4EAAA0[*barrelCollideTestObject4EAAA0]{
		loadFrame: func() uint32 { return 0 },
		loadLastFrame: func(obj *barrelCollideTestObject4EAAA0) uint32 {
			return obj.last
		},
		storeFrame: func(obj *barrelCollideTestObject4EAAA0, frame uint32) {
			obj.last = frame
		},
		audio: func(uint32, *barrelCollideTestObject4EAAA0) {},
	}
}

func TestBarrelCollide4EAAA0SuccessOrderAndCachedFrame(t *testing.T) {
	source := &barrelCollideTestObject4EAAA0{last: 0x11111111, marker: 0x89abcdef}
	target := &barrelCollideTestObject4EAAA0{marker: 0x22222222}
	collision := &struct{ marker uint32 }{marker: 0x33333333}
	liveFrame := uint32(0x10203040)
	events := make([]string, 0, 4)
	hooks := defaultBarrelCollideHooks4EAAA0()
	hooks.loadFrame = func() uint32 {
		events = append(events, "frame")
		source.last = 0x1020303c
		return liveFrame
	}
	hooks.loadLastFrame = func(got *barrelCollideTestObject4EAAA0) uint32 {
		events = append(events, "last")
		if got != source {
			t.Fatalf("last source = %p", got)
		}
		liveFrame = 0
		return got.last
	}
	hooks.storeFrame = func(got *barrelCollideTestObject4EAAA0, frame uint32) {
		events = append(events, "store")
		if got != source || frame != 0x10203040 {
			t.Fatalf("store = %p/%#x", got, frame)
		}
		got.last = frame
	}
	hooks.audio = func(id uint32, got *barrelCollideTestObject4EAAA0) {
		events = append(events, "audio")
		if id != barrelCollideSound4EAAA0 || got != source || got.last != 0x10203040 {
			t.Fatalf("audio = %d/%p last %#x", id, got, got.last)
		}
	}

	barrelCollide4EAAA0(source, target, collision, hooks)
	if want := []string{"frame", "last", "store", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.last != 0x10203040 || source.marker != 0x89abcdef || target.marker != 0x22222222 || collision.marker != 0x33333333 {
		t.Fatalf("state = source %+v target %+v collision %+v", source, target, collision)
	}
}

func TestBarrelCollide4EAAA0UnsignedBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		frame     uint32
		last      uint32
		wantAudio bool
	}{
		{name: "zero", frame: 0, last: 0},
		{name: "equal threshold", frame: 3, last: 0},
		{name: "strictly above", frame: 4, last: 0, wantAudio: true},
		{name: "threshold wraps to two equal", frame: 2, last: 0xffffffff},
		{name: "threshold wraps to two above", frame: 3, last: 0xffffffff, wantAudio: true},
		{name: "threshold wraps to one", frame: 2, last: 0xfffffffe, wantAudio: true},
		{name: "high frame above wrapped zero", frame: 0xffffffff, last: 0xfffffffd, wantAudio: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &barrelCollideTestObject4EAAA0{last: tc.last, marker: 0x12345678}
			audio := 0
			stores := 0
			hooks := defaultBarrelCollideHooks4EAAA0()
			hooks.loadFrame = func() uint32 { return tc.frame }
			hooks.storeFrame = func(obj *barrelCollideTestObject4EAAA0, frame uint32) {
				stores++
				obj.last = frame
			}
			hooks.audio = func(id uint32, obj *barrelCollideTestObject4EAAA0) {
				audio++
				if id != barrelCollideSound4EAAA0 || obj != source {
					t.Fatalf("audio = %d/%p", id, obj)
				}
			}

			barrelCollide4EAAA0(source, 0x7fffffff, [2]uint32{1, 2}, hooks)
			wantCalls := 0
			wantLast := tc.last
			if tc.wantAudio {
				wantCalls = 1
				wantLast = tc.frame
			}
			if stores != wantCalls || audio != wantCalls || source.last != wantLast || source.marker != 0x12345678 {
				t.Fatalf("calls/store/audio = %d/%d, last %#x, want %d/%#x", stores, audio, source.last, wantCalls, wantLast)
			}
		})
	}
}

func TestBarrelCollide4EAAA0NilSourceLoadsFrameBeforeFault(t *testing.T) {
	events := make([]string, 0, 2)
	hooks := defaultBarrelCollideHooks4EAAA0()
	hooks.loadFrame = func() uint32 {
		events = append(events, "frame")
		return 4
	}
	hooks.loadLastFrame = func(obj *barrelCollideTestObject4EAAA0) uint32 {
		events = append(events, "last")
		return obj.last
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"frame", "last"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	barrelCollide4EAAA0[*barrelCollideTestObject4EAAA0](nil, 1, 2, hooks)
}
