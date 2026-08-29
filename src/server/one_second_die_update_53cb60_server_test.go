package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestOneSecondDieUpdateNative53CB60UsesNativeObjectAndExactLoadOrder(t *testing.T) {
	source := new(Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}
	source.Field32 = 999
	events := make([]string, 0, 4)
	oneSecondDieUpdateNative53CB60(source, oneSecondDieUpdateNativeDeps53CB60{
		frame: func() uint32 {
			events = append(events, "frame")
			source.Field32 = 100
			return 130
		},
		fps: func() uint32 {
			events = append(events, "fps")
			// A reload after this callback would observe age 10 and skip the
			// deletion. GAME.EXE already cached the value 100 at this point.
			source.Field32 = 120
			return 30
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted source = %p, want %p", got, source)
			}
		},
	})
	if want := []string{"frame", "fps", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestOneSecondDieUpdateNative53CB60BoundaryAndUnsignedWrap(t *testing.T) {
	tests := []struct {
		name          string
		frame         uint32
		creationFrame uint32
		fps           uint32
		deleted       bool
	}{
		{name: "one tick early", frame: 129, creationFrame: 100, fps: 30},
		{name: "exactly one second", frame: 130, creationFrame: 100, fps: 30, deleted: true},
		{name: "later", frame: 131, creationFrame: 100, fps: 30, deleted: true},
		{name: "frame wrap", frame: 4, creationFrame: math.MaxUint32 - 20, fps: 25, deleted: true},
		{name: "zero fps", frame: 7, creationFrame: 7, fps: 0, deleted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{Field32: tc.creationFrame}
			var deleted *Object
			oneSecondDieUpdateNative53CB60(source, oneSecondDieUpdateNativeDeps53CB60{
				frame: func() uint32 { return tc.frame },
				fps:   func() uint32 { return tc.fps },
				delayedDelete: func(got *Object) {
					deleted = got
				},
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
