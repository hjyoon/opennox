package server

import (
	"fmt"
	"reflect"
	"testing"
)

type chakramUpdateTestObject53DCC0 struct {
	name        string
	flags       uint32
	owner       *chakramUpdateTestObject53DCC0
	item        *chakramUpdateTestObject53DCC0
	x           float32
	y           float32
	velX        float32
	velY        float32
	speed       float32
	createFrame uint32
}

type chakramUpdateTestData53DCC0 struct {
	last     *chakramUpdateTestObject53DCC0
	ownerX   float32
	ownerY   float32
	state    uint8
	returnTo *chakramUpdateTestObject53DCC0
}

type chakramUpdateTestFixture53DCC0 struct {
	events     []string
	update     *chakramUpdateTestData53DCC0
	ownerReads []*chakramUpdateTestObject53DCC0
	mapOK      bool
	frameValue uint32
	rateValue  uint32
}

func (f *chakramUpdateTestFixture53DCC0) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *chakramUpdateTestFixture53DCC0) hooks() chakramUpdateHooks53DCC0[
	*chakramUpdateTestObject53DCC0,
	*chakramUpdateTestData53DCC0,
] {
	return chakramUpdateHooks53DCC0[
		*chakramUpdateTestObject53DCC0,
		*chakramUpdateTestData53DCC0,
	]{
		loadUpdateData: func(obj *chakramUpdateTestObject53DCC0) *chakramUpdateTestData53DCC0 {
			f.event("update:%s", obj.name)
			return f.update
		},
		inventoryFirst: func(obj *chakramUpdateTestObject53DCC0) *chakramUpdateTestObject53DCC0 {
			f.event("inventory:%s", obj.name)
			return obj.item
		},
		loadFlags: func(obj *chakramUpdateTestObject53DCC0) uint32 {
			f.event("flags:%s", obj.name)
			return obj.flags
		},
		loadLastHit: func(update *chakramUpdateTestData53DCC0) *chakramUpdateTestObject53DCC0 {
			f.event("last")
			return update.last
		},
		storeLastHit: func(update *chakramUpdateTestData53DCC0, obj *chakramUpdateTestObject53DCC0) {
			f.event("store-last:nil")
			update.last = obj
		},
		loadOwner: func(obj *chakramUpdateTestObject53DCC0) *chakramUpdateTestObject53DCC0 {
			f.event("owner:%s", obj.name)
			if len(f.ownerReads) == 0 {
				return obj.owner
			}
			owner := f.ownerReads[0]
			f.ownerReads = f.ownerReads[1:]
			return owner
		},
		loadPosX: func(obj *chakramUpdateTestObject53DCC0) float32 {
			f.event("x:%s", obj.name)
			return obj.x
		},
		loadPosY: func(obj *chakramUpdateTestObject53DCC0) float32 {
			f.event("y:%s", obj.name)
			return obj.y
		},
		storeOwnerPosX: func(update *chakramUpdateTestData53DCC0, value float32) {
			f.event("store-owner-x")
			update.ownerX = value
		},
		storeOwnerPosY: func(update *chakramUpdateTestData53DCC0, value float32) {
			f.event("store-owner-y")
			update.ownerY = value
		},
		loadOwnerPosX: func(update *chakramUpdateTestData53DCC0) float32 {
			f.event("owner-x")
			return update.ownerX
		},
		loadOwnerPosY: func(update *chakramUpdateTestData53DCC0) float32 {
			f.event("owner-y")
			return update.ownerY
		},
		mapCheck: func(source, owner *chakramUpdateTestObject53DCC0) bool {
			ownerName := "nil"
			if owner != nil {
				ownerName = owner.name
			}
			f.event("map:%s:%s", source.name, ownerName)
			return f.mapOK
		},
		loadReturnState: func(update *chakramUpdateTestData53DCC0) uint8 {
			f.event("state:%d", update.state)
			return update.state
		},
		storeReturnState: func(update *chakramUpdateTestData53DCC0, value uint8) {
			f.event("store-state:%d", value)
			update.state = value
		},
		loadReturnTarget: func(update *chakramUpdateTestData53DCC0) *chakramUpdateTestObject53DCC0 {
			f.event("return-target")
			return update.returnTo
		},
		storeReturnTarget: func(update *chakramUpdateTestData53DCC0, obj *chakramUpdateTestObject53DCC0) {
			name := "nil"
			if obj != nil {
				name = obj.name
			}
			f.event("store-return:%s", name)
			update.returnTo = obj
		},
		loadSpeed: func(obj *chakramUpdateTestObject53DCC0) float32 {
			f.event("speed:%s", obj.name)
			return obj.speed
		},
		storeVelocityX: func(obj *chakramUpdateTestObject53DCC0, value float32) {
			f.event("vel-x")
			obj.velX = value
		},
		storeVelocityY: func(obj *chakramUpdateTestObject53DCC0, value float32) {
			f.event("vel-y")
			obj.velY = value
		},
		frame: func() uint32 {
			f.event("frame")
			return f.frameValue
		},
		loadCreateFrame: func(obj *chakramUpdateTestObject53DCC0) uint32 {
			f.event("created:%s", obj.name)
			return obj.createFrame
		},
		frameRate: func() uint32 {
			f.event("rate")
			return f.rateValue
		},
		delayedDelete: func(obj *chakramUpdateTestObject53DCC0) {
			f.event("delete:%s", obj.name)
		},
	}
}

