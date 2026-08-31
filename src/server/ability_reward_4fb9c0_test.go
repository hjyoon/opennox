package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type abilityRewardTestObject4FB9C0 struct {
	name   string
	class  uint8
	update *abilityRewardTestUpdate4FB9C0
}

type abilityRewardTestUpdate4FB9C0 struct {
	player *abilityRewardTestPlayer4FB9C0
}

type abilityRewardTestPlayer4FB9C0 struct {
	name       string
	levels     [6]uint32
	protection uint32
}

type abilityRewardTestWorld4FB9C0 struct {
	unit        *abilityRewardTestObject4FB9C0
	first       *abilityRewardTestObject4FB9C0
	next        map[*abilityRewardTestObject4FB9C0]*abilityRewardTestObject4FB9C0
	quest       int32
	state       int32
	events      []string
	faultAt     int
	afterStore  func(*abilityRewardTestPlayer4FB9C0, int32, uint32)
	playerLoads []*abilityRewardTestPlayer4FB9C0
	loadPlayerN int
}

func abilityRewardTestObjectName4FB9C0(obj *abilityRewardTestObject4FB9C0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *abilityRewardTestWorld4FB9C0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *abilityRewardTestWorld4FB9C0) hooks() abilityRewardHooks4FB9C0[
	*abilityRewardTestObject4FB9C0,
	*abilityRewardTestUpdate4FB9C0,
	*abilityRewardTestPlayer4FB9C0,
	string,
] {
	return abilityRewardHooks4FB9C0[
		*abilityRewardTestObject4FB9C0,
		*abilityRewardTestUpdate4FB9C0,
		*abilityRewardTestPlayer4FB9C0,
		string,
	]{
		loadUnitArg: func() *abilityRewardTestObject4FB9C0 {
			w.event("unit:" + abilityRewardTestObjectName4FB9C0(w.unit))
			return w.unit
		},
		loadClassLow: func(obj *abilityRewardTestObject4FB9C0) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", abilityRewardTestObjectName4FB9C0(obj), obj.class))
			return obj.class
		},
		loadUpdateData: func(obj *abilityRewardTestObject4FB9C0) *abilityRewardTestUpdate4FB9C0 {
			w.event("update:" + abilityRewardTestObjectName4FB9C0(obj))
			return obj.update
		},
		loadPlayer: func(update *abilityRewardTestUpdate4FB9C0) *abilityRewardTestPlayer4FB9C0 {
			var player *abilityRewardTestPlayer4FB9C0
			if w.loadPlayerN < len(w.playerLoads) {
				player = w.playerLoads[w.loadPlayerN]
			} else {
				player = update.player
			}
			w.loadPlayerN++
			w.event("player:" + player.name)
			return player
		},
		loadAbilityLevel: func(player *abilityRewardTestPlayer4FB9C0, ability int32) uint32 {
			value := player.levels[ability]
			w.event(fmt.Sprintf("level:%s:%d=%08x", player.name, ability, value))
			return value
		},
		storeAbilityLevel: func(player *abilityRewardTestPlayer4FB9C0, ability int32, value uint32) {
			w.event(fmt.Sprintf("store:%s:%d=%08x", player.name, ability, value))
			player.levels[ability] = value
			if w.afterStore != nil {
				w.afterStore(player, ability, value)
			}
		},
		loadProtection: func(player *abilityRewardTestPlayer4FB9C0) uint32 {
			w.event(fmt.Sprintf("protection:%s=%08x", player.name, player.protection))
			return player.protection
		},
		loadString: func(key, path string, line int) string {
			w.event(fmt.Sprintf("string:%s:%s:%d", key, path, line))
			return "localized:" + key
		},
		sendLineMessage: func(obj *abilityRewardTestObject4FB9C0, message string) {
			w.event("line:" + abilityRewardTestObjectName4FB9C0(obj) + ":" + message)
		},
		primaryMessage: func(obj *abilityRewardTestObject4FB9C0, message string, value uint8) {
			w.event(fmt.Sprintf("primary:%s:%s:%d", abilityRewardTestObjectName4FB9C0(obj), message, value))
		},
		awardProtection: func(token uint32, ability, level int32) {
			w.event(fmt.Sprintf("award:%08x:%d:%08x", token, ability, uint32(level)))
		},
		reportAbility: func(obj *abilityRewardTestObject4FB9C0, ability, rewardArg int32) {
			w.event(fmt.Sprintf("report:%s:%d:%08x", abilityRewardTestObjectName4FB9C0(obj), ability, uint32(rewardArg)))
		},
		gameFlagsCheck: func(mask uint32) int32 {
			w.event(fmt.Sprintf("flags:%08x=%08x", mask, uint32(w.quest)))
			return w.quest
		},
		rewardNotify: func(recipient *abilityRewardTestObject4FB9C0, kind int32, source *abilityRewardTestObject4FB9C0, ability int32) {
			w.event(fmt.Sprintf("notify:%s:%d:%s:%d", abilityRewardTestObjectName4FB9C0(recipient), kind, abilityRewardTestObjectName4FB9C0(source), ability))
		},
		checkPlayerState: func(obj *abilityRewardTestObject4FB9C0) int32 {
			w.event(fmt.Sprintf("state:%s=%08x", abilityRewardTestObjectName4FB9C0(obj), uint32(w.state)))
			return w.state
		},
		firstPlayerUnit: func() *abilityRewardTestObject4FB9C0 {
			w.event("first:" + abilityRewardTestObjectName4FB9C0(w.first))
			return w.first
		},
		nextPlayerUnit: func(obj *abilityRewardTestObject4FB9C0) *abilityRewardTestObject4FB9C0 {
			value := w.next[obj]
			w.event("next:" + abilityRewardTestObjectName4FB9C0(obj) + "=" + abilityRewardTestObjectName4FB9C0(value))
			return value
		},
	}
}

