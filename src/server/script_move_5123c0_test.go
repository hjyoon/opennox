package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	scriptMoveHighUnit5123C0           = uint64(0x7fae175b47d0)
	scriptMoveHighWaypoint5123C0       = uint64(0x7fae20920d80)
	scriptMoveHighMonsterData5123C0    = uint64(0x7fae300002ec)
	scriptMoveChangedMonsterData5123C0 = uint64(0x7fae400002ec)
	scriptMoveHighReport5123C0         = uint64(0x7fae50000020)
	scriptMoveHighRoam5123C0           = uint64(0x7fae60000010)
	scriptMoveHighFar5123C0            = uint64(0x7fae70000008)
)

type scriptMoveTestWorld5123C0 struct {
	events          []string
	flags           map[uint64]uint32
	classes         map[uint64]uint8
	types           map[uint64]uint16
	monsterUpdate   uint64
	changedUpdate   uint64
	mutateOnClear   bool
	points          map[uint64]uint8
	xbits           map[uint64]uint32
	ybits           map[uint64]uint32
	actions         map[uint32]uint64
	roamFlags       map[uint64]uint8
	storedU32       map[uint64][4]uint32
	storedLow       map[uint64][4]uint8
	storedWaypoints map[uint64][4]uint64
	moverType       uint32
	first           uint64
	next            map[uint64]uint64
	moverUpdates    map[uint64]uint64
	extents         map[uint64]uint32
	targetExtents   map[uint64]uint32
}

func newScriptMoveTestWorld5123C0() *scriptMoveTestWorld5123C0 {
	return &scriptMoveTestWorld5123C0{
		flags:         make(map[uint64]uint32),
		classes:       make(map[uint64]uint8),
		types:         make(map[uint64]uint16),
		monsterUpdate: scriptMoveHighMonsterData5123C0,
		changedUpdate: scriptMoveChangedMonsterData5123C0,
		points: map[uint64]uint8{
			scriptMoveHighWaypoint5123C0: 1,
		},
		xbits: map[uint64]uint32{
			scriptMoveHighWaypoint5123C0: 0x3fc00000,
		},
		ybits: map[uint64]uint32{
			scriptMoveHighWaypoint5123C0: 0xc0100000,
		},
		actions: map[uint32]uint64{
			scriptMoveReportAction5123C0: scriptMoveHighReport5123C0,
			scriptMoveRoamAction5123C0:   scriptMoveHighRoam5123C0,
			scriptMoveFarAction5123C0:    scriptMoveHighFar5123C0,
		},
		roamFlags: map[uint64]uint8{
			scriptMoveHighMonsterData5123C0:    0xa5,
			scriptMoveChangedMonsterData5123C0: 0x5a,
		},
		storedU32:       make(map[uint64][4]uint32),
		storedLow:       make(map[uint64][4]uint8),
		storedWaypoints: make(map[uint64][4]uint64),
		next:            make(map[uint64]uint64),
		moverUpdates:    make(map[uint64]uint64),
		extents:         make(map[uint64]uint32),
		targetExtents:   make(map[uint64]uint32),
	}
}

func (w *scriptMoveTestWorld5123C0) observe(event string) {
	w.events = append(w.events, event)
}

