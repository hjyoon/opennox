package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type poisonProtectionTestObject4E0040 struct {
	name       string
	flags      uint32
	class      uint32
	first      *poisonProtectionTestObject4E0040
	next       *poisonProtectionTestObject4E0040
	init       *poisonProtectionTestInit4E0040
	buffResult int32
	buffPower  uint32
}

type poisonProtectionTestInit4E0040 struct {
	name  string
	slots [poisonProtectionModifierSlots]*poisonProtectionTestModifier4E0040
}

type poisonProtectionTestModifier4E0040 struct {
	name  string
	match bool
	value float32
}

type poisonProtectionTestWorld4E0040 struct {
	unit    *poisonProtectionTestObject4E0040
	balance float64
	events  []string
	after   map[string]func()
}

func poisonProtectionTestObjectName4E0040(obj *poisonProtectionTestObject4E0040) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func poisonProtectionTestInitName4E0040(init *poisonProtectionTestInit4E0040) string {
	if init == nil {
		return "nil"
	}
	return init.name
}

func poisonProtectionTestModifierName4E0040(modifier *poisonProtectionTestModifier4E0040) string {
	if modifier == nil {
		return "nil"
	}
	return modifier.name
}

func (w *poisonProtectionTestWorld4E0040) record(event string) {
	w.events = append(w.events, event)
	if fn := w.after[event]; fn != nil {
		fn()
	}
}

func (w *poisonProtectionTestWorld4E0040) hooks() poisonProtectionHooks4E0040[
	*poisonProtectionTestObject4E0040,
	*poisonProtectionTestInit4E0040,
	*poisonProtectionTestModifier4E0040,
] {
	return poisonProtectionHooks4E0040[
		*poisonProtectionTestObject4E0040,
		*poisonProtectionTestInit4E0040,
		*poisonProtectionTestModifier4E0040,
	]{
		loadUnitArg: func() *poisonProtectionTestObject4E0040 {
			unit := w.unit
			w.record("arg")
			return unit
		},
		loadFirstItem: func(unit *poisonProtectionTestObject4E0040) *poisonProtectionTestObject4E0040 {
			item := unit.first
			w.record("first:" + unit.name)
			return item
		},
		loadFlags: func(obj *poisonProtectionTestObject4E0040) uint32 {
			value := obj.flags
			w.record("flags:" + obj.name)
			return value
		},
		loadClass: func(obj *poisonProtectionTestObject4E0040) uint32 {
			value := obj.class
			w.record("class:" + obj.name)
			return value
		},
		loadInitData: func(obj *poisonProtectionTestObject4E0040) *poisonProtectionTestInit4E0040 {
			init := obj.init
			w.record("init:" + obj.name)
			return init
		},
		loadModifier: func(init *poisonProtectionTestInit4E0040, slot int) *poisonProtectionTestModifier4E0040 {
			name := poisonProtectionTestInitName4E0040(init)
			if init == nil {
				w.record(fmt.Sprintf("modifier:%s:%d", name, slot))
				panic("nil modifier init data")
			}
			modifier := init.slots[slot]
			w.record(fmt.Sprintf("modifier:%s:%d", name, slot))
			return modifier
		},
		matchesProtection: func(modifier *poisonProtectionTestModifier4E0040) bool {
			match := modifier.match
			w.record("match:" + modifier.name)
			return match
		},
		loadModifierValue: func(modifier *poisonProtectionTestModifier4E0040) float32 {
			value := modifier.value
			w.record("value:" + modifier.name)
			return value
		},
		loadNextItem: func(item *poisonProtectionTestObject4E0040) *poisonProtectionTestObject4E0040 {
			next := item.next
			w.record("next:" + item.name)
			return next
		},
		testBuff: func(unit *poisonProtectionTestObject4E0040, enchant uint32) int32 {
			value := unit.buffResult
			w.record(fmt.Sprintf("buff:%s:%d", unit.name, enchant))
			return value
		},
		loadBuffPower: func(unit *poisonProtectionTestObject4E0040, enchant uint32) uint32 {
			value := unit.buffPower
			w.record(fmt.Sprintf("power:%s:%d", unit.name, enchant))
			return value
		},
		loadBalance: func(key string, index int32) float64 {
			value := w.balance
			w.record(fmt.Sprintf("balance:%s:%d", key, index))
			return value
		},
	}
}

