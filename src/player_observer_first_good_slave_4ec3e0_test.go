package opennox

import (
	"reflect"
	"testing"
)

type playerObserverGoodSlaveTestData4EC3E0 struct {
	name   string
	status uint8
}

type playerObserverGoodSlaveTestObject4EC3E0 struct {
	name       string
	class      uint8
	firstOwned *playerObserverGoodSlaveTestObject4EC3E0
	nextOwned  *playerObserverGoodSlaveTestObject4EC3E0
	data       *playerObserverGoodSlaveTestData4EC3E0
}

func defaultPlayerObserverGoodSlaveHooks4EC3E0() playerObserverFindGoodSlave2Hooks4EC3E0[
	*playerObserverGoodSlaveTestObject4EC3E0,
	*playerObserverGoodSlaveTestData4EC3E0,
] {
	return playerObserverFindGoodSlave2Hooks4EC3E0[
		*playerObserverGoodSlaveTestObject4EC3E0,
		*playerObserverGoodSlaveTestData4EC3E0,
	]{
		loadFirstOwned: func(owner *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
			return owner.firstOwned
		},
		loadClassByte: func(candidate *playerObserverGoodSlaveTestObject4EC3E0) uint8 {
			return candidate.class
		},
		loadUpdateData: func(candidate *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestData4EC3E0 {
			return candidate.data
		},
		loadStatusByte: func(data *playerObserverGoodSlaveTestData4EC3E0) uint8 {
			return data.status
		},
		loadNextOwned: func(candidate *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
			return candidate.nextOwned
		},
	}
}

func TestPlayerObserverFindGoodSlave2Contract4EC3E0NilAndEmptyExactLoads(t *testing.T) {
	events := make([]string, 0, 1)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC3E0()
	hooks.loadFirstOwned = func(owner *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
		events = append(events, "first:"+owner.name)
		return owner.firstOwned
	}
	hooks.loadClassByte = func(*playerObserverGoodSlaveTestObject4EC3E0) uint8 {
		t.Fatal("early-return path loaded a candidate class")
		return 0
	}

	if got := playerObserverFindGoodSlave2Contract4EC3E0(nil, hooks); got != nil {
		t.Fatalf("nil owner result = %p", got)
	}
	if len(events) != 0 {
		t.Fatalf("nil owner events = %v", events)
	}

	owner := &playerObserverGoodSlaveTestObject4EC3E0{name: "owner"}
	if got := playerObserverFindGoodSlave2Contract4EC3E0(owner, hooks); got != nil {
		t.Fatalf("empty owner result = %p", got)
	}
	if !reflect.DeepEqual(events, []string{"first:owner"}) {
		t.Fatalf("empty owner events = %v", events)
	}
}

func TestPlayerObserverFindGoodSlave2Contract4EC3E0TraversalOrder(t *testing.T) {
	nonMonster := &playerObserverGoodSlaveTestObject4EC3E0{name: "non-monster", class: 0xfc}
	dormantData := &playerObserverGoodSlaveTestData4EC3E0{name: "dormant-data", status: 0x7f}
	dormant := &playerObserverGoodSlaveTestObject4EC3E0{
		name:  "dormant",
		class: 0x82,
		data:  dormantData,
	}
	summonedData := &playerObserverGoodSlaveTestData4EC3E0{name: "summoned-data", status: 0x81}
	summoned := &playerObserverGoodSlaveTestObject4EC3E0{
		name:  "summoned",
		class: 0x02,
		data:  summonedData,
	}
	ignored := &playerObserverGoodSlaveTestObject4EC3E0{name: "ignored"}
	owner := &playerObserverGoodSlaveTestObject4EC3E0{name: "owner", firstOwned: nonMonster}
	nonMonster.nextOwned = dormant
	dormant.nextOwned = summoned
	summoned.nextOwned = ignored

	events := make([]string, 0, 12)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC3E0()
	hooks.loadFirstOwned = func(obj *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
		events = append(events, "first:"+obj.name)
		return obj.firstOwned
	}
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC3E0) uint8 {
		events = append(events, "class:"+obj.name)
		return obj.class
	}
	hooks.loadUpdateData = func(obj *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestData4EC3E0 {
		events = append(events, "data:"+obj.name)
		return obj.data
	}
	hooks.loadStatusByte = func(data *playerObserverGoodSlaveTestData4EC3E0) uint8 {
		events = append(events, "status:"+data.name)
		return data.status
	}
	hooks.loadNextOwned = func(obj *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
		events = append(events, "next:"+obj.name)
		return obj.nextOwned
	}

	if got := playerObserverFindGoodSlave2Contract4EC3E0(owner, hooks); got != summoned {
		t.Fatalf("result = %p, want %p", got, summoned)
	}
	want := []string{
		"first:owner",
		"class:non-monster", "next:non-monster",
		"class:dormant", "data:dormant", "status:dormant-data", "next:dormant",
		"class:summoned", "data:summoned", "status:summoned-data",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPlayerObserverFindGoodSlave2Contract4EC3E0UsesLiveNext(t *testing.T) {
	first := &playerObserverGoodSlaveTestObject4EC3E0{name: "first"}
	stale := &playerObserverGoodSlaveTestObject4EC3E0{name: "stale"}
	liveData := &playerObserverGoodSlaveTestData4EC3E0{name: "live-data", status: 0x80}
	live := &playerObserverGoodSlaveTestObject4EC3E0{name: "live", class: 0x02, data: liveData}
	owner := &playerObserverGoodSlaveTestObject4EC3E0{name: "owner", firstOwned: first}
	first.nextOwned = stale

	hooks := defaultPlayerObserverGoodSlaveHooks4EC3E0()
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC3E0) uint8 {
		if obj == first {
			first.nextOwned = live
		}
		return obj.class
	}
	if got := playerObserverFindGoodSlave2Contract4EC3E0(owner, hooks); got != live {
		t.Fatalf("result = %p, want live successor %p", got, live)
	}
}

func TestPlayerObserverFindGoodSlave2Contract4EC3E0NilMonsterDataFaultsBeforeNext(t *testing.T) {
	monster := &playerObserverGoodSlaveTestObject4EC3E0{name: "monster", class: 0x02}
	owner := &playerObserverGoodSlaveTestObject4EC3E0{name: "owner", firstOwned: monster}
	events := make([]string, 0, 4)
	hooks := defaultPlayerObserverGoodSlaveHooks4EC3E0()
	hooks.loadClassByte = func(obj *playerObserverGoodSlaveTestObject4EC3E0) uint8 {
		events = append(events, "class:"+obj.name)
		return obj.class
	}
	hooks.loadUpdateData = func(obj *playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestData4EC3E0 {
		events = append(events, "data:"+obj.name)
		return nil
	}
	hooks.loadStatusByte = func(data *playerObserverGoodSlaveTestData4EC3E0) uint8 {
		events = append(events, "status")
		return data.status
	}
	hooks.loadNextOwned = func(*playerObserverGoodSlaveTestObject4EC3E0) *playerObserverGoodSlaveTestObject4EC3E0 {
		events = append(events, "next")
		return nil
	}

	defer func() {
		if recover() == nil {
			t.Fatal("nil Monster update data returned without fault")
		}
		want := []string{"class:monster", "data:monster", "status"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	playerObserverFindGoodSlave2Contract4EC3E0(owner, hooks)
}
