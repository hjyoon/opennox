package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type inventoryServPlaceTestObject4F36F0 struct {
	name          string
	carryCapacity uint16
	flags         uint32
	typeInd       uint16
	classLow      uint8
	pickup        string
	collide       string
	scriptFunc    int32
}

type inventoryServPlaceTestWorld4F36F0 struct {
	allowedResult int32
	arg3          int32
	arg4          int32
	pickupResult  int32
	defaultResult int32
	events        []string
	faultAt       int

	afterAllowed    func(*inventoryServPlaceTestWorld4F36F0)
	afterLoadPickup func(*inventoryServPlaceTestWorld4F36F0)
	afterPickup     func(*inventoryServPlaceTestWorld4F36F0)
	afterStoreFlags func(*inventoryServPlaceTestWorld4F36F0)
	afterRefresh    func(*inventoryServPlaceTestWorld4F36F0)
	afterScript     func(*inventoryServPlaceTestWorld4F36F0)
}

func inventoryServPlaceTestName4F36F0(obj *inventoryServPlaceTestObject4F36F0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *inventoryServPlaceTestWorld4F36F0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *inventoryServPlaceTestWorld4F36F0) hooks() inventoryServPlaceHooks4F36F0[
	*inventoryServPlaceTestObject4F36F0,
	string,
	string,
] {
	return inventoryServPlaceHooks4F36F0[
		*inventoryServPlaceTestObject4F36F0,
		string,
		string,
	]{
		loadOwnerCarryCapacity: func(owner *inventoryServPlaceTestObject4F36F0) uint16 {
			value := owner.carryCapacity
			w.event(fmt.Sprintf("capacity:%s=%04x", inventoryServPlaceTestName4F36F0(owner), value))
			return value
		},
		loadItemFlagsLow: func(item *inventoryServPlaceTestObject4F36F0) uint8 {
			value := uint8(item.flags)
			w.event(fmt.Sprintf("item-flags-low:%s=%02x", inventoryServPlaceTestName4F36F0(item), value))
			return value
		},
		loadOwnerFlags: func(owner *inventoryServPlaceTestObject4F36F0) uint32 {
			value := owner.flags
			w.event(fmt.Sprintf("owner-flags:%s=%08x", inventoryServPlaceTestName4F36F0(owner), value))
			return value
		},
		loadItemType: func(item *inventoryServPlaceTestObject4F36F0) uint16 {
			value := item.typeInd
			w.event(fmt.Sprintf("item-type:%s=%04x", inventoryServPlaceTestName4F36F0(item), value))
			return value
		},
		itemTypeAllowed: func(typeInd uint16) int32 {
			w.event(fmt.Sprintf("allowed:%04x=%08x", typeInd, uint32(w.allowedResult)))
			if w.afterAllowed != nil {
				w.afterAllowed(w)
			}
			return w.allowedResult
		},
		loadOwnerClassLow: func(owner *inventoryServPlaceTestObject4F36F0) uint8 {
			value := owner.classLow
			w.event(fmt.Sprintf("owner-class:%s=%02x", inventoryServPlaceTestName4F36F0(owner), value))
			return value
		},
		loadPickup: func(item *inventoryServPlaceTestObject4F36F0) string {
			value := item.pickup
			w.event(fmt.Sprintf("pickup:%s=%s", inventoryServPlaceTestName4F36F0(item), value))
			if w.afterLoadPickup != nil {
				w.afterLoadPickup(w)
			}
			return value
		},
		loadArg4: func() int32 {
			w.event(fmt.Sprintf("arg4=%08x", uint32(w.arg4)))
			return w.arg4
		},
		loadArg3: func() int32 {
			w.event(fmt.Sprintf("arg3=%08x", uint32(w.arg3)))
			return w.arg3
		},
		callPickup: func(pickup string, owner, item *inventoryServPlaceTestObject4F36F0, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf(
				"call-pickup:%s:%s:%s:%08x:%08x=%08x",
				pickup,
				inventoryServPlaceTestName4F36F0(owner),
				inventoryServPlaceTestName4F36F0(item),
				uint32(arg3),
				uint32(arg4),
				uint32(w.pickupResult),
			))
			if w.afterPickup != nil {
				w.afterPickup(w)
			}
			return w.pickupResult
		},
		defaultPickup: func(owner, item *inventoryServPlaceTestObject4F36F0, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf(
				"default:%s:%s:%08x:%08x=%08x",
				inventoryServPlaceTestName4F36F0(owner),
				inventoryServPlaceTestName4F36F0(item),
				uint32(arg3),
				uint32(arg4),
				uint32(w.defaultResult),
			))
			if w.afterPickup != nil {
				w.afterPickup(w)
			}
			return w.defaultResult
		},
		loadItemFlags: func(item *inventoryServPlaceTestObject4F36F0) uint32 {
			value := item.flags
			w.event(fmt.Sprintf("item-flags:%s=%08x", inventoryServPlaceTestName4F36F0(item), value))
			return value
		},
		storeItemFlags: func(item *inventoryServPlaceTestObject4F36F0, value uint32) {
			w.event(fmt.Sprintf("store-item-flags:%s=%08x", inventoryServPlaceTestName4F36F0(item), value))
			item.flags = value
			if w.afterStoreFlags != nil {
				w.afterStoreFlags(w)
			}
		},
		loadCollide: func(item *inventoryServPlaceTestObject4F36F0) string {
			value := item.collide
			w.event(fmt.Sprintf("collide:%s=%s", inventoryServPlaceTestName4F36F0(item), value))
			return value
		},
		refreshCollide: func(item *inventoryServPlaceTestObject4F36F0) {
			w.event(fmt.Sprintf("refresh-collide:%s", inventoryServPlaceTestName4F36F0(item)))
			if w.afterRefresh != nil {
				w.afterRefresh(w)
			}
		},
		loadScriptPickupFunc: func(item *inventoryServPlaceTestObject4F36F0) int32 {
			value := item.scriptFunc
			w.event(fmt.Sprintf("script-func:%s=%08x", inventoryServPlaceTestName4F36F0(item), uint32(value)))
			return value
		},
		callScriptPickup: func(owner, item *inventoryServPlaceTestObject4F36F0) {
			w.event(fmt.Sprintf("script:%s:%s", inventoryServPlaceTestName4F36F0(owner), inventoryServPlaceTestName4F36F0(item)))
			if w.afterScript != nil {
				w.afterScript(w)
			}
		},
		storeScriptPickupFunc: func(item *inventoryServPlaceTestObject4F36F0, value int32) {
			w.event(fmt.Sprintf("store-script-func:%s=%08x", inventoryServPlaceTestName4F36F0(item), uint32(value)))
			item.scriptFunc = value
		},
	}
}

