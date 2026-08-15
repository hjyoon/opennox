package server

import (
	"reflect"
	"testing"
)

type unitClearOwnerTestPlayer4EC300 struct {
	name  string
	index uint8
}

type unitClearOwnerTestData4EC300 struct {
	name   string
	player *unitClearOwnerTestPlayer4EC300
}

type unitClearOwnerTestObject4EC300 struct {
	name       string
	class      uint32
	subClass   uint32
	owner      *unitClearOwnerTestObject4EC300
	nextOwned  *unitClearOwnerTestObject4EC300
	firstOwned *unitClearOwnerTestObject4EC300
	data       *unitClearOwnerTestData4EC300
}

func defaultUnitClearOwnerHooks4EC300() unitClearOwnerHooks4EC300[
	*unitClearOwnerTestObject4EC300,
	*unitClearOwnerTestData4EC300,
	*unitClearOwnerTestPlayer4EC300,
] {
	return unitClearOwnerHooks4EC300[
		*unitClearOwnerTestObject4EC300,
		*unitClearOwnerTestData4EC300,
		*unitClearOwnerTestPlayer4EC300,
	]{
		loadOwner: func(obj *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
			return obj.owner
		},
		loadClass: func(obj *unitClearOwnerTestObject4EC300) uint32 {
			return obj.class
		},
		isMonitored: func(*unitClearOwnerTestObject4EC300, *unitClearOwnerTestObject4EC300) bool {
			return false
		},
		loadSubClass: func(obj *unitClearOwnerTestObject4EC300) uint32 {
			return obj.subClass
		},
		loadPlayerData: func(owner *unitClearOwnerTestObject4EC300) *unitClearOwnerTestData4EC300 {
			return owner.data
		},
		storeSubClass: func(obj *unitClearOwnerTestObject4EC300, subClass uint32) {
			obj.subClass = subClass
		},
		loadPlayer: func(data *unitClearOwnerTestData4EC300) *unitClearOwnerTestPlayer4EC300 {
			return data.player
		},
		loadPlayerIndex: func(player *unitClearOwnerTestPlayer4EC300) uint8 {
			return player.index
		},
		netFxShield:   func(uint8, *unitClearOwnerTestObject4EC300) {},
		unmarkMinimap: func(uint8, *unitClearOwnerTestObject4EC300, uint32) {},
		loadFirstOwned: func(owner *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
			return owner.firstOwned
		},
		loadNextOwned: func(obj *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
			return obj.nextOwned
		},
		storeNextOwned: func(obj, next *unitClearOwnerTestObject4EC300) {
			obj.nextOwned = next
		},
		storeFirstOwned: func(owner, first *unitClearOwnerTestObject4EC300) {
			owner.firstOwned = first
		},
		storeOwner: func(obj, owner *unitClearOwnerTestObject4EC300) {
			obj.owner = owner
		},
		resetMonster:   func(*unitClearOwnerTestObject4EC300) {},
		markUnitUpdate: func(*unitClearOwnerTestObject4EC300) {},
	}
}

func TestUnitClearOwner4EC300NilAndUnownedReturnAtExactLoads(t *testing.T) {
	events := make([]string, 0, 1)
	hooks := defaultUnitClearOwnerHooks4EC300()
	hooks.loadOwner = func(obj *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "owner:"+obj.name)
		return obj.owner
	}
	hooks.loadClass = func(*unitClearOwnerTestObject4EC300) uint32 {
		t.Fatal("early-return path loaded class")
		return 0
	}

	unitClearOwner4EC300(nil, hooks)
	if len(events) != 0 {
		t.Fatalf("nil object events = %v", events)
	}

	obj := &unitClearOwnerTestObject4EC300{name: "object"}
	unitClearOwner4EC300(obj, hooks)
	if !reflect.DeepEqual(events, []string{"owner:object"}) {
		t.Fatalf("unowned events = %v", events)
	}
}

