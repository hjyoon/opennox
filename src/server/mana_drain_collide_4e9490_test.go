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

type manaSubTestUpdate4EEBF0 struct {
	current uint16
	prev    uint16
	player  *manaSubTestPlayer4EEBF0
}

type manaSubTestPlayer4EEBF0 struct {
	token uint32
}

type manaSubTestObject4EEBF0 struct {
	class  uint8
	update *manaSubTestUpdate4EEBF0
}

type manaSubTestState4EEBF0 struct {
	events  []string
	godMode bool
	token   uint32
	delta   int16
}

func (s *manaSubTestState4EEBF0) hooks() manaDrainManaSubHooks4EEBF0[
	*manaSubTestObject4EEBF0,
	*manaSubTestUpdate4EEBF0,
	*manaSubTestPlayer4EEBF0,
] {
	return manaDrainManaSubHooks4EEBF0[
		*manaSubTestObject4EEBF0,
		*manaSubTestUpdate4EEBF0,
		*manaSubTestPlayer4EEBF0,
	]{
		classLow: func(obj *manaSubTestObject4EEBF0) uint8 {
			s.events = append(s.events, "class")
			return obj.class
		},
		loadUpdateData: func(obj *manaSubTestObject4EEBF0) *manaSubTestUpdate4EEBF0 {
			s.events = append(s.events, "update")
			return obj.update
		},
		godMode: func() bool {
			s.events = append(s.events, "god-mode")
			return s.godMode
		},
		loadManaCurrent: func(update *manaSubTestUpdate4EEBF0) uint16 {
			s.events = append(s.events, "load-current")
			return update.current
		},
		storeManaPrev: func(update *manaSubTestUpdate4EEBF0, mana uint16) {
			s.events = append(s.events, "store-prev")
			update.prev = mana
		},
		storeManaCurrent: func(update *manaSubTestUpdate4EEBF0, mana uint16) {
			s.events = append(s.events, "store-current")
			update.current = mana
		},
		loadPlayer: func(update *manaSubTestUpdate4EEBF0) *manaSubTestPlayer4EEBF0 {
			s.events = append(s.events, "player")
			return update.player
		},
		loadProtection: func(player *manaSubTestPlayer4EEBF0) uint32 {
			s.events = append(s.events, "token")
			return player.token
		},
		protectMana: func(token uint32, delta int16) {
			s.events = append(s.events, "protect")
			s.token, s.delta = token, delta
		},
	}
}

func TestManaDrainManaSub4EEBF0Gates(t *testing.T) {
	state := &manaSubTestState4EEBF0{}
	manaDrainManaSub4EEBF0((*manaSubTestObject4EEBF0)(nil), 1, state.hooks())
	if len(state.events) != 0 {
		t.Fatalf("nil events = %v", state.events)
	}

	state = &manaSubTestState4EEBF0{}
	manaDrainManaSub4EEBF0(&manaSubTestObject4EEBF0{class: 0xfb}, 1, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"class"}) {
		t.Fatalf("non-player events = %v", state.events)
	}

	state = &manaSubTestState4EEBF0{godMode: true}
	obj := &manaSubTestObject4EEBF0{class: 4}
	manaDrainManaSub4EEBF0(obj, 1, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"class", "update", "god-mode"}) {
		t.Fatalf("god-mode events = %v", state.events)
	}
}

func TestManaDrainManaSub4EEBF0OriginalProtectionBranch(t *testing.T) {
	tests := []struct {
		name        string
		current     uint16
		amount      uint8
		wantCurrent uint16
		wantDelta   int16
	}{
		{name: "new above amount", current: 100, amount: 10, wantCurrent: 90, wantDelta: -10},
		{name: "new at amount", current: 20, amount: 10, wantCurrent: 10, wantDelta: -10},
		{name: "new below amount", current: 15, amount: 10, wantCurrent: 5, wantDelta: -5},
		{name: "insufficient mana", current: 5, amount: 10, wantCurrent: 0, wantDelta: 0},
		{name: "equal mana", current: 10, amount: 10, wantCurrent: 0, wantDelta: 0},
		{name: "zero amount", current: 0xffff, amount: 0, wantCurrent: 0xffff, wantDelta: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := &manaSubTestPlayer4EEBF0{token: 0xfedcba98}
			update := &manaSubTestUpdate4EEBF0{current: tc.current, prev: 0x7777, player: player}
			obj := &manaSubTestObject4EEBF0{class: 0x84, update: update}
			state := &manaSubTestState4EEBF0{}
			manaDrainManaSub4EEBF0(obj, tc.amount, state.hooks())
			wantEvents := []string{
				"class", "update", "god-mode", "load-current", "store-prev",
				"store-current", "load-current", "player", "token", "protect",
			}
			if !reflect.DeepEqual(state.events, wantEvents) {
				t.Fatalf("events = %v, want %v", state.events, wantEvents)
			}
			if update.prev != tc.current || update.current != tc.wantCurrent {
				t.Fatalf("mana prev/current = %d/%d, want %d/%d", update.prev, update.current, tc.current, tc.wantCurrent)
			}
			if state.token != player.token || state.delta != tc.wantDelta {
				t.Fatalf("protect = (%#x,%d), want (%#x,%d)", state.token, state.delta, player.token, tc.wantDelta)
			}
		})
	}
}

func TestManaDrainManaSub4EEBF0NilPlayerFaultAfterManaStores(t *testing.T) {
	update := &manaSubTestUpdate4EEBF0{current: 15, prev: 99}
	obj := &manaSubTestObject4EEBF0{class: 4, update: update}
	state := &manaSubTestState4EEBF0{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil player did not fault")
		}
		if update.prev != 15 || update.current != 5 {
			t.Fatalf("mana before fault = %d/%d, want 15/5", update.prev, update.current)
		}
		want := []string{
			"class", "update", "god-mode", "load-current", "store-prev",
			"store-current", "load-current", "player", "token",
		}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()
	manaDrainManaSub4EEBF0(obj, 10, state.hooks())
}
