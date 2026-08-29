package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestExpireUpdateNative53DB00UsesNativeObjectAndInclusiveInterval(t *testing.T) {
	source := new(Object)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
			t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
		}
	}

	tests := []struct {
		name    string
		start   uint32
		end     uint32
		frame   uint32
		deleted bool
	}{
		{name: "before start", start: 11, end: 20, frame: 10, deleted: true},
		{name: "at start", start: 11, end: 20, frame: 11},
		{name: "inside", start: 11, end: 20, frame: 15},
		{name: "at end", start: 11, end: 20, frame: 20},
		{name: "after end", start: 11, end: 20, frame: 21, deleted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source.Field32 = tc.start
			source.Field34 = tc.end
			var deleted *Object
			expireUpdateNative53DB00(source, tc.frame, ExpireUpdateRuntime53DB00{
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

func TestExpireUpdateNative53DB00UsesUnsignedFrameComparisons(t *testing.T) {
	source := &Object{Field32: math.MaxUint32 - 1, Field34: math.MaxUint32}
	var deleted *Object
	expireUpdateNative53DB00(source, 1, ExpireUpdateRuntime53DB00{
		DelayedDelete: func(got *Object) { deleted = got },
	})
	if deleted != source {
		t.Fatalf("deleted = %p, want %p for unsigned frame before start", deleted, source)
	}
}
