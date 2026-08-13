package server

import (
	"reflect"
	"testing"
)

type doorKeyTestObject4E8910 struct {
	name     string
	class    uint8
	typeName string
	data     *doorKeyTestData4E8910
	owner    *doorKeyTestObject4E8910
	first    *doorKeyTestObject4E8910
	next     *doorKeyTestObject4E8910
	typeInd  uint16
}

type doorKeyTestData4E8910 struct {
	lockCode uint8
}

type doorKeyTestState4E8910 struct {
	events       []string
	questMode    bool
	questState   int32
	fallback     *doorKeyTestObject4E8910
	onTypeName   func(*doorKeyTestObject4E8910)
	onQuestState func()
}

func (s *doorKeyTestState4E8910) hooks() doorCheckKeyHooks4E8910[
	*doorKeyTestObject4E8910,
	*doorKeyTestData4E8910,
] {
	return doorCheckKeyHooks4E8910[*doorKeyTestObject4E8910, *doorKeyTestData4E8910]{
		loadDoorData: func(obj *doorKeyTestObject4E8910) *doorKeyTestData4E8910 {
			s.events = append(s.events, "door-data")
			return obj.data
		},
		loadLockCode: func(data *doorKeyTestData4E8910) uint8 {
			s.events = append(s.events, "lock")
			return data.lockCode
		},
		loadOwner: func(obj *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "owner")
			return obj.owner
		},
		firstItem: func(obj *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "first:"+obj.name)
			return obj.first
		},
		loadClassByte: func(obj *doorKeyTestObject4E8910) uint8 {
			s.events = append(s.events, "class:"+obj.name)
			return obj.class
		},
		loadTypeName: func(obj *doorKeyTestObject4E8910) string {
			s.events = append(s.events, "name:"+obj.name)
			name := obj.typeName
			if s.onTypeName != nil {
				s.onTypeName(obj)
			}
			return name
		},
		nextItem: func(obj *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "next:"+obj.name)
			return obj.next
		},
		hasQuestGameMode: func() bool {
			s.events = append(s.events, "quest-mode")
			return s.questMode
		},
		loadQuestKeyState: func() int32 {
			s.events = append(s.events, "quest-state")
			if s.onQuestState != nil {
				s.onQuestState()
			}
			return s.questState
		},
		playersHaveSilverKey: func() *doorKeyTestObject4E8910 {
			s.events = append(s.events, "silver-fallback")
			return s.fallback
		},
	}
}

func doorKeyTestDoor4E8910(data *doorKeyTestData4E8910) *doorKeyTestObject4E8910 {
	return &doorKeyTestObject4E8910{data: data}
}

func TestDoorKeyNameMatches4E8910FixedCStrings(t *testing.T) {
	tests := []struct {
		code uint8
		name string
		want bool
	}{
		{1, "SilverKey", true},
		{2, "GoldKey", true},
		{3, "RubyKey", true},
		{4, "SapphireKey", true},
		{1, "SilverKey\x00ignored", true},
		{1, "SilverKeyX", false},
		{0, "SilverKey", false},
		{5, "SilverKey", false},
	}
	for _, tc := range tests {
		if got := doorKeyNameMatches4E8910(tc.code, tc.name); got != tc.want {
			t.Fatalf("match(%d, %q) = %v, want %v", tc.code, tc.name, got, tc.want)
		}
	}
}

