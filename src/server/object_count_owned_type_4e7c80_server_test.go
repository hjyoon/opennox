package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestObjectCountSubOfType4E7C80Adapter(t *testing.T) {
	if got := (*Object)(nil).CountSubOfType(7); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &Object{TypeInd: 7, ObjFlags: object.Flags(0x80000000)}
	third := &Object{TypeInd: 7, ObjFlags: object.FlagDestroyed, Field128: fourth}
	second := &Object{TypeInd: 7, Field128: third}
	first := &Object{TypeInd: 6, ObjFlags: object.FlagDestroyed, Field128: second}
	owner := &Object{Field129: first}

	if got := owner.CountSubOfType(7); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	if got := owner.CountSubOfType(0x00010007); got != 0 {
		t.Fatalf("high type result = %d, want 0", got)
	}
	if first.Field128 != second || second.Field128 != third || third.Field128 != fourth || fourth.Field128 != nil {
		t.Fatal("adapter mutated the owned list")
	}
	if first.TypeInd != 6 || second.TypeInd != 7 || third.TypeInd != 7 || fourth.TypeInd != 7 {
		t.Fatal("adapter mutated an object type")
	}
}
