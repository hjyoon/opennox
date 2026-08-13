package opennox

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

func TestUnitMoveTruncLow32(t *testing.T) {
	tests := []struct {
		name string
		in   float32
		want uint32
	}{
		{name: "zero", in: 0, want: 0},
		{name: "positive fraction", in: 1.99, want: 1},
		{name: "negative fraction", in: -1.99, want: math.MaxUint32},
		{name: "positive wrap", in: float32(math.Ldexp(1, 32)), want: 0},
		{name: "positive wrapped residue", in: float32(math.Ldexp(1, 32) + 512), want: 512},
		{name: "negative wrap", in: -float32(math.Ldexp(1, 32)), want: 0},
		{name: "minimum int64", in: -float32(math.Ldexp(1, 63)), want: 0},
		{name: "positive overflow", in: float32(math.Ldexp(1, 63)), want: 0},
		{name: "negative overflow", in: -float32(math.Ldexp(1, 64)), want: 0},
		{name: "positive infinity", in: float32(math.Inf(1)), want: 0},
		{name: "negative infinity", in: float32(math.Inf(-1)), want: 0},
		{name: "nan", in: float32(math.NaN()), want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := unitMoveTruncLow32(tc.in); got != tc.want {
				t.Fatalf("unitMoveTruncLow32(%v) = %#x, want %#x", tc.in, got, tc.want)
			}
		})
	}
}

func TestUnitMoveSameIntegerPosition4E7010(t *testing.T) {
	tests := []struct {
		name string
		old  types.Pointf
		next types.Pointf
		want bool
	}{
		{
			name: "equal",
			old:  types.Ptf(12.25, -8.75),
			next: types.Ptf(12.99, -8.01),
			want: true,
		},
		{
			name: "low 32 bit alias",
			old:  types.Ptf(0, 512),
			next: types.Ptf(float32(math.Ldexp(1, 32)), float32(math.Ldexp(1, 32)+512)),
			want: true,
		},
		{
			name: "different x",
			old:  types.Ptf(1, 2),
			next: types.Ptf(2, 2),
			want: false,
		},
		{
			name: "different y",
			old:  types.Ptf(1, 2),
			next: types.Ptf(1, 3),
			want: false,
		},
		{
			name: "invalid conversions alias zero",
			old:  types.Ptf(float32(math.NaN()), float32(math.Inf(1))),
			next: types.Ptf(0, 0),
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := unitMoveSameIntegerPosition4E7010(tc.old, tc.next); got != tc.want {
				t.Fatalf("same integer position for %v and %v = %v, want %v", tc.old, tc.next, got, tc.want)
			}
		})
	}
}

func TestUnitMoveOnline4E7010ReloadsPlayerInOriginalOrder(t *testing.T) {
	type updateData struct {
		moveFrame uint32
		player    int
	}
	ud := &updateData{player: 1}
	indices := map[int]uint8{1: 11, 2: 22, 3: 33, 4: 44}
	frames := []uint32{0x01020304, 0x44332211}
	var calls []string
	frameCall := 0

	h := unitMoveOnline4E7010Hooks[*updateData, int]{
		frame: func() uint32 {
			frameCall++
			calls = append(calls, "frame"+string(rune('0'+frameCall)))
			if frameCall == 1 {
				ud.player = 2
			} else {
				ud.player = 4
			}
			return frames[frameCall-1]
		},
		setMoveFrame: func(got *updateData, frame uint32) {
			calls = append(calls, "set-frame")
			if got != ud || frame != frames[0] {
				t.Fatalf("set frame got %p/%#x, want %p/%#x", got, frame, ud, frames[0])
			}
			got.moveFrame = frame
			got.player = 3
		},
		player: func(got *updateData) int {
			calls = append(calls, "player")
			if got != ud {
				t.Fatalf("player lookup used update data %p, want cached %p", got, ud)
			}
			return got.player
		},
		playerIndex: func(player int) uint8 {
			calls = append(calls, "index")
			return indices[player]
		},
		markPlayer: func(index uint8) {
			calls = append(calls, "mark")
			if index != 22 {
				t.Fatalf("marked player %d, want first reloaded player 22", index)
			}
		},
		sendPacket: func(index uint8, buf [5]byte) {
			calls = append(calls, "send")
			if index != 44 {
				t.Fatalf("sent to player %d, want second reloaded player 44", index)
			}
			if want := [5]byte{0xea, 0x11, 0x22, 0x33, 0x44}; buf != want {
				t.Fatalf("packet = % x, want % x", buf, want)
			}
		},
	}

	unitMoveOnline4E7010(ud, h)
	if ud.moveFrame != frames[0] {
		t.Fatalf("stored move frame = %#x, want %#x", ud.moveFrame, frames[0])
	}
	wantCalls := []string{"frame1", "player", "set-frame", "index", "mark", "frame2", "player", "index", "send"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("call order = %v, want %v", calls, wantCalls)
	}
}
