package server

import (
	"testing"
	"unsafe"
)

func TestScriptCallbackQueue5025A0PreservesNativePointerTriples(t *testing.T) {
	s := new(NoxScriptVM)
	blocks := make([]ScriptCallback, noxScriptCallbackQueueCap+1)
	callers := make([]Object, len(blocks))
	triggers := make([]Object, len(blocks))
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for name, ptr := range map[string]unsafe.Pointer{
			"block":   unsafe.Pointer(&blocks[0]),
			"caller":  unsafe.Pointer(&callers[0]),
			"trigger": unsafe.Pointer(&triggers[0]),
		} {
			if uintptr(ptr) <= uintptr(^uint32(0)) {
				t.Fatalf("%s pointer %p is not above the PE32 range", name, ptr)
			}
		}
	}

	for i := range blocks {
		s.scriptPushCallback(&blocks[i], &callers[i], &triggers[i])
	}
	if got := len(s.vm.callbacks); got != noxScriptCallbackQueueCap {
		t.Fatalf("queue length = %d, want %d", got, noxScriptCallbackQueueCap)
	}
	for i, got := range s.vm.callbacks {
		want := noxScriptCallback{Block: &blocks[i], Caller: &callers[i], Trigger: &triggers[i]}
		if got != want {
			t.Fatalf("entry %d = %+v, want %+v", i, got, want)
		}
	}
}

func TestScriptCallbackQueue5025E0MatchesAllThreePointers(t *testing.T) {
	block, otherBlock := new(ScriptCallback), new(ScriptCallback)
	caller, otherCaller := new(Object), new(Object)
	trigger, otherTrigger := new(Object), new(Object)
	match := noxScriptCallback{Block: block, Caller: caller, Trigger: trigger}
	nearTrigger := noxScriptCallback{Block: block, Caller: caller, Trigger: otherTrigger}
	nearCaller := noxScriptCallback{Block: block, Caller: otherCaller, Trigger: trigger}
	nearBlock := noxScriptCallback{Block: otherBlock, Caller: caller, Trigger: trigger}
	s := new(NoxScriptVM)
	s.vm.callbacks = []noxScriptCallback{match, nearTrigger, nearCaller, nearBlock, match}

	s.scriptPopCallback(block, caller, trigger)
	want := []noxScriptCallback{nearTrigger, nearCaller, nearBlock}
	if len(s.vm.callbacks) != len(want) {
		t.Fatalf("queue length = %d, want %d", len(s.vm.callbacks), len(want))
	}
	for i, got := range s.vm.callbacks {
		if got != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, got, want[i])
		}
	}
}

func TestScriptCallbackQueue5025E0SkipsShiftedAdjacentMatch(t *testing.T) {
	block := new(ScriptCallback)
	caller, trigger := new(Object), new(Object)
	match := noxScriptCallback{Block: block, Caller: caller, Trigger: trigger}
	other := noxScriptCallback{Block: new(ScriptCallback)}
	s := new(NoxScriptVM)
	s.vm.callbacks = make([]noxScriptCallback, 3, 3)
	copy(s.vm.callbacks, []noxScriptCallback{match, match, other})

	s.scriptPopCallback(block, caller, trigger)
	if got := len(s.vm.callbacks); got != 2 {
		t.Fatalf("queue length = %d, want 2", got)
	}
	if s.vm.callbacks[0] != match || s.vm.callbacks[1] != other {
		t.Fatalf("queue = %+v, want adjacent match retained then other", s.vm.callbacks)
	}
	if got := s.vm.callbacks[:cap(s.vm.callbacks)][2]; got != other {
		t.Fatalf("stale tail = %+v, want unchanged last record %+v", got, other)
	}
}
