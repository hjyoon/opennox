package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	unitOrderHighOwner533900 = uint64(0x7fae173753d0)
	unitOrderHighCreatureA   = uint64(0x7fae175b47d0)
	unitOrderHighCreatureB   = uint64(0x7fae175c47d0)
	unitOrderHighCreatureC   = uint64(0x7fae175d47d0)
	unitOrderHighCreatureD   = uint64(0x7fae175e47d0)
	unitOrderHighUpdateA     = uint64(0x7fae200002ec)
	unitOrderHighUpdateB     = uint64(0x7fae300002ec)
	unitOrderHighDefinition  = uint64(0x7fae40000010)
	unitOrderHighAction      = uint64(0x7fae50000020)
)

func TestUnitOrder533900DispatchPreservesLiveOwnedListAndNativeWidth(t *testing.T) {
	next := map[uint64]uint64{
		unitOrderHighCreatureA: unitOrderHighCreatureB,
		unitOrderHighCreatureB: unitOrderHighCreatureC,
		unitOrderHighCreatureC: unitOrderHighCreatureD,
	}
	classes := map[uint64]uint8{
		unitOrderHighOwner533900: unitOrderPlayerClass533900,
		unitOrderHighCreatureA:   unitOrderMonsterClass533900,
		unitOrderHighCreatureB:   unitOrderMonsterClass533900,
		unitOrderHighCreatureC:   0x20,
		unitOrderHighCreatureD:   unitOrderMonsterClass533900,
	}
	updates := map[uint64]uint64{
		unitOrderHighCreatureA: unitOrderHighUpdateA,
		unitOrderHighCreatureB: unitOrderHighUpdateB,
		unitOrderHighCreatureD: unitOrderHighUpdateB,
	}
	statuses := map[uint64]uint32{
		unitOrderHighUpdateA: unitOrderSummonedFlag533900,
		unitOrderHighUpdateB: 0,
	}
	var events []string
	hooks := unitOrderDispatchHooks533900[uint64, uint64]{
		loadClassLow: func(obj uint64) uint8 {
			events = append(events, fmt.Sprintf("class:%x", obj))
			return classes[obj]
		},
		firstOwned: func(obj uint64) uint64 {
			events = append(events, fmt.Sprintf("first:%x", obj))
			return unitOrderHighCreatureA
		},
		nextOwned: func(obj uint64) uint64 {
			events = append(events, fmt.Sprintf("next:%x", obj))
			return next[obj]
		},
		loadUpdate: func(obj uint64) uint64 {
			events = append(events, fmt.Sprintf("update:%x", obj))
			return updates[obj]
		},
		loadStatus: func(update uint64) uint32 {
			events = append(events, fmt.Sprintf("status:%x", update))
			return statuses[update]
		},
		localOrder: func(owner uint64, order uint32) {
			events = append(events, fmt.Sprintf("local:%x:%d", owner, order))
		},
		enactOrder: func(owner, unit uint64, order uint32) {
			events = append(events, fmt.Sprintf("enact:%x:%x:%d", owner, unit, order))
			// GAME.EXE loads Field128 after the callback. A callback may unlink
			// the old successor, so the loop must observe this live replacement.
			next[unit] = unitOrderHighCreatureC
		},
	}

	unitOrder533900(unitOrderHighOwner533900, 0, 3, hooks)
	want := []string{
		"class:7fae173753d0",
		"local:7fae173753d0:3",
		"first:7fae173753d0",
		"class:7fae175b47d0",
		"update:7fae175b47d0",
		"status:7fae200002ec",
		"enact:7fae173753d0:7fae175b47d0:3",
		"next:7fae175b47d0",
		"class:7fae175d47d0",
		"next:7fae175d47d0",
		"class:7fae175e47d0",
		"update:7fae175e47d0",
		"status:7fae300002ec",
		"next:7fae175e47d0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestUnitOrder533900NullAndExplicitCreatureBranches(t *testing.T) {
	var events []string
	hooks := unitOrderDispatchHooks533900[uint64, uint64]{
		enactOrder: func(owner, creature uint64, order uint32) {
			events = append(events, fmt.Sprintf("%x:%x:%08x", owner, creature, order))
		},
	}
	unitOrder533900(uint64(0), unitOrderHighCreatureA, 5, hooks)
	if len(events) != 0 {
		t.Fatalf("null-owner events = %v, want none", events)
	}
	unitOrder533900(unitOrderHighOwner533900, unitOrderHighCreatureA, 0x80000000, hooks)
	want := []string{"7fae173753d0:7fae175b47d0:80000000"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("explicit events = %v, want %v", events, want)
	}
}

type unitOrderEnactWorld5339A0 struct {
	events       []string
	updateLoads  int
	secondUpdate bool
	class        map[uint64]uint8
	zombie       bool
	flags        uint32
	defs         map[uint64]uint64
	lookupDef    uint64
	owner        uint64
	status       map[uint64]uint32
	moving       bool
	shooting     bool
	action       uint64
	aggression   map[uint64]float32
	sight        map[uint64]float32
	storedDef    map[uint64]uint64
	storedStatus map[uint64]uint32
}

func newUnitOrderEnactWorld5339A0() *unitOrderEnactWorld5339A0 {
	return &unitOrderEnactWorld5339A0{
		class: map[uint64]uint8{
			unitOrderHighOwner533900: unitOrderPlayerClass533900,
			unitOrderHighCreatureA:   unitOrderMonsterClass533900,
		},
		defs: map[uint64]uint64{
			unitOrderHighUpdateA: unitOrderHighDefinition,
			unitOrderHighUpdateB: unitOrderHighDefinition,
		},
		owner:        unitOrderHighOwner533900,
		status:       map[uint64]uint32{unitOrderHighUpdateA: 0xa5, unitOrderHighUpdateB: 0x66},
		moving:       true,
		shooting:     true,
		action:       unitOrderHighAction,
		aggression:   make(map[uint64]float32),
		sight:        make(map[uint64]float32),
		storedDef:    make(map[uint64]uint64),
		storedStatus: make(map[uint64]uint32),
	}
}

func (w *unitOrderEnactWorld5339A0) hooks() unitOrderEnactHooks5339A0[uint64, uint64, uint64, uint64] {
	return unitOrderEnactHooks5339A0[uint64, uint64, uint64, uint64]{
		loadUpdate: func(unit uint64) uint64 {
			w.updateLoads++
			update := unitOrderHighUpdateA
			if w.secondUpdate && w.updateLoads > 1 {
				update = unitOrderHighUpdateB
			}
			w.events = append(w.events, fmt.Sprintf("update:%x", update))
			return update
		},
		loadClassLow: func(obj uint64) uint8 {
			w.events = append(w.events, fmt.Sprintf("class:%x", obj))
			return w.class[obj]
		},
		isZombie: func(unit uint64) bool {
			w.events = append(w.events, fmt.Sprintf("zombie:%x", unit))
			return w.zombie
		},
		loadObjectFlags: func(unit uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("flags:%x", unit))
			return w.flags
		},
		loadMonsterDef: func(update uint64) uint64 {
			w.events = append(w.events, fmt.Sprintf("def:%x", update))
			return w.defs[update]
		},
		storeMonsterDef: func(update, def uint64) {
			w.events = append(w.events, fmt.Sprintf("store-def:%x:%x", update, def))
			w.defs[update] = def
			w.storedDef[update] = def
		},
		loadTypeIndex: func(unit uint64) uint16 {
			w.events = append(w.events, fmt.Sprintf("type:%x", unit))
			return 0x1234
		},
		monsterDefByType: func(typ uint16) uint64 {
			w.events = append(w.events, fmt.Sprintf("lookup:%04x", typ))
			return w.lookupDef
		},
		loadOwner: func(unit uint64) uint64 {
			w.events = append(w.events, fmt.Sprintf("owner:%x", unit))
			return w.owner
		},
		banish: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("banish:%x", unit))
		},
		observe: func(source, unit uint64) {
			w.events = append(w.events, fmt.Sprintf("observe:%x:%x", source, unit))
		},
		monsterCommand: func(unit, source uint64, command string, arg int16) {
			w.events = append(w.events, fmt.Sprintf("command:%s:%d", command, arg))
		},
		playOrderSound: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("sound:%x", unit))
		},
		loadStatus: func(update uint64) uint32 {
			w.events = append(w.events, fmt.Sprintf("status:%x", update))
			return w.status[update]
		},
		storeStatus: func(update uint64, status uint32) {
			w.events = append(w.events, fmt.Sprintf("store-status:%x:%08x", update, status))
			w.storedStatus[update] = status
		},
		storeAggression: func(update uint64, value float32) {
			w.events = append(w.events, fmt.Sprintf("aggression:%x:%08x", update, math.Float32bits(value)))
			w.aggression[update] = value
		},
		storeSightRange: func(update uint64, value float32) {
			w.events = append(w.events, fmt.Sprintf("sight:%x:%08x", update, math.Float32bits(value)))
			w.sight[update] = value
		},
		isMoving: func(unit uint64) bool {
			w.events = append(w.events, fmt.Sprintf("moving:%x", unit))
			return w.moving
		},
		canShoot: func(unit uint64) bool {
			w.events = append(w.events, fmt.Sprintf("shoot:%x", unit))
			return w.shooting
		},
		clearActionStack: func(unit uint64) {
			w.events = append(w.events, fmt.Sprintf("clear:%x", unit))
		},
		pushAction: func(unit uint64, action ai.ActionType) uint64 {
			w.events = append(w.events, fmt.Sprintf("push:%d", action))
			return w.action
		},
		setGuardArgs: func(action, unit uint64) {
			w.events = append(w.events, fmt.Sprintf("guard-args:%x:%x", action, unit))
		},
		setEscortArgs: func(action, source uint64) {
			w.events = append(w.events, fmt.Sprintf("escort-args:%x:%x", action, source))
		},
	}
}

