package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectFromNetCodeTestObject4ECCB0 struct {
	name  string
	flags uint8
	code  uint32
	next  *objectFromNetCodeTestObject4ECCB0
	item  *objectFromNetCodeTestObject4ECCB0
}

type objectFromNetCodeTestPlayer4ECCB0 struct {
	name string
	next *objectFromNetCodeTestPlayer4ECCB0
	unit *objectFromNetCodeTestObject4ECCB0
}

func TestObjectFromNetCode4ECCB0CacheHitStopsSearch(t *testing.T) {
	cached := &objectFromNetCodeTestObject4ECCB0{name: "cached", flags: objectDeadFlagLow4ECCB0}
	var events []string
	got := objectFromNetCode4ECCB0(uint32(0xfedcba98), objectFromNetCodeHooks4ECCB0[
		*objectFromNetCodeTestObject4ECCB0,
		*objectFromNetCodeTestPlayer4ECCB0,
	]{
		cacheLookup: func(code uint32) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, fmt.Sprintf("cache:%08x", code))
			return cached
		},
		cacheAdd:    func(*objectFromNetCodeTestObject4ECCB0) { t.Fatal("cache hit was added again") },
		firstObject: func() *objectFromNetCodeTestObject4ECCB0 { t.Fatal("cache hit scanned objects"); return nil },
		nextObject: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("cache hit advanced objects")
			return nil
		},
		firstItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("cache hit scanned inventory")
			return nil
		},
		nextItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("cache hit advanced inventory")
			return nil
		},
		firstPending: func() *objectFromNetCodeTestObject4ECCB0 { t.Fatal("cache hit scanned pending objects"); return nil },
		nextPending: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("cache hit advanced pending objects")
			return nil
		},
		firstPlayer: func() *objectFromNetCodeTestPlayer4ECCB0 { t.Fatal("cache hit scanned players"); return nil },
		nextPlayer: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestPlayer4ECCB0 {
			t.Fatal("cache hit advanced players")
			return nil
		},
		loadPlayerUnit: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("cache hit loaded a player unit")
			return nil
		},
		loadFlagsLow: func(*objectFromNetCodeTestObject4ECCB0) uint8 { t.Fatal("cache hit loaded flags"); return 0 },
		loadNetCode:  func(*objectFromNetCodeTestObject4ECCB0) uint32 { t.Fatal("cache hit loaded a network code"); return 0 },
	})
	if got != cached {
		t.Fatalf("result = %p, want cached %p", got, cached)
	}
	want := []string{"cache:fedcba98"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectFromNetCode4ECCB0SearchOrderAndPlayerLiveReload(t *testing.T) {
	const wanted = uint32(0x80000021)
	deadTop := &objectFromNetCodeTestObject4ECCB0{name: "dead-top", flags: objectDeadFlagLow4ECCB0, code: wanted}
	deadItem := &objectFromNetCodeTestObject4ECCB0{name: "dead-item", flags: objectDeadFlagLow4ECCB0, code: wanted}
	otherItem := &objectFromNetCodeTestObject4ECCB0{name: "other-item", code: wanted + 1}
	deadItem.next = otherItem
	deadTop.item = deadItem
	otherTop := &objectFromNetCodeTestObject4ECCB0{name: "other-top", code: wanted + 2}
	deadTop.next = otherTop

	deadPending := &objectFromNetCodeTestObject4ECCB0{name: "dead-pending", flags: objectDeadFlagLow4ECCB0, code: wanted}
	otherPending := &objectFromNetCodeTestObject4ECCB0{name: "other-pending", code: wanted + 3}
	deadPending.next = otherPending

	otherUnit := &objectFromNetCodeTestObject4ECCB0{name: "other-unit", code: wanted + 4}
	matchedUnit := &objectFromNetCodeTestObject4ECCB0{name: "matched-unit", code: wanted}
	liveUnit := &objectFromNetCodeTestObject4ECCB0{name: "live-unit", code: 7}
	firstPlayer := &objectFromNetCodeTestPlayer4ECCB0{name: "nil-player"}
	otherPlayer := &objectFromNetCodeTestPlayer4ECCB0{name: "other-player", unit: otherUnit}
	matchedPlayer := &objectFromNetCodeTestPlayer4ECCB0{name: "matched-player", unit: matchedUnit}
	firstPlayer.next = otherPlayer
	otherPlayer.next = matchedPlayer

	var events []string
	matchedLoads := 0
	hooks := objectFromNetCodeHooks4ECCB0[
		*objectFromNetCodeTestObject4ECCB0,
		*objectFromNetCodeTestPlayer4ECCB0,
	]{
		cacheLookup: func(code uint32) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, fmt.Sprintf("cache:%08x", code))
			return nil
		},
		cacheAdd: func(obj *objectFromNetCodeTestObject4ECCB0) {
			events = append(events, "add:"+obj.name)
		},
		firstObject: func() *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "first-object")
			return deadTop
		},
		nextObject: func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "next-object:"+obj.name)
			return obj.next
		},
		firstItem: func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "first-item:"+obj.name)
			return obj.item
		},
		nextItem: func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "next-item:"+obj.name)
			return obj.next
		},
		firstPending: func() *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "first-pending")
			return deadPending
		},
		nextPending: func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "next-pending:"+obj.name)
			return obj.next
		},
		firstPlayer: func() *objectFromNetCodeTestPlayer4ECCB0 {
			events = append(events, "first-player")
			return firstPlayer
		},
		nextPlayer: func(player *objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestPlayer4ECCB0 {
			events = append(events, "next-player:"+player.name)
			return player.next
		},
		loadPlayerUnit: func(player *objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "unit:"+player.name)
			if player == matchedPlayer {
				matchedLoads++
				if matchedLoads == 2 {
					return liveUnit
				}
			}
			return player.unit
		},
		loadFlagsLow: func(obj *objectFromNetCodeTestObject4ECCB0) uint8 {
			events = append(events, "flags:"+obj.name)
			return obj.flags
		},
		loadNetCode: func(obj *objectFromNetCodeTestObject4ECCB0) uint32 {
			events = append(events, "net:"+obj.name)
			return obj.code
		},
	}

	got := objectFromNetCode4ECCB0(wanted, hooks)
	if got != liveUnit {
		t.Fatalf("result = %p, want live reloaded unit %p", got, liveUnit)
	}
	want := []string{
		"cache:80000021",
		"first-object",
		"flags:dead-top",
		"first-item:dead-top",
		"flags:dead-item",
		"next-item:dead-item",
		"flags:other-item",
		"net:other-item",
		"next-item:other-item",
		"next-object:dead-top",
		"flags:other-top",
		"net:other-top",
		"first-item:other-top",
		"next-object:other-top",
		"first-pending",
		"flags:dead-pending",
		"next-pending:dead-pending",
		"flags:other-pending",
		"net:other-pending",
		"next-pending:other-pending",
		"first-player",
		"unit:nil-player",
		"next-player:nil-player",
		"unit:other-player",
		"flags:other-unit",
		"net:other-unit",
		"next-player:other-player",
		"unit:matched-player",
		"flags:matched-unit",
		"net:matched-unit",
		"unit:matched-player",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
	for _, event := range events {
		if len(event) >= 4 && event[:4] == "add:" {
			t.Fatalf("player-unit match unexpectedly cached: %v", events)
		}
	}
}

