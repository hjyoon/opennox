package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type mapEntryStateSetTestWorld4FC580 struct {
	events  []string
	faultAt int
	value   int32
	stored  int32
}

func (w *mapEntryStateSetTestWorld4FC580) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *mapEntryStateSetTestWorld4FC580) hooks() mapEntryStateSetHooks4FC580 {
	return mapEntryStateSetHooks4FC580{
		loadValueArg: func() int32 {
			value := w.value
			w.record(fmt.Sprintf("load:%#08x", uint32(value)))
			return value
		},
		storeValue: func(value int32) {
			w.record(fmt.Sprintf("store:%#08x", uint32(value)))
			w.stored = value
		},
	}
}

func TestMapEntryStateSet4FC580PreservesEveryBit(t *testing.T) {
	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
		-1985229329, // 0x89abcdef
	} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			w := mapEntryStateSetTestWorld4FC580{value: value}
			got := mapEntryStateSet4FC580(w.hooks())
			if uint32(got) != uint32(value) {
				t.Fatalf("return = %#08x, want %#08x", uint32(got), uint32(value))
			}
			if uint32(w.stored) != uint32(value) {
				t.Fatalf("stored = %#08x, want %#08x", uint32(w.stored), uint32(value))
			}
			wantEvents := []string{
				fmt.Sprintf("load:%#08x", uint32(value)),
				fmt.Sprintf("store:%#08x", uint32(value)),
			}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
			}
		})
	}
}

func TestMapEntryStateSet4FC580FaultPrefixes(t *testing.T) {
	for faultAt, wantEvents := range [][]string{
		{"load:0x89abcdef"},
		{"load:0x89abcdef", "store:0x89abcdef"},
	} {
		t.Run(fmt.Sprintf("fault-%d", faultAt+1), func(t *testing.T) {
			w := mapEntryStateSetTestWorld4FC580{
				faultAt: faultAt + 1,
				value:   -1985229329, // 0x89abcdef
				stored:  0x12345678,
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if !reflect.DeepEqual(w.events, wantEvents) {
					t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
				}
				if w.stored != 0x12345678 {
					t.Fatalf("stored = %#08x, want unchanged", uint32(w.stored))
				}
			}()
			mapEntryStateSet4FC580(w.hooks())
		})
	}
}
