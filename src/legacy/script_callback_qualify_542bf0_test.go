package legacy

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

type scriptCallbackQualifyTestObject542BF0 struct {
	next      *scriptCallbackQualifyTestObject542BF0
	flags     uint32
	id        *string
	kind      scriptCallbackObjectKind542BF0
	callbacks map[int]*string
}

type scriptCallbackQualifyTestWaypoint542BF0 struct {
	next  *scriptCallbackQualifyTestWaypoint542BF0
	flags uint32
	name  string
}

type scriptCallbackQualifyTestStore542BF0 struct {
	object *scriptCallbackQualifyTestObject542BF0
	event  int
	name   string
}

func scriptCallbackQualifyTestString542BF0(value string) *string {
	return &value
}

func scriptCallbackQualifyTestHooks542BF0(
	firstObject *scriptCallbackQualifyTestObject542BF0,
	firstWaypoint *scriptCallbackQualifyTestWaypoint542BF0,
	stores *[]scriptCallbackQualifyTestStore542BF0,
) scriptCallbackQualifyHooks542BF0[
	*scriptCallbackQualifyTestObject542BF0,
	*scriptCallbackQualifyTestWaypoint542BF0,
] {
	return scriptCallbackQualifyHooks542BF0[
		*scriptCallbackQualifyTestObject542BF0,
		*scriptCallbackQualifyTestWaypoint542BF0,
	]{
		firstObject: func() *scriptCallbackQualifyTestObject542BF0 {
			return firstObject
		},
		nextObject: func(object *scriptCallbackQualifyTestObject542BF0) *scriptCallbackQualifyTestObject542BF0 {
			return object.next
		},
		loadObjectFlags: func(object *scriptCallbackQualifyTestObject542BF0) uint32 {
			return object.flags
		},
		storeObjectFlags: func(object *scriptCallbackQualifyTestObject542BF0, flags uint32) {
			object.flags = flags
		},
		loadObjectID: func(object *scriptCallbackQualifyTestObject542BF0) (string, bool) {
			if object.id == nil {
				return "", false
			}
			return *object.id, true
		},
		storeObjectID: func(object *scriptCallbackQualifyTestObject542BF0, value string) {
			object.id = scriptCallbackQualifyTestString542BF0(value)
		},
		loadCallbackName: func(object *scriptCallbackQualifyTestObject542BF0, event int) (string, bool) {
			name, ok := object.callbacks[event]
			if !ok || name == nil {
				return "", false
			}
			return *name, true
		},
		objectKind: func(object *scriptCallbackQualifyTestObject542BF0) scriptCallbackObjectKind542BF0 {
			return object.kind
		},
		storeCallbackName: func(object *scriptCallbackQualifyTestObject542BF0, event int, name string) {
			object.callbacks[event] = scriptCallbackQualifyTestString542BF0(name)
			if stores != nil {
				*stores = append(*stores, scriptCallbackQualifyTestStore542BF0{
					object: object,
					event:  event,
					name:   name,
				})
			}
		},
		firstWaypoint: func() *scriptCallbackQualifyTestWaypoint542BF0 {
			return firstWaypoint
		},
		nextWaypoint: func(waypoint *scriptCallbackQualifyTestWaypoint542BF0) *scriptCallbackQualifyTestWaypoint542BF0 {
			return waypoint.next
		},
		loadWaypointFlags: func(waypoint *scriptCallbackQualifyTestWaypoint542BF0) uint32 {
			return waypoint.flags
		},
		storeWaypointFlags: func(waypoint *scriptCallbackQualifyTestWaypoint542BF0, flags uint32) {
			waypoint.flags = flags
		},
		loadWaypointName: func(waypoint *scriptCallbackQualifyTestWaypoint542BF0) string {
			return waypoint.name
		},
		storeWaypointName: func(waypoint *scriptCallbackQualifyTestWaypoint542BF0, name string) {
			waypoint.name = name
		},
	}
}