func newChakramUpdateFixture53DCC0() (
	*chakramUpdateTestFixture53DCC0,
	*chakramUpdateTestObject53DCC0,
	*chakramUpdateTestObject53DCC0,
) {
	item := &chakramUpdateTestObject53DCC0{name: "item"}
	owner := &chakramUpdateTestObject53DCC0{name: "owner", x: 13, y: 24}
	source := &chakramUpdateTestObject53DCC0{
		name: "source", owner: owner, item: item, x: 10, y: 20, speed: 11, createFrame: 100,
	}
	fixture := &chakramUpdateTestFixture53DCC0{
		update: &chakramUpdateTestData53DCC0{}, mapOK: true, frameValue: 120, rateValue: 30,
	}
	return fixture, source, owner
}

func TestChakramUpdate53DCC0InventoryGateCachesUpdateFirst(t *testing.T) {
	tests := []struct {
		name  string
		flags uint32
		item  bool
		want  []string
	}{
		{"missing", 0, false, []string{"update:source", "inventory:source", "delete:source"}},
		{"destroyed", 0x20, true, []string{"update:source", "inventory:source", "flags:item", "delete:source"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, source, _ := newChakramUpdateFixture53DCC0()
			if !tc.item {
				source.item = nil
			} else {
				source.item.flags = tc.flags
			}
			chakramUpdate53DCC0(source, f.hooks())
			if !reflect.DeepEqual(f.events, tc.want) {
				t.Fatalf("events = %q, want %q", f.events, tc.want)
			}
		})
	}
}

func TestChakramUpdate53DCC0ClearsLastHitThenHandlesInvalidOwner(t *testing.T) {
	f, source, owner := newChakramUpdateFixture53DCC0()
	last := &chakramUpdateTestObject53DCC0{name: "last", flags: chakramDestroyedFlag4EAF00}
	f.update.last = last
	f.update.returnTo = owner
	owner.flags = chakramDestroyedFlag4EAF00
	chakramUpdate53DCC0(source, f.hooks())
	want := []string{
		"update:source", "inventory:source", "flags:item", "last", "flags:last", "store-last:nil",
		"owner:source", "flags:owner", "store-state:1", "store-return:nil",
		"frame", "created:source", "rate",
	}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events = %q, want %q", f.events, want)
	}
	if f.update.last != nil || f.update.returnTo != nil || f.update.state != chakramReturnStateDrop4EAF00 {
		t.Fatalf("update = %+v, want cleared last/target and drop state", f.update)
	}
}

