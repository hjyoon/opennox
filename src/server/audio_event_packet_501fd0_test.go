package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	audioPacketUnit501FD0     = uint64(0x100000001)
	audioPacketOther501FD0    = uint64(0x200000002)
	audioPacketEvent501FD0    = uint64(0x300000003)
	audioPacketUpdate501FD0   = uint64(0x400000004)
	audioPacketPlayer501FD0   = uint64(0x500000005)
	audioPacketPlayer2_501FD0 = uint64(0x600000006)
)

type audioEventPacketWorld501FD0 struct {
	events  []string
	faultAt int

	updates       map[uint64]uint64
	eventObjects  map[uint64]uint64
	eventX        map[uint64]float32
	eventSound    map[uint64]int32
	players       map[uint64]uint64
	playerX       map[uint64]float32
	playerIndex   map[uint64]uint8
	windowWidth   int32
	floatResult   *int32
	afterFloat    func()
	enqueueResult bool

	gotIndex  uint8
	gotKind   uint8
	gotPacket [4]byte
}

func audioEventPacketTestFloatToInt501FD0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func newAudioEventPacketWorld501FD0() *audioEventPacketWorld501FD0 {
	return &audioEventPacketWorld501FD0{
		updates:       map[uint64]uint64{audioPacketUnit501FD0: audioPacketUpdate501FD0},
		eventObjects:  map[uint64]uint64{audioPacketEvent501FD0: audioPacketUnit501FD0},
		eventX:        map[uint64]float32{audioPacketEvent501FD0: 13.5},
		eventSound:    map[uint64]int32{audioPacketEvent501FD0: 0x12ab},
		players:       map[uint64]uint64{audioPacketUpdate501FD0: audioPacketPlayer501FD0},
		playerX:       map[uint64]float32{audioPacketPlayer501FD0: 10},
		playerIndex:   map[uint64]uint8{audioPacketPlayer501FD0: 5, audioPacketPlayer2_501FD0: 0x87},
		windowWidth:   2,
		enqueueResult: true,
	}
}

func (w *audioEventPacketWorld501FD0) observe(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *audioEventPacketWorld501FD0) hooks() audioEventPacketHooks501FD0[uint64, uint64, uint64, uint64] {
	return audioEventPacketHooks501FD0[uint64, uint64, uint64, uint64]{
		loadUpdate: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("update:%#x", unit))
			return w.updates[unit]
		},
		loadEventObject: func(event uint64) uint64 {
			w.observe(fmt.Sprintf("event-object:%#x", event))
			return w.eventObjects[event]
		},
		loadEventPositionX: func(event uint64) float32 {
			w.observe(fmt.Sprintf("event-x:%#x", event))
			return w.eventX[event]
		},
		loadEventSound: func(event uint64) int32 {
			w.observe(fmt.Sprintf("event-sound:%#x", event))
			return w.eventSound[event]
		},
		loadPlayer: func(update uint64) uint64 {
			w.observe(fmt.Sprintf("player:%#x", update))
			return w.players[update]
		},
		loadPlayerPositionX: func(player uint64) float32 {
			w.observe(fmt.Sprintf("player-x:%#x", player))
			return w.playerX[player]
		},
		floatToInt: func(value float32) int32 {
			w.observe(fmt.Sprintf("float-to-int:%g", value))
			result := audioEventPacketTestFloatToInt501FD0(value)
			if w.floatResult != nil {
				result = *w.floatResult
			}
			if w.afterFloat != nil {
				w.afterFloat()
			}
			return result
		},
		loadWindowWidth: func() int32 {
			w.observe("window-width")
			return w.windowWidth
		},
		loadPlayerIndex: func(player uint64) uint8 {
			w.observe(fmt.Sprintf("player-index:%#x", player))
			return w.playerIndex[player]
		},
		enqueue: func(index, kind uint8, packet [4]byte) bool {
			w.observe(fmt.Sprintf("enqueue:%02x:%d:%x", index, kind, packet))
			w.gotIndex, w.gotKind, w.gotPacket = index, kind, packet
			return w.enqueueResult
		},
	}
}

func (w *audioEventPacketWorld501FD0) run(percentage int32) bool {
	return sendAudioEventPacket501FD0(audioPacketUnit501FD0, audioPacketEvent501FD0, percentage, w.hooks())
}