func TestUnitClearOwner4EC300MonitoredPathUsesLiveOwnersAndPlayers(t *testing.T) {
	firstPlayer := &unitClearOwnerTestPlayer4EC300{name: "first", index: 0xfe}
	secondPlayer := &unitClearOwnerTestPlayer4EC300{name: "second", index: 0x81}
	data := &unitClearOwnerTestData4EC300{name: "data", player: firstPlayer}
	entryOwner := &unitClearOwnerTestObject4EC300{name: "entry-owner", class: 0x80000004}
	notifyOwner := &unitClearOwnerTestObject4EC300{name: "notify-owner", data: data}
	next := &unitClearOwnerTestObject4EC300{name: "next"}
	listOwner := &unitClearOwnerTestObject4EC300{name: "list-owner"}
	obj := &unitClearOwnerTestObject4EC300{
		name:      "object",
		class:     0x80000002,
		subClass:  0xaabbccff,
		owner:     entryOwner,
		nextOwned: next,
	}
	listOwner.firstOwned = obj
	events := make([]string, 0, 32)
	hooks := defaultUnitClearOwnerHooks4EC300()
	hooks.loadOwner = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "owner:"+got.name)
		return got.owner
	}
	hooks.loadClass = func(got *unitClearOwnerTestObject4EC300) uint32 {
		events = append(events, "class:"+got.name)
		return got.class
	}
	hooks.isMonitored = func(gotOwner, gotObj *unitClearOwnerTestObject4EC300) bool {
		events = append(events, "monitored:"+gotOwner.name+":"+gotObj.name)
		gotObj.owner = notifyOwner
		return true
	}
	hooks.loadSubClass = func(got *unitClearOwnerTestObject4EC300) uint32 {
		events = append(events, "subclass:"+got.name)
		return got.subClass
	}
	hooks.loadPlayerData = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestData4EC300 {
		events = append(events, "data:"+got.name)
		return got.data
	}
	hooks.storeSubClass = func(got *unitClearOwnerTestObject4EC300, subClass uint32) {
		events = append(events, "store-subclass")
		got.subClass = subClass
	}
	hooks.loadPlayer = func(got *unitClearOwnerTestData4EC300) *unitClearOwnerTestPlayer4EC300 {
		events = append(events, "player:"+got.name+":"+got.player.name)
		return got.player
	}
	hooks.loadPlayerIndex = func(got *unitClearOwnerTestPlayer4EC300) uint8 {
		events = append(events, "index:"+got.name)
		return got.index
	}
	hooks.netFxShield = func(index uint8, got *unitClearOwnerTestObject4EC300) {
		events = append(events, "shield")
		if index != firstPlayer.index || got != obj {
			t.Fatalf("shield args = %#x/%p", index, got)
		}
		data.player = secondPlayer
	}
	hooks.unmarkMinimap = func(index uint8, got *unitClearOwnerTestObject4EC300, flags uint32) {
		events = append(events, "unmark")
		if index != secondPlayer.index || got != obj || flags != 1 {
			t.Fatalf("unmark args = %#x/%p/%d", index, got, flags)
		}
		got.owner = listOwner
	}
	hooks.loadFirstOwned = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "first:"+got.name)
		return got.firstOwned
	}
	hooks.loadNextOwned = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "next:"+got.name)
		return got.nextOwned
	}
	hooks.storeFirstOwned = func(got, first *unitClearOwnerTestObject4EC300) {
		events = append(events, "store-first:"+got.name)
		got.firstOwned = first
	}
	hooks.storeOwner = func(got, owner *unitClearOwnerTestObject4EC300) {
		events = append(events, "store-owner")
		got.owner = owner
	}
	hooks.resetMonster = func(got *unitClearOwnerTestObject4EC300) {
		events = append(events, "reset")
		got.class = 0x04
	}
	hooks.markUnitUpdate = func(got *unitClearOwnerTestObject4EC300) {
		events = append(events, "mark")
		if got != obj {
			t.Fatalf("mark object = %p", got)
		}
	}

	unitClearOwner4EC300(obj, hooks)
	wantEvents := []string{
		"owner:object", "class:entry-owner", "monitored:entry-owner:object",
		"owner:object", "subclass:object", "data:notify-owner", "store-subclass",
		"player:data:first", "index:first", "shield",
		"player:data:second", "index:second", "unmark",
		"owner:object", "first:list-owner", "next:object", "store-first:list-owner",
		"class:object", "store-owner", "reset", "class:object", "mark",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if obj.subClass != 0xaabbcc7f {
		t.Fatalf("subclass = %#x", obj.subClass)
	}
	if obj.owner != nil || obj.nextOwned != next || listOwner.firstOwned != next {
		t.Fatalf("ownership = owner %p next %p head %p", obj.owner, obj.nextOwned, listOwner.firstOwned)
	}
}

