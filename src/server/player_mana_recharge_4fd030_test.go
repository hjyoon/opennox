package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerManaRechargeTestWorld4FD030 struct {
	unit       uint64
	classLow   uint8
	amount     int16
	addResult  uint16
	events     []string
	faultAt    int
	after      map[string]func()
	addedUnit  uint64
	addedValue int16
}

func newPlayerManaRechargeTestWorld4FD030() *playerManaRechargeTestWorld4FD030 {
	return &playerManaRechargeTestWorld4FD030{
		unit:      0x100001234,
		classLow:  playerManaRechargePlayerClass4FD030,
		amount:    -1234,
		addResult: 0xbeef,
		after:     make(map[string]func()),
	}
}

func (w *playerManaRechargeTestWorld4FD030) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerManaRechargeTestWorld4FD030) hooks() playerManaRechargeHooks4FD030[uint64] {
	return playerManaRechargeHooks4FD030[uint64]{
		loadUnitArg: func() (uint64, uint16) {
			unit := w.unit
			w.record(fmt.Sprintf("unit=%016x", unit))
			return unit, uint16(unit)
		},
		loadClassLow: func(unit uint64) uint8 {
			classLow := w.classLow
			w.record(fmt.Sprintf("class:%016x=%02x", unit, classLow))
			return classLow
		},
		loadAmountArg: func() int16 {
			amount := w.amount
			w.record(fmt.Sprintf("amount=%d", amount))
			return amount
		},
		addMana: func(unit uint64, amount int16) uint16 {
			w.addedUnit, w.addedValue = unit, amount
			w.record(fmt.Sprintf("add:%016x:%d", unit, amount))
			return w.addResult
		},
	}
}

func playerManaRechargeFullTrace4FD030() (*playerManaRechargeTestWorld4FD030, []string) {
	w := newPlayerManaRechargeTestWorld4FD030()
	w.after["class:0000000100001234=04"] = func() {
		w.unit = 0x200005678
		w.amount = math.MaxInt16
	}
	w.after["amount=32767"] = func() {
		w.unit = 0x300009abc
		w.amount = math.MinInt16
	}
	return w, []string{
		"unit=0000000100001234",
		"class:0000000100001234=04",
		"amount=32767",
		"add:0000000100001234:32767",
	}
}

func TestPlayerManaRecharge4FD030ExactTraceAndCachedArguments(t *testing.T) {
	w, want := playerManaRechargeFullTrace4FD030()
	if got := playerManaRecharge4FD030(w.hooks()); got != 0xbeef {
		t.Fatalf("result = %#x, want 0xbeef", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.addedUnit != 0x100001234 || w.addedValue != math.MaxInt16 {
		t.Fatalf("add args = %#x/%d, want cached %#x/%d", w.addedUnit, w.addedValue, uint64(0x100001234), math.MaxInt16)
	}
}

func TestPlayerManaRecharge4FD030AllPlayerPathFaultPrefixes(t *testing.T) {
	_, want := playerManaRechargeFullTrace4FD030()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w, _ := playerManaRechargeFullTrace4FD030()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %q, want %q", w.events, want[:faultAt])
				}
			}()
			_ = playerManaRecharge4FD030(w.hooks())
		})
	}
}

func TestPlayerManaRecharge4FD030NonPlayerReturnsPointerLowWordLazily(t *testing.T) {
	w := newPlayerManaRechargeTestWorld4FD030()
	w.unit = 0xabcdef0187654321
	w.classLow = 0xfb
	w.after["class:abcdef0187654321=fb"] = func() {
		w.amount = math.MaxInt16
		w.addResult = 0
	}

	if got := playerManaRecharge4FD030(w.hooks()); got != 0x4321 {
		t.Fatalf("result = %#x, want original pointer low word 0x4321", got)
	}
	want := []string{"unit=abcdef0187654321", "class:abcdef0187654321=fb"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.addedUnit != 0 || w.addedValue != 0 {
		t.Fatalf("non-Player invoked add with %#x/%d", w.addedUnit, w.addedValue)
	}
}

func TestPlayerManaRecharge4FD030AmountAndResultWidths(t *testing.T) {
	for _, test := range []struct {
		name   string
		amount int16
		result uint16
	}{
		{name: "minimum signed amount", amount: math.MinInt16, result: 0},
		{name: "minus one and maximum result", amount: -1, result: math.MaxUint16},
		{name: "maximum signed amount", amount: math.MaxInt16, result: 0x8000},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerManaRechargeTestWorld4FD030()
			w.amount, w.addResult = test.amount, test.result
			if got := playerManaRecharge4FD030(w.hooks()); got != test.result {
				t.Fatalf("result = %#x, want %#x", got, test.result)
			}
			if w.addedValue != test.amount {
				t.Fatalf("amount = %d, want %d", w.addedValue, test.amount)
			}
		})
	}
}
