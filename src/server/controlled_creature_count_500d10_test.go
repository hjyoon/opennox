package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	controlledCreatureHighOwner500D10 = uint64(0x7f03ec213490)
	controlledCreatureHighFirst500D10 = uint64(0x7ff71d949570)
	controlledCreatureHighNext500D10  = uint64(0x7fb3078c7430)
)

type controlledCreatureCountTestWorld500D10 struct {
	events      []string
	faultAt     int
	first       uint64
	next        map[uint64]uint64
	monitored   map[uint64]bool
	subclassLow map[uint64]uint8
	onMonitor   func(owner, unit uint64)
}

func (w *controlledCreatureCountTestWorld500D10) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *controlledCreatureCountTestWorld500D10) hooks() controlledCreatureCountHooks500D10[uint64] {
	return controlledCreatureCountHooks500D10[uint64]{
		loadFirstOwned: func(owner uint64) uint64 {
			w.observe(fmt.Sprintf("first:%x", owner))
			return w.first
		},
		isMonitored: func(owner, unit uint64) bool {
			w.observe(fmt.Sprintf("monitor:%x:%x", owner, unit))
			if w.onMonitor != nil {
				w.onMonitor(owner, unit)
			}
			return w.monitored[unit]
		},
		loadSubclassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("subclass:%x", unit))
			return w.subclassLow[unit]
		},
		loadNextOwned: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("next:%x", unit))
			return w.next[unit]
		},
	}
}

func TestControlledCreatureSize500D50BitPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name string
		low  uint8
		want uint32
	}{
		{name: "neither", low: 0x00, want: 4},
		{name: "small", low: 0x01, want: 1},
		{name: "medium", low: 0x02, want: 2},
		{name: "small wins", low: 0x03, want: 1},
		{name: "upper bits ignored", low: 0xfc, want: 4},
		{name: "small wins with upper bits", low: 0xff, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := controlledCreatureSize500D50(tc.low); got != tc.want {
				t.Fatalf("size(%#02x) = %d, want %d", tc.low, got, tc.want)
			}
		})
	}
}

func TestControlledCreatureCount500D10OrderAndNativeWidthHandles(t *testing.T) {
	w := &controlledCreatureCountTestWorld500D10{
		first: controlledCreatureHighFirst500D10,
		next: map[uint64]uint64{
			controlledCreatureHighFirst500D10: controlledCreatureHighNext500D10,
		},
		monitored: map[uint64]bool{
			controlledCreatureHighFirst500D10: true,
		},
		subclassLow: map[uint64]uint8{
			controlledCreatureHighFirst500D10: controlledCreatureSmallBit500D50,
		},
	}
	if got := controlledCreatureCount500D10(controlledCreatureHighOwner500D10, w.hooks()); got != 1 {
		t.Fatalf("count = %d, want 1", got)
	}
	want := []string{
		"first:7f03ec213490",
		"monitor:7f03ec213490:7ff71d949570",
		"subclass:7ff71d949570",
		"next:7ff71d949570",
		"monitor:7f03ec213490:7fb3078c7430",
		"next:7fb3078c7430",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, want)
	}
}

func TestControlledCreatureCount500D10ReadsLiveSuccessorAfterMonitor(t *testing.T) {
	const owner, first, stale, replacement uint64 = 0x100000001, 0x200000002, 0x300000003, 0x400000004
	w := &controlledCreatureCountTestWorld500D10{
		first: first,
		next: map[uint64]uint64{
			first: stale,
		},
		monitored: map[uint64]bool{
			first:       true,
			replacement: true,
		},
		subclassLow: map[uint64]uint8{
			first:       controlledCreatureMediumBit500D50,
			replacement: 0,
		},
	}
	w.onMonitor = func(_ uint64, unit uint64) {
		if unit == first {
			w.next[first] = replacement
		}
	}
	if got := controlledCreatureCount500D10(owner, w.hooks()); got != 6 {
		t.Fatalf("count = %d, want 6", got)
	}
	for _, event := range w.events {
		if event == "monitor:100000001:300000003" {
			t.Fatal("followed successor captured before the monitor callback")
		}
	}
}

func TestControlledCreatureCount500D10NoOwnerPreflightAndFaultPrefixes(t *testing.T) {
	baseline := &controlledCreatureCountTestWorld500D10{
		first:       controlledCreatureHighFirst500D10,
		next:        map[uint64]uint64{},
		monitored:   map[uint64]bool{controlledCreatureHighFirst500D10: true},
		subclassLow: map[uint64]uint8{controlledCreatureHighFirst500D10: 0},
	}
	controlledCreatureCount500D10(uint64(0), baseline.hooks())
	want := append([]string(nil), baseline.events...)
	if len(want) == 0 || want[0] != "first:0" {
		t.Fatalf("zero owner events = %q, want first load before any return", want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := &controlledCreatureCountTestWorld500D10{
				faultAt:     faultAt,
				first:       controlledCreatureHighFirst500D10,
				next:        map[uint64]uint64{},
				monitored:   map[uint64]bool{controlledCreatureHighFirst500D10: true},
				subclassLow: map[uint64]uint8{controlledCreatureHighFirst500D10: 0},
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				controlledCreatureCount500D10(uint64(0), w.hooks())
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

func TestControlledCreatureCountAdd500D10WrapsPE32Accumulator(t *testing.T) {
	if got := controlledCreatureCountAdd500D10(math.MaxUint32, 4); got != 3 {
		t.Fatalf("wrapping add = %#x, want 0x3", got)
	}
	signedBits := uint32(math.MaxInt32)
	signedBits++
	if got := int32(signedBits); got != math.MinInt32 {
		t.Fatalf("signed result bit pattern = %d, want %d", got, int32(math.MinInt32))
	}
}