func TestUnitClearOwner4EC300TraversesToPredecessorAndUsesLiveObjectNext(t *testing.T) {
	tail := &unitClearOwnerTestObject4EC300{name: "tail"}
	obj := &unitClearOwnerTestObject4EC300{name: "object", nextOwned: tail}
	second := &unitClearOwnerTestObject4EC300{name: "second"}
	first := &unitClearOwnerTestObject4EC300{name: "first", nextOwned: second}
	owner := &unitClearOwnerTestObject4EC300{name: "owner", firstOwned: first}
	obj.owner = owner
	events := make([]string, 0, 12)
	hooks := defaultUnitClearOwnerHooks4EC300()
	hooks.loadOwner = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "owner")
		return got.owner
	}
	hooks.loadClass = func(got *unitClearOwnerTestObject4EC300) uint32 {
		events = append(events, "class:"+got.name)
		return got.class
	}
	hooks.loadFirstOwned = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "first")
		return got.firstOwned
	}
	hooks.loadNextOwned = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "next:"+got.name)
		if got == second {
			return nil
		}
		return got.nextOwned
	}
	hooks.storeNextOwned = func(got, next *unitClearOwnerTestObject4EC300) {
		events = append(events, "store-next:"+got.name)
		got.nextOwned = next
	}
	hooks.storeOwner = func(got, gotOwner *unitClearOwnerTestObject4EC300) {
		events = append(events, "store-owner")
		got.owner = gotOwner
	}

	unitClearOwner4EC300(obj, hooks)
	wantEvents := []string{
		"owner", "class:owner", "owner", "first", "next:first", "next:second",
		"next:object", "store-next:second", "class:object", "store-owner", "class:object",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if second.nextOwned != tail || obj.nextOwned != tail || obj.owner != nil {
		t.Fatalf("links = second.next %p obj.next %p owner %p", second.nextOwned, obj.nextOwned, obj.owner)
	}
}

func TestUnitClearOwner4EC300MonitoredNilLiveOwnerFaultsAfterSubclassLoad(t *testing.T) {
	entryOwner := &unitClearOwnerTestObject4EC300{name: "entry-owner", class: 0x04}
	obj := &unitClearOwnerTestObject4EC300{name: "object", owner: entryOwner, subClass: 0xff}
	events := make([]string, 0, 8)
	hooks := defaultUnitClearOwnerHooks4EC300()
	hooks.loadOwner = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestObject4EC300 {
		events = append(events, "owner")
		return got.owner
	}
	hooks.loadClass = func(got *unitClearOwnerTestObject4EC300) uint32 {
		events = append(events, "class")
		return got.class
	}
	hooks.isMonitored = func(*unitClearOwnerTestObject4EC300, *unitClearOwnerTestObject4EC300) bool {
		events = append(events, "monitored")
		obj.owner = nil
		return true
	}
	hooks.loadSubClass = func(got *unitClearOwnerTestObject4EC300) uint32 {
		events = append(events, "subclass")
		return got.subClass
	}
	hooks.loadPlayerData = func(got *unitClearOwnerTestObject4EC300) *unitClearOwnerTestData4EC300 {
		events = append(events, "data")
		if got == nil {
			panic("nil live owner")
		}
		return got.data
	}
	hooks.storeSubClass = func(*unitClearOwnerTestObject4EC300, uint32) {
		events = append(events, "store-subclass")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected nil live-owner fault")
		}
		want := []string{"owner", "class", "monitored", "owner", "subclass", "data"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
		if obj.subClass != 0xff {
			t.Fatalf("subclass changed before owner update-data fault: %#x", obj.subClass)
		}
	}()
	unitClearOwner4EC300(obj, hooks)
}
