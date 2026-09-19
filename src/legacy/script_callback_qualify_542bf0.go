package legacy

import (
	"strconv"
	"strings"
)

const (
	scriptCallbackMarkedFlag542BF0 = uint32(0x80000000)
	scriptCallbackLiveFlags542BF0  = uint32(0x7fffffff)
	scriptCallbackError542BF0      = "ERROR_NAME_TOO_LONG!"
)

type scriptCallbackObjectKind542BF0 uint8

const (
	scriptCallbackObjectUnknown542BF0 scriptCallbackObjectKind542BF0 = iota
	scriptCallbackObjectTrigger542BF0
	scriptCallbackObjectMonster542BF0
	scriptCallbackObjectHole542BF0
	scriptCallbackObjectGenerator542BF0
)

type scriptCallbackQualifyHooks542BF0[Object, Waypoint comparable] struct {
	firstObject        func() Object
	nextObject         func(Object) Object
	loadObjectFlags    func(Object) uint32
	storeObjectFlags   func(Object, uint32)
	loadObjectID       func(Object) (string, bool)
	storeObjectID      func(Object, string)
	loadCallbackName   func(Object, int) (string, bool)
	objectKind         func(Object) scriptCallbackObjectKind542BF0
	storeCallbackName  func(Object, int, string)
	firstWaypoint      func() Waypoint
	nextWaypoint       func(Waypoint) Waypoint
	loadWaypointFlags  func(Waypoint) uint32
	storeWaypointFlags func(Waypoint, uint32)
	loadWaypointName   func(Waypoint) string
	storeWaypointName  func(Waypoint, string)
}

var (
	scriptCallbackTriggerEvents542BF0   = [...]int{1, 2, 0}
	scriptCallbackMonsterEvents542BF0   = [...]int{3, 5, 4, 6, 7, 8, 9, 10, 11}
	scriptCallbackHoleEvents542BF0      = [...]int{12}
	scriptCallbackGeneratorEvents542BF0 = [...]int{15, 16, 18, 17}
)

func scriptCallbackAppendQualifier542BF0(dst *strings.Builder, value int32) {
	dst.WriteByte('%')
	dst.WriteString(strconv.FormatInt(int64(value), 10))
}

func scriptCallbackQualifyName542BF0(name string, a1, a2, a3 int32) string {
	var result strings.Builder
	result.Grow(len(name) + 36)
	result.WriteString(name)
	scriptCallbackAppendQualifier542BF0(&result, a1)
	scriptCallbackAppendQualifier542BF0(&result, a2)
	scriptCallbackAppendQualifier542BF0(&result, a3)
	if result.Len() >= 128 {
		return scriptCallbackError542BF0
	}
	return result.String()
}

func scriptObjectQualifyName542BF0(name string, value int32) string {
	var result strings.Builder
	result.Grow(len(name) + 12)
	result.WriteString(name)
	scriptCallbackAppendQualifier542BF0(&result, value)
	if result.Len() >= 76 {
		return scriptCallbackError542BF0
	}
	return result.String()
}

func scriptCallbackQualifyEvents542BF0[Object comparable](
	object Object,
	events []int,
	a1, a2, a3 int32,
	nilAllowed bool,
	load func(Object, int) (string, bool),
	store func(Object, int, string),
) {
	for _, event := range events {
		name, ok := load(object, event)
		if !ok {
			if nilAllowed {
				continue
			}
			panic("sub_542BF0: required callback name is nil")
		}
		if name != "" {
			store(object, event, scriptCallbackQualifyName542BF0(name, a1, a2, a3))
		}
	}
}

// scriptCallbackQualify542BF0 preserves GAME.EXE 00542BF0. The object-list
// successor is loaded after each marked object has been rewritten. Event 14
// and monster-generator callbacks accept a nil name; trigger, monster, and
// hole callback slots retain the original required-name contract.
func scriptCallbackQualify542BF0[Object, Waypoint comparable](
	a1, a2, a3 int32,
	hooks scriptCallbackQualifyHooks542BF0[Object, Waypoint],
) {
	var zeroObject Object
	for object := hooks.firstObject(); object != zeroObject; object = hooks.nextObject(object) {
		if hooks.loadObjectFlags(object)&scriptCallbackMarkedFlag542BF0 == 0 {
			continue
		}
		if id, ok := hooks.loadObjectID(object); ok {
			hooks.storeObjectID(object, scriptObjectQualifyName542BF0(id, a1))
		}

		if name, ok := hooks.loadCallbackName(object, 14); ok && name != "" {
			hooks.storeCallbackName(object, 14, scriptCallbackQualifyName542BF0(name, a1, a2, a3))
		}

		var events []int
		nilAllowed := false
		switch hooks.objectKind(object) {
		case scriptCallbackObjectTrigger542BF0:
			events = scriptCallbackTriggerEvents542BF0[:]
		case scriptCallbackObjectMonster542BF0:
			events = scriptCallbackMonsterEvents542BF0[:]
		case scriptCallbackObjectHole542BF0:
			events = scriptCallbackHoleEvents542BF0[:]
		case scriptCallbackObjectGenerator542BF0:
			events = scriptCallbackGeneratorEvents542BF0[:]
			nilAllowed = true
		}
		scriptCallbackQualifyEvents542BF0(
			object, events, a1, a2, a3, nilAllowed,
			hooks.loadCallbackName, hooks.storeCallbackName,
		)
		hooks.storeObjectFlags(object, hooks.loadObjectFlags(object)&scriptCallbackLiveFlags542BF0)
	}

	var zeroWaypoint Waypoint
	for waypoint := hooks.firstWaypoint(); waypoint != zeroWaypoint; waypoint = hooks.nextWaypoint(waypoint) {
		if hooks.loadWaypointFlags(waypoint)&scriptCallbackMarkedFlag542BF0 == 0 {
			continue
		}
		if name := hooks.loadWaypointName(waypoint); name != "" {
			hooks.storeWaypointName(waypoint, scriptObjectQualifyName542BF0(name, a1))
		}
		hooks.storeWaypointFlags(waypoint, hooks.loadWaypointFlags(waypoint)&scriptCallbackLiveFlags542BF0)
	}
}