func TestObjectFromNetCode4ECCB0CachesMatchesFromObjectDomains(t *testing.T) {
	tests := []struct {
		name    string
		objects *objectFromNetCodeTestObject4ECCB0
		pending *objectFromNetCodeTestObject4ECCB0
		want    *objectFromNetCodeTestObject4ECCB0
	}{
		{
			name:    "top-level",
			objects: &objectFromNetCodeTestObject4ECCB0{name: "top", code: 9},
		},
		{
			name: "inventory",
			objects: &objectFromNetCodeTestObject4ECCB0{
				name: "owner",
				code: 8,
				item: &objectFromNetCodeTestObject4ECCB0{name: "item", code: 9},
			},
		},
		{
			name:    "pending",
			pending: &objectFromNetCodeTestObject4ECCB0{name: "pending", code: 9},
		},
	}
	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == nil {
				switch tc.name {
				case "top-level":
					tc.want = tc.objects
				case "inventory":
					tc.want = tc.objects.item
				case "pending":
					tc.want = tc.pending
				}
			}
			var added []*objectFromNetCodeTestObject4ECCB0
			hooks := objectFromNetCodeHooks4ECCB0[
				*objectFromNetCodeTestObject4ECCB0,
				*objectFromNetCodeTestPlayer4ECCB0,
			]{
				cacheLookup:  func(uint32) *objectFromNetCodeTestObject4ECCB0 { return nil },
				cacheAdd:     func(obj *objectFromNetCodeTestObject4ECCB0) { added = append(added, obj) },
				firstObject:  func() *objectFromNetCodeTestObject4ECCB0 { return tc.objects },
				nextObject:   func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 { return obj.next },
				firstItem:    func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 { return obj.item },
				nextItem:     func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 { return obj.next },
				firstPending: func() *objectFromNetCodeTestObject4ECCB0 { return tc.pending },
				nextPending:  func(obj *objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 { return obj.next },
				firstPlayer:  func() *objectFromNetCodeTestPlayer4ECCB0 { return nil },
				nextPlayer:   func(player *objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestPlayer4ECCB0 { return player.next },
				loadPlayerUnit: func(player *objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
					return player.unit
				},
				loadFlagsLow: func(obj *objectFromNetCodeTestObject4ECCB0) uint8 { return obj.flags },
				loadNetCode:  func(obj *objectFromNetCodeTestObject4ECCB0) uint32 { return obj.code },
			}
			got := objectFromNetCode4ECCB0(uint32(9), hooks)
			if got != tc.want {
				t.Fatalf("result = %p, want %p", got, tc.want)
			}
			if !reflect.DeepEqual(added, []*objectFromNetCodeTestObject4ECCB0{tc.want}) {
				t.Fatalf("cache adds = %v, want [%p]", added, tc.want)
			}
		})
	}
}

