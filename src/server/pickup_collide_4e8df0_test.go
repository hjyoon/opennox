package server

import (
	"fmt"
	"reflect"
	"testing"
)

type pickupCollideTestObject4E8DF0 struct {
	name        string
	class       uint8
	pickupFrame uint32
	update      *pickupCollideTestUpdate4E8DF0
}

type pickupCollideTestUpdate4E8DF0 struct {
	movement uint8
}

type pickupCollideTestState4E8DF0 struct {
	events      []string
	frame       uint32
	fps         uint32
	placeResult uint32
	onFrame     func()
}

func (s *pickupCollideTestState4E8DF0) hooks() pickupCollideHooks4E8DF0[
	*pickupCollideTestObject4E8DF0,
	*pickupCollideTestUpdate4E8DF0,
] {
	return pickupCollideHooks4E8DF0[
		*pickupCollideTestObject4E8DF0,
		*pickupCollideTestUpdate4E8DF0,
	]{
		loadClassByte: func(obj *pickupCollideTestObject4E8DF0) uint8 {
			s.events = append(s.events, "class:"+obj.name)
			return obj.class
		},
		frame: func() uint32 {
			s.events = append(s.events, "frame")
			value := s.frame
			if s.onFrame != nil {
				s.onFrame()
			}
			return value
		},
		loadPickupFrame: func(obj *pickupCollideTestObject4E8DF0) uint32 {
			s.events = append(s.events, "pickup")
			return obj.pickupFrame
		},
		fps: func() uint32 {
			s.events = append(s.events, "fps")
			return s.fps
		},
		loadUpdateData: func(obj *pickupCollideTestObject4E8DF0) *pickupCollideTestUpdate4E8DF0 {
			s.events = append(s.events, "update:"+obj.name)
			return obj.update
		},
		loadMovementFlagsByte: func(update *pickupCollideTestUpdate4E8DF0) uint8 {
			s.events = append(s.events, "movement")
			return update.movement
		},
		placeInventory: func(unit, item *pickupCollideTestObject4E8DF0, flag1, flag2 int32) uint32 {
			s.events = append(s.events, fmt.Sprintf("place:%s:%s:%d:%d", unit.name, item.name, flag1, flag2))
			return s.placeResult
		},
	}
}

func TestPickupCollide4E8DF0NilAndMonsterReturnOriginalUnit(t *testing.T) {
	state := &pickupCollideTestState4E8DF0{}
	got := pickupCollide4E8DF0(
		(*pickupCollideTestObject4E8DF0)(nil),
		(*pickupCollideTestObject4E8DF0)(nil),
		&struct{}{},
		state.hooks(),
	)
	if got.unit != nil || got.inventoryAttempted || len(state.events) != 0 {
		t.Fatalf("nil result = %+v, events = %v", got, state.events)
	}

	unit := &pickupCollideTestObject4E8DF0{name: "monster", class: pickupCollideMonsterClassByte4E8DF0}
	got = pickupCollide4E8DF0(
		(*pickupCollideTestObject4E8DF0)(nil),
		unit,
		&struct{}{},
		state.hooks(),
	)
	if got.unit != unit || got.inventoryAttempted || got.inventoryResult != 0 {
		t.Fatalf("monster result = %+v, want untouched unit", got)
	}
	if want := []string{"class:monster"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestPickupCollide4E8DF0UnsignedAgeAndLogicalFPSBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		frame       uint32
		pickupFrame uint32
		fps         uint32
		attempted   bool
	}{
		{name: "below half", frame: 101, pickupFrame: 100, fps: 4, attempted: false},
		{name: "exact half", frame: 102, pickupFrame: 100, fps: 5, attempted: true},
		{name: "unsigned wrap exact", frame: 1, pickupFrame: ^uint32(0), fps: 5, attempted: true},
		{name: "logical high bit", frame: 0, pickupFrame: 0xc0000001, fps: 0x80000000, attempted: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item := &pickupCollideTestObject4E8DF0{name: "item", pickupFrame: tc.pickupFrame}
			unit := &pickupCollideTestObject4E8DF0{name: "unit"}
			state := &pickupCollideTestState4E8DF0{
				frame: tc.frame, fps: tc.fps, placeResult: 0x89abcdef,
			}
			got := pickupCollide4E8DF0(item, unit, uintptr(0xfeedface), state.hooks())
			if got.unit != unit || got.inventoryAttempted != tc.attempted {
				t.Fatalf("result = %+v, attempted want %v", got, tc.attempted)
			}
			want := []string{"class:unit", "frame", "pickup", "fps"}
			if tc.attempted {
				want = append(want, "place:unit:item:1:1")
				if got.inventoryResult != 0x89abcdef {
					t.Fatalf("inventory result = %#x", got.inventoryResult)
				}
			} else if got.inventoryResult != 0 {
				t.Fatalf("guard result = %#x, want zero inventory word", got.inventoryResult)
			}
			if !reflect.DeepEqual(state.events, want) {
				t.Fatalf("events = %v, want %v", state.events, want)
			}
		})
	}
}

