package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupDefaultTestObject4F31E0 struct {
	name     string
	team     uint8
	class    uint32
	holder   *pickupDefaultTestObject4F31E0
	first    *pickupDefaultTestObject4F31E0
	next     *pickupDefaultTestObject4F31E0
	weight   uint8
	capacity uint16
	typeInd  uint16
	update   *pickupDefaultTestUpdate4F31E0
}

type pickupDefaultTestTeam4F31E0 struct {
	id    uint8
	color uint8
}

type pickupDefaultTestUpdate4F31E0 struct {
	player *pickupDefaultTestPlayer4F31E0
}

type pickupDefaultTestPlayer4F31E0 struct {
	index uint8
}

type pickupDefaultTestWorld4F31E0 struct {
	events []string
	flags  map[uint32]int32
	teams  map[uint8]*pickupDefaultTestTeam4F31E0
}

func pickupDefaultTestObjectName4F31E0(object *pickupDefaultTestObject4F31E0) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func (w *pickupDefaultTestWorld4F31E0) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *pickupDefaultTestWorld4F31E0) hooks() pickupDefaultHooks4F31E0[
	*pickupDefaultTestObject4F31E0,
	*pickupDefaultTestTeam4F31E0,
	*pickupDefaultTestUpdate4F31E0,
	*pickupDefaultTestPlayer4F31E0,
] {
	return pickupDefaultHooks4F31E0[
		*pickupDefaultTestObject4F31E0,
		*pickupDefaultTestTeam4F31E0,
		*pickupDefaultTestUpdate4F31E0,
		*pickupDefaultTestPlayer4F31E0,
	]{
		gameFlagsCheck: func(mask uint32) int32 {
			w.event("flags:%#x", mask)
			return w.flags[mask]
		},
		itemHasTeam: func(item *pickupDefaultTestObject4F31E0) bool {
			w.event("has-team:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.team != 0
		},
		teamsSame: func(owner, item *pickupDefaultTestObject4F31E0) bool {
			w.event("same-team:%s:%s", pickupDefaultTestObjectName4F31E0(owner), pickupDefaultTestObjectName4F31E0(item))
			return owner.team != 0 && owner.team == item.team
		},
		loadTeamID: func(item *pickupDefaultTestObject4F31E0) uint8 {
			w.event("team-id:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.team
		},
		findTeam: func(id uint8) *pickupDefaultTestTeam4F31E0 {
			w.event("find-team:%d", id)
			return w.teams[id]
		},
		loadClassLow: func(owner *pickupDefaultTestObject4F31E0) uint8 {
			w.event("class:%s", pickupDefaultTestObjectName4F31E0(owner))
			return uint8(owner.class)
		},
		loadUpdate: func(owner *pickupDefaultTestObject4F31E0) *pickupDefaultTestUpdate4F31E0 {
			w.event("update:%s", pickupDefaultTestObjectName4F31E0(owner))
			return owner.update
		},
		loadTeamColor: func(team *pickupDefaultTestTeam4F31E0) uint8 {
			w.event("team-color:%d", team.id)
			return team.color
		},
		loadPlayer: func(update *pickupDefaultTestUpdate4F31E0) *pickupDefaultTestPlayer4F31E0 {
			w.event("player")
			return update.player
		},
		loadPlayerInd: func(player *pickupDefaultTestPlayer4F31E0) uint8 {
			w.event("player-index:%d", player.index)
			return player.index
		},
		informTeam: func(index, code uint8, color uint32) {
			w.event("inform:%d:%d:%d", index, code, color)
		},
		loadInventoryHolder: func(item *pickupDefaultTestObject4F31E0) *pickupDefaultTestObject4F31E0 {
			w.event("holder:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.holder
		},
		loadCarryCapacity: func(owner *pickupDefaultTestObject4F31E0) uint16 {
			w.event("capacity:%s", pickupDefaultTestObjectName4F31E0(owner))
			return owner.capacity
		},
		loadInventoryFirst: func(owner *pickupDefaultTestObject4F31E0) *pickupDefaultTestObject4F31E0 {
			w.event("first:%s", pickupDefaultTestObjectName4F31E0(owner))
			return owner.first
		},
		loadWeight: func(item *pickupDefaultTestObject4F31E0) uint8 {
			w.event("weight:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.weight
		},
		loadInventoryNext: func(item *pickupDefaultTestObject4F31E0) *pickupDefaultTestObject4F31E0 {
			w.event("next:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.next
		},
		primaryMessage: func(owner *pickupDefaultTestObject4F31E0, message string, value uint8) {
			w.event("message:%s:%s:%d", pickupDefaultTestObjectName4F31E0(owner), message, value)
		},
		loadItemClass: func(item *pickupDefaultTestObject4F31E0) uint32 {
			w.event("item-class:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.class
		},
		loadItemType: func(item *pickupDefaultTestObject4F31E0) uint16 {
			w.event("item-type:%s", pickupDefaultTestObjectName4F31E0(item))
			return item.typeInd
		},
		countInventory: func(owner *pickupDefaultTestObject4F31E0, typeInd int32) int32 {
			w.event("count:%s:%d", pickupDefaultTestObjectName4F31E0(owner), typeInd)
			return int32(typeInd >> 8)
		},
		deleteWorldObject: func(item *pickupDefaultTestObject4F31E0) {
			w.event("delete:%s", pickupDefaultTestObjectName4F31E0(item))
		},
		inventoryPut: func(owner, item *pickupDefaultTestObject4F31E0, report int32) {
			w.event("put:%s:%s:%d", pickupDefaultTestObjectName4F31E0(owner), pickupDefaultTestObjectName4F31E0(item), report)
		},
	}
}

func TestPickupDefault4F31E0ExactOrdinarySuccessTrace(t *testing.T) {
	second := &pickupDefaultTestObject4F31E0{name: "second", weight: 17}
	first := &pickupDefaultTestObject4F31E0{name: "first", weight: 11, next: second}
	owner := &pickupDefaultTestObject4F31E0{name: "owner", capacity: 50, first: first}
	item := &pickupDefaultTestObject4F31E0{name: "item", weight: 20, class: 0x80000000}
	w := &pickupDefaultTestWorld4F31E0{flags: map[uint32]int32{}, teams: map[uint8]*pickupDefaultTestTeam4F31E0{}}

	if got := pickupDefault4F31E0(owner, item, -7, math.MaxInt32, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"flags:0x1000",
		"has-team:item",
		"holder:item",
		"capacity:owner",
		"first:owner",
		"weight:first", "next:first",
		"weight:second", "next:second",
		"weight:item",
		"item-class:item",
		"delete:item",
		"put:owner:item:-7",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
}

func TestPickupDefault4F31E0TeamRejectionLoadOrder(t *testing.T) {
	team := &pickupDefaultTestTeam4F31E0{id: 9, color: 6}
	player := &pickupDefaultTestPlayer4F31E0{index: 7}
	owner := &pickupDefaultTestObject4F31E0{
		name: "owner", team: 2, class: uint32(pickupDefaultPlayerClass4F31E0),
		update: &pickupDefaultTestUpdate4F31E0{player: player},
	}
	item := &pickupDefaultTestObject4F31E0{name: "item", team: team.id}
	w := &pickupDefaultTestWorld4F31E0{
		flags: map[uint32]int32{},
		teams: map[uint8]*pickupDefaultTestTeam4F31E0{team.id: team},
	}

	if got := pickupDefault4F31E0(owner, item, 1, 2, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"flags:0x1000", "has-team:item", "same-team:owner:item",
		"team-id:item", "find-team:9", "class:owner", "update:owner",
		"team-color:9", "player", "player-index:7", "inform:7:16:6",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
}

func TestPickupDefault4F31E0QuestSkipsTeamAndCachesCapacity(t *testing.T) {
	owner := &pickupDefaultTestObject4F31E0{name: "owner", capacity: 10}
	item := &pickupDefaultTestObject4F31E0{name: "item", team: 9, weight: 20}
	w := &pickupDefaultTestWorld4F31E0{
		flags: map[uint32]int32{pickupDefaultQuestFlag4F31E0: -1},
		teams: map[uint8]*pickupDefaultTestTeam4F31E0{},
	}
	hooks := w.hooks()
	originalCapacity := hooks.loadCarryCapacity
	hooks.loadCarryCapacity = func(owner *pickupDefaultTestObject4F31E0) uint16 {
		capacity := originalCapacity(owner)
		owner.capacity = 0
		return capacity
	}

	if got := pickupDefault4F31E0(owner, item, 0, 99, hooks); got != 1 {
		t.Fatalf("result = %d, want cached-capacity success", got)
	}
	want := []string{
		"flags:0x1000", "holder:item", "capacity:owner", "first:owner",
		"weight:item", "item-class:item", "delete:item", "put:owner:item:0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
}

func TestPickupDefault4F31E0HolderCapacityAndWeightRejections(t *testing.T) {
	tests := []struct {
		name      string
		owner     *pickupDefaultTestObject4F31E0
		item      *pickupDefaultTestObject4F31E0
		wantTrace []string
	}{
		{
			name:      "already held",
			owner:     &pickupDefaultTestObject4F31E0{name: "owner", capacity: 10},
			item:      &pickupDefaultTestObject4F31E0{name: "item", holder: &pickupDefaultTestObject4F31E0{name: "holder"}},
			wantTrace: []string{"flags:0x1000", "has-team:item", "holder:item"},
		},
		{
			name:      "zero capacity",
			owner:     &pickupDefaultTestObject4F31E0{name: "owner"},
			item:      &pickupDefaultTestObject4F31E0{name: "item"},
			wantTrace: []string{"flags:0x1000", "has-team:item", "holder:item", "capacity:owner"},
		},
		{
			name: "too heavy",
			owner: &pickupDefaultTestObject4F31E0{
				name: "owner", capacity: 10,
				first: &pickupDefaultTestObject4F31E0{name: "first", weight: 12},
			},
			item: &pickupDefaultTestObject4F31E0{name: "item", weight: 9},
			wantTrace: []string{
				"flags:0x1000", "has-team:item", "holder:item", "capacity:owner",
				"first:owner", "weight:first", "next:first", "weight:item",
				"message:owner:pickup.c:CarryingTooMuch:0",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &pickupDefaultTestWorld4F31E0{flags: map[uint32]int32{}, teams: map[uint8]*pickupDefaultTestTeam4F31E0{}}
			if got := pickupDefault4F31E0(test.owner, test.item, 1, 1, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, test.wantTrace) {
				t.Fatalf("events =\n%v\nwant =\n%v", w.events, test.wantTrace)
			}
		})
	}
}

func TestPickupDefault4F31E0FoodLimitsAndOrder(t *testing.T) {
	tests := []struct {
		name      string
		mode      int32
		encoded   uint16
		want      int32
		wantLimit int32
	}{
		{name: "normal below", encoded: 2 << 8, want: 1, wantLimit: 3},
		{name: "normal equal", encoded: 3 << 8, want: 0, wantLimit: 3},
		{name: "quest below", mode: 1, encoded: 8 << 8, want: 1, wantLimit: 9},
		{name: "quest equal", mode: -1, encoded: 9 << 8, want: 0, wantLimit: 9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := &pickupDefaultTestObject4F31E0{name: "owner", capacity: 10}
			item := &pickupDefaultTestObject4F31E0{
				name: "item", class: pickupDefaultFoodClass4F31E0, typeInd: test.encoded,
			}
			w := &pickupDefaultTestWorld4F31E0{
				flags: map[uint32]int32{pickupDefaultQuestCoopFlags4F31E0: test.mode},
				teams: map[uint8]*pickupDefaultTestTeam4F31E0{},
			}
			got := pickupDefault4F31E0(owner, item, 4, 5, w.hooks())
			if got != test.want {
				t.Fatalf("result = %d, want %d (limit %d)", got, test.want, test.wantLimit)
			}
			prefix := []string{
				"flags:0x1000", "has-team:item", "holder:item", "capacity:owner",
				"first:owner", "weight:item", "item-class:item", "item-type:item",
				fmt.Sprintf("count:owner:%d", test.encoded), "flags:0x1800",
			}
			if !reflect.DeepEqual(w.events[:len(prefix)], prefix) {
				t.Fatalf("prefix =\n%v\nwant =\n%v", w.events[:len(prefix)], prefix)
			}
			if got == 0 {
				if last := w.events[len(w.events)-1]; last != "message:owner:pickup.c:MaxSameItem:0" {
					t.Fatalf("last event = %q", last)
				}
			} else if tail := w.events[len(w.events)-2:]; !reflect.DeepEqual(tail, []string{"delete:item", "put:owner:item:4"}) {
				t.Fatalf("success tail = %v", tail)
			}
		})
	}
}

func TestPickupDefault4F31E0WrappingHelpers(t *testing.T) {
	if got := pickupDefaultWeightAdd4F31E0(math.MaxUint32, 1); got != 0 {
		t.Fatalf("wrapped sum = %#x, want 0", got)
	}
	if got := pickupDefaultWeightBudget4F31E0(0xffff, math.MaxUint32); got != 0x1ffff {
		t.Fatalf("wrapped budget = %#x, want 0x1ffff", got)
	}
	if got := pickupDefaultWeightBudget4F31E0(1, 3); got != -1 {
		t.Fatalf("signed budget = %d, want -1", got)
	}
	if pickupDefaultFoodBelowLimit4F31E0(math.MinInt32, 3) {
		t.Fatal("INT32_MIN - 3 wraps positive and must reject")
	}
	if pickupDefaultFoodBelowLimit4F31E0(math.MaxInt32, 9) {
		t.Fatal("INT32_MAX - 9 stays positive and must reject")
	}
}

func TestPickupDefault4F31E0FaultPrefixStartsAfterModeCall(t *testing.T) {
	w := &pickupDefaultTestWorld4F31E0{flags: map[uint32]int32{}, teams: map[uint8]*pickupDefaultTestTeam4F31E0{}}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not fault")
		}
		if want := []string{"flags:0x1000", "has-team:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	pickupDefault4F31E0(
		(*pickupDefaultTestObject4F31E0)(nil),
		(*pickupDefaultTestObject4F31E0)(nil),
		0, 0, w.hooks(),
	)
}
