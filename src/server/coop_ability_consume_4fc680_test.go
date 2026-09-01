package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type coopAbilityExecution4FC680 struct {
	unit  uintptr
	state int32
}

type coopAbilityConsumeTestWorld4FC680 struct {
	events     []string
	faultAt    int
	coopFlag   int32
	flag20     int32
	state      int32
	stateLoads int
	unit       uintptr
	executions []coopAbilityExecution4FC680
	onFindUnit func()
	onExecute  func()
}

func (w *coopAbilityConsumeTestWorld4FC680) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *coopAbilityConsumeTestWorld4FC680) hooks() coopAbilityConsumeHooks4FC680[uintptr] {
	return coopAbilityConsumeHooks4FC680[uintptr]{
		loadCoopFlag: func() int32 {
			value := w.coopFlag
			w.record(fmt.Sprintf("coop:%#08x", uint32(value)))
			return value
		},
		loadFlag20: func() int32 {
			value := w.flag20
			w.record(fmt.Sprintf("flag20:%#08x", uint32(value)))
			return value
		},
		loadState: func() int32 {
			w.stateLoads++
			value := w.state
			w.record(fmt.Sprintf("state%d:%#08x", w.stateLoads, uint32(value)))
			return value
		},
		firstPlayerUnit: func() uintptr {
			unit := w.unit
			w.record(fmt.Sprintf("unit:%#x", unit))
			if w.onFindUnit != nil {
				w.onFindUnit()
			}
			return unit
		},
		executeAbility: func(unit uintptr, state int32) {
			w.record(fmt.Sprintf("execute:%#x:%#08x", unit, uint32(state)))
			w.executions = append(w.executions, coopAbilityExecution4FC680{unit: unit, state: state})
			if w.onExecute != nil {
				w.onExecute()
			}
		},
		storeState: func(value int32) {
			w.record(fmt.Sprintf("store:%#08x", uint32(value)))
			w.state = value
		},
	}
}

func TestCoopAbilityConsume4FC680ExactFlagAndStateGates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		coopFlag   int32
		flag20     int32
		state      int32
		unit       uintptr
		wantState  int32
		wantEvents []string
	}{
		{
			name:       "coop zero",
			state:      -1985229329,
			unit:       0x1234,
			wantState:  -1985229329,
			wantEvents: []string{"coop:0x00000000"},
		},
		{
			name:       "flag20 exact one",
			coopFlag:   1,
			flag20:     1,
			state:      -1985229329,
			unit:       0x1234,
			wantState:  -1985229329,
			wantEvents: []string{"coop:0x00000001", "flag20:0x00000001"},
		},
		{
			name:       "state zero",
			coopFlag:   1,
			wantEvents: []string{"coop:0x00000001", "flag20:0x00000000", "state1:0x00000000"},
		},
		{
			name:       "no player unit",
			coopFlag:   1,
			state:      -1985229329,
			wantState:  -1985229329,
			wantEvents: []string{"coop:0x00000001", "flag20:0x00000000", "state1:0x89abcdef", "unit:0x0"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := coopAbilityConsumeTestWorld4FC680{
				coopFlag: tc.coopFlag,
				flag20:   tc.flag20,
				state:    tc.state,
				unit:     tc.unit,
			}
			coopAbilityConsume4FC680(w.hooks())
			if uint32(w.state) != uint32(tc.wantState) {
				t.Fatalf("state = %#08x, want %#08x", uint32(w.state), uint32(tc.wantState))
			}
			if len(w.executions) != 0 {
				t.Fatalf("executions = %#v, want none", w.executions)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %#v, want %#v", w.events, tc.wantEvents)
			}
		})
	}
}

