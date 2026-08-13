package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCountSlavesNative4E7CF0(t *testing.T) {
	if got := Nox_xxx_unitCountSlaves_4E7CF0(nil, uint32(object.ClassMonster), 0x2000); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000)}
	third := &server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x2000),
		ObjFlags: object.FlagDestroyed, Field128: fourth,
	}
	second := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(0x1000), Field128: third}
	first := &server.Object{ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(0x2000), Field128: second}
	owner := &server.Object{Field129: first}

	if got := Nox_xxx_unitCountSlaves_4E7CF0(owner, uint32(object.ClassMonster), 0x2000); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	if got := Nox_xxx_unitCountSlaves_4E7CF0(owner, uint32(object.ClassMonster), 0); got != 0 {
		t.Fatalf("zero-subclass result = %d, want 0", got)
	}
	if first.Field128 != second || second.Field128 != third || third.Field128 != fourth || fourth.Field128 != nil {
		t.Fatal("native adapter mutated the owned list")
	}
	if first.ObjSubClass != object.SubClass(0x2000) || second.ObjSubClass != object.SubClass(0x1000) ||
		third.ObjSubClass != object.SubClass(0x2000) || fourth.ObjSubClass != object.SubClass(0x2000) {
		t.Fatal("native adapter mutated an object subclass")
	}
}
