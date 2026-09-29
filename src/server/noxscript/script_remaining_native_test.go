package noxscript

import (
	"math"
	"slices"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
	ns "github.com/opennox/noxscript/ns/v4"
)

type remainingBuiltinTestObject struct {
	ns.Obj
	trace *[]string
	xp    float32
}

func (o *remainingBuiltinTestObject) GiveXp(xp float32) {
	*o.trace = append(*o.trace, "give-xp")
	o.xp = xp
}

type remainingBuiltinTestImpl struct {
	ns.Implementation
	trace         *[]string
	unknownB8     bool
	unknownB9     bool
	unknownID     int
	halberd       ns.HalberdLevel
	halberdCalled bool
}

func (s *remainingBuiltinTestImpl) Unknownb8(id int) bool {
	*s.trace = append(*s.trace, "unknown-b8")
	s.unknownID = id
	return s.unknownB8
}

func (s *remainingBuiltinTestImpl) Unknownb9(id int) bool {
	*s.trace = append(*s.trace, "unknown-b9")
	s.unknownID = id
	return s.unknownB9
}

func (s *remainingBuiltinTestImpl) SetHalberd(upgrade ns.HalberdLevel) {
	*s.trace = append(*s.trace, "set-halberd")
	s.halberd = upgrade
	s.halberdCalled = true
}

type remainingBuiltinTestVM struct {
	VM
	impl    *remainingBuiltinTestImpl
	trace   []string
	ints    []int32
	floats  []float32
	objects []ns.Obj
	bools   []bool
}

func (s *remainingBuiltinTestVM) NoxScript() ns.Implementation { return s.impl }

func (s *remainingBuiltinTestVM) PopI32() int32 {
	s.trace = append(s.trace, "pop-i32")
	v := s.ints[0]
	s.ints = s.ints[1:]
	return v
}

func (s *remainingBuiltinTestVM) PopF32() float32 {
	s.trace = append(s.trace, "pop-f32")
	v := s.floats[0]
	s.floats = s.floats[1:]
	return v
}

func (s *remainingBuiltinTestVM) PopObjectNS() ns.Obj {
	s.trace = append(s.trace, "pop-object")
	v := s.objects[0]
	s.objects = s.objects[1:]
	return v
}

func (s *remainingBuiltinTestVM) PushBool(v bool) {
	s.trace = append(s.trace, "push-bool")
	s.bools = append(s.bools, v)
}

func TestGiveXpBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	vm := &remainingBuiltinTestVM{floats: []float32{math.Float32frombits(0x42f68000)}}
	obj := &remainingBuiltinTestObject{trace: &vm.trace}
	vm.objects = []ns.Obj{obj}
	vm.impl = &remainingBuiltinTestImpl{trace: &vm.trace}

	result, ok := CallBuiltin(vm, asm.BuiltinGiveXp)
	if !ok || result != 0 {
		t.Fatalf("GiveXp dispatch = %d/%v, want 0/true", result, ok)
	}
	if want := []string{"pop-f32", "pop-object", "give-xp"}; !slices.Equal(vm.trace, want) {
		t.Fatalf("GiveXp trace = %v, want %v", vm.trace, want)
	}
	if got := math.Float32bits(obj.xp); got != 0x42f68000 {
		t.Fatalf("GiveXp bits = %08x, want 42f68000", got)
	}
}

func TestGiveXpBuiltinIgnoresMissingObject(t *testing.T) {
	vm := &remainingBuiltinTestVM{floats: []float32{1}, objects: []ns.Obj{nil}}
	vm.impl = &remainingBuiltinTestImpl{trace: &vm.trace}
	result, ok := CallBuiltin(vm, asm.BuiltinGiveXp)
	if !ok || result != 0 {
		t.Fatalf("GiveXp dispatch = %d/%v, want 0/true", result, ok)
	}
	if want := []string{"pop-f32", "pop-object"}; !slices.Equal(vm.trace, want) {
		t.Fatalf("GiveXp trace = %v, want %v", vm.trace, want)
	}
}

func TestUnknownB8B9BuiltinsUseNativeDispatch(t *testing.T) {
	tests := []struct {
		name string
		op   asm.Builtin
		call string
		want bool
	}{
		{name: "b8", op: asm.BuiltinUnknownb8, call: "unknown-b8", want: true},
		{name: "b9", op: asm.BuiltinUnknownb9, call: "unknown-b9", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := &remainingBuiltinTestVM{ints: []int32{-2}}
			vm.impl = &remainingBuiltinTestImpl{trace: &vm.trace, unknownB8: true, unknownB9: false}
			result, ok := CallBuiltin(vm, tc.op)
			if !ok || result != 0 {
				t.Fatalf("dispatch = %d/%v, want 0/true", result, ok)
			}
			if want := []string{"pop-i32", tc.call, "push-bool"}; !slices.Equal(vm.trace, want) {
				t.Fatalf("trace = %v, want %v", vm.trace, want)
			}
			if vm.impl.unknownID != -2 || !slices.Equal(vm.bools, []bool{tc.want}) {
				t.Fatalf("call id/result = %d/%v, want -2/%v", vm.impl.unknownID, vm.bools, tc.want)
			}
		})
	}
}

func TestSetHalberdBuiltinNativeDispatch(t *testing.T) {
	vm := &remainingBuiltinTestVM{ints: []int32{int32(ns.OblivionWierdling)}}
	vm.impl = &remainingBuiltinTestImpl{trace: &vm.trace}
	result, ok := CallBuiltin(vm, asm.BuiltinSetHalberd)
	if !ok || result != 0 {
		t.Fatalf("SetHalberd dispatch = %d/%v, want 0/true", result, ok)
	}
	if want := []string{"pop-i32", "set-halberd"}; !slices.Equal(vm.trace, want) {
		t.Fatalf("SetHalberd trace = %v, want %v", vm.trace, want)
	}
	if !vm.impl.halberdCalled || vm.impl.halberd != ns.OblivionWierdling {
		t.Fatalf("SetHalberd call = %v/%d, want true/%d", vm.impl.halberdCalled, vm.impl.halberd, ns.OblivionWierdling)
	}
}
