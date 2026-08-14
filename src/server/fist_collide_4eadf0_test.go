package server

import (
	"math"
	"reflect"
	"testing"
)

type fistCollideTestData4EADF0 struct {
	damage int32
}

type fistCollideTestDamageFn4EADF0 struct {
	name string
}

type fistCollideTestObject4EADF0 struct {
	name     string
	z        float32
	update   *fistCollideTestData4EADF0
	parent   *fistCollideTestObject4EADF0
	damageFn *fistCollideTestDamageFn4EADF0
}

type fistCollideTestState4EADF0 struct {
	events   []string
	onParent func()
	onDamage func()
	gotFn    *fistCollideTestDamageFn4EADF0
	gotArgs  [3]*fistCollideTestObject4EADF0
	gotValue int32
	gotType  uint32
}

func (s *fistCollideTestState4EADF0) hooks() fistCollideHooks4EADF0[
	*fistCollideTestObject4EADF0,
	*fistCollideTestData4EADF0,
	*fistCollideTestDamageFn4EADF0,
] {
	return fistCollideHooks4EADF0[
		*fistCollideTestObject4EADF0,
		*fistCollideTestData4EADF0,
		*fistCollideTestDamageFn4EADF0,
	]{
		loadZ: func(obj *fistCollideTestObject4EADF0) float32 {
			s.events = append(s.events, "z")
			return obj.z
		},
		loadUpdateData: func(obj *fistCollideTestObject4EADF0) *fistCollideTestData4EADF0 {
			s.events = append(s.events, "update")
			return obj.update
		},
		loadDamage: func(data *fistCollideTestData4EADF0) int32 {
			s.events = append(s.events, "damage")
			return data.damage
		},
		findParentPlayer: func(obj *fistCollideTestObject4EADF0) *fistCollideTestObject4EADF0 {
			s.events = append(s.events, "parent")
			if s.onParent != nil {
				s.onParent()
			}
			return obj.parent
		},
		loadTargetDamage: func(obj *fistCollideTestObject4EADF0) *fistCollideTestDamageFn4EADF0 {
			s.events = append(s.events, "damage-fn")
			return obj.damageFn
		},
		callTargetDamage: func(
			fn *fistCollideTestDamageFn4EADF0,
			target, parent, source *fistCollideTestObject4EADF0,
			damage int32,
			damageType uint32,
		) int32 {
			s.events = append(s.events, "call")
			s.gotFn = fn
			s.gotArgs = [3]*fistCollideTestObject4EADF0{target, parent, source}
			s.gotValue = damage
			s.gotType = damageType
			if s.onDamage != nil {
				s.onDamage()
			}
			return -0x1234567
		},
	}
}

func TestFistCollide4EADF0HeightGateMatchesX87UnorderedPath(t *testing.T) {
	tests := []struct {
		name string
		z    float32
		call bool
	}{
		{name: "positive", z: 1},
		{name: "positive-infinity", z: float32(math.Inf(1))},
		{name: "positive-zero", z: 0, call: true},
		{name: "negative-zero", z: math.Float32frombits(0x80000000), call: true},
		{name: "negative", z: -1, call: true},
		{name: "negative-infinity", z: float32(math.Inf(-1)), call: true},
		{name: "quiet-nan", z: math.Float32frombits(0x7fc12345), call: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &fistCollideTestData4EADF0{damage: 9}
			parent := &fistCollideTestObject4EADF0{name: "parent"}
			source := &fistCollideTestObject4EADF0{name: "source", z: tc.z, update: data, parent: parent}
			target := &fistCollideTestObject4EADF0{
				name:     "target",
				damageFn: &fistCollideTestDamageFn4EADF0{name: "damage"},
			}
			state := &fistCollideTestState4EADF0{}
			fistCollide4EADF0(source, target, (*int)(nil), state.hooks())
			want := []string{"z"}
			if tc.call {
				want = []string{"z", "update", "damage", "parent", "damage-fn", "call"}
			}
			if !reflect.DeepEqual(state.events, want) {
				t.Fatalf("events = %v, want %v", state.events, want)
			}
		})
	}
}

func TestFistCollide4EADF0CachesUpdateBeforeNilTarget(t *testing.T) {
	source := &fistCollideTestObject4EADF0{name: "source", update: nil}
	state := &fistCollideTestState4EADF0{}
	fistCollide4EADF0(source, nil, &struct{}{}, state.hooks())
	if want := []string{"z", "update"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestFistCollide4EADF0CachesDamageBeforeParentAndLoadsLiveCallback(t *testing.T) {
	data := &fistCollideTestData4EADF0{damage: 0x71234567}
	parent := &fistCollideTestObject4EADF0{name: "parent"}
	oldFn := &fistCollideTestDamageFn4EADF0{name: "old"}
	newFn := &fistCollideTestDamageFn4EADF0{name: "new"}
	source := &fistCollideTestObject4EADF0{name: "source", update: data, parent: parent}
	target := &fistCollideTestObject4EADF0{name: "target", damageFn: oldFn}
	state := &fistCollideTestState4EADF0{}
	state.onParent = func() {
		data.damage = -9
		source.update = &fistCollideTestData4EADF0{damage: -10}
		target.damageFn = newFn
	}
	state.onDamage = func() {
		data.damage = -11
	}

	fistCollide4EADF0(source, target, &struct{ unread uint32 }{unread: 0xffffffff}, state.hooks())

	wantEvents := []string{"z", "update", "damage", "parent", "damage-fn", "call"}
	if !reflect.DeepEqual(state.events, wantEvents) {
		t.Fatalf("events = %v, want %v", state.events, wantEvents)
	}
	if state.gotFn != newFn {
		t.Fatalf("Damage callback = %p, want live %p", state.gotFn, newFn)
	}
	if want := [3]*fistCollideTestObject4EADF0{target, parent, source}; state.gotArgs != want {
		t.Fatalf("Damage args = %v, want %v", state.gotArgs, want)
	}
	if state.gotValue != 0x71234567 || state.gotType != fistCollideDamageType4EADF0 {
		t.Fatalf("Damage value/type = %#x/%d", state.gotValue, state.gotType)
	}
}

func TestFistCollide4EADF0NilDataFaultsBeforeParentLookup(t *testing.T) {
	source := &fistCollideTestObject4EADF0{name: "source"}
	target := &fistCollideTestObject4EADF0{name: "target"}
	state := &fistCollideTestState4EADF0{}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		fistCollide4EADF0(source, target, (*int)(nil), state.hooks())
	}()
	if recovered == nil {
		t.Fatal("nil update-data did not fault")
	}
	if want := []string{"z", "update", "damage"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestFistCollide4EADF0NilSourceFaultsBeforeTargetOrUpdate(t *testing.T) {
	target := &fistCollideTestObject4EADF0{name: "target"}
	state := &fistCollideTestState4EADF0{}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		fistCollide4EADF0(nil, target, (*int)(nil), state.hooks())
	}()
	if recovered == nil {
		t.Fatal("nil source did not fault")
	}
	if want := []string{"z"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}
