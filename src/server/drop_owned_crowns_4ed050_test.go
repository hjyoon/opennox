package server

import (
	"reflect"
	"testing"
)

type dropOwnedCrownsTestData4ED050 struct {
	name         string
	pickupTarget *dropOwnedCrownsTestObject4ED050
}

type dropOwnedCrownsTestObject4ED050 struct {
	name      string
	typeIndex uint16
	first     *dropOwnedCrownsTestObject4ED050
	next      *dropOwnedCrownsTestObject4ED050
	update    *dropOwnedCrownsTestData4ED050
}

type dropOwnedCrownsTestWorld4ED050 struct {
	cache       uint32
	lookup      uint32
	owner       *dropOwnedCrownsTestObject4ED050
	target      *dropOwnedCrownsTestObject4ED050
	events      []string
	faultAt     int
	firstMatch  *dropOwnedCrownsTestObject4ED050
	secondMatch *dropOwnedCrownsTestObject4ED050
	replacement *dropOwnedCrownsTestData4ED050
}

func (w *dropOwnedCrownsTestWorld4ED050) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func dropOwnedCrownsObjectName4ED050(obj *dropOwnedCrownsTestObject4ED050) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *dropOwnedCrownsTestWorld4ED050) hooks() dropOwnedCrownsHooks4ED050[
	*dropOwnedCrownsTestObject4ED050,
	*dropOwnedCrownsTestData4ED050,
	string,
] {
	return dropOwnedCrownsHooks4ED050[
		*dropOwnedCrownsTestObject4ED050,
		*dropOwnedCrownsTestData4ED050,
		string,
	]{
		loadCrownTypeCache: func() uint32 {
			w.event("cache")
			return w.cache
		},
		lookupCrownType: func() uint32 {
			w.event("lookup:Crown")
			return w.lookup
		},
		storeCrownTypeCache: func(value uint32) {
			w.event("store-cache")
			w.cache = value
		},
		loadOwnerArg: func() *dropOwnedCrownsTestObject4ED050 {
			w.event("owner-arg")
			return w.owner
		},
		firstOwned: func(owner *dropOwnedCrownsTestObject4ED050) *dropOwnedCrownsTestObject4ED050 {
			w.event("first:" + dropOwnedCrownsObjectName4ED050(owner))
			return owner.first
		},
		loadTargetArg: func() *dropOwnedCrownsTestObject4ED050 {
			w.event("target-arg")
			return w.target
		},
		loadTypeIndex: func(item *dropOwnedCrownsTestObject4ED050) uint16 {
			w.event("type:" + item.name)
			return item.typeIndex
		},
		loadUpdate: func(item *dropOwnedCrownsTestObject4ED050) *dropOwnedCrownsTestData4ED050 {
			w.event("update:" + item.name)
			return item.update
		},
		ownerPosition: func(owner *dropOwnedCrownsTestObject4ED050) string {
			w.event("position:" + owner.name)
			return "position-of-" + owner.name
		},
		dropCrown: func(
			owner, item *dropOwnedCrownsTestObject4ED050,
			position string,
		) uint32 {
			w.event("drop:" + owner.name + ":" + item.name + ":" + position)
			if item == w.firstMatch {
				item.update = w.replacement
				item.next = w.secondMatch
				w.cache = uint32(w.secondMatch.typeIndex)
			}
			return 0xf1234567
		},
		storePickupTarget: func(
			update *dropOwnedCrownsTestData4ED050,
			target *dropOwnedCrownsTestObject4ED050,
		) {
			w.event("store-target:" + update.name + ":" + dropOwnedCrownsObjectName4ED050(target))
			update.pickupTarget = target
		},
		nextOwned: func(item *dropOwnedCrownsTestObject4ED050) *dropOwnedCrownsTestObject4ED050 {
			w.event("next:" + item.name)
			return item.next
		},
	}
}