func (w *scriptMoveTestWorld5123C0) hooks() scriptMoveHooks5123C0[uint64, uint64, uint64, uint64, uint64] {
	return scriptMoveHooks5123C0[uint64, uint64, uint64, uint64, uint64]{
		loadFlags: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("flags:%x", unit))
			return w.flags[unit]
		},
		loadClassLow: func(unit uint64) uint8 {
			w.observe(fmt.Sprintf("class:%x", unit))
			return w.classes[unit]
		},
		loadMonsterUpdate: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("monster-update:%x", unit))
			return w.monsterUpdate
		},
		clearActionStack: func(unit uint64) {
			w.observe(fmt.Sprintf("clear:%x", unit))
			if w.mutateOnClear {
				w.monsterUpdate = w.changedUpdate
			}
		},
		pushAction: func(unit uint64, action uint32) uint64 {
			w.observe(fmt.Sprintf("push:%x:%d", unit, action))
			return w.actions[action]
		},
		storeActionArgU32: func(action uint64, index int, value uint32) {
			w.observe(fmt.Sprintf("store-u32:%x:%d:%08x", action, index, value))
			args := w.storedU32[action]
			args[index] = value
			w.storedU32[action] = args
		},
		loadWaypointPoints: func(waypoint uint64) uint8 {
			w.observe(fmt.Sprintf("points:%x", waypoint))
			return w.points[waypoint]
		},
		storeActionWaypoint: func(action uint64, index int, waypoint uint64) {
			w.observe(fmt.Sprintf("store-waypoint:%x:%d:%x", action, index, waypoint))
			args := w.storedWaypoints[action]
			args[index] = waypoint
			w.storedWaypoints[action] = args
		},
		loadRoamFlagsLow: func(update uint64) uint8 {
			w.observe(fmt.Sprintf("roam-flags:%x", update))
			return w.roamFlags[update]
		},
		storeActionArgLow: func(action uint64, index int, value uint8) {
			w.observe(fmt.Sprintf("store-low:%x:%d:%02x", action, index, value))
			args := w.storedLow[action]
			args[index] = value
			w.storedLow[action] = args
		},
		loadWaypointXBits: func(waypoint uint64) uint32 {
			w.observe(fmt.Sprintf("x:%x", waypoint))
			return w.xbits[waypoint]
		},
		loadWaypointYBits: func(waypoint uint64) uint32 {
			w.observe(fmt.Sprintf("y:%x", waypoint))
			return w.ybits[waypoint]
		},
		loadMoverType: func() uint32 {
			w.observe("mover-type")
			return w.moverType
		},
		loadTypeInd: func(unit uint64) uint16 {
			w.observe(fmt.Sprintf("type:%x", unit))
			return w.types[unit]
		},
		moverGoTo: func(unit, waypoint uint64) {
			w.observe(fmt.Sprintf("mover-go:%x:%x", unit, waypoint))
		},
		firstObject: func() uint64 {
			w.observe("first")
			return w.first
		},
		loadMoverUpdate: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("mover-update:%x", unit))
			return w.moverUpdates[unit]
		},
		loadExtent: func(unit uint64) uint32 {
			w.observe(fmt.Sprintf("extent:%x", unit))
			return w.extents[unit]
		},
		loadMoverTargetExtent: func(update uint64) uint32 {
			w.observe(fmt.Sprintf("target-extent:%x", update))
			return w.targetExtents[update]
		},
		nextObject: func(unit uint64) uint64 {
			w.observe(fmt.Sprintf("next:%x", unit))
			return w.next[unit]
		},
	}
}

func TestScriptMove5123C0MonsterExactOrderAndNativeWidth(t *testing.T) {
	w := newScriptMoveTestWorld5123C0()
	w.classes[scriptMoveHighUnit5123C0] = scriptMoveMonsterClassLow5123C0
	w.mutateOnClear = true
	scriptMoveTo5123C0(scriptMoveHighUnit5123C0, scriptMoveHighWaypoint5123C0, w.hooks())

	want := []string{
		"flags:7fae175b47d0",
		"class:7fae175b47d0",
		"monster-update:7fae175b47d0",
		"clear:7fae175b47d0",
		"push:7fae175b47d0:32",
		"store-u32:7fae50000020:0:00000008",
		"points:7fae20920d80",
		"push:7fae175b47d0:10",
		"store-waypoint:7fae60000010:0:7fae20920d80",
		"roam-flags:7fae300002ec",
		"store-low:7fae60000010:2:a5",
		"push:7fae175b47d0:8",
		"x:7fae20920d80",
		"store-u32:7fae70000008:0:3fc00000",
		"y:7fae20920d80",
		"store-u32:7fae70000008:1:c0100000",
		"store-u32:7fae70000008:2:00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if got := w.storedWaypoints[scriptMoveHighRoam5123C0][0]; got != scriptMoveHighWaypoint5123C0 {
		t.Fatalf("ROAM waypoint = %#x, want native handle %#x", got, scriptMoveHighWaypoint5123C0)
	}
	if got := w.storedLow[scriptMoveHighRoam5123C0][2]; got != 0xa5 {
		t.Fatalf("ROAM flags = %#x, want cached UpdateData byte 0xa5", got)
	}
	if got := w.storedU32[scriptMoveHighFar5123C0]; got != [4]uint32{0x3fc00000, 0xc0100000, 0, 0} {
		t.Fatalf("FAR_MOVE_TO args = %#v", got)
	}
}

