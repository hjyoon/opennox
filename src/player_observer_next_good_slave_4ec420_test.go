package opennox

import (
	"reflect"
	"testing"
)

type playerObserverGoodSlaveTestData4EC420 struct {
	name   string
	status uint8
}

type playerObserverGoodSlaveTestObject4EC420 struct {
	name      string
	class     uint8
	owner     *playerObserverGoodSlaveTestObject4EC420
	nextOwned *playerObserverGoodSlaveTestObject4EC420
	data      *playerObserverGoodSlaveTestData4EC420
}

func defaultPlayerObserverGoodSlaveHooks4EC420() playerObserverFindGoodSlaveHooks4EC420[
	*playerObserverGoodSlaveTestObject4EC420,
	*playerObserverGoodSlaveTestData4EC420,
] {
	return playerObserverFindGoodSlaveHooks4EC420[
		*playerObserverGoodSlaveTestObject4EC420,
		*playerObserverGoodSlaveTestData4EC420,
	]{
		loadOwner: func(current *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
			return current.owner
		},
		loadNextOwned: func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
			return obj.nextOwned
		},
		loadClassByte: func(candidate *playerObserverGoodSlaveTestObject4EC420) uint8 {
			return candidate.class
		},
		loadUpdateData: func(candidate *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestData4EC420 {
			return candidate.data
		},
		loadStatusByte: func(data *playerObserverGoodSlaveTestData4EC420) uint8 {
			return data.status
		},
	}
}

func TestPlayerObserverFindGoodSlaveContract4EC420GuardLoadOrder(t *testing.T) {
	events := make([]string, 0, 2)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC420()
	hooks.loadOwner = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "owner:"+obj.name)
		return obj.owner
	}
	hooks.loadNextOwned = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "next:"+obj.name)
		return obj.nextOwned
	}
	hooks.loadClassByte = func(*playerObserverGoodSlaveTestObject4EC420) uint8 {
		t.Fatal("guard path loaded a candidate class")
		return 0
	}

	if got := playerObserverFindGoodSlaveContract4EC420(nil, hooks); got != nil {
		t.Fatalf("nil current result = %p", got)
	}
	if len(events) != 0 {
		t.Fatalf("nil current events = %v", events)
	}

	current := &playerObserverGoodSlaveTestObject4EC420{name: "current"}
	if got := playerObserverFindGoodSlaveContract4EC420(current, hooks); got != nil {
		t.Fatalf("ownerless current result = %p", got)
	}
	if want := []string{"owner:current"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("ownerless events = %v, want %v", events, want)
	}

	events = events[:0]
	current.owner = &playerObserverGoodSlaveTestObject4EC420{name: "opaque-owner"}
	if got := playerObserverFindGoodSlaveContract4EC420(current, hooks); got != nil {
		t.Fatalf("terminal current result = %p", got)
	}
	if want := []string{"owner:current", "next:current"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("terminal events = %v, want %v", events, want)
	}
}

func TestPlayerObserverFindGoodSlaveContract4EC420TraversalOrder(t *testing.T) {
	owner := &playerObserverGoodSlaveTestObject4EC420{name: "owner"}
	nonMonster := &playerObserverGoodSlaveTestObject4EC420{name: "non-monster", class: 0xfc}
	dormantData := &playerObserverGoodSlaveTestData4EC420{name: "dormant-data", status: 0x7f}
	dormant := &playerObserverGoodSlaveTestObject4EC420{name: "dormant", class: 0x82, data: dormantData}
	summonedData := &playerObserverGoodSlaveTestData4EC420{name: "summoned-data", status: 0x81}
	summoned := &playerObserverGoodSlaveTestObject4EC420{name: "summoned", class: 0x02, data: summonedData}
	ignored := &playerObserverGoodSlaveTestObject4EC420{name: "ignored"}
	current := &playerObserverGoodSlaveTestObject4EC420{name: "current", owner: owner, nextOwned: nonMonster}
	nonMonster.nextOwned = dormant
	dormant.nextOwned = summoned
	summoned.nextOwned = ignored

	events := make([]string, 0, 14)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC420()
	hooks.loadOwner = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "owner:"+obj.name)
		return obj.owner
	}
	hooks.loadNextOwned = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "next:"+obj.name)
		return obj.nextOwned
	}
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC420) uint8 {
		events = append(events, "class:"+obj.name)
		return obj.class
	}
	hooks.loadUpdateData = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestData4EC420 {
		events = append(events, "data:"+obj.name)
		return obj.data
	}
	hooks.loadStatusByte = func(data *playerObserverGoodSlaveTestData4EC420) uint8 {
		events = append(events, "status:"+data.name)
		return data.status
	}

	if got := playerObserverFindGoodSlaveContract4EC420(current, hooks); got != summoned {
		t.Fatalf("result = %p, want %p", got, summoned)
	}
	want := []string{
		"owner:current", "next:current",
		"class:non-monster", "next:non-monster",
		"class:dormant", "data:dormant", "status:dormant-data", "next:dormant",
		"class:summoned", "data:summoned", "status:summoned-data",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPlayerObserverFindGoodSlaveContract4EC420UsesLiveNext(t *testing.T) {
	owner := &playerObserverGoodSlaveTestObject4EC420{name: "owner"}
	first := &playerObserverGoodSlaveTestObject4EC420{name: "first"}
	stale := &playerObserverGoodSlaveTestObject4EC420{name: "stale"}
	liveData := &playerObserverGoodSlaveTestData4EC420{name: "live-data", status: 0x80}
	live := &playerObserverGoodSlaveTestObject4EC420{name: "live", class: 0x02, data: liveData}
	current := &playerObserverGoodSlaveTestObject4EC420{name: "current", owner: owner, nextOwned: first}
	first.nextOwned = stale

	hooks := defaultPlayerObserverGoodSlaveHooks4EC420()
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC420) uint8 {
		if obj == first {
			first.nextOwned = live
		}
		return obj.class
	}
	if got := playerObserverFindGoodSlaveContract4EC420(current, hooks); got != live {
		t.Fatalf("result = %p, want live successor %p", got, live)
	}
}

func TestPlayerObserverFindGoodSlaveContract4EC420NilMonsterDataFaultsBeforeNext(t *testing.T) {
	owner := &playerObserverGoodSlaveTestObject4EC420{name: "owner"}
	monster := &playerObserverGoodSlaveTestObject4EC420{name: "monster", class: 0x02}
	current := &playerObserverGoodSlaveTestObject4EC420{name: "current", owner: owner, nextOwned: monster}
	events := make([]string, 0, 6)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC420()
	hooks.loadOwner = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "owner:"+obj.name)
		return obj.owner
	}
	hooks.loadNextOwned = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestObject4EC420 {
		events = append(events, "next:"+obj.name)
		return obj.nextOwned
	}
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC420) uint8 {
		events = append(events, "class:"+obj.name)
		return obj.class
	}
	hooks.loadUpdateData = func(obj *playerObserverGoodSlaveTestObject4EC420) *playerObserverGoodSlaveTestData4EC420 {
		events = append(events, "data:"+obj.name)
		return nil
	}
	hooks.loadStatusByte = func(data *playerObserverGoodSlaveTestData4EC420) uint8 {
		events = append(events, "status")
		return data.status
	}

	defer func() {
		if recover() == nil {
			t.Fatal("nil Monster update data returned without fault")
		}
		want := []string{"owner:current", "next:current", "class:monster", "data:monster", "status"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	playerObserverFindGoodSlaveContract4EC420(current, hooks)
}
