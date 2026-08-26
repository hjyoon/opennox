package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func elevatorTestObject53B5D0(t *testing.T, shaft bool) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	if shaft {
		update := new(ElevatorShaftUpdateData)
		unit.UpdateData = unsafe.Pointer(update)
		unit.ObjClass = object.ClassElevatorShaft
	} else {
		update := new(ElevatorUpdateData)
		unit.UpdateData = unsafe.Pointer(update)
		unit.ObjClass = object.ClassElevator
	}
	return unit
}

func elevatorTestHooks53B5D0(frame, tickRate uint32, links map[*Object]*Object, events *[]string) elevatorUpdateHooks53B5D0 {
	return elevatorUpdateHooks53B5D0{
		frame:    func() uint32 { return frame },
		tickRate: func() uint32 { return tickRate },
		link:     func(unit *Object) *Object { return links[unit] },
		needSync: func(unit *Object) { *events = append(*events, fmt.Sprintf("sync:%p", unit)) },
		audio: func(unit *Object, upward bool) {
			*events = append(*events, fmt.Sprintf("audio:%p:%t", unit, upward))
		},
		eachInCircle: func(types.Pointf, float32, func(*Object) bool) {},
		pointInBox:   func(*types.Pointf, *Shape, *types.Pointf) bool { return true },
		move: func(unit *Object, pos types.Pointf) {
			*events = append(*events, fmt.Sprintf("move:%p:%v", unit, pos))
		},
		raise: func(unit *Object, z float32) {
			*events = append(*events, fmt.Sprintf("raise:%p:%g", unit, z))
		},
		queueCollision: func(unit *Object) {
			*events = append(*events, fmt.Sprintf("collision:%p", unit))
		},
	}
}

func TestElevatorUpdate53B5D0StateCycle(t *testing.T) {
	elevator := elevatorTestObject53B5D0(t, false)
	shaft := elevatorTestObject53B5D0(t, true)
	elevator.ObjFlags = object.FlagEnabled | object.FlagNoCollide
	links := map[*Object]*Object{elevator: shaft, shaft: elevator}
	var events []string

	hooks := elevatorTestHooks53B5D0(31, 30, links, &events)
	elevatorUpdate53B5D0(elevator, hooks)
	update := elevator.UpdateDataElevator()
	if update.Field_3 != 3 || fmt.Sprint(events) != fmt.Sprint([]string{fmt.Sprintf("audio:%p:true", elevator)}) {
		t.Fatalf("start-up = state %d events %v", update.Field_3, events)
	}

	events = nil
	hooks = elevatorTestHooks53B5D0(32, 30, links, &events)
	update.Field_4 = 18
	elevatorUpdate53B5D0(elevator, hooks)
	if update.Field_4 != 20 || elevator.ObjFlags.Has(object.FlagNoCollide) {
		t.Fatalf("rising = height %d flags %#x", update.Field_4, elevator.ObjFlags)
	}
	wantSync := []string{fmt.Sprintf("sync:%p", elevator), fmt.Sprintf("sync:%p", shaft)}
	if fmt.Sprint(events) != fmt.Sprint(wantSync) {
		t.Fatalf("rising events = %v, want %v", events, wantSync)
	}

	update.Field_4 = 63
	elevatorUpdate53B5D0(elevator, hooks)
	if update.Field_4 != 64 || update.Field_3 != 2 || elevator.Field34 != 32 {
		t.Fatalf("upper stop = (%d,%d,%d), want (64,2,32)", update.Field_4, update.Field_3, elevator.Field34)
	}

	events = nil
	hooks = elevatorTestHooks53B5D0(63, 30, links, &events)
	elevatorUpdate53B5D0(elevator, hooks)
	if update.Field_3 != 1 || fmt.Sprint(events) != fmt.Sprint([]string{fmt.Sprintf("audio:%p:false", elevator)}) {
		t.Fatalf("start-down = state %d events %v", update.Field_3, events)
	}

	update.Field_4 = 2
	events = nil
	elevatorUpdate53B5D0(elevator, hooks)
	if update.Field_4 != 0 || !elevator.ObjFlags.Has(object.FlagNoCollide) {
		t.Fatalf("lowering = height %d flags %#x", update.Field_4, elevator.ObjFlags)
	}
	elevatorUpdate53B5D0(elevator, hooks)
	if update.Field_3 != 0 || elevator.Field34 != 63 {
		t.Fatalf("lower stop = (%d,%d), want (0,63)", update.Field_3, elevator.Field34)
	}
}

func TestElevatorShaftUpdate53B380TracksElevator(t *testing.T) {
	elevator := elevatorTestObject53B5D0(t, false)
	shaft := elevatorTestObject53B5D0(t, true)
	elevator.UpdateDataElevator().Field_3 = 3
	links := map[*Object]*Object{shaft: elevator}
	var events []string
	hooks := elevatorTestHooks53B5D0(1, 30, links, &events)
	elevatorShaftUpdate53B380(shaft, hooks)
	want := []string{fmt.Sprintf("collision:%p", shaft), fmt.Sprintf("audio:%p:true", shaft)}
	if shaft.UpdateDataElevatorShaft().Field_3 != 3 || fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("shaft update = state %d events %v, want %v", shaft.UpdateDataElevatorShaft().Field_3, events, want)
	}
}

func TestElevatorSound53B490(t *testing.T) {
	tests := []struct {
		material uint16
		subclass object.SubClass
		up, down uint32
	}{
		{8, 0, 257, 258},
		{16, 0x20, 253, 254},
		{16, 0x40, 259, 260},
		{16, 0, 251, 252},
		{32, 0x2, 255, 256},
		{32, 0, 249, 250},
	}
	for _, tc := range tests {
		unit := &Object{Material: tc.material, ObjSubClass: tc.subclass}
		if up, down := uint32(elevatorSound53B490(unit, true)), uint32(elevatorSound53B490(unit, false)); up != tc.up || down != tc.down {
			t.Errorf("material %d subclass %#x = (%d,%d), want (%d,%d)", tc.material, tc.subclass, up, down, tc.up, tc.down)
		}
	}
}
