package server

import (
	"fmt"
	"reflect"
	"testing"
)

const summonManaCostHighUnit500CA0 = uint64(0x7ff71d949570)

var summonManaCostOracleTable500CA0 = [...]int32{
	15, 60, 60, 60, 30, 30, 30, 15, 60, 60,
	15, 30, 30, 30, 85, 85, 30, 60, 85, 60,
	30, 30, 60, 30, 15, 30, 85, 30, 30, 15,
	60, 30, 30, 30, 30, 60, 60, 60, 60, 60,
}

type summonManaCostTestWorld500CA0 struct {
	events   []string
	faultAt  int
	classLow uint8
	costs    map[uint32]int32
}

func (w *summonManaCostTestWorld500CA0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *summonManaCostTestWorld500CA0) hooks() summonManaCostHooks500CA0[uint64] {
	return summonManaCostHooks500CA0[uint64]{
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classLow
		},
		loadCost: func(address uint32) int32 {
			w.observe(fmt.Sprintf("cost:%08x", address))
			return w.costs[address]
		},
	}
}

func newSummonManaCostTestWorld500CA0() *summonManaCostTestWorld500CA0 {
	return &summonManaCostTestWorld500CA0{
		classLow: 0x84,
		costs: map[uint32]int32{
			summonManaCostAddress500CA0(75): -123456789,
		},
	}
}

func TestSummonManaCost500CA0ExactOrderAndNativeWidthHandle(t *testing.T) {
	w := newSummonManaCostTestWorld500CA0()
	got := summonManaCost500CA0(int32(75), summonManaCostHighUnit500CA0, w.hooks())
	if want := int32(-123456789); got != want {
		t.Fatalf("cost = %d, want %d", got, want)
	}
	wantEvents := []string{
		"class:7ff71d949570",
		"cost:005bc370",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %q, want exact oracle order %q", w.events, wantEvents)
	}
}

func TestSummonManaCost500CA0ExactGates(t *testing.T) {
	tests := []struct {
		name       string
		unit       uint64
		classLow   uint8
		want       int32
		wantEvents []string
	}{
		{name: "null unit"},
		{
			name: "non-player", unit: summonManaCostHighUnit500CA0, classLow: 0xfb,
			wantEvents: []string{"class:7ff71d949570"},
		},
		{
			name: "player low bit", unit: summonManaCostHighUnit500CA0, classLow: 0x04,
			want:       -123456789,
			wantEvents: []string{"class:7ff71d949570", "cost:005bc370"},
		},
		{
			name: "player high bits", unit: summonManaCostHighUnit500CA0, classLow: 0xfc,
			want:       -123456789,
			wantEvents: []string{"class:7ff71d949570", "cost:005bc370"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSummonManaCostTestWorld500CA0()
			w.classLow = tc.classLow
			got := summonManaCost500CA0(int32(75), tc.unit, w.hooks())
			if got != tc.want {
				t.Fatalf("cost = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
		})
	}
}

func TestSummonManaCost500CA0OracleTable(t *testing.T) {
	w := newSummonManaCostTestWorld500CA0()
	w.costs = make(map[uint32]int32, len(summonManaCostOracleTable500CA0))
	for i, cost := range summonManaCostOracleTable500CA0 {
		spellID := int32(75 + i)
		address := uint32(0x005bc370 + 4*i)
		w.costs[address] = cost
		if gotAddress := summonManaCostAddress500CA0(spellID); gotAddress != address {
			t.Fatalf("spell %d address = %#08x, want %#08x", spellID, gotAddress, address)
		}
		if got := summonManaCost500CA0(spellID, summonManaCostHighUnit500CA0, w.hooks()); got != cost {
			t.Fatalf("spell %d cost = %d, want %d", spellID, got, cost)
		}
	}
}

func TestSummonManaCostAddress500CA0UsesPE32Wraparound(t *testing.T) {
	tests := []struct {
		spellID int32
		want    uint32
	}{
		{spellID: 0, want: 0x005bc244},
		{spellID: -1, want: 0x005bc240},
		{spellID: 1 << 30, want: 0x005bc244},
		{spellID: -1 << 31, want: 0x005bc244},
		{spellID: 1<<31 - 1, want: 0x005bc240},
	}
	for _, tc := range tests {
		if got := summonManaCostAddress500CA0(tc.spellID); got != tc.want {
			t.Errorf("spell ID %d address = %#08x, want %#08x", tc.spellID, got, tc.want)
		}
	}
}

func TestSummonManaCost500CA0AllFaultPrefixes(t *testing.T) {
	baseline := newSummonManaCostTestWorld500CA0()
	summonManaCost500CA0(int32(75), summonManaCostHighUnit500CA0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSummonManaCostTestWorld500CA0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				summonManaCost500CA0(int32(75), summonManaCostHighUnit500CA0, w.hooks())
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