func TestPoisonProtection4E0040ExactTraversalOrder(t *testing.T) {
	matchingItem := &poisonProtectionTestModifier4E0040{name: "item-match", match: true, value: 0.1}
	nonmatchingItem := &poisonProtectionTestModifier4E0040{name: "item-other", value: 9}
	matchingUnit := &poisonProtectionTestModifier4E0040{name: "unit-match", match: true, value: 0.2}
	itemInit := &poisonProtectionTestInit4E0040{name: "item-init"}
	itemInit.slots[0] = matchingItem
	itemInit.slots[2] = nonmatchingItem
	unitInit := &poisonProtectionTestInit4E0040{name: "unit-init"}
	unitInit.slots[0] = matchingUnit

	second := &poisonProtectionTestObject4E0040{name: "second"}
	first := &poisonProtectionTestObject4E0040{
		name:  "first",
		flags: poisonProtectionEquippedFlag4E0040,
		class: poisonProtectionClassMask4E0040,
		init:  itemInit,
		next:  second,
	}
	unit := &poisonProtectionTestObject4E0040{
		name:       "unit",
		class:      poisonProtectionClassMask4E0040,
		first:      first,
		init:       unitInit,
		buffResult: -7,
		buffPower:  2,
	}
	world := &poisonProtectionTestWorld4E0040{unit: unit, balance: 0.25}
	got := poisonProtection4E0040(world.hooks())
	want := 0.25 + float64(float32(float64(float32(0.1))+float64(float32(0.2))))
	if got != want {
		t.Fatalf("result bits = %#016x, want %#016x", math.Float64bits(got), math.Float64bits(want))
	}
	wantEvents := []string{
		"arg", "first:unit",
		"flags:first", "class:first", "init:first",
		"modifier:item-init:0", "match:item-match", "value:item-match",
		"modifier:item-init:1",
		"modifier:item-init:2", "match:item-other",
		"modifier:item-init:3", "next:first",
		"flags:second", "next:second",
		"class:unit", "init:unit",
		"modifier:unit-init:0", "match:unit-match", "value:unit-match",
		"modifier:unit-init:1", "modifier:unit-init:2", "modifier:unit-init:3",
		"buff:unit:18", "power:unit:18", "balance:PoisonSpellProtection:1",
	}
	if !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("events = %q, want %q", world.events, wantEvents)
	}
}

func TestPoisonProtection4E0040TraversalUsesCachedLoadsAndLiveSuccessors(t *testing.T) {
	original := &poisonProtectionTestModifier4E0040{name: "original", match: true, value: 1}
	replacement := &poisonProtectionTestModifier4E0040{name: "replacement", match: false, value: 50}
	oldInit := &poisonProtectionTestInit4E0040{name: "old-init"}
	newInit := &poisonProtectionTestInit4E0040{name: "new-init"}
	newInit.slots[0] = original
	oldNext := &poisonProtectionTestObject4E0040{name: "old-next"}
	liveNext := &poisonProtectionTestObject4E0040{name: "live-next"}
	item := &poisonProtectionTestObject4E0040{
		name:  "item",
		flags: poisonProtectionEquippedFlag4E0040,
		class: poisonProtectionClassMask4E0040,
		init:  oldInit,
		next:  oldNext,
	}
	unit := &poisonProtectionTestObject4E0040{name: "unit", first: item}
	world := &poisonProtectionTestWorld4E0040{
		unit: unit,
		after: map[string]func(){
			"flags:item": func() {
				item.flags = 0
			},
			"class:item": func() {
				item.class = 0
				item.init = newInit
			},
			"modifier:new-init:0": func() {
				newInit.slots[0] = replacement
			},
			"match:original": func() {
				original.value = -2
				item.next = liveNext
			},
		},
	}
	if got := poisonProtection4E0040(world.hooks()); got != -2 {
		t.Fatalf("result = %v, want -2", got)
	}
	if containsPoisonProtectionEvent4E0040(world.events, "flags:old-next") {
		t.Fatalf("cached old successor was traversed: %q", world.events)
	}
	if !containsPoisonProtectionEvent4E0040(world.events, "flags:live-next") {
		t.Fatalf("live successor was not traversed: %q", world.events)
	}
}

func containsPoisonProtectionEvent4E0040(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func TestPoisonProtection4E0040NilUnitReturnsPositiveZero(t *testing.T) {
	world := &poisonProtectionTestWorld4E0040{}
	got := poisonProtection4E0040(world.hooks())
	if math.Float64bits(got) != 0 {
		t.Fatalf("nil result bits = %#016x, want positive zero", math.Float64bits(got))
	}
	if !reflect.DeepEqual(world.events, []string{"arg"}) {
		t.Fatalf("nil events = %q, want [arg]", world.events)
	}
}

func TestPoisonProtection4E0040NilInitFaultPrefix(t *testing.T) {
	item := &poisonProtectionTestObject4E0040{
		name:  "item",
		flags: poisonProtectionEquippedFlag4E0040,
		class: poisonProtectionClassMask4E0040,
	}
	world := &poisonProtectionTestWorld4E0040{
		unit: &poisonProtectionTestObject4E0040{name: "unit", first: item},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil modifier init data did not fault")
		}
		want := []string{"arg", "first:unit", "flags:item", "class:item", "init:item", "modifier:nil:0"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("fault events = %q, want %q", world.events, want)
		}
	}()
	poisonProtection4E0040(world.hooks())
}