func verifyInventoryServPlaceFaultPrefixes4F36F0(
	t *testing.T,
	want []string,
	build func() (*inventoryServPlaceTestWorld4F36F0, *inventoryServPlaceTestObject4F36F0, *inventoryServPlaceTestObject4F36F0),
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			inventoryServPlace4F36F0(owner, item, w.hooks())
		})
	}
}

func TestInventoryServPlace4F36F0CustomPickupExactResultAndLivePostState(t *testing.T) {
	build := func() (*inventoryServPlaceTestWorld4F36F0, *inventoryServPlaceTestObject4F36F0, *inventoryServPlaceTestObject4F36F0) {
		owner := &inventoryServPlaceTestObject4F36F0{
			name:          "owner",
			carryCapacity: 0x4321,
			flags:         0x10203040,
			classLow:      0x86,
		}
		item := &inventoryServPlaceTestObject4F36F0{
			name:       "item",
			flags:      0x00000001,
			typeInd:    0xabcd,
			pickup:     "custom",
			collide:    "old-collide",
			scriptFunc: -1,
		}
		w := &inventoryServPlaceTestWorld4F36F0{
			allowedResult: math.MinInt32,
			arg3:          0x12345678,
			arg4:          math.MinInt32,
			pickupResult:  math.MinInt32,
		}
		w.afterAllowed = func(*inventoryServPlaceTestWorld4F36F0) {
			owner.classLow = 0x06
		}
		w.afterLoadPickup = func(w *inventoryServPlaceTestWorld4F36F0) {
			item.pickup = "mutated"
			w.arg4 = -2
			w.arg3 = math.MaxInt32
		}
		w.afterPickup = func(*inventoryServPlaceTestWorld4F36F0) {
			item.flags = 0xa5a500c3
			item.collide = "live-collide"
		}
		w.afterStoreFlags = func(*inventoryServPlaceTestWorld4F36F0) {
			item.collide = "after-store"
		}
		w.afterRefresh = func(*inventoryServPlaceTestWorld4F36F0) {
			item.scriptFunc = 0x10203040
		}
		w.afterScript = func(*inventoryServPlaceTestWorld4F36F0) {
			item.scriptFunc = math.MinInt32
		}
		return w, owner, item
	}

	w, owner, item := build()
	if got := inventoryServPlace4F36F0(owner, item, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := []string{
		"capacity:owner=4321",
		"item-flags-low:item=01",
		"owner-flags:owner=10203040",
		"item-type:item=abcd",
		"allowed:abcd=80000000",
		"owner-class:owner=06",
		"pickup:item=custom",
		"arg4=fffffffe",
		"arg3=7fffffff",
		"call-pickup:custom:owner:item:7fffffff:fffffffe=80000000",
		"item-flags:item=a5a500c3",
		"store-item-flags:item=a5a50083",
		"collide:item=after-store",
		"refresh-collide:item",
		"script-func:item=10203040",
		"script:owner:item",
		"store-script-func:item=ffffffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if item.flags != 0xa5a50083 || item.scriptFunc != -1 {
		t.Fatalf("post state flags/script = %#08x/%d", item.flags, item.scriptFunc)
	}
	verifyInventoryServPlaceFaultPrefixes4F36F0(t, want, build)
}

func TestInventoryServPlace4F36F0NilPickupUsesDefaultAndExactResult(t *testing.T) {
	owner := &inventoryServPlaceTestObject4F36F0{name: "owner", carryCapacity: 1, classLow: 2}
	item := &inventoryServPlaceTestObject4F36F0{name: "item", typeInd: 0xffff, scriptFunc: -1}
	w := &inventoryServPlaceTestWorld4F36F0{
		allowedResult: 1,
		arg3:          math.MinInt32,
		arg4:          math.MaxInt32,
		defaultResult: -2,
	}
	if got := inventoryServPlace4F36F0(owner, item, w.hooks()); got != -2 {
		t.Fatalf("result = %d, want -2", got)
	}
	want := []string{
		"capacity:owner=0001",
		"item-flags-low:item=00",
		"owner-flags:owner=00000000",
		"item-type:item=ffff",
		"allowed:ffff=00000001",
		"owner-class:owner=02",
		"pickup:item=",
		"arg4=7fffffff",
		"arg3=80000000",
		"default:owner:item:80000000:7fffffff=fffffffe",
		"item-flags:item=00000000",
		"script-func:item=ffffffff",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestInventoryServPlace4F36F0ZeroPickupResultSkipsPostState(t *testing.T) {
	owner := &inventoryServPlaceTestObject4F36F0{name: "owner", carryCapacity: 1, classLow: 4}
	item := &inventoryServPlaceTestObject4F36F0{name: "item", typeInd: 7, pickup: "custom", flags: 0x40, collide: "collide", scriptFunc: 5}
	w := &inventoryServPlaceTestWorld4F36F0{allowedResult: 1, arg3: 3, arg4: 4}
	if got := inventoryServPlace4F36F0(owner, item, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"capacity:owner=0001",
		"item-flags-low:item=40",
		"owner-flags:owner=00000000",
		"item-type:item=0007",
		"allowed:0007=00000001",
		"owner-class:owner=04",
		"pickup:item=custom",
		"arg4=00000004",
		"arg3=00000003",
		"call-pickup:custom:owner:item:00000003:00000004=00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestInventoryServPlace4F36F0NoCollideAndScriptBranches(t *testing.T) {
	tests := []struct {
		name       string
		flags      uint32
		collide    string
		scriptFunc int32
		wantTail   []string
	}{
		{
			name:       "flag-clear-skips-collide-load",
			flags:      0x80000001,
			collide:    "ignored",
			scriptFunc: -1,
			wantTail: []string{
				"item-flags:item=80000001",
				"script-func:item=ffffffff",
			},
		},
		{
			name:       "set-flag-nil-collide",
			flags:      0x80000041,
			scriptFunc: -1,
			wantTail: []string{
				"item-flags:item=80000041",
				"store-item-flags:item=80000001",
				"collide:item=",
				"script-func:item=ffffffff",
			},
		},
		{
			name:       "script-zero-is-active",
			flags:      0x00000001,
			scriptFunc: 0,
			wantTail: []string{
				"item-flags:item=00000001",
				"script-func:item=00000000",
				"script:owner:item",
				"store-script-func:item=ffffffff",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := &inventoryServPlaceTestObject4F36F0{name: "owner", carryCapacity: 1, classLow: 4}
			item := &inventoryServPlaceTestObject4F36F0{
				name:       "item",
				typeInd:    1,
				pickup:     "custom",
				flags:      test.flags,
				collide:    test.collide,
				scriptFunc: test.scriptFunc,
			}
			w := &inventoryServPlaceTestWorld4F36F0{allowedResult: 1, pickupResult: 1}
			if got := inventoryServPlace4F36F0(owner, item, w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			prefixLen := len(w.events) - len(test.wantTail)
			if prefixLen < 0 || !reflect.DeepEqual(w.events[prefixLen:], test.wantTail) {
				t.Fatalf("tail = %v, want %v (all %v)", w.events[max(prefixLen, 0):], test.wantTail, w.events)
			}
		})
	}
}

func TestInventoryServPlace4F36F0GuardOrder(t *testing.T) {
	base := func() (*inventoryServPlaceTestWorld4F36F0, *inventoryServPlaceTestObject4F36F0, *inventoryServPlaceTestObject4F36F0) {
		return &inventoryServPlaceTestWorld4F36F0{allowedResult: 1},
			&inventoryServPlaceTestObject4F36F0{name: "owner", carryCapacity: 1, classLow: 4},
			&inventoryServPlaceTestObject4F36F0{name: "item", typeInd: 1, scriptFunc: -1}
	}
	tests := []struct {
		name   string
		mutate func(**inventoryServPlaceTestObject4F36F0, **inventoryServPlaceTestObject4F36F0, *inventoryServPlaceTestWorld4F36F0)
		want   []string
	}{
		{name: "nil-owner", mutate: func(owner, _ **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) {
			*owner = nil
		}},
		{name: "nil-item", mutate: func(_, item **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) { *item = nil }},
		{name: "zero-capacity", mutate: func(owner, _ **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) {
			(*owner).carryCapacity = 0
		}, want: []string{"capacity:owner=0000"}},
		{name: "destroyed-item", mutate: func(_, item **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) {
			(*item).flags = 0xa5a50020
		}, want: []string{"capacity:owner=0001", "item-flags-low:item=20"}},
		{name: "dead-owner", mutate: func(owner, _ **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) {
			(*owner).flags = 0x00008000
		}, want: []string{"capacity:owner=0001", "item-flags-low:item=00", "owner-flags:owner=00008000"}},
		{name: "disallowed-item", mutate: func(_, _ **inventoryServPlaceTestObject4F36F0, w *inventoryServPlaceTestWorld4F36F0) {
			w.allowedResult = 0
		}, want: []string{"capacity:owner=0001", "item-flags-low:item=00", "owner-flags:owner=00000000", "item-type:item=0001", "allowed:0001=00000000"}},
		{name: "non-unit-owner", mutate: func(owner, _ **inventoryServPlaceTestObject4F36F0, _ *inventoryServPlaceTestWorld4F36F0) {
			(*owner).classLow = 0xf8
		}, want: []string{"capacity:owner=0001", "item-flags-low:item=00", "owner-flags:owner=00000000", "item-type:item=0001", "allowed:0001=00000001", "owner-class:owner=f8"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w, owner, item := base()
			test.mutate(&owner, &item, w)
			if got := inventoryServPlace4F36F0(owner, item, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %v, want %v", w.events, test.want)
			}
		})
	}
}
