package server

import (
	"reflect"
	"testing"
)

type damageCollideTestData4E9430 struct {
	damage     uint8
	damageType int32
}

type damageCollideTestHealth4E9430 struct{}

type damageCollideTestObject4E9430 struct {
	data   *damageCollideTestData4E9430
	health *damageCollideTestHealth4E9430
	parent *damageCollideTestObject4E9430
}

type damageCollideTestArgs4E9430 struct {
	target, source, attacker *damageCollideTestObject4E9430
	damage                   int32
	damageType               int32
}

type damageCollideTestState4E9430 struct {
	events         []string
	frameLow       uint8
	callbackResult int32
	got            damageCollideTestArgs4E9430
}

func (s *damageCollideTestState4E9430) hooks() damageCollideHooks4E9430[
	*damageCollideTestObject4E9430,
	*damageCollideTestData4E9430,
	*damageCollideTestHealth4E9430,
] {
	return damageCollideHooks4E9430[
		*damageCollideTestObject4E9430,
		*damageCollideTestData4E9430,
		*damageCollideTestHealth4E9430,
	]{
		loadCollideData: func(obj *damageCollideTestObject4E9430) *damageCollideTestData4E9430 {
			s.events = append(s.events, "collide-data")
			return obj.data
		},
		loadHealth: func(obj *damageCollideTestObject4E9430) *damageCollideTestHealth4E9430 {
			s.events = append(s.events, "health")
			return obj.health
		},
		loadDamage: func(data *damageCollideTestData4E9430) uint8 {
			s.events = append(s.events, "damage")
			return data.damage
		},
		loadFrameLow: func() uint8 {
			s.events = append(s.events, "frame")
			return s.frameLow
		},
		loadDamageType: func(data *damageCollideTestData4E9430) int32 {
			s.events = append(s.events, "damage-type")
			return data.damageType
		},
		findParent: func(obj *damageCollideTestObject4E9430) *damageCollideTestObject4E9430 {
			s.events = append(s.events, "parent")
			return obj.parent
		},
		damage: func(target, source, attacker *damageCollideTestObject4E9430, damage, damageType int32) int32 {
			s.events = append(s.events, "callback")
			s.got = damageCollideTestArgs4E9430{
				target: target, source: source, attacker: attacker,
				damage: damage, damageType: damageType,
			}
			return s.callbackResult
		},
	}
}

func TestDamageCollide4E9430TargetAndHealthGates(t *testing.T) {
	source := &damageCollideTestObject4E9430{}
	state := &damageCollideTestState4E9430{}
	damageCollide4E9430(source, nil, struct{ forbidden *int }{}, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"collide-data"}) {
		t.Fatalf("nil-target events = %v", state.events)
	}

	state = &damageCollideTestState4E9430{}
	target := &damageCollideTestObject4E9430{}
	damageCollide4E9430(source, target, struct{ forbidden *int }{}, state.hooks())
	if !reflect.DeepEqual(state.events, []string{"collide-data", "health"}) {
		t.Fatalf("missing-health events = %v", state.events)
	}
}

func TestDamageCollide4E9430DamageByteAndFrameRules(t *testing.T) {
	tests := []struct {
		name      string
		damage    uint8
		frameLow  uint8
		want      int32
		wantFrame bool
	}{
		{name: "zero", damage: 0, frameLow: 1, want: 0},
		{name: "one even", damage: 1, frameLow: 0xfe, want: 0, wantFrame: true},
		{name: "one odd", damage: 1, frameLow: 0xff, want: 1, wantFrame: true},
		{name: "two", damage: 2, frameLow: 1, want: 1},
		{name: "three floors", damage: 3, frameLow: 1, want: 1},
		{name: "maximum", damage: 0xff, frameLow: 1, want: 0x7f},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parent := &damageCollideTestObject4E9430{}
			source := &damageCollideTestObject4E9430{
				data:   &damageCollideTestData4E9430{damage: tc.damage, damageType: -17},
				parent: parent,
			}
			target := &damageCollideTestObject4E9430{health: &damageCollideTestHealth4E9430{}}
			state := &damageCollideTestState4E9430{frameLow: tc.frameLow, callbackResult: -99}
			damageCollide4E9430(source, target, [2]float32{1, -2}, state.hooks())

			wantEvents := []string{"collide-data", "health", "damage"}
			if tc.wantFrame {
				wantEvents = append(wantEvents, "frame")
			}
			wantEvents = append(wantEvents, "damage-type", "parent", "callback")
			if !reflect.DeepEqual(state.events, wantEvents) {
				t.Fatalf("events = %v, want %v", state.events, wantEvents)
			}
			wantArgs := damageCollideTestArgs4E9430{
				target: target, source: parent, attacker: source,
				damage: tc.want, damageType: -17,
			}
			if state.got != wantArgs {
				t.Fatalf("callback args = %+v, want %+v", state.got, wantArgs)
			}
		})
	}
}

func TestDamageCollide4E9430CachesDataAndLoadsTypeBeforeParent(t *testing.T) {
	initial := &damageCollideTestData4E9430{damage: 1, damageType: 7}
	replacement := &damageCollideTestData4E9430{damage: 200, damageType: 300}
	parent := &damageCollideTestObject4E9430{}
	source := &damageCollideTestObject4E9430{data: initial, parent: parent}
	target := &damageCollideTestObject4E9430{health: &damageCollideTestHealth4E9430{}}
	state := &damageCollideTestState4E9430{frameLow: 1}
	hooks := state.hooks()
	baseHealth := hooks.loadHealth
	hooks.loadHealth = func(obj *damageCollideTestObject4E9430) *damageCollideTestHealth4E9430 {
		health := baseHealth(obj)
		source.data = replacement
		return health
	}
	baseFrame := hooks.loadFrameLow
	hooks.loadFrameLow = func() uint8 {
		frame := baseFrame()
		initial.damageType = 19
		return frame
	}
	baseParent := hooks.findParent
	hooks.findParent = func(obj *damageCollideTestObject4E9430) *damageCollideTestObject4E9430 {
		initial.damageType = 23
		return baseParent(obj)
	}

	damageCollide4E9430(source, target, (*int)(nil), hooks)
	if state.got.damage != 1 || state.got.damageType != 19 {
		t.Fatalf("damage/type = %d/%d, want 1/19", state.got.damage, state.got.damageType)
	}
	if got := countDamageCollideEvent4E9430(state.events, "collide-data"); got != 1 {
		t.Fatalf("collide-data loads = %d, events %v", got, state.events)
	}
}

func TestDamageCollide4E9430NilDataFaultIsDelayed(t *testing.T) {
	source := &damageCollideTestObject4E9430{}
	target := &damageCollideTestObject4E9430{health: &damageCollideTestHealth4E9430{}}
	state := &damageCollideTestState4E9430{}
	defer func() {
		if recover() == nil {
			t.Fatal("valid target with nil collide data did not fault")
		}
		if !reflect.DeepEqual(state.events, []string{"collide-data", "health", "damage"}) {
			t.Fatalf("fault events = %v", state.events)
		}
	}()
	damageCollide4E9430(source, target, (*int)(nil), state.hooks())
}

func countDamageCollideEvent4E9430(events []string, want string) int {
	n := 0
	for _, event := range events {
		if event == want {
			n++
		}
	}
	return n
}
