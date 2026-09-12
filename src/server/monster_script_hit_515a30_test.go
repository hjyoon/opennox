package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

const monsterScriptHitHighUnit515A30 = uint64(0x7faedbb66e70)

type monsterScriptHitTestWorld515A30 struct {
	events   []string
	classLow uint8
	flags    uint32
	can      bool
	pos      *types.Pointf
	nilPush  map[uint32]bool
	args     map[uint64][4]uint32
	pushes   []uint32
}

func (w *monsterScriptHitTestWorld515A30) hooks() monsterScriptHitHooks515A30[uint64, uint64] {
	return monsterScriptHitHooks515A30[uint64, uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		canAttack: func(unit uint64) bool {
			w.events = append(w.events, fmt.Sprintf("can:%x", unit))
			return w.can
		},
		clearActionStack: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("clear:%x", unit))
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%x:%d", unit, action))
			w.pushes = append(w.pushes, action)
			if action == 17 && w.pos != nil {
				w.pos.X = math.Float32frombits(0x7fa12345)
				w.pos.Y = math.Float32frombits(0x80000000)
			}
			if w.nilPush[action] {
				return 0
			}
			return 0x7fae00000000 + uint64(len(w.pushes))*0x100
		},
		storeArgBits: func(item uint64, index int, value uint32) {
			w.events = append(w.events, fmt.Sprintf("store:%x:%d:%08x", item, index, value))
			args := w.args[item]
			args[index] = value
			w.args[item] = args
		},
		loadMeleeRange: func(unit uint64) float32 {
			w.events = append(w.events, fmt.Sprintf("range:%x", unit))
			return 12.5
		},
		loadRadius: func(unit uint64) float32 {
			w.events = append(w.events, fmt.Sprintf("radius:%x", unit))
			return 3.25
		},
	}
}

func newMonsterScriptHitTestWorld515A30() *monsterScriptHitTestWorld515A30 {
	return &monsterScriptHitTestWorld515A30{
		classLow: 2,
		can:      true,
		args:     make(map[uint64][4]uint32),
	}
}

func TestMonsterScriptHitMissile515B80ExactOrderAndBits(t *testing.T) {
	w := newMonsterScriptHitTestWorld515A30()
	pos := types.Ptf(1, 2)
	w.pos = &pos
	monsterScriptHitMissile515B80(monsterScriptHitHighUnit515A30, &pos, w.hooks())
	want := []string{
		"class:7faedbb66e70", "flags:7faedbb66e70", "can:7faedbb66e70",
		"clear:7faedbb66e70", "push:7faedbb66e70:32",
		"store:7fae00000100:0:00000011", "push:7faedbb66e70:17",
		"store:7fae00000200:0:7fa12345", "store:7fae00000200:1:80000000",
		"store:7fae00000200:2:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if got := w.args[0x7fae00000200]; got != [4]uint32{0x7fa12345, 0x80000000, 0, 0} {
		t.Fatalf("missile args = %#v", got)
	}
}

func TestMonsterScriptHitMelee515A30ExactOrderAndBits(t *testing.T) {
	w := newMonsterScriptHitTestWorld515A30()
	pos := types.Ptf(12.5, -7.25)
	monsterScriptHitMelee515A30(monsterScriptHitHighUnit515A30, &pos, w.hooks())
	want := []string{
		"class:7faedbb66e70", "flags:7faedbb66e70", "can:7faedbb66e70",
		"clear:7faedbb66e70", "push:7faedbb66e70:32",
		"store:7fae00000100:0:00000010", "push:7faedbb66e70:16",
		"push:7faedbb66e70:51", "range:7faedbb66e70", "radius:7faedbb66e70",
		"store:7fae00000300:0:417c0000", "store:7fae00000300:2:41480000",
		"store:7fae00000300:3:c0e80000", "push:7faedbb66e70:7",
		"store:7fae00000400:0:41480000", "store:7fae00000400:1:c0e80000",
		"store:7fae00000400:2:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestMonsterScriptHit515A30OriginalGatesAndNilPush(t *testing.T) {
	pos := types.Ptf(1, 2)
	for _, tc := range []struct {
		name    string
		unit    uint64
		pos     *types.Pointf
		class   uint8
		flags   uint32
		can     bool
		missile bool
		want    []string
	}{
		{name: "nil unit", missile: true, pos: &pos},
		{name: "nil missile position", missile: true, unit: monsterScriptHitHighUnit515A30, class: 2, can: true},
		{name: "not monster", missile: true, unit: monsterScriptHitHighUnit515A30, pos: &pos, class: 1, can: true, want: []string{"class:7faedbb66e70"}},
		{name: "dead", missile: true, unit: monsterScriptHitHighUnit515A30, pos: &pos, class: 2, flags: 0x8000, can: true, want: []string{"class:7faedbb66e70", "flags:7faedbb66e70"}},
		{name: "cannot shoot", missile: true, unit: monsterScriptHitHighUnit515A30, pos: &pos, class: 2, want: []string{"class:7faedbb66e70", "flags:7faedbb66e70", "can:7faedbb66e70"}},
		{name: "cannot melee", unit: monsterScriptHitHighUnit515A30, pos: &pos, class: 2, want: []string{"class:7faedbb66e70", "flags:7faedbb66e70", "can:7faedbb66e70"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMonsterScriptHitTestWorld515A30()
			w.classLow, w.flags, w.can = tc.class, tc.flags, tc.can
			if tc.missile {
				monsterScriptHitMissile515B80(tc.unit, tc.pos, w.hooks())
			} else {
				monsterScriptHitMelee515A30(tc.unit, tc.pos, w.hooks())
			}
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %q, want %q", w.events, tc.want)
			}
		})
	}
	for _, missile := range []bool{false, true} {
		w := newMonsterScriptHitTestWorld515A30()
		w.nilPush = map[uint32]bool{32: true, 51: true, 17: true, 7: true}
		if missile {
			monsterScriptHitMissile515B80(monsterScriptHitHighUnit515A30, &pos, w.hooks())
			if !reflect.DeepEqual(w.pushes, []uint32{32, 17}) || len(w.args) != 0 {
				t.Fatalf("nil missile pushes/args = %v/%v", w.pushes, w.args)
			}
		} else {
			monsterScriptHitMelee515A30(monsterScriptHitHighUnit515A30, &pos, w.hooks())
			if !reflect.DeepEqual(w.pushes, []uint32{32, 16, 51, 7}) || len(w.args) != 0 {
				t.Fatalf("nil melee pushes/args = %v/%v", w.pushes, w.args)
			}
		}
	}
}
