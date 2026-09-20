package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestArachnaphobiaUpdate53DA60SpawnsWithNativeOwner(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	s.SetTickRate(30)
	owner := new(Object)
	spider := new(Object)
	source := &Object{
		Field32:  95,
		Field34:  99,
		ObjOwner: owner,
		PosVec:   types.Pointf{X: 12.5, Y: -7.25},
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	var gotObject, gotOwner *Object
	var gotPosition types.Pointf
	s.ArachnaphobiaUpdate53DA60(source, ArachnaphobiaUpdateRuntime53DA60{
		NewObjectByTypeID: func(id string) *Object {
			if id != "SmallSpider" {
				t.Fatalf("type ID = %q, want SmallSpider", id)
			}
			return spider
		},
		CreateAt: func(object, owner *Object, position types.Pointf) {
			gotObject, gotOwner, gotPosition = object, owner, position
		},
		RandomInt: func(minimum, maximum int) int {
			if minimum != 1 || maximum != 5 {
				t.Fatalf("random range = %d..%d, want 1..5", minimum, maximum)
			}
			return 4
		},
		DelayedDelete: func(*Object) { t.Fatal("young source was deleted") },
	})
	if gotObject != spider || gotOwner != owner || gotPosition != source.PosVec {
		t.Fatalf("create = %p/%p/%+v, want %p/%p/%+v", gotObject, gotOwner, gotPosition, spider, owner, source.PosVec)
	}
	if source.Field34 != 104 {
		t.Fatalf("next spawn frame = %d, want 104", source.Field34)
	}
}

func TestArachnaphobiaUpdate53DA60StrictBoundariesAndWrap(t *testing.T) {
	s := new(Server)
	s.SetTickRate(30)
	s.SetFrame(100)
	source := &Object{Field32: 10, Field34: 100}
	spawned := 0
	deleted := 0
	runtime := ArachnaphobiaUpdateRuntime53DA60{
		NewObjectByTypeID: func(string) *Object { spawned++; return nil },
		CreateAt:          func(*Object, *Object, types.Pointf) { t.Fatal("nil spider was created") },
		RandomInt:         func(int, int) int { return 5 },
		DelayedDelete:     func(*Object) { deleted++ },
	}
	s.ArachnaphobiaUpdate53DA60(source, runtime)
	if spawned != 0 || deleted != 0 || source.Field34 != 100 {
		t.Fatalf("at boundary = spawned %d, deleted %d, next %d", spawned, deleted, source.Field34)
	}

	s.SetFrame(101)
	s.ArachnaphobiaUpdate53DA60(source, runtime)
	if spawned != 1 || deleted != 1 || source.Field34 != 106 {
		t.Fatalf("after boundary = spawned %d, deleted %d, next %d; want 1/1/106", spawned, deleted, source.Field34)
	}

	s.SetFrame(math.MaxUint32 - 1)
	source.Field32 = math.MaxUint32 - 1
	source.Field34 = math.MaxUint32 - 2
	s.ArachnaphobiaUpdate53DA60(source, runtime)
	if source.Field34 != 3 || deleted != 1 {
		t.Fatalf("wrapped state = next %d, deleted %d; want 3/1", source.Field34, deleted)
	}
	s.ArachnaphobiaUpdate53DA60(nil, runtime)
}
