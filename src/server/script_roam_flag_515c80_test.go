package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSetRoamFlag515C80WritesOnlyLowByte(t *testing.T) {
	update := &MonsterUpdateData{Field333: 0xdeadbeef}
	monster := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	monster.SetRoamFlag515C80(0x42)
	if update.Field333 != 0xdeadbe42 {
		t.Fatalf("Field333 = %#x, want 0xdeadbe42", update.Field333)
	}
	monster.SetRoamFlag515C80(0xff)
	if update.Field333 != 0xdeadbeff {
		t.Fatalf("Field333 = %#x, want 0xdeadbeff", update.Field333)
	}
	other := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	other.SetRoamFlag515C80(0)
	(*Object)(nil).SetRoamFlag515C80(0)
	(&Object{ObjClass: object.ClassMonster}).SetRoamFlag515C80(0)
	if update.Field333 != 0xdeadbeff {
		t.Fatalf("non-monster or nil target changed Field333 to %#x", update.Field333)
	}
}
