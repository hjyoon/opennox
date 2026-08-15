package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type crownUpdateTestObject53E1D0 struct {
	name      string
	flags     uint32
	radius    float32
	pos       types.Pointf
	direction [2]int16
	update    *crownUpdateTestData53E1D0
	owner     *crownUpdateTestObject53E1D0
}

type crownUpdateTestData53E1D0 struct {
	name   string
	field0 *crownUpdateTestObject53E1D0
	target *crownUpdateTestObject53E1D0
}

type crownUpdateTestState53E1D0 struct {
	events           []string
	traceResult      bool
	directionLoads   int
	positionXLoads   int
	positionYLoads   int
	afterTargetLoad  func()
	afterField0Clear func()
	tracedFrom       types.Pointf
	tracedTo         types.Pointf
	movedTo          types.Pointf
}

func crownUpdateTestObjectName53E1D0(obj *crownUpdateTestObject53E1D0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (s *crownUpdateTestState53E1D0) hooks() crownUpdateHooks53E1D0[
	*crownUpdateTestObject53E1D0,
	*crownUpdateTestData53E1D0,
] {
	return crownUpdateHooks53E1D0[
		*crownUpdateTestObject53E1D0,
		*crownUpdateTestData53E1D0,
	]{
		loadUpdate: func(obj *crownUpdateTestObject53E1D0) *crownUpdateTestData53E1D0 {
			s.events = append(s.events, "update:"+crownUpdateTestObjectName53E1D0(obj))
			return obj.update
		},
		loadPickupTarget: func(data *crownUpdateTestData53E1D0) *crownUpdateTestObject53E1D0 {
			s.events = append(s.events, "target:"+data.name)
			if s.afterTargetLoad != nil {
				s.afterTargetLoad()
			}
			return data.target
		},
		loadFlags: func(obj *crownUpdateTestObject53E1D0) uint32 {
			s.events = append(s.events, "flags:"+crownUpdateTestObjectName53E1D0(obj))
			return obj.flags
		},
		pickup: func(who, crown *crownUpdateTestObject53E1D0, flag1, flag2 int32) uint32 {
			s.events = append(s.events, fmt.Sprintf(
				"pickup:%s:%s:%d:%d",
				crownUpdateTestObjectName53E1D0(who),
				crownUpdateTestObjectName53E1D0(crown),
				flag1,
				flag2,
			))
			return 0xf1234567
		},
		loadField0: func(data *crownUpdateTestData53E1D0) *crownUpdateTestObject53E1D0 {
			s.events = append(s.events, "field0:"+data.name)
			return data.field0
		},
		loadFlagsLow: func(obj *crownUpdateTestObject53E1D0) uint8 {
			s.events = append(s.events, "flags-low:"+crownUpdateTestObjectName53E1D0(obj))
			return uint8(obj.flags)
		},
		clearField0: func(data *crownUpdateTestData53E1D0) {
			s.events = append(s.events, "clear-field0:"+data.name)
			data.field0 = nil
			if s.afterField0Clear != nil {
				s.afterField0Clear()
			}
		},
		loadOwner: func(obj *crownUpdateTestObject53E1D0) *crownUpdateTestObject53E1D0 {
			s.events = append(s.events, "owner:"+crownUpdateTestObjectName53E1D0(obj))
			return obj.owner
		},
		clearOwner: func(obj *crownUpdateTestObject53E1D0) {
			s.events = append(s.events, "clear-owner:"+obj.name)
			obj.owner = nil
		},
		loadRadius: func(obj *crownUpdateTestObject53E1D0) float32 {
			s.events = append(s.events, "radius:"+obj.name)
			return obj.radius
		},
		loadPosX: func(obj *crownUpdateTestObject53E1D0) float32 {
			s.positionXLoads++
			s.events = append(s.events, fmt.Sprintf("pos-x:%s:%d", obj.name, s.positionXLoads))
			if s.positionXLoads == 1 {
				return obj.pos.X
			}
			return obj.pos.X + 10
		},
		loadPosY: func(obj *crownUpdateTestObject53E1D0) float32 {
			s.positionYLoads++
			s.events = append(s.events, fmt.Sprintf("pos-y:%s:%d", obj.name, s.positionYLoads))
			if s.positionYLoads == 1 {
				return obj.pos.Y
			}
			return obj.pos.Y - 4
		},
		loadDirection: func(obj *crownUpdateTestObject53E1D0) int16 {
			s.directionLoads++
			s.events = append(s.events, fmt.Sprintf("direction:%s:%d", obj.name, s.directionLoads))
			return obj.direction[s.directionLoads-1]
		},
		loadDirectionCos: func(direction int16) float32 {
			s.events = append(s.events, fmt.Sprintf("cos:%d", direction))
			return 0.5
		},
		loadDirectionSin: func(direction int16) float32 {
			s.events = append(s.events, fmt.Sprintf("sin:%d", direction))
			return -0.25
		},
		trace: func(from, to types.Pointf, flags uint8) bool {
			s.events = append(s.events, fmt.Sprintf("trace:%d", flags))
			s.tracedFrom = from
			s.tracedTo = to
			return s.traceResult
		},
		move: func(obj *crownUpdateTestObject53E1D0, pos types.Pointf) {
			s.events = append(s.events, "move:"+obj.name)
			s.movedTo = pos
		},
	}
}

func assertCrownUpdateEvents53E1D0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestCrownUpdate53E1D0LiveTargetCallsFourArgumentPickupAndReturns(t *testing.T) {
	target := &crownUpdateTestObject53E1D0{name: "target", flags: 0x40000000}
	data := &crownUpdateTestData53E1D0{name: "cached", target: target}
	crown := &crownUpdateTestObject53E1D0{name: "crown", update: data}
	state := &crownUpdateTestState53E1D0{}

	crownUpdate53E1D0(crown, state.hooks())
	assertCrownUpdateEvents53E1D0(t, state.events, []string{
		"update:crown",
		"target:cached",
		"flags:target",
		"pickup:target:crown:1:1",
	})
}

func TestCrownUpdate53E1D0BlockedTargetUsesCachedFallbackAndLiveOwner(t *testing.T) {
	target := &crownUpdateTestObject53E1D0{name: "target", flags: 0x8000}
	field0 := &crownUpdateTestObject53E1D0{name: "field0", flags: 0x12340020}
	oldData := &crownUpdateTestData53E1D0{name: "old", target: target, field0: field0}
	newData := &crownUpdateTestData53E1D0{name: "new", field0: field0}
	newOwner := &crownUpdateTestObject53E1D0{name: "new-owner", flags: 0x20}
	crown := &crownUpdateTestObject53E1D0{name: "crown", update: oldData}
	state := &crownUpdateTestState53E1D0{
		afterTargetLoad: func() {
			crown.update = newData
		},
		afterField0Clear: func() {
			crown.owner = newOwner
		},
	}

	crownUpdate53E1D0(crown, state.hooks())
	if oldData.field0 != nil || newData.field0 != field0 {
		t.Fatalf("field0 values = (%p, %p), want (nil, %p)", oldData.field0, newData.field0, field0)
	}
	if crown.owner != nil {
		t.Fatal("blocked live owner was not cleared")
	}
	assertCrownUpdateEvents53E1D0(t, state.events, []string{
		"update:crown",
		"target:old",
		"flags:target",
		"field0:old",
		"flags-low:field0",
		"clear-field0:old",
		"owner:crown",
		"flags:new-owner",
		"clear-owner:crown",
	})
}

func TestCrownUpdate53E1D0OwnerGeometryUsesOriginalReadOrder(t *testing.T) {
	owner := &crownUpdateTestObject53E1D0{
		name:      "owner",
		radius:    2.25,
		pos:       types.Pointf{X: 10, Y: -4},
		direction: [2]int16{-32767, -2},
	}
	crown := &crownUpdateTestObject53E1D0{
		name:   "crown",
		radius: 3.5,
		update: &crownUpdateTestData53E1D0{name: "data"},
		owner:  owner,
	}
	state := &crownUpdateTestState53E1D0{traceResult: true}

	crownUpdate53E1D0(crown, state.hooks())
	distance := float64(2.25 + 3.5 + 10)
	wantFrom := types.Pointf{X: 10, Y: -4}
	wantTo := types.Pointf{
		X: float32(distance*0.5 + 20),
		Y: float32(distance*-0.25 - 8),
	}
	if state.tracedFrom != wantFrom || state.tracedTo != wantTo || state.movedTo != wantTo {
		t.Fatalf(
			"trace/move = (%+v, %+v, %+v), want (%+v, %+v, %+v)",
			state.tracedFrom, state.tracedTo, state.movedTo,
			wantFrom, wantTo, wantTo,
		)
	}
	assertCrownUpdateEvents53E1D0(t, state.events, []string{
		"update:crown",
		"target:data",
		"field0:data",
		"owner:crown",
		"flags:owner",
		"radius:owner",
		"pos-x:owner:1",
		"radius:crown",
		"pos-y:owner:1",
		"direction:owner:1",
		"cos:-32767",
		"pos-x:owner:2",
		"direction:owner:2",
		"sin:-2",
		"pos-y:owner:2",
		"trace:5",
		"move:crown",
	})
}

func TestCrownUpdate53E1D0BlockedField0BitUsesLowByteOnly(t *testing.T) {
	field0 := &crownUpdateTestObject53E1D0{name: "field0", flags: 0x00002000}
	data := &crownUpdateTestData53E1D0{name: "data", field0: field0}
	crown := &crownUpdateTestObject53E1D0{name: "crown", update: data}
	state := &crownUpdateTestState53E1D0{}

	crownUpdate53E1D0(crown, state.hooks())
	if data.field0 != field0 {
		t.Fatal("high-byte 0x20 bit cleared Field0")
	}
	assertCrownUpdateEvents53E1D0(t, state.events, []string{
		"update:crown",
		"target:data",
		"field0:data",
		"flags-low:field0",
		"owner:crown",
	})
}

func TestCrownUpdate53E1D0FailedTraceDoesNotMove(t *testing.T) {
	owner := &crownUpdateTestObject53E1D0{name: "owner"}
	crown := &crownUpdateTestObject53E1D0{
		name:   "crown",
		update: &crownUpdateTestData53E1D0{name: "data"},
		owner:  owner,
	}
	state := &crownUpdateTestState53E1D0{}

	crownUpdate53E1D0(crown, state.hooks())
	if len(state.events) == 0 || state.events[len(state.events)-1] != "trace:5" {
		t.Fatalf("last event = %#v, want trace", state.events)
	}
}