func newAbilityRewardTestWorld4FB9C0() *abilityRewardTestWorld4FB9C0 {
	player := &abilityRewardTestPlayer4FB9C0{name: "player", protection: 0x89abcdef}
	unit := &abilityRewardTestObject4FB9C0{name: "unit", class: 0xf4}
	unit.update = &abilityRewardTestUpdate4FB9C0{player: player}
	other := &abilityRewardTestObject4FB9C0{name: "other", class: 4}
	return &abilityRewardTestWorld4FB9C0{
		unit:  unit,
		first: unit,
		next: map[*abilityRewardTestObject4FB9C0]*abilityRewardTestObject4FB9C0{
			unit:  other,
			other: nil,
		},
		quest: 1,
	}
}

func abilityRewardSuccessTrace4FB9C0() []string {
	return []string{
		"unit:unit",
		"class:unit=f4",
		"update:unit",
		"player:player",
		"level:player:3=00000000",
		"store:player:3=00000005",
		"player:player",
		"level:player:3=00000005",
		"player:player",
		"level:player:3=00000005",
		"protection:player=89abcdef",
		"award:89abcdef:3:00000005",
		"report:unit:3:80001234",
		"flags:00001000=00000001",
		"notify:unit:2:unit:3",
		"state:unit=00000000",
		"first:unit",
		"next:unit=other",
		"notify:other:2:unit:3",
		"next:other=nil",
	}
}

func TestAbilityRewardServ4FB9C0ExactSuccessTraceAndFaultPrefixes(t *testing.T) {
	want := abilityRewardSuccessTrace4FB9C0()
	build := newAbilityRewardTestWorld4FB9C0

	w := build()
	if got := abilityRewardServ4FB9C0(3, math.MinInt32+0x1234, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
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
			abilityRewardServ4FB9C0(3, math.MinInt32+0x1234, w.hooks())
		})
	}
}

func TestAbilityRewardServ4FB9C0ClassIDAndOwnedGates(t *testing.T) {
	t.Run("non-player", func(t *testing.T) {
		w := newAbilityRewardTestWorld4FB9C0()
		w.unit.class = 0xf0
		if got := abilityRewardServ4FB9C0(3, 0, w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"unit:unit", "class:unit=f0"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	for _, ability := range []int32{math.MinInt32, -1, 0, 6, math.MaxInt32} {
		t.Run(fmt.Sprintf("invalid-%08x", uint32(ability)), func(t *testing.T) {
			w := newAbilityRewardTestWorld4FB9C0()
			if got := abilityRewardServ4FB9C0(ability, -1, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			want := []string{
				"unit:unit",
				"class:unit=f4",
				fmt.Sprintf("string:AwardAbilityError:%s:108", abilityRewardMessagePath4FB9C0),
				"line:unit:localized:AwardAbilityError",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		})
	}

	t.Run("already-owned", func(t *testing.T) {
		w := newAbilityRewardTestWorld4FB9C0()
		w.unit.update.player.levels[5] = 0x80000000
		if got := abilityRewardServ4FB9C0(5, 1, w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{
			"unit:unit", "class:unit=f4", "update:unit", "player:player",
			"level:player:5=80000000", "primary:unit:use.c:HadAbility:0",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})
}

func TestAbilityRewardServ4FB9C0ReloadsPlayerAndUsesSignedClamp(t *testing.T) {
	for _, tc := range []struct {
		name      string
		liveLevel uint32
		wantLevel uint32
	}{
		{name: "six", liveLevel: 6, wantLevel: 5},
		{name: "max-signed", liveLevel: math.MaxInt32, wantLevel: 5},
		{name: "min-signed", liveLevel: 0x80000000, wantLevel: 0x80000000},
		{name: "minus-one", liveLevel: math.MaxUint32, wantLevel: math.MaxUint32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newAbilityRewardTestWorld4FB9C0()
			w.quest = 0
			initial := w.unit.update.player
			live := &abilityRewardTestPlayer4FB9C0{name: "live", protection: 0x10203040}
			live.levels[2] = tc.liveLevel
			w.afterStore = func(player *abilityRewardTestPlayer4FB9C0, ability int32, value uint32) {
				if player == initial && ability == 2 && value == 5 {
					w.unit.update.player = live
				}
			}
			if got := abilityRewardServ4FB9C0(2, -7, w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if initial.levels[2] != 5 || live.levels[2] != tc.wantLevel {
				t.Fatalf("initial/live level = %#x/%#x, want 5/%#x", initial.levels[2], live.levels[2], tc.wantLevel)
			}
			wantAward := fmt.Sprintf("award:10203040:2:%08x", tc.wantLevel)
			found := false
			for _, event := range w.events {
				if event == wantAward {
					found = true
				}
			}
			if !found {
				t.Fatalf("events = %v, missing %q", w.events, wantAward)
			}
		})
	}
}

func TestAbilityRewardServ4FB9C0QuestSuppressionStillNotifiesSource(t *testing.T) {
	w := newAbilityRewardTestWorld4FB9C0()
	w.state = math.MinInt32
	if got := abilityRewardServ4FB9C0(1, 0, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantTail := []string{
		"flags:00001000=00000001",
		"notify:unit:2:unit:1",
		"state:unit=80000000",
	}
	if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}
