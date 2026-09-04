package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type spellManaChargeCostKey4FCF90 struct {
	spellID  int32
	costType int32
}

type spellManaChargeTestUpdate4FCF90 struct {
	mana          uint16
	rechargeCost  uint16
	rechargeFrame uint16
}

type spellManaChargeTestWorld4FCF90 struct {
	unit          uint64
	classLow      uint8
	update        uint64
	spellID       int32
	costType      int32
	godMode       bool
	tickRate      uint32
	updates       map[uint64]*spellManaChargeTestUpdate4FCF90
	summonCosts   map[int32]int32
	ordinaryCosts map[spellManaChargeCostKey4FCF90]int32
	events        []string
	faultAt       int
	after         map[string]func()
}

func newSpellManaChargeTestWorld4FCF90() *spellManaChargeTestWorld4FCF90 {
	const update = uint64(0x200005678)
	return &spellManaChargeTestWorld4FCF90{
		unit:          0x100001234,
		classLow:      spellManaChargePlayerClass4FCF90,
		update:        update,
		spellID:       74,
		costType:      2,
		tickRate:      30,
		updates:       map[uint64]*spellManaChargeTestUpdate4FCF90{update: {mana: 100}},
		summonCosts:   make(map[int32]int32),
		ordinaryCosts: make(map[spellManaChargeCostKey4FCF90]int32),
		after:         make(map[string]func()),
	}
}

func (w *spellManaChargeTestWorld4FCF90) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellManaChargeTestWorld4FCF90) updateState(update uint64) *spellManaChargeTestUpdate4FCF90 {
	state := w.updates[update]
	if state == nil {
		state = new(spellManaChargeTestUpdate4FCF90)
		w.updates[update] = state
	}
	return state
}

func (w *spellManaChargeTestWorld4FCF90) hooks() spellManaChargeHooks4FCF90[uint64, uint64] {
	return spellManaChargeHooks4FCF90[uint64, uint64]{
		loadUnitArg: func() uint64 {
			value := w.unit
			w.record(fmt.Sprintf("unit=%016x", value))
			return value
		},
		loadClassLow: func(unit uint64) uint8 {
			value := w.classLow
			w.record(fmt.Sprintf("class:%016x=%02x", unit, value))
			return value
		},
		loadUpdateData: func(unit uint64) uint64 {
			value := w.update
			w.record(fmt.Sprintf("update:%016x=%016x", unit, value))
			return value
		},
		loadSpellArg: func() int32 {
			value := w.spellID
			w.record(fmt.Sprintf("spell=%d", value))
			return value
		},
		loadGodMode: func() bool {
			value := w.godMode
			w.record(fmt.Sprintf("god=%t", value))
			return value
		},
		loadCostTypeArg: func() int32 {
			value := w.costType
			w.record(fmt.Sprintf("cost-type=%d", value))
			return value
		},
		summonCost: func(spellID int32, unit uint64) int32 {
			value := w.summonCosts[spellID]
			w.record(fmt.Sprintf("summon:%d:%016x=%d", spellID, unit, value))
			return value
		},
		spellManaCost: func(spellID, costType int32) int32 {
			value := w.ordinaryCosts[spellManaChargeCostKey4FCF90{spellID: spellID, costType: costType}]
			w.record(fmt.Sprintf("ordinary:%d:%d=%d", spellID, costType, value))
			return value
		},
		loadCurrentMana: func(update uint64) uint16 {
			value := w.updateState(update).mana
			w.record(fmt.Sprintf("mana:%016x=%d", update, value))
			return value
		},
		subtractMana: func(unit uint64, cost int32) {
			w.record(fmt.Sprintf("subtract:%016x:%d", unit, cost))
		},
		storeRechargeCost: func(update uint64, value uint16) {
			w.record(fmt.Sprintf("store-cost:%016x=%d", update, value))
			w.updateState(update).rechargeCost = value
		},
		loadTickRate: func() uint32 {
			value := w.tickRate
			w.record(fmt.Sprintf("tick=%d", value))
			return value
		},
		storeRechargeFrame: func(update uint64, value uint16) {
			w.record(fmt.Sprintf("store-frame:%016x=%d", update, value))
			w.updateState(update).rechargeFrame = value
		},
	}
}

