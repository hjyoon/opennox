package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type questInventoryTestOwner4F2C30 struct {
	class  uint32
	counts map[int32]int32
}

type questInventoryTestState4F2C30 struct {
	events       []string
	typeIDs      map[string]uint32
	balance      float32
	converted    int32
	lookupFault  string
	countFault   int32
	balanceFault bool
	convertFault bool
}

func newQuestInventoryTestState4F2C30() *questInventoryTestState4F2C30 {
	state := &questInventoryTestState4F2C30{
		typeIDs:    make(map[string]uint32),
		balance:    3.5,
		converted:  4,
		countFault: math.MinInt32,
	}
	for index, name := range questInventoryTypeNames4F2C30 {
		state.typeIDs[name] = uint32(index + 101)
	}
	return state
}

func (state *questInventoryTestState4F2C30) hooks() questInventoryLimitsHooks4F2C30[*questInventoryTestOwner4F2C30] {
	return questInventoryLimitsHooks4F2C30[*questInventoryTestOwner4F2C30]{
		objectTypeID: func(name string) uint32 {
			state.events = append(state.events, "type:"+name)
			if name == state.lookupFault {
				panic("type lookup fault")
			}
			return state.typeIDs[name]
		},
		isNil: func(owner *questInventoryTestOwner4F2C30) bool {
			return owner == nil
		},
		loadClass: func(owner *questInventoryTestOwner4F2C30) uint32 {
			state.events = append(state.events, "class")
			return owner.class
		},
		countInventory: func(owner *questInventoryTestOwner4F2C30, typeID int32) int32 {
			state.events = append(state.events, fmt.Sprintf("count:%08x", uint32(typeID)))
			if typeID == state.countFault {
				panic("inventory count fault")
			}
			return owner.counts[typeID]
		},
		loadBalance: func(key string) float32 {
			state.events = append(state.events, "balance:"+key)
			if state.balanceFault {
				panic("balance fault")
			}
			return state.balance
		},
		floatToInt: func(value float32) int32 {
			state.events = append(state.events, fmt.Sprintf("convert:%08x", math.Float32bits(value)))
			if state.convertFault {
				panic("conversion fault")
			}
			return state.converted
		},
	}
}

func questInventoryWarmCache4F2C30(state *questInventoryTestState4F2C30) questInventoryLimitsCache4F2C30 {
	var cache questInventoryLimitsCache4F2C30
	for index, name := range questInventoryTypeNames4F2C30 {
		cache.typeIDs[index] = state.typeIDs[name]
	}
	return cache
}

func TestQuestInventoryTypeCacheOrderAndNilBypass4F2C30(t *testing.T) {
	state := newQuestInventoryTestState4F2C30()
	var cache questInventoryLimitsCache4F2C30
	if got := questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks()); got != 1 {
		t.Fatalf("nil owner = %d, want 1", got)
	}
	wantEvents := make([]string, len(questInventoryTypeNames4F2C30))
	for index, name := range questInventoryTypeNames4F2C30 {
		wantEvents[index] = "type:" + name
		if cache.typeIDs[index] != state.typeIDs[name] {
			t.Fatalf("cache[%d] = %d, want %d", index, cache.typeIDs[index], state.typeIDs[name])
		}
	}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("cold nil events = %v, want %v", state.events, wantEvents)
	}
	state.events = nil
	if got := questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks()); got != 1 {
		t.Fatalf("warm nil owner = %d, want 1", got)
	}
	if len(state.events) != 0 {
		t.Fatalf("warm nil events = %v, want none", state.events)
	}
}

func TestQuestInventoryTypeCacheSentinelRetry4F2C30(t *testing.T) {
	state := newQuestInventoryTestState4F2C30()
	state.typeIDs[questInventoryTypeNames4F2C30[0]] = 0
	state.typeIDs[questInventoryTypeNames4F2C30[5]] = 0
	var cache questInventoryLimitsCache4F2C30
	for call := 0; call < 2; call++ {
		if got := questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks()); got != 1 {
			t.Fatalf("nil retry call %d = %d, want 1", call, got)
		}
	}
	if got, want := len(state.events), 2*len(questInventoryTypeNames4F2C30); got != want {
		t.Fatalf("lookup events = %d, want %d: %v", got, want, state.events)
	}

	state = newQuestInventoryTestState4F2C30()
	state.typeIDs[questInventoryTypeNames4F2C30[5]] = 0
	cache = questInventoryLimitsCache4F2C30{}
	questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks())
	state.events = nil
	state.typeIDs[questInventoryTypeNames4F2C30[5]] = 999
	questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks())
	if len(state.events) != 0 || cache.typeIDs[5] != 0 {
		t.Fatalf("non-sentinel zero was retried: events %v cache %v", state.events, cache.typeIDs)
	}
}

func TestQuestInventoryTypeCacheFaultPrefix4F2C30(t *testing.T) {
	for faultIndex, faultName := range questInventoryTypeNames4F2C30 {
		t.Run(faultName, func(t *testing.T) {
			state := newQuestInventoryTestState4F2C30()
			state.lookupFault = faultName
			var cache questInventoryLimitsCache4F2C30
			defer func() {
				if recover() == nil {
					t.Fatal("lookup did not fault")
				}
				if got, want := len(state.events), faultIndex+1; got != want {
					t.Fatalf("event prefix length = %d, want %d: %v", got, want, state.events)
				}
				for index := 0; index < faultIndex; index++ {
					if cache.typeIDs[index] != state.typeIDs[questInventoryTypeNames4F2C30[index]] {
						t.Fatalf("cache[%d] was not committed before fault", index)
					}
				}
				if cache.typeIDs[faultIndex] != 0 {
					t.Fatalf("faulting cache[%d] = %d, want 0", faultIndex, cache.typeIDs[faultIndex])
				}
			}()
			questInventoryLimits4F2C30[*questInventoryTestOwner4F2C30](nil, &cache, state.hooks())
		})
	}
}