func TestPickupCollide4E8DF0PlayerMovementAndCachedClass(t *testing.T) {
	item := &pickupCollideTestObject4E8DF0{name: "item"}
	unit := &pickupCollideTestObject4E8DF0{
		name: "player", class: pickupCollidePlayerClassByte4E8DF0,
		update: &pickupCollideTestUpdate4E8DF0{},
	}
	state := &pickupCollideTestState4E8DF0{frame: 10, fps: 20, placeResult: 1}
	state.onFrame = func() {
		// ECX retains the entry class byte across the original frame/FPS
		// loads, so this live mutation must not bypass the Player gate.
		unit.class = 0
	}

	got := pickupCollide4E8DF0(item, unit, (*uint32)(nil), state.hooks())
	if got.unit != unit || got.inventoryAttempted {
		t.Fatalf("result = %+v, want cached-Player movement rejection", got)
	}
	want := []string{"class:player", "frame", "pickup", "fps", "update:player", "movement"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}

	unit.class = pickupCollidePlayerClassByte4E8DF0
	unit.update.movement = 0x81
	state.events = nil
	state.onFrame = nil
	state.placeResult = 0
	got = pickupCollide4E8DF0(item, unit, (*uint32)(nil), state.hooks())
	if !got.inventoryAttempted || got.inventoryResult != 0 {
		t.Fatalf("false inventory result = %+v, want attempted zero", got)
	}
	want = append(want[:0], "class:player", "frame", "pickup", "fps", "update:player", "movement", "place:player:item:1:1")
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestPickupCollide4E8DF0NonPlayerSkipsUpdateData(t *testing.T) {
	item := &pickupCollideTestObject4E8DF0{name: "item", pickupFrame: 7}
	unit := &pickupCollideTestObject4E8DF0{name: "unit", class: 0x80}
	state := &pickupCollideTestState4E8DF0{frame: 8, fps: 2, placeResult: 7}

	got := pickupCollide4E8DF0(item, unit, "ignored", state.hooks())
	if !got.inventoryAttempted || got.inventoryResult != 7 {
		t.Fatalf("result = %+v, want direct inventory attempt", got)
	}
	want := []string{"class:unit", "frame", "pickup", "fps", "place:unit:item:1:1"}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestPickupCollide4E8DF0NilItemFaultsBeforeFPSRead(t *testing.T) {
	unit := &pickupCollideTestObject4E8DF0{name: "unit"}
	state := &pickupCollideTestState4E8DF0{frame: 10, fps: 20}
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not fault at pickup-frame load")
		}
		want := []string{"class:unit", "frame", "pickup"}
		if !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()
	pickupCollide4E8DF0(
		(*pickupCollideTestObject4E8DF0)(nil),
		unit,
		(*uint32)(nil),
		state.hooks(),
	)
}