func TestSpellManaCharge4FCF90EntryOrderAndGates(t *testing.T) {
	t.Run("non-Player still caches update after class load", func(t *testing.T) {
		w := newSpellManaChargeTestWorld4FCF90()
		w.classLow = 0x82
		w.after["class:0000000100001234=82"] = func() { w.update = 0x300009abc }

		if got := spellManaCharge4FCF90(w.hooks()); got != -1 {
			t.Fatalf("result = %d, want canonical -1", got)
		}
		want := []string{
			"unit=0000000100001234",
			"class:0000000100001234=82",
			"update:0000000100001234=0000000300009abc",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	tests := []struct {
		name   string
		mutate func(*spellManaChargeTestWorld4FCF90)
		want   int32
		events []string
	}{
		{
			name:   "zero spell",
			mutate: func(w *spellManaChargeTestWorld4FCF90) { w.spellID = 0 },
			want:   -1,
			events: []string{
				"unit=0000000100001234", "class:0000000100001234=04",
				"update:0000000100001234=0000000200005678", "spell=0",
			},
		},
		{
			name:   "GodMode",
			mutate: func(w *spellManaChargeTestWorld4FCF90) { w.godMode = true },
			want:   0,
			events: []string{
				"unit=0000000100001234", "class:0000000100001234=04",
				"update:0000000100001234=0000000200005678", "spell=74", "god=true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newSpellManaChargeTestWorld4FCF90()
			test.mutate(w)
			if got := spellManaCharge4FCF90(w.hooks()); got != test.want {
				t.Fatalf("result = %d, want %d", got, test.want)
			}
			if !reflect.DeepEqual(w.events, test.events) {
				t.Fatalf("events = %q, want %q", w.events, test.events)
			}
		})
	}
}

func TestSpellManaCharge4FCF90CostDispatchBoundaries(t *testing.T) {
	tests := []struct {
		spellID int32
		events  []string
	}{
		{74, []string{"cost-type=-2147483648", "ordinary:74:-2147483648=7"}},
		{75, []string{"summon:75:0000000100001234=7"}},
		{114, []string{"summon:114:0000000100001234=7"}},
		{115, []string{"cost-type=-2147483648", "ordinary:115:-2147483648=7"}},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("spell-%d", test.spellID), func(t *testing.T) {
			w := newSpellManaChargeTestWorld4FCF90()
			w.spellID = test.spellID
			w.costType = math.MinInt32
			w.summonCosts[test.spellID] = 7
			w.ordinaryCosts[spellManaChargeCostKey4FCF90{test.spellID, math.MinInt32}] = 7

			if got := spellManaCharge4FCF90(w.hooks()); got != 7 {
				t.Fatalf("result = %d, want 7", got)
			}
			wantTail := append(test.events, "mana:0000000200005678=100", "subtract:0000000100001234:7")
			if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("event tail = %q, want %q", got, wantTail)
			}
		})
	}
}

func TestSpellManaCharge4FCF90CachesArgumentsAndUpdateAcrossCallbacks(t *testing.T) {
	w := newSpellManaChargeTestWorld4FCF90()
	w.ordinaryCosts[spellManaChargeCostKey4FCF90{74, 2}] = 10
	w.after["ordinary:74:2=10"] = func() {
		w.unit = 0x40000def0
		w.update = 0x300009abc
		w.spellID = 115
		w.costType = 9
	}

	if got := spellManaCharge4FCF90(w.hooks()); got != 10 {
		t.Fatalf("result = %d, want cached cost 10", got)
	}
	wantTail := []string{
		"ordinary:74:2=10",
		"mana:0000000200005678=100",
		"subtract:0000000100001234:10",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want cached arguments %q", got, wantTail)
	}
}

func TestSpellManaCharge4FCF90SignedNegativeCostStillSubtracts(t *testing.T) {
	w := newSpellManaChargeTestWorld4FCF90()
	w.updateState(w.update).mana = 0
	w.ordinaryCosts[spellManaChargeCostKey4FCF90{74, 2}] = math.MinInt32

	if got := spellManaCharge4FCF90(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want INT32_MIN", got)
	}
	wantTail := []string{
		"ordinary:74:2=-2147483648",
		"mana:0000000200005678=0",
		"subtract:0000000100001234:-2147483648",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestSpellManaCharge4FCF90InsufficientSummonUsesOrdinaryRechargeAndLiveTick(t *testing.T) {
	w := newSpellManaChargeTestWorld4FCF90()
	w.spellID = 75
	w.updateState(w.update).mana = 9
	w.summonCosts[75] = 10
	w.ordinaryCosts[spellManaChargeCostKey4FCF90{75, 1}] = 0x12345
	w.after["summon:75:0000000100001234=10"] = func() {
		w.spellID = 74
		w.update = 0x300009abc
	}
	w.after["store-cost:0000000200005678=9029"] = func() {
		w.tickRate = 0xffff8001
	}

	if got := spellManaCharge4FCF90(w.hooks()); got != -1 {
		t.Fatalf("result = %d, want canonical -1", got)
	}
	want := []string{
		"unit=0000000100001234", "class:0000000100001234=04",
		"update:0000000100001234=0000000200005678", "spell=75", "god=false",
		"summon:75:0000000100001234=10", "mana:0000000200005678=9",
		"ordinary:75:1=74565", "store-cost:0000000200005678=9029",
		"tick=4294934529", "store-frame:0000000200005678=32769",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	state := w.updateState(0x200005678)
	if state.rechargeCost != 0x2345 || state.rechargeFrame != 0x8001 {
		t.Fatalf("recharge cost/frame = %#x/%#x, want 0x2345/0x8001", state.rechargeCost, state.rechargeFrame)
	}
	if other := w.updates[0x300009abc]; other != nil && (other.rechargeCost != 0 || other.rechargeFrame != 0) {
		t.Fatalf("live replacement update changed: %#v", other)
	}
}

func TestSpellManaCharge4FCF90AllFaultPrefixes(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*spellManaChargeTestWorld4FCF90)
		events []string
	}{
		{
			name: "enough",
			setup: func(w *spellManaChargeTestWorld4FCF90) {
				w.ordinaryCosts[spellManaChargeCostKey4FCF90{74, 2}] = 10
			},
			events: []string{
				"unit=0000000100001234", "class:0000000100001234=04",
				"update:0000000100001234=0000000200005678", "spell=74", "god=false",
				"cost-type=2", "ordinary:74:2=10", "mana:0000000200005678=100",
				"subtract:0000000100001234:10",
			},
		},
		{
			name: "insufficient",
			setup: func(w *spellManaChargeTestWorld4FCF90) {
				w.ordinaryCosts[spellManaChargeCostKey4FCF90{74, 2}] = 101
				w.ordinaryCosts[spellManaChargeCostKey4FCF90{74, 1}] = 2
				w.tickRate = 3
			},
			events: []string{
				"unit=0000000100001234", "class:0000000100001234=04",
				"update:0000000100001234=0000000200005678", "spell=74", "god=false",
				"cost-type=2", "ordinary:74:2=101", "mana:0000000200005678=100",
				"ordinary:74:1=2", "store-cost:0000000200005678=2", "tick=3",
				"store-frame:0000000200005678=3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for faultAt := 1; faultAt <= len(test.events); faultAt++ {
				t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
					w := newSpellManaChargeTestWorld4FCF90()
					test.setup(w)
					w.faultAt = faultAt

					var recovered any
					func() {
						defer func() { recovered = recover() }()
						_ = spellManaCharge4FCF90(w.hooks())
					}()
					if recovered != test.events[faultAt-1] {
						t.Fatalf("recovered = %#v, want %q", recovered, test.events[faultAt-1])
					}
					if want := test.events[:faultAt]; !reflect.DeepEqual(w.events, want) {
						t.Fatalf("events = %q, want prefix %q", w.events, want)
					}
				})
			}
		})
	}
}