func TestScriptMove5123C0MonsterPushFailuresMatchOracle(t *testing.T) {
	tests := []struct {
		name       string
		action     uint32
		wantSuffix []string
	}{
		{
			name:   "report failure continues",
			action: scriptMoveReportAction5123C0,
			wantSuffix: []string{
				"push:7fae175b47d0:32", "points:7fae20920d80",
				"push:7fae175b47d0:10", "store-waypoint:7fae60000010:0:7fae20920d80",
				"roam-flags:7fae300002ec", "store-low:7fae60000010:2:a5",
				"push:7fae175b47d0:8", "x:7fae20920d80",
				"store-u32:7fae70000008:0:3fc00000", "y:7fae20920d80",
				"store-u32:7fae70000008:1:c0100000", "store-u32:7fae70000008:2:00000000",
			},
		},
		{
			name:   "roam failure continues",
			action: scriptMoveRoamAction5123C0,
			wantSuffix: []string{
				"push:7fae175b47d0:32", "store-u32:7fae50000020:0:00000008",
				"points:7fae20920d80", "push:7fae175b47d0:10",
				"push:7fae175b47d0:8", "x:7fae20920d80",
				"store-u32:7fae70000008:0:3fc00000", "y:7fae20920d80",
				"store-u32:7fae70000008:1:c0100000", "store-u32:7fae70000008:2:00000000",
			},
		},
		{
			name:   "far failure returns",
			action: scriptMoveFarAction5123C0,
			wantSuffix: []string{
				"push:7fae175b47d0:32", "store-u32:7fae50000020:0:00000008",
				"points:7fae20920d80", "push:7fae175b47d0:10",
				"store-waypoint:7fae60000010:0:7fae20920d80",
				"roam-flags:7fae300002ec", "store-low:7fae60000010:2:a5",
				"push:7fae175b47d0:8",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newScriptMoveTestWorld5123C0()
			w.classes[scriptMoveHighUnit5123C0] = scriptMoveMonsterClassLow5123C0
			w.actions[tc.action] = 0
			scriptMoveTo5123C0(scriptMoveHighUnit5123C0, scriptMoveHighWaypoint5123C0, w.hooks())
			prefix := []string{
				"flags:7fae175b47d0", "class:7fae175b47d0",
				"monster-update:7fae175b47d0", "clear:7fae175b47d0",
			}
			want := append(prefix, tc.wantSuffix...)
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		})
	}
}

