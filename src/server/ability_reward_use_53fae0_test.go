package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type abilityRewardUseTestData53FAE0 struct {
	ability uint8
}

type abilityRewardUseTestPlayer53FAE0 struct {
	name   string
	class  uint8
	levels [6]uint32
}

type abilityRewardUseTestUpdate53FAE0 struct {
	player *abilityRewardUseTestPlayer53FAE0
}

type abilityRewardUseTestObject53FAE0 struct {
	name    string
	class   uint8
	netCode uint32
	update  *abilityRewardUseTestUpdate53FAE0
	data    *abilityRewardUseTestData53FAE0
}

type abilityRewardUseTestWorld53FAE0 struct {
	owner        *abilityRewardUseTestObject53FAE0
	item         *abilityRewardUseTestObject53FAE0
	flags        int32
	rewardResult int32
	events       []string
	faultAt      int
	afterAbility func(int)
	abilityLoads int
	afterFlags   func()
	observedArgs [2]int32
}

func abilityRewardUseTestName53FAE0(obj *abilityRewardUseTestObject53FAE0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *abilityRewardUseTestWorld53FAE0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *abilityRewardUseTestWorld53FAE0) hooks() abilityRewardUseHooks53FAE0[
	*abilityRewardUseTestObject53FAE0,
	*abilityRewardUseTestUpdate53FAE0,
	*abilityRewardUseTestPlayer53FAE0,
	*abilityRewardUseTestData53FAE0,
] {
	return abilityRewardUseHooks53FAE0[
		*abilityRewardUseTestObject53FAE0,
		*abilityRewardUseTestUpdate53FAE0,
		*abilityRewardUseTestPlayer53FAE0,
		*abilityRewardUseTestData53FAE0,
	]{
		loadItemArg: func() *abilityRewardUseTestObject53FAE0 {
			w.event("item:" + abilityRewardUseTestName53FAE0(w.item))
			return w.item
		},
		loadOwnerArg: func() *abilityRewardUseTestObject53FAE0 {
			w.event("owner:" + abilityRewardUseTestName53FAE0(w.owner))
			return w.owner
		},
		loadUseData: func(item *abilityRewardUseTestObject53FAE0) *abilityRewardUseTestData53FAE0 {
			w.event("data:" + abilityRewardUseTestName53FAE0(item))
			return item.data
		},
		loadClassLow: func(owner *abilityRewardUseTestObject53FAE0) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", abilityRewardUseTestName53FAE0(owner), owner.class))
			return owner.class
		},
		loadUpdateData: func(owner *abilityRewardUseTestObject53FAE0) *abilityRewardUseTestUpdate53FAE0 {
			w.event("update:" + abilityRewardUseTestName53FAE0(owner))
			return owner.update
		},
		loadPlayer: func(update *abilityRewardUseTestUpdate53FAE0) *abilityRewardUseTestPlayer53FAE0 {
			w.event("player:" + update.player.name)
			return update.player
		},
		loadPlayerClass: func(player *abilityRewardUseTestPlayer53FAE0) uint8 {
			w.event(fmt.Sprintf("player-class:%s=%02x", player.name, player.class))
			return player.class
		},
		primaryMessage: func(owner *abilityRewardUseTestObject53FAE0, message string, value uint8) {
			w.event(fmt.Sprintf("primary:%s:%s:%d", abilityRewardUseTestName53FAE0(owner), message, value))
		},
		loadNetCode: func(owner *abilityRewardUseTestObject53FAE0) uint32 {
			w.event(fmt.Sprintf("netcode:%s=%08x", abilityRewardUseTestName53FAE0(owner), owner.netCode))
			return owner.netCode
		},
		audit: func(sound int32, owner *abilityRewardUseTestObject53FAE0, kind int32, code uint32) {
			w.event(fmt.Sprintf("audit:%d:%s:%d:%08x", sound, abilityRewardUseTestName53FAE0(owner), kind, code))
		},
		gameFlagsCheck: func(mask uint32) int32 {
			w.event(fmt.Sprintf("flags:%08x=%08x", mask, uint32(w.flags)))
			if w.afterFlags != nil {
				w.afterFlags()
			}
			return w.flags
		},
		loadAbility: func(data *abilityRewardUseTestData53FAE0) uint8 {
			value := data.ability
			w.abilityLoads++
			w.event(fmt.Sprintf("ability:%d=%02x", w.abilityLoads, value))
			if w.afterAbility != nil {
				w.afterAbility(w.abilityLoads)
			}
			return value
		},
		loadAbilityLevel: func(player *abilityRewardUseTestPlayer53FAE0, ability int32) uint32 {
			value := player.levels[ability]
			w.event(fmt.Sprintf("level:%s:%d=%08x", player.name, ability, value))
			return value
		},
		rewardAbility: func(owner *abilityRewardUseTestObject53FAE0, ability, rewardArg int32) int32 {
			w.observedArgs = [2]int32{ability, rewardArg}
			w.event(fmt.Sprintf("reward:%s:%d:%d=%08x", abilityRewardUseTestName53FAE0(owner), ability, rewardArg, uint32(w.rewardResult)))
			return w.rewardResult
		},
		delayedDelete: func(item *abilityRewardUseTestObject53FAE0) {
			w.event("delete:" + abilityRewardUseTestName53FAE0(item))
		},
	}
}