func TestEnactUnitOrder5339A0GatesAndDefinitionCache(t *testing.T) {
	t.Run("initial update precedes null source", func(t *testing.T) {
		w := newUnitOrderEnactWorld5339A0()
		enactUnitOrder5339A0(uint64(0), unitOrderHighCreatureA, 0, w.hooks())
		if want := []string{"update:7fae200002ec"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("non-zombie blocked flag", func(t *testing.T) {
		w := newUnitOrderEnactWorld5339A0()
		w.flags = unitOrderBlockedFlag5339A0
		enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, 0, w.hooks())
		want := []string{
			"update:7fae200002ec", "class:7fae175b47d0",
			"zombie:7fae175b47d0", "flags:7fae175b47d0",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("zombie banish skips flags", func(t *testing.T) {
		w := newUnitOrderEnactWorld5339A0()
		w.zombie = true
		w.flags = unitOrderBlockedFlag5339A0
		enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, 0, w.hooks())
		want := []string{
			"update:7fae200002ec", "class:7fae175b47d0",
			"zombie:7fae175b47d0", "def:7fae200002ec",
			"owner:7fae175b47d0", "banish:7fae175b47d0",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("failed lookup is stored before return", func(t *testing.T) {
		w := newUnitOrderEnactWorld5339A0()
		w.defs[unitOrderHighUpdateA] = 0
		enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, 2, w.hooks())
		wantTail := []string{
			"def:7fae200002ec", "type:7fae175b47d0", "lookup:1234",
			"store-def:7fae200002ec:0",
		}
		if !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
			t.Fatalf("events = %v, want tail %v", w.events, wantTail)
		}
		if _, ok := w.storedDef[unitOrderHighUpdateA]; !ok {
			t.Fatal("nil definition assignment was not observed")
		}
	})
}

func TestEnactUnitOrder5339A0ActionSequences(t *testing.T) {
	tests := []struct {
		name         string
		order        uint32
		secondUpdate bool
		wantTail     []string
	}{
		{
			name:  "idle",
			order: 2,
			wantTail: []string{
				"class:7fae173753d0", "command:MonUtil.c:idle:0", "sound:7fae175b47d0",
				"status:7fae200002ec", "aggression:7fae200002ec:3f000000",
				"store-status:7fae200002ec:000000a5", "clear:7fae175b47d0", "push:0",
			},
		},
		{
			name:  "guard",
			order: 3,
			wantTail: []string{
				"moving:7fae175b47d0", "class:7fae173753d0", "command:MonUtil.c:guarding:0",
				"sound:7fae175b47d0", "aggression:7fae200002ec:3f000000",
				"shoot:7fae175b47d0", "status:7fae200002ec",
				"store-status:7fae200002ec:000000e5", "sight:7fae200002ec:437a0000",
				"clear:7fae175b47d0", "push:4", "guard-args:7fae50000020:7fae175b47d0",
			},
		},
		{
			name:         "escort recaches update",
			order:        4,
			secondUpdate: true,
			wantTail: []string{
				"update:7fae300002ec", "moving:7fae175b47d0", "class:7fae173753d0",
				"command:MonUtil.c:escorting:0", "sound:7fae175b47d0", "status:7fae300002ec",
				"aggression:7fae300002ec:3f547ae1", "store-status:7fae300002ec:00000026",
				"clear:7fae175b47d0", "push:3", "escort-args:7fae50000020:7fae173753d0",
			},
		},
		{
			name:  "hunt plays sound before message",
			order: 5,
			wantTail: []string{
				"moving:7fae175b47d0", "sound:7fae175b47d0", "class:7fae173753d0",
				"command:MonUtil.c:Hunting:0", "status:7fae200002ec",
				"aggression:7fae200002ec:3f547ae1", "store-status:7fae200002ec:000000a5",
				"clear:7fae175b47d0", "push:5",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitOrderEnactWorld5339A0()
			w.secondUpdate = tc.secondUpdate
			enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, tc.order, w.hooks())
			if !reflect.DeepEqual(w.events[len(w.events)-len(tc.wantTail):], tc.wantTail) {
				t.Fatalf("events = %q, want tail %q", w.events, tc.wantTail)
			}
		})
	}
}

func TestEnactUnitOrder5339A0MovementAndNullActionBranches(t *testing.T) {
	w := newUnitOrderEnactWorld5339A0()
	w.moving = false
	enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, 3, w.hooks())
	if got := w.events[len(w.events)-1]; got != "moving:7fae175b47d0" {
		t.Fatalf("last event = %q, want movement gate", got)
	}

	w = newUnitOrderEnactWorld5339A0()
	w.action = 0
	enactUnitOrder5339A0(unitOrderHighOwner533900, unitOrderHighCreatureA, 3, w.hooks())
	for _, event := range w.events {
		if len(event) >= len("guard-args:") && event[:len("guard-args:")] == "guard-args:" {
			t.Fatalf("null action unexpectedly stored arguments: %v", w.events)
		}
	}
}

func TestUnitOrder5339A0OracleConstants(t *testing.T) {
	checks := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"moving threshold", math.Float32bits(unitOrderMovingSpeed5339A0), 0x3c23d70a},
		{"calm aggression", math.Float32bits(unitOrderCalm5339A0), 0x3f000000},
		{"aggressive order", math.Float32bits(unitOrderAggressive5339A0), 0x3f547ae1},
		{"guard sight", math.Float32bits(unitOrderGuardSight5339A0), 0x437a0000},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %#08x, want %#08x", check.name, check.got, check.want)
		}
	}
	if unitOrderGuardSight5339A0 != 250 {
		t.Fatalf("guard sight = %g, want 250", unitOrderGuardSight5339A0)
	}
}
