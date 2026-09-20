package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestSelfDestructUpdate53CC90BoundaryAndNativePointer(t *testing.T) {
	tests := []struct {
		name          string
		frame         uint32
		creationFrame uint32
		deleted       bool
	}{
		{name: "creation tick", frame: 100, creationFrame: 100},
		{name: "two ticks", frame: 102, creationFrame: 100},
		{name: "three ticks", frame: 103, creationFrame: 100, deleted: true},
		{name: "later", frame: 150, creationFrame: 100, deleted: true},
		{name: "frame wrap", frame: 1, creationFrame: math.MaxUint32 - 1, deleted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{Field32: tc.creationFrame}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
				t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
			}
			s := new(Server)
			s.SetFrame(tc.frame)
			var deleted *Object
			s.SelfDestructUpdate53CC90(source, SelfDestructUpdateRuntime53CC90{
				DelayedDelete: func(got *Object) { deleted = got },
			})
			if (deleted != nil) != tc.deleted {
				t.Fatalf("deleted = %p, want deleted=%t", deleted, tc.deleted)
			}
			if deleted != nil && deleted != source {
				t.Fatalf("deleted source = %p, want %p", deleted, source)
			}
		})
	}
}

func TestSelfDestructUpdate53CC90RejectsMissingState(t *testing.T) {
	called := false
	new(Server).SelfDestructUpdate53CC90(nil, SelfDestructUpdateRuntime53CC90{
		DelayedDelete: func(*Object) { called = true },
	})
	new(Server).SelfDestructUpdate53CC90(new(Object), SelfDestructUpdateRuntime53CC90{})
	if called {
		t.Fatal("missing state invoked deletion")
	}
}