func TestChakramUpdate53DCC0MapFailureStillSteersToCachedOwnerPosition(t *testing.T) {
	f, source, cachedOwner := newChakramUpdateFixture53DCC0()
	liveOwner := &chakramUpdateTestObject53DCC0{name: "live-owner"}
	f.ownerReads = []*chakramUpdateTestObject53DCC0{cachedOwner, liveOwner}
	f.mapOK = false
	chakramUpdate53DCC0(source, f.hooks())
	if f.update.ownerX != cachedOwner.x || f.update.ownerY != cachedOwner.y || f.update.returnTo != nil {
		t.Fatalf("owner snapshot/target = (%v, %v, %p)", f.update.ownerX, f.update.ownerY, f.update.returnTo)
	}
	if source.velX == 0 || source.velY == 0 {
		t.Fatalf("velocity = (%v, %v), want cached-owner steering", source.velX, source.velY)
	}
	wantSubsequence := []string{
		"owner:source", "flags:owner", "x:owner", "store-owner-x", "y:owner", "store-owner-y",
		"owner:source", "map:source:live-owner", "store-return:nil", "state:0", "return-target",
	}
	start := 4 // update, inventory, item flags, last
	if !reflect.DeepEqual(f.events[start:start+len(wantSubsequence)], wantSubsequence) {
		t.Fatalf("owner/map sequence = %q, want %q", f.events[start:start+len(wantSubsequence)], wantSubsequence)
	}
}

func TestChakramUpdate53DCC0NonzeroStateSkipsLiveTargetWriteAndSteering(t *testing.T) {
	f, source, cachedOwner := newChakramUpdateFixture53DCC0()
	liveOwner := &chakramUpdateTestObject53DCC0{name: "live-owner"}
	f.ownerReads = []*chakramUpdateTestObject53DCC0{cachedOwner, liveOwner}
	f.update.state = chakramReturnStateSeek4EAF00
	f.update.returnTo = cachedOwner
	chakramUpdate53DCC0(source, f.hooks())
	if f.update.returnTo != cachedOwner || source.velX != 0 || source.velY != 0 {
		t.Fatalf("result = (target %p, velocity %v/%v), want unchanged target and velocity", f.update.returnTo, source.velX, source.velY)
	}
	for _, event := range f.events {
		if event == "return-target" || event == "vel-x" || event == "vel-y" || event == "store-return:live-owner" {
			t.Fatalf("unexpected event %q in %q", event, f.events)
		}
	}
}

func TestChakramUpdate53DCC0InvalidLiveReturnTargetClearsBeforeState(t *testing.T) {
	f, source, cachedOwner := newChakramUpdateFixture53DCC0()
	mapOwner := &chakramUpdateTestObject53DCC0{name: "map-owner"}
	liveReturn := &chakramUpdateTestObject53DCC0{name: "live-return", flags: chakramUntargetableFlag4EAF00}
	f.ownerReads = []*chakramUpdateTestObject53DCC0{cachedOwner, mapOwner, liveReturn}
	chakramUpdate53DCC0(source, f.hooks())
	wantTailBeforeLifetime := []string{
		"store-return:live-return", "state:0", "return-target", "flags:live-return",
		"store-return:nil", "store-state:1",
	}
	frameIndex := len(f.events) - 3
	if !reflect.DeepEqual(f.events[frameIndex-len(wantTailBeforeLifetime):frameIndex], wantTailBeforeLifetime) {
		t.Fatalf("target invalidation order = %q", f.events)
	}
	if f.update.returnTo != nil || f.update.state != chakramReturnStateDrop4EAF00 {
		t.Fatalf("update = %+v, want cleared target/drop state", f.update)
	}
}

func TestChakramUpdate53DCC0LifetimeUsesWrappingUint32(t *testing.T) {
	tests := []struct {
		name       string
		frame      uint32
		wantExpire bool
	}{
		{"equal limit", 2, false},
		{"above limit", 3, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, source, _ := newChakramUpdateFixture53DCC0()
			source.createFrame = ^uint32(0) - 2
			f.frameValue = tc.frame
			f.rateValue = 1
			chakramUpdate53DCC0(source, f.hooks())
			if got := f.update.state == chakramReturnStateDrop4EAF00; got != tc.wantExpire {
				t.Fatalf("expired = %t, want %t; update %+v", got, tc.wantExpire, f.update)
			}
		})
	}
}
