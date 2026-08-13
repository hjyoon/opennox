package server

import (
	"testing"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestObjectPlayerMasksClearNative4E80C0UsesNativeFields(t *testing.T) {
	second := &Object{Field34: 0x55667788, Field35: 0xffffffff, Field36: 0xffffffff, Field37: 0x99aabbcc}
	first := &Object{Field34: 0x11223344, Field35: 0x1f, Field36: 0x0f, Field37: 0x55667788, ObjNext: second}
	got := objectPlayerMasksClearNative4E80C0(34, objectPlayerMasksClearNativeDeps4E80C0{
		firstObject: func() *Object { return first },
		nextObject:  func(obj *Object) *Object { return obj.Next() },
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if first.Field35 != 0x1b || first.Field36 != 0x0b || second.Field35 != 0xfffffffb || second.Field36 != 0xfffffffb {
		t.Fatalf("masks = (%#x, %#x), (%#x, %#x)", first.Field35, first.Field36, second.Field35, second.Field36)
	}
	if first.Field34 != 0x11223344 || first.Field37 != 0x55667788 || second.Field34 != 0x55667788 || second.Field37 != 0x99aabbcc {
		t.Fatal("native adapter changed an adjacent field")
	}
}

func TestServerClearObjectPlayerMasks4E80C0UsesActualObjectList(t *testing.T) {
	s := &Server{}
	second := &Object{Field35: 0xffffffff, Field36: 0xffffffff}
	first := &Object{Field35: 0xffffffff, Field36: 0xffffffff, ObjNext: second}
	s.Objs.List = first
	s.ClearObjectPlayerMasks4E80C0(ntype.PlayerInd(-1))
	if first.Field35 != 0x7fffffff || first.Field36 != 0x7fffffff || second.Field35 != 0x7fffffff || second.Field36 != 0x7fffffff {
		t.Fatalf("masks = (%#x, %#x), (%#x, %#x), want all 0x7fffffff", first.Field35, first.Field36, second.Field35, second.Field36)
	}
}

func TestServerClearObjectPlayerMasks4E80C0EmptyList(t *testing.T) {
	s := &Server{}
	s.ClearObjectPlayerMasks4E80C0(7)
}