func TestScriptMove5123C0BlockedGatePrecedesClassLoad(t *testing.T) {
	w := newScriptMoveTestWorld5123C0()
	w.flags[scriptMoveHighUnit5123C0] = scriptMoveBlockedFlag5123C0
	scriptMoveTo5123C0(scriptMoveHighUnit5123C0, scriptMoveHighWaypoint5123C0, w.hooks())
	if want := []string{"flags:7fae175b47d0"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestScriptMove5123C0DirectMoverDispatch(t *testing.T) {
	w := newScriptMoveTestWorld5123C0()
	w.moverType = 0x1234
	w.types[scriptMoveHighUnit5123C0] = 0x1234
	scriptMoveTo5123C0(scriptMoveHighUnit5123C0, scriptMoveHighWaypoint5123C0, w.hooks())
	want := []string{
		"flags:7fae175b47d0", "class:7fae175b47d0",
		"mover-type", "type:7fae175b47d0",
		"mover-go:7fae175b47d0:7fae20920d80",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestScriptMove5123C0ScansMatchingMoversInLiveOrder(t *testing.T) {
	const (
		plain    = uint64(0x7fae81000001)
		matching = uint64(0x7fae82000002)
		other    = uint64(0x7fae83000003)
		update1  = uint64(0x7fae91000001)
		update2  = uint64(0x7fae92000002)
	)
	w := newScriptMoveTestWorld5123C0()
	w.moverType = 0x1234
	w.types[scriptMoveHighUnit5123C0] = 1
	w.types[plain] = 2
	w.types[matching] = 0x1234
	w.types[other] = 0x1234
	w.first = plain
	w.next[plain] = matching
	w.next[matching] = other
	w.moverUpdates[matching] = update1
	w.moverUpdates[other] = update2
	w.extents[scriptMoveHighUnit5123C0] = 0xaabbccdd
	w.targetExtents[update1] = 0xaabbccdd
	w.targetExtents[update2] = 0x11223344

	scriptMoveTo5123C0(scriptMoveHighUnit5123C0, scriptMoveHighWaypoint5123C0, w.hooks())
	want := []string{
		"flags:7fae175b47d0", "class:7fae175b47d0",
		"mover-type", "type:7fae175b47d0", "first",
		"mover-type", "type:7fae81000001", "next:7fae81000001",
		"mover-type", "type:7fae82000002", "mover-update:7fae82000002",
		"extent:7fae175b47d0", "target-extent:7fae91000001",
		"mover-go:7fae82000002:7fae20920d80", "next:7fae82000002",
		"mover-type", "type:7fae83000003", "mover-update:7fae83000003",
		"extent:7fae175b47d0", "target-extent:7fae92000002", "next:7fae83000003",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestMoverGoTo5124C0CachesUpdateAndLoadsWaypointLate(t *testing.T) {
	const (
		unit           = uint64(0x7faea1000001)
		waypoint       = uint64(0x7faea2000002)
		entryUpdate    = uint64(0x7faea3000003)
		replacedUpdate = uint64(0x7faea4000004)
	)
	var events []string
	liveUpdate := entryUpdate
	waypointIndex := uint32(0x10203040)
	storedState := make(map[uint64]uint8)
	storedIndex := make(map[uint64]uint32)
	moverGoTo5124C0(unit, waypoint, moverGoToHooks5124C0[uint64, uint64, uint64]{
		loadUpdate: func(got uint64) uint64 {
			events = append(events, fmt.Sprintf("update:%x", got))
			return liveUpdate
		},
		setOn: func(got uint64) {
			events = append(events, fmt.Sprintf("set-on:%x", got))
			liveUpdate = replacedUpdate
			waypointIndex = 0xa1b2c3d4
		},
		storeVelocityXBits: func(got uint64, bits uint32) {
			events = append(events, fmt.Sprintf("vel-x:%x:%08x", got, bits))
		},
		storeVelocityYBits: func(got uint64, bits uint32) {
			events = append(events, fmt.Sprintf("vel-y:%x:%08x", got, bits))
		},
		storeStateLow: func(update uint64, value uint8) {
			events = append(events, fmt.Sprintf("state:%x:%02x", update, value))
			storedState[update] = value
		},
		loadWaypointIndex: func(got uint64) uint32 {
			events = append(events, fmt.Sprintf("waypoint-index:%x", got))
			return waypointIndex
		},
		storeWaypointIndex: func(update uint64, value uint32) {
			events = append(events, fmt.Sprintf("store-index:%x:%08x", update, value))
			storedIndex[update] = value
		},
		addToUpdatable: func(got uint64) {
			events = append(events, fmt.Sprintf("updatable:%x", got))
		},
	})
	want := []string{
		"update:7faea1000001", "set-on:7faea1000001",
		"vel-x:7faea1000001:00000000", "vel-y:7faea1000001:00000000",
		"state:7faea3000003:00", "waypoint-index:7faea2000002",
		"store-index:7faea3000003:a1b2c3d4", "updatable:7faea1000001",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
	if _, ok := storedIndex[replacedUpdate]; ok {
		t.Fatal("stores followed the replacement UpdateData instead of the entry-time handle")
	}
	if storedState[entryUpdate] != 0 || storedIndex[entryUpdate] != 0xa1b2c3d4 {
		t.Fatalf("cached update stores = %#x/%#x", storedState[entryUpdate], storedIndex[entryUpdate])
	}
}

func TestScriptMove5123C0OracleConstants(t *testing.T) {
	checks := []struct {
		got  ai.ActionType
		want ai.ActionType
	}{
		{ai.ActionType(scriptMoveFarAction5123C0), ai.ACTION_FAR_MOVE_TO},
		{ai.ActionType(scriptMoveRoamAction5123C0), ai.ACTION_ROAM},
		{ai.ActionType(scriptMoveReportAction5123C0), ai.ACTION_REPORT},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("action = %s, want %s", check.got, check.want)
		}
	}
}
