package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type pentagramUpdateTestData53BEF0 struct {
	state          uint8
	triggered      uint32
	animationFrame uint8
	animationTick  uint8
	animationStep  uint8
}

type pentagramUpdateTestObject53BEF0 struct {
	name        string
	data        *pentagramUpdateTestData53BEF0
	destination *pentagramUpdateTestObject53BEF0
	position    types.Pointf
	radius      float32
	class       uint32
	enabled     bool
	field34     uint32
}

func pentagramUpdateTestHooks53BEF0(
	events *[]string,
	frames *[]uint32,
	units []*pentagramUpdateTestObject53BEF0,
) pentagramUpdateHooks53BEF0[
	*pentagramUpdateTestObject53BEF0,
	*pentagramUpdateTestData53BEF0,
	*types.Pointf,
] {
	event := func(format string, args ...any) {
		*events = append(*events, fmt.Sprintf(format, args...))
	}
	return pentagramUpdateHooks53BEF0[
		*pentagramUpdateTestObject53BEF0,
		*pentagramUpdateTestData53BEF0,
		*types.Pointf,
	]{
		loadUpdate: func(obj *pentagramUpdateTestObject53BEF0) *pentagramUpdateTestData53BEF0 {
			return obj.data
		},
		loadState: func(data *pentagramUpdateTestData53BEF0) uint8 {
			return data.state
		},
		storeState: func(data *pentagramUpdateTestData53BEF0, value uint8) {
			data.state = value
			event("state:%d", value)
		},
		loadTriggered: func(data *pentagramUpdateTestData53BEF0) uint32 {
			return data.triggered
		},
		storeTriggered: func(data *pentagramUpdateTestData53BEF0, value uint32) {
			data.triggered = value
			event("trigger:%d", value)
		},
		loadAnimationFrame: func(data *pentagramUpdateTestData53BEF0) uint8 {
			return data.animationFrame
		},
		storeAnimationFrame: func(data *pentagramUpdateTestData53BEF0, value uint8) {
			data.animationFrame = value
			event("animation-frame:%d", value)
		},
		loadAnimationTick: func(data *pentagramUpdateTestData53BEF0) uint8 {
			return data.animationTick
		},
		storeAnimationTick: func(data *pentagramUpdateTestData53BEF0, value uint8) {
			data.animationTick = value
			event("animation-tick:%d", value)
		},
		loadAnimationStep: func(data *pentagramUpdateTestData53BEF0) uint8 {
			return data.animationStep
		},
		storeAnimationStep: func(data *pentagramUpdateTestData53BEF0, value uint8) {
			data.animationStep = value
			event("animation-step:%d", value)
		},
		needSync: func(obj *pentagramUpdateTestObject53BEF0) {
			event("sync:%s", obj.name)
		},
		loadDestination: func(
			obj *pentagramUpdateTestObject53BEF0,
			_ *pentagramUpdateTestData53BEF0,
		) *pentagramUpdateTestObject53BEF0 {
			event("destination:%s", obj.name)
			return obj.destination
		},
		loadRadius: func(obj *pentagramUpdateTestObject53BEF0) float32 {
			return obj.radius
		},
		loadPosX: func(obj *pentagramUpdateTestObject53BEF0) float32 {
			return obj.position.X
		},
		loadPosY: func(obj *pentagramUpdateTestObject53BEF0) float32 {
			return obj.position.Y
		},
		cachePosition: func(obj *pentagramUpdateTestObject53BEF0) *types.Pointf {
			event("position:%s", obj.name)
			return &obj.position
		},
		eachInRect: func(rect types.Rectf, callback func(*pentagramUpdateTestObject53BEF0)) {
			event("rect:%g,%g:%g,%g", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
			for _, unit := range units {
				callback(unit)
			}
		},
		teleportVisible: func(unit *pentagramUpdateTestObject53BEF0, destination *types.Pointf) {
			event("visible:%s:%g,%g", unit.name, destination.X, destination.Y)
		},
		teleportInvisible: func(unit *pentagramUpdateTestObject53BEF0, destination *types.Pointf) {
			event("invisible:%s:%g,%g", unit.name, destination.X, destination.Y)
		},
		isEnabled: func(obj *pentagramUpdateTestObject53BEF0) bool {
			event("enabled:%s", obj.name)
			return obj.enabled
		},
		frame: func() uint32 {
			value := (*frames)[0]
			*frames = (*frames)[1:]
			event("frame:%d", value)
			return value
		},
		storeField34: func(obj *pentagramUpdateTestObject53BEF0, value uint32) {
			obj.field34 = value
			event("field34:%s:%d", obj.name, value)
		},
	}
}

func TestPentagramTeleport53C060OrderAndLivePosition(t *testing.T) {
	unit := &pentagramUpdateTestObject53BEF0{name: "unit", position: types.Ptf(1, 2)}
	destination := types.Ptf(30, 40)
	var events []string
	hooks := pentagramTeleportHooks53C060[*pentagramUpdateTestObject53BEF0, *types.Pointf]{
		loadClass: func(obj *pentagramUpdateTestObject53BEF0) uint32 {
			events = append(events, "class:"+obj.name)
			return obj.class
		},
		cachePosition: func(obj *pentagramUpdateTestObject53BEF0) *types.Pointf {
			events = append(events, "position:"+obj.name)
			return &obj.position
		},
		pointFX: func(id uint32, position *types.Pointf) {
			events = append(events, fmt.Sprintf("fx:%d:%g,%g", id, position.X, position.Y))
		},
		audio: func(id uint32, obj *pentagramUpdateTestObject53BEF0) {
			events = append(events, fmt.Sprintf("audio:%d:%s", id, obj.name))
		},
		teleport: func(obj *pentagramUpdateTestObject53BEF0, target *types.Pointf) {
			events = append(events, fmt.Sprintf("teleport:%s:%g,%g", obj.name, target.X, target.Y))
			obj.position = *target
		},
	}
	pentagramTeleport53C060(unit, &destination, hooks)
	want := []string{
		"class:unit",
		"position:unit",
		"fx:137:1,2",
		"audio:147:unit",
		"teleport:unit:30,40",
		"fx:137:30,40",
		"audio:147:unit",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}

	for _, class := range []uint32{0x00020000, 0x00400000, 0x00420000} {
		events = nil
		unit.class = class
		pentagramTeleport53C060(unit, &destination, hooks)
		if !reflect.DeepEqual(events, []string{"class:unit"}) {
			t.Fatalf("class %#x events = %#v", class, events)
		}
	}
}

func TestPentagramTeleportInvisible53C140ClassGate(t *testing.T) {
	destination := types.Ptf(7, 9)
	for _, tc := range []struct {
		class uint32
		want  bool
	}{{0, true}, {0x00020000, false}, {0x00400000, false}} {
		called := false
		unit := &pentagramUpdateTestObject53BEF0{class: tc.class}
		pentagramTeleportInvisible53C140(
			unit,
			&destination,
			func(obj *pentagramUpdateTestObject53BEF0) uint32 { return obj.class },
			func(*pentagramUpdateTestObject53BEF0, *types.Pointf) { called = true },
		)
		if called != tc.want {
			t.Fatalf("class %#x called = %t, want %t", tc.class, called, tc.want)
		}
	}
}

func TestPentagramUpdate53BEF0ActivatesPairWithIndependentFrames(t *testing.T) {
	sourceData := &pentagramUpdateTestData53BEF0{triggered: 1, animationStep: 3}
	destinationData := &pentagramUpdateTestData53BEF0{state: 9, animationFrame: 8, animationTick: 7}
	destination := &pentagramUpdateTestObject53BEF0{name: "destination", data: destinationData}
	source := &pentagramUpdateTestObject53BEF0{
		name: "source", data: sourceData, destination: destination, enabled: true,
	}
	events := make([]string, 0, 16)
	frames := []uint32{100, 101}
	hooks := pentagramUpdateTestHooks53BEF0(&events, &frames, nil)

	if got := pentagramUpdate53BEF0(source, hooks); got != 101 {
		t.Fatalf("return = %d, want 101", got)
	}
	if sourceData.state != 1 || sourceData.animationFrame != 0 || sourceData.animationTick != 0 ||
		sourceData.animationStep != 0 || sourceData.triggered != 0 || source.field34 != 100 {
		t.Fatalf("source state = %+v field34=%d", sourceData, source.field34)
	}
	if destinationData.state != 2 || destinationData.animationFrame != 0 ||
		destinationData.animationTick != 0 || destination.field34 != 101 {
		t.Fatalf("destination state = %+v field34=%d", destinationData, destination.field34)
	}
	want := []string{
		"sync:source", "animation-step:0", "destination:source", "enabled:source",
		"state:1", "animation-frame:0", "animation-tick:0", "frame:100",
		"field34:source:100", "state:2", "animation-frame:0", "animation-tick:0",
		"frame:101", "field34:destination:101", "trigger:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestPentagramUpdate53BEF0TeleportsOnStepEight(t *testing.T) {
	sourceData := &pentagramUpdateTestData53BEF0{
		state: 1, triggered: 0x11223344, animationFrame: 0, animationTick: 0, animationStep: 7,
	}
	destination := &pentagramUpdateTestObject53BEF0{
		name: "destination", data: new(pentagramUpdateTestData53BEF0), position: types.Ptf(90, 80),
	}
	source := &pentagramUpdateTestObject53BEF0{
		name: "source", data: sourceData, destination: destination,
		position: types.Ptf(20, 30), radius: 5,
	}
	unitA := &pentagramUpdateTestObject53BEF0{name: "a"}
	unitB := &pentagramUpdateTestObject53BEF0{name: "b"}
	var events []string
	frames := []uint32(nil)
	hooks := pentagramUpdateTestHooks53BEF0(&events, &frames, []*pentagramUpdateTestObject53BEF0{unitA, unitB})

	if got := pentagramUpdate53BEF0(source, hooks); got != 1 {
		t.Fatalf("return = %d, want 1", got)
	}
	if sourceData.animationStep != 8 || sourceData.animationTick != 0 || sourceData.triggered != 0 {
		t.Fatalf("source state = %+v", sourceData)
	}
	want := []string{
		"animation-step:8", "sync:source", "animation-tick:0", "destination:source",
		"position:destination", "rect:15,25:25,35", "visible:a:90,80", "visible:b:90,80",
		"trigger:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestPentagramUpdate53BEF0TerminalAnimationStates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		state      uint8
		frame      uint8
		wantReturn int32
	}{
		{name: "outgoing", state: 1, frame: 4, wantReturn: 4},
		{name: "incoming", state: 2, frame: 4, wantReturn: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &pentagramUpdateTestData53BEF0{
				state: tc.state, triggered: 7, animationFrame: tc.frame, animationTick: 1, animationStep: 6,
			}
			obj := &pentagramUpdateTestObject53BEF0{name: tc.name, data: data}
			var events []string
			frames := []uint32(nil)
			hooks := pentagramUpdateTestHooks53BEF0(&events, &frames, nil)
			if got := pentagramUpdate53BEF0(obj, hooks); got != tc.wantReturn {
				t.Fatalf("return = %d, want %d", got, tc.wantReturn)
			}
			if data.state != 0 || data.triggered != 0 {
				t.Fatalf("data = %+v", data)
			}
			if tc.state == 2 && data.animationStep != 0 {
				t.Fatalf("incoming animation step = %d, want 0", data.animationStep)
			}
		})
	}
}

func TestPentagramInvisibleUpdate53C0C0EnumeratesAndConsumesTrigger(t *testing.T) {
	data := &pentagramUpdateTestData53BEF0{triggered: 0xffffffff}
	destination := &pentagramUpdateTestObject53BEF0{name: "destination", position: types.Ptf(70, 60)}
	source := &pentagramUpdateTestObject53BEF0{
		name: "source", data: data, destination: destination, enabled: true,
		position: types.Ptf(10, 20), radius: 3,
	}
	unit := &pentagramUpdateTestObject53BEF0{name: "unit"}
	var events []string
	frames := []uint32(nil)
	hooks := pentagramUpdateTestHooks53BEF0(&events, &frames, []*pentagramUpdateTestObject53BEF0{unit})

	if got := pentagramInvisibleUpdate53C0C0(source, hooks); got != 0 {
		t.Fatalf("return = %d, want 0", got)
	}
	if data.triggered != 0 {
		t.Fatalf("trigger = %#x, want 0", data.triggered)
	}
	want := []string{
		"destination:source", "enabled:source", "position:destination",
		"rect:7,17:13,23", "invisible:unit:70,60", "trigger:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestPentagramInvisibleUpdate53C0C0DisabledSkipsPosition(t *testing.T) {
	data := &pentagramUpdateTestData53BEF0{triggered: 1}
	source := &pentagramUpdateTestObject53BEF0{
		name: "source", data: data,
		destination: &pentagramUpdateTestObject53BEF0{name: "destination"},
	}
	var events []string
	frames := []uint32(nil)
	hooks := pentagramUpdateTestHooks53BEF0(&events, &frames, nil)
	pentagramInvisibleUpdate53C0C0(source, hooks)
	want := []string{"destination:source", "enabled:source", "trigger:0"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
