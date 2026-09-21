package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/spell"
)

const (
	lesserHealHighTarget52DD50 = uint64(0x7fb8ed1723d0)
	lesserHealHighCaster52DD50 = uint64(0x2ad51258fad0)
)

type lesserHealTestArg52DD50 struct {
	target uint64
}

type lesserHealTestWorld52DD50 struct {
	hp       map[uint64]uint16
	maxHP    map[uint64]uint16
	classLow map[uint64]uint8
	class    map[uint64]uint8
	events   []string
}

func (w *lesserHealTestWorld52DD50) hooks() lesserHealHooks52DD50[uint64, *lesserHealTestArg52DD50] {
	return lesserHealHooks52DD50[uint64, *lesserHealTestArg52DD50]{
		loadTarget: func(arg *lesserHealTestArg52DD50) uint64 {
			w.events = append(w.events, fmt.Sprintf("target:%#x", arg.target))
			return arg.target
		},
		getHP: func(obj uint64) uint16 {
			w.events = append(w.events, fmt.Sprintf("hp:%#x", obj))
			return w.hp[obj]
		},
		getMaxHP: func(obj uint64) uint16 {
			w.events = append(w.events, fmt.Sprintf("max:%#x", obj))
			return w.maxHP[obj]
		},
		manaCost: func(id spell.ID, level int) int {
			w.events = append(w.events, fmt.Sprintf("cost:%s:%d", id, level))
			return 17
		},
		refundMana: func(obj uint64, amount int) {
			w.events = append(w.events, fmt.Sprintf("refund:%#x:%d", obj, amount))
		},
		balance: func(key string) float32 {
			w.events = append(w.events, "balance:"+key)
			return 12.5
		},
		loadClassLow: func(obj uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class-low:%#x", obj))
			return w.classLow[obj]
		},
		loadPlayerCls: func(obj uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("player-class:%#x", obj))
			return w.class[obj]
		},
		classHealth: func(class uint8) float32 {
			w.events = append(w.events, fmt.Sprintf("class-health:%d", class))
			return 1.6
		},
		floatToInt: func(value float32) int32 {
			w.events = append(w.events, fmt.Sprintf("round:%g", value))
			return lesserHealFloatToInt52DD50(value)
		},
		adjustHP: func(obj uint64, amount int32) {
			w.events = append(w.events, fmt.Sprintf("adjust:%#x:%d", obj, amount))
		},
		audio: func(id spell.ID, obj uint64) {
			w.events = append(w.events, fmt.Sprintf("audio:%s:%#x", id, obj))
		},
	}
}

func newLesserHealTestWorld52DD50() *lesserHealTestWorld52DD50 {
	return &lesserHealTestWorld52DD50{
		hp:       map[uint64]uint16{lesserHealHighTarget52DD50: 20},
		maxHP:    map[uint64]uint16{lesserHealHighTarget52DD50: 100},
		classLow: map[uint64]uint8{lesserHealHighCaster52DD50: 4},
		class:    map[uint64]uint8{lesserHealHighCaster52DD50: 1},
	}
}

func TestLesserHeal52DD50PreservesNativeWidthAndOriginalOrder(t *testing.T) {
	w := newLesserHealTestWorld52DD50()
	arg := &lesserHealTestArg52DD50{target: lesserHealHighTarget52DD50}
	if got := lesserHeal52DD50(spell.SPELL_LESSER_HEAL, lesserHealHighTarget52DD50, lesserHealHighCaster52DD50, arg, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"target:0x7fb8ed1723d0",
		"target:0x7fb8ed1723d0",
		"target:0x7fb8ed1723d0",
		"hp:0x7fb8ed1723d0",
		"max:0x7fb8ed1723d0",
		"balance:LesserHealAmount",
		"class-low:0x2ad51258fad0",
		"player-class:0x2ad51258fad0",
		"class-health:1",
		"target:0x7fb8ed1723d0",
		"round:20",
		"adjust:0x7fb8ed1723d0:20",
		"target:0x7fb8ed1723d0",
		"audio:SPELL_LESSER_HEAL:0x7fb8ed1723d0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
}

func TestLesserHeal52DD50FullSelfTargetRefundsMana(t *testing.T) {
	w := newLesserHealTestWorld52DD50()
	w.hp[lesserHealHighTarget52DD50] = 100
	arg := &lesserHealTestArg52DD50{target: lesserHealHighTarget52DD50}
	if got := lesserHeal52DD50(spell.SPELL_LESSER_HEAL, lesserHealHighTarget52DD50, lesserHealHighCaster52DD50, arg, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"target:0x7fb8ed1723d0",
		"target:0x7fb8ed1723d0",
		"target:0x7fb8ed1723d0",
		"hp:0x7fb8ed1723d0",
		"max:0x7fb8ed1723d0",
		"target:0x7fb8ed1723d0",
		"cost:SPELL_LESSER_HEAL:1",
		"refund:0x2ad51258fad0:17",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestLesserHeal52DD50NilTargetStopsAfterSecondLiveLoad(t *testing.T) {
	w := newLesserHealTestWorld52DD50()
	arg := &lesserHealTestArg52DD50{}
	if got := lesserHeal52DD50(spell.SPELL_LESSER_HEAL, uint64(0), lesserHealHighCaster52DD50, arg, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{"target:0x0", "target:0x0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestLesserHealFloatToInt52DD50(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  int32
	}{
		{2.5, 2},
		{3.5, 4},
		{-2.5, -2},
		{float32(math.NaN()), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{2147483648, math.MinInt32},
	} {
		if got := lesserHealFloatToInt52DD50(tc.value); got != tc.want {
			t.Errorf("lesserHealFloatToInt52DD50(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