func TestPoisonProtection4E0040NumericContract(t *testing.T) {
	modifierLimit := math.Float32frombits(poisonProtectionModifierLimitBits)
	finalLimit := math.Float32frombits(poisonProtectionFinalLimitBits)
	tests := []struct {
		name       string
		inventory  []float32
		unit       []float32
		buff       int32
		power      uint32
		balance    float64
		want       float64
		wantNaN    bool
		wantIndex  int32
		wantLookup bool
	}{
		{name: "zero", want: 0},
		{
			name: "single binary32 spill",
			unit: []float32{0.1, 0.2},
			want: float64(float32(float64(float32(0.1)) + float64(float32(0.2)))),
		},
		{name: "modifier upper clamp", unit: []float32{0.5, 0.3}, want: float64(modifierLimit)},
		{
			name:      "inventory spill does not replace accumulator",
			inventory: []float32{-16777216, -1},
			unit:      []float32{16777216},
			want:      -1,
		},
		{
			name: "final upper clamp",
			unit: []float32{0.6}, buff: 1, power: 3, balance: 0.4,
			want: float64(finalLimit), wantIndex: 2, wantLookup: true,
		},
		{
			name: "zero power wraps index",
			buff: 1, power: 0, balance: -0.25,
			want: -0.25, wantIndex: -1, wantLookup: true,
		},
		{
			name: "power uses low byte",
			buff: -1, power: 0x102, balance: 0.125,
			want: 0.125, wantIndex: 1, wantLookup: true,
		},
		{name: "positive infinity modifier clamps", unit: []float32{float32(math.Inf(1))}, want: float64(modifierLimit)},
		{name: "negative infinity remains", unit: []float32{float32(math.Inf(-1))}, want: math.Inf(-1)},
		{name: "unordered modifier remains", unit: []float32{float32(math.NaN())}, wantNaN: true},
		{name: "unordered spell result remains", buff: 1, power: 1, balance: math.NaN(), wantNaN: true, wantIndex: 0, wantLookup: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unitInit := &poisonProtectionTestInit4E0040{name: "unit-init"}
			for i, value := range test.unit {
				unitInit.slots[i] = &poisonProtectionTestModifier4E0040{name: fmt.Sprintf("unit-%d", i), match: true, value: value}
			}
			unit := &poisonProtectionTestObject4E0040{
				name:       "unit",
				buffResult: test.buff,
				buffPower:  test.power,
			}
			if len(test.unit) != 0 {
				unit.class = poisonProtectionClassMask4E0040
				unit.init = unitInit
			}
			if len(test.inventory) != 0 {
				itemInit := &poisonProtectionTestInit4E0040{name: "item-init"}
				for i, value := range test.inventory {
					itemInit.slots[i] = &poisonProtectionTestModifier4E0040{name: fmt.Sprintf("item-%d", i), match: true, value: value}
				}
				unit.first = &poisonProtectionTestObject4E0040{
					name:  "item",
					flags: poisonProtectionEquippedFlag4E0040,
					class: poisonProtectionClassMask4E0040,
					init:  itemInit,
				}
			}
			world := &poisonProtectionTestWorld4E0040{unit: unit, balance: test.balance}
			got := poisonProtection4E0040(world.hooks())
			if test.wantNaN {
				if !math.IsNaN(got) {
					t.Fatalf("result = %v, want NaN", got)
				}
			} else if math.Float64bits(got) != math.Float64bits(test.want) {
				t.Fatalf("result bits = %#016x, want %#016x", math.Float64bits(got), math.Float64bits(test.want))
			}
			balancePrefix := "balance:" + poisonProtectionBalanceKey4E0040 + ":"
			lookupFound := false
			for _, event := range world.events {
				if len(event) >= len(balancePrefix) && event[:len(balancePrefix)] == balancePrefix {
					lookupFound = true
					if event != fmt.Sprintf("%s%d", balancePrefix, test.wantIndex) {
						t.Fatalf("balance event = %q, want index %d", event, test.wantIndex)
					}
				}
			}
			if lookupFound != test.wantLookup {
				t.Fatalf("balance lookup = %v, want %v; events %q", lookupFound, test.wantLookup, world.events)
			}
		})
	}
}

func TestPoisonProtection4E0040GatesDeferDependentLoads(t *testing.T) {
	item := &poisonProtectionTestObject4E0040{
		name:  "item",
		flags: 0,
		class: poisonProtectionClassMask4E0040,
	}
	unit := &poisonProtectionTestObject4E0040{name: "unit", first: item, buffResult: 0, buffPower: 99}
	world := &poisonProtectionTestWorld4E0040{unit: unit}
	if got := poisonProtection4E0040(world.hooks()); got != 0 {
		t.Fatalf("result = %v, want 0", got)
	}
	want := []string{"arg", "first:unit", "flags:item", "next:item", "class:unit", "buff:unit:18"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %q, want %q", world.events, want)
	}
}