func newDropOwnedCrownsFullTrace4ED050() (
	*dropOwnedCrownsTestWorld4ED050,
	*dropOwnedCrownsTestData4ED050,
	*dropOwnedCrownsTestData4ED050,
	[]string,
) {
	owner := &dropOwnedCrownsTestObject4ED050{name: "owner"}
	target := &dropOwnedCrownsTestObject4ED050{name: "target"}
	other := &dropOwnedCrownsTestObject4ED050{name: "other", typeIndex: 6}
	firstData := &dropOwnedCrownsTestData4ED050{name: "first-data"}
	firstMatch := &dropOwnedCrownsTestObject4ED050{name: "first-match", typeIndex: 7, update: firstData}
	stale := &dropOwnedCrownsTestObject4ED050{name: "stale", typeIndex: 7}
	secondData := &dropOwnedCrownsTestData4ED050{name: "second-data"}
	secondMatch := &dropOwnedCrownsTestObject4ED050{name: "second-match", typeIndex: 9, update: secondData}
	replacement := &dropOwnedCrownsTestData4ED050{name: "replacement"}
	other.next = firstMatch
	firstMatch.next = stale
	owner.first = other
	w := &dropOwnedCrownsTestWorld4ED050{
		lookup:      7,
		owner:       owner,
		target:      target,
		firstMatch:  firstMatch,
		secondMatch: secondMatch,
		replacement: replacement,
	}
	wantEvents := []string{
		"cache",
		"lookup:Crown",
		"store-cache",
		"owner-arg",
		"first:owner",
		"target-arg",
		"cache",
		"type:other",
		"next:other",
		"cache",
		"type:first-match",
		"update:first-match",
		"position:owner",
		"drop:owner:first-match:position-of-owner",
		"store-target:first-data:target",
		"next:first-match",
		"cache",
		"type:second-match",
		"update:second-match",
		"position:owner",
		"drop:owner:second-match:position-of-owner",
		"store-target:second-data:target",
		"next:second-match",
	}
	return w, firstData, secondData, wantEvents
}

func TestDropOwnedCrowns4ED050ExactTrace(t *testing.T) {
	w, firstData, secondData, wantEvents := newDropOwnedCrownsFullTrace4ED050()
	dropOwnedCrowns4ED050(w.hooks())
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, wantEvents)
	}
	if firstData.pickupTarget != w.target || secondData.pickupTarget != w.target {
		t.Fatalf("cached update targets = (%p,%p), want %p", firstData.pickupTarget, secondData.pickupTarget, w.target)
	}
	if w.replacement.pickupTarget != nil {
		t.Fatalf("replacement update target = %p, want nil", w.replacement.pickupTarget)
	}
}

func TestDropOwnedCrowns4ED050FaultOrder(t *testing.T) {
	_, _, _, wantEvents := newDropOwnedCrownsFullTrace4ED050()
	for faultAt := range wantEvents {
		faultAt++
		t.Run(wantEvents[faultAt-1], func(t *testing.T) {
			w, _, _, _ := newDropOwnedCrownsFullTrace4ED050()
			w.faultAt = faultAt
			defer func() {
				gotPanic := recover()
				if gotPanic != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", gotPanic, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %v, want %v", w.events, want)
				}
			}()
			dropOwnedCrowns4ED050(w.hooks())
		})
	}
}

func TestDropOwnedCrowns4ED050EmptyListSkipsTarget(t *testing.T) {
	w := &dropOwnedCrownsTestWorld4ED050{
		lookup: 7,
		owner:  &dropOwnedCrownsTestObject4ED050{name: "owner"},
		target: &dropOwnedCrownsTestObject4ED050{name: "unread"},
	}
	dropOwnedCrowns4ED050(w.hooks())
	want := []string{"cache", "lookup:Crown", "store-cache", "owner-arg", "first:owner"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestDropOwnedCrowns4ED050StoresZeroLookupAndMatchesTypeZero(t *testing.T) {
	data := &dropOwnedCrownsTestData4ED050{name: "zero-data"}
	item := &dropOwnedCrownsTestObject4ED050{name: "zero", update: data}
	owner := &dropOwnedCrownsTestObject4ED050{name: "owner", first: item}
	target := &dropOwnedCrownsTestObject4ED050{name: "target"}
	w := &dropOwnedCrownsTestWorld4ED050{owner: owner, target: target}
	dropOwnedCrowns4ED050(w.hooks())
	if data.pickupTarget != target {
		t.Fatalf("zero-type target = %p, want %p", data.pickupTarget, target)
	}
	wantPrefix := []string{"cache", "lookup:Crown", "store-cache", "owner-arg", "first:owner", "target-arg"}
	if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix = %v, want %v", w.events[:len(wantPrefix)], wantPrefix)
	}
}

func TestDropOwnedCrowns4ED050UsesFullCacheAgainstZeroExtendedType(t *testing.T) {
	data := &dropOwnedCrownsTestData4ED050{name: "data"}
	item := &dropOwnedCrownsTestObject4ED050{name: "item", typeIndex: 1, update: data}
	owner := &dropOwnedCrownsTestObject4ED050{name: "owner", first: item}
	w := &dropOwnedCrownsTestWorld4ED050{
		cache:  0x00010001,
		lookup: 1,
		owner:  owner,
		target: &dropOwnedCrownsTestObject4ED050{name: "target"},
	}
	dropOwnedCrowns4ED050(w.hooks())
	if data.pickupTarget != nil {
		t.Fatalf("high-word cache matched 16-bit type: target = %p", data.pickupTarget)
	}
	want := []string{"cache", "owner-arg", "first:owner", "target-arg", "cache", "type:item", "next:item"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
