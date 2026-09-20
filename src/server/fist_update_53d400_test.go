package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestFistUpdate53D400ImpactEffectsWithNativePointer(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)
	source := &Object{
		ObjFlags: object.FlagEnabled,
		PosVec:   types.Ptf(10, 20),
		ZVal:     0,
		Field32:  100,
	}
	source.Shape.Circle.R = 5
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	var events []string
	s.FistUpdate53D400(source, FistUpdateRuntime53D400{
		AudioEvent: func(id uint32, got *Object) {
			events = append(events, fmt.Sprintf("audio:%d:%t", id, got == source))
		},
		MakeScorch: func(position types.Pointf, kind int) {
			events = append(events, fmt.Sprintf("scorch:%v:%d", position, kind))
		},
		SendPointFX: func(code uint8, position types.Pointf) {
			events = append(events, fmt.Sprintf("fx:%d:%v", code, position))
		},
		Earthquake: func(position types.Pointf, magnitude int) {
			events = append(events, fmt.Sprintf("quake:%v:%d", position, magnitude))
		},
		DelayedDelete: func(*Object) { t.Fatal("impact tick deleted fist") },
	})

	want := []string{
		"audio:48:true",
		"scorch:{10 20}:2",
		"fx:138:{15 20}",
		"fx:138:{5 20}",
		"fx:138:{10 25}",
		"quake:{10 20}:30",
	}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.ObjFlags != object.FlagEnabled|object.FlagMarked {
		t.Fatalf("flags = %#08x, want enabled|marked", uint32(source.ObjFlags))
	}
}

func TestFistUpdate53D400MarkedFistDoesNotImpactTwice(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)
	source := &Object{
		ObjFlags: object.FlagMarked,
		PosVec:   types.Ptf(10, 20),
		ZVal:     0,
		Field32:  100,
	}
	runtime := FistUpdateRuntime53D400{
		AudioEvent:    func(uint32, *Object) { t.Fatal("replayed impact audio") },
		MakeScorch:    func(types.Pointf, int) { t.Fatal("replayed impact scorch") },
		SendPointFX:   func(uint8, types.Pointf) { t.Fatal("replayed impact point effect") },
		Earthquake:    func(types.Pointf, int) { t.Fatal("replayed impact earthquake") },
		DelayedDelete: func(*Object) { t.Fatal("grounded fist was deleted") },
	}
	s.FistUpdate53D400(source, runtime)
	s.FistUpdate53D400(nil, runtime)
}

func TestFistUpdate53D400HeightAndLifetimeDeletion(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	source := &Object{
		ObjFlags: object.FlagMarked,
		ZVal:     fistDeleteHeight53D400,
		Field32:  100,
	}
	deleted := 0
	runtime := FistUpdateRuntime53D400{
		DelayedDelete: func(got *Object) {
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
			deleted++
		},
	}

	source.ObjFlags = 0
	s.SetFrame(190)
	s.FistUpdate53D400(source, runtime)
	if deleted != 0 {
		t.Fatalf("deletes before impact mark = %d, want 0", deleted)
	}

	source.ObjFlags = object.FlagMarked
	s.SetFrame(190)
	s.FistUpdate53D400(source, runtime)
	if deleted != 1 {
		t.Fatalf("deletes at strict lifetime boundary = %d, want height delete only", deleted)
	}

	s.SetFrame(191)
	s.FistUpdate53D400(source, runtime)
	if deleted != 3 {
		t.Fatalf("deletes after height and lifetime checks = %d, want 3", deleted)
	}

	source.ZVal = 1
	source.Field32 = math.MaxUint32 - 10
	s.SetFrame(80)
	s.FistUpdate53D400(source, runtime)
	if deleted != 4 {
		t.Fatalf("deletes after frame wrap = %d, want 4", deleted)
	}
}