func TestObjectFromNetCode4ECCB0EmptyDomainsReturnNilInOrder(t *testing.T) {
	var events []string
	got := objectFromNetCode4ECCB0(uint32(17), objectFromNetCodeHooks4ECCB0[
		*objectFromNetCodeTestObject4ECCB0,
		*objectFromNetCodeTestPlayer4ECCB0,
	]{
		cacheLookup: func(uint32) *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "cache")
			return nil
		},
		cacheAdd: func(*objectFromNetCodeTestObject4ECCB0) { t.Fatal("empty search added to cache") },
		firstObject: func() *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "objects")
			return nil
		},
		nextObject: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("empty objects advanced")
			return nil
		},
		firstItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("empty objects loaded inventory")
			return nil
		},
		nextItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("empty inventory advanced")
			return nil
		},
		firstPending: func() *objectFromNetCodeTestObject4ECCB0 {
			events = append(events, "pending")
			return nil
		},
		nextPending: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("empty pending objects advanced")
			return nil
		},
		firstPlayer: func() *objectFromNetCodeTestPlayer4ECCB0 {
			events = append(events, "players")
			return nil
		},
		nextPlayer: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestPlayer4ECCB0 {
			t.Fatal("empty players advanced")
			return nil
		},
		loadPlayerUnit: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("empty players loaded a unit")
			return nil
		},
		loadFlagsLow: func(*objectFromNetCodeTestObject4ECCB0) uint8 { t.Fatal("empty search loaded flags"); return 0 },
		loadNetCode: func(*objectFromNetCodeTestObject4ECCB0) uint32 {
			t.Fatal("empty search loaded a network code")
			return 0
		},
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	want := []string{"cache", "objects", "pending", "players"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectFromNetCode4ECCB0FullWidthCodeAndCacheAddFault(t *testing.T) {
	obj := &objectFromNetCodeTestObject4ECCB0{name: "object", code: 0xffffffff}
	const fault = "cache add fault"
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
	}()
	objectFromNetCode4ECCB0(^uint32(0), objectFromNetCodeHooks4ECCB0[
		*objectFromNetCodeTestObject4ECCB0,
		*objectFromNetCodeTestPlayer4ECCB0,
	]{
		cacheLookup: func(uint32) *objectFromNetCodeTestObject4ECCB0 { return nil },
		cacheAdd: func(got *objectFromNetCodeTestObject4ECCB0) {
			if got != obj {
				t.Fatalf("cache object = %p", got)
			}
			panic(fault)
		},
		firstObject: func() *objectFromNetCodeTestObject4ECCB0 { return obj },
		nextObject: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("match advanced objects")
			return nil
		},
		firstItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("match loaded inventory")
			return nil
		},
		nextItem: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("match advanced inventory")
			return nil
		},
		firstPending: func() *objectFromNetCodeTestObject4ECCB0 { t.Fatal("match scanned pending objects"); return nil },
		nextPending: func(*objectFromNetCodeTestObject4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("match advanced pending objects")
			return nil
		},
		firstPlayer: func() *objectFromNetCodeTestPlayer4ECCB0 { t.Fatal("match scanned players"); return nil },
		nextPlayer: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestPlayer4ECCB0 {
			t.Fatal("match advanced players")
			return nil
		},
		loadPlayerUnit: func(*objectFromNetCodeTestPlayer4ECCB0) *objectFromNetCodeTestObject4ECCB0 {
			t.Fatal("match loaded player unit")
			return nil
		},
		loadFlagsLow: func(got *objectFromNetCodeTestObject4ECCB0) uint8 {
			if got != obj {
				t.Fatalf("flags object = %p", got)
			}
			return 0
		},
		loadNetCode: func(got *objectFromNetCodeTestObject4ECCB0) uint32 {
			if got != obj {
				t.Fatalf("net object = %p", got)
			}
			return got.code
		},
	})
}