func TestQuestInventoryNonPlayerBypassesCounts4F2C30(t *testing.T) {
	state := newQuestInventoryTestState4F2C30()
	cache := questInventoryWarmCache4F2C30(state)
	owner := &questInventoryTestOwner4F2C30{class: 0x100}
	if got := questInventoryLimits4F2C30(owner, &cache, state.hooks()); got != 1 {
		t.Fatalf("non-player = %d, want 1", got)
	}
	if want := []string{"class"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestQuestInventoryPotionAndStaffBoundaries4F2C30(t *testing.T) {
	for _, test := range []struct {
		name       string
		potionAt   int
		potion     int32
		staff      int32
		limit      int32
		want       int32
		wantCounts int
	}{
		{name: "all exact", potionAt: -1, potion: 9, staff: 3, limit: 3, want: 1, wantCounts: 13},
		{name: "first potion over", potionAt: 0, potion: 10, staff: 0, limit: 0, wantCounts: 1},
		{name: "middle potion over", potionAt: 5, potion: 10, staff: 0, limit: 0, wantCounts: 6},
		{name: "last potion over", potionAt: 11, potion: 10, staff: 0, limit: 0, wantCounts: 12},
		{name: "negative counts pass", potionAt: -1, potion: math.MinInt32, staff: -2, limit: -1, want: 1, wantCounts: 13},
		{name: "staff over", potionAt: -1, potion: 9, staff: 4, limit: 3, wantCounts: 13},
		{name: "negative limit", potionAt: -1, potion: 0, staff: 0, limit: -1, wantCounts: 13},
	} {
		t.Run(test.name, func(t *testing.T) {
			state := newQuestInventoryTestState4F2C30()
			state.converted = test.limit
			cache := questInventoryWarmCache4F2C30(state)
			owner := &questInventoryTestOwner4F2C30{class: questInventoryPlayerClass4F2C30, counts: make(map[int32]int32)}
			for index, typeID := range cache.typeIDs {
				value := test.potion
				if index == questInventoryPotionTypes4F2C30 {
					value = test.staff
				} else if test.potionAt >= 0 && index != test.potionAt {
					value = 0
				}
				owner.counts[int32(typeID)] = value
			}
			if got := questInventoryLimits4F2C30(owner, &cache, state.hooks()); got != test.want {
				t.Fatalf("result = %d, want %d; events %v", got, test.want, state.events)
			}
			counts := 0
			for _, event := range state.events {
				if len(event) >= len("count:") && event[:len("count:")] == "count:" {
					counts++
				}
			}
			if counts != test.wantCounts {
				t.Fatalf("count calls = %d, want %d; events %v", counts, test.wantCounts, state.events)
			}
			balanceCalls := test.wantCounts == len(questInventoryTypeNames4F2C30)
			wantBalanceEvents := 0
			if balanceCalls {
				wantBalanceEvents = 2
			}
			if got := len(state.events) - 1 - counts; got != wantBalanceEvents {
				t.Fatalf("non-class/count events = %d, want %d; events %v", got, wantBalanceEvents, state.events)
			}
		})
	}
}

func TestQuestInventoryPreservesFullCachedTypeBits4F2C30(t *testing.T) {
	state := newQuestInventoryTestState4F2C30()
	cache := questInventoryWarmCache4F2C30(state)
	cache.typeIDs[0] = 0x80000001
	cache.typeIDs[questInventoryPotionTypes4F2C30] = 0xffffffff
	owner := &questInventoryTestOwner4F2C30{class: 4, counts: map[int32]int32{-2147483647: 9, -1: 0}}
	state.converted = 0
	if got := questInventoryLimits4F2C30(owner, &cache, state.hooks()); got != 1 {
		t.Fatalf("high-bit cached IDs = %d, want 1; events %v", got, state.events)
	}
	if state.events[1] != "count:80000001" || state.events[len(state.events)-1] != "count:ffffffff" {
		t.Fatalf("cached type events = %v", state.events)
	}
}

func TestQuestInventoryBalanceAndConversionFaultOrder4F2C30(t *testing.T) {
	for _, test := range []struct {
		name         string
		balanceFault bool
		convertFault bool
		wantTail     []string
	}{
		{name: "balance", balanceFault: true, wantTail: []string{"balance:" + questInventoryBalanceKey4F2C30}},
		{name: "convert", convertFault: true, wantTail: []string{"balance:" + questInventoryBalanceKey4F2C30, "convert:40600000"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			state := newQuestInventoryTestState4F2C30()
			state.balanceFault = test.balanceFault
			state.convertFault = test.convertFault
			cache := questInventoryWarmCache4F2C30(state)
			owner := &questInventoryTestOwner4F2C30{class: 4, counts: make(map[int32]int32)}
			defer func() {
				if recover() == nil {
					t.Fatal("hook did not fault")
				}
				gotTail := state.events[len(state.events)-len(test.wantTail):]
				if !reflect.DeepEqual(gotTail, test.wantTail) {
					t.Fatalf("tail = %v, want %v; all events %v", gotTail, test.wantTail, state.events)
				}
				for _, event := range state.events {
					if event == fmt.Sprintf("count:%08x", cache.typeIDs[questInventoryPotionTypes4F2C30]) {
						t.Fatalf("staff count occurred after fault: %v", state.events)
					}
				}
			}()
			questInventoryLimits4F2C30(owner, &cache, state.hooks())
		})
	}
}
