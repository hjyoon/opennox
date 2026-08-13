package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCountOwnedClassNative4E7CC0(t *testing.T) {
	if got := Sub_4E7CC0(nil, uint32(object.ClassMonster)); got != 0 {
		t.Fatalf("nil result = %d, want 0", got)
	}

	fourth := &server.Object{ObjClass: object.ClassPlayer}
	third := &server.Object{ObjClass: object.ClassMonster | object.ClassPlayer, Field128: fourth}
	second := &server.Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDestroyed, Field128: third}
	first := &server.Object{ObjClass: object.ClassMissile, Field128: second}
	owner := &server.Object{Field129: first}

	if got := Sub_4E7CC0(owner, uint32(object.ClassMonster)); got != 2 {
		t.Fatalf("result = %d, want 2", got)
	}
	if got := Sub_4E7CC0(owner, 0); got != 0 {
		t.Fatalf("zero-mask result = %d, want 0", got)
	}
	if first.Field128 != second || second.Field128 != third || third.Field128 != fourth || fourth.Field128 != nil {
		t.Fatal("native adapter mutated the owned list")
	}
	if first.ObjClass != object.ClassMissile || second.ObjClass != object.ClassMonster ||
		third.ObjClass != object.ClassMonster|object.ClassPlayer || fourth.ObjClass != object.ClassPlayer {
		t.Fatal("native adapter mutated an object class")
	}
}
