package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCountInventoryClassNative4E7D70(t *testing.T) {
	if got := Sub_4E7D70(nil, uint32(object.ClassMonster)); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &server.Object{ObjClass: object.ClassPlayer}
	third := &server.Object{ObjClass: object.ClassMonster | object.ClassPlayer, InvNextItem: fourth}
	second := &server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDestroyed, InvNextItem: third}
	first := &server.Object{ObjClass: object.ClassMissile, InvNextItem: second}
	owner := &server.Object{InvFirstItem: first}

	if got := Sub_4E7D70(owner, uint32(object.ClassMonster)); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	if got := Sub_4E7D70(owner, 0); got != 0 {
		t.Fatalf("zero-mask result = %d, want 0", got)
	}
	if first.InvNextItem != second || second.InvNextItem != third || third.InvNextItem != fourth || fourth.InvNextItem != nil {
		t.Fatal("native adapter mutated the inventory list")
	}
	if first.ObjClass != object.ClassMissile || second.ObjClass != object.ClassMonster ||
		third.ObjClass != object.ClassMonster|object.ClassPlayer || fourth.ObjClass != object.ClassPlayer {
		t.Fatal("native adapter mutated an object class")
	}
}