func TestCoopAbilityConsume4FC680UsesPE32FlagComparisons(t *testing.T) {
	for _, tc := range []struct {
		name     string
		coopFlag int32
		flag20   int32
		wantRun  bool
	}{
		{name: "coop min", coopFlag: math.MinInt32, wantRun: true},
		{name: "coop negative one", coopFlag: -1, wantRun: true},
		{name: "coop positive two", coopFlag: 2, wantRun: true},
		{name: "flag20 zero", coopFlag: 1, wantRun: true},
		{name: "flag20 negative one", coopFlag: 1, flag20: -1, wantRun: true},
		{name: "flag20 positive two", coopFlag: 1, flag20: 2, wantRun: true},
		{name: "flag20 max", coopFlag: 1, flag20: math.MaxInt32, wantRun: true},
		{name: "flag20 exact one", coopFlag: 1, flag20: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := coopAbilityConsumeTestWorld4FC680{
				coopFlag: tc.coopFlag,
				flag20:   tc.flag20,
				state:    math.MinInt32,
				unit:     0x1234,
			}
			coopAbilityConsume4FC680(w.hooks())
			if got := len(w.executions) == 1; got != tc.wantRun {
				t.Fatalf("executed = %t, want %t; events = %#v", got, tc.wantRun, w.events)
			}
			if tc.wantRun && w.state != 0 {
				t.Fatalf("state = %#08x, want cleared", uint32(w.state))
			}
			if !tc.wantRun && w.state != math.MinInt32 {
				t.Fatalf("state = %#08x, want unchanged", uint32(w.state))
			}
		})
	}
}

func TestCoopAbilityConsume4FC680ReloadsStateAfterUnitLookup(t *testing.T) {
	for _, reloaded := range []int32{0, math.MinInt32, -1, math.MaxInt32, -1985229329} {
		t.Run(fmt.Sprintf("%08x", uint32(reloaded)), func(t *testing.T) {
			w := coopAbilityConsumeTestWorld4FC680{
				coopFlag: 1,
				state:    1,
				unit:     0x5678,
			}
			w.onFindUnit = func() { w.state = reloaded }
			coopAbilityConsume4FC680(w.hooks())

			wantExecution := []coopAbilityExecution4FC680{{unit: 0x5678, state: reloaded}}
			if !reflect.DeepEqual(w.executions, wantExecution) {
				t.Fatalf("executions = %#v, want %#v", w.executions, wantExecution)
			}
			if w.state != 0 {
				t.Fatalf("state = %#08x, want cleared", uint32(w.state))
			}
			wantEvents := []string{
				"coop:0x00000001",
				"flag20:0x00000000",
				"state1:0x00000001",
				"unit:0x5678",
				fmt.Sprintf("state2:%#08x", uint32(reloaded)),
				fmt.Sprintf("execute:0x5678:%#08x", uint32(reloaded)),
				"store:0x00000000",
			}
			if !reflect.DeepEqual(w.events, wantEvents) {
				t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
			}
		})
	}
}

func TestCoopAbilityConsume4FC680ClearsCallbackMutationAfterReturn(t *testing.T) {
	w := coopAbilityConsumeTestWorld4FC680{
		coopFlag: 1,
		state:    -1985229329,
		unit:     0x1234,
	}
	w.onExecute = func() { w.state = math.MaxInt32 }
	coopAbilityConsume4FC680(w.hooks())
	if w.state != 0 {
		t.Fatalf("state = %#08x, want cleared after callback", uint32(w.state))
	}
}

func TestCoopAbilityConsume4FC680FaultPrefixesPreserveState(t *testing.T) {
	wantEvents := []string{
		"coop:0x00000001",
		"flag20:0x00000000",
		"state1:0x89abcdef",
		"unit:0x1234",
		"state2:0x89abcdef",
		"execute:0x1234:0x89abcdef",
		"store:0x00000000",
	}
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := coopAbilityConsumeTestWorld4FC680{
				faultAt:  faultAt,
				coopFlag: 1,
				state:    -1985229329,
				unit:     0x1234,
			}
			defer func() {
				if got := recover(); got != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
				if uint32(w.state) != 0x89abcdef {
					t.Fatalf("state = %#08x, want unchanged", uint32(w.state))
				}
			}()
			coopAbilityConsume4FC680(w.hooks())
		})
	}
}