func TestScriptCallbackQualifyNames542BF0PreserveSignedDwordsAndExactBounds(t *testing.T) {
	if got, want := scriptCallbackQualifyName542BF0(
		"Callback", math.MinInt32, -1, math.MaxInt32,
	), "Callback%-2147483648%-1%2147483647"; got != want {
		t.Fatalf("callback name = %q, want %q", got, want)
	}
	if got := scriptCallbackQualifyName542BF0(strings.Repeat("a", 121), 0, 0, 0); len(got) != 127 {
		t.Fatalf("127-byte callback result = %q (len %d), want preserved", got, len(got))
	}
	if got := scriptCallbackQualifyName542BF0(strings.Repeat("a", 122), 0, 0, 0); got != scriptCallbackError542BF0 {
		t.Fatalf("128-byte callback result = %q, want %q", got, scriptCallbackError542BF0)
	}

	if got, want := scriptObjectQualifyName542BF0("Object", math.MinInt32), "Object%-2147483648"; got != want {
		t.Fatalf("object name = %q, want %q", got, want)
	}
	if got := scriptObjectQualifyName542BF0(strings.Repeat("b", 73), 0); len(got) != 75 {
		t.Fatalf("75-byte object result = %q (len %d), want preserved", got, len(got))
	}
	if got := scriptObjectQualifyName542BF0(strings.Repeat("b", 74), 0); got != scriptCallbackError542BF0 {
		t.Fatalf("76-byte object result = %q, want %q", got, scriptCallbackError542BF0)
	}
}

func TestScriptCallbackQualify542BF0RestoresObjectAndWaypointOrder(t *testing.T) {
	unknown := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0 | 0x40000123,
		id:        scriptCallbackQualifyTestString542BF0("Unknown"),
		callbacks: map[int]*string{14: scriptCallbackQualifyTestString542BF0("OnGlobal")},
	}
	trigger := &scriptCallbackQualifyTestObject542BF0{
		flags: scriptCallbackMarkedFlag542BF0 | 0x45,
		id:    scriptCallbackQualifyTestString542BF0("Trigger"),
		kind:  scriptCallbackObjectTrigger542BF0,
		callbacks: map[int]*string{
			1: scriptCallbackQualifyTestString542BF0("OnPressed"),
			2: scriptCallbackQualifyTestString542BF0(""),
			0: scriptCallbackQualifyTestString542BF0("OnReleased"),
		},
	}
	monster := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0 | 0x66,
		kind:      scriptCallbackObjectMonster542BF0,
		callbacks: make(map[int]*string),
	}
	for _, event := range scriptCallbackMonsterEvents542BF0 {
		monster.callbacks[event] = scriptCallbackQualifyTestString542BF0("Monster" + string(rune('A'+event)))
	}
	hole := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0 | 0x77,
		kind:      scriptCallbackObjectHole542BF0,
		callbacks: map[int]*string{12: scriptCallbackQualifyTestString542BF0("OnEnter")},
	}
	generator := &scriptCallbackQualifyTestObject542BF0{
		flags: scriptCallbackMarkedFlag542BF0 | 0x88,
		id:    scriptCallbackQualifyTestString542BF0(""),
		kind:  scriptCallbackObjectGenerator542BF0,
		callbacks: map[int]*string{
			15: scriptCallbackQualifyTestString542BF0("OnGenerate"),
			18: scriptCallbackQualifyTestString542BF0(""),
		},
	}
	unmarked := &scriptCallbackQualifyTestObject542BF0{
		flags:     0x99,
		id:        scriptCallbackQualifyTestString542BF0("Unmarked"),
		kind:      scriptCallbackObjectTrigger542BF0,
		callbacks: map[int]*string{1: scriptCallbackQualifyTestString542BF0("Untouched")},
	}
	unknown.next = trigger
	trigger.next = monster
	monster.next = hole
	hole.next = generator
	generator.next = unmarked

	waypoint := &scriptCallbackQualifyTestWaypoint542BF0{
		flags: scriptCallbackMarkedFlag542BF0 | 0x123,
		name:  "Arrival",
	}
	emptyWaypoint := &scriptCallbackQualifyTestWaypoint542BF0{
		flags: scriptCallbackMarkedFlag542BF0 | 0x456,
	}
	unmarkedWaypoint := &scriptCallbackQualifyTestWaypoint542BF0{
		flags: 0x789,
		name:  "UntouchedWaypoint",
	}
	waypoint.next = emptyWaypoint
	emptyWaypoint.next = unmarkedWaypoint

	var stores []scriptCallbackQualifyTestStore542BF0
	scriptCallbackQualify542BF0(
		-7, 8, 9,
		scriptCallbackQualifyTestHooks542BF0(unknown, waypoint, &stores),
	)

	var gotEvents []int
	for _, store := range stores {
		gotEvents = append(gotEvents, store.event)
		if !strings.HasSuffix(store.name, "%-7%8%9") {
			t.Errorf("event %d name = %q, want callback qualifier", store.event, store.name)
		}
	}
	wantEvents := []int{14, 1, 0, 3, 5, 4, 6, 7, 8, 9, 10, 11, 12, 15}
	if !reflect.DeepEqual(gotEvents, wantEvents) {
		t.Fatalf("stored event order = %v, want %v", gotEvents, wantEvents)
	}
	if got, want := *unknown.id, "Unknown%-7"; got != want {
		t.Errorf("unknown ID = %q, want %q", got, want)
	}
	if got, want := *generator.id, "%-7"; got != want {
		t.Errorf("empty-but-present generator ID = %q, want %q", got, want)
	}
	if got := *unmarked.id; got != "Unmarked" {
		t.Errorf("unmarked ID = %q, want untouched", got)
	}
	for name, object := range map[string]*scriptCallbackQualifyTestObject542BF0{
		"unknown": unknown, "trigger": trigger, "monster": monster, "hole": hole, "generator": generator,
	} {
		if object.flags&scriptCallbackMarkedFlag542BF0 != 0 {
			t.Errorf("%s flags = %#x, marked bit was not cleared", name, object.flags)
		}
	}
	if unmarked.flags != 0x99 {
		t.Errorf("unmarked flags = %#x, want %#x", unmarked.flags, uint32(0x99))
	}
	if got, want := waypoint.name, "Arrival%-7"; got != want {
		t.Errorf("waypoint name = %q, want %q", got, want)
	}
	if waypoint.flags != 0x123 || emptyWaypoint.flags != 0x456 {
		t.Errorf("marked waypoint flags = %#x/%#x, want 0x123/0x456", waypoint.flags, emptyWaypoint.flags)
	}
	if unmarkedWaypoint.name != "UntouchedWaypoint" || unmarkedWaypoint.flags != 0x789 {
		t.Errorf("unmarked waypoint = %q/%#x, want untouched", unmarkedWaypoint.name, unmarkedWaypoint.flags)
	}
}

