package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const unitBuffTestObject4FF350 = uint64(0x1234567889abcdef)

type unitBuffTestWorld4FF350 struct {
	events  []string
	faultAt int
	buff    int32
	buffs   uint32
}

func (w *unitBuffTestWorld4FF350) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitBuffTestWorld4FF350) hooks() UnitBuffTestHooks4FF350[uint64] {
	return UnitBuffTestHooks4FF350[uint64]{
		LoadBuffArg: func() int32 {
			w.observe(fmt.Sprintf("buff=%08x", uint32(w.buff)))
			return w.buff
		},
		LoadBuffs: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("buffs:%016x=%08x", unit, w.buffs))
			return w.buffs
		},
	}
}

func TestUnitBuffTest4FF350NullShortCircuit(t *testing.T) {
	w := &unitBuffTestWorld4FF350{
		buff:  math.MinInt32,
		buffs: math.MaxUint32,
	}
	if got := UnitBuffTest4FF350(uint64(0), w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}
	if len(w.events) != 0 {
		t.Fatalf("events = %q, want no argument or buff-word load", w.events)
	}
}

func TestUnitBuffTest4FF350ExactLoadOrderAndNativeToken(t *testing.T) {
	w := &unitBuffTestWorld4FF350{buff: -1, buffs: uint32(1) << 31}
	if got := UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical one", got)
	}
	want := []string{
		"buff=ffffffff",
		"buffs:1234567889abcdef=80000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact load order and full token %q", w.events, want)
	}
}

func TestUnitBuffTest4FF350MasksShiftCountLikeX86CL(t *testing.T) {
	tests := []struct {
		buff  int32
		index uint32
	}{
		{0, 0},
		{1, 1},
		{31, 31},
		{32, 0},
		{33, 1},
		{-1, 31},
		{-32, 0},
		{-33, 31},
		{math.MinInt32, 0},
		{math.MaxInt32, 31},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("buff-%08x", uint32(tc.buff)), func(t *testing.T) {
			w := &unitBuffTestWorld4FF350{
				buff:  tc.buff,
				buffs: uint32(1) << tc.index,
			}
			if got := UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks()); got != 1 {
				t.Fatalf("matching result = %d, want canonical one for low-five-bit index %d", got, tc.index)
			}

			w.events = nil
			w.buffs = uint32(1) << ((tc.index + 1) & 31)
			if got := UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks()); got != 0 {
				t.Fatalf("adjacent-bit result = %d, want canonical zero", got)
			}
		})
	}
}

func TestUnitBuffTest4FF350CanonicalBoolean(t *testing.T) {
	for _, buffs := range []uint32{1, 3, 0x7fffffff, math.MaxUint32} {
		w := &unitBuffTestWorld4FF350{buff: 0, buffs: buffs}
		if got := UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks()); got != 1 {
			t.Fatalf("buffs %#08x result = %d, want exactly one", buffs, got)
		}
	}

	w := &unitBuffTestWorld4FF350{buff: 5, buffs: math.MaxUint32 &^ (uint32(1) << 5)}
	if got := UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks()); got != 0 {
		t.Fatalf("clear selected bit result = %d, want exactly zero", got)
	}
}

func TestUnitBuffTest4FF350AllFaultPrefixes(t *testing.T) {
	baseline := &unitBuffTestWorld4FF350{buff: 7, buffs: uint32(1) << 7}
	UnitBuffTest4FF350(unitBuffTestObject4FF350, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &unitBuffTestWorld4FF350{
				faultAt: faultAt,
				buff:    7,
				buffs:   uint32(1) << 7,
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				UnitBuffTest4FF350(unitBuffTestObject4FF350, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}
