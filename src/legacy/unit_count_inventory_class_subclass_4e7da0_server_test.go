package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCountInventoryClassSubclassNative4E7DA0(t *testing.T) {
	if got := Sub_4E7DA0(nil, uint32(object.ClassMonster), 0x2000); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000)}
	third := &server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000),
		ObjFlags: object.FlagDestroyed, InvNextItem: fourth,
	}
	second := &server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x1000), InvNextItem: third,
	}
	first := &server.Object{
		ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(0x2000), InvNextItem: second,
	}
	owner := &server.Object{InvFirstItem: first}

	if got := Sub_4E7DA0(owner, uint32(object.ClassMonster), 0x2000); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	if got := Sub_4E7DA0(owner, uint32(object.ClassMonster), 0); got != 0 {
		t.Fatalf("zero-subclass result = %d, want 0", got)
	}
	if first.InvNextItem != second || second.InvNextItem != third || third.InvNextItem != fourth || fourth.InvNextItem != nil {
		t.Fatal("native adapter mutated the inventory list")
	}
	if first.ObjSubClass != object.SubClass(0x2000) || second.ObjSubClass != object.SubClass(0x1000) ||
		third.ObjSubClass != object.SubClass(0x2000) || fourth.ObjSubClass != object.SubClass(0x2000) {
		t.Fatal("native adapter mutated an object subclass")
	}
}