func TestScriptCallbackQualify542BF0UsesLiveSuccessorAfterMutation(t *testing.T) {
	first := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0,
		callbacks: map[int]*string{14: scriptCallbackQualifyTestString542BF0("First")},
	}
	skipped := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0,
		callbacks: map[int]*string{14: scriptCallbackQualifyTestString542BF0("Skipped")},
	}
	live := &scriptCallbackQualifyTestObject542BF0{
		flags:     scriptCallbackMarkedFlag542BF0,
		callbacks: map[int]*string{14: scriptCallbackQualifyTestString542BF0("Live")},
	}
	first.next = skipped
	skipped.next = live

	hooks := scriptCallbackQualifyTestHooks542BF0(first, nil, nil)
	baseStore := hooks.storeCallbackName
	var visited []string
	hooks.storeCallbackName = func(object *scriptCallbackQualifyTestObject542BF0, event int, name string) {
		visited = append(visited, strings.Split(name, "%")[0])
		baseStore(object, event, name)
		if object == first {
			first.next = live
		}
	}
	scriptCallbackQualify542BF0(1, 2, 3, hooks)
	if want := []string{"First", "Live"}; !reflect.DeepEqual(visited, want) {
		t.Fatalf("visited callbacks = %v, want live successor %v", visited, want)
	}
}

func TestScriptCallbackQualify542BF0RetainsRequiredCallbackContract(t *testing.T) {
	for name, kind := range map[string]scriptCallbackObjectKind542BF0{
		"trigger": scriptCallbackObjectTrigger542BF0,
		"monster": scriptCallbackObjectMonster542BF0,
		"hole":    scriptCallbackObjectHole542BF0,
	} {
		t.Run(name, func(t *testing.T) {
			object := &scriptCallbackQualifyTestObject542BF0{
				flags:     scriptCallbackMarkedFlag542BF0,
				kind:      kind,
				callbacks: map[int]*string{},
			}
			defer func() {
				if got := recover(); got != "sub_542BF0: required callback name is nil" {
					t.Fatalf("panic = %#v, want required callback contract", got)
				}
			}()
			scriptCallbackQualify542BF0(
				1, 2, 3,
				scriptCallbackQualifyTestHooks542BF0(object, nil, nil),
			)
		})
	}
}
