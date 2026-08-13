package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestObjectCountInventoryWithType4E7D30Adapter(t *testing.T) {
	if got := (*Object)(nil).CountInventoryWithType(7); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &Object{TypeInd: 7, ObjFlags: object.Flags(0x80000000)}
	third := &Object{TypeInd: 7, ObjFlags: object.FlagDestroyed, InvNextItem: fourth}
	second := &Object{TypeInd: 7, InvNextItem: third}
	first := &Object{TypeInd: 6, ObjFlags: object.FlagDestroyed, InvNextItem: second}
	owner := &Object{InvFirstItem: first}

	if got := owner.CountInventoryWithType(0); got != 4 {
		t.Fatalf("zero-query result = %d, want 4", got)
	}
	if got := owner.CountInventoryWithType(7); got != 2 {
		t.Fatalf("typed result = %d, want 2", got)
	}
	if got := owner.CountInventoryWithType(0x00010007); got != 0 {
		t.Fatalf("high type result = %d, want 0", got)
	}
	if first.InvNextItem != second || second.InvNextItem != third || third.InvNextItem != fourth || fourth.InvNextItem != nil {
		t.Fatal("adapter mutated the inventory list")
	}
}