func newAbilityRewardUseTestWorld53FAE0() *abilityRewardUseTestWorld53FAE0 {
	player := &abilityRewardUseTestPlayer53FAE0{name: "player"}
	owner := &abilityRewardUseTestObject53FAE0{
		name:    "owner",
		class:   0xf4,
		netCode: 0x89abcdef,
		update:  &abilityRewardUseTestUpdate53FAE0{player: player},
	}
	item := &abilityRewardUseTestObject53FAE0{
		name: "item",
		data: &abilityRewardUseTestData53FAE0{ability: 2},
	}
	return &abilityRewardUseTestWorld53FAE0{
		owner:        owner,
		item:         item,
		flags:        math.MinInt32,
		rewardResult: math.MinInt32,
	}
}

func abilityRewardUseSuccessTrace53FAE0() []string {
	return []string{
		"item:item",
		"owner:owner",
		"data:item",
		"class:owner=f4",
		"update:owner",
		"player:player",
		"player-class:player=00",
		"flags:00001800=80000000",
		"player:player",
		"ability:1=02",
		"level:player:2=00000000",
		"ability:2=05",
		"reward:owner:5:1=80000000",
		"delete:item",
	}
}

func TestUseAbilityReward53FAE0ExactSuccessTraceLiveAbilityAndFaultPrefixes(t *testing.T) {
	want := abilityRewardUseSuccessTrace53FAE0()
	build := func() *abilityRewardUseTestWorld53FAE0 {
		w := newAbilityRewardUseTestWorld53FAE0()
		w.afterAbility = func(load int) {
			if load == 1 {
				w.item.data.ability = 5
			}
		}
		return w
	}

	w := build()
	if got := useAbilityReward53FAE0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.observedArgs != [2]int32{5, 1} {
		t.Fatalf("reward args = %v, want [5 1]", w.observedArgs)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			useAbilityReward53FAE0(w.hooks())
		})
	}
}

func TestUseAbilityReward53FAE0LoadsUpdateBeforeNonPlayerReturn(t *testing.T) {
	w := newAbilityRewardUseTestWorld53FAE0()
	w.owner.class = 0xf0
	if got := useAbilityReward53FAE0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"item:item", "owner:owner", "data:item", "class:owner=f0", "update:owner",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestUseAbilityReward53FAE0ClassFailureMessagesThenAuditsLiveNetCode(t *testing.T) {
	w := newAbilityRewardUseTestWorld53FAE0()
	w.owner.update.player.class = 0xff
	w.owner.netCode = 0x10203040
	if got := useAbilityReward53FAE0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantTail := []string{
		"player:player",
		"player-class:player=ff",
		"primary:owner:pickup.c:ObjectEquipClassFail:0",
		"netcode:owner=10203040",
		"audit:925:owner:2:10203040",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}

func TestUseAbilityReward53FAE0FlagAndServiceBranches(t *testing.T) {
	t.Run("zero-flags-skips-level-and-audits-zero-result", func(t *testing.T) {
		w := newAbilityRewardUseTestWorld53FAE0()
		w.flags = 0
		w.rewardResult = 0
		w.owner.netCode = math.MaxUint32
		if got := useAbilityReward53FAE0(w.hooks()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.observedArgs != [2]int32{2, 0} {
			t.Fatalf("reward args = %v, want [2 0]", w.observedArgs)
		}
		wantTail := []string{
			"flags:00001800=00000000",
			"ability:1=02",
			"reward:owner:2:0=00000000",
			"netcode:owner=ffffffff",
			"audit:925:owner:2:ffffffff",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})

	t.Run("owned-level-clears-reward-argument", func(t *testing.T) {
		w := newAbilityRewardUseTestWorld53FAE0()
		w.owner.update.player.levels[2] = math.MaxUint32
		w.rewardResult = 1
		if got := useAbilityReward53FAE0(w.hooks()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.observedArgs != [2]int32{2, 0} {
			t.Fatalf("reward args = %v, want [2 0]", w.observedArgs)
		}
	})
}

func TestUseAbilityReward53FAE0ReloadsPlayerAfterFlagCallback(t *testing.T) {
	w := newAbilityRewardUseTestWorld53FAE0()
	replacement := &abilityRewardUseTestPlayer53FAE0{name: "replacement"}
	replacement.levels[2] = 9
	w.afterFlags = func() { w.owner.update.player = replacement }
	w.rewardResult = 1
	if got := useAbilityReward53FAE0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.observedArgs != [2]int32{2, 0} {
		t.Fatalf("reward args = %v, want [2 0]", w.observedArgs)
	}
	want := []string{"player:replacement", "ability:1=02", "level:replacement:2=00000009"}
	for i := range w.events {
		if i+len(want) <= len(w.events) && reflect.DeepEqual(w.events[i:i+len(want)], want) {
			return
		}
	}
	t.Fatalf("events = %v, missing %v", w.events, want)
}