func TestSendAudioEventPacket501FD0OrderPacketAndPlayerReload(t *testing.T) {
	w := newAudioEventPacketWorld501FD0()
	w.afterFloat = func() {
		w.players[audioPacketUpdate501FD0] = audioPacketPlayer2_501FD0
	}
	if !w.run(-3) {
		t.Fatal("enqueue result = false, want true")
	}
	wantEvents := []string{
		"update:0x100000001",
		"event-object:0x300000003",
		"event-x:0x300000003",
		"event-sound:0x300000003",
		"player:0x400000004",
		"player-x:0x500000005",
		"float-to-int:3.5",
		"window-width",
		"player:0x400000004",
		"player-index:0x600000006",
		"enqueue:87:1:a7c8abf6",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, wantEvents)
	}
	if w.gotIndex != 0x87 || w.gotKind != audioEventPacketListKind501FD0 || w.gotPacket != [4]byte{0xa7, 0xc8, 0xab, 0xf6} {
		t.Fatalf("enqueue = index %#x kind %d packet %x", w.gotIndex, w.gotKind, w.gotPacket)
	}
}

func TestSendAudioEventPacket501FD0GenericHandlesAndSignedMovement(t *testing.T) {
	w := newAudioEventPacketWorld501FD0()
	w.eventObjects[audioPacketEvent501FD0] = audioPacketOther501FD0
	w.eventX[audioPacketEvent501FD0] = 6.5
	w.eventSound[audioPacketEvent501FD0] = 0x123
	w.enqueueResult = false
	if w.run(0) {
		t.Fatal("enqueue result = true, want propagated false")
	}
	if w.gotPacket != [4]byte{0xa6, 0x38, 0x23, 0x01} {
		t.Fatalf("packet = %x, want a6382301", w.gotPacket)
	}
}

func TestAudioEventPackedWord501FD0(t *testing.T) {
	tests := []struct {
		sound      int32
		percentage int32
		want       uint16
	}{
		{0x12ab, -3, 0xf6ab},
		{0, 63, 0xfc00},
		{0xffff, math.MinInt32, 0xffff},
		{-32768, 0, 0x8000},
	}
	for _, test := range tests {
		if got := audioEventPackedWord501FD0(test.sound, test.percentage); got != test.want {
			t.Errorf("packed(%#x, %d) = %#x, want %#x", test.sound, test.percentage, got, test.want)
		}
	}
}

func TestSendAudioEventPacket501FD0SignedDwordMultiply(t *testing.T) {
	w := newAudioEventPacketWorld501FD0()
	converted := int32(math.MaxInt32)
	w.floatResult = &converted
	w.run(0)
	// MaxInt32 * 50 wraps to -50 in the PE32 EAX register.
	if w.gotPacket[1] != 0xce {
		t.Fatalf("movement = %#x, want low byte of -50 (0xce)", w.gotPacket[1])
	}
}

func TestAudioEventSignedDivide501FD0(t *testing.T) {
	for _, test := range []struct {
		numerator   int32
		denominator int32
		want        int32
	}{
		{201, 2, 100},
		{-201, 2, -100},
		{201, -2, -100},
		{math.MinInt32, 2, -1073741824},
	} {
		if got := audioEventSignedDivide501FD0(test.numerator, test.denominator); got != test.want {
			t.Errorf("divide(%d, %d) = %d, want %d", test.numerator, test.denominator, got, test.want)
		}
	}
	for _, test := range []struct {
		name        string
		numerator   int32
		denominator int32
	}{
		{"zero", 1, 0},
		{"overflow", math.MinInt32, -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("divide did not panic")
				}
			}()
			audioEventSignedDivide501FD0(test.numerator, test.denominator)
		})
	}
}

func TestSendAudioEventPacket501FD0DivideFaultPrecedesPlayerReload(t *testing.T) {
	w := newAudioEventPacketWorld501FD0()
	w.windowWidth = 1
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("zero half-width did not panic")
			}
		}()
		w.run(0)
	}()
	want := []string{
		"update:0x100000001", "event-object:0x300000003",
		"event-x:0x300000003", "event-sound:0x300000003",
		"player:0x400000004", "player-x:0x500000005",
		"float-to-int:3.5", "window-width",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("fault prefix = %v, want %v", w.events, want)
	}
}

func TestSendAudioEventPacket501FD0InjectedFaultPrefixes(t *testing.T) {
	baseline := newAudioEventPacketWorld501FD0()
	baseline.run(-3)
	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		w := newAudioEventPacketWorld501FD0()
		w.faultAt = faultAt
		panicked := false
		func() {
			defer func() {
				panicked = recover() != nil
			}()
			w.run(-3)
		}()
		if !panicked {
			t.Fatalf("fault %d did not panic", faultAt)
		}
		want := baseline.events[:faultAt]
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("fault %d prefix = %v, want %v", faultAt, w.events, want)
		}
	}
}
