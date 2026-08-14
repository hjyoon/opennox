package server

import (
	"reflect"
	"testing"
)

type audioEventCollideTestObject4EAAD0 struct {
	class uint8
	last  uint32
	data  *AudioEventCollideData
	guard uint32
}

func defaultAudioEventCollideHooks4EAAD0() audioEventCollideHooks4EAAD0[*audioEventCollideTestObject4EAAD0, *AudioEventCollideData] {
	return audioEventCollideHooks4EAAD0[*audioEventCollideTestObject4EAAD0, *AudioEventCollideData]{
		classLow: func(obj *audioEventCollideTestObject4EAAD0) uint8 {
			return obj.class
		},
		loadFrame: func() uint32 { return 0 },
		loadLastFrame: func(obj *audioEventCollideTestObject4EAAD0) uint32 {
			return obj.last
		},
		storeFrame: func(obj *audioEventCollideTestObject4EAAD0, frame uint32) {
			obj.last = frame
		},
		loadCollideData: func(obj *audioEventCollideTestObject4EAAD0) *AudioEventCollideData {
			return obj.data
		},
		loadSound: func(data *AudioEventCollideData) uint32 {
			return data.Sound
		},
		audio: func(uint32, *audioEventCollideTestObject4EAAD0) {},
	}
}

func TestAudioEventCollide4EAAD0TargetGuardsBeforeSource(t *testing.T) {
	tests := []struct {
		name       string
		target     *audioEventCollideTestObject4EAAD0
		wantEvents []string
	}{
		{name: "nil target"},
		{name: "non player", target: &audioEventCollideTestObject4EAAD0{class: 0x80}, wantEvents: []string{"class"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := defaultAudioEventCollideHooks4EAAD0()
			hooks.classLow = func(obj *audioEventCollideTestObject4EAAD0) uint8 {
				events = append(events, "class")
				return obj.class
			}
			hooks.loadFrame = func() uint32 {
				t.Fatal("source path reached")
				return 0
			}
			audioEventCollide4EAAD0[*audioEventCollideTestObject4EAAD0](nil, tc.target, 123, hooks)
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestAudioEventCollide4EAAD0SuccessOrderAndCachedValues(t *testing.T) {
	data := &AudioEventCollideData{Sound: 0x11223344}
	source := &audioEventCollideTestObject4EAAD0{last: 0x10203020, data: data, guard: 0x89abcdef}
	target := &audioEventCollideTestObject4EAAD0{class: 0x84, guard: 0x76543210}
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	liveFrame := uint32(0x10203040)
	events := make([]string, 0, 7)
	hooks := defaultAudioEventCollideHooks4EAAD0()
	hooks.classLow = func(obj *audioEventCollideTestObject4EAAD0) uint8 {
		events = append(events, "class")
		if obj != target {
			t.Fatalf("class object = %p", obj)
		}
		return obj.class
	}
	hooks.loadFrame = func() uint32 {
		events = append(events, "frame")
		return liveFrame
	}
	hooks.loadLastFrame = func(obj *audioEventCollideTestObject4EAAD0) uint32 {
		events = append(events, "last")
		if obj != source {
			t.Fatalf("last object = %p", obj)
		}
		liveFrame = 0
		return obj.last
	}
	hooks.storeFrame = func(obj *audioEventCollideTestObject4EAAD0, frame uint32) {
		events = append(events, "store")
		if obj != source || frame != 0x10203040 {
			t.Fatalf("store = %p/%#x", obj, frame)
		}
		obj.last = frame
	}
	hooks.loadCollideData = func(obj *audioEventCollideTestObject4EAAD0) *AudioEventCollideData {
		events = append(events, "data")
		if obj != source || obj.last != 0x10203040 {
			t.Fatalf("data object = %p, last %#x", obj, obj.last)
		}
		return obj.data
	}
	hooks.loadSound = func(got *AudioEventCollideData) uint32 {
		events = append(events, "sound")
		if got != data {
			t.Fatalf("data = %p", got)
		}
		data.Sound = 0
		return 0x11223344
	}
	hooks.audio = func(id uint32, obj *audioEventCollideTestObject4EAAD0) {
		events = append(events, "audio")
		if id != 0x11223344 || obj != source || obj.last != 0x10203040 {
			t.Fatalf("audio = %#x/%p, last %#x", id, obj, obj.last)
		}
	}

	audioEventCollide4EAAD0(source, target, collision, hooks)
	want := []string{"class", "frame", "last", "store", "data", "sound", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.last != 0x10203040 || source.guard != 0x89abcdef || target.guard != 0x76543210 || collision.guard != 0x31415926 {
		t.Fatalf("state = source %+v target %+v collision %+v", source, target, collision)
	}
}

func TestAudioEventCollide4EAAD0UnsignedBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		frame     uint32
		last      uint32
		wantAudio bool
	}{
		{name: "below", frame: 29, last: 0},
		{name: "equal threshold", frame: 30, last: 0},
		{name: "strictly above", frame: 31, last: 0, wantAudio: true},
		{name: "wrapped equal", frame: 29, last: 0xffffffff},
		{name: "wrapped above", frame: 30, last: 0xffffffff, wantAudio: true},
		{name: "high frame above wrapped threshold", frame: 0xffffffff, last: 0xfffffff0, wantAudio: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &audioEventCollideTestObject4EAAD0{
				last: tc.last,
				data: &AudioEventCollideData{Sound: 0},
			}
			target := &audioEventCollideTestObject4EAAD0{class: audioEventCollidePlayerClass4EAAD0}
			stores := 0
			audio := 0
			hooks := defaultAudioEventCollideHooks4EAAD0()
			hooks.loadFrame = func() uint32 { return tc.frame }
			hooks.storeFrame = func(obj *audioEventCollideTestObject4EAAD0, frame uint32) {
				stores++
				obj.last = frame
			}
			hooks.audio = func(id uint32, obj *audioEventCollideTestObject4EAAD0) {
				audio++
				if id != 0 || obj != source {
					t.Fatalf("audio = %d/%p", id, obj)
				}
			}
			audioEventCollide4EAAD0(source, target, 0, hooks)
			wantCalls := 0
			wantLast := tc.last
			if tc.wantAudio {
				wantCalls = 1
				wantLast = tc.frame
			}
			if stores != wantCalls || audio != wantCalls || source.last != wantLast {
				t.Fatalf("stores/audio/last = %d/%d/%#x, want %d/%d/%#x", stores, audio, source.last, wantCalls, wantCalls, wantLast)
			}
		})
	}
}

func TestAudioEventCollide4EAAD0NilSourceLoadsFrameBeforeFault(t *testing.T) {
	target := &audioEventCollideTestObject4EAAD0{class: audioEventCollidePlayerClass4EAAD0}
	events := make([]string, 0, 3)
	hooks := defaultAudioEventCollideHooks4EAAD0()
	hooks.classLow = func(obj *audioEventCollideTestObject4EAAD0) uint8 {
		events = append(events, "class")
		return obj.class
	}
	hooks.loadFrame = func() uint32 {
		events = append(events, "frame")
		return 31
	}
	hooks.loadLastFrame = func(obj *audioEventCollideTestObject4EAAD0) uint32 {
		events = append(events, "last")
		return obj.last
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"class", "frame", "last"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	audioEventCollide4EAAD0[*audioEventCollideTestObject4EAAD0](nil, target, 0, hooks)
}
