package noxscript

import (
	"slices"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
	ns "github.com/opennox/noxscript/ns/v4"
)

type allegianceBuiltinTestObject struct {
	ns.Obj
}

type allegianceBuiltinTestImpl struct {
	ns.Implementation
	trace *[]string
	obj   ns.Obj
}

func (s *allegianceBuiltinTestImpl) MakeFriendly(obj ns.Obj) {
	*s.trace = append(*s.trace, "make-friendly")
	s.obj = obj
}

func (s *allegianceBuiltinTestImpl) MakeEnemy(obj ns.Obj) {
	*s.trace = append(*s.trace, "make-enemy")
	s.obj = obj
}

func (s *allegianceBuiltinTestImpl) BecomePet(obj ns.Obj) {
	*s.trace = append(*s.trace, "become-pet")
	s.obj = obj
}

func (s *allegianceBuiltinTestImpl) BecomeEnemy(obj ns.Obj) {
	*s.trace = append(*s.trace, "become-enemy")
	s.obj = obj
}

type allegianceBuiltinTestVM struct {
	VM
	impl    *allegianceBuiltinTestImpl
	trace   []string
	objects []ns.Obj
}

func (s *allegianceBuiltinTestVM) NoxScript() ns.Implementation {
	return s.impl
}

func (s *allegianceBuiltinTestVM) PopObjectNS() ns.Obj {
	s.trace = append(s.trace, "pop-object")
	v := s.objects[0]
	s.objects = s.objects[1:]
	return v
}

func TestAllegianceBuiltinsUseNativeObjectDispatch(t *testing.T) {
	tests := []struct {
		name string
		op   asm.Builtin
		call string
	}{
		{name: "MakeFriendly", op: asm.BuiltinMakeFriendly, call: "make-friendly"},
		{name: "MakeEnemy", op: asm.BuiltinMakeEnemy, call: "make-enemy"},
		{name: "BecomePet", op: asm.BuiltinBecomePet, call: "become-pet"},
		{name: "BecomeEnemy", op: asm.BuiltinBecomeEnemy, call: "become-enemy"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &allegianceBuiltinTestObject{}
			vm := &allegianceBuiltinTestVM{objects: []ns.Obj{obj}}
			vm.impl = &allegianceBuiltinTestImpl{trace: &vm.trace}

			result, ok := CallBuiltin(vm, tc.op)
			if !ok || result != 0 {
				t.Fatalf("dispatch = %d/%v, want 0/true", result, ok)
			}
			if want := []string{"pop-object", tc.call}; !slices.Equal(vm.trace, want) {
				t.Fatalf("trace = %v, want %v", vm.trace, want)
			}
			if vm.impl.obj != obj {
				t.Fatalf("object = %p, want %p", vm.impl.obj, obj)
			}
		})
	}
}

func TestAllegianceBuiltinsIgnoreMissingObject(t *testing.T) {
	for _, op := range []asm.Builtin{
		asm.BuiltinMakeFriendly,
		asm.BuiltinMakeEnemy,
		asm.BuiltinBecomePet,
		asm.BuiltinBecomeEnemy,
	} {
		vm := &allegianceBuiltinTestVM{objects: []ns.Obj{nil}}
		vm.impl = &allegianceBuiltinTestImpl{trace: &vm.trace}

		result, ok := CallBuiltin(vm, op)
		if !ok || result != 0 {
			t.Fatalf("builtin %d dispatch = %d/%v, want 0/true", op, result, ok)
		}
		if want := []string{"pop-object"}; !slices.Equal(vm.trace, want) {
			t.Fatalf("builtin %d trace = %v, want %v", op, vm.trace, want)
		}
	}
}