func TestDoorCheckKey4E8910EarlyGatesAndNilUnitFault(t *testing.T) {
	for _, tc := range []struct {
		name   string
		lock   uint8
		owner  bool
		events []string
	}{
		{"mechanism", 5, false, []string{"door-data", "lock"}},
		{"owned", 1, true, []string{"door-data", "lock", "owner"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &doorKeyTestData4E8910{lockCode: tc.lock}
			door := doorKeyTestDoor4E8910(data)
			if tc.owner {
				door.owner = &doorKeyTestObject4E8910{name: "owner"}
			}
			state := &doorKeyTestState4E8910{}
			if got := doorCheckKey4E8910[*doorKeyTestObject4E8910](nil, door, state.hooks()); got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			if !reflect.DeepEqual(state.events, tc.events) {
				t.Fatalf("events = %v, want %v", state.events, tc.events)
			}
		})
	}

	data := &doorKeyTestData4E8910{lockCode: 1}
	door := doorKeyTestDoor4E8910(data)
	state := &doorKeyTestState4E8910{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not fault at first inventory load")
		}
		want := []string{"door-data", "lock", "owner"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()
	doorCheckKey4E8910[*doorKeyTestObject4E8910](nil, door, state.hooks())
}

func TestDoorCheckKey4E8910LiveLockAndPostMatchUnitClass(t *testing.T) {
	data := &doorKeyTestData4E8910{lockCode: 1}
	door := doorKeyTestDoor4E8910(data)
	key := &doorKeyTestObject4E8910{name: "key", class: doorKeyClassByte4E8910, typeName: "GoldKey"}
	unit := &doorKeyTestObject4E8910{name: "unit", class: doorPlayerClassByte4E8910, first: key}
	state := &doorKeyTestState4E8910{}
	state.onTypeName = func(*doorKeyTestObject4E8910) { data.lockCode = 2 }

	got := doorCheckKey4E8910(unit, door, state.hooks())
	if got != key {
		t.Fatalf("result = %p, want key %p", got, key)
	}
	want := []string{"door-data", "lock", "owner", "first:unit", "class:key", "name:key", "lock", "class:unit"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestDoorCheckKey4E8910InventoryNextAndQuestFallbackOrder(t *testing.T) {
	data := &doorKeyTestData4E8910{lockCode: 2}
	door := doorKeyTestDoor4E8910(data)
	nonKey := &doorKeyTestObject4E8910{name: "non-key", class: 0x80}
	wrong := &doorKeyTestObject4E8910{name: "wrong", class: doorKeyClassByte4E8910, typeName: "RubyKey"}
	nonKey.next = wrong
	unit := &doorKeyTestObject4E8910{name: "unit", class: doorPlayerClassByte4E8910, first: nonKey}
	fallback := &doorKeyTestObject4E8910{name: "fallback"}
	state := &doorKeyTestState4E8910{questMode: true, questState: 1, fallback: fallback}
	state.onQuestState = func() { data.lockCode = 1 }

	got := doorCheckKey4E8910(unit, door, state.hooks())
	if got != fallback {
		t.Fatalf("result = %p, want fallback %p", got, fallback)
	}
	want := []string{
		"door-data", "lock", "owner", "first:unit",
		"class:non-key", "next:non-key", "class:wrong", "name:wrong", "lock", "next:wrong",
		"class:unit", "quest-mode", "quest-state", "lock", "silver-fallback",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestDoorCheckKey4E8910QuestStateMustEqualOne(t *testing.T) {
	data := &doorKeyTestData4E8910{lockCode: 1}
	door := doorKeyTestDoor4E8910(data)
	unit := &doorKeyTestObject4E8910{name: "unit", class: doorPlayerClassByte4E8910}
	state := &doorKeyTestState4E8910{questMode: true, questState: 2, fallback: &doorKeyTestObject4E8910{}}

	if got := doorCheckKey4E8910(unit, door, state.hooks()); got != nil {
		t.Fatalf("quest state 2 returned %p", got)
	}
	want := []string{"door-data", "lock", "owner", "first:unit", "class:unit", "quest-mode", "quest-state"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

type silverKeyTestState4E8A10 struct {
	events      []string
	cached      uint32
	lookup      uint32
	firstUnit   *doorKeyTestObject4E8910
	onCacheLoad func(int)
	cacheLoads  int
}

func (s *silverKeyTestState4E8A10) hooks() playersHaveSilverKeyHooks4E8A10[*doorKeyTestObject4E8910] {
	return playersHaveSilverKeyHooks4E8A10[*doorKeyTestObject4E8910]{
		loadCachedTypeID: func() uint32 {
			s.events = append(s.events, "cache")
			value := s.cached
			if s.onCacheLoad != nil {
				s.onCacheLoad(s.cacheLoads)
			}
			s.cacheLoads++
			return value
		},
		lookupTypeID: func() uint32 {
			s.events = append(s.events, "lookup")
			return s.lookup
		},
		storeCachedTypeID: func(value uint32) {
			s.events = append(s.events, "store")
			s.cached = value
		},
		firstPlayerUnit: func() *doorKeyTestObject4E8910 {
			s.events = append(s.events, "first-player")
			return s.firstUnit
		},
		nextPlayerUnit: func(unit *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "next-player:"+unit.name)
			return unit.next
		},
		firstItem: func(unit *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "first-item:"+unit.name)
			return unit.first
		},
		nextItem: func(item *doorKeyTestObject4E8910) *doorKeyTestObject4E8910 {
			s.events = append(s.events, "next-item:"+item.name)
			return item.next
		},
		loadTypeInd: func(item *doorKeyTestObject4E8910) uint16 {
			s.events = append(s.events, "type:"+item.name)
			return item.typeInd
		},
	}
}

func TestPlayersHaveSilverKey4E8A10InitializesCacheSelectsLastUnitAndFirstKey(t *testing.T) {
	u1key := &doorKeyTestObject4E8910{name: "u1-key", typeInd: 7}
	u1 := &doorKeyTestObject4E8910{name: "u1", first: u1key}
	u2wrong := &doorKeyTestObject4E8910{name: "u2-wrong", typeInd: 9}
	u2key1 := &doorKeyTestObject4E8910{name: "u2-key1", typeInd: 7}
	u2key2 := &doorKeyTestObject4E8910{name: "u2-key2", typeInd: 7}
	u2wrong.next, u2key1.next = u2key1, u2key2
	u2 := &doorKeyTestObject4E8910{name: "u2", first: u2wrong}
	u1.next = u2
	state := &silverKeyTestState4E8A10{lookup: 7, firstUnit: u1}

	got := playersHaveSilverKey4E8A10(state.hooks())
	if got != u2key1 {
		t.Fatalf("result = %p, want last unit's first key %p", got, u2key1)
	}
	wantPrefix := []string{"cache", "lookup", "store", "first-player", "first-item:u1", "cache", "type:u1-key"}
	if len(state.events) < len(wantPrefix) || !reflect.DeepEqual(state.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix = %v, want %v", state.events, wantPrefix)
	}
	if state.cached != 7 {
		t.Fatalf("cached type = %d, want 7", state.cached)
	}
	if state.events[len(state.events)-2] != "cache" || state.events[len(state.events)-1] != "type:u2-key1" {
		t.Fatalf("second-pass tail = %v, want cache then first matching type", state.events[len(state.events)-2:])
	}
}

func TestPlayersHaveSilverKey4E8A10ReadsCacheBeforeTypeAndSkipsLookupWhenSet(t *testing.T) {
	item := &doorKeyTestObject4E8910{name: "item", typeInd: 8}
	unit := &doorKeyTestObject4E8910{name: "unit", first: item}
	state := &silverKeyTestState4E8A10{cached: 7, lookup: 99, firstUnit: unit}
	state.onCacheLoad = func(index int) {
		if index == 1 {
			item.typeInd = 7
		}
	}

	got := playersHaveSilverKey4E8A10(state.hooks())
	if got != item {
		t.Fatalf("result = %p, want item %p", got, item)
	}
	for _, event := range state.events {
		if event == "lookup" || event == "store" {
			t.Fatalf("nonzero cache unexpectedly performed %q: %v", event, state.events)
		}
	}
	firstCache, firstType := -1, -1
	for i, event := range state.events {
		if event == "cache" && firstCache < 0 && i > 0 {
			firstCache = i
		}
		if event == "type:item" && firstType < 0 {
			firstType = i
		}
	}
	if firstCache < 0 || firstType != firstCache+1 {
		t.Fatalf("cache/type order = %v", state.events)
	}
}

func TestPlayersHaveSilverKey4E8A10NoUnitStopsAfterFirstPlayerLookup(t *testing.T) {
	state := &silverKeyTestState4E8A10{cached: 3}
	if got := playersHaveSilverKey4E8A10(state.hooks()); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{"cache", "first-player"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}
