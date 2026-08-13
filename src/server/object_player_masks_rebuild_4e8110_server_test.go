package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestObjectPlayerMasksRebuildNative4E8110UsesNativeFields(t *testing.T) {
	unit := &Object{TypeInd: 17}
	player := &Player{PlayerUnit: unit}
	second := &Object{ObjClass: object.ClassPlayer, Field34: 0x55667788, Field35: 0xffffffff, Field36: 0xffffffff, Field37: 0x99aabbcc}
	first := &Object{ObjClass: object.ClassMonster, Field34: 0x11223344, Field35: 0x1f, Field36: 0x0f, Field37: 0x55667788, ObjNext: second}
	got := objectPlayerMasksRebuildNative4E8110(34, objectPlayerMasksRebuildNativeDeps4E8110{
		playerByInd: func(ind int32) *Player {
			if ind != 34 {
				t.Fatalf("player index = %d, want 34", ind)
			}
			return player
		},
		firstObject: func() *Object { return first },
		nextObject:  func(obj *Object) *Object { return obj.Next() },
		isHostile: func(gotUnit, obj *Object) int32 {
			if gotUnit != unit || (obj != first && obj != second) {
				t.Fatalf("hostility args = (%p, %p)", gotUnit, obj)
			}
			return 1
		},
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if first.Field35 != 0x1f || first.Field36 != 0x0f || second.Field35 != 0xffffffff || second.Field36 != 0xffffffff {
		t.Fatalf("masks = (%#x, %#x), (%#x, %#x)", first.Field35, first.Field36, second.Field35, second.Field36)
	}
	if first.Field34 != 0x11223344 || first.Field37 != 0x55667788 || second.Field34 != 0x55667788 || second.Field37 != 0x99aabbcc {
		t.Fatal("native adapter changed an adjacent field")
	}
}

func TestServerRebuildObjectPlayerMasks4E8110UsesActualLists(t *testing.T) {
	s := &Server{}
	s.Players.list = []Player{{Active: 1}}
	second := &Object{Field35: 0xffffffff, Field36: 0xffffffff}
	first := &Object{Field35: 0xffffffff, Field36: 0xffffffff, ObjNext: second}
	s.Objs.List = first
	got := s.RebuildObjectPlayerMasks4E8110(0)
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if first.Field35 != 0xfffffffe || first.Field36 != 0xfffffffe || second.Field35 != 0xfffffffe || second.Field36 != 0xfffffffe {
		t.Fatalf("masks = (%#x, %#x), (%#x, %#x), want all 0xfffffffe", first.Field35, first.Field36, second.Field35, second.Field36)
	}
}

func TestServerRebuildObjectPlayerMasks4E8110MissingPlayerSkipsObjects(t *testing.T) {
	s := &Server{}
	s.Objs.List = &Object{Field35: 0x11223344, Field36: 0x55667788}
	if got := s.RebuildObjectPlayerMasks4E8110(-1); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if s.Objs.List.Field35 != 0x11223344 || s.Objs.List.Field36 != 0x55667788 {
		t.Fatal("missing player touched the object list")
	}
}
