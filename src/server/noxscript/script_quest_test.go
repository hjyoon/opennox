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
	trace        *[]string
	name         string
	result       int
	float        float32
	setInt       int
	setFloat     float32
	setIntSeen   bool
	setFloatSeen bool
}

func (s *questStatusBuiltinTestImpl) SetQuestStatus(status int, name string) {
	*s.trace = append(*s.trace, "set-quest-status")
	s.name = name
	s.setInt = status
	s.setIntSeen = true
}

func (s *questStatusBuiltinTestImpl) SetQuestStatusFloat(status float32, name string) {
	*s.trace = append(*s.trace, "set-quest-status-float")
	s.name = name
	s.setFloat = status
	s.setFloatSeen = true
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

func (s *questStatusBuiltinTestImpl) ResetQuestStatus(name string) {
	*s.trace = append(*s.trace, "reset-quest-status")
	s.name = name
}

type questStatusBuiltinTestVM struct {
	VM
	impl      *questStatusBuiltinTestImpl
	trace     []string
	ints      []int32
	strings   []string
	pushed    []int32
	popFloats []float32
	floats    []float32
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

func (s *questStatusBuiltinTestVM) PopI32() int32 {
	s.trace = append(s.trace, "pop-i32")
	v := s.ints[0]
	s.ints = s.ints[1:]
	return v
}

func (s *questStatusBuiltinTestVM) PopF32() float32 {
	s.trace = append(s.trace, "pop-f32")
	v := s.popFloats[0]
	s.popFloats = s.popFloats[1:]
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

func TestSetQuestStatusBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	vm := &questStatusBuiltinTestVM{
		ints:    []int32{-2147483648},
		strings: []string{"War01a:Value"},
	}
	vm.impl = &questStatusBuiltinTestImpl{trace: &vm.trace}

	result, ok := CallBuiltin(vm, asm.BuiltinSetQuestStatus)
	if !ok || result != 0 {
		t.Fatalf("SetQuestStatus dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "pop-i32", "set-quest-status"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("SetQuestStatus trace = %v, want %v", vm.trace, wantTrace)
	}
	if !vm.impl.setIntSeen || vm.impl.name != "War01a:Value" || vm.impl.setInt != -2147483648 {
		t.Fatalf("SetQuestStatus call = seen %v, name %q, value %d", vm.impl.setIntSeen, vm.impl.name, vm.impl.setInt)
	}
}

func TestSetQuestStatusFloatBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	want := math.Float32frombits(0xffc12345)
	vm := &questStatusBuiltinTestVM{
		strings:   []string{"War01a:Value"},
		popFloats: []float32{want},
	}
	vm.impl = &questStatusBuiltinTestImpl{trace: &vm.trace}

	result, ok := CallBuiltin(vm, asm.BuiltinSetQuestStatusFloat)
	if !ok || result != 0 {
		t.Fatalf("SetQuestStatusFloat dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "pop-f32", "set-quest-status-float"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("SetQuestStatusFloat trace = %v, want %v", vm.trace, wantTrace)
	}
	if !vm.impl.setFloatSeen || vm.impl.name != "War01a:Value" || math.Float32bits(vm.impl.setFloat) != 0xffc12345 {
		t.Fatalf("SetQuestStatusFloat call = seen %v, name %q, bits %08x", vm.impl.setFloatSeen, vm.impl.name, math.Float32bits(vm.impl.setFloat))
	}
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

func TestResetQuestStatusBuiltinNativeDispatchAndStackOrder(t *testing.T) {
	vm := &questStatusBuiltinTestVM{strings: []string{"War01a:*"}}
	vm.impl = &questStatusBuiltinTestImpl{trace: &vm.trace}

	result, ok := CallBuiltin(vm, asm.BuiltinResetQuestStatus)
	if !ok || result != 0 {
		t.Fatalf("ResetQuestStatus dispatch = %d/%v, want 0/true", result, ok)
	}
	wantTrace := []string{"pop-string", "reset-quest-status"}
	if !slices.Equal(vm.trace, wantTrace) {
		t.Fatalf("ResetQuestStatus trace = %v, want %v", vm.trace, wantTrace)
	}
	if vm.impl.name != "War01a:*" {
		t.Fatalf("ResetQuestStatus name = %q, want War01a:*", vm.impl.name)
	}
}
