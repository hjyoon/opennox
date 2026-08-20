package server

import (
	"reflect"
	"testing"
)

type manaDrainTestUpdate4E9490 struct {
	mana uint16
}

type manaDrainTestData4E9490 struct {
	amount uint8
}

type manaDrainTestObject4E9490 struct {
	class  uint8
	update *manaDrainTestUpdate4E9490
	data   *manaDrainTestData4E9490
	timer  int16
}

type manaDrainTestState4E9490 struct {
	events []string
	frames []uint32
	fps    uint32
}

func (s *manaDrainTestState4E9490) hooks() manaDrainCollideHooks4E9490[
	*manaDrainTestObject4E9490,
	*manaDrainTestUpdate4E9490,
	*manaDrainTestData4E9490,
] {
	return manaDrainCollideHooks4E9490[
		*manaDrainTestObject4E9490,
		*manaDrainTestUpdate4E9490,
		*manaDrainTestData4E9490,
	]{
		classLow: func(obj *manaDrainTestObject4E9490) uint8 {
			s.events = append(s.events, "class")
			return obj.class
		},
		loadUpdateData: func(obj *manaDrainTestObject4E9490) *manaDrainTestUpdate4E9490 {
			s.events = append(s.events, "update")
			return obj.update
		},
		loadManaCurrent: func(update *manaDrainTestUpdate4E9490) uint16 {
			s.events = append(s.events, "mana")
			return update.mana
		},
		loadCollideData: func(obj *manaDrainTestObject4E9490) *manaDrainTestData4E9490 {
			s.events = append(s.events, "collide-data")
			return obj.data
		},
		loadAmount: func(data *manaDrainTestData4E9490) uint8 {
			s.events = append(s.events, "amount")
			return data.amount
		},
		subtractMana: func(_ *manaDrainTestObject4E9490, amount uint8) {
			s.events = append(s.events, "subtract")
			if amount != 0xa5 {
				panic("wrong amount")
			}
		},
		loadSharedTimer: func(obj *manaDrainTestObject4E9490) int16 {
			s.events = append(s.events, "timer")
			return obj.timer
		},
		loadFrame: func() uint32 {
			s.events = append(s.events, "frame")
			frame := s.frames[0]
			s.frames = s.frames[1:]
			return frame
		},
		loadFPS: func() uint32 {
			s.events = append(s.events, "fps")
			return s.fps
		},
		audio: func(*manaDrainTestObject4E9490) {
			s.events = append(s.events, "audio")
		},
		storeSharedTimer: func(obj *manaDrainTestObject4E9490, frame uint16) {
			s.events = append(s.events, "store")
			obj.timer = int16(frame)
		},
	}
}

func TestManaDrainCollide4E9490TargetGatesPrecedeSource(t *testing.T) {
	source := (*manaDrainTestObject4E9490)(nil)
	state := &manaDrainTestState4E9490{}
	manaDrainCollide4E9490(source, nil, struct{ forbidden *int }{}, state.hooks())
	if len(state.events) != 0 {
		t.Fatalf("nil-target events = %v", state.events)
	}

	state = &manaDrainTestState4E9490{}
	target := &manaDrainTestObject4E9490{class: 0xfb}
	manaDrainCollide4E9490(source, target, struct{ forbidden *int }{}, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"class"}) {
		t.Fatalf("non-player events = %v", state.events)
	}

	state = &manaDrainTestState4E9490{}
	target = &manaDrainTestObject4E9490{class: 0x04, update: &manaDrainTestUpdate4E9490{}}
	manaDrainCollide4E9490(source, target, struct{ forbidden *int }{}, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"class", "update", "mana"}) {
		t.Fatalf("zero-mana events = %v", state.events)
	}
}

func TestManaDrainCollide4E9490FullOrderAndLiveFrame(t *testing.T) {
	source := &manaDrainTestObject4E9490{
		data:  &manaDrainTestData4E9490{amount: 0xa5},
		timer: -1,
	}
	target := &manaDrainTestObject4E9490{
		class:  0x84,
		update: &manaDrainTestUpdate4E9490{mana: 1},
	}
	state := &manaDrainTestState4E9490{frames: []uint32{0, 0x12345}, fps: 0}
	manaDrainCollide4E9490(source, target, [2]float32{3, -4}, state.hooks())
	want := []string{
		"class", "update", "mana", "collide-data", "amount", "subtract",
		"timer", "frame", "fps", "audio", "frame", "store",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
	if got := uint16(source.timer); got != 0x2345 {
		t.Fatalf("stored live frame = %#x, want 0x2345", got)
	}
}

func TestManaDrainCollide4E9490StrictUnsignedThrottle(t *testing.T) {
	tests := []struct {
		name  string
		last  int16
		frame uint32
		fps   uint32
		want  bool
	}{
		{name: "exact half rejected", last: 100, frame: 130, fps: 60},
		{name: "above half accepted", last: 100, frame: 131, fps: 60, want: true},
		{name: "signed negative timer", last: -1, frame: 0, fps: 0, want: true},
		{name: "wrapped subtraction accepted", last: 1, frame: 0, fps: 0xffffffff, want: true},
		{name: "logical high-bit fps shift", last: 0, frame: 0x40000000, fps: 0x80000000},
		{name: "strictly above logical high-bit half", last: 0, frame: 0x40000001, fps: 0x80000000, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &manaDrainTestObject4E9490{
				data:  &manaDrainTestData4E9490{amount: 0xa5},
				timer: tc.last,
			}
			target := &manaDrainTestObject4E9490{
				class:  4,
				update: &manaDrainTestUpdate4E9490{mana: 1},
			}
			frames := []uint32{tc.frame}
			if tc.want {
				frames = append(frames, 7)
			}
			state := &manaDrainTestState4E9490{frames: frames, fps: tc.fps}
			manaDrainCollide4E9490(source, target, (*int)(nil), state.hooks())
			got := false
			for _, event := range state.events {
				got = got || event == "audio"
			}
			if got != tc.want {
				t.Fatalf("audio = %v, events %v", got, state.events)
			}
		})
	}
}

func TestManaDrainCollide4E9490NilPointersFaultAtOriginalLoads(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		state := &manaDrainTestState4E9490{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil update did not fault")
			}
			if !reflect.DeepEqual(state.events, []string{"class", "update", "mana"}) {
				t.Fatalf("events = %v", state.events)
			}
		}()
		manaDrainCollide4E9490(
			(*manaDrainTestObject4E9490)(nil),
			&manaDrainTestObject4E9490{class: 4},
			(*int)(nil),
			state.hooks(),
		)
	})

	t.Run("nil source", func(t *testing.T) {
		state := &manaDrainTestState4E9490{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil source did not fault")
			}
			want := []string{"class", "update", "mana", "collide-data"}
			if !reflect.DeepEqual(state.events, want) {
				t.Fatalf("events = %v, want %v", state.events, want)
			}
		}()
		manaDrainCollide4E9490(
			(*manaDrainTestObject4E9490)(nil),
			&manaDrainTestObject4E9490{class: 4, update: &manaDrainTestUpdate4E9490{mana: 1}},
			(*int)(nil),
			state.hooks(),
		)
	})
}
