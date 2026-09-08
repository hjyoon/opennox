package noxscript

import (
	"math"
	"slices"
	"testing"

	"github.com/opennox/noxscript/ns/asm"
	ns "github.com/opennox/noxscript/ns/v4"
)

type questStatusBuiltinTestImpl struct {
	ns.Implementation
	trace  *[]string
	name   string
	result int
	float  float32
}

func (s *questStatusBuiltinTestImpl) GetQuestStatus(name string) int {
	*s.trace = append(*s.trace, "get-quest-status")
	s.name = name
	return s.result
}

func (s *questStatusBuiltinTestImpl) GetQuestStatusFloat(name string) float32 {
	*s.trace = append(*s.trace, "get-quest-status-float")
	s.name = name
	return s.float
}

type questStatusBuiltinTestVM struct {
	VM
	impl    *questStatusBuiltinTestImpl
	trace   []string
	strings []string
	pushed  []int32
	floats  []float32
}

func (s *questStatusBuiltinTestVM) NoxScript() ns.Implementation {
	return s.impl
}

func (s *questStatusBuiltinTestVM) PopString() string {
	s.trace = append(s.trace, "pop-string")
	v := s.strings[0]
	s.strings = s.strings[1:]
	return v
}

func (s *questStatusBuiltinTestVM) PushI32(v int32) {
	s.trace = append(s.trace, "push-i32")
	s.pushed = append(s.pushed, v)
}

func (s *questStatusBuiltinTestVM) PushF32(v float32) {
	s.trace = append(s.trace, "push-f32")
	s.floats = append(s.floats, v)
}

func TestGetQuestStatusBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	vm := &questStatusBuiltinTestVM{strings: []string{"War01a:Value"}}
	vm.impl = &questStatusBuiltinTestImpl{trace: &vm.trace, result: -2147483648}

	result, ok := CallBuiltin(vm, asm.BuiltinGetQuestStatus)
	if !ok || result != 0 {
		t.Fatalf("GetQuestStatus dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "get-quest-status", "push-i32"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("GetQuestStatus trace = %v, want %v", vm.trace, wantTrace)
	}
	if vm.impl.name != "War01a:Value" {
		t.Fatalf("GetQuestStatus name = %q, want War01a:Value", vm.impl.name)
	}
	if !slices.Equal(vm.pushed, []int32{-2147483648}) {
		t.Fatalf("GetQuestStatus pushed = %v, want [-2147483648]", vm.pushed)
	}
}

func TestGetQuestStatusFloatBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	vm := &questStatusBuiltinTestVM{strings: []string{"War01a:Value"}}
	vm.impl = &questStatusBuiltinTestImpl{
		trace: &vm.trace,
		float: math.Float32frombits(0xffc12345),
	}

	result, ok := CallBuiltin(vm, asm.BuiltinGetQuestStatusFloat)
	if !ok || result != 0 {
		t.Fatalf("GetQuestStatusFloat dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "get-quest-status-float", "push-f32"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("GetQuestStatusFloat trace = %v, want %v", vm.trace, wantTrace)
	}
	if vm.impl.name != "War01a:Value" {
		t.Fatalf("GetQuestStatusFloat name = %q, want War01a:Value", vm.impl.name)
	}
	if len(vm.floats) != 1 {
		t.Fatalf("GetQuestStatusFloat pushed %d values, want 1", len(vm.floats))
	}
	if got := math.Float32bits(vm.floats[0]); got != 0xffc12345 {
		t.Fatalf("GetQuestStatusFloat pushed bits = %08x, want ffc12345", got)
	}
}
